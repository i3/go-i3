package i3

import "strings"

func (n *Node) FindFocusedLeaf() *Node {
	return n.FindFocused(func(n *Node) bool { return n.Focused })
}

func (child *Node) FindParent() *Node {
	tree, err := GetTree()
	if err != nil {
		return nil
	}
	parent := tree.Root.FindChild(func(n *Node) bool {
		for _, f := range n.Focus {
			if f == child.ID {
				return true
			}
		}
		return false
	})

	return parent
}

func (n *Node) IsFloating() bool {
	if len(n.Floating) < 4 {
		return false
	}
	return strings.HasSuffix(string(n.Floating), "_on")
}
