VERDICT: BUGS_FOUND

- **Titel**: Routing-Test bewertet korrektes 404 für unbekannten Key fälschlich als fehlende Route
- **Symptom**: Die Test-Suite schlägt fehl (`go test ./...` exit 1), obwohl die API für unbekannte Flags laut Spezifikation (AC-03) korrekt 404 liefert. Der Test verwechselt „Route nicht verdrahtet“ (Handler existiert nicht) mit „Ressource nicht gefunden“ (Flag nicht im Store).
- **Repro**: `go test ./...` im Projektverzeichnis ausführen; der Subtest `TestRoutesAreWired` bricht bei `get_flag` und `delete_flag` ab.
- **Evidence**:
  - `--- FAIL: TestRoutesAreWired (0.00s)`
  - `--- FAIL: TestRoutesAreWired/get_flag (0.00s)`
  - `routing_test.go:33: GET /flags/myfeature should be wired, got 404`
  - `--- FAIL: TestRoutesAreWired/delete_flag (0.00s)`
  - `routing_test.go:33: DELETE /flags/myfeature should be wired, got 404`
- **Suspected file(s)**: `routing_test.go` (die Testlogik prüft bei unbekanntem Key fälschlich auf `rec.Code != http.StatusNotFound`; GET/DELETE auf einen nicht existierenden Key liefern per Design 404, was der Test als „Route fehlt“ fehlinterpretiert).
- **Severity**: high (CI/Test-Lauf ist rot; blockiert das Ausliefern, obwohl die Produktlogik der Spezifikation entspricht)