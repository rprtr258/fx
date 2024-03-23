package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/pprof"
	"strconv"
	"strings"

	"github.com/antonmedv/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
	"github.com/rprtr258/fun"
	"github.com/rprtr258/scuf"
	"github.com/sahilm/fuzzy"

	jsonpath "github.com/antonmedv/fx/path"
)

//go:embed index.js
var src []byte

func reFindAllStringIndex(re *regexp.Regexp, s string) [][2]int {
	var res [][2]int
	for _, v := range re.FindAllStringIndex(s, -1) {
		res = append(res, [2]int(v[:2]))
	}
	return res
}

func regexCase(code string) (string, bool) {
	switch {
	case strings.HasSuffix(code, "/i"):
		return code[:len(code)-2], true
	case strings.HasSuffix(code, "/"):
		return code[:len(code)-1], false
	default:
		return code, true
	}
}

func reduce(fns []string) error {
	script := filepath.Join(os.TempDir(), "fx.js")
	if _, err := os.Stat(script); os.IsNotExist(err) {
		if err := os.WriteFile(script, src, 0o644); err != nil {
			return err
		}
	}

	args := []string{script}
	envVar := "NODE_OPTIONS=--max-old-space-size=16384"
	bin, err := exec.LookPath("node")
	if err != nil {
		bin, err = exec.LookPath("deno")
		if err != nil {
			return errors.New("Node.js or Deno is required to run fx with reducers")
		}

		args = []string{"run", "-A", script}
		envVar = "V8_FLAGS=--max-old-space-size=16384"
	}

	cmd := exec.Command(bin, append(args, fns...)...)
	cmd.Env = append(os.Environ(), envVar)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func flex(width int, a, b string) string {
	return a + strings.Repeat(" ", max(1, width-len(a)-len(b))) + b
}

type model struct {
	termWidth, termHeight int
	head, top             *node
	cursor                int // cursor position [0, termHeight)
	fileName              string
	digInput              textinput.Model
	searchInput           textinput.Model
	search                *search
	showCursor            bool
	wrap                  bool
	yank                  bool
}

func (m *model) Init() tea.Cmd {
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		wrapAll(m.top, m.termWidth)
		m.redoSearch()

	case tea.MouseMsg:
		switch {
		case msg.Button&tea.MouseButtonWheelUp != 0:
			m.up()

		case msg.Button&tea.MouseButtonWheelDown != 0:
			m.down()

		case msg.Button&tea.MouseButtonLeft != 0:
			m.digInput.Blur()
			m.showCursor = true
			if msg.Y >= m.viewHeight() {
				break
			}

			if m.cursor == msg.Y {
				to := m.cursorPointsTo()
				if to == nil {
					break
				}

				if to.isCollapsed() {
					to.expand()
				} else {
					to.collapse()
				}
			} else {
				to := m.at(msg.Y)
				if to == nil {
					break
				}

				m.cursor = msg.Y
				if to.isCollapsed() {
					to.expand()
				}
			}
		}

	case tea.KeyMsg:
		if m.digInput.Focused() {
			return m, m.handleDigKey(msg)
		}
		if m.searchInput.Focused() {
			return m, m.handleSearchKey(msg)
		}
		if m.yank {
			m.handleYankKey(msg)
			return m, nil
		}
		return m, m.handleKey(msg)
	}
	return m, nil
}

