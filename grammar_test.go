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

func TestTwoEmptyGrammars(t *testing.T) {
	res, err := OhmGrammar.MatchesRule("Grammars", `
		G1 {
		}
		G2 {
		}
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !res {
		t.Errorf("expected=true actual=false")
	}
}

func TestGrammarWithSuperGrammar(t *testing.T) {
	res, err := OhmGrammar.MatchesRule("Grammars", `
		G1 <: G2 {
		}
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !res {
		t.Errorf("expected=true actual=false")
	}
}

func TestOhmGrammar(t *testing.T) {
	res, err := OhmGrammar.MatchesRule("Grammars", `
		Ohm {

			Grammars
				= Grammar*

			Grammar
				= ident SuperGrammar? "{" Rule* "}"

			SuperGrammar
				= "<:" ident

			Rule
				= ident Formals? ruleDescr? "="  RuleBody  -- define
				| ident Formals?            ":=" OverrideRuleBody  -- override
				| ident Formals?            "+=" RuleBody  -- extend

			RuleBody
				= "|"? NonemptyListOf<TopLevelTerm, "|">

			TopLevelTerm
				= Seq caseName  -- inline
				| Seq

			OverrideRuleBody
				= "|"? NonemptyListOf<OverrideTopLevelTerm, "|">

			OverrideTopLevelTerm
				= "..."  -- superSplice
				| TopLevelTerm

			Formals
				= "<" ListOf<ident, ","> ">"

			Params
				= "<" ListOf<Seq, ","> ">"

			Alt
				= NonemptyListOf<Seq, "|">

			Seq
				= Iter*

			Iter
				= Pred "*"  -- star
				| Pred "+"  -- plus
				| Pred "?"  -- opt
				| Pred

			Pred
				= "~" Lex  -- not
				| "&" Lex  -- lookahead
				| Lex

			Lex
				= "#" Base  -- lex
				| Base

			Base
				= ident Params? ~(ruleDescr? "=" | ":=" | "+=")  -- application
				| oneCharTerminal ".." oneCharTerminal           -- range
				| terminal                                       -- terminal
				| "(" Alt ")"                                    -- paren

			ruleDescr  (a rule description)
				= "(" ruleDescrText ")"

			ruleDescrText
				= (~")" any)*

			caseName
				= "--" (~"\n" space)* name (~"\n" space)* ("\n" | &"}")

			name  (a name)
				= nameFirst nameRest*

			nameFirst
				= "_"
				| letter

			nameRest
				= "_"
				| alnum

			ident  (an identifier)
				= name

			terminal
				= "\"" terminalChar* "\""

			oneCharTerminal
				= "\"" terminalChar "\""

			terminalChar
				= escapeChar
				| ~"\\" ~"\"" ~"\n" "\u{0}".."\u{10FFFF}"

			escapeChar  (an escape sequence)
				= "\\\\"                                     -- backslash
				| "\\\""                                     -- doubleQuote
				| "\\\'"                                     -- singleQuote
				| "\\b"                                      -- backspace
				| "\\n"                                      -- lineFeed
				| "\\r"                                      -- carriageReturn
				| "\\t"                                      -- tab
				| "\\u{" hexDigit hexDigit? hexDigit?
						hexDigit? hexDigit? hexDigit? "}"   -- unicodeCodePoint
				| "\\u" hexDigit hexDigit hexDigit hexDigit  -- unicodeEscape
				| "\\x" hexDigit hexDigit                    -- hexEscape

			space
			+= comment

			comment
				= "//" (~"\n" any)* &("\n" | end)  -- singleLine
				| "/*" (~"*/" any)* "*/"  -- multiLine

			tokens = token*

			token = caseName | comment | ident | operator | punctuation | terminal | any

			operator = "<:" | "=" | ":=" | "+=" | "*" | "+" | "?" | "~" | "&"

			punctuation = "<" | ">" | "," | "--"
		}
	`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if !res {
		t.Errorf("expected=true actual=false")
	}
}
