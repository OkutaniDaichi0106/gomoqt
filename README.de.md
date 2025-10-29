# gomoqt

Eine Go-Implementierung von Media over QUIC Transport (MOQT), die speziell die MOQ Lite-Spezifikation für effizientes Medienstreaming über QUIC implementiert.

[![codecov](https://codecov.io/gh/OkutaniDaichi0106/gomoqt/branch/main/graph/badge.svg?token=4LZCD3FEU3)](https://codecov.io/gh/OkutaniDaichi0106/gomoqt)

## Übersicht

Diese Implementierung folgt der [MOQ Lite-Spezifikation](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html) und bietet eine Grundlage zum Aufbau von Echtzeit-Medien-Streaming-Anwendungen unter Verwendung des QUIC-Transports. MOQ Lite ist eine vereinfachte Version des Media over QUIC Transport-Protokolls, das auf niedrigere Latenz und geringere Komplexität ausgelegt ist, während die Kernvorteile der QUIC-basierten Medienübertragung erhalten bleiben.

## Funktionen

Diese Implementierung enthält:

- **MOQ Lite-Protokoll**: Kernimplementierung der MOQ Lite-Spezifikation
- **WebTransport-Unterstützung**: Volle Unterstützung für WebTransport-Verbindungen in Browsern
- **Roh-QUIC-Unterstützung**: Direkte QUIC-Verbindungen für native Anwendungen
- **Track-Management**: Publisher/Subscriber-Muster zur Verwaltung von Medientracks
- **Multiplexed Streaming**: Effizientes Multiplexing mehrerer Medientracks
- **Beispielanwendungen**: Vollständige Beispiele für verschiedene Anwendungsfälle

- **Echtzeit-Streaming**:
  Minimierte End-to-End-Latenz für interaktive Anwendungsfälle (Live-Events, Konferenzen, Überwachung mit geringer Latenz). Geeignet für Szenarien, in denen Reaktionsfähigkeit für die Benutzererfahrung wichtig ist.

- **Unterbrechungsfreies Streaming**:
  Robuste Wiedergabe bei verschiedenen Netzwerkbedingungen. Basiert auf QUIC/WebTransport-Primitiven, um Stalls zu reduzieren und die Wiederherstellung nach Paketverlust zu verbessern.

- **Effiziente Inhaltsübertragung**:
  Protokolloptimierungen und Multiplexing reduzieren den Verbindungs-Overhead und die Infrastrukturkosten bei vielen gleichzeitigen Zuschauern oder Streams.

- **Nahtlose Wiedergabe**:
  Jitter- und Puffermanagement zur Reduzierung von Rebuffering und für eine gleichmäßige, kontinuierliche Wiedergabe.

- **Optimierte Qualität**:
  Adaptive Übertragungsmuster priorisieren nutzbare Qualität bei eingeschränkter Bandbreite, um ein konsistentes Benutzererlebnis auf allen Gerätetypen zu gewährleisten.

## Komponenten

| Komponente | Beschreibung |
|-----------|------|
| **moqt** | Das zentrale Go-Paket zur Implementierung und Handhabung des Media over QUIC (MOQ)-Protokolls. |
| **moq-web** | TypeScript-Implementierung des MOQ-Protokolls für das Web. |
| **interop** | Tools und Beispiele zur Interoperabilitätsprüfung verschiedener MOQ-Implementierungen über Plattformen hinweg. |

## Entwicklung

### Voraussetzungen

- Go 1.25.0 oder neuer
- [Mage](https://magefile.org/) Build-Tool (Installation mit `go install github.com/magefile/mage@latest`)

### Erste Schritte

1. Repository klonen:
   ```bash
   git clone https://github.com/OkutaniDaichi0106/gomoqt.git
   cd gomoqt
   ```

2. Paket installieren:
   ```bash
   go get github.com/OkutaniDaichi0106/gomoqt
   ```

3. Mage-Tool installieren:
   ```bash
   go install github.com/magefile/mage@latest
   ```

Hinweis: Entwicklungs-Setup-Befehle (dev-setup, Zertifikaterstellung usw.) sind weiterhin über die Justfile verfügbar. Die Kernbefehle zum Bauen (test, lint, fmt, build, clean) wurden nach Mage migriert.

### Entwicklungsbefehle

#### Beispiele ausführen

```bash
# Interop-Server starten
just interop-server

# In einem anderen Terminal den Interop-Client starten
just interop-client
```

#### Codequalität
```bash
# Code formatieren
mage fmt

# Linter ausführen (benötigt golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
mage lint

# Tests ausführen
mage test
```

#### Build & Clean
```bash
# Code bauen
mage build

# Generierte Dateien bereinigen
mage clean

# Verfügbare Befehle anzeigen
mage help
```

### Beispiele

Das Verzeichnis [examples](examples) enthält Beispielanwendungen zur Verwendung von gomoqt:

* **Interop Server und Client** (`interop/`): Interoperabilitätstests zwischen verschiedenen MOQ-Implementierungen
* **Broadcast-Beispiel** (`examples/broadcast/`): Demonstration der Broadcast-Funktionalität
* **Echo-Beispiel** (`examples/echo/`): Einfache Echo-Server- und Client-Implementierung
* **Native QUIC** (`examples/native_quic/`): Beispiele für direkte QUIC-Verbindungen
* **Relay** (`examples/relay/`): Relay-Funktionalität für Medienstreaming

### Dokumentation

* [GoDoc](https://pkg.go.dev/github.com/OkutaniDaichi0106/gomoqt)
* [MOQ Lite-Spezifikation](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html)
* [Implementierungsstatus](moqt/README.md) – Detaillierter Fortschritt der Implementierung

## Spezifikationskonformität

Diese Implementierung zielt auf die MOQ Lite-Spezifikation ab, die einen vereinfachten Ansatz für Media over QUIC Transport bietet. Der aktuelle Implementierungsstatus kann in der [moqt-Paket-README](moqt/README.md) eingesehen werden, die eine detaillierte Nachverfolgung der implementierten Funktionen nach Spezifikationsabschnitten enthält.

## Zum Projekt beitragen

Beiträge sind willkommen! So kannst du helfen:

1. Repository forken.
2. Feature-Branch erstellen (`git checkout -b feature/amazing-feature`).
3. Änderungen durchführen.
4. Codequalität überprüfen:
   ```bash
   mage fmt
   mage lint
   mage test
   ```
5. Änderungen committen (`git commit -m 'Add amazing feature'`).
6. Branch pushen (`git push origin feature/amazing-feature`).
7. Einen Pull Request eröffnen.

## Lizenz

Dieses Projekt ist unter der MIT-Lizenz lizenziert; siehe [LICENSE](LICENSE) für Details.

## Danksagungen

* [quic-go](https://github.com/quic-go/quic-go) – QUIC-Implementierung in Go
* [webtransport-go](https://github.com/quic-go/webtransport-go) – WebTransport-Implementierung in Go
* [MOQ Lite-Spezifikation](https://kixelated.github.io/moq-drafts/draft-lcurley-moq-transfork.html) – Die Spezifikation, der diese Implementierung folgt