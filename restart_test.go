package i3

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestRestartSubprocess runs in a process which has been started with
// DISPLAY= pointing to an Xvfb instance with i3 -c testdata/i3.config running.
func TestRestartSubprocess(t *testing.T) {
	if os.Getenv("GO_WANT_XVFB") != "1" {
		t.Skip("parent process")
	}

	// received is buffered so that we can blockingly read on tick.
	received := make(chan *ShutdownEvent, 1)
	tick := make(chan *TickEvent)
	go func() {
		defer close(tick)
		defer close(received)
		recv := Subscribe(ShutdownEventType, TickEventType)
		defer recv.Close()
		log.Printf("reading events")
		for recv.Next() {
			log.Printf("received: %#v", recv.Event())
			switch ev := recv.Event().(type) {
			case *ShutdownEvent:
				received <- ev
			case *TickEvent:
				tick <- ev
			}
		}
		log.Printf("done reading events")
		t.Fatal(recv.Close()) // should not be reached
	}()

	log.Printf("read initial tick")
	<-tick // Wait until the subscription is ready
	log.Printf("restart")
	if err := Restart(); err != nil {
		t.Fatal(err)
	}

	// Restarting i3 triggered a close of the connection, i.e. also a new
	// subscribe and initial tick event:
	log.Printf("read next initial tick")
	ev := <-tick
	if !ev.First {
		t.Fatalf("expected first tick after restart, got %#v instead", ev)
	}

	if _, err := SendTick(""); err != nil {
		t.Fatal(err)
	}
	log.Printf("read tick")
	<-tick // Wait until tick was received
	log.Printf("read received")
	<-received // Verify shutdown event was received
	log.Printf("getversion")
	if _, err := GetVersion(); err != nil {
		t.Fatal(err)
	}
}

func TestRestart(t *testing.T) {
	t.Parallel()

	ctx, canc := context.WithCancel(context.Background())
	defer canc()

	_, DISPLAY, err := launchXvfb(ctx)
	if err != nil {
		t.Fatal(err)
	}

	dir, err := ioutil.TempDir("", "i3restart")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	I3SOCK := filepath.Join(dir, "i3.sock")

	cleanup, err := launchI3(ctx, DISPLAY, I3SOCK)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	cmd := exec.Command(os.Args[0], "-test.run=TestRestartSubprocess", "-test.v")
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
