package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/muesli/termenv"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/scuf"
)

type theme struct {
	Cursor    scuf.Modifier
	Syntax    scuf.Modifier
	Preview   scuf.Modifier
	StatusBar scuf.Modifier
	Search    scuf.Modifier
	Key       scuf.Modifier
	String    scuf.Modifier
	Null      scuf.Modifier
	Bool      scuf.Modifier
	Number    scuf.Modifier
}

func isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func isWhitespace(ch rune) bool {
	return fun.Contains(ch, ' ', '\t', '\n', '\r')
}

func valueStyle(bb *string, selected, chunk bool) scuf.Modifier {
	if selected {
		return currentTheme.Cursor
	}

	if chunk {
		return currentTheme.String
	}

	b := []byte(*bb)
	if isDigit(rune(b[0])) || b[0] == '-' {
		return currentTheme.Number
	}

	switch b[0] {
	case '"':
		return currentTheme.String
	case 't', 'f':
		return currentTheme.Bool
	case 'n':
		return currentTheme.Null
	case '{', '[', '}', ']':
		return currentTheme.Syntax
	default:
		return nil
	}
}

func or(s1, s2 string) string {
	if s1 != "" {
		return s1
	}
	return s2
}

func readCurrentTheme() theme {
	if termenv.ColorProfile() == termenv.Ascii {
		return themes["0"]
	}

	themeID := or(os.Getenv("FX_THEME"), "1")

	currentTheme, ok := themes[themeID]
	if !ok {
		panic(fmt.Sprintf("fx: unknown theme %q, available themes: %v", themeID, themeNames))
	}

	return currentTheme
}

var (
	themeNames = func() []string {
		themeNames := fun.Keys(themes)
		sort.Strings(themeNames)
		return themeNames
	}()
	currentTheme     = readCurrentTheme()
	defaultCursor    = scuf.ModReverse
	defaultPreview   = scuf.Combine(scuf.FgANSI(8))
	defaultStatusBar = scuf.Combine(scuf.BgANSI(7), scuf.FgANSI(0))
	defaultSearch    = scuf.Combine(scuf.BgANSI(11), scuf.FgANSI(16))
	defaultNull      = scuf.FgANSI(243)
)

var (
	colon              = scuf.String(": ", currentTheme.Syntax)
	colonPreview       = scuf.String(":", currentTheme.Preview)
	comma              = scuf.String(",", currentTheme.Syntax)
	empty              = scuf.String("~", currentTheme.Preview)
	dot3               = scuf.String("…", currentTheme.Preview)
	closeCurlyBracket  = scuf.String("}", currentTheme.Syntax)
	closeSquareBracket = scuf.String("]", currentTheme.Syntax)
)

var themes = map[string]theme{
	"0": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   nil,
		StatusBar: nil,
		Search:    defaultSearch,
		Key:       nil,
		String:    nil,
		Null:      nil,
		Bool:      nil,
		Number:    nil,
	},
	"1": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       bold(scuf.FgANSI(4)),
		String:    scuf.FgANSI(2),
		Null:      defaultNull,
		Bool:      scuf.FgANSI(5),
		Number:    scuf.FgANSI(6),
	},
	"2": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       scuf.FgANSI(2),
		String:    scuf.FgANSI(4),
		Null:      defaultNull,
		Bool:      scuf.FgANSI(5),
		Number:    scuf.FgANSI(6),
	},
	"3": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       scuf.FgANSI(13),
		String:    scuf.FgANSI(11),
		Null:      defaultNull,
		Bool:      scuf.FgANSI(1),
		Number:    scuf.FgANSI(14),
	},
	"4": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       scuf.FgRGB(scuf.MustParseHexRGB("#00F5D4")),
		String:    scuf.FgRGB(scuf.MustParseHexRGB("#00BBF9")),
		Null:      defaultNull,
		Bool:      scuf.FgRGB(scuf.MustParseHexRGB("#F15BB5")),
		Number:    scuf.FgRGB(scuf.MustParseHexRGB("#9B5DE5")),
	},
	"5": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       scuf.FgRGB(scuf.MustParseHexRGB("#faf0ca")),
		String:    scuf.FgRGB(scuf.MustParseHexRGB("#f4d35e")),
		Null:      defaultNull,
		Bool:      scuf.FgRGB(scuf.MustParseHexRGB("#ee964b")),
		Number:    scuf.FgRGB(scuf.MustParseHexRGB("#ee964b")),
	},
	"6": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       scuf.FgRGB(scuf.MustParseHexRGB("#4D96FF")),
		String:    scuf.FgRGB(scuf.MustParseHexRGB("#6BCB77")),
		Null:      defaultNull,
		Bool:      scuf.FgRGB(scuf.MustParseHexRGB("#FF6B6B")),
		Number:    scuf.FgRGB(scuf.MustParseHexRGB("#FFD93D")),
	},
	"7": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       bold(scuf.FgANSI(42)),
		String:    bold(scuf.FgANSI(213)),
		Null:      defaultNull,
		Bool:      bold(scuf.FgANSI(201)),
		Number:    bold(scuf.FgANSI(201)),
	},
	"8": {
		Cursor:    defaultCursor,
		Syntax:    nil,
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       bold(scuf.FgANSI(51)),
		String:    scuf.FgANSI(195),
		Null:      defaultNull,
		Bool:      scuf.FgANSI(50),
		Number:    scuf.FgANSI(123),
	},
	"🔵": {
		Cursor:    scuf.Combine(scuf.FgANSI(15), scuf.BgANSI(33)),
		Syntax:    bold(scuf.FgANSI(33)),
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       scuf.FgANSI(33),
		String:    nil,
		Null:      nil,
		Bool:      nil,
		Number:    nil,
	},
	"🥝": {
		Cursor:    defaultCursor,
		Syntax:    scuf.FgANSI(179),
		Preview:   defaultPreview,
		StatusBar: defaultStatusBar,
		Search:    defaultSearch,
		Key:       bold(scuf.FgANSI(154)),
		String:    scuf.FgANSI(82),
		Null:      scuf.FgANSI(230),
		Bool:      scuf.FgANSI(226),
		Number:    scuf.FgANSI(226),
	},
}

func bold(bg scuf.Modifier) scuf.Modifier {
	return scuf.Combine(scuf.ModBold, bg)
}

func themeTester() {
	for _, name := range themeNames {
		t := themes[name]
		comma := scuf.String(",", t.Syntax)
		colon := scuf.String(":", t.Syntax)

		fmt.Println(scuf.String(fmt.Sprintf("Theme %q", name), scuf.ModBold))
		fmt.Println(scuf.String("{", t.Syntax))
		for _, kv := range []struct {
			key   string
			color scuf.Modifier
			value string
		}{
			{`"string"`, t.String, `"Fox jumps over the lazy dog"`},
			{`"number"`, t.Number, "1234567890"},
			{`"boolean"`, t.Bool, "true"},
			{`"null"`, t.Null, "null"},
		} {
			fmt.Printf("  %v%v %v%v\n",
				scuf.String(kv.key, t.Key),
				colon,
				scuf.String(kv.value, kv.color),
				comma)
		}
		fmt.Println(scuf.String("}", t.Syntax))
		fmt.Println()
	}
}