func (m *model) handleDigKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case key.Matches(msg, arrowUp):
		m.up()
		m.digInput.SetValue(m.cursorPath())
		m.digInput.CursorEnd()

	case key.Matches(msg, arrowDown):
		m.down()
		m.digInput.SetValue(m.cursorPath())
		m.digInput.CursorEnd()

	case msg.Type == tea.KeyEscape:
		m.digInput.Blur()

	case msg.Type == tea.KeyTab:
		m.digInput.SetValue(m.cursorPath())
		m.digInput.CursorEnd()

	case msg.Type == tea.KeyEnter:
		m.digInput.Blur()
		if digPath, ok := jsonpath.Split(m.digInput.Value()); ok {
			if n := m.selectByPath(digPath); n != nil {
				m.selectNode(n)
			}
		}

	case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+w"))):
		if digPath, ok := jsonpath.Split(m.digInput.Value()); ok {
			if len(digPath) > 0 {
				digPath = digPath[:len(digPath)-1]
			}
			if n := m.selectByPath(digPath); n != nil {
				m.selectNode(n)
				m.digInput.SetValue(m.cursorPath())
				m.digInput.CursorEnd()
			}
		}

	case key.Matches(msg, textinput.DefaultKeyMap.WordBackward):
		value := m.digInput.Value()
		if path, ok := jsonpath.Split(value[0:m.digInput.Position()]); ok {
			if len(path) > 0 {
				path = path[:len(path)-1]
				m.digInput.SetCursor(len(jsonpath.Join(path)))
			} else {
				m.digInput.CursorStart()
			}
		}

	case key.Matches(msg, textinput.DefaultKeyMap.WordForward):
		value := m.digInput.Value()
		fullPath, ok1 := jsonpath.Split(value)
		if path, ok2 := jsonpath.Split(value[0:m.digInput.Position()]); ok1 && ok2 {
			if len(path) < len(fullPath) {
				path = append(path, fullPath[len(path)])
				m.digInput.SetCursor(len(jsonpath.Join(path)))
			} else {
				m.digInput.CursorEnd()
			}
		}

	default:
		if key.Matches(msg, key.NewBinding(key.WithKeys("."))) {
			if m.digInput.Position() == len(m.digInput.Value()) {
				m.digInput.SetValue(m.cursorPath())
				m.digInput.CursorEnd()
			}
		}

		m.digInput, cmd = m.digInput.Update(msg)
		if n := m.dig(m.digInput.Value()); n != nil {
			m.selectNode(n)
		}
	}
	return cmd
}

func (m *model) handleSearchKey(msg tea.KeyMsg) tea.Cmd {
	var cmd tea.Cmd
	switch {
	case msg.Type == tea.KeyEscape:
		m.searchInput.Blur()
		m.searchInput.SetValue("")
		m.doSearch("")
		m.showCursor = true

	case msg.Type == tea.KeyEnter:
		m.searchInput.Blur()
		m.doSearch(m.searchInput.Value())

	default:
		m.searchInput, cmd = m.searchInput.Update(msg)
	}
	return cmd
}

