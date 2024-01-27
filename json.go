package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"

	"github.com/rprtr258/fun"
	"github.com/samber/lo"
)

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

type jsonParser struct {
	depth int
}

func parse(data []byte) (_ *node, err error) {
	p := &jsonParser{}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	var m any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}

	return p.parseValue(m), nil
}

func (p *jsonParser) parseValue(m any) *node {
	switch m := m.(type) {
	case string:
		return p.parseString(m)
	case float64:
		return p.parseNumber(m)
	case map[string]any:
		return p.parseObject(m)
	case []any:
		return p.parseArray(m)
	case bool:
		return p.parseKeyword(strconv.FormatBool(m))
	case nil:
		return p.parseKeyword("null")
	default:
		panic(fmt.Sprintf("Unexpected type %T", m))
	}
}

func (p *jsonParser) parseString(m string) *node {
	return &node{
		depth: p.depth,
		value: fun.Ptr(strconv.Quote(m)),
	}
}

func (p *jsonParser) parseNumber(m float64) *node {
	return &node{
		depth: p.depth,
		// TODO: somehow get raw number representation
		value: fun.Ptr(strconv.FormatFloat(m, 'f', -1, 64)),
	}
}

func (p *jsonParser) parseObject(m map[string]any) *node {
	object := &node{
		depth: p.depth,
		value: fun.Ptr("{"),
	}

	keys := lo.Keys(m)
	sort.Strings(keys)

	for i, k := range keys {
		p.depth++
		key := p.parseString(k)
		value := p.parseValue(m[k])
		p.depth--

		key.key, key.value = key.value, value.value
		key.directParent = object
		key.next = value.next
		if key.next != nil {
			key.next.prev = key
		}
		key.end = value.end
		value.indirectParent = key
		object.append(key)

		if i < len(m)-1 {
			object.end.comma = true
		}
	}

	object.append(&node{
		depth:        p.depth,
		value:        fun.Ptr("}"),
		directParent: object,
		index:        -1,
	})
	return object
}

func (p *jsonParser) parseArray(m []any) *node {
	arr := &node{
		depth: p.depth,
		value: fun.Ptr("["),
	}

	for i := range m {
		p.depth++
		value := p.parseValue(m[i])
		value.directParent = arr
		value.index = i
		p.depth--

		arr.append(value)

		if i < len(m)-1 {
			arr.end.comma = true
		}
	}

	arr.append(&node{
		depth:        p.depth,
		value:        fun.Ptr("]"),
		directParent: arr,
		index:        -1,
	})
	return arr
}

func (p *jsonParser) parseKeyword(name string) *node {
	return &node{
		depth: p.depth,
		value: fun.Ptr(name),
	}
}
