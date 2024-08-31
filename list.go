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
func (d *LinkedList[T]) AddBetween(data T, predecessor *Node[T], successor *Node[T]) (*Node[T], *LinkedList[T]) {
	n := &Node[T]{Data: data, Prev: predecessor, Next: successor}

	if predecessor != nil {
		predecessor.Next = n
	} else {
		// If predecessor is nil, n becomes the new head
		d.header.Next = n
	}

	if successor != nil {
		successor.Prev = n
	} else {
		// If successor is nil, n becomes the new tail
		d.trailer.Next = n
	}

	if d.IsLast(successor) {
		successor.Next = n
	}

	d.Size++
	return n, d
}

func (d *LinkedList[T]) InsertPrev(data T, current *Node[T]) (*Node[T], *LinkedList[T]) {
	n := &Node[T]{Data: data, Next: nil, Prev: nil}

	n.Next = current
	n.Prev = current.Prev
	current.Prev = n

	return n, d

}

// AddFirst adds a new node to the beginning of the list.
// time-complexity: O(1)
func (d *LinkedList[T]) AddFirst(data T) *Node[T] {
	new_node, _ := d.AddBetween(data, d.header, d.header.Next)
	return new_node
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

	if predecessor != nil {
		predecessor.Next = successor
	} else {
		// If n is the head node, update the head pointer
		d.header.Next = successor
	}

	if successor != nil {
		successor.Prev = predecessor
	} else {
		// If n is the tail node, update the tail pointer
		d.trailer.Next = predecessor
	}

	n.Next = nil
	n.Prev = nil

	d.Size--

	return n.Data
}

func (d *LinkedList[T]) RemoveFirstFound(e T) (*LinkedList[T], *Node[T]) {
    current_node := d.header.Next

    if current_node == nil {
        // The list is empty
        return d, nil
    }

    for {
        if current_node.Data == e {
            // Determine the next node before removal
            next_node := current_node.Next

            prev_node := current_node.Prev

            // Update the pointers
            if prev_node != nil {
                prev_node.Next = next_node
            } else {
                // If there's no previous node, this is the head node
                d.header.Next = next_node
            }

            if next_node != nil {
                next_node.Prev = prev_node
            } else {
                // If there's no next node, this is the tail node
                d.trailer.Prev = prev_node
            }

            // Handle the circular nature
            if current_node == d.header.Next {
                d.header.Next = next_node
                if d.header.Next == d.header {
                    // List is now empty
                    d.header.Next = nil
                    d.trailer.Prev = nil
                }
            }
            if current_node == d.trailer.Prev {
                d.trailer.Prev = prev_node
            }

            // Remove the current node
            d.Remove(current_node)

            // Return the next available node after removal
            return d, next_node
        }

        current_node = current_node.Next

        // If we are back at the header, the element was not found
        if current_node == d.header.Next {
            break
        }
    }

    return d, nil
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
