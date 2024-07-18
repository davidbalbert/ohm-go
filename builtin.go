package ohm

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type Grammar struct {
	super *Grammar
	rules map[string]PExpr
}

func (g *Grammar) MatchesRule(name, input string) (bool, error) {
	// TODO: allow matching rules with args
	body := &Seq{[]PExpr{&Apply{name: name}, &Apply{name: "end"}}}
	root := call{}

	state := &MatchState{
		g:     g,
		input: input,
		pos:   0,
		stack: []call{root},
	}

	return state.eval(body)
}

type call struct {
	app *Apply
	pos int
}

type MatchState struct {
	g     *Grammar
	input string
	pos   int
	stack []call
}

func (m *MatchState) eval(expr PExpr) (bool, error) {
	pos := m.pos

	res, err := expr.Eval(m)
	if err != nil {
		return false, err
	}

	if !res {
		m.pos = pos
		return false, nil
	}
	return true, nil
}

type PExpr interface {
	Eval(m *MatchState) (bool, error)
}

type Any struct{}

func (*Any) Eval(m *MatchState) (bool, error) {
	if m.pos >= len(m.input) {
		return false, nil
	}

	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		return false, fmt.Errorf("invalid rune at pos %d", m.pos)
	}

	m.pos += size
	return true, nil
}

type Terminal struct {
	s string
}

func (t *Terminal) Eval(m *MatchState) (bool, error) {
	if strings.HasPrefix(m.input[m.pos:], t.s) {
		m.pos += len(t.s)
		return true, nil
	}
	return false, nil
}

type Param struct {
	idx int
}

func (p *Param) Eval(m *MatchState) (bool, error) {
	call := m.stack[len(m.stack)-1]
	return m.eval(call.app.args[p.idx])
}

type Chars struct {
	runes []rune
}

func (c *Chars) Eval(m *MatchState) (bool, error) {
	if m.pos >= len(m.input) {
		return false, nil
	}

	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		return false, fmt.Errorf("invalid rune at pos %d", m.pos)
	}

	for _, rune := range c.runes {
		if r == rune {
			m.pos += size
			return true, nil
		}
	}

	return false, nil
}

type Range struct {
	start rune
	end   rune
}

func (r *Range) Eval(m *MatchState) (bool, error) {
	if m.pos >= len(m.input) {
		return false, nil
	}

	actual, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if actual == utf8.RuneError {
		return false, fmt.Errorf("invalid rune at pos %d", m.pos)
	}

	if actual < r.start || actual > r.end {
		return false, nil
	}

	m.pos += size
	return true, nil
}

type Alt struct {
	exprs []PExpr
}

func (a *Alt) Eval(m *MatchState) (bool, error) {
	for _, expr := range a.exprs {
		res, err := m.eval(expr)
		if err != nil {
			return false, err
		}
		if res {
			return true, nil
		}
	}

	return false, nil
}

type Seq struct {
	exprs []PExpr
}

func (s *Seq) Eval(m *MatchState) (bool, error) {
	for _, expr := range s.exprs {
		res, err := m.eval(expr)
		if err != nil {
			return false, err
		}
		if !res {
			return false, nil
		}
	}

	return true, nil
}

type Opt struct {
	expr PExpr
}

func (o *Opt) Eval(m *MatchState) (bool, error) {
	_, err := m.eval(o.expr)
	if err != nil {
		return false, err
	}
	return true, nil
}

type Star struct {
	expr PExpr
}

func (s *Star) Eval(m *MatchState) (bool, error) {
	for {
		res, err := m.eval(s.expr)
		if err != nil {
			return false, err
		}
		if !res {
			break
		}
	}
	return true, nil
}

type Plus struct {
	expr PExpr
}

func (p *Plus) Eval(m *MatchState) (bool, error) {
	res, err := m.eval(p.expr)
	if err != nil || !res {
		return res, err
	}

	for {
		res, err := m.eval(p.expr)
		if err != nil {
			return false, err
		}
		if !res {
			break
		}
	}
	return true, nil
}

type Apply struct {
	name string
	args []PExpr
}

func (a *Apply) Eval(m *MatchState) (bool, error) {
	m.stack = append(m.stack, call{app: a, pos: m.pos})
	defer func() {
		m.stack = m.stack[:len(m.stack)-1]
	}()

	g := m.g
	for g != nil {
		expr := g.rules[a.name]
		if expr != nil {
			return m.eval(expr)
		}

		g = g.super
	}
	return false, fmt.Errorf("unknown rule \"%s\"", a.name)
}

type Lookahead struct {
	expr PExpr
}

func (l *Lookahead) Eval(m *MatchState) (bool, error) {
	pos := m.pos
	defer func() { m.pos = pos }()
	return m.eval(l.expr)
}

type Not struct {
	expr PExpr
}

func (n *Not) Eval(m *MatchState) (bool, error) {
	pos := m.pos
	defer func() { m.pos = pos }()
	res, err := m.eval(n.expr)
	if err != nil {
		return false, err
	}
	return !res, nil
}

type ucType int

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

func (c *UnicodeCategories) Eval(m *MatchState) (bool, error) {
	if m.pos >= len(m.input) {
		return false, nil
	}

	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		return false, fmt.Errorf("invalid rune at pos %d", m.pos)
	}

	// Special case lower and upper so we can use Go's IsLower and IsUpper functions
	// which have optimizations for ASCII.
	switch c.kind {
	case ucTypeLower:
		if !unicode.IsLower(r) {
			return false, nil
		}
	case ucTypeUpper:
		if !unicode.IsUpper(r) {
			return false, nil
		}
	case ucTypeRanges:
		if !unicode.In(r, c.ranges...) {
			return false, nil
		}
	}

	m.pos += size
	return true, nil
}

var lower UnicodeCategories = UnicodeCategories{kind: ucTypeLower}
var upper UnicodeCategories = UnicodeCategories{kind: ucTypeUpper}
var ltmo UnicodeCategories = UnicodeCategories{
	kind:   ucTypeRanges,
	ranges: []*unicode.RangeTable{unicode.Lt, unicode.Lm, unicode.Lo},
	names:  []string{"Lt", "Lm", "Lo"},
}

var primitiveRules Grammar = Grammar{
	super: nil,
	rules: map[string]PExpr{
		"any":         &Any{},
		"lower":       &lower,
		"upper":       &upper,
		"unicodeLtmo": &ltmo,
	},
}

// This will be generated from built-in-rules.ohm

var BuiltInRules Grammar = Grammar{
	super: &primitiveRules,
	rules: map[string]PExpr{
		"alnum":    &Alt{[]PExpr{&Apply{name: "letter"}, &Apply{name: "digit"}}},
		"letter":   &Alt{[]PExpr{&Apply{name: "lower"}, &Apply{name: "upper"}, &Apply{name: "unicodeLtmo"}}},
		"digit":    &Range{'0', '9'},
		"hexDigit": &Alt{[]PExpr{&Apply{name: "digit"}, &Range{'a', 'f'}, &Range{'A', 'F'}}},
		"end":      &Not{&Any{}},
		"spaces":   &Star{&Apply{name: "space"}},
		"space":    &Chars{[]rune(" \t\n\r")},
	},
}
