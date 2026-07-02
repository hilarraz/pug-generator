package slug

import "testing"

func TestOf(t *testing.T) {
	cases := map[string]string{
		"D.Va":              "dva",
		"Lúcio":             "lucio",
		"Torbjörn":          "torbjorn",
		"Wrecking Ball":     "wreckingball",
		"Soldier: 76":       "soldier76",
		"Junker Queen":      "junkerqueen",
		"King's Row":        "kingsrow",
		"Lijiang Tower":     "lijiangtower",
		"Wuxing University": "wuxinguniversity",
	}
	for in, want := range cases {
		if got := Of(in); got != want {
			t.Errorf("Of(%q) = %q, want %q", in, got, want)
		}
	}
}
