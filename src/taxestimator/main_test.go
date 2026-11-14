package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func get(t *testing.T, url string) (*httptest.ResponseRecorder, string) {
	t.Helper()
	h, err := newHandler()
	if err != nil {
		t.Fatal(err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", url, nil))
	return rec, rec.Body.String()
}

func TestIndexForm(t *testing.T) {
	rec, body := get(t, "/")
	if rec.Code != 200 || !strings.Contains(body, "<form") || strings.Contains(body, "<table") || !strings.Contains(body, "Age 50") {
		t.Fatalf("code=%d", rec.Code)
	}
}

func TestEstimate(t *testing.T) {
	// 60,000 - 14,600 standard = 45,400 taxable: 1,160 + 4,056 = 5,216
	rec, body := get(t, "/?income=60,000&status=single&year=2024&deduction=standard")
	if rec.Code != 200 {
		t.Fatalf("code=%d", rec.Code)
	}
	for _, want := range []string{"$5,216.00", "$1,160.00", "$4,056.00", "$45,400.00", "Standard deduction", "2024"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
}

func TestBadInput(t *testing.T) {
	for _, u := range []string{"/?income=abc&status=single&year=2024", "/?income=-5&status=single&year=2024", "/?income=1&status=x&year=2024", "/?income=1&status=joint&year=1900", "/?income=NaN&year=2024", "/?income=Inf&year=2024"} {
		rec, body := get(t, u)
		if rec.Code != http.StatusBadRequest || !strings.Contains(body, `role="alert"`) {
			t.Errorf("%s: code=%d", u, rec.Code)
		}
	}
	if rec, _ := get(t, "/nope"); rec.Code != 404 {
		t.Errorf("want 404")
	}
}

func TestItemized(t *testing.T) {
	// 60,000 - 20,000 charity = 40,000 taxable: 1,160 + 12% of 28,400 = 4,568.00 -> 4,568
	_, body := get(t, "/?income=60000&status=single&year=2024&deduction=itemized&charity=20,000")
	for _, want := range []string{"$4,568.00", "$40,000.00", "Itemized charitable contributions"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	if strings.Contains(body, "capped") {
		t.Error("unexpected cap note")
	}
	if strings.Contains(body, "is higher than your itemized amount") {
		t.Error("unexpected standard-higher note")
	}
	// Small contribution: standard is better, and a hint is shown.
	_, body = get(t, "/?income=60000&status=single&year=2024&deduction=itemized&charity=1000")
	if !strings.Contains(body, "is higher than your itemized amount") {
		t.Error("missing standard-higher hint")
	}
	// Cap at 60% of income.
	_, body = get(t, "/?income=10000&status=single&year=2024&deduction=itemized&charity=9000")
	if !strings.Contains(body, "limited to 60%") || !strings.Contains(body, "$4,000.00") {
		t.Error("missing cap")
	}
	rec, _ := get(t, "/?income=1000&status=single&year=2024&deduction=itemized&charity=abc")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bad charity: code=%d", rec.Code)
	}
}

func Test401k(t *testing.T) {
	// 100k - 10k 401k - 14,600 std = 75,400 taxable; tax 11,641
	_, body := get(t, "/?income=100000&status=single&year=2024&deduction=standard&k401=10,000")
	for _, want := range []string{"$11,641.00", "$75,400.00", "Traditional 401(k)", "−$10,000.00"} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q", want)
		}
	}
	// over the limit is capped; age50 raises it
	_, body = get(t, "/?income=200000&status=single&year=2024&k401=40000")
	if !strings.Contains(body, "limited to $23,000.00") {
		t.Error("missing 401k cap note")
	}
	_, body = get(t, "/?income=200000&status=single&year=2024&k401=40000&age50=1")
	if !strings.Contains(body, "limited to $30,500.00") || !strings.Contains(body, `value="1" checked> Age 50`) {
		t.Error("catch-up not applied")
	}
	rec, _ := get(t, "/?income=1000&k401=abc&year=2024")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("bad 401k: code=%d", rec.Code)
	}
}
