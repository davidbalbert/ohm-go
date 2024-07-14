package ohm

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type Grammar interface {
	Super() Grammar
	Rule(name string) PExpr
}

type MatchState struct {
	g     Grammar
	input string
	pos   int
	err   error
}

func (m *MatchState) apply(name string) bool {
	g := m.g
	for g != nil {
		expr := g.Rule(name)
		if expr != nil {
			return expr.Eval(m)
		}

		g = g.Super()
	}

	m.err = fmt.Errorf("rule '%s' not found", name)
	return false
}

type PExpr interface {
	Eval(m *MatchState) bool
}

type Any struct{}

func (*Any) Eval(m *MatchState) bool {
	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		m.err = fmt.Errorf("invalid rune at pos %d", m.pos)
		return false
	}

	m.pos += size
	return true
}

type Chars struct {
	runes []rune
}

func (c *Chars) Eval(m *MatchState) bool {
	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		m.err = fmt.Errorf("invalid rune at pos %d", m.pos)
		return false
	}

	for _, rune := range c.runes {
		if r == rune {
			m.pos += size
			return true
		}
	}

	m.err = fmt.Errorf("expected one of '%s', got '%c'", c.runes, r)
	return false
}

type Range struct {
	start rune
	end   rune
}

func (r *Range) Eval(m *MatchState) bool {
	actual, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if actual == utf8.RuneError {
		m.err = fmt.Errorf("invalid rune at pos %d", m.pos)
		return false
	}

	if actual < r.start || actual > r.end {
		m.err = fmt.Errorf("expected '%c'..'%c', got '%c'", r.start, r.end, actual)
		return false
	}

	m.pos += size
	return true
}

type Alt struct {
	exprs []PExpr
}

func (a *Alt) Eval(m *MatchState) bool {
	for _, expr := range a.exprs {
		pos := m.pos
		if expr.Eval(m) {
			return true
		}

		m.pos = pos
	}

	return false
}

type Seq struct {
	exprs []PExpr
}

func (s *Seq) Eval(m *MatchState) bool {
	for _, expr := range s.exprs {
		if !expr.Eval(m) {
			return false
		}
	}

	return true
}

type Opt struct {
	expr PExpr
}

func (o *Opt) Eval(m *MatchState) bool {
	o.expr.Eval(m)
	return true
}

type Star struct {
	expr PExpr
}

func (s *Star) Eval(m *MatchState) bool {
	for s.expr.Eval(m) {
	}

	return true
}

type Plus struct {
	expr PExpr
}

func (p *Plus) Eval(m *MatchState) bool {
	if !p.expr.Eval(m) {
		return false
	}

	for p.expr.Eval(m) {
	}

	return true
}

type Apply struct {
	name string
}

func (a *Apply) Eval(m *MatchState) bool {
	return m.apply(a.name)
}

type Lookahead struct {
	expr PExpr
}

func (l *Lookahead) Eval(m *MatchState) bool {
	pos := m.pos
	defer func() { m.pos = pos }()
	return l.expr.Eval(m)
}

type Not struct {
	expr PExpr
}

func (n *Not) Eval(m *MatchState) bool {
	pos := m.pos
	if n.expr.Eval(m) {
		m.pos = pos
		return false
	}

	return true
}

type ucType int

// Special case lower and upper case so we can use Go's IsLower and IsUpper functions
// which have optimizations for ASCII.
const (
	ucTypeLower ucType = iota
	ucTypeUpper
	ucTypeRanges
)

type UnicodeCategories struct {
	kind   ucType
	ranges []*unicode.RangeTable
	names  []string
}

func (c *UnicodeCategories) Eval(m *MatchState) bool {
	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		m.err = fmt.Errorf("invalid rune at pos %d", m.pos)
		return false
	}

	switch c.kind {
	case ucTypeLower:
		if !unicode.IsLower(r) {
			m.err = fmt.Errorf("expected lower-case letter, got '%c'", r)
			return false
		}
	case ucTypeUpper:
		if !unicode.IsUpper(r) {
			m.err = fmt.Errorf("expected upper-case letter, got '%c'", r)
			return false
		}
	case ucTypeRanges:
		if !unicode.In(r, c.ranges...) {
			m.err = fmt.Errorf("expected character in %v, got '%c'", c.names, r)
			return false
		}
	}

	m.pos += size
	return true
}

var lower UnicodeCategories = UnicodeCategories{kind: ucTypeLower}
var upper UnicodeCategories = UnicodeCategories{kind: ucTypeUpper}
var ltmo UnicodeCategories = UnicodeCategories{
	kind:   ucTypeRanges,
	ranges: []*unicode.RangeTable{unicode.Lt, unicode.Lm, unicode.Lo},
	names:  []string{"Lt", "Lm", "Lo"},
}

type PrimitiveRules struct{}

func (*PrimitiveRules) Super() Grammar {
	return nil
}

func (*PrimitiveRules) Rule(name string) PExpr {
	switch name {
	case "any":
		return &Any{}
	case "lower":
		return &lower
	case "upper":
		return &upper
	case "unicodeLtmo":
		return &ltmo
	}

	return nil
}

// This will be generated from built-in-rules.ohm

type BuiltInRules struct{}

func (*BuiltInRules) Super() Grammar {
	return &PrimitiveRules{}
}

func (g *BuiltInRules) Rule(name string) PExpr {
	switch name {
	case "alnum":
		return g.rule_alnum()
	case "letter":
		return g.rule_letter()
	case "digit":
		return g.rule_digit()
	case "hexDigit":
		return g.rule_hexDigit()
	case "end":
		return g.rule_end()
	case "spaces":
		return g.rule_spaces()
	case "space":
		return g.rule_space()
	}

	return nil
}

func (*BuiltInRules) rule_alnum() PExpr {
	return &Alt{[]PExpr{&Apply{"letter"}, &Apply{"digit"}}}
}

func (*BuiltInRules) rule_letter() PExpr {
	return &Alt{[]PExpr{&Apply{"lower"}, &Apply{"upper"}, &Apply{"unicodeLtmo"}}}
}

func (*BuiltInRules) rule_digit() PExpr {
	return &Alt{[]PExpr{&Range{'0', '9'}}}
}

func (*BuiltInRules) rule_hexDigit() PExpr {
	return &Alt{[]PExpr{&Apply{"digit"}, &Range{'a', 'f'}, &Range{'A', 'F'}}}
}

func (*BuiltInRules) rule_end() PExpr {
	return &Not{&Any{}}
}

func (*BuiltInRules) rule_spaces() PExpr {
	return &Star{&Apply{"space"}}
}

func (*BuiltInRules) rule_space() PExpr {
	return &Chars{[]rune(" \t\n\r")}
}
