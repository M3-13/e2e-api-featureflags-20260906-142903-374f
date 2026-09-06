VERDICT: CHANGES_REQUESTED

## Prüfrahmen

Geprüft wurde der sichtbare Stand des Go-Backends „Feature-Flag-Service REST-API“. Es handelt sich um ein reines Backend ohne Endbenutzer-Web-UI. Damit entfallen die Pflichten zu Impressum, Cookie-Banner, AGB/Widerrufsbelehrung und die unmittelbaren Barrierefreiheitspflichten einer öffentlichen Web-UI. Relevant sind dagegen DSGVO/GDPR sowie der Cyber Resilience Act, soweit das Produkt als digitale Lösung in Verkehr gebracht oder betrieben wird.

Die Prüfung stützt sich ausschließlich auf den sichtbaren Code und die Sprint-Spezifikation. Die vorhandenen Dateien `COMPLIANCE.md` und `SECURITY.md` wurden nicht inhaltlich vorgelegt und daher nicht als erfüllte Anforderung gewertet.

---

## 1. DSGVO / GDPR

### Positiv sichtbar

- `internal/api/middleware.go` protokolliert ausschließlich Methode, Pfad und Statuscode. Der Test `internal/api/middleware_test.go` belegt, dass weder Query-Parameter wie `user` noch Body-Inhalte im Log erscheinen.
- `internal/store/store.go` speichert nur `Flag`-Objekte. Evaluierungsanfragen verändern den Store nicht und speichern keine Nutzer-IDs oder Ergebnisse.
- `internal/api/flags_evaluate.go` setzt bei der Evaluierung `Cache-Control: no-store`, was Caching personenbeziehbarer Antworten verhindert.
- Fehlerantworten sind generisch und enthalten keine internen Fehlermeldungen oder Stacktraces.
- Es gibt eine Body-Begrenzung auf 1 MiB, was Datenminimierung und Ressourcenschutz dient.

### Befund GDPR-1 — Mittel

**Titel:** Rechtsgrundlage und Datenschutzdokumentation fehlen im sichtbaren Stand.

**Beschreibung:** Der Parameter `user` in `GET /flags/{key}/evaluate` ist ein personenbeziehbarer Identifikator, auch wenn er nur transient für die Hash-Berechnung verwendet wird. Die DSGVO verlangt für die Verarbeitung eine dokumentierte Rechtsgrundlage sowie geeignete Informationen für die betroffenen Personen. Im sichtbaren Code und in der Spec ist dafür keine Dokumentation erkennbar.

**Maßnahme:**
- In `COMPLIANCE.md` oder `README.md` einen Abschnitt „Datenverarbeitung / DSGVO“ ergänzen.
- Beispieltext:
  - „Der Endpunkt `GET /flags/{key}/evaluate` verarbeitet den Query-Parameter `user` als personenbeziehbaren Identifikator ausschließlich flüchtig im Arbeitsspeicher zur deterministischen Rollout-Entscheidung. Der Wert wird nicht gespeichert, nicht geloggt, nicht an Dritte übermittelt und nach der Antwort verworfen. Rechtsgrundlage ist Art. 6 Abs. 1 lit. b DSGVO, soweit die Verarbeitung zur Erfüllung eines Vertrags mit dem Kunden erfolgt, oder Art. 6 Abs. 1 lit. f DSGVO auf Basis des berechtigten Interesses an einer funktionsfähigen, pseudonymen Feature-Auslieferung. Soweit der Betreiber als Auftragsverarbeiter handelt, ist ein Auftragsverarbeitungsvertrag nach Art. 28 DSGVO erforderlich.“
- Zusätzlich dokumentieren, dass keine Nutzerprofile, keine persistente Speicherung und keine automatisierten Einzelfallentscheidungen nach Art. 22 DSGVO erfolgen.

### Befund GDPR-2 — Niedrig

**Titel:** Unnötiges Echo des `user`-Parameters in der API-Antwort.

**Beschreibung:** `internal/api/flags_evaluate.go` gibt den personenbezogenen Parameter `user` über das Feld `evaluateResponse.User` unverändert an den Aufrufer zurück. Für die Funktionalität ist das nicht erforderlich; Datenminimierung nach Art. 5 Abs. 1 lit. c DSGVO spricht dagegen.

