package textutil

import (
	"unicode"

	"github.com/jmigpin/editor/util/iout"
	"github.com/jmigpin/editor/util/uiutil/widget"
)

func StartOfLine(te *widget.TextEdit, sel bool) error {
	tc := te.TextCursor

	ci := tc.Index()
	i, err := iout.LineStartIndex(tc.RW(), ci)
	if err != nil {
		return err
	}

	// stop at first non blank rune from the left
	for j := 0; j < 500; j++ {
		ru, _, err := tc.RW().ReadRuneAt(i + j)
		if err != nil {
			return err
		}
		if !unicode.IsSpace(ru) {
			i += j
			break
		}
	}

	tc.SetSelectionUpdate(sel, i)
	return nil
}
