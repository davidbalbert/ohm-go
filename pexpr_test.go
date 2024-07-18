package ohm

import "testing"

func grammar(rules map[string]PExpr) Grammar {
	return Grammar{
		super: &BuiltInRules,
		rules: rules,
	}
}

func lit(s string) PExpr {
	seq := Seq{}
	for _, r := range s {
		seq.exprs = append(seq.exprs, &Range{r, r})
	}
	return &seq
}

func seq(exprs ...PExpr) PExpr {
	return &Seq{exprs}
}

func alt(exprs ...PExpr) PExpr {
	return &Alt{exprs}
}

type test struct {
	input   string
	matches bool
}

func TestLiteral(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": lit("foo"),
	})

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

func TestLexSeq(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": seq(lit("foo"), lit("bar")),
	})

	tests := []test{
		{"foobar", true},
		{"foo bar", false},
		{"fooba", false},
		{"foobarr", false},
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

func TestSyntacticSeq(t *testing.T) {
	g := grammar(map[string]PExpr{
		"Start": seq(lit("foo"), lit("bar")),
	})

	tests := []test{
		{"foobar", true},
		{"foo bar", true},
		{"fooba", false},
		{"foobarr", false},
	}

	for _, test := range tests {
		res, err := g.MatchesRule("Start", test.input)
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

func TestLexAlt(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": alt(lit("foo"), lit("bar")),
	})

	tests := []test{
		{"foo", true},
		{"bar", true},
		{" foo ", false},
		{"foobar", false},
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

func TestSyntacticAlt(t *testing.T) {
	g := Grammar{
		super: &BuiltInRules,
		rules: map[string]PExpr{
			"Start": alt(lit("foo"), lit("bar")),
		},
	}

	tests := []test{
		{"foo", true},
		{"bar", true},
		{" foo ", true},
		{"foobar", false},
	}

	for _, test := range tests {
		res, err := g.MatchesRule("Start", test.input)
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
