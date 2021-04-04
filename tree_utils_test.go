package i3

import (
	"context"
	"os"
	"os/exec"
	"testing"
)

// TestTreeUtilsSubprocess runs in a process which has been started with
// DISPLAY= pointing to an Xvfb instance with i3 -c testdata/i3.config running.
func TestTreeUtilsSubprocess(t *testing.T) {
	if os.Getenv("GO_WANT_XVFB") != "1" {
		t.Skip("parent process")
	}

	mark_name := "foo"
	ws_name := "1:test_space"

	if _, err := RunCommand("rename workspace to " + ws_name); err != nil {
		t.Fatal(err)
	}

	if _, err := RunCommand("open; mark " + mark_name); err != nil {
		t.Fatal(err)
	}

	t.Run("FindParent", func(t *testing.T) {
		t.Parallel()
		got, err := GetTree()
		if err != nil {
			t.Fatal(err)
		}

		node := got.Root.FindFocused(func(n *Node) bool { return n.Focused })
		if node == nil {
			t.Fatal("unexpectedly could not find any focused node in GetTree reply")
		}

		// Exercise FindParent to locate parent for given node.
		parent := node.FindParent()

		if parent == nil {
			t.Fatal("no parent found")
		}
		if parent.Name != ws_name {
			t.Fatal("wrong parent found: " + parent.Name)
		}
	})

	t.Run("IsFloating", func(t *testing.T) {
		// do not run in parallel because 'floating toggle' breaks other tests

		got, err := GetTree()
		if err != nil {
			t.Fatal(err)
		}

		node := got.Root.FindFocused(func(n *Node) bool { return n.Focused })
		if node == nil {
			t.Fatal("unexpectedly could not find any focused node in GetTree reply")
		}

		if node.IsFloating() == true {
			t.Fatal("node is floating")
		}

		if _, err := RunCommand("floating toggle"); err != nil {
			t.Fatal(err)
		}

		got, err = GetTree()
		if err != nil {
			t.Fatal(err)
		}

		node = got.Root.FindFocused(func(n *Node) bool { return n.Focused })
		if node == nil {
			t.Fatal("unexpectedly could not find any focused node in GetTree reply")
		}

		if node.IsFloating() == false {
			t.Fatal("node is not floating")
		}

		RunCommand("floating toggle")
	})
}

func TestTreeUtils(t *testing.T) {
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

	cmd := exec.Command(os.Args[0], "-test.run=TestTreeUtilsSubprocess", "-test.v")
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
