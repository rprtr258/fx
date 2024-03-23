package main

import (
	"github.com/rprtr258/fun"
)

type tokenKind int

const (
	tokenKindUnknown     tokenKind = iota
	tokenKindObjectStart           // {
	tokenKindObjectEnd             // }
	tokenKindArrayStart            // [
	tokenKindArrayEnd              // ]
	tokenKindString                // "..."
	tokenKindNumber                // 3.1415
	tokenKindTrue                  // true
	tokenKindFalse                 // false
	tokenKindNull                  // null
	tokenKindComma                 // ,
	tokenKindColon                 // :
)

type token struct {
	kind       tokenKind
	start, end int
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isWhitespace(ch rune) bool {
	return fun.Contains(ch, ' ', '\t', '\n', '\r')
}

type jsonParser struct {
	src []rune
	i   int
}

func (p *jsonParser) cur() (rune, bool) {
	if p.i >= len(p.src) {
		return 0, false
	}
	return p.src[p.i], true
}

func (p *jsonParser) next() {
	if p.i < len(p.src)-1 {
		p.i++
	}
}

func (p *jsonParser) rest() token {
	return token{
		kind:  tokenKindUnknown,
		start: p.i,
		end:   len(p.src),
	}
}

func (p *jsonParser) skipWhitespace() {
	for {
		if c, ok := p.cur(); !ok || !isWhitespace(c) {
			return
		}
		p.next()
	}
}

func (p *jsonParser) parseString() token {
	i := p.i
	p.next() // "
	for p.i < len(p.src) && p.src[p.i] != '"' {
		p.next() // TODO: also check for \X, \u0000, \U0000
	}
	p.next() // "
	return token{
		kind:  tokenKindString,
		start: i,
		end:   p.i,
	}
}

// TODO: test exhaustively
func (p *jsonParser) parseNumber() token {
	i := p.i
	if p.src[i] == '-' {
		p.next()
	}
	for p.i < len(p.src) && isDigit(p.src[p.i]) {
		p.next()
	}
	if p.i < len(p.src) && p.src[p.i] == '.' {
		p.next()
		for p.i < len(p.src) && isDigit(p.src[p.i]) {
			p.next()
		}
	}
	if p.i < len(p.src) && (p.src[p.i] == 'e' || p.src[p.i] == 'E') {
		p.next()
		if p.i < len(p.src) && (p.src[p.i] == '-' || p.src[p.i] == '+') {
			p.next()
		}
		for p.i < len(p.src) && isDigit(p.src[p.i]) {
			p.next()
		}
	}
	return token{
		kind:  tokenKindNumber,
		start: i,
		end:   p.i,
	}
}

func (p *jsonParser) parseObject() []token {
	kvs := []token{p.parseKeyword(tokenKindObjectStart, "{")}
	for {
		p.skipWhitespace()
		kvs = append(kvs, p.parseString()) // key
		p.skipWhitespace()
		kvs = append(kvs, p.parseKeyword(tokenKindColon, ":")) // :
		kvs = append(kvs, p.parseValue()...)                   // value
		if c, ok := p.cur(); !ok || c != ',' {
			break
		}
		kvs = append(kvs, p.parseKeyword(tokenKindComma, ",")) // ,
	}
	return append(kvs, p.parseKeyword(tokenKindObjectEnd, "}"))
}

func (p *jsonParser) parseArray() []token {
	elems := []token{p.parseKeyword(tokenKindArrayStart, "[")}
	for {
		elems = append(elems, p.parseValue()...) // value
		p.skipWhitespace()
		if c, ok := p.cur(); !ok || c != ',' {
			break
		}
		elems = append(elems, p.parseKeyword(tokenKindComma, ",")) // ,
	}
	return append(elems, p.parseKeyword(tokenKindArrayEnd, "]"))
}

func (p *jsonParser) parseKeyword(kind tokenKind, name string) token {
	i := p.i
	for _, r := range name {
		c, ok := p.cur()
		if !ok || c != r {
			return p.rest()
		}
		p.next()
	}
	return token{
		kind:  kind,
		start: i,
		end:   p.i,
	}
}

func (p *jsonParser) parseValue() []token {
	p.skipWhitespace()

	c, ok := p.cur()
	if !ok {
		return []token{p.rest()}
	}

	switch c {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		return []token{p.parseString()}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
		return []token{p.parseNumber()}
	case 't':
		return []token{p.parseKeyword(tokenKindTrue, "true")}
	case 'f':
		return []token{p.parseKeyword(tokenKindFalse, "false")}
	case 'n':
		return []token{p.parseKeyword(tokenKindNull, "null")}
	default:
		return []token{p.rest()}
	}
}

// TODO: return generator
func parse(s string) []token {
	p := &jsonParser{
		src: []rune(s),
		i:   0,
	}
	return p.parseValue()
}
