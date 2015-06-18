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
	"strconv"
)

/***** Basic Tree Operations *****/

// Tree node, can be the parent for a subtree.
type Node struct {
	Fare uint16 // index from original data structure

	// Axis for plane of bisection for this node, determined when added to a tree.
	axis        int
	Coordinates [4]float64
	leftChild   *Node // Nodes < Location on this axis.
	rightChild  *Node // Nodes >= Location on this axis.
}

// Create a new node from a set of coordinates.
func NewNode(coords [4]float64) *Node {
	n := new(Node)
	n.Coordinates = coords

	return n
}

func String(list [4]float64) string {
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
