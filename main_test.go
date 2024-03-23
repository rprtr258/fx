package main

import (
	"bytes"
	_ "embed"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

//go:embed testdata/example.json
var _json []byte

func prepare(t *testing.T) *teatest.TestModel {
	t.Helper()

	head := nodeparse(string(_json), parse(string(_json)))
	return teatest.NewTestModel(
		t,
		&model{
			top:         head,
			head:        head,
			wrap:        true,
			showCursor:  true,
			digInput:    textinput.New(),
			searchInput: textinput.New(),
			search:      newSearch(),
		},
		teatest.WithInitialTermSize(80, 40),
	)
}

func read(t *testing.T, tm *teatest.TestModel) []byte {
	t.Helper()

	var out []byte
	teatest.WaitFor(t,
		tm.Output(),
		func(b []byte) bool {
			out = b
			return bytes.Contains(b, []byte("{"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second),
	)
	return out
}

func TestOutput(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestNavigation(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	tm.Send(tea.KeyMsg{Type: tea.KeyDown})
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestDig(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(".")})
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("year")})
	tm.Send(tea.KeyMsg{Type: tea.KeyEnter})
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestCollapseRecursive(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	tm.Send(tea.KeyMsg{Type: tea.KeyShiftLeft})
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
