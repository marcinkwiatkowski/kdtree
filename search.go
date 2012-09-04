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

/***** Tree Search Functions *****/

// Searches tree for node at exact coords. Returns nil if no node matching coords found,
// or if len(coords) != tree dimensions.
func (n *Node) Find(coords []float64) *Node {
	if len(coords) != len(n.Coordinates) {
		return nil
	}

	axis := n.axis
	if coords[axis] < n.Coordinates[axis] {
		if n.leftChild == nil {
			return nil
		} else {
			return n.leftChild.Find(coords)
		}
	} else if coords[axis] == n.Coordinates[axis] {
		if equal(coords, n.Coordinates) {
			return n
		}
	}
	// implicit else
	if n.rightChild == nil {
		return nil
	}
	// implicit else
	return n.rightChild.Find(coords)
}

// Finds the root of the tree from an arbitrary node.
func (n *Node) Root() *Node {
	if n.parent == nil {
		return n
	}
	return n.parent.Root()
}

// Range parameter, used to search the k-d tree.
type Range struct {
	Axis int
	Min  float64
	Max  float64
}

// Find a list of nodes matching the supplied list of dimensional
// Ranges. Ranges are considered to be required, so if two exclude each other,
// this will result in an empty result set. If an axis outside of the tree's
// dimensions is specified, nil is returned.
//func (n *Node) GetRange([]Range) []*Node {

//}

// Tests equality of float slices, returns false if lengths or any values contained within differ.
func equal(a, b []float64) bool {
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
