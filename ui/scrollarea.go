package ui

import (
	"image"

	"golang.org/x/image/math/fixed"

	"github.com/jmigpin/editor/uiutil/widget"
	"github.com/jmigpin/editor/xgbutil/evreg"
	"github.com/jmigpin/editor/xgbutil/xinput"
)

type ScrollArea struct {
	widget.ScrollArea
	ui            *UI
	ta            *TextArea
	buttonPressed bool
	evUnreg       evreg.Unregister

	disableTextAreaOffsetYEvent bool
}

func NewScrollArea(ui *UI, ta *TextArea) *ScrollArea {
	sa := &ScrollArea{ui: ui, ta: ta}

	sa.ScrollArea.Init(ui)
	widget.AppendChilds(sa, ta)

	r1 := ui.EvReg.Add(xinput.ButtonPressEventId,
		&evreg.Callback{sa.onButtonPress})
	r2 := ui.EvReg.Add(xinput.ButtonReleaseEventId,
		&evreg.Callback{sa.onButtonRelease})
	r3 := ui.EvReg.Add(xinput.MotionNotifyEventId,
		&evreg.Callback{sa.onMotionNotify})
	sa.evUnreg.Add(r1, r2, r3)

	// textarea set text
	sa.ta.EvReg.Add(TextAreaSetStrEventId,
		&evreg.Callback{func(ev0 interface{}) {
			sa.CalcChildsBounds()
			sa.MarkNeedsPaint()
		}})
	// textarea set offset y
	sa.ta.EvReg.Add(TextAreaSetOffsetYEventId,
		&evreg.Callback{func(ev0 interface{}) {
			if !sa.disableTextAreaOffsetYEvent {
				sa.CalcChildsBounds()
				sa.MarkNeedsPaint()
			}
		}})

	return sa
}
func (sa *ScrollArea) Close() {
	sa.evUnreg.UnregisterAll()
}

func (sa *ScrollArea) CalcChildsBounds() {
	// measure textarea to have accurate str height
	// TODO: needs improvement, using scrollwidth from widget.scrollarea
	b := sa.Bounds()
	b.Max.X -= sa.ScrollWidth
	_ = sa.ta.Measure(b.Sub(b.Min).Max)

	// calc position using int26_6 values cast to floats
	dy := float64(fixed.I(sa.Bounds().Dy()))
	offset := float64(sa.ta.OffsetY())
	height := float64(sa.taHeight())
	sa.ScrollArea.CalcPosition(offset, height, dy)

	sa.ScrollArea.CalcChildsBounds()
}

func (sa *ScrollArea) CalcPositionFromPoint(p *image.Point) {
	// Dragging the scrollbar, updates textarea offset

	sa.ScrollArea.CalcPositionFromPoint(p)

	sa.disableTextAreaOffsetYEvent = true // ignore loop event

	// set textarea offset
	pp := sa.VBarPositionPercent()
	oy := fixed.Int26_6(pp * float64(sa.taHeight()))
	sa.setTaOffsetY(oy)

	sa.disableTextAreaOffsetYEvent = false

	sa.CalcChildsBounds()
	sa.MarkNeedsPaint()
}

func (sa *ScrollArea) CalcPositionFromScroll(up bool) {
	mult := 1
	if up {
		mult = -1
	}

	sa.disableTextAreaOffsetYEvent = true // ignore loop event

	// set textarea offset
	scrollLines := 4
	v := fixed.Int26_6(scrollLines*mult) * sa.ta.LineHeight()
	sa.setTaOffsetY(sa.ta.OffsetY() + v)

	sa.disableTextAreaOffsetYEvent = false

	sa.CalcChildsBounds()
	sa.MarkNeedsPaint()
}

func (sa *ScrollArea) setTaOffsetY(v fixed.Int26_6) {
	dy := fixed.I(sa.Bounds().Dy())
	max := sa.taHeight() - dy
	if v > max {
		v = max
	}
	sa.ta.SetOffsetY(v)
}

func (sa *ScrollArea) taHeight() fixed.Int26_6 {
	// extra height allows to scroll past the str height
	dy := fixed.I(sa.Bounds().Dy())
	extra := dy - 2*sa.ta.LineHeight() // keep something visible

	return sa.ta.StrHeight() + extra
}

func (sa *ScrollArea) onButtonPress(ev0 interface{}) {
	ev := ev0.(*xinput.ButtonPressEvent)

	if ev.Point.In(sa.Bounds()) && !ev.Point.In(*sa.VBarBounds()) {
		// TODO: use ta.scrollup
		switch {
		case ev.Button.Button(4): // wheel up
			sa.CalcPositionFromScroll(true)
		case ev.Button.Button(5): // wheel down
			sa.CalcPositionFromScroll(false)
		}
		return
	}

	if !ev.Point.In(*sa.VBarBounds()) {
		return
	}
	sa.buttonPressed = true
	switch {
	case ev.Button.Button(1):
		sa.SetVBarOrigPad(ev.Point) // keep pad for drag calc
		sa.CalcPositionFromPoint(ev.Point)
	case ev.Button.Button(4): // wheel up
		sa.ta.PageUp()
	case ev.Button.Button(5): // wheel down
		sa.ta.PageDown()
	}
}
func (sa *ScrollArea) onButtonRelease(ev0 interface{}) {
	if !sa.buttonPressed {
		return
	}
	sa.buttonPressed = false
	ev := ev0.(*xinput.ButtonReleaseEvent)
	if ev.Button.Button(1) {
		sa.CalcPositionFromPoint(ev.Point)
	}
}
func (sa *ScrollArea) onMotionNotify(ev0 interface{}) {
	if !sa.buttonPressed {
		return
	}
	ev := ev0.(*xinput.MotionNotifyEvent)
	switch {
	case ev.Mods.HasButton(1):
		sa.CalcPositionFromPoint(ev.Point)
	}
}