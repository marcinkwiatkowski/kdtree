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
)

/***** Tree Management Functions *****/

// Returns a slice of all distinct nodes in the tree. This is done by a tree traversal,
// and will be equally as slow.
func (n *Node) NodeList() []*Node {
	nodelist := make([]*Node, 0, 100)
	f := func(n *Node) {
		nodelist = append(nodelist, n)
	}
	n.Traverse(f)

	return nodelist
}

// Wrapper for a slice of nodes implementing sort.Interface for different dimensional axes.
type SortableNodeList struct {
	// dimension axis to sort on
	Axis  int
	Nodes []*Node
}

func (snl *SortableNodeList) Len() int {
	return len(snl.Nodes)
}

func (snl *SortableNodeList) Less(i, j int) bool {
	return snl.Nodes[i].Coordinates[snl.Axis] < snl.Nodes[j].Coordinates[snl.Axis]
}

func (snl *SortableNodeList) Swap(i, j int) {
	tmp := snl.Nodes[i]
	snl.Nodes[i] = snl.Nodes[j]
	snl.Nodes[j] = tmp
}

// Builds a tree from a list of nodes. Returns the root Node of the new tree.
// This is destructive, and will break any existing tree these nodes may be a member of.
// This is intended to be used to build an new tree, or as part of a tree Balance.
// This is a recursive function, you should always call it with depth = 0, parent = nil.
func BuildTree(nodes []*Node, depth int, parent *Node) *Node {
	var root *Node
	// special case handling first
	switch len(nodes) {
	case 0:
		root = nil
	case 1:
		dimensions := len(nodes[0].Coordinates)
		root = nodes[0]
		root.Mutex.Lock()
		defer root.Mutex.Unlock()

		root.axis = depth % dimensions
		root.parent = parent
		root.leftChild = nil
		root.rightChild = nil
	default:
		median := (len(nodes) / 2) - 1 // -1 so that it's a slice index
		dimensions := len(nodes[0].Coordinates)

		snl := new(SortableNodeList)
		snl.Axis = depth % dimensions
		snl.Nodes = make([]*Node, len(nodes))
		copy(snl.Nodes, nodes)
		sort.Sort(snl)

		root = snl.Nodes[median]

		root.Mutex.Lock()
		defer root.Mutex.Unlock()
		root.axis = snl.Axis
		root.parent = parent
		root.leftChild = BuildTree(snl.Nodes[0:median], depth+1, root)
		root.rightChild = BuildTree(snl.Nodes[median+1:], depth+1, root)
	}

	return root
}

// Balances a tree by re-inserting all nodes into a new tree.
// Returns the root Node for the new tree.
func (n *Node) Balance() *Node {
	nodelist := n.NodeList()
	return BuildTree(nodelist, 0, nil)
}

// Checks that the (sub)tree below this node is valid:
// - All children to the left of it are < it on the axis.
// - All children to the right of it are >= it on the axis.
// - All child axes are their parent's axis + 1 (mod # dimensions).
// - All children have the correct parent.
//
// Returns nil if valid, or an error describing something broken in the tree.
func (n *Node) Validate() error {
	var err error = nil
	if n.leftChild != nil {
		f := func(check *Node) {
			if check.Coordinates[n.axis] >= n.Coordinates[n.axis] {
				err = errors.New(check.String() + " is right of " + n.String() + " on axis " + strconv.FormatInt(int64(n.axis), 10))
			}
		}
		n.leftChild.Traverse(f)
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
		if err = n.leftChild.Validate(); err != nil {
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
		n.rightChild.Traverse(f)
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
		if err = n.rightChild.Validate(); err != nil {
			return err
		}
	}
	return err
}

// Returns depth of the deepest branch of this (sub)tree.
func (n *Node) Depth() int {
	if n == nil {
		return 0
	}
	left_depth := n.leftChild.Depth() + 1
	right_depth := n.rightChild.Depth() + 1
	if left_depth > right_depth {
		return left_depth
	}
	return right_depth
}

// Returns number of nodes in this (sub)tree.
func (n *Node) Size() int {
	return len(n.NodeList())
}
