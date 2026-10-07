package acquisitionuniverse

import (
	"strings"
	"testing"
)

func TestComposite(t *testing.T) {
	full := Company{ID: "a", MandateFit: 100, OutreachSuitability: 100, TransitionContext: 100}
	if got := Composite(full, DefaultWeights); got != 100 {
		t.Fatalf("got %v", got)
	}
	owned := Company{MandateFit: 100, OutreachSuitability: 100, GroupOwned: true}
	if got := Composite(owned, DefaultWeights); got != 70 {
		t.Fatalf("got %v", got)
	}
}

func TestRankCoverageBrief(t *testing.T) {
	r := Rank([]Company{{ID: "a", MandateFit: 50}, {ID: "b", MandateFit: 90, OutreachSuitability: 50}}, DefaultWeights)
	if r[0].Company.ID != "b" {
		t.Fatal("wrong order")
	}
	v, c := SignalCoverage(map[string]string{Signals[0]: "x", Signals[1]: ""})
	if v != 1 || c != 2 {
		t.Fatalf("coverage %d/%d", v, c)
	}
	b := ThesisBrief{Vertical: "Water", Regions: []string{"US"}}
	if len(b.Validate()) != 1 || !strings.Contains(b.Text(), "free-pilot.php") {
		t.Fatal("brief")
	}
}
