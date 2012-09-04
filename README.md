kdtree - a k-dimensional tree implementation in Go
===================================================

This library was inspired by both the [Wikipedia K-d Tree article](http://en.wikipedia.org/wiki/K-d_tree)
and the [Haskell KdTree package](http://hackage.haskell.org/package/KdTree).

It implements a k-dimensional B-tree with float64 coordinates in Go. K-dimensional trees are a reasonably efficient
way of searching K-dimensional space for matching items by using bisecting planes at each binary tree branch.
The Wikipedia article explains this concept in greater detail.

This library implements most (all?) basic functionality you would expect to be available from such a
data structure, and every major operation includes unit tests and benchmarks.

Install
-------

This library is fully installable with the go command:

	go get -u -v github.com/unit3/kdtree

Tests and Benchmarks
--------------------

Unit tests and benchmarks are implemented in *_test.go files using the fantastic builtin
"testing" library. To run the full suite of tests and benchmarks, do the following in the source
directory:

	go test -v -bench=.*

Planned Improvements
--------------------

The Remove() function is a pretty naive implementation, and is about an order of magnitude slower
than other operations:

	BenchmarkBuildTree       1000000      7459 ns/op
	BenchmarkAddNodes        2000000      1040 ns/op
	BenchmarkRemoveNodes      200000     18795 ns/op


License
-------

kdtree is free software: you can redistribute it and/or modify
it under the terms of the GNU Lesser General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

kdtree is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Lesser General Public License for more details.

You should have received a copy of the GNU Lesser General Public License
along with kdtree.  If not, see [http://www.gnu.org/licenses/](http://www.gnu.org/licenses/).
