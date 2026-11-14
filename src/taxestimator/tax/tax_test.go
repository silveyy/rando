package tax

import (
	"math"
	"testing"
)

func TestEstimate(t *testing.T) {
	tests := []struct {
		name    string
		year    int
		status  FilingStatus
		income  float64
		total   float64
		nonzero int
	}{
		{"zero", 2024, Single, 0, 0, 0},
		{"within first bracket", 2024, Single, 10000, 1000, 1},
		// 1160 + 12% of (47150-11600)=4266 + 22% of (60000-47150)=2827
		{"single 2024 60k", 2024, Single, 60000, 8253, 3},
		// 2320 + 12% of 71100=8532 + 22% of (120000-94300)=5654
		{"joint 2024 120k", 2024, Joint, 120000, 16506, 3},
		{"single 2025 exactly at bracket top", 2025, Single, 11925, 1192.5, 1},
		// 1192.50+4386+12072.50+22548+17032+131538.75+138250.50
		{"single 2025 1M", 2025, Single, 1000000, 327020.25, 7},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := Estimate(tc.year, tc.status, tc.income, 0)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(r.TotalTax-tc.total) > 0.005 {
				t.Errorf("total = %v, want %v", r.TotalTax, tc.total)
			}
			var sum, inc float64
			n := 0
			for _, b := range r.Brackets {
				sum += b.Tax
				inc += b.IncomeInBracket
				if b.Tax > 0 {
					n++
				}
			}
			if math.Abs(sum-r.TotalTax) > 0.005 {
				t.Errorf("bracket sum %v != total %v", sum, r.TotalTax)
			}
			if math.Abs(inc-tc.income) > 0.005 {
				t.Errorf("income in brackets %v != income %v", inc, tc.income)
			}
			if n != tc.nonzero {
				t.Errorf("nonzero brackets = %d, want %d", n, tc.nonzero)
			}
		})
	}
}

func TestErrors(t *testing.T) {
	if _, err := Estimate(1999, Single, 1, 0); err == nil {
		t.Error("want error for unsupported year")
	}
	if _, err := Estimate(2024, "bogus", 1, 0); err == nil {
		t.Error("want error for bad status")
	}
	if _, err := Estimate(2024, Single, -1, 0); err == nil {
		t.Error("want error for negative income")
	}
}

func TestTablesWellFormed(t *testing.T) {
	for _, y := range Years() {
		for _, s := range []FilingStatus{Single, Joint} {
			bs, err := Brackets(y, s)
			if err != nil || len(bs) != 7 || !bs[6].Unbounded {
				t.Fatalf("%d %s malformed", y, s)
			}
			for i := 1; i < len(bs); i++ {
				if bs[i].Lower != bs[i-1].Upper || bs[i].Upper < bs[i].Lower && !bs[i].Unbounded {
					t.Errorf("%d %s bracket %d not contiguous", y, s, i)
				}
			}
		}
	}
}

func TestDeduction(t *testing.T) {
	// 2024 single, 60k gross, 14,600 standard => 45,400 taxable:
	// 1160 + 12% of (45400-11600)=4056
	r, err := Estimate(2024, Single, 60000, 14600)
	if err != nil {
		t.Fatal(err)
	}
	if r.TaxableIncome != 45400 || math.Abs(r.TotalTax-5216) > 0.005 {
		t.Errorf("taxable=%v tax=%v", r.TaxableIncome, r.TotalTax)
	}
	if math.Abs(r.EffectiveRate-5216.0/60000) > 1e-9 {
		t.Errorf("effective rate %v", r.EffectiveRate)
	}
	// Deduction larger than income: no tax, no negative taxable income.
	r, _ = Estimate(2024, Single, 10000, 14600)
	if r.TaxableIncome != 0 || r.TotalTax != 0 || r.Deduction != 10000 {
		t.Errorf("%+v", r)
	}
	if _, err := Estimate(2024, Single, 1, -1); err == nil {
		t.Error("want error for negative deduction")
	}
}

func TestStandardAndCharitable(t *testing.T) {
	for _, y := range Years() {
		for _, s := range []FilingStatus{Single, Joint} {
			if d, err := StandardDeduction(y, s); err != nil || d <= 0 {
				t.Errorf("%d %s: %v %v", y, s, d, err)
			}
		}
	}
	if got := CharitableDeduction(100000, 5000); got != 5000 {
		t.Errorf("got %v", got)
	}
	if got := CharitableDeduction(100000, 90000); got != 60000 {
		t.Errorf("cap: got %v", got)
	}
}

func TestCalculate401k(t *testing.T) {
	// 2024 single, 100k gross, 10k 401k, standard 14,600 => taxable 75,400
	// 1160 + 12%*(47150-11600)=4266 + 22%*(75400-47150)=6215 = 11641
	d, err := Calculate(Inputs{Year: 2024, Status: Single, Gross: 100000, Traditional401k: 10000})
	if err != nil {
		t.Fatal(err)
	}
	if d.AdjustedIncome != 90000 || d.TaxableIncome != 75400 || math.Abs(d.TotalTax-11641) > 0.005 {
		t.Errorf("%+v tax=%v", d, d.TotalTax)
	}
	if d.Income != 100000 || math.Abs(d.EffectiveRate-0.11641) > 1e-9 || d.Capped401k {
		t.Errorf("income=%v eff=%v capped=%v", d.Income, d.EffectiveRate, d.Capped401k)
	}
}

func TestCalculate401kLimit(t *testing.T) {
	d, _ := Calculate(Inputs{Year: 2024, Status: Single, Gross: 200000, Traditional401k: 40000})
	if !d.Capped401k || d.Contribution401k != 23000 || d.Limit401k != 23000 {
		t.Errorf("%+v", d)
	}
	d, _ = Calculate(Inputs{Year: 2024, Status: Single, Gross: 200000, Traditional401k: 40000, Age50: true})
	if d.Contribution401k != 30500 || !d.Capped401k {
		t.Errorf("catch-up: %+v", d)
	}
	d, _ = Calculate(Inputs{Year: 2024, Status: Single, Gross: 5000, Traditional401k: 20000})
	if d.Contribution401k != 5000 || d.TotalTax != 0 {
		t.Errorf("limited by gross: %+v", d)
	}
}

func TestCalculateCharityUsesAdjustedIncome(t *testing.T) {
	// adjusted = 10000-5000 = 5000; 60% cap = 3000
	d, err := Calculate(Inputs{Year: 2024, Status: Single, Gross: 10000, Traditional401k: 5000, Itemize: true, Charity: 4000})
	if err != nil {
		t.Fatal(err)
	}
	if d.Deduction != 3000 || !d.CharityCapped || d.TaxableIncome != 2000 {
		t.Errorf("%+v", d.Result)
	}
	if !d.StandardIsHigher {
		t.Error("want StandardIsHigher")
	}
}
