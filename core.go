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

// A K-Dimensional Tree library, based on an algorithmic description from:
// http://en.wikipedia.org/wiki/K-d_tree
// and implementation ideas from:
// http://hackage.haskell.org/package/KdTree
//
// Licensed under the LGPL Version 3: http://www.gnu.org/licenses/
package kdtree

import (
	"errors"
	"strconv"
)

/***** Basic Tree Operations *****/

// Tree node, can be the parent for a subtree.
type Node struct {
	Index int // index from original data structure

	// Axis for plane of bisection for this node, determined when added to a tree.
	axis        int
	Coordinates []float64
	tree        *Tree // Tree this node belongs to, avoids reverse scan for root node.
	parent      *Node // Parent == nil is a tree root.
	leftChild   *Node // Nodes < Location on this axis.
	rightChild  *Node // Nodes >= Location on this axis.
}

// Create a new node from a set of coordinates.
func NewNode(coords []float64) *Node {
	n := new(Node)
	n.Coordinates = coords

	return n
}

func String(list []float64) string {
	out := "("
	for i := 0; i < len(list); i++ {
		out += " " + strconv.FormatFloat(list[i], 'G', 5, 64)
	}
	out += " )"
	return out
}

// String representation of a node.
func (n *Node) String() string {
	out := "[ " + String(n.Coordinates) + ": axis = " + strconv.FormatInt(int64(n.axis), 10) + " ]"
	return out
}

// Adds new Node or subtree to existing (sub)tree. Returns an error if the Node can't be added to the tree.
func (n *Node) add(newnode *Node) error {
	// Check dimensions of new node at tree root.
	if n.parent == nil {
		if len(n.Coordinates) != len(newnode.Coordinates) {
			return errors.New("Node with " + string(len(newnode.Coordinates)) + " dimensions can't be added to tree with " + string(len(n.Coordinates)) + " dimensions.")
		}
	}

	// erase any existing parent to node being added
	if newnode.parent != nil {
		newnode.parent = nil
	}

	// re-add any children first
	if newnode.leftChild != nil {
		if err := n.add(newnode.leftChild); err != nil {
			return err
		}
		newnode.leftChild = nil
	}
	if newnode.rightChild != nil {
		if err := n.add(newnode.rightChild); err != nil {
			return err
		}
		newnode.rightChild = nil
	}

	// now place this node
	if newnode.Coordinates[n.axis] < n.Coordinates[n.axis] {
		if n.leftChild == nil {
			newnode.axis = (n.axis + 1) % len(n.Coordinates)
			n.leftChild = newnode
			newnode.parent = n
			return nil
		} else {
			n.leftChild.add(newnode)
		}
	} else {
		if n.rightChild == nil {
			newnode.axis = (n.axis + 1) % len(n.Coordinates)
			n.rightChild = newnode
			newnode.parent = n
			return nil
		} else {
			n.rightChild.add(newnode)
		}
	}

	return nil
}

// Removes node from the tree it belongs to, adjusting other nodes as necessary.
// If this operation creates a new tree root, it is returned, otherwise nil.
func (n *Node) remove() *Node {
	if n.parent != nil {
		if !(n.parent.leftChild == n || n.parent.rightChild == n) {
			panic(n.String() + " to be removed not attached to its parent: " + n.parent.String())
		}
		parent := n.parent
		// remove references to this node from the parent
		if parent.leftChild == n {
			parent.leftChild = nil
		}
		// avoiding "else" auto-corrects the potential error case where parent.leftChild == parent.rightChild == n
		if parent.rightChild == n {
			parent.rightChild = nil
		}
		// remove reference to parent
		n.parent = nil

		// re-add any children to the previous level
		if n.leftChild != nil {
			if err := parent.add(n.leftChild); err != nil {
				panic("Unexpected error while removing node: " + err.Error())
			}
		}
		if n.rightChild != nil {
			if err := parent.add(n.rightChild); err != nil {
				panic("Unexpected error while removing node: " + err.Error())
			}
		}

		// remove references to children
		n.leftChild = nil
		n.rightChild = nil
	} else { // tree root
		switch {
		// arbitrarily rebalance so n.rightChild is the new tree root
		case n.leftChild != nil && n.rightChild != nil:
			n.rightChild.parent = nil
			if err := n.rightChild.add(n.leftChild); err != nil {
				// should never be an error on internal tree ops
				panic("Unexpected error adding subtree to new root: " + err.Error())
			}
			return n.rightChild
		case n.leftChild != nil: // implied n.rightChild == nil
			n.leftChild.parent = nil // new tree root
			return n.leftChild
		case n.rightChild != nil: // implied n.leftChild == nil
			n.rightChild.parent = nil // new tree root
			return n.rightChild
		}
		// case: n.leftChild == nil && n.rightChild == nil means empty tree
		return nil
	}
	n.tree = nil
	return nil
}

// Performs a left depth first tree traversal, running function f on every Node found.
func (n *Node) traverse(f func(*Node)) {
	if n != nil {
		if n.leftChild != nil {
			n.leftChild.traverse(f)
		}
		if n.rightChild != nil {
			n.rightChild.traverse(f)
		}
		f(n)
	}
}
