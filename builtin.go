package ohm

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type Grammar struct {
	super *Grammar
	rules map[string]PExpr
}

func (g *Grammar) MatchesRule(name, input string) (bool, error) {
	// TODO: allow matching rules with args
	a := Apply{name: name}
	islex, err := a.isLexical()
	if err != nil {
		return false, err
	}

	body := &Seq{[]PExpr{&Apply{name: name}, &Apply{name: "end"}}}
	root := call{app: &Apply{}, lexical: islex}

	state := &MatchState{
		g:     g,
		input: input,
		pos:   0,
		stack: []call{root},
	}

	return state.eval(body)
}

type call struct {
	app     *Apply
	pos     int
	lexical bool
}

type MatchState struct {
	g     *Grammar
	input string
	pos   int
	stack []call
}

var spaces Apply = Apply{name: "spaces"}

func (m *MatchState) eval(expr PExpr) (bool, error) {
	pos := m.pos

	if !m.stack[len(m.stack)-1].lexical && expr != &spaces {
		m.eval(&spaces)
	}

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
	substituteParams(args []PExpr) (PExpr, error)
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

func (a *Any) substituteParams(args []PExpr) (PExpr, error) {
	return a, nil
}

type Char struct {
	r rune
}

func (c *Char) Eval(m *MatchState) (bool, error) {
	if m.pos >= len(m.input) {
		return false, nil
	}

	r, size := utf8.DecodeRuneInString(m.input[m.pos:])
	if r == utf8.RuneError {
		return false, fmt.Errorf("invalid rune at pos %d", m.pos)
	}

	if r != c.r {
		return false, nil
	}
	m.pos += size
	return true, nil
}

func (c *Char) substituteParams(args []PExpr) (PExpr, error) {
	return c, nil
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

func (c *Chars) substituteParams(args []PExpr) (PExpr, error) {
	return c, nil
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

func (r *Range) substituteParams(args []PExpr) (PExpr, error) {
	return r, nil
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

func (a *Alt) substituteParams(args []PExpr) (PExpr, error) {
	newExprs := make([]PExpr, len(a.exprs))
	for i, expr := range a.exprs {
		newExpr, err := expr.substituteParams(args)
		if err != nil {
			return nil, err
		}
		newExprs[i] = newExpr
	}
	return &Alt{newExprs}, nil
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

func (s *Seq) substituteParams(args []PExpr) (PExpr, error) {
	newExprs := make([]PExpr, len(s.exprs))
	for i, expr := range s.exprs {
		newExpr, err := expr.substituteParams(args)
		if err != nil {
			return nil, err
		}
		newExprs[i] = newExpr
	}
	return &Seq{newExprs}, nil
}

type Maybe struct {
	expr PExpr
}

func (o *Maybe) Eval(m *MatchState) (bool, error) {
	_, err := m.eval(o.expr)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (o *Maybe) substituteParams(args []PExpr) (PExpr, error) {
	newExpr, err := o.expr.substituteParams(args)
	if err != nil {
		return nil, err
	}
	return &Maybe{newExpr}, nil
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

func (s *Star) substituteParams(args []PExpr) (PExpr, error) {
	newExpr, err := s.expr.substituteParams(args)
	if err != nil {
		return nil, err
	}
	return &Star{newExpr}, nil
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

func (p *Plus) substituteParams(args []PExpr) (PExpr, error) {
	newExpr, err := p.expr.substituteParams(args)
	if err != nil {
		return nil, err
	}
	return &Plus{newExpr}, nil
}

type Apply struct {
	name string
	args []PExpr
}

func (a *Apply) Eval(m *MatchState) (bool, error) {
	islex, err := a.isLexical()
	if err != nil {
		return false, err
	}

	caller := m.stack[len(m.stack)-1]
	app, err := a.substituteParams(caller.app.args)
	if err != nil {
		return false, err
	}

	m.stack = append(m.stack, call{app: app.(*Apply), pos: m.pos, lexical: islex})

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

func (a *Apply) substituteParams(args []PExpr) (PExpr, error) {
	newArgs := make([]PExpr, len(a.args))
	for i, arg := range a.args {
		newArg, err := arg.substituteParams(args)
		if err != nil {
			return nil, err
		}
		newArgs[i] = newArg
	}
	return &Apply{a.name, newArgs}, nil
}

func (a *Apply) isLexical() (bool, error) {
	r, _ := utf8.DecodeRuneInString(a.name)
	if r == utf8.RuneError {
		return false, fmt.Errorf("invalid rule name \"%s\"", a.name)
	}

	return unicode.IsLower(r), nil
}

type Param struct {
	idx int
}

func (p *Param) Eval(m *MatchState) (bool, error) {
	call := m.stack[len(m.stack)-1]
	if p.idx >= len(call.app.args) {
		return false, fmt.Errorf("param index out of range: %d", p.idx)
	}
	return m.eval(call.app.args[p.idx])
}

func (p *Param) substituteParams(args []PExpr) (PExpr, error) {
	if p.idx >= len(args) {
		return nil, fmt.Errorf("param index out of range: %d", p.idx)
	}
	return args[p.idx], nil
}

type Lookahead struct {
	expr PExpr
}

func (l *Lookahead) Eval(m *MatchState) (bool, error) {
	pos := m.pos
	defer func() { m.pos = pos }()
	return m.eval(l.expr)
}

func (l *Lookahead) substituteParams(args []PExpr) (PExpr, error) {
	newExpr, err := l.expr.substituteParams(args)
	if err != nil {
		return nil, err
	}
	return &Lookahead{newExpr}, nil
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

func (n *Not) substituteParams(args []PExpr) (PExpr, error) {
	newExpr, err := n.expr.substituteParams(args)
	if err != nil {
		return nil, err
	}
	return &Not{newExpr}, nil
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

func (c *UnicodeCategories) substituteParams(args []PExpr) (PExpr, error) {
	return c, nil
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
		"ListOf": &Alt{[]PExpr{
			&Apply{"NonemptyListOf", []PExpr{&Param{0}, &Param{1}}},
			&Apply{"EmptyListOf", []PExpr{&Param{0}, &Param{1}}},
		}},
		"NonemptyListOf": &Seq{[]PExpr{&Param{0}, &Star{&Seq{[]PExpr{&Param{1}, &Param{0}}}}}},
		"EmptyListOf":    &Seq{},
		"listOf": &Alt{[]PExpr{
			&Apply{"nonemptyListOf", []PExpr{&Param{0}, &Param{1}}},
			&Apply{name: "emptyListOf", args: []PExpr{&Param{0}, &Param{1}}},
		}},
		"nonemptyListOf": &Seq{[]PExpr{&Param{0}, &Star{&Seq{[]PExpr{&Param{1}, &Param{0}}}}}},
		"emptyListOf":    &Seq{},
		"end":            &Not{&Any{}},
		"spaces":         &Star{&Apply{name: "space"}},
		"space":          &Chars{[]rune(" \t\n\r")},
	},
}
