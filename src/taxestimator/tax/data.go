package tax

// Federal ordinary-income brackets, from IRS annual inflation-adjustment
// revenue procedures. Upper bounds are in whole dollars.
var tables = map[int]map[FilingStatus][]Bracket{
	2023: {
		Single: build([]int64{11000, 44725, 95375, 182100, 231250, 578125}),
		Joint:  build([]int64{22000, 89450, 190750, 364200, 462500, 693750}),
	},
	2024: {
		Single: build([]int64{11600, 47150, 100525, 191950, 243725, 609350}),
		Joint:  build([]int64{23200, 94300, 201050, 383900, 487450, 731200}),
	},
	2025: {
		Single: build([]int64{11925, 48475, 103350, 197300, 250525, 626350}),
		Joint:  build([]int64{23850, 96950, 206700, 394600, 501050, 751600}),
	},
	2026: {
		Single: build([]int64{12400, 50400, 105700, 201775, 256225, 640600}),
		Joint:  build([]int64{24800, 100800, 211400, 403550, 512450, 768700}),
	},
}

// Standard deductions by year and filing status.
var standardDeductions = map[int]map[FilingStatus]float64{
	2023: {Single: 13850, Joint: 27700},
	2024: {Single: 14600, Joint: 29200},
	2025: {Single: 15750, Joint: 31500},
	2026: {Single: 16100, Joint: 32200},
}

// 401(k) employee elective deferral limits: base and age-50+ catch-up.
var limits401k = map[int]struct{ Base, CatchUp float64 }{
	2023: {22500, 7500},
	2024: {23000, 7500},
	2025: {23500, 7500},
	2026: {24500, 8000},
}

var rates = []float64{0.10, 0.12, 0.22, 0.24, 0.32, 0.35, 0.37}

// build turns the six bracket thresholds into seven brackets.
func build(uppers []int64) []Bracket {
	bs := make([]Bracket, 0, len(rates))
	var lower int64
	for i, r := range rates {
		b := Bracket{Rate: r, Lower: lower}
		if i < len(uppers) {
			b.Upper = uppers[i]
			lower = uppers[i]
		} else {
			b.Unbounded = true
		}
		bs = append(bs, b)
	}
	return bs
}
