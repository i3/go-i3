package i3

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

// TestCloseSubprocess runs in a process which has been started with
// DISPLAY= pointing to an Xvfb instance with i3 -c testdata/i3.config running.
func TestCloseSubprocess(t *testing.T) {
	if os.Getenv("GO_WANT_XVFB") != "1" {
		t.Skip("parent process")
	}

	ws := Subscribe(WorkspaceEventType)
	received := make(chan Event)
	go func() {
		defer close(received)
		for ws.Next() {
		}
		received <- nil
	}()
	ws.Close()
	select {
	case <-received:
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for a Close()")
	}
}

func TestClose(t *testing.T) {
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

	cmd := exec.Command(os.Args[0], "-test.run=TestCloseSubprocess", "-test.v")
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
