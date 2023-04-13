package tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-auxiliaries/selenium/tests/testconf"

	"github.com/go-auxiliaries/selenium"

	"github.com/BurntSushi/xgbutil"
	"github.com/google/go-cmp/cmp"
)

func TestFrameBuffer(t *testing.T) {
	// Note on FrameBuffer and xgb.Conn:
	// There appears to be a race condition when closing a Conn instance before
	// a FrameBuffer instance.  A short sleep solves the problem.
	t.Run("Default behavior", func(t *testing.T) {
		// The default Xvfb screen size is "1280x1024x8".
		if !testconf.Xvfb() {
			t.Skipf("skipping - xvfb is disabled")
		}
		frameBuffer, err := selenium.NewFrameBuffer()
		if err != nil {
			t.Fatalf("Could not create frame buffer: %s", err.Error())
		}
		defer func() {
			_ = frameBuffer.Stop()
		}()

		if frameBuffer.Display == "" {
			t.Fatalf("frameBuffer.Display is empty")
		}

		d, err := xgbutil.NewConnDisplay(":" + frameBuffer.Display)
		if err != nil {
			t.Fatalf("could not connect to display %q: %s", frameBuffer.Display, err.Error())
		}
		defer time.Sleep(time.Second * 2)
		defer d.Conn().Close()
		s := d.Screen()
		if diff := cmp.Diff(1280, int(s.WidthInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
		if diff := cmp.Diff(1024, int(s.HeightInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
	})
	t.Run("With bad screen size", func(t *testing.T) {
		if !testconf.Xvfb() {
			t.Skipf("skipping - xvfb is disabled")
		}
		options := selenium.FrameBufferOptions{
			ScreenSize: "not a screen size",
		}
		_, err := selenium.NewFrameBufferWithOptions(options)
		if err == nil {
			t.Fatalf("Expected an error about the screen size")
		}
	})
	t.Run("With screen size", func(t *testing.T) {
		if !testconf.Xvfb() {
			t.Skipf("skipping - xvfb is disabled")
		}
		desiredWidth := 1024
		desiredHeight := 768
		desiredDepth := 24
		options := selenium.FrameBufferOptions{
			ScreenSize: fmt.Sprintf("%dx%dx%d", desiredWidth, desiredHeight, desiredDepth),
		}
		frameBuffer, err := selenium.NewFrameBufferWithOptions(options)
		if err != nil {
			t.Fatalf("Could not create frame buffer: %s", err.Error())
		}
		defer func() {
			_ = frameBuffer.Stop()
		}()

		if frameBuffer.Display == "" {
			t.Fatalf("frameBuffer.Display is empty")
		}

		d, err := xgbutil.NewConnDisplay(":" + frameBuffer.Display)
		if err != nil {
			t.Fatalf("could not connect to display %q: %s", frameBuffer.Display, err.Error())
		}
		defer time.Sleep(time.Second * 2)
		defer d.Conn().Close()
		s := d.Screen()
		if diff := cmp.Diff(desiredWidth, int(s.WidthInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
		if diff := cmp.Diff(desiredHeight, int(s.HeightInPixels)); diff != "" {
			t.Fatalf("args returned diff (-want/+got):\n%s", diff)
		}
	})
}
