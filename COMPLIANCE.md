VERDICT: CHANGES_REQUESTED

## 1. DSGVO

### 1.1 Positiv: keine Speicherung und kein Logging personenbezogener Nutzer-IDs  
Der Query-Parameter `user` wird in `internal/api/flags_evaluate.go` nur für die Hash-Berechnung verwendet.  
Der In-Memory-Store (`internal/store/store.go`) enthält ausschließlich `Flag`-Objekte.  
Die Logging-Middleware (`internal/api/middleware.go`) protokolliert nach sichtbarem Code ausschließlich Methode, Pfad und Statuscode. Query-Parameter und Bodies werden nicht geloggt.  
**Bewertung:** konform. Keine Maßnahme erforderlich.

### 1.2 Cache-Control für die Evaluierungsantwort fehlt  
**Schweregrad:** niedrig  
In `internal/api/flags_evaluate.go` wird die Nutzer-ID im JSON-Response-Objekt `evaluateResponse.User` an den Client zurückgegeben. Da es sich um einen GET-Endpunkt handelt, können Zwischenspeicher die Antwort unter Umständen cachen.  
**Remedium:** In `EvaluateFlag` vor `writeJSON` setzen:
```go
w.Header().Set("Cache-Control", "no-store")
```
Begründung: Verhindert eine unbeabsichtigte Zwischenspeicherung personenbezogener Nutzer-IDs.

### 1.3 Freitext `description` ohne Längenbegrenzung  
**Schweregrad:** niedrig  
In `internal/api/flags_create.go` und `internal/api/flags_update.go` kann `description` beliebig lang sein. Damit können unbeabsichtigt personenbezogene Daten in den Store gelangen und unkontrolliert im Speicher gehalten werden.  
**Remedium:** Längenbegrenzung einfügen, z. B.:
```go
if len(req.Description) > 2000 {
    writeError(w, http.StatusBadRequest, "description too long")
    return
}
```
Zusätzlich in der README/API-Doku dokumentieren, dass `description` keine personenbezogenen Daten enthalten darf.

### 1.4 Rechtsgrundlage für die Verarbeitung der Nutzer-ID  
**Schweregrad:** niedrig  
Der Code verarbeitet die übergebene `user`-ID transient, speichert sie nicht und protokolliert sie nicht. Eine Rechtsgrundlage ist deployment-abhängig und im Code nicht sichtbar.  
**Remedium:** In der Betreiberdokumentation (`README.md`) festhalten, dass die Nutzer-ID ausschließlich zur deterministischen Rollout-Berechnung verarbeitet wird, nicht gespeichert wird und die Verarbeitung im Rahmen eines Vertrags oder berechtigten Interesses erfolgt. Falls der Dienst als Auftragsverarbeiter betrieben wird, ist ein AV-Vertrag erforderlich.

## 2. EU Cyber Resilience Act (CRA)

### 2.1 Mutierende Endpunkte ohne Authentifizierung  
**Schweregrad:** hoch  
POST `/flags`, PUT `/flags/{key}` und DELETE `/flags/{key}` sind ohne sichtbare Authentifizierung oder Autorisierung erreichbar. Jeder Netzwerkteilnehmer kann Feature-Flags anlegen, ändern oder löschen. Das verletzt die CRA-Anforderung an Sicherheit durch Gestaltung und Schutz vor unbefugtem Zugriff.  
**Remedium:** Authentifizierung mindestens für die mutierenden Endpunkte einführen. Beispiel in `internal/api/middleware.go`:
```go
func RequireAuth(next http.Handler) http.Handler {
    token := os.Getenv("FLAGS_API_TOKEN")
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if token == "" || r.Header.Get("Authorization") != "Bearer "+token {
            writeError(w, http.StatusUnauthorized, "unauthorized")
            return
        }
        next.ServeHTTP(w, r)
    })
}
```
Die Tests in `internal/api/*_test.go` und `routing_test.go` müssen dann einen gültigen Token-Header senden. Soll die Authentifizierung bewusst über ein vorgelagertes Gateway erfolgen, ist das im sichtbaren Code ebenfalls zu dokumentieren; aktuell fehlt eine solche sichtbare Dokumentation.

### 2.2 HTTP-Server ohne Timeouts  
**Schweregrad:** mittel  
In `main.go` wird direkt `http.ListenAndServe(addr, handler)` verwendet. Damit laufen die Default-Timeouts, was das Produkt anfällig für langsame Verbindungen und Ressourcenerschöpfung macht.  
**Remedium:** In `main.go` einen expliziten `http.Server` mit Timeouts konfigurieren:
```go
srv := &http.Server{
    Addr:              addr,
    Handler:           handler,
    ReadHeaderTimeout: 5 * time.Second,
    ReadTimeout:       15 * time.Second,
    WriteTimeout:      15 * time.Second,
    IdleTimeout:       60 * time.Second,
    MaxHeaderBytes:    1 << 20,
}
if err := srv.ListenAndServe(); err != nil {
    log.Fatal(err)
}
```
Dafür `time` importieren.

### 2.3 Kein Rate-Limiting sichtbar  
**Schweregrad:** niedrig  
Das Body-Limit (`internal/api/respond.go`) begrenzt einzelne Requests, aber der Dienst hat keinen sichtbaren Schutz gegen eine hohe Anzahl paralleler Anfragen.  
**Remedium:** Entweder ein Rate-Limiting-Middleware ergänzen oder in der `README.md` dokumentieren, dass ein vorgelagertes API-Gateway diese Aufgabe übernimmt. Mindestens die geplanten Betriebsgrenzen dokumentieren.

### 2.4 Abhängigkeiten und SBOM  
**Schweregrad:** niedrig  
`go.mod` ist laut Dateiliste vorhanden und enthält keine externen Frameworks. Damit bestehen derzeit keine problematischen Drittanbieter-Abhängigkeiten.  
**Remedium:** In `README.md` einen kurzen Abschnitt „Security Properties“ ergänzen: Standardbibliothek, keine externen Abhängigkeiten, Body-Limit, keine PII-Logs, Update über normalen Go-Build/Deployment. Dies dokumentiert die geforderten Sicherheitseigenschaften und die Update-/Nachbesserungsfähigkeit.

## 3. EU AI Act

Kein KI-Feature sichtbar. Die Anwendung ist ein deterministischer Feature-Flag-Dienst ohne trainierte Modelle, generative Funktionen oder automatisierte Entscheidungsfindung im Sinne des AI Act.  
**Bewertung:** nicht anwendbar.

## 4. Pflichttexte und UI

Das Produkt ist ein reines Go-Backend ohne Endnutzer-UI. Impressumspflicht, Cookie-Banner und vergleichbare Web-UI-Pflichten sind hier nicht anwendbar.  
**Bewertung:** nicht anwendbar.

## 5. Barrierefreiheit

Keine öffentliche Web-UI vorhanden.  
**Bewertung:** nicht anwendbar.

## Gesamtbild

Die Datenminimierung ist im Kern sauber umgesetzt: Es werden keine Nutzer-IDs gespeichert oder geloggt, Fehlerantworten sind generisch, und das Body-Limit ist vorhanden. Offen sind vor allem CRA-bezogene Sicherheitslücken: fehlende Authentifizierung an mutierenden Endpunkten, fehlende HTTP-Timeouts und fehlende sichtbare Betriebssicherheits-Dokumentation. Diese Lücken sind behebbar und erfordern keine grundsätzliche Neuarchitektur.