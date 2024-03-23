package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNode_paths(t *testing.T) {
	t.Parallel()

	s := `{"a": 1, "b": {"f": 2}, "c": [3, 4]}`
	n := nodeparse(s, parse(s))

	var paths []string
	var nodes []*node
	n.paths("", &paths, &nodes)
	assert.Equal(t, []string{".a", ".b", ".b.f", ".c", ".c[0]", ".c[1]"}, paths)
}

func TestNode_children(t *testing.T) {
	t.Parallel()

	s := `{"a": 1, "b": {"f": 2}, "c": [3, 4]}`
	n := nodeparse(s, parse(s))
	paths, _ := n.children()
	assert.Equal(t, []string{"a", "b", "c"}, paths)
}
