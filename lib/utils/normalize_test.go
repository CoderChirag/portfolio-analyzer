package utils

import "testing"

func TestNormalizeFundName(t *testing.T) {
	cases := []struct{ in, want string }{
		// formatting differences collapse
		{"SBI Contra Fund - Direct Plan - Growth", "sbi contra direct growth"},
		{"SBI Contra Direct Plan Growth", "sbi contra direct growth"},
		{"  sbi   CONTRA  direct growth ", "sbi contra direct growth"},

		// abbreviations and letter/digit boundaries
		{"Nippon India Nifty50 Index Dir Gr", "nippon india nifty 50 index direct growth"},
		{"Nippon India Index Fund Nifty 50 Plan Direct Growth", "nippon india index nifty 50 direct growth"},

		// "&" and "and" are the same token
		{"ICICI Prudential Banking & PSU Debt", "icici prudential banking & psu debt"},
		{"ICICI Prudential Banking and PSU Debt", "icici prudential banking & psu debt"},

		// dividend == idcw
		{"HDFC Balanced Advantage Dividend Payout", "hdfc balanced advantage idcw payout"},
		{"HDFC Balanced Advantage IDCW Payout", "hdfc balanced advantage idcw payout"},

		// accents fold to ASCII
		{"Fondé Équité Growth", "fonde equite growth"},

		// nothing meaningful left
		{"Fund - Plan", ""},
		{"", ""},
	}

	for _, c := range cases {
		if got := NormalizeFundName(c.in, 255); got != c.want {
			t.Errorf("NormalizeFundName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestNormalizeFundNameKeepsShareClassesDistinct(t *testing.T) {
	distinct := []string{
		"Tata Small Cap Fund Direct Growth",
		"Tata Small Cap Fund Regular Growth",
		"Tata Small Cap Fund Direct IDCW",
		"Tata Small Cap Fund Direct IDCW Reinvestment",
	}
	seen := map[string]string{}
	for _, name := range distinct {
		k := NormalizeFundName(name, 255)
		if prev, dup := seen[k]; dup {
			t.Fatalf("%q and %q both normalize to %q", prev, name, k)
		}
		seen[k] = name
	}
}
