package ohm

import "testing"

// TODO: get rid of Terminal and have a sequence of chars.
func TestTerminal(t *testing.T) {
	g := Grammar{
		super: &BuiltInRules,
		rules: map[string]PExpr{
			"start": &Terminal{"foo"},
		},
	}

	matches, err := g.MatchesRule("start", "foo")
	if err != nil {
		t.Fatal(err)
	}
	if !matches {
		t.Errorf("\"foo\" should succeed")
	}
}
