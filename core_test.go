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
	"math/rand"
	"strconv"
	"testing"
	"time"
)

/****** Library Unit Tests and Benchmarks *****/
func init() {
	rand.Seed(time.Now().Unix())
}

// Generate a random set of coordinates with arbitrary number of
// dimensions. Returned slice has len() == dimensions.
func rndCoords(dimensions int) []float64 {
	coords := make([]float64, dimensions)
	for i := 0; i < dimensions; i++ {
		coords[i] = rand.Float64()
	}
	return coords
}

// Generate a random node list with given dimensions and size.
func genlist(dimensions, size int) []*Node {
	nodelist := make([]*Node, size)
	for i := 0; i < size; i++ {
		nn := NewNode(rndCoords(dimensions))
		nodelist[i] = nn
	}
	return nodelist
}

// Generates a list of random nodes, then inserts them into a k-d tree using BuildTree.
// Test fails if any nodes from the list are missing from the tree.
// This test also implicitely tests Node.Find().
func TestBuildTree(t *testing.T) {
	// test size: 6 dimensions == 2 * normal 3d, 100000 nodes == "5 9s" of accuracy.
	nl := genlist(6, 100000)
	tree := BuildTree(nl)
	if tree == nil {
		t.Fatal("Tree not generated!")
	}
	if err := tree.Validate(); err != nil {
		t.Fatal("Tree is not valid: " + err.Error())
	}
	for k, n := range nl {
		if search, err := tree.Find(n.Coordinates); err == nil && search == nil {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + " not found!")
		} else if search != n {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + ", found " + search.String())
		} else if err != nil {
			t.Fatal("Error while searching tree:", err)
		}
	}
}

// Test speed to create a new tree from randomly generated nodes.
func BenchmarkBuildTree(b *testing.B) {
	// We're benchmarking tree generation, not node list generation, pause until
	// nl is created.
	b.StopTimer()
	nl := genlist(6, b.N)
	b.StartTimer()

	BuildTree(nl)
}

func TestFindRoot(t *testing.T) {
	nl := genlist(6, 100000)
	tree := BuildTree(nl)
	for _, n := range nl {
		go func() {
			root := n.root()
			if root != tree.Root {
				defer t.Fatal("Found incorrect root " + root.String() + " from node " + n.String())
			}
		}()
	}
}

func TestAddNodes(t *testing.T) {
	nl := genlist(6, 100000)
	tree := BuildTree(nl)
	// insert 1000 nodes
	donechan := make(chan bool, 100)
	for i := 0; i < 1000; i++ {
		n := NewNode(rndCoords(6))
		nl = append(nl, n)
		go func() {
			if err := tree.Add(n); err != nil {
				defer t.Fatal("Failed to add node " + n.String() + ": " + err.Error())
			}
			donechan <- true
		}()
	}
	// wait for goroutines to finish
	for i := 0; i < 1000; i++ {
		<-donechan
	}
	if err := tree.Validate(); err != nil {
		t.Fatal("Tree is not valid after adding nodes: " + err.Error())
	}
	// compare nodes against list
	for k, n := range nl {
		if search, err := tree.Find(n.Coordinates); err == nil && search == nil {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + " not found!")
		} else if search != n {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + ", found " + search.String())
		} else if err != nil {
			t.Fatal("Error while searching tree:", err)
		}
	}
}

func BenchmarkAddNodes(b *testing.B) {
	tree := new(Tree)
	for i := 0; i < b.N/2; i++ {
		go tree.Add(NewNode(rndCoords(6)))
	}
}

func BenchmarkFind(b *testing.B) {
	b.StopTimer()
	nl := genlist(6, b.N)
	tree := BuildTree(nl)
	donechan := make(chan bool, 100)
	b.StartTimer()
	for k, n := range nl {
		go func() {
			if search, err := tree.Find(n.Coordinates); err == nil && search == nil {
				defer b.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + " not found!")
			} else if search != n {
				defer b.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + ", found " + search.String())
			} else if err != nil {
				defer b.Fatal("Error while searching tree:", err)
			}
			donechan <- true
		}()
	}
	// wait for goroutines to finish
	for _, _ = range nl {
		<-donechan
	}
}

func TestAddSubtree(t *testing.T) {
	nl1 := genlist(6, 75000)
	nl2 := genlist(6, 25000)
	tree1 := BuildTree(nl1)
	tree2 := BuildTree(nl2)
	tree1.Add(tree2.Root)
	for k, n := range nl1 {
		if search, err := tree1.Find(n.Coordinates); err == nil && search == nil {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + " not found!")
		} else if search != n {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + ", found " + search.String())
		} else if err != nil {
			t.Fatal("Error while searching tree:", err)
		}
	}
	for k, n := range nl2 {
		if search, err := tree1.Find(n.Coordinates); err == nil && search == nil {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + " not found!")
		} else if search != n {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + ", found " + search.String())
		} else if err != nil {
			t.Fatal("Error while searching tree:", err)
		}
	}
}

