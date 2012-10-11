// Copyright 2012 by Graeme Humphries <graeme@sudo.ca>
//
// kdtree is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// kdtree is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with kdtree.  If not, see http://www.gnu.org/licenses/.

package kdtree

import (
	"errors"
	"sort"
	"strconv"
	"sync"
)

/***** Tree Object *****/
// Tree is needed for locking, to prevent syncronization issues.
type Tree struct {
	Mutex sync.RWMutex

	Root *Node
}

/***** Tree Functions *****/
// These functions wrap the private Node functions in lock operations so that
// they're thread-safe.

// Adds new Node to existing Tree. Returns an error if the Node can't be added to the tree.
func (t *Tree) Add(newnode *Node) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if t.Root == nil {
		t.Root = newnode
		return nil
	}
	return t.Root.add(newnode)
}

// Removes node from the Tree, rebalancoing other nodes as necessary.
// Returns an error if Node isn't a member of this Tree.
func (t *Tree) Remove(n *Node) error {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	if n.tree != t {
		return errors.New("Node not a member of tree.")
	}
	if newroot := n.remove(); newroot != nil {
		t.Root = newroot
	}

	return nil
}

// Performs a left depth first tree traversal, running function f on every Node found.
func (t *Tree) Traverse(f func(*Node)) {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	f(t.Root)
}

/***** Tree Management Functions *****/

// Builds a new tree from a list of nodes. This is destructive, and
// will remove any existing tree membership from nodes passed to it.
func BuildTree(nodes []*Node) *Tree {
	tree := new(Tree)
	tree.Mutex.Lock()
	defer tree.Mutex.Unlock()
	tree.Root = buildRootNode(nodes, 0, nil)
	f := func(n *Node) {
		n.tree = tree
	}
	tree.Root.traverse(f)

	return tree
}

// Builds a tree from a list of nodes. Returns the root Node of the new tree.
// This is destructive, and will break any existing tree these nodes may be a member of.
// This is intended to be used to build an new tree, or as part of a tree Balance.
// This is a recursive function, you should always call it with depth = 0, parent = nil.
func buildRootNode(nodes []*Node, depth int, parent *Node) *Node {
	var root *Node
	// special case handling first
	switch len(nodes) {
	case 0:
		root = nil
	case 1:
		dimensions := len(nodes[0].Coordinates)
		root = nodes[0]

		root.axis = depth % dimensions
		root.parent = parent
		root.leftChild = nil
		root.rightChild = nil
	default:
		median := (len(nodes) / 2) - 1 // -1 so that it's a slice index
		dimensions := len(nodes[0].Coordinates)

		snl := new(sortableNodeList)
		snl.Axis = depth % dimensions
		snl.Nodes = make([]*Node, len(nodes))
		copy(snl.Nodes, nodes)
		sort.Sort(snl)

		root = snl.Nodes[median]

		root.axis = snl.Axis
		root.parent = parent
		root.leftChild = buildRootNode(snl.Nodes[0:median], depth+1, root)
		root.rightChild = buildRootNode(snl.Nodes[median+1:], depth+1, root)
	}

	return root
}

// Rebalances a whole Tree.
func (t *Tree) Balance() {
	t.Mutex.Lock()
	defer t.Mutex.Unlock()
	nodelist := t.Root.nodeList()
	t.Root = buildRootNode(nodelist, 0, nil)
}

// Checks that Tree is a valid kdtree. Returns an error of there are problems.
func (t *Tree) Validate() error {
	t.Mutex.RLock()
	t.Mutex.RUnlock()
	return t.Root.validate()
}

// Checks that the (sub)tree below this node is valid:
// - All children to the left of it are < it on the axis.
// - All children to the right of it are >= it on the axis.
// - All child axes are their parent's axis + 1 (mod # dimensions).
// - All children have the correct parent.
//
// Returns nil if valid, or an error describing something broken in the tree.
func (n *Node) validate() error {
	var err error = nil
	if n.leftChild != nil {
		f := func(check *Node) {
			if check.Coordinates[n.axis] >= n.Coordinates[n.axis] {
				err = errors.New(check.String() + " is right of " + n.String() + " on axis " + strconv.FormatInt(int64(n.axis), 10))
			}
		}
		n.leftChild.traverse(f)
		// check all subtrees / dimensions
		if err != nil {
			return err
		}
		// make sure axes are sensible
		if expected := (n.axis + 1) % len(n.Coordinates); n.leftChild.axis != expected {
			return errors.New("Child axis " + strconv.FormatInt(int64(n.leftChild.axis), 10) + " isn't parent axis + 1 (" + strconv.FormatInt(int64(expected), 10))
		}
		// make sure parental relationships are correct
		if n.leftChild.parent == nil {
			return errors.New("Child " + n.leftChild.String() + " is missing parent " + n.String())
		} else if n.leftChild.parent != n {
			return errors.New("Child " + n.leftChild.String() + " is has incorrect parent " + n.leftChild.parent.String() + ", should be " + n.String())
		}

		// finally check all subtrees
		if err = n.leftChild.validate(); err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	if n.rightChild != nil {
		f := func(check *Node) {
			if check.Coordinates[n.axis] < n.Coordinates[n.axis] {
				err = errors.New(check.String() + " is left of " + n.String() + " on axis " + strconv.FormatInt(int64(n.axis), 10))
			}
		}
		n.rightChild.traverse(f)
		// check all subtrees / dimensions
		if err != nil {
			return err
		}
		// make sure axes are sensible
		if expected := (n.axis + 1) % len(n.Coordinates); n.rightChild.axis != expected {
			return errors.New("Child axis " + strconv.FormatInt(int64(n.rightChild.axis), 10) + " isn't parent axis + 1 (" + strconv.FormatInt(int64(expected), 10))
		}
		// make sure parental relationships are correct
		if n.rightChild.parent == nil {
			return errors.New("Child " + n.rightChild.String() + " is missing parent " + n.String())
		} else if n.rightChild.parent != n {
			return errors.New("Child " + n.rightChild.String() + " is has incorrect parent " + n.rightChild.parent.String() + ", should be " + n.String())
		}

		// finally check all subtrees
		if err = n.rightChild.validate(); err != nil {
			return err
		}
	}
	return err
}

// Returns Depth of the deepest branch of this Tree.
func (t *Tree) Depth() int {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.depth()
}

// Returns depth of the deepest branch of this (sub)tree.
func (n *Node) depth() int {
	if n == nil {
		return 0
	}
	left_depth := n.leftChild.depth() + 1
	right_depth := n.rightChild.depth() + 1
	if left_depth > right_depth {
		return left_depth
	}
	return right_depth
}

// Returns number of nodes in the Tree.
func (t *Tree) Size() int {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.size()
}

// Returns number of nodes in this (sub)tree.
func (n *Node) size() int {
	return len(n.nodeList())
}
