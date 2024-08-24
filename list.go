package main

import (
	"fmt"
	"strings"
)

type Node[T comparable] struct {
	Data T
	Prev *Node[T]
	Next *Node[T]
}

type LinkedList[T comparable] struct {
	header  *Node[T] // header.Next is the first node in the list.
	trailer *Node[T] // trailer.Prev is the last node in the list.
	Size    int
}


// New constructs and returns an empty doubly linked-list.
// time-complexity: O(1)
func NewList[T comparable]() LinkedList[T] {
	var d LinkedList[T]

	d.header = &Node[T]{}
	d.trailer = &Node[T]{Prev: d.header}
	d.header.Next = d.trailer

	return d
}

// IsEmpty returns true if the linked-list doesn't contain any nodes.
// time-complexity: O(1)
func (d *LinkedList[T]) IsEmpty() bool {
	return d.Size == 0
}

// First returns the first element of the list. It returns false if the list is empty.
// time-complexity: O(1)
func (d *LinkedList[T]) First() (data T, ok bool) {
	if d.IsEmpty() {
		return
	}
	return d.header.Next.Data, true
}

// Last returns the last element of the list. It returns false if the list is empty.
// time-complexity: O(1)
func (d *LinkedList[T]) Last() (data T, ok bool) {
	if d.IsEmpty() {
		return
	}
	return d.trailer.Prev.Data, true
}

// Is Last
func (d *LinkedList[T]) IsLast(node *Node[T]) bool {
	if d.IsEmpty() || d.Size <= 1 {
		return true
	}

	return node == d.trailer
}

// AddBetween constructs a new node out of the given data and inserts it between the given two nodes.
// and returns the newly inserted node
// time-complexity: O(1)
func (d *LinkedList[T]) AddBetween(data T, predecessor *Node[T], successor *Node[T]) *Node[T] {
	n := &Node[T]{Data: data, Next: successor, Prev: predecessor}

	predecessor.Next = n
	successor.Prev = n

	d.Size++
	return n
}

// AddFirst adds a new node to the beginning of the list.
// time-complexity: O(1)
func (d *LinkedList[T]) AddFirst(data T) *Node[T] {
	return d.AddBetween(data, d.header, d.header.Next)
}

// AddLast adds a new node to the end of the list.
// time-complexity: O(1)
func (d *LinkedList[T]) AddLast(data T) {
	d.AddBetween(data, d.trailer.Prev, d.trailer)
}

// Remove removes the given node from the list. It returns the removed node's data.
// time-complexity: O(1)
func (d *LinkedList[T]) Remove(n *Node[T]) T {
	predecessor := n.Prev
	successor := n.Next

	predecessor.Next = successor
	successor.Prev = predecessor

	n.Next = nil
	n.Prev = nil

	d.Size--

	return n.Data
}

func (d *LinkedList[T]) RemoveFirstFound(e T) (*LinkedList[T], *Node[T]){
				current_node := d.header.Next

				for {
								if current_node.Data == e {
												d.Remove(current_node)
												current_node = current_node.Next
												break
								}
								current_node = current_node.Next
				}

				return d, current_node
}

// RemoveFirst removes and returns the first element of the list. It returns false if the list is empty.
// time-complexity: O(1)
func (d *LinkedList[T]) RemoveFirste() (data T, ok bool) {
	if d.IsEmpty() {
		return
	}
	return d.Remove(d.header.Next), true
}

// RemoveLast removes and returns the last element of the list. It returns false if the list empty.
// time-complexity: O(1)
func (d *LinkedList[T]) RemoveLast() (data T, ok bool) {
	if d.IsEmpty() {
		return
	}
	return d.Remove(d.trailer.Prev), true
}

// String returns the string representation of the list.
// time-complexity: O(n)
func (d *LinkedList[T]) String() string {
	var b strings.Builder

	b.WriteString("[ ")

	for current := d.header.Next; current != d.trailer; current = current.Next {
		b.WriteString(fmt.Sprint(current.Data))
		b.WriteString(" ")
	}

	b.WriteString("]")

	return b.String()
}

// ToSlice returns the linked-list as a slice.
// time-complexity: O(n)
func (d *LinkedList[T]) ToSlice() []T {
	r := make([]T, d.Size)

	for i, cur := 0, d.header.Next; cur != d.trailer && i < len(r); i, cur = i+1, cur.Next {
		r[i] = cur.Data
	}

	return r
}