func TestRemoveNodes(t *testing.T) {
	// order of magnitude smaller, because removals are an order of magnitude slower.
	nl := genlist(6, 10000)
	tree := BuildTree(nl)
	// remove nodes from end of nodelist
	for i := len(nl) - 500; i < len(nl); i++ {
		size := tree.Size()
		if err := tree.Remove(nl[i]); err != nil {
			t.Fatal("Failed to remove node " + nl[i].String() + ", " + err.Error())
		}
		newsize := tree.Size()
		diff := size - newsize
		if diff != 1 {
			t.Fatal("Removing one node shrunk the tree by " + strconv.FormatInt(int64(diff), 10) + " nodes.")
		}
	}
	if err := tree.Validate(); err != nil {
		t.Fatal("Tree is not valid after removing nodes: " + err.Error())
	}
	if curlist := tree.NodeList(); len(curlist) != (10000 - 500) {
		t.Fatal("Tree has incorrect number of nodes after removal: " + strconv.FormatInt(int64(len(curlist)), 10))
	}

	// compare nodes against shortened list
	for k := 0; k < len(nl)-500; k++ {
		n := nl[k]
		search, _ := tree.Find(n.Coordinates)
		if search == nil {
			t.Error(strconv.FormatInt(int64(k), 10) + ": " + n.String() + " not found!")
			if n.parent != nil {
				t.Log("Parent is " + n.parent.String())
				if n.parent.leftChild != nil {
					t.Log("p.leftchild = " + n.parent.leftChild.String())
				}
				if n.parent.rightChild != nil {
					t.Log("p.rightchild = " + n.parent.rightChild.String())
				}
			}
			t.FailNow()
		} else if search != n {
			t.Fatal(strconv.FormatInt(int64(k), 10) + ": " + n.String() + ", found " + search.String())
		}
	}
}

func BenchmarkRemoveNodes(b *testing.B) {
	b.StopTimer()
	nl := genlist(6, b.N*2)
	tree := BuildTree(nl)
	donechan := make(chan bool)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		go func() {
			tree.Remove(nl[i])
			donechan <- true
		}()
	}
	// wait for goroutines to finish
	for i := 0; i < b.N; i++ {
		<-donechan
	}
}

func TestBalance(t *testing.T) {
	// first, generate an unbalanced tree on purpose
	tree := new(Tree)
	tree.Add(NewNode([]float64{0.0, 0.0, 0.0, 0.0, 0.0, 0.0}))
	for i := 0; i < 100000; i++ {
		// Because the tree root is (0.0...), and math.rand generates numbers in [0.0,1.0), these nodes
		// will all fall to the right of the root.
		n := NewNode(rndCoords(6))
		if err := tree.Add(n); err != nil {
			t.Fatal(err)
		}
	}
	if tree.Root.leftChild.size() > 0 {
		t.Fatal("Left branch should always be empty after unbalanced generation.")
	}
	start_depth := tree.Depth()
	tree.Balance()
	end_depth := tree.Depth()
	depth_diff := start_depth - end_depth

	// check for balancing errors
	if depth_diff <= 0 {
		t.Fatal("New tree has a depth >= old tree.")
	}
	left_size := tree.Root.leftChild.size()
	right_size := tree.Root.rightChild.size()
	size_diff := left_size - right_size
	if size_diff > 10 || size_diff < -10 {
		t.Error("Left and right branches have a node difference > 10 (" + strconv.FormatInt(int64(size_diff), 10))
	}
}

func BenchmarkBalance(b *testing.B) {
	b.StopTimer()
	nl := genlist(6, b.N)
	tree := BuildTree(nl)
	b.StartTimer()

	tree.Balance()
}

func TestFindRange(t *testing.T) {
	nl := genlist(6, 20000)
	tree := BuildTree(nl)
	donechan := make(chan bool)

	for i := 0; i < 100; i++ {
		go func() {
			ranges := make(map[int]Range)
			for axis := rand.Intn(6); len(ranges) < rand.Intn(6)+1; axis = rand.Intn(6) {
				r := Range{rand.Float64(), rand.Float64()}
				if r.Min > r.Max {
					r.Min, r.Max = r.Max, r.Min
				}
				ranges[axis] = r
			}
			results1, err := tree.FindRange(ranges)
			if err != nil {
				t.Fatal(err)
			}
			snl := sortableNodeList{0, nl}
			results2, err := snl.findrange(ranges)
			if err != nil {
				defer t.Fatal(err)
			}

			if len(results1) != len(results2) {
				defer t.Fatal("Tree FindRange returned", len(results1), "nodes, list findrange returned", len(results2))
			}
			for _, n := range results1 {
				if _, ok := find_nl(results2, n); !ok {
					defer t.Fatal("Node from tree results not found in results list:", n)
				}
			}
			for _, n := range results2 {
				if _, ok := find_nl(results1, n); !ok {
					defer t.Fatal("Node from results list not found in tree results:", n)
				}
			}
			donechan <- true
		}()
	}
	// wait for goroutines to complete	
	for i := 0; i < 100; i++ {
		<-donechan
	}
}

func BenchmarkFindRange(b *testing.B) {
	b.StopTimer()
	nl := genlist(6, b.N*2)
	tree := BuildTree(nl)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		ranges := make(map[int]Range)
		for axis := rand.Intn(6); len(ranges) < 2; axis = rand.Intn(6) {
			r := Range{rand.Float64(), rand.Float64()}
			if r.Min > r.Max {
				r.Min, r.Max = r.Max, r.Min
			}
			ranges[axis] = r
		}
		if _, err := tree.FindRange(ranges); err != nil {
			b.Fatal(err)
		}
	}
}
