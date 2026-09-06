# Feature-Flag-Service REST-API

Ein schlanker Feature-Flag-Service als REST-API in Go. Flags lassen sich anlegen,
listen, einzeln abrufen, ändern und löschen; die Evaluierung liefert pro Nutzer eine
deterministische Ja/Nein-Entscheidung anhand eines stabilen Hashs über Flag-Key und
Nutzer-ID. Die Datenhaltung erfolgt thread-sicher im In-Memory-Store mit Mutex,
ergänzt um Eingabevalidierung, JSON-Fehlerobjekte und Zugriffs-Logging. Der Service
kommt ohne externe Abhängigkeiten aus und nutzt ausschließlich die Go-Standardbibliothek.

## Tech-Stack

- **Sprache**: Go (1.22+)
- **Runtime**: `net/http` aus der Standardbibliothek (kein externes Web-Framework)
- **Build**: `go build ./...`
- **Tests**: `go test ./...` mit `httptest`

## Installation

Voraussetzung: Go 1.22 oder neuer installiert. Keine weiteren Abhängigkeiten.

```sh
git clone <repo-url>
cd <repo>
```

## Ausführen (Entwicklung)

```sh
go run .
```

Der Server lauscht standardmäßig auf Port `8080`. Über die Umgebungsvariable `PORT`
lässt sich ein anderer Port wählen:

```sh
PORT=9000 go run .
```

## Build für Produktion

```sh
go build -o featureflags .
./featureflags
```

## Endpunkte

| Methode | Pfad                        | Beschreibung                                          |
|---------|-----------------------------|-------------------------------------------------------|
| POST    | `/flags`                    | Flag anlegen (201; 400/409 bei Fehler)                |
| GET     | `/flags`                    | Alle Flags auflisten (200; ohne Flags `[]`)           |
| GET     | `/flags/{key}`              | Einzelnes Flag abrufen (200; 404 unbekannter Key)     |
| PUT     | `/flags/{key}`              | Flag ändern (200; 400/404 bei Fehler)                 |
| DELETE  | `/flags/{key}`              | Flag löschen (204; 404 unbekannter Key)               |
| GET     | `/flags/{key}/evaluate`     | Deterministische Evaluierung für `?user={id}`         |
| GET     | `/healthz`                  | Health-Check (200, `{"status":"ok"}`)                 |

Fehlerantworten haben immer das Format `{"error": "<text>"}`. Der Request-Body ist
auf 1 MiB begrenzt; das Zugriffs-Log protokolliert ausschließlich Methode, Pfad und
Statuscode.

## Konfiguration

| Variable                   | Beschreibung                                                    | Default |
|----------------------------|-----------------------------------------------------------------|---------|
| `PORT`                     | Port, auf dem der Server lauscht                                | `8080`  |
| `FEATUREFLAGS_API_TOKEN`   | Bearer-Token für mutierende Endpunkte (POST/PUT/DELETE)          | –       |

## Features

- In-Memory-Speicherung der Flags (thread-sicher mit `sync.RWMutex`)
- CRUD-Endpunkte für Flags mit JSON-Ein-/Ausgabe
- Deterministische Evaluierung pro Nutzer
- Eingabevalidierung mit 1-MiB-Body-Limit
- JSON-Fehlerobjekte ohne interne Details
- Zugriffs-Logging (nur Methode, Pfad, Statuscode)

## Security Properties

- **Nur Standardbibliothek**: Der Service nutzt ausschließlich die
  Go-Standardbibliothek (`net/http`); es gibt keine externen
  Laufzeit-Abhängigkeiten, die gepflegt oder auf Sicherheitslücken überprüft
  werden müssten.
- **1-MiB-Body-Limit**: Eingehende JSON-Request-Bodies werden über einen
  `LimitedReader` auf maximal 1 MiB begrenzt und bei Überschreitung abgelehnt
  (Status 400), bevor der Body vollständig gepuffert wird.
- **Keine PII in Logs**: Das Zugriffs-Log protokolliert ausschließlich Methode,
  Pfad und Statuscode — niemals Query-Parameter (insbesondere nicht den
  `user`-Parameter) oder Request-Bodies.
- **Geschützte mutierende Endpunkte**: Die mutierenden Endpunkte
  (`POST /flags`, `PUT /flags/{key}`, `DELETE /flags/{key}`) sind über einen
  Bearer-Token geschützt. Der Token wird aus der Umgebungsvariable
  `FEATUREFLAGS_API_TOKEN` gelesen und im `Authorization`-Header als
  `Bearer <token>` übermittelt.
- **Updates über normalen Go-Build**: Aktualisierungen erfolgen über den
  regulären Go-Build-Prozess (`go build ./...`); es sind keine besonderen
  Deployment-Werkzeuge oder Build-Schritte erforderlich.

## Datenschutz

### Verarbeitung der Nutzer-ID

Die Nutzer-ID (`?user={id}`) wird **ausschließlich** für die deterministische
Rollout-Berechnung verwendet: Zusammen mit dem Flag-Key fließt sie in einen
FNV-1a-Hash ein, dessen Ergebnis bestimmt, ob das Flag für diesen Nutzer
aktiviert ist. Die Nutzer-ID wird dabei **nicht gespeichert** — der
In-Memory-Store enthält ausschließlich Flag-Objekte, und Evaluierungsanfragen
verändern den Store nicht und protokollieren keine Nutzer-IDs oder
Evaluierungsergebnisse.

- **Rechtsgrundlage**: Vertragserfüllung (Art. 6 Abs. 1 lit. b DSGVO) bzw.
  berechtigtes Interesse (Art. 6 Abs. 1 lit. f DSGVO) an der konsistenten
  Auslieferung von Feature-Flags.
- **Auftragsverarbeitung**: Wird der Service als Auftragsverarbeiter betrieben,
  ist ein Vertrag zur Auftragsverarbeitung (AV-Vertrag, Art. 28 DSGVO)
  erforderlich.

### Feldbeschränkungen

- `description` darf **keine personenbezogenen Daten** enthalten und ist auf
  maximal **2000 Zeichen** begrenzt.
- `key` darf maximal **128 Zeichen** lang sein.

## Betrieb & Infrastruktur

TLS-Terminierung und Rate-Limiting erfolgen nicht im Service selbst, sondern
über ein vorgelagertes Reverse-Proxy bzw. API-Gateway. Der Service lauscht
unverschlüsselt auf dem konfigurierten Port; HTTPS und Zugriffsbegrenzung
müssen am vorgelagerten Gateway eingerichtet werden.
