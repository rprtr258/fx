package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	"github.com/rprtr258/tea"
	"github.com/rprtr258/tea/components/headless/hierachy"
	"github.com/rprtr258/tea/components/textinput"
	"github.com/rprtr258/tea/teatest"
)

//go:embed testdata/example.json
var _json []byte

var _original = func() any {
	var v any
	if err := json.Unmarshal(_json, &v); err != nil {
		panic(err.Error())
	}
	return v
}()

func prepare(t *testing.T) *teatest.TestModel[*model] {
	t.Helper()

	return teatest.NewTestModelFixture(
		t,
		&model{
			tree:       hierachy.New(fromJSON(_original)),
			original:   _original,
			result:     _original,
			queryError: "",
			digInput:   textinput.New(),
		},
		teatest.WithInitialTermSize(80, 40),
	)
}

func read(t *testing.T, tm *teatest.TestModel[*model]) []byte {
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

	tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestNavigation(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	tm.Send(tea.MsgKey{Type: tea.KeyDown})
	tm.Send(tea.MsgKey{Type: tea.KeyDown})
	tm.Send(tea.MsgKey{Type: tea.KeyDown})
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestDig(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune(".")})
	tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune("year")})
	tm.Send(tea.MsgKey{Type: tea.KeyEnter})
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}

func TestCollapseRecursive(t *testing.T) {
	t.Parallel()

	tm := prepare(t)

	tm.Send(tea.MsgKey{Type: tea.KeyShiftLeft})
	teatest.RequireEqualOutput(t, read(t, tm))

	tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune("q")})
	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
