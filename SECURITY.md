VERDICT: CHANGES_REQUESTED

Scanner-Hinweis: Für diesen Projekttyp wurden keine automatisierten Security-Scanner ausgeführt (`no applicable security scanners`). Das Fehlen von Scanner-Ergebnissen ist kein Nachweis für Abwesenheit von Schwachstellen. Die folgende Bewertung beruht auf manueller Codeanalyse.

## Befunde

### 1. Mittel — Inkonsistente Zugriffskontrolle auf lesende Verwaltungs-Endpunkte
**Betroffene Stelle:** `main.go`, Routen für `GET /flags` und `GET /flags/{key}`.

Die Endpunkte `GET /flags` und `GET /flags/{key}` sind ohne Authentifizierung erreichbar, während die mutierenden Endpunkte `POST`, `PUT` und `DELETE` durch `api.RequireAuth` geschützt sind. Ein anonymer Client kann dadurch sämtliche Flag-Metadaten (`key`, `enabled`, `description`, `rollout_percent`) auslesen. Das ermöglicht Reconnaissance und legt die interne Feature-Konfiguration offen.

**Fix:**
- Management-Endpunkte konsistent mit `api.RequireAuth` absichern, z. B.:
  ```go
  mux.Handle("GET /flags", api.RequireAuth(api.ListFlags(s)))
  mux.Handle("GET /flags/{key}", api.RequireAuth(api.GetFlag(s)))
  ```
- Öffentlich bleibt bewusst nur `GET /flags/{key}/evaluate` (und optional `GET /healthz`).
- Tests in `flags_list_test.go`, `flags_get_test.go` und `routing_test.go` entsprechend um Autorisierungsheader bzw. 401-Tests ergänzen.

### 2. Niedrig — API-Token-Vergleich nicht zeitkonstant
**Betroffene Stelle:** `internal/api/middleware.go`, Funktion `RequireAuth`.

Der Token wird mit einem normalen Stringvergleich (`auth != "Bearer "+token`) geprüft. Das erlaubt theoretisch einen Timing-Angriff auf den Bearer-Token. Praktisch ist das Ausbeuten über ein Netzwerk bei einem hoch-entropen Token erschwert, dennoch sollte eine zeitkonstante Prüfung verwendet werden.

**Fix:**
- `crypto/subtle.ConstantTimeCompare` einsetzen, z. B.:
  ```go
  import "crypto/subtle"

  // ...
  if token == "" || subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+token)) != 1 {
      writeError(w, http.StatusUnauthorized, "unauthorized")
      return
  }
  ```

### 3. Niedrig — Flag-Key-Format nicht validiert
**Betroffene Stelle:** `internal/api/flags_create.go`, Validierungsblock.

Der Key wird nur auf leer und auf maximal 128 Runes geprüft. Zeichen wie `/`, `?`, `#`, Leerzeichen oder Steuerzeichen sind erlaubt. Ein Key mit `/` kann anschließend über die Pfad-Routen (`GET /flags/{key}`, `PUT`, `DELETE`, `Evaluate`) nicht adressiert werden, was zu inkonsistentem Verhalten führen kann. Sonderzeichen können zudem in URLs, Logs oder bei der Integration mit anderen Systemen Probleme verursachen.

**Fix:**
- Erlaubten Zeichensatz definieren und validieren, z. B.:
  ```go
  import "regexp"

  var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

  // nach der Längenprüfung:
  if !keyPattern.MatchString(req.Key) {
      writeError(w, http.StatusBadRequest, "key contains invalid characters")
      return
  }
  ```
- Test für ungültige Zeichen ergänzen.

### 4. Niedrig — Transportverschlüsselung nicht erzwungen
**Betroffene Stelle:** `main.go`, `server.ListenAndServe()`.

Der Server lauscht ohne TLS. Standardmäßig ist der Host zwar auf `127.0.0.1` gesetzt, aber per `HOST=0.0.0.0` kann der Dienst im Netzwerk exponiert werden. In diesem Fall würden der Bearer-Token und die Flag-Daten im Klartext übertragen.

**Fix:**
- Standard-Loopback beibehalten und dokumentieren.
- Für Exposition im Netzwerk TLS aktivieren, z. B. `ListenAndServeTLS` mit Zertifikat/Key, oder einen TLS-terminierenden Reverse Proxy verbindlich vorschreiben.
- Optional: beim Start eine Warnung ausgeben, wenn `HOST != 127.0.0.1` und keine TLS-Konfiguration erkannt wird.

### 5. Niedrig — Caching-Header für Flag-Antworten nicht gesetzt
**Betroffene Stellen:** `internal/api/respond.go` (`writeJSON`), `internal/api/flags_list.go`, `internal/api/flags_get.go`, `internal/api/flags_update.go`, `internal/api/flags_create.go`.

Nur `EvaluateFlag` setzt `Cache-Control: no-store`. Andere JSON-Antworten mit Flag-Metadaten (Liste, Einzelabruf, nach Änderung) können von zwischengeschalteten Caches gespeichert werden. Feature-Flag-Konfiguration kann vertraulich sein.

**Fix:**
- In `writeJSON` generell `Cache-Control: no-store` setzen:
  ```go
  func writeJSON(w http.ResponseWriter, status int, v any) {
      w.Header().Set("Content-Type", "application/json")
      w.Header().Set("Cache-Control", "no-store")
      w.WriteHeader(status)
      _ = json.NewEncoder(w).Encode(v)
  }
  ```
- Die bereits vorhandene explizite `no-store`-Setzung in `flags_evaluate.go` kann bleiben oder entfernt werden, da sie dann redundant ist.

## Positive Beobachtungen

- **Keine hartkodierten Secrets:** `FEATUREFLAGS_API_TOKEN` wird aus der Umgebung gelesen; keine Passwörter, Token oder Schlüssel im Repository.
- **Body-Limit:** `maxBodyBytes` (1 MiB) wird über einen `io.LimitedReader` durchgesetzt; übermäßig große Bodies führen zu 400, ohne dass der gesamte Body gepuffert wird.
- **Fehlerantworten:** JSON-Fehlerantworten enthalten generische Texte ohne interne Fehlermeldungen oder Stacktraces.
- **Datenschutz im Logging:** Die Logging-Middleware protokolliert ausschließlich Methode, Pfad und Statuscode; Query-Parameter (insbesondere `user`) und Request-Bodies werden nicht geloggt.
- **Thread-Sicherheit:** Der In-Memory-Store verwendet `sync.RWMutex`; Evaluierungen verändern den Store nicht.
- **Keine externen Abhängigkeiten:** Der sichtbare Code nutzt ausschließlich die Go-Standardbibliothek.
- **Server-Härtung:** `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout` und `MaxHeaderBytes` sind gesetzt.

## Gesamtbewertung

Es wurden keine kritischen oder hohen Sicherheitslücken wie hartkodierte Secrets, Injection/RCE, Auth-Bypass oder ausnutzbare Dependency-Schwachstellen festgestellt. Die vorhandenen Befunde sind überwiegend Härtungsmaßnahmen und Zugriffskontrollverbesserungen. Daher: `CHANGES_REQUESTED`.