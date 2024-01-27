package main

import "github.com/samber/lo"

type search struct {
	err     error
	results []*node
	cursor  int
	values  map[*node][]match
	keys    map[*node][]match
}

func newSearch() *search {
	return &search{
		results: make([]*node, 0),
		values:  make(map[*node][]match),
		keys:    make(map[*node][]match),
	}
}

type match struct {
	start, end int
	index      int
}

type piece struct {
	b     []byte
	index int
}

func safeSlice(b []byte, start, end int) []byte {
	length := len(b)
	start = max(min(start, length), 0)
	end = max(min(end, length), 0)
	start = min(start, end)
	return b[start:end]
}

func splitBytesByIndexes(bb *string, indexes []match) []piece {
	b := []byte(lo.FromPtr(bb))
	out := make([]piece, 0, 1)
	pos := 0
	for _, pair := range indexes {
		out = append(out,
			piece{safeSlice(b, pos, pair.start), -1},
			piece{safeSlice(b, pair.start, pair.end), pair.index},
		)
		pos = pair.end
	}
	out = append(out, piece{safeSlice(b, pos, len(b)), -1})
	return out
}

func splitIndexesToChunks(chunks [][]byte, indexes [][]int, searchIndex int) [][]match {
	chunkIndexes := make([][]match, len(chunks))
	for index, idx := range indexes {
		i, j := idx[0], idx[1]
		position := 0
		for k, chunk := range chunks {
			// If start index lies in this chunk
			if i >= position+len(chunk) {
				continue
			}

			chunkIndexes[k] = append(chunkIndexes[k], match{
				start: i - position,
				end:   min(len(chunk), j-position),
				index: searchIndex + index,
			})

			if j-position <= len(chunk) {
				break
			}

			// Adjust the starting index for the next chunk
			position += len(chunk)
			i = position
		}
	}
	return chunkIndexes
}
