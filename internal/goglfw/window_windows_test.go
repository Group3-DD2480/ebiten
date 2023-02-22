package goglfw_test

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2/internal/goglfw"
)

func TestWindow_SetSizeLimits(t *testing.T) {

	// Test case 1:
	window := goglfw.Window{}
	minwidth := 2
	minheight := 2
	maxwidth := 3
	maxheight := 3
	err := window.SetSizeLimits(minwidth, minheight, maxwidth, maxheight)
	if err != goglfw.NotInitialized {
		t.Error()
	}

	goglfw.Init()
	// Test case 2:
	window = goglfw.Window{}
	minwidth = 0
	minheight = -2
	maxwidth = 1
	maxheight = 1
	err = window.SetSizeLimits(minwidth, minheight, maxwidth, maxheight)
	if fmt.Sprint(err) != fmt.Sprint("goglfw: invalid window minimum size 0x-2: ", goglfw.InvalidValue) {
		t.Error()
	}

	// Test case 3:
	window = goglfw.Window{}
	minwidth = 1
	minheight = 1
	maxwidth = -2
	maxheight = -2
	err = window.SetSizeLimits(minwidth, minheight, maxwidth, maxheight)
	if fmt.Sprint(err) != fmt.Sprint("goglfw: invalid window maximum size -2x-2: ", goglfw.InvalidValue) {
		t.Error()
	}

	// Test case 4: resizable set to false
	window = goglfw.Window{}
	//(window).resizable = false
	minwidth = -1
	minheight = -1
	maxwidth = -1
	maxheight = -1
	err = window.SetSizeLimits(minwidth, minheight, maxwidth, maxheight)
	if err != nil {
		t.Error()
	}

	// Test case 5:
	window = goglfw.Window{}
	minwidth = 2
	minheight = 2
	maxwidth = 3
	maxheight = 3
	err = window.SetSizeLimits(minwidth, minheight, maxwidth, maxheight)
	if err != nil {
		t.Error()
	}

}
