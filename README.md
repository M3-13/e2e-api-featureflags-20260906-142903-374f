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

| Variable | Beschreibung                        | Default |
|----------|-------------------------------------|---------|
| `PORT`   | Port, auf dem der Server lauscht    | `8080`  |

## Features

- In-Memory-Speicherung der Flags (thread-sicher mit `sync.RWMutex`)
- CRUD-Endpunkte für Flags mit JSON-Ein-/Ausgabe
- Deterministische Evaluierung pro Nutzer
- Eingabevalidierung mit 1-MiB-Body-Limit
- JSON-Fehlerobjekte ohne interne Details
- Zugriffs-Logging (nur Methode, Pfad, Statuscode)
