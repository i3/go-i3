package i3

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

// TestSubscribeSubprocess runs in a process which has been started with
// DISPLAY= pointing to an Xvfb instance with i3 -c testdata/i3.config running.
func TestSubscribeSubprocess(t *testing.T) {
	if os.Getenv("GO_WANT_XVFB") != "1" {
		t.Skip("parent process")
	}

	// TODO(https://github.com/i3/i3/issues/2988): as soon as we are targeting
	// i3 4.15, use SendTick to eliminate race conditions in this test.

	t.Run("subscribe", func(t *testing.T) {
		var eg errgroup.Group
		ws := Subscribe(WorkspaceEventType)
		received := make(chan Event)
		eg.Go(func() error {
			defer close(received)
			if ws.Next() {
				received <- ws.Event()
			}
			return ws.Close()
		})
		// As we can’t know when EventReceiver.Next() actually subscribes, we
		// just continuously switch workspaces.
		ctx, canc := context.WithCancel(context.Background())
		defer canc()
		go func() {
			cnt := 2
			for ctx.Err() == nil {
				RunCommand(fmt.Sprintf("workspace %d", cnt))
				cnt++
				time.Sleep(10 * time.Millisecond)
			}
		}()
		select {
		case <-received:
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for an event from i3")
		}
		if err := eg.Wait(); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("subscribeParallel", func(t *testing.T) {
		var mu sync.Mutex
		received := make(map[string]int)

		recv1 := Subscribe(WorkspaceEventType)
		go func() {
			for recv1.Next() {
				ev := recv1.Event().(*WorkspaceEvent)
				if ev.Change == "init" {
					mu.Lock()
					received[ev.Current.Name]++
					mu.Unlock()
				}
			}
		}()

		recv2 := Subscribe(WorkspaceEventType)
		go func() {
			for recv2.Next() {
				ev := recv2.Event().(*WorkspaceEvent)
				if ev.Change == "init" {
					mu.Lock()
					received[ev.Current.Name]++
					mu.Unlock()
				}
			}
		}()

		cnt := 2
		start := time.Now()
		for time.Since(start) < 5*time.Second {
			mu.Lock()
			done := received[fmt.Sprintf("%d", cnt-1)] == 2
			mu.Unlock()
			if done {
				return // success
			}
			if _, err := RunCommand(fmt.Sprintf("workspace %d", cnt)); err != nil {
				t.Fatal(err)
			}
			cnt++
			time.Sleep(10 * time.Millisecond)
		}
	})

	t.Run("subscribeMultiple", func(t *testing.T) {
		var eg errgroup.Group
		ws := Subscribe(WorkspaceEventType, ModeEventType)
		received := make(chan struct{})
		eg.Go(func() error {
			defer close(received)
			defer ws.Close()
			seen := map[EventType]bool{
				WorkspaceEventType: false,
				ModeEventType:      false,
			}
		Outer:
			for ws.Next() {
				switch ws.Event().(type) {
				case *WorkspaceEvent:
					seen[WorkspaceEventType] = true
				case *ModeEvent:
					seen[ModeEventType] = true
				}

				for _, seen := range seen {
					if !seen {
						continue Outer
					}
				}

				return nil
			}
			return ws.Close()
		})
		// As we can’t know when EventReceiver.Next() actually subscribes, we
		// just continuously switch workspaces and modes.
		ctx, canc := context.WithCancel(context.Background())
		defer canc()
		go func() {
			modes := []string{"default", "conf"}
			cnt := 2
			for ctx.Err() == nil {
				RunCommand(fmt.Sprintf("workspace %d", cnt))
				cnt++
				RunCommand(fmt.Sprintf("mode %s", modes[cnt%len(modes)]))
				time.Sleep(10 * time.Millisecond)
			}
		}()
		select {
		case <-received:
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout waiting for an event from i3")
		}
		if err := eg.Wait(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSubscribe(t *testing.T) {
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

	cmd := exec.Command(os.Args[0], "-test.run=TestSubscribeSubprocess", "-test.v")
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
