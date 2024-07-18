package ohm

import "testing"

func literal(s string) PExpr {
	seq := Seq{}
	for _, r := range s {
		seq.exprs = append(seq.exprs, &Range{r, r})
	}
	return &seq
}

func TestLiteral(t *testing.T) {
	g := Grammar{
		super: &BuiltInRules,
		rules: map[string]PExpr{
			"start": literal("foo"),
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