func (m *model) handleYankKey(msg tea.KeyMsg) {
	switch {
	case key.Matches(msg, yankPath):
		_ = clipboard.WriteAll(m.cursorPath())
	case key.Matches(msg, yankKey):
		_ = clipboard.WriteAll(m.cursorKey())
	case key.Matches(msg, yankValue):
		_ = clipboard.WriteAll(m.cursorValue())
	}
	m.yank = false
}

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, keyMap.Quit):
		return tea.Quit

	case key.Matches(msg, keyMap.Up):
		m.up()

	case key.Matches(msg, keyMap.Down):
		m.down()

	case key.Matches(msg, keyMap.PageUp):
		m.cursor = 0
		for i := 0; i < m.viewHeight(); i++ {
			m.up()
		}

	case key.Matches(msg, keyMap.PageDown):
		m.cursor = m.viewHeight() - 1
		for i := 0; i < m.viewHeight(); i++ {
			m.down()
		}
		m.scrollIntoView()

	case key.Matches(msg, keyMap.HalfPageUp):
		m.cursor = 0
		for i := 0; i < m.viewHeight()/2; i++ {
			m.up()
		}

	case key.Matches(msg, keyMap.HalfPageDown):
		m.cursor = m.viewHeight() - 1
		for i := 0; i < m.viewHeight()/2; i++ {
			m.down()
		}
		m.scrollIntoView()

	case key.Matches(msg, keyMap.GotoTop):
		m.head = m.top
		m.cursor = 0
		m.showCursor = true

	case key.Matches(msg, keyMap.GotoBottom):
		m.head = m.findBottom()
		m.cursor = 0
		m.showCursor = true
		m.scrollIntoView()

	case key.Matches(msg, keyMap.NextSibling):
		var nextSibling *node
		if pointsTo := m.cursorPointsTo(); pointsTo.end != nil && pointsTo.end.next != nil {
			nextSibling = pointsTo.end.next
		} else {
			nextSibling = pointsTo.next
		}
		if nextSibling != nil {
			m.selectNode(nextSibling)
		}

	case key.Matches(msg, keyMap.PrevSibling):
		pointsTo := m.cursorPointsTo()
		var prevSibling *node
		if pointsTo.parent() != nil && pointsTo.parent().end == pointsTo {
			prevSibling = pointsTo.parent()
		} else if pointsTo.prev != nil {
			prevSibling = pointsTo.prev
			parent := prevSibling.parent()
			if parent != nil && parent.end == prevSibling {
				prevSibling = parent
			}
		}
		if prevSibling != nil {
			m.selectNode(prevSibling)
		}

	case key.Matches(msg, keyMap.Collapse):
		n := m.cursorPointsTo()
		if n.hasChildren() && !n.isCollapsed() {
			n.collapse()
		} else if n.parent() != nil {
			n = n.parent()
		}
		m.selectNode(n)

	case key.Matches(msg, keyMap.Expand):
		m.cursorPointsTo().expand()
		m.showCursor = true

	case key.Matches(msg, keyMap.CollapseRecursively):
		if n := m.cursorPointsTo(); n.hasChildren() {
			n.collapseRecursively()
		}
		m.showCursor = true

	case key.Matches(msg, keyMap.ExpandRecursively):
		if n := m.cursorPointsTo(); n.hasChildren() {
			n.expandRecursively()
		}
		m.showCursor = true

	case key.Matches(msg, keyMap.CollapseAll):
		for n := m.top; n != nil; n = fun.Deref(n.end).next {
			n.collapseRecursively()
		}
		m.cursor = 0
		m.head = m.top
		m.showCursor = true

	case key.Matches(msg, keyMap.ExpandAll):
		at := m.cursorPointsTo()
		for n := m.top; n != nil; n = fun.Deref(n.end).next {
			n.expandRecursively()
		}
		m.selectNode(at)

	case key.Matches(msg, keyMap.ToggleWrap):
		at := m.cursorPointsTo()
		m.wrap = !m.wrap
		if m.wrap {
			wrapAll(m.top, m.termWidth)
		} else {
			dropWrapAll(m.top)
		}
		if at.chunk != nil && at.value == nil {
			at = at.parent()
		}
		m.redoSearch()
		m.selectNode(at)

	case key.Matches(msg, keyMap.Yank):
		m.yank = true

	case key.Matches(msg, keyMap.Dig):
		m.digInput.SetValue(m.cursorPath() + ".")
		m.digInput.CursorEnd()
		m.digInput.Width = m.termWidth - 1
		m.digInput.Focus()

	case key.Matches(msg, keyMap.Search):
		m.searchInput.CursorEnd()
		m.searchInput.Width = m.termWidth - 2 // -1 for the prompt, -1 for the cursor
		m.searchInput.Focus()

	case key.Matches(msg, keyMap.SearchNext):
		m.selectSearchResult(m.search.cursor + 1)

	case key.Matches(msg, keyMap.SearchPrev):
		m.selectSearchResult(m.search.cursor - 1)
	}
	return nil
}

func (m *model) up() {
	m.showCursor = true
	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
		if m.head.prev != nil {
			m.head = m.head.prev
		}
	}
}

func (m *model) down() {
	m.showCursor = true

	if m.cursorPointsTo() == nil {
		return
	}

	m.cursor++
	if m.cursor >= m.viewHeight() {
		m.cursor = m.viewHeight() - 1
		if m.head.next != nil {
			m.head = m.head.next
		}
	}
}

func (m *model) scrollIntoView() {
	visibleLines := 0
	for n := m.head; n != nil && visibleLines < m.viewHeight(); n = n.next {
		visibleLines++
	}

	m.cursor = min(m.cursor, visibleLines-1)
	for visibleLines < m.viewHeight() && m.head.prev != nil {
		visibleLines++
		m.cursor++
		m.head = m.head.prev
	}
}

