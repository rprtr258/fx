package main

import (
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
	"github.com/rprtr258/fun"
)

func dropWrapAll(n *node) {
	for ; n != nil; n = fun.IF(n.isCollapsed(), n.collapsed, n.next) {
		if fun.Deref(n.value) != "" && fun.Deref(n.value)[0] == '"' {
			n.dropChunks()
		}
	}
}

func wrapAll(n *node, termWidth int) {
	if termWidth <= 0 {
		return
	}

	for ; n != nil; n = fun.IF(n.isCollapsed(), n.collapsed, n.next) {
		if fun.Deref(n.value) == "" || fun.Deref(n.value)[0] != '"' {
			continue
		}

		n.dropChunks()
		lines := doWrap(n, termWidth)
		if len(lines) <= 1 {
			continue
		}

		n.chunk = fun.Ptr(lines[0])
		for i := 1; i < len(lines); i++ {
			n.insertChunk(&node{
				directParent: n,
				depth:        n.depth,
				chunk:        fun.Ptr(lines[i]),
				comma:        n.comma && i == len(lines)-1,
			})
		}
	}
}

func doWrap(n *node, termWidth int) []string {
	width := n.depth * 2
	if n.key != nil {
		for _, ch := range *n.key {
			width += runewidth.RuneWidth(ch)
		}
		width += 2 // for ": "
	}

	lines := make([]string, 0, 1)
	start, end := 0, 0
	value := fun.Deref(n.value)
	for _, r := range value {
		w := runewidth.RuneWidth(r)
		if width+w > termWidth {
			lines = append(lines, string([]byte(value)[start:end]))
			start = end
			width = n.depth * 2
		}
		width += w
		end += utf8.RuneLen(r)
	}
	if start < end {
		lines = append(lines, string([]byte(value)[start:]))
	}
	return lines
}