**Maßnahme:**
- In `internal/api/flags_evaluate.go` das Feld `User string` aus `evaluateResponse` entfernen und nur `Enabled` und `Key` zurückgeben.
- In `internal/api/flags_evaluate_test.go` den Test `TestEvaluateDeterministicSameUser` entsprechend anpassen, sodass kein `body.User` mehr erwartet wird.
- Diese Änderung bricht keine legitime Produktanforderung: Die ACs verlangen lediglich eine deterministische Ja/Nein-Entscheidung, kein User-Echo.

---

## 2. EU Cyber Resilience Act

### Positiv sichtbar

- Keine externen Web-Frameworks/Router; die Implementierung nutzt die Go-Standardbibliothek.
- `main.go` setzt serverhärtende Timeouts (`ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, `IdleTimeout`, `MaxHeaderBytes`).
- Eingaben werden validiert, JSON-Bodies werden größenbegrenzt, und mutierende Endpunkte sind per Bearer-Token abgesichert.
- Angriffsfläche und Datenverarbeitung sind durch den flüchtigen In-Memory-Store klein.

### Befund CRA-1 — Hoch

**Titel:** Keine TLS-Absicherung des HTTP-Servers sichtbar.

**Beschreibung:** `main.go` startet ausschließlich `server.ListenAndServe()`. Dadurch wird der Bearer-Token bei mutierenden Anfragen sowie die Evaluierungsantworten bei Netzwerkexposition im Klartext übertragen. Der Default-Bind `127.0.0.1` mindert das Risiko, aber sobald `HOST` auf eine nicht-lokale Schnittstelle gesetzt wird, besteht ein erhebliches Risiko. Der Cyber Resilience Act verlangt Sicherheit durch Standard und Vorgabe, insbesondere Schutz der Vertraulichkeit übertragener Daten.

**Maßnahme:**
- In `main.go` einen TLS-Pfad ergänzen, z. B. `ListenAndServeTLS` mit konfigurierbaren Umgebungsvariablen `TLS_CERT_FILE` und `TLS_KEY_FILE`.
- Alternativ in `README.md`/`SECURITY.md` verbindlich dokumentieren, dass der Dienst ausschließlich hinter einer TLS-terminierenden Reverse-Proxy-Schicht betrieben werden darf.
- Den Default `HOST=127.0.0.1` beibehalten und in der Dokumentation klarstellen, dass `HOST=0.0.0.0` oder eine öffentliche Adresse nur mit TLS zulässig ist.

### Befund CRA-2 — Mittel

**Titel:** Keine SBOM-/Lieferketten-Dokumentation sichtbar.

**Beschreibung:** Der Cyber Resilience Act verlangt für Produkte mit digitalen Elementen eine nachvollziehbare Software-Stückliste oder mindestens eine dokumentierte Abhängigkeitsliste. Die Datei `go.mod` existiert und ist laut Spec klein und ohne externe Module, aber eine maschinenlesbare SBOM oder eine dokumentierte Lieferketten-Übersicht ist im sichtbaren Stand nicht vorhanden.

**Maßnahme:**
- In der Build-/CI-Pipeline eine SBOM erzeugen, z. B. mit `cyclonedx-gomod` oder einem SPDX-Generator, und als Artefakt z. B. `sbom.cdx.json` bereitstellen.
- In `SECURITY.md` oder `COMPLIANCE.md` festhalten: „Dieses Release verwendet ausschließlich die Go-Standardbibliothek; die SBOM ist im Build-Artefakt enthalten.“
- Da keine externen Abhängigkeiten sichtbar sind, ist das Risiko niedrig, aber die Dokumentationspflicht bleibt bestehen.

### Befund CRA-3 — Mittel

**Titel:** Sicherheits- und Update-/Patch-Dokumentation nicht sichtbar.

**Beschreibung:** Im sichtbaren Stand fehlt eine dokumentierte Beschreibung der Sicherheitseigenschaften sowie des Update- und Patch-Prozesses. Der Cyber Resilience Act verlangt dokumentierte Sicherheitseigenschaften und eine klar geregelte Bereitstellung von Sicherheitsupdates.

**Maßnahme:**
- In `SECURITY.md` einen Abschnitt „Security properties“ ergänzen, z. B.:
  - Body-Limit 1 MiB
  - Mutierende Endpunkte per Bearer-Token geschützt
  - Kein Logging von Query-Parametern oder Bodies
  - Server-Timeouts gesetzt
  - Kein persistentes Speichern von Nutzer-IDs
- In `SECURITY.md` einen Abschnitt „Update/Patch-Prozess“ ergänzen: Zuständigkeit, Release-Kanal, Patch-Zeitraum, Meldung von Schwachstellen.

### Befund CRA-4 — Niedrig

**Titel:** Token-Vergleich nicht in konstanter Zeit.

**Beschreibung:** `internal/api/middleware.go` vergleicht `auth != "Bearer "+token` mit normalem String-Vergleich. Bei einem netzwerkexponierten Dienst kann das theoretisch Timing-Seitenkanäle ermöglichen.

**Maßnahme:**
- In `internal/api/middleware.go` den Vergleich auf `crypto/subtle.ConstantTimeCompare` oder `hmac.Equal` umstellen, z. B.:
  - `subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+token)) == 1`
- Bereits vorhandene Tests weiterverwenden; das Verhalten bleibt identisch.

### Befund CRA-5 — Niedrig

**Titel:** Keine Ratenbegrenzung/Bruteforce-Abwehr an den authentifizierten Endpunkten sichtbar.

**Beschreibung:** Die mutierenden Endpunkte sind mit einem statischen Token geschützt, aber ohne sichtbare Begrenzung der Anfragen pro Zeiteinheit. Bei öffentlicher Exposition ist das ein zusätzliches Risiko.

**Maßnahme:**
- Optional eine kleine Rate-Limit-Middleware ergänzen, die auf Client-IP oder Authentifizierungskennung basiert und produktiv über ein vorgelagertes Gateway konfigurierbar bleibt.
- Wichtig dabei: Die legitime Evaluierungs- und Verwaltungsnutzung muss möglich bleiben; das Rate-Limit darf nicht so streng sein, dass der bestimmungsgemäße Betrieb scheitert.

---

## 3. EU AI Act

**Befund AI-1 — Keine Relevanz erkennbar (Information).**

Im sichtbaren Produkt ist keine KI-Funktion, kein maschinelles Lernen, kein automatisiertes Entscheidungssystem im Sinne des AI Act enthalten. Die Rollout-Entscheidung ist deterministischer Hash-basierter Code, keine KI. Es ergeben sich derzeit keine Pflichten aus dem EU AI Act.

---

## 4. Pflichttexte und Web-UI

**Befund UI-1 — Keine Relevanz (Information).**

Als reines `go-backend` ohne Endbenutzer-UI bestehen keine Verpflichtungen zu Impressum, Cookie-Banner, Nutzungsbedingungen/AGB oder gesetzlichen Widerrufsbelehrungen. Die API-Antworten sind JSON; es gibt keine Webseite. Deshalb ist dieser Prüfbereich nicht anwendbar.

---

## 5. Barrierefreiheit

**Befund ACCESS-1 — Keine Relevanz (Information).**

Als reines Backend ohne öffentliche Web-UI ist die WCAG/BITV/EAA-Prüfung nicht anwendbar. Sollte später ein Admin-Web-UI ergänzt werden, wären WCAG 2.1 AA und der European Accessibility Act zu beachten.

---

## Zusammenfassung

Der sichtbare Code erfüllt die wichtigsten Datenschutzanforderungen im engeren Sinne: kein Logging personenbezogener Daten, keine persistente Speicherung von Nutzer-IDs, generische Fehlerantworten und ein begrenzter, flüchtiger Verarbeitungszweck. Ein fundamentaler DSGVO-Verstoß, der eine Blockierung rechtfertigen würde, liegt nicht vor.

Behebbar sind vor allem:

- **CRA-1:** TLS-Absicherung bzw. dokumentierte TLS-Terminierung.
- **CRA-2/CRA-3:** SBOM- und Sicherheits-/Update-Dokumentation.
- **GDPR-1:** Dokumentierte Rechtsgrundlage für die transiente Verarbeitung des `user`-Parameters.
- **GDPR-2:** Entfernen des unnötigen User-Echos aus der Evaluierungsantwort.
- **CRA-4/CRA-5:** Härtungsmaßnahmen mit geringem Aufwand.

Diese Punkte sind durch konkrete Code- und Dokumentationsänderungen lösbar und stehen der Produktfunktion nicht entgegen. Daher: **CHANGES_REQUESTED**.