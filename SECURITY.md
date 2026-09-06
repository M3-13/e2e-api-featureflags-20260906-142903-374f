VERDICT: BLOCKED

## Sicherheitsbericht

### Prüfumfang
Es lagen keine automatisierten Scanner-Ergebnisse vor („no applicable security scanners for this project type“). Die Bewertung basiert daher ausschließlich auf der manuellen Analyse des vorliegenden Go-Quellcodes. Der Code erfüllt die funktionalen und die in den Acceptance Criteria genannten Sicherheitskriterien (Body-Limit, reduzierte Logausgabe, generische Fehlerantworten) grundsätzlich. Es besteht jedoch eine schwerwiegende Lücke bei der Zugriffskontrolle.

### 1. Hoch: Fehlende Authentifizierung und Autorisierung für administrative Endpunkte

**Betroffene Stellen:** `main.go` (`newHandler`, `main`), `internal/api/flags_create.go`, `internal/api/flags_update.go`, `internal/api/flags_delete.go`, `internal/api/flags_list.go`

**Beschreibung:**  
Der Service bindet in `main.go` an `:8080` und damit an alle Netzwerkschnittstellen. Kein einziger Endpunkt verlangt eine Authentifizierung oder prüft Berechtigungen. Insbesondere die mutierenden Endpunkte:

- `POST /flags`
- `PUT /flags/{key}`
- `DELETE /flags/{key}`

sind für jeden erreichbar, der Netzwerkzugriff auf den Dienst hat. Ein Angreifer kann Feature Flags anlegen, verändern oder löschen und dadurch unmittelbar Geschäftslogik bzw. Anwendungskonfiguration manipulieren. Dies entspricht einem Auth-Bypass im Sinne von „kein Zugriffsschutz vorhanden“.

