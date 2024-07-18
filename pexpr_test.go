package ohm

import "testing"

func grammar(rules map[string]PExpr) *Grammar {
	return &Grammar{
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

func opt(expr PExpr) PExpr {
	return &Opt{expr}
}

type test struct {
	input   string
	matches bool
}

func testMatchesRule(t *testing.T, g *Grammar, rule string, tests []test) {
	t.Helper()

	for _, test := range tests {
		res, err := g.MatchesRule(rule, test.input)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if test.matches != res {
			t.Errorf("input=\"%s\" expected=%v actual=%v", test.input, test.matches, res)
		}
	}
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
	testMatchesRule(t, g, "start", tests)
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
	testMatchesRule(t, g, "start", tests)
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
	testMatchesRule(t, g, "Start", tests)
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
	testMatchesRule(t, g, "start", tests)
}

func TestSyntacticAlt(t *testing.T) {
	g := grammar(map[string]PExpr{
		"Start": alt(lit("foo"), lit("bar")),
	})

	tests := []test{
		{"foo", true},
		{"bar", true},
		{" foo ", true},
		{"foobar", false},
	}
	testMatchesRule(t, g, "Start", tests)
}

func TestLexOpt(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": seq(lit("aa"), opt(lit("bb")), lit("cc")),
	})

	tests := []test{
		{"aacc", true},
		{"aabbcc", true},
		{"aa bb cc", false},
		{"aa cc", false},
		{"aabcc", false},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestSyntacticOpt(t *testing.T) {
	g := grammar(map[string]PExpr{
		"Start": seq(lit("aa"), opt(lit("bb")), lit("cc")),
	})

	tests := []test{
		{"aacc", true},
		{"aabbcc", true},
		{"aa bb cc", true},
		{"aa cc", true},
		{"aabcc", false},
	}
	testMatchesRule(t, g, "Start", tests)
}

func TestAny(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": &Any{},
	})

	tests := []test{
		{"a", true},
		{"", false},
		{"ab", false}, // matching adds an implicit "end"
	}
	testMatchesRule(t, g, "start", tests)
}
