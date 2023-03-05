package ui_test

import (
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	"os/exec"
	"strings"
	"testing"
	"time"
)

type Game struct {
	T *testing.T
}

var regularTermination = errors.New("regular termination")

func (g *Game) Update() error {
    if err := changeKeyboardLayout("us"); err != nil {
        g.T.Fatal(err)
    }
	expected := ebiten.KeyName(ebiten.KeyQ)
    if err := changeKeyboardLayout("ru"); err != nil {
        g.T.Fatal(err)
    }
	time.Sleep(100 * time.Millisecond) // Give the OS some time to find the update
	got := ebiten.KeyName(ebiten.KeyQ)
	if strings.Compare(expected, got) != 0 {
		g.T.Fatalf("expected %s, got %s", expected, got)
	}
	return regularTermination
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 1, 1
}

func changeKeyboardLayout(s string, opts ...string) error {
	layouttool := "setxkbmap"
	binary, lookErr := exec.LookPath(layouttool)
	if lookErr != nil {
		return lookErr
	}
	var cmd *exec.Cmd
	switch len(opts) {
	case 0:
		cmd = exec.Command(binary, s)
	case 1:
		cmd = exec.Command(binary, s, opts[0])
	case 2:
		cmd = exec.Command(binary, s, opts[0], opts[1])
	case 3:
		cmd = exec.Command(binary, s, opts[0], opts[1], opts[2])
	case 4:
		cmd = exec.Command(binary, s, opts[0], opts[1], opts[2], opts[3])
	case 5:
		cmd = exec.Command(binary, s, opts[0], opts[1], opts[2], opts[3], opts[4])
	case 6:
		cmd = exec.Command(binary, s, opts[0], opts[1], opts[2], opts[3], opts[4], opts[5])
	default:
		return errors.New("Unsupported amount of args")
	}
	if runErr := cmd.Run(); runErr != nil {
		return runErr
	}
	return nil
}

func getKeyboardLayout() (string, error) {
	layouttool := "setxkbmap"
	binary, lookErr := exec.LookPath(layouttool)
	if lookErr != nil {
		return "", lookErr
	}
	cmd := exec.Command(binary, "-query")
	out, runErr := cmd.Output()
	if runErr != nil {
		return "", runErr
	}
	sout := strings.Fields(string(out[:]))
	for i, s := range sout {
		if strings.Compare(s, "layout:") == 0 {
			return sout[i+1], nil
		}
	}
	return "", errors.New("Unable to find \"layout:\" option")
}

func TestKeyboardFrozenOverTick(t *testing.T) {
	layout, err := getKeyboardLayout()
	if err != nil {
		t.Skip("Skipping test due to missing system dependencies needed to run the test")
	}
	if err := ebiten.RunGame(&Game{t}); err != nil && err != regularTermination {
		t.Fatal(err)
	}
	changeKeyboardLayout(layout)
}
