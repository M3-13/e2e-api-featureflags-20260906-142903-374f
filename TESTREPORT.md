VERDICT: BUGS_FOUND

Der Build (`go build ./...`) ist sauber, und die Paket-Tests für `internal/api`, `internal/evaluate` sowie `internal/store` sind grün. Der Gesamtlauf `go test ./...` schlägt jedoch mit Exit 1 fehl. Ursache ist ein falscher Routing-Test: Er wertet die spezifikationskonformen 404-Antworten für unbekannte Flag-`key`s fälschlich als „Route nicht verdrahtet“. `GET /flags/{key}` und `DELETE /flags/{key}` müssen laut AC-03 für einen unbekannten `key` mit 404 und JSON-Fehlerobjekt antworten. `TestRoutesAreWired` fragt genau den unbekannten `myfeature` ab, ohne vorher ein Flag anzulegen, und erwartet deshalb zu Unrecht keinen 404.

**Bugliste**

- **Titel**: `TestRoutesAreWired` wertet legitime 404-Antworten für unbekannte Keys als Routingfehler
- **Symptom**: `go test ./...` endet mit Exit 1; die Untertests `get_flag` und `delete_flag` schlagen fehl, obwohl die Handler korrekt registriert sind und das von der Spezifikation verlangte 404 liefern. Damit ist AC-09 („laufen mit go test grün durch“) verletzt und das CI-Gate bricht.
- **Repro**: Im Projektstamm `go test ./...` ausführen.
- **Beleg**:
  ```
  --- FAIL: TestRoutesAreWired (0.00s)
      --- FAIL: TestRoutesAreWired/get_flag (0.00s)
          routing_test.go:33: GET /flags/myfeature should be wired, got 404
      --- FAIL: TestRoutesAreWired/delete_flag (0.00s)
          routing_test.go:33: DELETE /flags/myfeature should be wired, got 404
  FAIL
  FAIL	featureflags	0.621s
  ```
- **Verdächtige Datei(en)**: `routing_test.go` (Logik in `TestRoutesAreWired`)
- **Schweregrad**: high