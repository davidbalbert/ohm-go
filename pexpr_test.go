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
		seq.exprs = append(seq.exprs, &Char{r})
	}
	return &seq
}

func seq(exprs ...PExpr) PExpr {
	return &Seq{exprs}
}

func alt(exprs ...PExpr) PExpr {
	return &Alt{exprs}
}

func maybe(expr PExpr) PExpr {
	return &Maybe{expr}
}

func apply(name string, args ...PExpr) PExpr {
	return &Apply{name, args}
}

func param(n int) PExpr {
	return &Param{n}
}

func chars(rs ...rune) PExpr {
	return &Chars{rs}
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
		"start": seq(lit("aa"), maybe(lit("bb")), lit("cc")),
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
		"Start": seq(lit("aa"), maybe(lit("bb")), lit("cc")),
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

func TestRange(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": &Range{'b', 'd'},
	})

	tests := []test{
		{"a", false},
		{"b", true},
		{"c", true},
		{"d", true},
		{"e", false},
		{"bb", false},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestChars(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": chars('a', 'c', 'e'),
	})

	tests := []test{
		{"a", true},
		{"b", false},
		{"c", true},
		{"d", false},
		{"e", true},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestLookahead(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": seq(&Lookahead{lit("fo")}, lit("foo")),
	})

	tests := []test{
		{"fo", false},
		{"foo", true},
		{"fooo", false},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestNot(t *testing.T) {
	g := grammar(map[string]PExpr{
		"while": seq(lit("while"), &Not{&Range{'a', 'z'}}),
	})

	tests := []test{
		{"while", true},
		{"while1", false},
		{"whilea", false},
		{"whil", false},
	}
	testMatchesRule(t, g, "while", tests)
}

func TestStar(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": seq(&Star{lit("a")}, lit("b")),
	})

	tests := []test{
		{"b", true},
		{"ab", true},
		{"aab", true},
		{"aaab", true},
		{"aaaab", true},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestPlus(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": seq(&Plus{lit("a")}, lit("b")),
	})

	tests := []test{
		{"b", false},
		{"ab", true},
		{"aab", true},
		{"aaab", true},
		{"aaaab", true},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestMaybe(t *testing.T) {
	g := grammar(map[string]PExpr{
		"start": seq(&Maybe{lit("a")}, lit("b")),
	})

	tests := []test{
		{"b", true},
		{"ab", true},
		{"aab", false},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestApply(t *testing.T) {
	g := grammar(map[string]PExpr{
		"Start": seq(apply("foo"), lit("bar")),
		"foo":   lit("foo"),
	})

	tests := []test{
		{"foobar", true},
		{"foo bar", true},
		{"fooba", false},
		{"foobarr", false},
	}
	testMatchesRule(t, g, "Start", tests)
}

func TestLexApplyWithArgs(t *testing.T) {
	g := grammar(map[string]PExpr{
		// commaListOf<elem> = nonemptyCommaListOf<elem> | emptyCommaListOf<elem>
		// nonemptyCommaListOf<elem> = elem ("," elem)*
		// emptyCommaListOf<elem> = /* nothing */
		"start":             apply("commaList", lit("a")),
		"commaList":         alt(apply("nonemptyCommaList", param(0)), apply("emptyCommaList", param(0))),
		"nonemptyCommaList": seq(param(0), &Star{seq(lit(","), param(0))}),
		"emptyCommaList":    &Seq{},
	})

	tests := []test{
		{"", true},
		{"a", true},
		{"a,a", true},
		{"a,a,a", true},
		{"a,", false},
		{",a", false},
		{"a,b", false},
		{"a, a", false},
	}
	testMatchesRule(t, g, "start", tests)
}

func TestSyntacticApplyWithArgs(t *testing.T) {
	g := grammar(map[string]PExpr{
		"Start":             apply("CommaList", lit("a")),
		"CommaList":         alt(apply("NonemptyCommaList", param(0)), apply("EmptyCommaList", param(0))),
		"NonemptyCommaList": seq(param(0), &Star{seq(lit(","), param(0))}),
		"EmptyCommaList":    &Seq{},
	})

	tests := []test{
		{"", true},
		{"a", true},
		{"a,a", true},
		{"a,a,a", true},
		{"a, a", true},
		{"a, a, a", true},
		{"a,", false},
		{",a", false},
		{"a, b", false},
	}
	testMatchesRule(t, g, "Start", tests)
}

// TODO:
// - Param
// - left recursion
