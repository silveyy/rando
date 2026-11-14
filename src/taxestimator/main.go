// Command taxestimator serves a small federal income tax estimator.
package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"taxestimator/tax"
)

//go:embed web/templates web/static
var webFS embed.FS

var funcs = template.FuncMap{
	"money":   money,
	"pct":     func(f float64) string { return strconv.FormatFloat(f*100, 'f', -1, 64) + "%" },
	"pct2":    func(f float64) string { return strconv.FormatFloat(f*100, 'f', 2, 64) + "%" },
	"dollars": func(n int64) string { return money(float64(n)) },
}

// money formats a dollar amount like $1,234.56.
func money(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	intPart, frac, _ := strings.Cut(s, ".")
	var b strings.Builder
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	out := "$" + b.String() + "." + frac
	if neg {
		out = "-" + out
	}
	return out
}

type page struct {
	Years  []int
	Year   int
	Status tax.FilingStatus
	// Raw form text echoed back into the nav.
	IncomeIn  string
	K401In    string
	CharityIn string
	Itemize   bool
	Age50     bool

	Error  string
	Result *tax.Details
}

// parseAmount accepts values like "85000", "85,000" or "$85,000.50".
// Blank input is zero.
func parseAmount(s string) (float64, error) {
	s = strings.NewReplacer("$", "", ",", "", " ", "").Replace(s)
	if s == "" {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f != f || f > 1e12 {
		return 0, strconv.ErrSyntax
	}
	return f, nil
}

type server struct{ tmpl *template.Template }

// compute validates the form values and fills in Result or Error.
func (p *page) compute() {
	income, err := parseAmount(p.IncomeIn)
	if err != nil || strings.TrimSpace(p.IncomeIn) == "" {
		p.Error = "Enter your gross income as a number, e.g. 85000."
		return
	}
	k401, err := parseAmount(p.K401In)
	if err != nil {
		p.Error = "Enter 401(k) contributions as a number, e.g. 10000."
		return
	}
	charity, err := parseAmount(p.CharityIn)
	if err != nil {
		p.Error = "Enter charitable contributions as a number, e.g. 5000."
		return
	}
	if income < 0 || k401 < 0 || charity < 0 {
		p.Error = "Amounts can't be negative."
		return
	}
	res, err := tax.Calculate(tax.Inputs{
		Year: p.Year, Status: p.Status, Gross: income,
		Traditional401k: k401, Age50: p.Age50,
		Itemize: p.Itemize, Charity: charity,
	})
	if err != nil {
		p.Error = err.Error()
		return
	}
	p.Result = res
}

func (s *server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	years := tax.Years()
	p := page{Years: years, Year: years[0], Status: tax.Single}

	q := r.URL.Query()
	if len(q) > 0 {
		if y, err := strconv.Atoi(q.Get("year")); err == nil {
			p.Year = y
		}
		if st := tax.FilingStatus(q.Get("status")); st != "" {
			p.Status = st
		}
		p.IncomeIn = q.Get("income")
		p.K401In = q.Get("k401")
		p.CharityIn = q.Get("charity")
		p.Itemize = q.Get("deduction") == "itemized"
		p.Age50 = q.Get("age50") == "1"
		p.compute()
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if p.Error != "" {
		w.WriteHeader(http.StatusBadRequest)
	}
	if err := s.tmpl.ExecuteTemplate(w, "index.html", p); err != nil {
		log.Printf("render: %v", err)
	}
}

func newHandler() (http.Handler, error) {
	tmpl, err := template.New("").Funcs(funcs).ParseFS(webFS, "web/templates/*.html")
	if err != nil {
		return nil, err
	}
	static, err := fs.Sub(webFS, "web/static")
	if err != nil {
		return nil, err
	}
	s := &server{tmpl: tmpl}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.index)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	return mux, nil
}

func main() {
	h, err := newHandler()
	if err != nil {
		log.Fatal(err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, h))
}
