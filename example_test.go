package i3_test

import (
	"fmt"
	"log"
	"strings"

	"go.i3wm.org/i3/v4"
)

func ExampleIsUnsuccessful() {
	cr, err := i3.RunCommand("norp")
	// “norp” is not implemented, so this command is expected to fail.
	if err != nil && !i3.IsUnsuccessful(err) {
		log.Fatal(err)
	}
	log.Printf("error for norp: %v", cr[0].Error)
}

func ExampleSubscribe() {
	recv := i3.Subscribe(i3.WindowEventType)
	for recv.Next() {
		ev := recv.Event().(*i3.WindowEvent)
		log.Printf("change: %s", ev.Change)
	}
	log.Fatal(recv.Close())
}

func ExampleGetTree() {
	// Focus or start Google Chrome on the focused workspace.

	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}

	ws := tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Type == i3.WorkspaceNode
	})
	if ws == nil {
		log.Fatalf("could not locate workspace")
	}

	chrome := ws.FindChild(func(n *i3.Node) bool {
		return strings.HasSuffix(n.Name, "- Google Chrome")
	})
	if chrome != nil {
		_, err = i3.RunCommand(fmt.Sprintf(`[con_id="%d"] focus`, chrome.ID))
	} else {
		_, err = i3.RunCommand(`exec google-chrome`)
	}
	if err != nil {
		log.Fatal(err)
	}
}
