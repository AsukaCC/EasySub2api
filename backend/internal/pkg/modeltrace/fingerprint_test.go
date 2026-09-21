package modeltrace

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"testing"
)

// Golden probabilities were generated with the pinned upstream JavaScript scorer
// against public GPT, Claude and mixed reference outputs, not the Go port.
func TestUpstreamParity(t *testing.T) {
	body, err := os.ReadFile("testdata/parity.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Name     string
		Outputs  []string
		Expected []Attribution
	}
	if err := json.Unmarshal(body, &cases); err != nil {
		t.Fatal(err)
	}
	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			var samples [][]int
			for _, output := range tc.Outputs {
				samples = append(samples, ParseNumbers(output))
			}
			result, err := Analyze(samples)
			if err != nil {
				t.Fatal(err)
			}
			if len(result.Models) != len(tc.Expected) {
				t.Fatal("candidate count mismatch")
			}
			total := 0.0
			for i, got := range result.Models {
				want := tc.Expected[i]
				if got.Model != want.Model || math.Abs(got.Probability-want.Probability) > 1e-10 {
					t.Errorf("got %+v, want %+v", got, want)
				}
				total += got.Probability
			}
			if math.Abs(total-1) > 1e-12 {
				t.Fatal("probabilities do not sum to one")
			}
			familyTotal := 0.0
			for _, family := range result.Families {
				familyTotal += family.Probability
			}
			if math.Abs(familyTotal-1) > 1e-12 {
				t.Fatal("family probabilities do not sum to one")
			}
		})
	}
}

func TestParserAndValidation(t *testing.T) {
	for _, tc := range []struct {
		text string
		want []int
	}{
		{"requested 300 numbers: 12, 42, 355, 0, 356", []int{12, 42, 355}},
		{"1,2,3 abc 4 5", []int{1, 2, 3}},
		{"12 个数字：4，5，6", []int{4, 5, 6}},
		{"No numeric output", nil},
	} {
		if got := ParseNumbers(tc.text); !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%q: got %v want %v", tc.text, got, tc.want)
		}
	}
	if _, err := Analyze([][]int{{1}, {2}, {3}}); err == nil {
		t.Fatal("accepted insufficient numbers")
	}
	if _, err := Analyze(nil); err == nil {
		t.Fatal("accepted incomplete samples")
	}
	if MinimumNumbers(300) != 165 {
		t.Fatal("minimum diverged from upstream")
	}
	seen := map[int]bool{}
	for _, challenge := range Challenges() {
		if seen[challenge.Expected] || challenge.Expected < 292 || challenge.Expected > 332 {
			t.Fatal("invalid challenge length")
		}
		seen[challenge.Expected] = true
	}
	if len(seen) != 3 {
		t.Fatal("three distinct challenges required")
	}
}
