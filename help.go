package main

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

var ErrUsage = errors.New("usage")

var usage = func() string {
	v := reflect.ValueOf(keyMap)
	fields := lo.Map(
		reflect.VisibleFields(v.Type()),
		func(_ reflect.StructField, i int) key.Binding {
			return v.Field(i).Interface().(key.Binding)
		})

	keyMapInfo := lipgloss.NewStyle().PaddingLeft(2).Render(lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(lo.Map(
			fields,
			func(b key.Binding, _ int) string {
				return or(b.Help().Key, strings.Join(b.Keys(), ", ")) + "    "
			}), "\n"),
		strings.Join(lo.Map(
			fields,
			func(b key.Binding, _ int) string {
				return b.Help().Desc
			}), "\n"),
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
