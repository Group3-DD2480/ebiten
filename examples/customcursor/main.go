// Copyright 2021 The Ebiten Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"image"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/internal/glfw"
	"github.com/hajimehoshi/ebiten/v2/vector"
	resources "github.com/hajimehoshi/ebiten/v2/examples/resources/images/flappy"
)

const (
	screenWidth  = 640
	screenHeight = 480
)

type Game struct {
	grids      map[image.Rectangle]*glfw.Cursor
	gridColors map[image.Rectangle]color.Color
}

type Sprite struct {
	image *ebiten.Image
	x     int
	y     int
}

func (g *Game) Update() error {
	pt := image.Pt(ebiten.CursorPosition())
	for r, cursor := range g.grids {
		if pt.In(r) {
			ebiten.SetCursor(cursor)
			currentCursor = cursor
			break
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		switch ebiten.CursorMode() {
		case ebiten.CursorModeVisible:
			ebiten.SetCursorMode(ebiten.CursorModeHidden)
		case ebiten.CursorModeHidden:
			ebiten.SetCursorMode(ebiten.CursorModeVisible)
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for r, c := range g.gridColors {
		vector.DrawFilledRect(screen, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), c)
	}

	if currentCursor == ebitenCursor {
		ebitenutil.DebugPrint(screen, "Cursor: Ebiten")
	} else if currentCursor == gopherCursor {
		ebitenutil.DebugPrint(screen, "Cursor: Gopher")
	}

}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func createCustomCursor(sourceImage []byte) *glfw.Cursor {
	// Decode an image from the image file's byte slice.
	img, _, err := image.Decode(bytes.NewReader(sourceImage))
	if err != nil {
		log.Fatal(err)
	}
	return ebiten.CreateCursor(&img, 0, 0)
}

var ebitenCursor *glfw.Cursor
var gopherCursor *glfw.Cursor
var currentCursor *glfw.Cursor

func init() {
	ebitenCursor = createCustomCursor(images.Ebiten_png)
	gopherCursor = createCustomCursor(resources.Gopher_png)
}

func main() {

	//ebiten.SetCursor(ebitenCursor)

	marginedWidth := 50

	g := &Game{
		grids: map[image.Rectangle]*glfw.Cursor{
			image.Rect(marginedWidth, marginedWidth, screenWidth/2, screenHeight-marginedWidth): ebitenCursor,
			image.Rect(screenWidth/2, marginedWidth, screenWidth-marginedWidth, screenHeight-marginedWidth): gopherCursor,
			// image.Rect(300, 100, 400, 200): ebitenCursor,
			// image.Rect(100, 200, 200, 300): gopherCursor,
			// image.Rect(200, 200, 300, 300): ebitenCursor,
			// image.Rect(300, 200, 400, 300): gopherCursor,
		},
		gridColors: map[image.Rectangle]color.Color{},
	}
	for rect, c := range g.grids {
		clr := color.RGBA{0x40, 0x40, 0x40, 0xff}
		if c == ebitenCursor {
			clr.R = 0x80
		} else {
			clr.G = 0x80
		}
		g.gridColors[rect] = clr
	}

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("CustomCursor (Ebitengine Demo)")
	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
