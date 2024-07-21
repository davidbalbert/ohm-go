package ohm

import "testing"

func TestNoGrammars(t *testing.T) {
	res, err := OhmGrammar.MatchesRule("Grammars", `
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !res {
		t.Errorf("expected=true actual=false")
	}
}

func TestEmptyGrammar(t *testing.T) {
	res, err := OhmGrammar.MatchesRule("Grammars", `
		G1 {}
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !res {
		t.Errorf("expected=true actual=false")
	}
}
