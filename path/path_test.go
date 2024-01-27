package path_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/antonmedv/fx/path"
)

func Test_SplitPath(t *testing.T) {
	t.Parallel()

	for input, want := range map[string][]any{
		"":                      {},
		".":                     {},
		"x":                     {},
		".foo":                  {"foo"},
		"x.foo":                 {"foo"},
		"x[42]":                 {42},
		".[42]":                 {42},
		".42":                   {"42"},
		".физ":                  {"физ"},
		".foo.bar":              {"foo", "bar"},
		".foo[42]":              {"foo", 42},
		".foo[42].bar":          {"foo", 42, "bar"},
		".foo[1][2]":            {"foo", 1, 2},
		`.foo["bar"]`:           {"foo", "bar"},
		`.foo["bar\""]`:         {"foo", `bar"`},
		".foo['bar']['baz\\'']": {"foo", "bar", "baz\\'"},
		"[42]":                  {42},
		"[42].foo":              {42, "foo"},
	} {
		input := input
		want := want

		t.Run(input, func(t *testing.T) {
			t.Parallel()

			p, ok := path.Split(input)
			require.True(t, ok)
			require.Equal(t, want, p)
		})
	}
}

func Test_SplitPath_negative(t *testing.T) {
	t.Parallel()

	for _, input := range []string{
		"./",
		"x/",
		"1+1",
		"x[42",
		".i % 2",
		"x[for x]",
		"x['y'.",
		"x[0?",
		"x[\"\\u",
		"x['\\n",
		"x[9999999999999999999999999999999999999]",
		"x[]",
	} {
		input := input
		t.Run(input, func(t *testing.T) {
			t.Parallel()

			_, ok := path.Split(input)
			require.False(t, ok)
		})
	}
}

func TestJoin(t *testing.T) {
	t.Parallel()

	for want, input := range map[string][]any{
		"":                        {},
		".foo":                    {"foo"},
		".foo.bar":                {"foo", "bar"},
		".foo[42]":                {"foo", 42},
		".foo.bar[42]":            {"foo", "bar", 42},
		".foo.bar[42].baz":        {"foo", "bar", 42, "baz"},
		".foo.bar[42].baz[1]":     {"foo", "bar", 42, "baz", 1},
		".foo.bar[42].baz[1].qux": {"foo", "bar", 42, "baz", 1, "qux"},
		`["foo bar"]`:             {"foo bar"},
	} {
		input := input
		want := want

		t.Run(want, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, want, path.Join(input))
		})
	}
}
