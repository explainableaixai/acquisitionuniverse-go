// Package acquisitionuniverse is a local toolkit for acquisition target screening:
// thesis briefs, a weighted composite score, ranking and signal coverage.
// It makes no network calls.
package acquisitionuniverse

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// PilotURL is where a screening pilot is requested.
const PilotURL = "https://www.acquisitionuniverse.com/free-pilot.php"

// Signals are the 15 common signals, each recorded with a verbatim quote and a source URL, or "not visible".
var Signals = [15]string{
	"Founder or family association",
	"Visible leadership bench depth",
	"Operating history and continued independence",
	"Strategic fit to thesis",
	"Geographic and branch footprint",
	"Service-led vs product-led model",
	"Recurring-offering indicators",
	"Vertical specialization and end-market exposure",
	"Acquisition-program or roll-up readiness",
	"Management professionalization",
	"Hiring posture and functional investment",
	"Website and news activity trajectory",
	"Partner and channel ecosystem position",
	"Compliance and regulated-market readiness",
	"Digital-commercial maturity",
}

// Weights of the three composite scores.
type Weights struct{ MandateFit, OutreachSuitability, TransitionContext float64 }

// DefaultWeights are 70, 20 and 10 percent.
var DefaultWeights = Weights{0.7, 0.2, 0.1}

// ThesisBrief describes what a screening should look for.
type ThesisBrief struct {
	Name         string
	Vertical     string
	Subverticals []string
	Regions      []string
	EmployeesMin *int
	EmployeesMax *int
	MustHave     []string
	Exclusions   []string
	Notes        string
}

func orDefault(v []string, sep, empty string) string {
	if len(v) == 0 {
		return empty
	}
	return strings.Join(v, sep)
}

// Validate lists problems that would make the brief ambiguous to a screening team.
func (b ThesisBrief) Validate() []string {
	var out []string
	if b.Vertical == "" {
		out = append(out, "vertical is empty")
	}
	if len(b.Regions) == 0 {
		out = append(out, "no regions listed")
	}
	if b.EmployeesMin != nil && b.EmployeesMax != nil && *b.EmployeesMin > *b.EmployeesMax {
		out = append(out, "EmployeesMin is above EmployeesMax")
	}
	if len(b.Exclusions) == 0 {
		out = append(out, "no exclusions listed, group-owned companies are the usual first one")
	}
	return out
}

// Text renders a plain text brief, ready to paste into a pilot request.
func (b ThesisBrief) Text() string {
	size := "not specified"
	if b.EmployeesMin != nil || b.EmployeesMax != nil {
		lo, hi := "?", "?"
		if b.EmployeesMin != nil {
			lo = fmt.Sprint(*b.EmployeesMin)
		}
		if b.EmployeesMax != nil {
			hi = fmt.Sprint(*b.EmployeesMax)
		}
		size = lo + " to " + hi + " employees"
	}
	lines := []string{
		"Thesis: " + b.Name,
		"Vertical: " + b.Vertical,
		"Subverticals: " + orDefault(b.Subverticals, ", ", "all"),
		"Regions: " + orDefault(b.Regions, ", ", "not specified"),
		"Size: " + size,
		"Must have: " + orDefault(b.MustHave, "; ", "none"),
		"Exclude: " + orDefault(b.Exclusions, "; ", "none"),
	}
	if b.Notes != "" {
		lines = append(lines, "Notes: "+b.Notes)
	}
	lines = append(lines, "", "Request a pilot: "+PilotURL)
	return strings.Join(lines, "\n")
}

// Company holds the three composite inputs on a 0 to 100 scale.
type Company struct {
	ID                  string
	MandateFit          float64
	OutreachSuitability float64
	TransitionContext   float64
	GroupOwned          bool
}

// Composite is the weighted score on a 0 to 100 scale. A group-owned company scores zero on outreach suitability.
func Composite(c Company, w Weights) float64 {
	outreach := c.OutreachSuitability
	if c.GroupOwned {
		outreach = 0
	}
	total := w.MandateFit + w.OutreachSuitability + w.TransitionContext
	v := (c.MandateFit*w.MandateFit + outreach*w.OutreachSuitability + c.TransitionContext*w.TransitionContext) / total
	return math.Round(v*10) / 10
}

// Ranked pairs a company with its composite.
type Ranked struct {
	Company   Company
	Composite float64
}

// Rank returns the companies with their composite, highest first.
func Rank(companies []Company, w Weights) []Ranked {
	out := make([]Ranked, len(companies))
	for i, c := range companies {
		out[i] = Ranked{c, Composite(c, w)}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Composite > out[j].Composite })
	return out
}

// SignalCoverage counts visible and checked signals. An empty value or "not visible" counts as silent.
func SignalCoverage(found map[string]string) (visible, checked int) {
	for _, s := range Signals {
		v, ok := found[s]
		if !ok {
			continue
		}
		checked++
		if v != "" && v != "not visible" {
			visible++
		}
	}
	return
}
