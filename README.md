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

## FAQ

**Is there a CSV reader?** Use `encoding/csv` and map your columns.

**Go version?** 1.20 or later.

**License?** MIT. info@alpha-quantum.com