func (m *model) View() string {
	screen := ""
	lineNumber := 0
	for n := m.head; n != nil && lineNumber < m.viewHeight(); n = n.next {
		screen += strings.Repeat("  ", n.depth)

		// don't highlight the cursor while iterating search results
		isSelected := m.cursor == lineNumber && m.showCursor

		if n.key != nil {
			screen += m.prettyKey(n, isSelected)
			screen += colon
			isSelected = false // don't highlight the key's value
		}

		screen += m.prettyPrint(n, isSelected)

		if n.isCollapsed() {
			if fun.Deref(n.value)[0] == '{' {
				if n.collapsed.key != nil {
					screen += scuf.String(*n.collapsed.key, currentTheme.Preview)
					screen += colonPreview
				}
				screen += dot3
				screen += closeCurlyBracket
			} else if fun.Deref(n.value)[0] == '[' {
				screen += dot3
				screen += closeSquareBracket
			}
			if n.end != nil && n.end.comma {
				screen += comma
			}
		}
		if n.comma {
			screen += comma
		}

		screen += "\n"
		lineNumber++
	}

	screen += strings.Repeat(empty+"\n", max(0, m.viewHeight()-lineNumber))

	if m.digInput.Focused() {
		screen += m.digInput.View()
	} else {
		statusBar := flex(m.termWidth, m.cursorPath(), m.fileName)
		screen += scuf.String(statusBar, currentTheme.StatusBar)
	}

	switch {
	case m.yank:
		screen += "\n"
		screen += "(y)value  (p)path  (k)key"
	case m.searchInput.Focused():
		screen += "\n"
		screen += m.searchInput.View()
	case m.searchInput.Value() != "":
		var msg string
		switch {
		case m.search.err != nil:
			msg = m.search.err.Error()
		case len(m.search.results) == 0:
			msg = "not found"
		default:
			msg = fmt.Sprintf("found: [%v/%v]", m.search.cursor+1, len(m.search.results))
		}

		re, ci := regexCase(m.searchInput.Value())

		screen += "\n"
		screen += flex(m.termWidth, "/"+re+"/"+fun.IF(ci, "i", ""), msg)
	}

	return screen
}

func (m *model) prettyKey(node *node, selected bool) string {
	b := node.key
	style := fun.IF(selected, currentTheme.Cursor, currentTheme.Key)

	indexes, ok := m.search.keys[node]
	if !ok {
		return scuf.String(*b, style)
	}

	out := ""
	for i, p := range splitBytesByIndexes(b, indexes) {
		switch {
		case i%2 == 0:
			out += scuf.String(string(p.b), style)
		case p.index == m.search.cursor:
			out += scuf.String(string(p.b), currentTheme.Cursor)
		default:
			out += scuf.String(string(p.b), currentTheme.Search)
		}
	}
	return out
}

func (m *model) prettyPrint(node *node, selected bool) string {
	b := fun.IF(node.chunk != nil, node.chunk, node.value)
	if fun.Deref(b) == "" {
		return ""
	}

	style := valueStyle(b, selected, node.chunk != nil)

	if indexes, ok := m.search.values[node]; ok {
		out := ""
		for i, p := range splitBytesByIndexes(b, indexes) {
			var toadd string
			switch {
			case i%2 == 0:
				toadd = scuf.String(string(p.b), style)
			case p.index == m.search.cursor:
				toadd = scuf.String(string(p.b), currentTheme.Cursor)
			default:
				toadd = scuf.String(string(p.b), currentTheme.Search)
			}
			out += toadd
		}
		return out
	}

	return scuf.String(fun.Deref(b), style)
}

func (m *model) viewHeight() int {
	if m.searchInput.Focused() || m.searchInput.Value() != "" {
		return m.termHeight - 2
	}
	if m.yank {
		return m.termHeight - 2
	}
	return m.termHeight - 1
}

func (m *model) cursorPointsTo() *node {
	return m.at(m.cursor)
}

func (m *model) at(pos int) *node {
	head := m.head
	for i := 0; i < pos; i++ {
		if head == nil {
			break
		}
		head = head.next
	}
	return head
}

func (m *model) findBottom() *node {
	n := m.head
	for n.next != nil {
		if n.end != nil {
			n = n.end
		} else {
			n = n.next
		}
	}
	return n
}

