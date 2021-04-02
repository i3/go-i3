package i3

import "strings"

func (n *Node) FindFocusedLeaf() *Node {
	return n.FindFocused(func(n *Node) bool { return n.Focused })
}

// FindParent method returns the parent node of the curren one
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

// IsFloating method returns true if the current node is floating
func (n *Node) IsFloating() bool {
	return strings.HasSuffix(string(n.Floating), "_on")
}
