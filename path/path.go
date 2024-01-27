package path

import (
	"regexp"
	"strconv"
	"unicode"
)

type state int

const (
	stateStart state = iota
	stateUnknown
	statePropOrIndex
	stateProp
	stateIndex
	statetIndexEnd
	stateNumber
	stateDoubleQuote
	stateDoubleQuoteEscape
	stateSingleQuote
	stateSingleQuoteEscape
)

func Split(p string) ([]any, bool) {
	path := make([]any, 0)
	s := ""
	state := stateStart
	for _, ch := range p {
		switch state {
		case stateStart:
			switch {
			case ch == 'x':
				state = stateUnknown
			case ch == '.':
				state = statePropOrIndex
			case ch == '[':
				state = stateIndex
			default:
				return path, false
			}

		case stateUnknown:
			switch {
			case ch == '.':
				state = stateProp
				s = ""
			case ch == '[':
				state = stateIndex
				s = ""
			default:
				return path, false
			}

		case statePropOrIndex:
			switch {
			case isProp(ch):
				state = stateProp
				s = string(ch)
			case ch == '[':
				state = stateIndex
			default:
				return path, false
			}

		case stateProp:
			switch {
			case isProp(ch):
				s += string(ch)
			case ch == '.':
				state = stateProp
				path = append(path, s)
				s = ""
			case ch == '[':
				state = stateIndex
				path = append(path, s)
				s = ""
			default:
				return path, false
			}

		case stateIndex:
			switch {
			case unicode.IsDigit(ch):
				state = stateNumber
				s = string(ch)
			case ch == '"':
				state = stateDoubleQuote
				s = ""
			case ch == '\'':
				state = stateSingleQuote
				s = ""
			default:
				return path, false
			}

		case statetIndexEnd:
			switch {
			case ch == ']':
				state = stateUnknown
			default:
				return path, false
			}

		case stateNumber:
			switch {
			case unicode.IsDigit(ch):
				s += string(ch)
			case ch == ']':
				state = stateUnknown
				n, err := strconv.Atoi(s)
				if err != nil {
					return path, false
				}
				path = append(path, n)
				s = ""
			default:
				return path, false
			}

		case stateDoubleQuote:
			switch ch {
			case '"':
				state = statetIndexEnd
				path = append(path, s)
				s = ""
			case '\\':
				state = stateDoubleQuoteEscape
			default:
				s += string(ch)
			}

		case stateDoubleQuoteEscape:
			switch ch {
			case '"':
				state = stateDoubleQuote
				s += string(ch)
			default:
				return path, false
			}

		case stateSingleQuote:
			switch ch {
			case '\'':
				state = statetIndexEnd
				path = append(path, s)
				s = ""
			case '\\':
				state = stateSingleQuoteEscape
				s += string(ch)
			default:
				s += string(ch)
			}

		case stateSingleQuoteEscape:
			switch ch {
			case '\'':
				state = stateSingleQuote
				s += string(ch)
			default:
				return path, false
			}
		}
	}
	if s != "" && state == stateProp {
		path = append(path, s)
	}
	return path, s == "" || state == stateProp
}

func isProp(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$'
}

var Identifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func Join(path []any) string {
	s := ""
	for _, v := range path {
		switch v := v.(type) {
		case string:
			if Identifier.MatchString(v) {
				s += "." + v
			} else {
				s += "[" + strconv.Quote(v) + "]"
			}
		case int:
			s += "[" + strconv.Itoa(v) + "]"
		}
	}
	return s
}
