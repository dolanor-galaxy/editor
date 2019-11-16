package wimage

import (
	"image"
	"image/draw"

	"github.com/BurntSushi/xgb/shm"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/pkg/errors"
)

type ShmWImage struct {
	opt     *Options
	segId   shm.Seg
	imgWrap *ShmImgWrap
}

func NewShmWImage(opt *Options) (*ShmWImage, error) {
	// init shared memory extension
	if err := shm.Init(opt.Conn); err != nil {
		return nil, err
	}

	wi := &ShmWImage{opt: opt}

	// server segment id
	segId, err := shm.NewSegId(wi.opt.Conn)
	if err != nil {
		return nil, err
	}
	wi.segId = segId

	// initial image
	r := image.Rect(0, 0, 1, 1)
	if err := wi.Resize(r); err != nil {
		return nil, err
	}

	return wi, nil
}
func (wi *ShmWImage) Close() error {
	return wi.imgWrap.Close()
}

func (wi *ShmWImage) Resize(r image.Rectangle) error {
	imgWrap, err := NewShmImgWrap(r)
	if err != nil {
		return err
	}
	old := wi.imgWrap
	wi.imgWrap = imgWrap
	// clean old img
	if old != nil {
		// need to detach to attach a new img id later
		_ = shm.Detach(wi.opt.Conn, wi.segId)

		err := old.Close()
		if err != nil {
			return err
		}
	}
	// attach to segId
	readOnly := false
	shmId := uint32(wi.imgWrap.shmId)
	cookie := shm.AttachChecked(wi.opt.Conn, wi.segId, shmId, readOnly)
	if err := cookie.Check(); err != nil {
		return errors.Wrap(err, "shmwimage.resize.attach")
	}

	return nil
}

func (wi *ShmWImage) Image() draw.Image {
	return wi.imgWrap.Img
}

func (wi *ShmWImage) PutImage(r image.Rectangle) (bool, error) {
	gctx := wi.opt.GCtx
	img := wi.imgWrap.Img
	drawable := xproto.Drawable(wi.opt.Window)
	depth := wi.opt.ScreenInfo.RootDepth
	b := img.Bounds()
	_ = shm.PutImage(
		wi.opt.Conn,
		drawable,
		gctx,
		uint16(b.Dx()), uint16(b.Dy()), // total width/height
		uint16(r.Min.X), uint16(r.Min.Y), uint16(r.Dx()), uint16(r.Dy()), // src x,y,w,h
		int16(r.Min.X), int16(r.Min.Y), // dst x,y
		depth,
		xproto.ImageFormatZPixmap,
		1, // send shm.CompletionEvent when done
		wi.segId,
		0) // offset
	return false, nil
}