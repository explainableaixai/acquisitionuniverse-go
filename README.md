# acquisitionuniverse-go

Go helpers for acquisition target work: brief validation, composite scoring, ranking and signal coverage. Standard library only, zero network calls.

```bash
go get github.com/explainableaixai/acquisitionuniverse-go
```

## When you would reach for it

You receive a screened list and want a nightly job that re-ranks it, flags thin records and renders a brief for the deal team. Go is a good fit for that kind of small, boring service. The functions below are the parts you would otherwise copy between projects.

The same ordering logic supports [buy-side long lists](https://www.acquisitionuniverse.com/use-cases/buy-side-long-lists.php): many companies, a few clear leaders, and a record of why.

## Package contents

```go
var Signals [15]string
var DefaultWeights Weights            // 0.7, 0.2, 0.1
const PilotURL string

type ThesisBrief struct { ... }       // Validate() []string, Text() string
type Company struct { ... }
func Composite(c Company, w Weights) float64
func Rank(companies []Company, w Weights) []Ranked
func SignalCoverage(found map[string]string) (visible, checked int)
```

## Example

```go
package main

import (
	"fmt"

	au "github.com/explainableaixai/acquisitionuniverse-go"
)

func main() {
	list := []au.Company{
		{ID: "H-01", MandateFit: 90, OutreachSuitability: 82, TransitionContext: 55},
		{ID: "H-02", MandateFit: 97, OutreachSuitability: 92, TransitionContext: 75, GroupOwned: true},
	}
	for _, r := range au.Rank(list, au.DefaultWeights) {
		fmt.Printf("%s %.1f\n", r.Company.ID, r.Composite)
	}
}
```

H-02 would win on fit alone. The group-owned flag removes its outreach points, so H-01 comes first.

## Brief validation in a pipeline step

```go
lo, hi := 20, 200
b := au.ThesisBrief{
	Name:         "Commercial HVAC, Mountain West",
	Vertical:     "Commercial HVAC and mechanical",
	Regions:      []string{"Colorado", "Utah"},
	EmployeesMin: &lo,
	EmployeesMax: &hi,
	Exclusions:   []string{"group-owned"},
}
if problems := b.Validate(); len(problems) > 0 {
	log.Fatalf("fix the brief first: %v", problems)
}
fmt.Println(b.Text())
```

## Coverage report

```go
found := map[string]string{
	au.Signals[0]:  "family associated",
	au.Signals[4]:  "Colorado, Utah",
	au.Signals[13]: "",
}
v, c := au.SignalCoverage(found)
fmt.Printf("%d of %d signals visible\n", v, c)
```

An empty string or `not visible` counts as silent.

## Serving the ranked list over HTTP

```go
http.HandleFunc("/ranked", func(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(au.Rank(load(), au.DefaultWeights))
})
```

## Elsewhere in the portfolio

The same company maintains [audience personas for advertising](https://www.cookielessaudiences.com/features/audience-personas-advertising.php) data for ad tech and a [skill normalization API](https://www.resumereaderapi.com/normalization/skills.php) for recruiting products.

<!--expanded-->
## A scoring service in Go

Go is a good language for the plumbing of a data team: small services, scheduled jobs and command line tools. This module gives those programs a shared definition of how a screened company is scored, so the ranking in your nightly job, your API and your report all agree.

The code is a single package with no dependencies and no network calls. It is easy to vendor, easy to audit and quick to test.

## One definition of the score

Disagreements about rankings usually trace back to two programs that implement the formula slightly differently. One rounds, the other truncates. One forgets the group ownership rule. Importing a shared package removes that class of bug.

```go
func Composite(c Company, w Weights) float64 {
	outreach := c.OutreachSuitability
	if c.GroupOwned {
		outreach = 0
	}
	total := w.MandateFit + w.OutreachSuitability + w.TransitionContext
	v := (c.MandateFit*w.MandateFit + outreach*w.OutreachSuitability + c.TransitionContext*w.TransitionContext) / total
	return math.Round(v*10) / 10
}
```

Fit counts for 70 percent, suitability for 20 percent and context for 10 percent by default. A company already owned by a group has no suitability, whatever its fit.

## A small HTTP service

Expose the ranking as JSON so that other tools can use it. The handler below reads companies from a request body and returns them in order.

```go
func rankHandler(w http.ResponseWriter, r *http.Request) {
	var in []au.Company
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	out := au.Rank(in, au.DefaultWeights)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}
```

Add basic authentication and TLS before you put it on a network. Target lists are sensitive.

## Command line tool

A short program that reads a CSV and prints the top rows is a good first use. It also doubles as documentation for new colleagues, who can read the code and see exactly how the ranking works.

```go
func main() {
	f, _ := os.Open(os.Args[1])
	rows, _ := csv.NewReader(f).ReadAll()
	var cs []au.Company
	for _, r := range rows[1:] {
		fit, _ := strconv.ParseFloat(r[1], 64)
		out, _ := strconv.ParseFloat(r[2], 64)
		tr, _ := strconv.ParseFloat(r[3], 64)
		cs = append(cs, au.Company{ID: r[0], MandateFit: fit, OutreachSuitability: out, TransitionContext: tr, GroupOwned: r[4] == "yes"})
	}
	for i, r := range au.Rank(cs, au.DefaultWeights) {
		if i == 20 {
			break
		}
		fmt.Printf("%-12s %5.1f\n", r.Company.ID, r.Composite)
	}
}
```

In real code, handle the errors that the snippet ignores. A list with a malformed row should fail loudly, not silently drop the company.

<!--extra-->
## Vertical signals in Go

A screening adds extra signals for the trade you are screening, on top of the fifteen common ones. For commercial HVAC and mechanical contractors they include service contract language, named equipment partnerships and multi-branch footprint. For industrial coatings they include NADCAP accreditation and ITAR registration. Represent them as `map[string]string` on your own record type, and keep the fifteen common signals in a separate map. `SignalCoverage` counts only the names in `Signals`, so your extras never distort the figure, and you can write a ten line helper that computes a second ratio for them.

Write the helper once and test it with a table, including a case where every extra is silent and a case where every extra is visible. Those two cases catch most mistakes in coverage code.

Finally, remember to expose the vertical in your reports. A ranking that mixes verticals without saying so produces confusing lists, because the signals that matter differ. Group by vertical, rank within each group and put the name of the vertical in the file name.

A closing thought: small services last when they are boring. Keep the module pinned, keep the weights in configuration, log the version and the weights in every report and keep a short runbook. Boring software is software that nobody has to think about at night.

<!--further-->
## Further reading and practical notes

The [Go language site](https://go.dev/) documents the standard library and the tooling used by this module. For acquirers who plan to use bank financing, the [US Small Business Administration](https://www.sba.gov/) explains the loan programmes that many small business buyers rely on.

Notes for production. Treat the ranking as a pure function and keep it that way. Do not read configuration or the clock inside it, and it will stay trivial to test. Put the I/O at the edges: a loader that turns a file into `[]Company` and a writer that turns `[]Ranked` into a report. When a stakeholder asks for a new column, you add it in the writer, and the formula stays untouched.

If you serve the ranking over HTTP, set limits. Cap the request body, cap the number of companies and set timeouts. Return the weights you used in the response, and the package version, so a saved response explains itself.

Fuzz the formula. A fuzz test that feeds random floats, including negative numbers, `NaN` and very large values, will show you where your inputs need validation. Decide what the function should do and write it down: clamp, reject or treat as zero. Any of the three is fine, as long as it is a decision and not an accident.

Finally, log a hash of the input file next to each report. Months later, the hash tells you whether two reports used the same data, which settles many arguments before they begin.

## Testing the formula

Table driven tests suit this module well. Write cases for the extremes (all zeros, all hundreds), for the group ownership rule, for custom weights, and for rounding. Add a fuzz test that generates random scores between 0 and 100 and asserts that the composite stays in range and that setting `GroupOwned` never increases it.

## Interpreting coverage

`SignalCoverage` takes a map from signal name to a value and returns the number visible and the number checked. An empty string or the text `not visible` counts as silent. The 15 names are available in `Signals`. Coverage helps you decide where to spend human attention, since a high scoring company with thin evidence is a better candidate for a quick manual read than a mid scoring company with rich evidence.

## Reading list

Buy side teams often want to see how a long list is built before they decide what to do with it. The page on [buy side long lists](https://www.acquisitionuniverse.com/use-cases/buy-side-long-lists.php) describes it. The same company publishes [audience personas for advertising](https://www.cookielessaudiences.com/features/audience-personas-advertising.php) for ad tech teams and a [skill normalization API](https://www.resumereaderapi.com/normalization/skills.php) for recruiting products.

## Versioning

The module follows semantic versioning. Tags are published from the repository, and the first release is v1.0.0. Pin it in `go.mod` and read the changelog before you upgrade.

## Questions

**Can I use it with generics or custom score types?** The types are concrete on purpose. Convert your own types into `Company` at the edge.

**Does it allocate a lot?** `Rank` allocates one slice for the result. Everything else is on the stack.

**License?** MIT. Questions to info@alpha-quantum.com.

## FAQ

**Is there a CSV reader?** Use `encoding/csv` and map your columns.

**Go version?** 1.20 or later.

**License?** MIT. info@alpha-quantum.com
