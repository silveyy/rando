# Tax Estimator

Small Go web app: enter gross income, filing status and tax year to see federal tax per bracket.

    go run .          # serves on :8080 (set PORT to change)
    go test ./...

Bracket data lives in `tax/data.go`; add a year by adding a map entry.