**Konkrete Lösung:**  
- Standardmäßig nur an `127.0.0.1` binden, nicht an `:8080` (alle Interfaces). Die Bind-Adresse über eine Umgebungsvariable wie `HOST` und `PORT` konfigurierbar machen, mit sicherem Default `127.0.0.1:[PORT]`.
- Zusätzlich eine Authentifizierungs-Middleware vor die administrativen Routen schalten, z. B. konfigurierbares Bearer-Token oder mTLS. Beispiel:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := os.Getenv("FEATUREFLAGS_API_TOKEN")
        if token != "" {
            auth := r.Header.Get("Authorization")
            if auth != "Bearer "+token {
                api.WriteError(w, http.StatusUnauthorized, "unauthorized")
                return
            }
        }
        next.ServeHTTP(w, r)
    })
}
```

- Entsprechende Tests für `401` bei fehlendem/ungültigem Token ergänzen.

### 2. Mittel: HTTP-Server ohne Timeouts – Slowloris/Resource-DoS

**Betroffene Stelle:** `main.go`, `http.ListenAndServe(addr, handler)`

**Beschreibung:**  
`http.ListenAndServe` verwendet intern einen `http.Server` ohne `ReadTimeout`, `ReadHeaderTimeout`, `WriteTimeout` oder `IdleTimeout`. Dadurch kann ein Angreifer Verbindungen langsam offen halten oder Anfragen nur sehr langsam senden und so Ressourcen aufbrauchen.

**Konkrete Lösung:**  
Den `http.Server` explizit mit Timeouts konfigurieren:

```go
srv := &http.Server{
    Addr:              addr,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       10 * time.Second,
    WriteTimeout:      10 * time.Second,
    IdleTimeout:       120 * time.Second,
}
if err := srv.ListenAndServe(); err != nil { ... }
```

### 3. Niedrig: Transport unverschlüsselt

**Betroffene Stelle:** `main.go`

**Beschreibung:**  
Der Service verwendet ausschließlich `http` ohne TLS. Flag-Konfigurationen und ggf. später hinzugefügte API-Tokens wären im Klartext transportiert. Aktuell werden keine hochsensitiven persönlichen Daten übertragen, daher niedrig eingestuft.

**Konkrete Lösung:**  
- TLS-Terminierung über einen vorgeschalteten Reverse Proxy (empfohlen) **oder**
- direkte Verwendung von `srv.ListenAndServeTLS` mit konfigurierbaren Zertifikatspfaden.

### 4. Niedrig: Unzureichende Eingabelängen-Validierung für `key` und `description`

**Betroffene Stellen:** `internal/api/flags_create.go`, `internal/api/flags_update.go`, `internal/store/store.go`

**Beschreibung:**  
Bei `CreateFlag` wird lediglich `key == ""` geprüft. Es gibt keine Begrenzung für die Länge von `key` oder `description`. Das 1-MiB-Body-Limit begrenzt zwar einen einzelnen Request, aber wiederholte Anfragen können den In-Memory-Store mit unnötig großen Objekten füllen und so zu Speicherdruck führen.

**Konkrete Lösung:**  
Sinnvolle Längenlimits validieren und bei Überschreitung mit `400` ablehnen, z. B.:

```go
if utf8.RuneCountInString(req.Key) > 128 {
    writeError(w, http.StatusBadRequest, "key is too long")
    return
}
if len(req.Description) > 4096 {
    writeError(w, http.StatusBadRequest, "description is too long")
    return
}
```

Analog auch in `UpdateFlag` für `Description` und ggf. `Key`.

### 5. Niedrig: Mögliche Log-Injection über URL-Pfad

**Betroffene Stelle:** `internal/api/middleware.go`

**Beschreibung:**  
`accessLog.Printf("%s %s %d", r.Method, r.URL.Path, rec.status)` verwendet `r.URL.Path` als unformatierten `%s`-String. `r.URL.Path` ist URL-dekodiert und kann potenziell Steuerzeichen wie `\n` enthalten (`%0A` im Request-Pfad). Ein Angreifer könnte damit Log-Einträge verfälschen oder eigene Zeilen einschleusen. Die Kernanforderung „nur Methode, Pfad und Status loggen“ bleibt dabei formal erfüllt, aber die Ausgabe kann manipuliert werden.

**Konkrete Lösung:**  
Path mit `%q` maskieren oder Steuerzeichen vor der Ausgabe entfernen:

```go
accessLog.Printf("%s %q %d", r.Method, r.URL.Path, rec.status)
```

oder explizit `strings.Map(...)` zum Entfernen von Steuerzeichen verwenden.

### 6. Niedrig: `decodeJSON` erzwingt weder Content-Type noch vollständigen Body-Konsum

**Betroffene Stelle:** `internal/api/respond.go`

**Beschreibung:**  
`decodeJSON` akzeptiert JSON unabhängig vom `Content-Type` und prüft nicht, ob nach dem ersten JSON-Objekt noch weitere nicht-Whitespace-Daten folgen. `json.Decoder.Decode` liest nur das erste Objekt; ein Body wie `{"key":"x"}{"enabled":true}` würde für den ersten Wert akzeptiert. Das ist derzeit nicht unmittelbar ausnutzbar, kann aber zu Content-Confusion und unerwartetem Verhalten führen.

**Konkrete Lösung:**  
- `Content-Type: application/json` prüfen und bei Abweichung `415`/`400` zurückgeben.
- Nach der ersten Decode sicherstellen, dass kein weiterer nicht-Whitespace-Token folgt, z. B.:

```go
if dec.More() {
    return errors.New("unexpected trailing data")
}
```

oder einen zweiten `Decode`-Versuch auf `io.EOF` prüfen.

### Abschließende Hinweise
- Keine harten Secret-/Token-Funde, keine SQL/Command-Injection, keine unsichere Deserialisierung, kein SSRF und keine XSS-Lücke, da JSON-Ausgaben durch `encoding/json` korrekt escaped werden.
- Die In-Memory-Synchronisierung über `sync.RWMutex` ist korrekt; Evaluierungsanfragen verändern den Store nicht und speichern keine Nutzer-IDs.
- Die oben als „Hoch“ eingestufte fehlende Authentifizierung/Autorisierung begründet den `BLOCKED`-Status und muss vor einem produktiven Betrieb behoben werden.