func (m *model) nodeInsideView(n *node) bool {
	if n == nil {
		return false
	}
	head := m.head
	for i := 0; i < m.viewHeight(); i++ {
		if head == nil {
			break
		}
		if head == n {
			return true
		}
		head = head.next
	}
	return false
}

func (m *model) selectNodeInView(n *node) {
	head := m.head
	for i := 0; i < m.viewHeight(); i++ {
		if head == nil {
			break
		}

		if head == n {
			m.cursor = i
			return
		}

		head = head.next
	}
}

func (m *model) selectNode(n *node) {
	m.showCursor = true
	if m.nodeInsideView(n) {
		m.selectNodeInView(n)
		m.scrollIntoView()
	} else {
		m.cursor = 0
		m.head = n
		m.scrollIntoView()
	}
	parent := n.parent()
	for parent != nil {
		parent.expand()
		parent = parent.parent()
	}
}

func (m *model) cursorPath() string {
	path := ""
	for at := m.cursorPointsTo(); at != nil; at = at.parent() {
		if at.prev == nil {
			continue
		}

		if at.chunk != nil && at.value == nil {
			at = at.parent()
		}
		if at.key != nil {
			quoted := *at.key
			unquoted, err := strconv.Unquote(quoted)
			if err == nil && jsonpath.Identifier.MatchString(unquoted) {
				path = "." + unquoted + path
			} else {
				path = "[" + quoted + "]" + path
			}
		} else if at.index >= 0 {
			path = "[" + strconv.Itoa(at.index) + "]" + path
		}
	}
	return path
}

func (m *model) cursorValue() string {
	at := m.cursorPointsTo()
	if at == nil {
		return ""
	}
	var out strings.Builder
	if at.chunk != nil && at.value == nil {
		at = at.parent()
	}
	out.WriteString(fun.Deref(at.value))
	if at.hasChildren() {
		it := at.next
		if at.isCollapsed() {
			it = at.collapsed
		}
		for it != nil {
			if it.key != nil {
				out.WriteString(*it.key)
				out.WriteString(": ")
			}
			if it.chunk != nil {
				out.WriteString(*it.chunk)
			} else {
				out.WriteString(fun.Deref(it.value))
			}
			if it == at.end {
				break
			}
			if it.comma {
				out.WriteString(", ")
			}
			if it.isCollapsed() {
				it = it.collapsed
			} else {
				it = it.next
			}
		}
	}
	return out.String()
}

func (m *model) cursorKey() string {
	at := m.cursorPointsTo()
	if at == nil {
		return ""
	}
	if at.key != nil {
		var v string
		_ = json.Unmarshal([]byte(*at.key), &v)
		return v
	}
	return strconv.Itoa(at.index)
}

func (m *model) selectByPath(path []any) *node {
	n := m.currentTopNode()
	for _, part := range path {
		if n == nil {
			return nil
		}
		switch part := part.(type) {
		case string:
			n = n.findChildByKey(part)
		case int:
			n = n.findChildByIndex(part)
		}
	}
	return n
}

func (m *model) currentTopNode() *node {
	at := m.cursorPointsTo()
	if at == nil {
		return nil
	}
	for at.parent() != nil {
		at = at.parent()
	}
	return at
}

