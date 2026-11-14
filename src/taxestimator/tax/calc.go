package tax

import "fmt"

// Inputs are all the user-configurable options for an estimate.
type Inputs struct {
	Year   int
	Status FilingStatus
	Gross  float64 // pay before 401(k) deferrals

	Traditional401k float64 // pre-tax 401(k) contributions requested
	Age50           bool    // eligible for the 401(k) catch-up

	Itemize bool    // itemize charitable contributions instead of the standard deduction
	Charity float64 // cash charitable contributions (used when Itemize)
}

// Limit401k returns the employee 401(k) deferral limit for the year.
func Limit401k(year int, age50 bool) (float64, error) {
	l, ok := limits401k[year]
	if !ok {
		return 0, fmt.Errorf("unsupported tax year %d", year)
	}
	if age50 {
		return l.Base + l.CatchUp, nil
	}
	return l.Base, nil
}

// Details extends Result with how the taxable income was reached.
type Details struct {
	*Result
	Contribution401k  float64 // 401(k) amount actually applied
	Limit401k         float64
	Capped401k        bool // requested amount was reduced to the limit
	AdjustedIncome    float64
	DeductionLabel    string
	CharityCapped     bool    // charity reduced to the 60% limit
	StandardDeduction float64 // standard deduction for comparison
	StandardIsHigher  bool    // itemizing, but the standard deduction is larger
}

// Calculate applies the 401(k) reduction, then the chosen deduction, then
// the brackets.
func Calculate(in Inputs) (*Details, error) {
	if in.Gross < 0 || in.Traditional401k < 0 || in.Charity < 0 {
		return nil, fmt.Errorf("amounts must not be negative")
	}
	limit, err := Limit401k(in.Year, in.Age50)
	if err != nil {
		return nil, err
	}
	std, err := StandardDeduction(in.Year, in.Status)
	if err != nil {
		return nil, err
	}
	d := &Details{Limit401k: limit, StandardDeduction: std}

	k := in.Traditional401k
	if k > limit {
		k = limit
		d.Capped401k = true
	}
	if k > in.Gross {
		k = in.Gross
	}
	d.Contribution401k = round2(k)
	d.AdjustedIncome = round2(in.Gross - d.Contribution401k)

	deduction := std
	d.DeductionLabel = "Standard deduction"
	if in.Itemize {
		deduction = CharitableDeduction(d.AdjustedIncome, in.Charity)
		d.CharityCapped = deduction < in.Charity
		d.DeductionLabel = "Itemized charitable contributions"
		d.StandardIsHigher = std > deduction
	}

	res, err := Estimate(in.Year, in.Status, d.AdjustedIncome, deduction)
	if err != nil {
		return nil, err
	}
	res.Income = in.Gross
	if in.Gross > 0 {
		res.EffectiveRate = res.TotalTax / in.Gross
	}
	d.Result = res
	return d, nil
}
