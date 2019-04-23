package parseutil

import (
	"os"
	"strings"
	"unicode"

	"github.com/jmigpin/editor/util/iout/iorw"
	"github.com/jmigpin/editor/util/osutil"
	"github.com/jmigpin/editor/util/statemach"
)

// parsed formats:
// 	<filename:line?:col?>
type Resource struct {
	Path         string
	RawPath      string
	Line, Column int

	ExpandedMin, ExpandedMax int
}

func ParseResource(rd iorw.Reader, index int) (*Resource, error) {
	escape := osutil.EscapeRune

	l, r := ExpandIndexesEscape(rd, index, false, isResourceRune, escape)

	res := &Resource{ExpandedMin: l, ExpandedMax: r}

	p := &ResParser{res: res, escape: escape, pathSep: os.PathSeparator}
	rd2 := iorw.NewLimitedReader(rd, l, r, 0)
	err := p.start(rd2)
	if err != nil {
		return nil, err
	}
	return res, nil
}

//----------

type ResParser struct {
	sc  *statemach.Scanner
	st  func() bool // state
	err error

	escape  rune
	pathSep rune

	res *Resource
}

func (p *ResParser) start(r iorw.Reader) error {
	p.sc = statemach.NewScanner(r)
	// state loop
	p.st = p.pathHeader
	for {
		if p.st == nil || !p.st() {
			break
		}
	}
	return p.err
}

//----------

func (p *ResParser) pathHeader() bool {
	if !p.path() {
		p.err = p.sc.Errorf("path")
		return false
	}
	_ = p.lineCol()
	p.st = nil
	return true
}

func (p *ResParser) path() bool {
	ok := p.sc.RewindOnFalse(func() bool {
		_ = p.pathItem() // optional
		pathSepFn := func(ru rune) bool { return ru == p.pathSep }
		for {
			if p.sc.Match.End() {
				break
			}
			if !p.sc.Match.FnLoop(pathSepFn) { // any number of pathsep
				break
			}
			if !p.pathItem() {
				break
			}
		}
		return !p.sc.Empty()
	})
	if ok {
		s := p.sc.Value()
		p.res.RawPath = s

		// filter
		s = RemoveEscapes(s, p.escape)
		s = CleanMultiplePathSeps(s, p.pathSep)
		p.res.Path = s

		p.sc.Advance()
		return true
	}
	return false
}

func (p *ResParser) pathItem() bool {
	return p.sc.RewindOnFalse(func() bool {
		isPathItemRune := isPathItemRuneFn(p.escape, p.pathSep)
		for p.sc.Match.Escape(p.escape) ||
			p.sc.Match.Fn(isPathItemRune) {
		}
		return !p.sc.Empty()
	})
}

//----------

func (p *ResParser) lineCol() bool {
	return p.sc.RewindOnFalse(func() bool {
		// line sep
		if !p.sc.Match.Rune(':') {
			return false
		}
		p.sc.Advance()
		// line
		v, err := p.sc.Match.IntValueAdvance()
		if err != nil {
			return false
		}
		p.res.Line = v

		_ = p.sc.RewindOnFalse(func() bool {
			// column sep
			if !p.sc.Match.Rune(':') {
				return false
			}
			p.sc.Advance()
			// column
			v, err = p.sc.Match.IntValueAdvance()
			if err != nil {
				return false
			}
			p.res.Column = v
			return true
		})

		return true
	})
}

//----------

func CleanMultiplePathSeps(str string, sep rune) string {
	w := []rune{}
	added := false
	for _, ru := range str {
		if ru == sep {
			if !added {
				added = true
				w = append(w, ru)
			}
		} else {
			added = false
			w = append(w, ru)
		}
	}
	return string(w)
}

//----------

var ExtraRunes = `_-~.%@&?=#` + `\\^` + `/` + ` ` + `()[]{}<>` + `:`

var ResourceExtraRunes = RunesExcept(ExtraRunes, ""+
	" "+ // space must be escaped
	"()[]<>"+ // usually used around filenames in various outputs
	"")

var PathItemExtraRunes = RunesExcept(ExtraRunes, ""+
	" "+ // space must be escaped
	"()[]<>"+ // usually used around filenames in various outputs
	":"+ // line/column
	"")

//----------

func isResourceRune(ru rune) bool {
	return unicode.IsLetter(ru) || unicode.IsDigit(ru) ||
		strings.ContainsRune(ResourceExtraRunes, ru)
}

func isPathItemRuneFn(escape, pathSep rune) func(ru rune) bool {
	// must be escaped:
	// 	escape: must be escaped
	// 	pathSeparator: not part of path item
	runes := RunesExcept(PathItemExtraRunes, string(escape)+string(pathSep))

	// return function
	return func(ru rune) bool {
		return unicode.IsLetter(ru) || unicode.IsDigit(ru) ||
			strings.ContainsRune(runes, ru)
	}
}

func EscapeFilename(str string) string {
	// windows note: if ':' is escaped, then it might have problems parsing compiler output lines with line/col. This way the volume name (ex: "C://") needs an escape (ex: "C^://") and parsing <filename:line:col> works.

	escape := osutil.EscapeRune
	mustBeEscaped := " ()[]<>:" + string(escape)
	return AddEscapes(str, escape, mustBeEscaped)
}

//----------

func RunesExcept(runes, except string) string {
	drop := func(ru rune) rune {
		if strings.ContainsRune(except, ru) {
			return -1
		}
		return ru
	}
	return strings.Map(drop, runes)
}

//----------