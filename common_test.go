package i3

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

func displayLikelyAvailable(display int) bool {
	// The path to this lock is hard-coded to /tmp in the Xorg source code, at
	// least in xorg-server-1.19.3. If the path ever changes, that’s no big
	// deal. We’ll fall through to starting Xvfb and having Xvfb fail, which is
	// only a performance hit, no failure.
	b, err := ioutil.ReadFile(fmt.Sprintf("/tmp/.X%d-lock", display))
	if err != nil {
		if os.IsNotExist(err) {
			return true
		}
		// Maybe a starting process is just replacing the file? The display
		// is likely not available.
		return false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil {
		// No pid inside the lock file, so Xvfb will remove the file.
		return true
	}

	return !pidValid(pid)
}

var signalMu sync.Mutex

func launchXvfb(ctx context.Context) (xvfb *exec.Cmd, DISPLAY string, _ error) {
	// Only one goroutine can wait for Xvfb to start at any point in time, as
	// signal handlers are global (per-process, not per-goroutine).
	signalMu.Lock()
	defer signalMu.Unlock()

	var lastErr error
	display := 0 // :0 is usually an active session
	for attempt := 0; attempt < 100; attempt++ {
		display++
		if !displayLikelyAvailable(display) {
			continue
		}
		// display likely available, try to start Xvfb
		DISPLAY := fmt.Sprintf(":%d", display)
		// Indicate we implement Xvfb’s readiness notification mechanism.
		signal.Ignore(syscall.SIGUSR1)
		xvfb := exec.CommandContext(ctx, "Xvfb", DISPLAY, "-screen", "0", "1280x800x24")
		if attempt == 99 { // last attempt
			xvfb.Stderr = os.Stderr
		}
		if lastErr = xvfb.Start(); lastErr != nil {
			continue
		}

		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGUSR1)
		// The buffer of 1 allows the Wait() goroutine to return.
		status := make(chan error, 1)
		go func() {
			defer signal.Stop(ch)
			for sig := range ch {
				if sig == syscall.SIGUSR1 {
					status <- nil // success
					return
				}
			}
		}()
		go func() {
			defer func() {
				signal.Stop(ch)
				close(ch) // avoid leaking the other goroutine
			}()
			ps, err := xvfb.Process.Wait()
			if err != nil {
				status <- err
				return
			}
			if ps.Exited() {
				status <- fmt.Errorf("Xvfb exited: %v", ps)
				return
			}
			status <- fmt.Errorf("BUG: Wait returned, but !ps.Exited()")
		}()
		if lastErr = <-status; lastErr == nil {
			return xvfb, DISPLAY, nil // Xvfb ready
		}
	}
	return nil, "", lastErr
}
