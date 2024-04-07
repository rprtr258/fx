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

	digInput := textinput.New()
	digInput.SetValue(".")

	return teatest.NewTestModelFixture(
		t,
		&model{
			tree:       hierachy.New(fromJSON(_original)),
			original:   _original,
			result:     _original,
			queryError: "",
			digInput:   digInput,
		},
		teatest.WithInitialTermSize(80, 40),
	)
}

func Test(t *testing.T) {
	t.Parallel()

	for name, keys := range map[string][]tea.MsgKey{
		"Output": nil,
		"Navigation": {
			{Type: tea.KeyDown},
			{Type: tea.KeyDown},
			{Type: tea.KeyDown},
		},
		"Dig": {
			{Type: tea.KeyRunes, Runes: []rune(".")},
			{Type: tea.KeyRunes, Runes: []rune("year")},
			{Type: tea.KeyEnter},
		},
		"CollapseRecursive": {
			{Type: tea.KeyShiftLeft},
		},
	} {
		keys := keys
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tm := prepare(t)

			for _, key := range keys {
				tm.Send(key)
			}

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
			teatest.RequireEqualOutput(t, out)

			tm.Send(tea.MsgKey{Type: tea.KeyRunes, Runes: []rune("q")})
			tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
		})
	}
}
