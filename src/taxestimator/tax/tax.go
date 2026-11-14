// Package tax computes US federal income tax per bracket.
package tax

import (
	"errors"
	"fmt"
	"sort"
)

// FilingStatus is a federal filing status.
type FilingStatus string

const (
	Single FilingStatus = "single"
	Joint  FilingStatus = "joint"
)

// Label returns a human-readable name.
func (f FilingStatus) Label() string {
	switch f {
	case Single:
		return "Single"
	case Joint:
		return "Married filing jointly"
	}
	return string(f)
}

// Valid reports whether f is a supported filing status.
func (f FilingStatus) Valid() bool { return f == Single || f == Joint }

// Bracket is one marginal bracket. Upper is the top of the bracket in whole
// dollars; Unbounded marks the final bracket, which has no upper limit.
type Bracket struct {
	Rate      float64 // e.g. 0.10
	Lower     int64
	Upper     int64
	Unbounded bool
}

// BracketResult is the tax owed within one bracket.
type BracketResult struct {
	Bracket
	IncomeInBracket float64
	Tax             float64
}

// Result is the full estimate.
type Result struct {
	Year          int
	Status        FilingStatus
	Income        float64 // gross income
	Deduction     float64 // deduction applied
	TaxableIncome float64
	Brackets      []BracketResult
	TotalTax      float64
	EffectiveRate float64 // TotalTax / gross Income, 0 when Income is 0
	MarginalRate  float64 // rate of the highest bracket reached
}

// Years returns the supported tax years, newest first.
func Years() []int {
	ys := make([]int, 0, len(tables))
	for y := range tables {
		ys = append(ys, y)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(ys)))
	return ys
}

// StandardDeduction returns the standard deduction for a year and status.
func StandardDeduction(year int, status FilingStatus) (float64, error) {
	if _, err := Brackets(year, status); err != nil {
		return 0, err
	}
	return standardDeductions[year][status], nil
}

// CharitableLimit is the maximum share of gross income deductible for cash
// charitable contributions (60% of AGI).
const CharitableLimit = 0.60

// CharitableDeduction returns the itemized deduction for cash charitable
// contributions, capped at CharitableLimit of gross income.
func CharitableDeduction(gross, contributions float64) float64 {
	limit := round2(gross * CharitableLimit)
	if contributions > limit {
		return limit
	}
	return contributions
}

// Brackets returns the brackets for a year and status.
func Brackets(year int, status FilingStatus) ([]Bracket, error) {
	t, ok := tables[year]
	if !ok {
		return nil, fmt.Errorf("unsupported tax year %d", year)
	}
	if !status.Valid() {
		return nil, fmt.Errorf("unsupported filing status %q", status)
	}
	return t[status], nil
}

// Estimate computes the tax owed in each bracket on gross income (dollars)
// after subtracting deduction. The deduction is capped at gross income, so
// taxable income is never negative. Amounts are rounded to the cent.
func Estimate(year int, status FilingStatus, gross, deduction float64) (*Result, error) {
	if gross < 0 {
		return nil, errors.New("income must not be negative")
	}
	if deduction < 0 {
		return nil, errors.New("deduction must not be negative")
	}
	bs, err := Brackets(year, status)
	if err != nil {
		return nil, err
	}
	if deduction > gross {
		deduction = gross
	}
	income := round2(gross - deduction)
	res := &Result{Year: year, Status: status, Income: gross, Deduction: round2(deduction), TaxableIncome: income}
	var totalCents int64
	for _, b := range bs {
		br := BracketResult{Bracket: b}
		if income > float64(b.Lower) {
			top := income
			if !b.Unbounded && float64(b.Upper) < top {
				top = float64(b.Upper)
			}
			br.IncomeInBracket = round2(top - float64(b.Lower))
			cents := int64(br.IncomeInBracket*100 + 0.5) // non-negative
			taxCents := (cents*int64(b.Rate*10000+0.5) + 5000) / 10000
			br.Tax = float64(taxCents) / 100
			totalCents += taxCents
			res.MarginalRate = b.Rate
		}
		res.Brackets = append(res.Brackets, br)
	}
	res.TotalTax = float64(totalCents) / 100
	if gross > 0 {
		res.EffectiveRate = res.TotalTax / gross
	}
	return res, nil
}

func round2(f float64) float64 { return float64(int64(f*100+0.5)) / 100 }
