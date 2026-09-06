VERDICT: BUGS_FOUND

**Bug 1: Routing-Test `TestRoutesAreWired` verwechselt 404 für unbekannte Flag-Keys mit fehlender Route**
- **Titel**: Routing-Test verwechselt 404 für unbekannte Flag-Keys mit fehlender Route
- **Symptom**: `go test ./...` endet mit Exit 1. Der Test `TestRoutesAreWired` meldet, `GET /flags/myfeature` und `DELETE /flags/myfeature` seien „not wired“ (nicht verdrahtet), obwohl beide Endpunkte laut AC-03 für einen unbekannten Key korrekt 404 liefern. Dadurch wird AC-09 („go test grün“) verletzt und die CI blockiert.
- **Repro**: Im Projekt-Root `go test ./...` ausführen.
- **Evidence**:
  ```
  --- FAIL: TestRoutesAreWired (0.00s)
      --- FAIL: TestRoutesAreWired/get_flag (0.00s)
          routing_test.go:33: GET /flags/myfeature should be wired, got 404
      --- FAIL: TestRoutesAreWired/delete_flag (0.00s)
          routing_test.go:33: DELETE /flags/myfeature should be wired, got 404
  FAIL
  FAIL	featureflags	0.638s
  ```
  Zugehöriges Log des fehlgeschlagenen Tests:
  `POST /flags 400`, `GET /flags/myfeature 404`, `DELETE /flags/myfeature 404`
- **Suspected file(s)**: `routing_test.go` — der Test legt das Flag vermutlich mit ungültigem Request-Body an (`POST /flags` liefert 400), sodass `myfeature` nie existiert; die folgenden GET/DELETE liefern daher korrekt 404, aber der Test interpretiert das als „Route nicht verdrahtet“. Alternativ müssen die Test-Assertions an die 404-Semantik aus AC-03 angepasst werden.
- **Severity**: high