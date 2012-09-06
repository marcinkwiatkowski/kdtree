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

/***** Node list management functions *****/

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
	snl.Nodes[i], snl.Nodes[j] = snl.Nodes[j], snl.Nodes[i]
}

// Perform the same search as Node.FindRange() on a list of nodes, used in
// unit testing. Axis is ignored in this function.
func (snl *SortableNodeList) findrange(ranges map[int]Range) ([]*Node, error) {
	result := make([]*Node, 0, len(snl.Nodes))
	for _, n := range snl.Nodes {
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
			}
		}
		if add {
			result = append(result, n)
		}
	}

	return result, nil
}

// Find a Node in a slice of Nodes. Returns (slice index, true) if found, or (_, false) if missing.
func find_nl(nl []*Node, n1 *Node) (int, bool) {
	for i, n2 := range nl {
		if n1 == n2 {
			return i, true
		}
	}
	return 0, false
}
