package glfw_test

import (
	"bytes"
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/examples/resources/images"
	"image/png"
	"testing"
)

type Game struct {
	T *testing.T
}

var regularTermination = errors.New("regular termination")

func (g *Game) Update() error {
	img, err := png.Decode(bytes.NewReader(images.Ebiten_png))
	if err != nil {
		g.T.Fatal(err)
	}
	cursor := ebiten.CreateCursor(&img, 0, 0)
	if cursor == nil {
		g.T.Fatal("Failed to create cursor")
	}
	ebiten.SetCursor(cursor)
	expected := ebiten.CursorShapeCustom
	got := ebiten.CursorShape()
	if got != expected {
		g.T.Fatalf("Expected %d, got %d", expected, got)
	}
	return regularTermination
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1, 1
}

func TestSetCursor(t *testing.T) {
	if err := ebiten.RunGame(&Game{t}); err != nil && err != regularTermination {
		t.Fatal(err)
	}
}
