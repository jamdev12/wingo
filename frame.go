package main

import "code.google.com/p/jamslam-x-go-binding/xgb"

import (
    "github.com/BurntSushi/xgbutil/xgraphics"
    "github.com/BurntSushi/xgbutil/xrect"
)

const (
    StateActive = iota
    StateInactive
)

type Frame interface {
    Client() *client
    ConfigureClient(flags uint16, x, y int16, w, h uint16,
                    sibling xgb.Id, stackMode byte, ignoreHints bool)
    ConfigureFrame(flags uint16, x, y int16, w, h uint16,
                   sibling xgb.Id, stackMode byte, ignoreHints bool)
    Destroy()
    Geom() xrect.Rect // the geometry of the parent window
    Map()
    Off()
    On()
    Parent() *frameParent
    ParentId() xgb.Id
    ParentWin() *window
    State() int
    StateActive()
    StateInactive()
    Unmap()
    ValidateHeight(height uint16) uint16
    ValidateWidth(width uint16) uint16

    // The margins of this frame's decorations.
    Top() int16
    Bottom() int16
    Left() int16
    Right() int16

    // These are temporary. I think they will move to 'layout'
    Moving() bool
    MovingState() *moveState
    // moveBegin(rx, ry, ex, ey int16) 
    // moveStep(rx, ry, ex, ey int16) 
    // moveEnd(rx, ry, ex, ey int16) 

    Resizing() bool
    ResizingState() *resizeState
    // resizeBegin(direction uint32, rx, ry, ex, ey int16) (bool, xgb.Id) 
    // resizeStep(rx, ry, ex, ey int16) 
    // resizeEnd(rx, ry, ex, ey int16) 
}

type frameParent struct {
    window *window
    client *client
}

func newParent(c *client) *frameParent {
    mask := uint32(xgb.CWEventMask)
    val := []uint32{xgb.EventMaskSubstructureRedirect |
                    xgb.EventMaskButtonPress |
                    xgb.EventMaskButtonRelease}
    parent := createWindow(X.RootWin(), mask, val)
    p := &frameParent{
        window: parent,
        client: c,
    }

    X.Conn().ReparentWindow(c.Id(), parent.id, 0, 0)

    return p
}

func (p *frameParent) Win() *window {
    return p.window
}

// framePiece is similar to clientOffset in concept, but it represents
// any other piece of the decorations. As such, it requires a bit more state.
// The corresponding "new" functions for this type can be found in *pieces.go
type framePiece struct {
    win *window
    imgActive xgb.Id
    imgInactive xgb.Id
    xoff, yoff int16
    woff, hoff uint16
}

func (p *framePiece) destroy() {
    p.win.destroy() // also detaches all event handlers
    xgraphics.FreePixmap(X, p.imgActive)
    xgraphics.FreePixmap(X, p.imgInactive)
}

func (p *framePiece) active() {
    p.win.change(xgb.CWBackPixmap, uint32(p.imgActive))
    p.win.clear()
}

func (p *framePiece) inactive() {
    p.win.change(xgb.CWBackPixmap, uint32(p.imgInactive))
    p.win.clear()
}

func (p *framePiece) initialGeom(flags uint16) {
    p.win.moveresize(flags, p.xoff, p.yoff, p.woff, p.hoff)
}

func (p *framePiece) x() int16 {
    return p.win.geom.X()
}

func (p *framePiece) y() int16 {
    return p.win.geom.Y()
}

func (p *framePiece) w() uint16 {
    return p.win.geom.Width()
}

func (p *framePiece) h() uint16 {
    return p.win.geom.Height()
}

// The relative geometry of the client window in the frame parent window.
// x and y are relative to the top-left corner of the parent window.
// w and h are values that satisfy these properties:
// parent_width - w = client_width
// parent_height - h = client_height
// Where client_width and client_height is the width and height of the client
// window inside the frame.
type clientOffset struct {
    x, y int16
    w, h uint16
}

type moveState struct {
    moving bool
    lastRootX int16
    lastRootY int16
}

type resizeState struct {
    resizing bool
    rootX, rootY int16
    x, y int16
    width, height uint16
    xs, ys, ws, hs bool
}

// Frame related functions that can be defined using only the Frame interface.

func FrameReset(f Frame) {
    geom := f.Client().Geom()
    FrameMR(f, DoW | DoH, 0, 0, geom.Width(), geom.Height(), false)
}

// FrameMR is short for FrameMoveresize.
func FrameMR(f Frame, flags uint16, x, y int16, w, h uint16, ignoreHints bool) {
    f.ConfigureClient(flags, x, y, w, h, xgb.Id(0), 0, ignoreHints)
}

