package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// fundFillerWords carry no identity and are dropped.
// Never add words that distinguish share classes: direct, regular,
// growth, idcw, payout, reinvestment, bonus must survive.
var fundFillerWords = map[string]struct{}{
	"fund":   {},
	"plan":   {},
	"option": {},
	"scheme": {},
	"the":    {},
}

// fundSynonyms maps alternate spellings to one canonical token.
var fundSynonyms = map[string]string{
	"dir":      "direct",
	"reg":      "regular",
	"gr":       "growth",
	"dividend": "idcw", // SEBI renamed "dividend" options to IDCW in 2021
	"div":      "idcw",
	"and":      "&", // "&" survives as its own token, see tokenize
	"funds":    "fund",
}

// asciiFolder strips accents: "é" -> "e".
var asciiFolder = transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// NormalizeFundName returns the matching key stored in
// instrument_aliases.normalized_alias. Two names that differ only in case,
// punctuation, spacing, accents, filler words or known abbreviations produce
// the same key:
//
//	"SBI Contra Fund - Direct Plan - Growth" -> "sbi contra direct growth"
//	"SBI Contra Direct Plan Growth"          -> "sbi contra direct growth"
//	"Nippon India Nifty50 Index Dir Gr"      -> "nippon india nifty 50 index direct growth"
//
// It returns "" if nothing meaningful remains; callers must treat "" as
// "no match" and never look it up or store it as a match key.
func NormalizeFundName(name string, maxNormalizedLen int) string {
	folded, _, err := transform.String(asciiFolder, name)
	if err != nil {
		folded = name
	}

	tokens := Tokenize(strings.ToLower(folded))
	tokens = applySynonyms(tokens)
	tokens = collapseFoF(tokens)
	tokens = dropFillers(tokens)

	return TruncateKey(strings.TrimSpace(strings.Join(tokens, " ")), maxNormalizedLen)
}

func applySynonyms(tokens []string) []string {
	out := make([]string, len(tokens))
	for i, t := range tokens {
		if s, ok := fundSynonyms[t]; ok {
			t = s
		}
		out[i] = t
	}
	return out
}

// collapseFoF turns "fund of fund" into the single token "fof", so a
// fund-of-funds never normalizes to the same key as its underlying fund.
// Must run after applySynonyms ("funds" -> "fund") and before dropFillers
// (which would otherwise remove "fund").
func collapseFoF(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for i := 0; i < len(tokens); i++ {
		if i+2 < len(tokens) && tokens[i] == "fund" && tokens[i+1] == "of" && tokens[i+2] == "fund" {
			out = append(out, "fof")
			i += 2
			continue
		}
		out = append(out, tokens[i])
	}
	return out
}

func dropFillers(tokens []string) []string {
	out := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if _, filler := fundFillerWords[t]; !filler {
			out = append(out, t)
		}
	}
	return out
}
