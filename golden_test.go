package i3

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// TestGoldensSubprocess runs in a process which has been started with
// DISPLAY= pointing to an Xvfb instance with i3 -c testdata/i3.config running.
func TestGoldensSubprocess(t *testing.T) {
	if os.Getenv("GO_WANT_XVFB") != "1" {
		t.Skip("parent process")
	}

	if _, err := RunCommand("open; mark foo"); err != nil {
		t.Fatal(err)
	}

	t.Run("GetVersion", func(t *testing.T) {
		t.Parallel()
		got, err := GetVersion()
		if err != nil {
			t.Fatal(err)
		}
		got.HumanReadable = "" // too brittle to compare
		got.Patch = 0          // the IPC interface does not change across patch releases
		abs, err := filepath.Abs("testdata/i3.config")
		if err != nil {
			t.Fatal(err)
		}
		want := Version{
			Major:                4,
			Minor:                17,
			Patch:                0,
			LoadedConfigFileName: abs,
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetVersion reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("AtLeast", func(t *testing.T) {
		t.Parallel()
		if err := AtLeast(4, 14); err != nil {
			t.Errorf("AtLeast(4, 14) unexpectedly returned an error: %v", err)
		}
		if err := AtLeast(4, 0); err != nil {
			t.Errorf("AtLeast(4, 0) unexpectedly returned an error: %v", err)
		}
		if err := AtLeast(4, 999); err == nil {
			t.Errorf("AtLeast(4, 999) unexpectedly did not return an error")
		}
	})

	t.Run("GetBarIDs", func(t *testing.T) {
		t.Parallel()
		got, err := GetBarIDs()
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"bar-0"}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetBarIDs reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetBarConfig", func(t *testing.T) {
		t.Parallel()
		got, err := GetBarConfig("bar-0")
		if err != nil {
			t.Fatal(err)
		}
		want := BarConfig{
			ID:                   "bar-0",
			Mode:                 "dock",
			Position:             "bottom",
			StatusCommand:        "i3status",
			Font:                 "-misc-fixed-medium-r-normal--13-120-75-75-C-70-iso10646-1",
			WorkspaceButtons:     true,
			BindingModeIndicator: true,
			Colors: BarConfigColors{
				Background:                  "#000000",
				Statusline:                  "#ffffff",
				Separator:                   "#666666",
				FocusedBackground:           "#000000",
				FocusedStatusline:           "#ffffff",
				FocusedSeparator:            "#666666",
				FocusedWorkspaceText:        "#4c7899",
				FocusedWorkspaceBackground:  "#285577",
				FocusedWorkspaceBorder:      "#ffffff",
				ActiveWorkspaceText:         "#333333",
				ActiveWorkspaceBackground:   "#5f676a",
				ActiveWorkspaceBorder:       "#ffffff",
				InactiveWorkspaceText:       "#333333",
				InactiveWorkspaceBackground: "#222222",
				InactiveWorkspaceBorder:     "#888888",
				UrgentWorkspaceText:         "#2f343a",
				UrgentWorkspaceBackground:   "#900000",
				UrgentWorkspaceBorder:       "#ffffff",
				BindingModeText:             "#2f343a",
				BindingModeBackground:       "#900000",
				BindingModeBorder:           "#ffffff",
			},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetBarConfig reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetBindingModes", func(t *testing.T) {
		t.Parallel()
		got, err := GetBindingModes()
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"default"}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetBindingModes reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetMarks", func(t *testing.T) {
		t.Parallel()
		got, err := GetMarks()
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"foo"}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetMarks reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetOutputs", func(t *testing.T) {
		t.Parallel()
		got, err := GetOutputs()
		if err != nil {
			t.Fatal(err)
		}
		want := []Output{
			{
				Name: "xroot-0",
				Rect: Rect{Width: 1280, Height: 800},
			},
			{
				Name:             "screen",
				Active:           true,
				CurrentWorkspace: "1",
				Rect:             Rect{Width: 1280, Height: 800},
			},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetOutputs reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetWorkspaces", func(t *testing.T) {
		t.Parallel()
		got, err := GetWorkspaces()
		if err != nil {
			t.Fatal(err)
		}
		want := []Workspace{
			{
				Num:     1,
				Name:    "1",
				Visible: true,
				Focused: true,
				Rect:    Rect{Width: 1280, Height: 800},
				Output:  "screen",
			},
		}
		cmpopts := []cmp.Option{
			cmp.FilterPath(
				func(p cmp.Path) bool {
					return p.Last().String() == ".ID"
				},
				cmp.Ignore()),
		}
		if diff := cmp.Diff(got, want, cmpopts...); diff != "" {
			t.Fatalf("unexpected GetWorkspaces reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("RunCommand", func(t *testing.T) {
		t.Parallel()
		got, err := RunCommand("norp")
		if err != nil && !IsUnsuccessful(err) {
			t.Fatal(err)
		}
		if !IsUnsuccessful(err) {
			t.Fatalf("command unexpectedly succeeded")
		}
		want := []CommandResult{
			{
				Success: false,
				Error:   "Expected one of these tokens: <end>, '[', 'move', 'exec', 'exit', 'restart', 'reload', 'shmlog', 'debuglog', 'border', 'layout', 'append_layout', 'workspace', 'focus', 'kill', 'open', 'fullscreen', 'sticky', 'split', 'floating', 'mark', 'unmark', 'resize', 'rename', 'nop', 'scratchpad', 'swap', 'title_format', 'mode', 'bar'",
			},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected RunCommand reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetConfig", func(t *testing.T) {
		t.Parallel()
		got, err := GetConfig()
		if err != nil {
			t.Fatal(err)
		}
		configBytes, err := ioutil.ReadFile("testdata/i3.config")
		if err != nil {
			t.Fatal(err)
		}
		want := Config{
			Config: string(configBytes),
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("unexpected GetConfig reply: (-got +want)\n%s", diff)
		}
	})

	t.Run("GetTree", func(t *testing.T) {
		t.Parallel()
		got, err := GetTree()
		if err != nil {
			t.Fatal(err)
		}

		// Basic sanity checks:
		if got.Root == nil {
			t.Fatalf("tree.Root unexpectedly is nil")
		}

		if got, want := got.Root.Name, "root"; got != want {
			t.Fatalf("unexpected tree root name: got %q, want %q", got, want)
		}

		// Exercise FindFocused to locate at least one workspace.
		if node := got.Root.FindFocused(func(n *Node) bool { return n.Type == WorkspaceNode }); node == nil {
			t.Fatalf("unexpectedly could not find any workspace node in GetTree reply")
		}

		// Exercise FindChild to locate at least one workspace.
		if node := got.Root.FindChild(func(n *Node) bool { return n.Type == WorkspaceNode }); node == nil {
			t.Fatalf("unexpectedly could not find any workspace node in GetTree reply")
		}
	})
}

func TestGoldens(t *testing.T) {
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

	cmd := exec.Command(os.Args[0], "-test.run=TestGoldensSubprocess", "-test.v")
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