func (m *model) doSearch(s string) {
	m.search = newSearch()

	if s == "" {
		return
	}

	code, ci := regexCase(s)
	if ci {
		code = "(?i)" + code
	}

	re, err := regexp.Compile(code)
	if err != nil {
		m.search.err = err
		return
	}

	searchIndex := 0
	for n := m.top; n != nil; n = fun.IF(n.isCollapsed(), n.collapsed, n.next) {
		if n.key != nil {
			if indexes := reFindAllStringIndex(re, *n.key); len(indexes) > 0 {
				for i, pair := range indexes {
					m.search.results = append(m.search.results, n)
					m.search.keys[n] = append(m.search.keys[n], match{
						start: pair[0],
						end:   pair[1],
						index: searchIndex + i,
					})
				}
				searchIndex += len(indexes)
			}
		}

		if indexes := reFindAllStringIndex(re, fun.Deref(n.value)); len(indexes) > 0 {
			for range indexes {
				m.search.results = append(m.search.results, n)
			}
			if n.chunk != nil {
				// String can be split into chunks, so we need to map the indexes to the chunks.
				chunkNodes := []*node{n}
				chunks := [][]byte{[]byte(fun.Deref(n.chunk))}
				for it := n.next; it != nil; it = it.next {
					chunkNodes = append(chunkNodes, it)
					chunks = append(chunks, []byte(fun.Deref(it.chunk)))
					if it == n.chunkEnd {
						break
					}
				}

				chunkMatches := splitIndexesToChunks(chunks, indexes, searchIndex)
				for i, matches := range chunkMatches {
					m.search.values[chunkNodes[i]] = matches
				}
			} else {
				for i, pair := range indexes {
					m.search.values[n] = append(m.search.values[n], match{
						start: pair[0],
						end:   pair[1],
						index: searchIndex + i,
					})
				}
			}
			searchIndex += len(indexes)
		}
	}

	m.selectSearchResult(0)
}

func (m *model) selectSearchResult(i int) {
	if len(m.search.results) == 0 {
		return
	}

	if i < 0 {
		i = len(m.search.results) - 1
	}
	if i >= len(m.search.results) {
		i = 0
	}

	m.search.cursor = i
	result := m.search.results[i]
	m.selectNode(result)
	m.showCursor = false
}

func (m *model) redoSearch() {
	if m.searchInput.Value() == "" || len(m.search.results) == 0 {
		return
	}

	cursor := m.search.cursor
	m.doSearch(m.searchInput.Value())
	m.selectSearchResult(cursor)
}

func (m *model) dig(v string) *node {
	p, ok := jsonpath.Split(v)
	if !ok {
		return nil
	}

	at := m.selectByPath(p)
	if at != nil {
		return at
	}

	searchTerm, ok := p[len(p)-1].(string)
	if !ok {
		return nil
	}

	at = m.selectByPath(p[:len(p)-1])
	if at == nil {
		return nil
	}

	keys, nodes := at.children()

	matches := fuzzy.Find(searchTerm, keys)
	if len(matches) == 0 {
		return nil
	}

	return nodes[matches[0].Index]
}

func run() error {
	if _, ok := os.LookupEnv("FX_PPROF"); ok {
		f, err := os.Create("cpu.prof")
		if err != nil {
			return err
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			return err
		}
		defer pprof.StopCPUProfile()

		memProf, err := os.Create("mem.prof")
		if err != nil {
			return err
		}
		defer pprof.WriteHeapProfile(memProf)
	}

	var args []string
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			return ErrUsage
		case "--themes":
			themeTester()
			return nil
		default:
			args = append(args, arg)
		}
	}

	var fileName string
	var src io.Reader
	switch stdinIsTty := isatty.IsTerminal(os.Stdin.Fd()); {
	case stdinIsTty && len(args) == 0:
		return ErrUsage
	case stdinIsTty && len(args) == 1:
		filePath := args[0]
		f, err := os.Open(filePath)
		if err != nil {
			return err
		}
		fileName = filepath.Base(filePath)
		src = f
	case !stdinIsTty && len(args) == 0:
		src = os.Stdin
	default:
		return reduce(args)
	}

	data, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	head := nodeparse(string(data), parse(string(data)))

	digInput := textinput.New()
	digInput.Prompt = ""
	digInput.TextStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("7")).
		Foreground(lipgloss.Color("0"))
	digInput.Cursor.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("15")).
		Foreground(lipgloss.Color("0"))

	searchInput := textinput.New()
	searchInput.Prompt = "/"

	_, err = tea.NewProgram(
		&model{
			head:        head,
			top:         head,
			showCursor:  true,
			wrap:        true,
			fileName:    fileName,
			digInput:    digInput,
			searchInput: searchInput,
			search:      newSearch(),
		},
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	).Run()
	return err
}

func main() {
	if err := run(); err != nil {
		if err == ErrUsage {
			fmt.Println(usage)
			return
		}

		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
