VERDICT: BUGS_FOUND

**Title:** Routing-Smoke-Test bewertet legitime 404-Antworten für unbekannte Flag-Keys als „Route nicht verdrahtet“

**Symptom:** Die Go-Test-Suite ist rot: `go test ./...` endet mit Exit-Code 1. Der Produktivcode der API-Pakete (`internal/api`, `internal/evaluate`, `internal/store`) ist zwar grün, aber der Routingtest `TestRoutesAreWired` schlägt für `GET /flags/myfeature` und `DELETE /flags/myfeature` fehl, weil er eine 404-Antwort als fehlende Route interpretiert. Für einen unbekannten Key ist 404 gemäß AC-03 jedoch korrektes Verhalten. Damit ist AC-09 verletzt („laufen mit go test grün durch“).

**Repro:** `go test ./...` im Projektstamm ausführen.

**Evidence:**
```
--- FAIL: TestRoutesAreWired (0.00s)
    --- FAIL: TestRoutesAreWired/get_flag (0.00s)
        routing_test.go:33: GET /flags/myfeature should be wired, got 404
    --- FAIL: TestRoutesAreWired/delete_flag (0.00s)
        routing_test.go:33: DELETE /flags/myfeature should be wired, got 404
FAIL
FAIL	featureflags	0.617s
```

**Suspected file(s):** `routing_test.go`

**Severity:** high