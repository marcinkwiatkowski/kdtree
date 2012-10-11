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
)

/***** Tree Search Functions *****/

// Searches Tree for node at exact coords. Returns (nil, nil) if no node matching coords found,
// or (nil, error) if len(coords) != tree dimensions.
func (t *Tree) Find(coords []float64) (*Node, error) {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.find(coords)
}

// Searches (sub)tree for node at exact coords. Returns (nil, nil) if no node matching coords found,
// or (nil, error) if len(coords) != tree dimensions.
func (n *Node) find(coords []float64) (*Node, error) {
	if len(coords) != len(n.Coordinates) {
		return nil, errors.New("Search coordinates have " + string(len(coords)) + " dimensions, tree has " + string(len(n.Coordinates)) + " dimensions.")
	}

	axis := n.axis
	if coords[axis] < n.Coordinates[axis] {
		if n.leftChild == nil {
			return nil, nil
		} else {
			return n.leftChild.find(coords)
		}
	} else if coords[axis] == n.Coordinates[axis] {
		if equal_fl(coords, n.Coordinates) {
			return n, nil
		}
	}
	// implicit else
	if n.rightChild == nil {
		return nil, nil
	}
	// implicit else
	return n.rightChild.find(coords)
}

// Finds the root of the tree from an arbitrary node.
func (n *Node) root() *Node {
	if n.parent == nil {
		return n
	}
	return n.parent.root()
}

// Range parameter, used to search the k-d tree.
type Range struct {
	Min float64
	Max float64
}

// Find a list of Nodes in Tree matching the supplied map of dimensional
// Ranges. The map index is used as the axis to restrict. 
// Use math.Inf() to create remove the restriction on Min or Max.
//
// If no results are found, (nil, nil) is returned.
// If an axis outside of the tree's dimensions is specified, nil is returned with an error.
func (t *Tree) FindRange(ranges map[int]Range) ([]*Node, error) {
	t.Mutex.RLock()
	defer t.Mutex.RUnlock()
	return t.Root.findRange(ranges)
}

// Find a list of nodes matching the supplied map of dimensional
// Ranges. The map index is used as the axis to restrict. 
// Use math.Inf() to create remove the restriction on Min or Max.
//
// If no results are found, (nil, nil) is returned.
// If an axis outside of the tree's dimensions is specified, nil is returned with an error.
func (n *Node) findRange(ranges map[int]Range) ([]*Node, error) {
	if n == nil {
		return nil, nil
	}

	result := make([]*Node, 0, 10)
	// check to see if the current node should be returned
	add := true
	for a, r := range ranges {
		if a >= len(n.Coordinates) {
			return nil, errors.New("Range on axis " + string(a) + " exceeds tree dimensions.")
		}
		if a < 0 {
			return nil, errors.New("Negative axes are invalid.")
		}

		if n.Coordinates[a] < r.Min || n.Coordinates[a] > r.Max {
			add = false
			break
		}
	}
	if add {
		result = append(result, n)
	}

	// search subtrees
	r, ok := ranges[n.axis]
	// search subtree if we're not restricting this axis, or if restrictions match.
	if !ok || r.Min < n.Coordinates[n.axis] {
		if left, err := n.leftChild.findRange(ranges); err == nil {
			result = append(result, left...)
		} else {
			return result, err
		}
	}
	if !ok || r.Max >= n.Coordinates[n.axis] {
		if right, err := n.rightChild.findRange(ranges); err == nil {
			result = append(result, right...)
		} else {
			return result, err
		}
	}

	return result, nil
}

// Tests equality of float slices, returns false if lengths or any values contained within differ.
func equal_fl(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
