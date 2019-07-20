package i3

import (
	"context"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"testing"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
)

// TestSyncSubprocess runs in a process which has been started with
// DISPLAY= pointing to an Xvfb instance with i3 -c testdata/i3.config running.
func TestSyncSubprocess(t *testing.T) {
	if os.Getenv("GO_WANT_XVFB") != "1" {
		t.Skip("parent process")
	}

	xu, err := xgbutil.NewConn()
	if err != nil {
		t.Fatalf("NewConn: %v", err)
	}
	defer xu.Conn().Close()

	// Create an X11 window
	X := xu.Conn()
	wid, err := xproto.NewWindowId(X)
	if err != nil {
		t.Fatal(err)
	}
	screen := xproto.Setup(X).DefaultScreen(X)
	cookie := xproto.CreateWindowChecked(
		X,
		screen.RootDepth,
		wid,
		screen.Root,
		0, // x
		0, // y
		1, // width
		1, // height
		0, // border width
		xproto.WindowClassInputOutput,
		screen.RootVisual,
		xproto.CwBackPixel|xproto.CwEventMask,
		[]uint32{ // values must be in the order defined by the protocol
			0xffffffff,
			xproto.EventMaskStructureNotify |
				xproto.EventMaskKeyPress |
				xproto.EventMaskKeyRelease})
	if err := cookie.Check(); err != nil {
		t.Fatal(err)
	}

	// Synchronize i3 with that X11 window
	rnd := rand.Uint32()
	resp, err := Sync(SyncRequest{
		Rnd:    rnd,
		Window: uint32(wid),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := resp.Success, true; got != want {
		t.Fatalf("SyncResult.Success: got %v, want %v", got, want)
	}

	for {
		ev, xerr := X.WaitForEvent()
		if xerr != nil {
			t.Fatalf("WaitEvent: got X11 error %v", xerr)
		}
		cm, ok := ev.(xproto.ClientMessageEvent)
		if !ok {
			t.Logf("ignoring non-ClientMessage %v", ev)
			continue
		}
		if got, want := cm.Window, wid; got != want {
			t.Errorf("sync ClientMessage.Window: got %v, want %v", got, want)
		}
		if got, want := cm.Data.Data32[:2], []uint32{uint32(wid), rnd}; !reflect.DeepEqual(got, want) {
			t.Errorf("sync ClientMessage.Data: got %x, want %x", got, want)
		}
		break
	}
}

func TestSync(t *testing.T) {
	t.Parallel()

	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	_, DISPLAY, err := launchXvfb(ctx)
	if err != nil {
		t.Fatal(err)
	}

	cleanup, err := launchI3(ctx, DISPLAY, "")
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cmd := exec.Command(os.Args[0], "-test.run=TestSyncSubprocess", "-test.v")
	cmd.Env = []string{
		"GO_WANT_XVFB=1",
		"DISPLAY=" + DISPLAY,
		"PATH=" + os.Getenv("PATH"),
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatal(err.Error())
	}
}
