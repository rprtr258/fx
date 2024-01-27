package main

import (
	"strconv"

	jsonpath "github.com/antonmedv/fx/path"
	"github.com/samber/lo"
)

type node struct {
	prev, next, end *node
	directParent    *node
	indirectParent  *node
	collapsed       *node
	depth           int
	key             *string
	value           *string
	chunk           *string
	chunkEnd        *node
	comma           bool
	index           int
}

// append ands a node as a child to the current node (body of {...} or [...]).
func (n *node) append(child *node) {
	if n.end == nil {
		n.end = n
	}
	n.end.next = child
	child.prev = n.end
	n.end = lo.Ternary(child.end == nil, child, child.end)
}

func (n *node) insertChunk(chunk *node) {
	if n.chunkEnd == nil {
		n.insertAfter(chunk)
	} else {
		n.chunkEnd.insertAfter(chunk)
	}
	n.chunkEnd = chunk
}

func (n *node) insertAfter(child *node) {
	if n.next == nil {
		n.next = child
		child.prev = n
	} else {
		old := n.next
		n.next = child
		child.prev = n
		child.next = old
		old.prev = child
	}
}

func (n *node) dropChunks() {
	if n.chunkEnd == nil {
		return
	}

	n.chunk = nil

	n.next = n.chunkEnd.next
	if n.next != nil {
		n.next.prev = n
	}

	n.chunkEnd = nil
}

func (n *node) hasChildren() bool {
	return n.end != nil
}

func (n *node) parent() *node {
	if n.directParent == nil {
		return nil
	}

	if n.directParent.indirectParent != nil {
		return n.directParent.indirectParent
	}

	return n.directParent
}

func (n *node) isCollapsed() bool {
	return n.collapsed != nil
}

func (n *node) collapse() {
	if n.end == nil || n.isCollapsed() {
		return
	}

	n.collapsed = n.next
	n.next = n.end.next
	if n.next != nil {
		n.next.prev = n
	}
}

func (n *node) collapseRecursively() {
	for at := lo.Ternary(n.isCollapsed(), n.collapsed, n.next); at != nil && at != n.end; at = at.next {
		if !at.hasChildren() {
			continue
		}

		at.collapseRecursively()
		at.collapse()
	}
}

func (n *node) expand() {
	if n.isCollapsed() {
		if n.next != nil {
			n.next.prev = n.end
		}
		n.next = n.collapsed
		n.collapsed = nil
	}
}

func (n *node) expandRecursively() {
	for at := n; at != nil && at != n.end; at = at.next {
		at.expand()
	}
}

func (n *node) findChildByKey(key string) *node {
	for it := n.next; it != nil && it != n.end; it = lo.Switch[bool, *node](true).
		Case(it.chunkEnd != nil, it.chunkEnd).
		Case(it.end != nil, it.end).
		Default(it).next {
		if it.key == nil {
			continue
		}

		k, err := strconv.Unquote(*it.key)
		if err != nil {
			return nil
		}

		if k == key {
			return it
		}
	}
	return nil
}

func (n *node) findChildByIndex(index int) *node {
	for at := n.next; at != nil && at != n.end; at = lo.Ternary(at.end != nil, at.end, at).next {
		if at.index == index {
			return at
		}
	}
	return nil
}

func (n *node) paths(prefix string, paths *[]string, nodes *[]*node) {
	it := n.next
	for it != nil && it != n.end {
		var path string
		if it.key != nil {
			quoted := *it.key
			unquoted, err := strconv.Unquote(quoted)
			path = lo.Ternary(
				err == nil && jsonpath.Identifier.MatchString(unquoted),
				"."+unquoted,
				"["+quoted+"]",
			)
		} else if it.index >= 0 {
			path = "[" + strconv.Itoa(it.index) + "]"
		}
		path = prefix + path

		*paths = append(*paths, path)
		*nodes = append(*nodes, it)

		if it.hasChildren() {
			it.paths(path, paths, nodes)
			it = it.end.next
		} else {
			it = it.next
		}
	}
}

func (n *node) children() ([]string, []*node) {
	if !n.hasChildren() {
		return nil, nil
	}

	var paths []string
	var nodes []*node
	for it := lo.Ternary(n.isCollapsed(), n.collapsed, n.next); it != nil && it != n.end; it = lo.Ternary(it.hasChildren(), it.end, it).next {
		if it.key == nil {
			continue
		}

		key := *it.key
		unquoted, err := strconv.Unquote(key)
		if err == nil {
			key = unquoted
		}
		paths = append(paths, key)
		nodes = append(nodes, it)
	}
	return paths, nodes
}
