package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/rprtr258/fun"
)

var ErrUsage = errors.New("usage")

var usage = func() string {
	v := reflect.ValueOf(keyMap)
	fields := fun.Map[key.Binding](
		func(_ reflect.StructField, i int) key.Binding {
			return v.Field(i).Interface().(key.Binding)
		},
		reflect.VisibleFields(v.Type())...,
	)

	keyMapInfo := lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(fun.Map[string](
			func(b key.Binding, _ int) string {
				return or(b.Help().Key, strings.Join(b.Keys(), ", ")) + "    "
			}, fields...), "\n"),

		strings.Join(fun.Map[string](
			func(b key.Binding, _ int) string {
				return b.Help().Desc
			}, fields...), "\n"),
	))

	return fmt.Sprintf(`fx terminal JSON viewer
Usage: fx [FILENAME] [SELECTOR]

Examples:
  fx data.json          # view JSON
  fx data.json .field   # view JSON field
  curl ... | fx         # view JSON from curl

Flags:
  -h, --help            print help
  --themes              print themes
  -r, --raw             treat input as a raw string
  -s, --slurp           read all inputs into an array

Key bindings:
%v`,
		keyMapInfo,
	)
}()
