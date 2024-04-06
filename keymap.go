package main

import "github.com/rprtr258/tea/components/key"

type KeyMap struct {
	Quit                key.Binding
	PageDown            key.Binding
	PageUp              key.Binding
	HalfPageUp          key.Binding
	HalfPageDown        key.Binding
	GotoTop             key.Binding
	GotoBottom          key.Binding
	Down                key.Binding
	Up                  key.Binding
	Expand              key.Binding
	Collapse            key.Binding
	ExpandRecursively   key.Binding
	CollapseRecursively key.Binding
	ExpandAll           key.Binding
	CollapseAll         key.Binding
	NextSibling         key.Binding
	PrevSibling         key.Binding
	ToggleWrap          key.Binding
	Yank                key.Binding
	Search              key.Binding
	SearchNext          key.Binding
	SearchPrev          key.Binding
	Dig                 key.Binding
}

var keyMap = KeyMap{
	Quit: key.Binding{
		Keys: []string{"q", "ctrl+c", "esc"},
		Help: key.Help{"", "exit program"},
	},
	PageDown: key.Binding{
		Keys: []string{"pgdown", " ", "f"},
		Help: key.Help{"pgdown, space, f", "page down"},
	},
	PageUp: key.Binding{
		Keys: []string{"pgup", "b"},
		Help: key.Help{"pgup, b", "page up"},
	},
	HalfPageUp: key.Binding{
		Keys: []string{"u", "ctrl+u"},
		Help: key.Help{"", "half page up"},
	},
	HalfPageDown: key.Binding{
		Keys: []string{"d", "ctrl+d"},
		Help: key.Help{"", "half page down"},
	},
	GotoTop: key.Binding{
		Keys: []string{"g", "home"},
		Help: key.Help{"", "goto top"},
	},
	GotoBottom: key.Binding{
		Keys: []string{"G", "end"},
		Help: key.Help{"", "goto bottom"},
	},
	Down: key.Binding{
		Keys: []string{"down", "j"},
		Help: key.Help{"", "down"},
	},
	Up: key.Binding{
		Keys: []string{"up", "k"},
		Help: key.Help{"", "up"},
	},
	Expand: key.Binding{
		Keys: []string{"right", "l", "enter"},
		Help: key.Help{"", "expand"},
	},
	Collapse: key.Binding{
		Keys: []string{"left", "h", "backspace"},
		Help: key.Help{"", "collapse"},
	},
	ExpandRecursively: key.Binding{
		Keys: []string{"L", "shift+right"},
		Help: key.Help{"", "expand recursively"},
	},
	CollapseRecursively: key.Binding{
		Keys: []string{"H", "shift+left"},
		Help: key.Help{"", "collapse recursively"},
	},
	ExpandAll: key.Binding{
		Keys: []string{"e"},
		Help: key.Help{"", "expand all"},
	},
	CollapseAll: key.Binding{
		Keys: []string{"E"},
		Help: key.Help{"", "collapse all"},
	},
	NextSibling: key.Binding{
		Keys: []string{"J", "shift+down"},
		Help: key.Help{"", "next sibling"},
	},
	PrevSibling: key.Binding{
		Keys: []string{"K", "shift+up"},
		Help: key.Help{"", "previous sibling"},
	},
	ToggleWrap: key.Binding{
		Keys: []string{"z"},
		Help: key.Help{"", "toggle strings wrap"},
	},
	Yank: key.Binding{
		Keys: []string{"y"},
		Help: key.Help{"", "yank/copy"},
	},
	Search: key.Binding{
		Keys: []string{"/"},
		Help: key.Help{"", "search regexp"},
	},
	SearchNext: key.Binding{
		Keys: []string{"n"},
		Help: key.Help{"", "next search result"},
	},
	SearchPrev: key.Binding{
		Keys: []string{"N"},
		Help: key.Help{"", "prev search result"},
	},
	Dig: key.Binding{
		Keys: []string{"."},
		Help: key.Help{"", "dig json"},
	},
}

var (
	yankValue = key.Binding{Keys: []string{"y"}}
	yankKey   = key.Binding{Keys: []string{"k"}}
	yankPath  = key.Binding{Keys: []string{"p"}}
	arrowUp   = key.Binding{Keys: []string{"up"}}
	arrowDown = key.Binding{Keys: []string{"down"}}
)
