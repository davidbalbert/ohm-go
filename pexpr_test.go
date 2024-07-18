package ohm

import "testing"

func lit(s string) PExpr {
	seq := Seq{}
	for _, r := range s {
		seq.exprs = append(seq.exprs, &Range{r, r})
	}
	return &seq
}

type test struct {
	input   string
	matches bool
}

func TestLiteral(t *testing.T) {
	g := Grammar{
		super: &BuiltInRules,
		rules: map[string]PExpr{
			"start": lit("foo"),
		},
	}

	tests := []test{
		{"foo", true},
		{"fooo", false},
		{"fo", false},
	}

	for _, test := range tests {
		res, err := g.MatchesRule("start", test.input)
		if err != nil {
			t.Fatal(err)
		}
		if test.matches && !res {
			t.Errorf("\"%s\" should succeed", test.input)
		} else if !test.matches && res {
			t.Errorf("\"%s\" should fail", test.input)
		}
	}
}
