# CrateDesk — Funktionsspezifikation

> **Zweck dieses Dokuments:** Beschreibt pro Kernfunktion, was die Software leisten muss (SOLL), was der Code heute tatsächlich tut (IST) und welche konkreten Ausbauschritte nötig sind. Geschrieben, damit eine KI einzelne Kernfunktionen eigenständig fertigstellen kann, ohne das gesamte Projekt zu überblicken.
>
> **Stand:** 07.09.2026 · Analyse-Grundlage: Quellcode der Repositories `CrateDesk` (Commit `dc852c0`) und `CrateDesk_scanner` (Commit `8c6cab2`).
>
> **Wichtig:** Alle IST-Aussagen stammen aus gelesenem Code, nicht aus der vorhandenen Projektdokumentation. Wo README, `IMPLEMENTATION.md` oder Architekturdokumente dem widersprechen, gilt dieses Dokument — die Doku beschreibt vielfach Absichten, nicht Zustände. Es wurde nichts zur Laufzeit ausgeführt; Aussagen über Laufzeitfehler sind statisch aus dem Code abgeleitet.

---

## Inhalt

- [So ist dieses Dokument zu benutzen](#so-ist-dieses-dokument-zu-benutzen)
- [Teil A — Produktverständnis](#teil-a--produktverständnis)
- [Teil B — Die 13 Kernfunktionen](#teil-b--die-13-kernfunktionen)
- [Teil C — Querschnittsbefunde](#teil-c--querschnittsbefunde)
- [Teil D — Priorisierte Roadmap](#teil-d--priorisierte-roadmap)
- [Teil E — Blocker-Anhang mit Fundstellen](#teil-e--blocker-anhang-mit-fundstellen)

---

## So ist dieses Dokument zu benutzen

Jede Kernfunktion in Teil B hat denselben Aufbau:

| Abschnitt | Inhalt |
|---|---|
| **Zweck** | Wofür die Funktion fachlich da ist |
| **Fachregeln (SOLL)** | Was gelten muss, damit ein Betrieb real damit arbeiten kann |
| **Ist-Zustand** | Was der Code heute tut, mit Fundstellen |
| **Ausbauaufträge** | Nummerierte, einzeln abarbeitbare Arbeitspakete |

**Statusmarker:**

| Marker | Bedeutung |
|---|---|
| ✅ | Funktioniert wie beschrieben |
| ⚠️ | Teilweise gebaut, aber unvollständig oder fachlich falsch |
| 🔌 | **Fertig gebaut, aber nicht angeschlossen** — der günstigste Fix |
| ❌ | Existiert nicht oder ist eine Attrappe |
| 🚫 | Blockiert die Nutzung anderer Funktionen |

Ein Ausbauauftrag ist so formuliert, dass er als einzelner Task an eine KI übergeben werden kann. Abhängigkeiten sind explizit genannt.

---

# Teil A — Produktverständnis

## Was das Produkt ist

Eine **selbst gehostete Vermiet- und Rechnungssoftware für Vermietbetriebe** in der DACH-Region. Ursprünglich für Veranstaltungstechnik, inzwischen über Branchenprofile auf 19 Branchen erweitert (Film & TV, Baumaschinen, Werkzeugverleih, Gerüstbau, Messebau, Gastro, Fahrzeugvermietung u. a.).

Positionierung gegen Rentman und Easyjob: keine Lizenzkosten, kein Vendor Lock-in, Daten auf eigenem Server, DATEV-Anbindung, und als Alleinstellungsmerkmal eine **Peer-to-Peer-Federation** für unternehmensübergreifendes Equipment-Sharing ohne zentrale Instanz.

Zielgröße laut Planung: 1–100 Mitarbeiter.

## Der Kern-Kreislauf

```
Artikelstamm  →  Angebot  →  Auftrag/Projekt  →  Packliste
                                                     ↓
Rechnung  ←  Rückgabe/Prüfung  ←  Einsatz  ←  Check-Out (Scanner)
    ↓
DATEV / Buchhaltung
```

Quer dazu: Personal, Transport, Wartung, Fehlmengen/Zumietung.

## Die zentrale Einsicht

Eine Vermietsoftware ist **kein Lagerverwaltungssystem**. Ein Lagersystem beantwortet „wie viel liegt da?". Eine Vermietsoftware muss beantworten:

> **„Wie viel habe ich am 14. Juni zwischen 08:00 und 22:00 noch frei — unter Berücksichtigung aller Reservierungen, Rüstzeiten, Wartungssperren und Zumietungen?"**

Das ist eine **zeitraumbezogene Verfügbarkeitsrechnung**. Sie ist das Herzstück, an dem Angebot, Packliste, Fehlmengen, Zumietung und Auslastungs-KPI hängen.

**Diese Rechnung existiert im Code nicht.** Die Prüfung nimmt hart „1 Stück" an und kennt keine Datumsparameter (`services/inventory-service/internal/application/availability_service.go:62-64`). Solange sie fehlt, sind alle darauf aufbauenden Funktionen Fassade.

Deshalb ist sie Kernfunktion 1 und Voraussetzung für die Kernfunktionen 3, 4, 9 und 12.

## Das durchgängige Muster

Über alle vier analysierten Domänen zieht sich ein Befund:

> **Der Code ist deutlich besser, als er funktioniert.**

Session-Manager, Digest-Versand, Workflow-Aktionsausführung, PDF-Generator, Zertifikatsverwaltung, Permission-Modell, Equipment-Historien-Repository, ZPL-Etikettendienst — alle sauber gebaut, keiner davon angeschlossen. Das ist typisch für schnell parallelisierte Entwicklung: Bausteine entstehen isoliert und werden am Ende nicht zusammengesteckt.

Die praktische Konsequenz: **Anschließen ist billiger als Bauen.** Ein erheblicher Teil des Funktionsumfangs lässt sich mit wenigen Zeilen Verdrahtung aktivieren (siehe Teil D, Phase 0).

---

# Teil B — Die 13 Kernfunktionen

---

## 1. Verfügbarkeits- und Dispositionsengine 🚫

### Zweck

Für jeden Artikel und jeden Zeitraum die freie Menge berechnen. Grundlage für Angebot, Packliste, Fehlmengen und Auslastung.

### Fachregeln (SOLL)

1. **Mengen statt Einzelstücke.** Ein Artikeltyp hat eine Bestandsmenge; Seriennummern sind optional darunter. Eine Reservierung muss ein `quantity`-Feld haben.
2. **Zeitraum-Overlap halboffen:** `start_a < end_b AND end_a > start_b`. Ende gleich Beginn ist **kein** Konflikt.
3. **Rüst- und Pufferzeiten** pro Kategorie konfigurierbar (Reinigung, Prüfung, Transportweg). Ein Gerät ist nicht sofort nach Rückgabe wieder verfügbar.
4. **Verfügbarkeitsstufen:**
   - `frei` — buchbar
   - `vorgemerkt` — Angebot blockt weich, verfällt mit Angebotsablauf
   - `reserviert` — Auftrag blockt hart
   - `ausgegeben` — physisch draußen
5. **Sperren einrechnen:** Wartung, defekt, verloren, ausgesondert reduzieren den verfügbaren Bestand.
6. **Überbuchung bewusst erlauben**, aber als Fehlmenge markieren — nicht stillschweigend zulassen.
7. **Race-Sicherheit:** Zwei gleichzeitige Buchungen des letzten Geräts dürfen nicht beide durchgehen. Absicherung gehört in die Datenbank, nicht nur in die Anwendung.
8. Zumietware (Federation) zählt als temporärer Bestand mit Kosten.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| Zeitraumprüfung | ❌ Endpunkt nimmt keine Datumsparameter entgegen | `inventory-service/internal/application/availability_service.go:41-99` |
| Bestandsmenge | ❌ `totalQty` hart auf `1` gesetzt | `availability_service.go:64` |
| Reservierungsmenge | ❌ `Reservation` hat kein `Quantity`-Feld | `project-service/internal/domain/reservation.go` |
| Overlap-Prüfung | ⚠️ Existiert isoliert im project-service, ohne Bestandsbezug | `reservation_postgres.go:126-136` |
| Race-Schutz | ❌ Rein applikativ, kein EXCLUDE-Constraint | `project-service/migrations/003_create_reservations.sql` |
| Überbuchung | ⚠️ `force`-Flag hebelt Prüfung aus, ohne Fehlmenge zu erzeugen | `reservation_service.go:37` |
| Rüstzeiten | ❌ Nicht vorhanden | — |
| Wartungssperre | ⚠️ Sperre wird gesetzt, Verfügbarkeit ignoriert sie | `maintenance-service/…/equipment_http_client.go:53-56` |

Zusätzlich: `FindOverlapping` ist implementiert (`reservation_postgres.go:164-199`) und hat keinen Aufrufer.

### Ausbauaufträge

**1.1 — Datenmodell auf Mengen umstellen**
Reservierung um `quantity INT NOT NULL DEFAULT 1` erweitern. Artikelbestand als eigenes Konzept einführen: entweder `equipment_types.stock_quantity` oder Zählung der zugehörigen Einzelstücke. Migration inkl. Backfill bestehender Reservierungen auf `1`.

**1.2 — Verfügbarkeits-Endpunkt neu schneiden**
`POST /api/v1/availability/check` mit Body `{items:[{equipment_type_id, quantity}], start, end, exclude_project_id?}`. Antwort pro Position: `{available, reserved, blocked, shortage}`. Der alte Endpunkt `GET /api/v1/equipment/{id}/availability` bleibt als Kompatibilitätsschicht bestehen.

**1.3 — Kernabfrage implementieren**
Verfügbare Menge = Gesamtbestand − Summe überlappender Reservierungen − gesperrte Stücke. Overlap halboffen. Vorgemerkt und reserviert getrennt ausweisen. Abhängig von 1.1.

**1.4 — Rüstzeiten**
`turnaround_hours` je Kategorie in `tenant_config`. In die Overlap-Berechnung einbeziehen: Reservierungsende + Rüstzeit blockt weiter.

**1.5 — Race-Sicherheit**
PostgreSQL-`EXCLUDE`-Constraint mit `btree_gist` über `(equipment_type_id WITH =, tstzrange(start,end) WITH &&)`, oder Serialisierung über `SELECT ... FOR UPDATE` auf eine Bestandszeile innerhalb einer echten Transaktion. Abhängig von 1.1.

**1.6 — Überbuchung sichtbar machen**
Bei `force: true` einen Fehlmengensatz erzeugen statt stillschweigend zu buchen. Abhängig von Kernfunktion 9.

---

## 2. Artikelstamm und Lagerstruktur ⚠️

### Zweck

Wissen, was man besitzt, was es kostet und wo es liegt.

### Fachregeln (SOLL)

1. **Artikeltyp vs. Einzelstück** trennen. Typ trägt Preis, Gewicht, Maße, Ersatzwert. Einzelstück trägt Seriennummer, Barcode, RFID, Zustand, Historie.
2. **Kits/Flightcases** als buchbare Einheit, deren Inhalt mit-reserviert und beim Check-In auf Vollständigkeit geprüft wird.
3. **Verbrauchsmaterial** mengenmäßig ohne Einzelverfolgung.
4. **Preise mandantenkonfigurierbar:** Tages-/Wochensatz, Staffelrabatte, Kundenkonditionen, Saison.
5. **Lagerorte hierarchisch:** Standort → Raum → Regal → Fach, jeder Platz mit eigenem Barcode.
6. **Etiketten** für Geräte und Lagerplätze, für Thermodrucker (ZPL) und A4.
7. **Zustandshistorie** je Einzelstück, lückenlos.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| Artikeltyp/Einzelstück | ✅ Sauber getrennt, inkl. Typ→Instanz-Generierung | `inventory-service/internal/domain/equipment_type.go` |
| Flightcases | ⚠️ CRUD vorhanden, keine Mit-Reservierung des Inhalts | `domain/flightcase.go` |
| Preis-Engine | ⚠️ Staffelrabatte (5/10/15/20 %) hart im Code | `application/price_engine.go:60-69` |
| Steuersatz | ⚠️ `0.19` hart im Handler | `adapters/http/handlers.go:782` |
| Wochensatz | ❌ `RentalPriceWeek` wird von der Engine nicht verwendet | `price_engine.go` |
| Lagerstruktur | 🚫 **Zwei parallele, unverbundene Modelle** | siehe unten |
| Etiketten Backend | 🔌 ZPL-Dienst gebaut, kein Endpunkt ruft ihn auf | `warehouse-service/…/zpl_service.go` |
| Etiketten Frontend | ⚠️ Erzeugt Etiketten clientseitig per jsPDF, ignoriert Backend | `frontend/…/EquipmentLabels.tsx` |
| Zustandshistorie | 🔌 Repository gebaut, nie instanziiert; Endpunkt gibt leere Liste | `handlers.go:1077-1087` |
| Bildupload | ❌ StorageAdapter wird als `nil` übergeben | `cmd/server/main.go:52` |

**Der Lagerstruktur-Blocker:** Der warehouse-service enthält zwei Modelle nebeneinander — einen flachen `Location`-Baum (`site/building/room/rack/shelf/bin`) und eine fünfstufige Hierarchie (`Warehouse→Zone→Rack→Bay→StockLocation`). Sie sind nicht miteinander verbunden. Gravierender: Es existieren **zwei widersprüchliche Migrationsverzeichnisse**. Angewendet wird `migrations/` (fünfstufiges Modell), die Repositories sprechen aber gegen `internal/infrastructure/migrations/` (flaches Modell). Damit laufen **alle Lagerplatz- und Inventur-Endpunkte auf Laufzeitfehler**.

Zusätzlich: `equipment.location_id` ist ein freier String ohne Fremdschlüssel gegen eines der beiden Modelle.

### Ausbauaufträge

**2.1 — Lagerstruktur-Konflikt auflösen** 🚫
Entscheidung für **ein** Modell treffen (Empfehlung: die fünfstufige Hierarchie, weil das angewendete Schema sie bereits abbildet). Das unterlegene Modell samt Repositories und Migrationsverzeichnis entfernen. Repositories auf das verbleibende Schema umschreiben. `internal/infrastructure/migrations/` löschen, damit keine Verwechslung mehr möglich ist. **Blockiert 2.2, 2.5 und Kernfunktion 5.**

**2.2 — Lesende Endpunkte für die Hierarchie**
GET/PUT/DELETE für Zone, Rack, Bay, StockLocation ergänzen — aktuell ist das Modell nur beschreibbar, nicht pflegbar. Abhängig von 2.1.

**2.3 — Preise konfigurierbar machen**
Rabattstaffel und Steuersatz aus `auth.tenant_config` lesen statt aus Konstanten. Wochensatz in die Preisberechnung einbeziehen (Regel: ab 7 Tagen Wochensatz statt 7× Tagessatz, sofern günstiger).

**2.4 — Zustandshistorie aktivieren** 🔌
`EquipmentHistoryRepository` in `main.go` instanziieren und an den `EquipmentService` binden. Bei jeder Statusänderung, Zustandsänderung und Standortänderung einen Satz schreiben. Den Platzhalter-Handler durch echte Abfrage ersetzen.

**2.5 — Etikettendruck zusammenführen**
ZPL-Endpunkte für Gerät und Lagerplatz freischalten. Frontend so umbauen, dass es bei vorhandenem Thermodrucker das Backend nutzt und nur als Fallback clientseitig erzeugt. Abhängig von 2.1.

**2.6 — Flightcase-Inhalt mit-reservieren**
Beim Reservieren eines Flightcases den gesamten Inhalt mit-blocken; beim Check-In Vollständigkeit prüfen und Abweichungen melden. Abhängig von Kernfunktion 1.

**2.7 — Bildupload verdrahten**
`StorageAdapter` (lokal oder NAS) in `main.go` instanziieren und übergeben.

---

## 3. Angebot → Auftrag → Projekt ⚠️

### Zweck

Die kaufmännische Kette ohne Medienbruch abbilden.

### Fachregeln (SOLL)

1. Angebot mit Positionen, Mietzeitraum, Rabatt, Gültigkeit — erzeugt **weiche Reservierungen**.
2. **Auftragsbestätigung erzeugt automatisch das Projekt** und wandelt weiche in harte Reservierungen.
3. Angebotsstatus nach Umwandlung auf `converted` setzen — eine erneute Umwandlung muss ausgeschlossen sein.
4. Projekt-Statusmodell mit erzwungenen Übergängen, auch beim Anlegen und Ändern.
5. **Nachträgliche Änderungen** ziehen Reservierungen, Packliste und Rechnungsentwurf konsistent nach.
6. **Dry Hire vs. Full Service** als Projektart, die steuert, welche Module greifen.
7. Projekt kopieren muss Packlisten und Reservierungen mitkopieren.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| Projekt-Statusmodell | ✅ Sauber definiert, strikte Übergänge | `project-service/internal/domain/project.go:112-146` |
| Statusmodell erzwungen | ❌ `CreateProject` erlaubt beliebigen Initialstatus | `project_service.go:78-83` |
| Statusmodell bei Update | ❌ `UpdateProject` umgeht Domain-Methoden, keine Datumsvalidierung | `project_service.go:107-136` |
| Angebot → Projekt | ❌ `ConfirmQuote` legt kein Projekt an | `invoice-service/…/quote_service.go:227-248` |
| Angebot → Rechnung | ⚠️ Funktioniert, aber Angebotsstatus wird nicht gesetzt → beliebig oft wiederholbar | `quote_service.go:274-363` |
| Weiche Reservierung | ❌ Nicht vorhanden | — |
| Projekt kopieren | ⚠️ Kopiert nur Stammdaten, keine Packliste/Reservierung; Frontend sendet Pflichtfelder nicht → immer Fehler | `project_service.go:304-353`, `frontend/…/api.ts` |
| Dry Hire | ✅ Als Flag vorhanden, steuert Unterschriftspflicht im Scanner | `domain/project.go` |
| Default-Währung | ⚠️ `"USD"` in einem DACH-Produkt | `domain/project.go:57` |

### Ausbauaufträge

**3.1 — Auftragsbestätigung erzeugt Projekt**
`ConfirmQuote` erweitern: HTTP-Call an project-service, Projekt aus Angebotsdaten anlegen, `quote.project_id` setzen, Angebotspositionen in harte Reservierungen überführen. Fehlerfall sauber behandeln (kein halb angelegter Zustand). Abhängig von Kernfunktion 1.

**3.2 — Angebotsstatus `converted` einführen**
Statuswert ergänzen, DB-CHECK-Constraint anpassen (aktuell fehlt bereits `confirmed`, siehe E-2), Mehrfachumwandlung verhindern.

**3.3 — Statusmodell durchsetzen**
`CreateProject` auf Initialstatus `draft` festnageln. `UpdateProject` auf die Domain-Methoden umstellen, damit Datums- und Budgetvalidierung greifen.

**3.4 — Weiche Reservierung**
Reservierungsstatus `tentative` einführen, der bei Angebotsablauf automatisch verfällt. In die Verfügbarkeitsrechnung getrennt einrechnen. Abhängig von 1.3.

**3.5 — Projekt kopieren vervollständigen**
Packlisten und Reservierungen mitkopieren, Datumsversatz anwenden. Frontend-Aufruf um die Pflichtfelder `new_start_date`/`new_end_date` ergänzen.

**3.6 — Default-Währung auf EUR**
Trivial, aber wirkt sich auf jede neue Rechnung aus.

---

## 4. Packliste und Kommissionierung ⚠️

### Zweck

Aus dem Auftrag wird eine physisch abarbeitbare Liste.

### Fachregeln (SOLL)

1. **Automatisch aus den Reservierungen erzeugen.**
2. **Nach Lagerplatz sortiert**, damit man einmal durchs Lager läuft.
3. Teilmengen abbilden: gepackt / geladen / zurück, mit Differenzanzeige.
4. Statusübergänge: `draft → confirmed → packed → loaded → returned`.
5. Druckbar (PDF) und auf dem Scanner abarbeitbar.
6. Rückgabe mit Fehlmengen- und Schadenserfassung.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| CRUD + Statusmodell | ✅ Vollständig | `project-service/internal/domain/packlist.go` |
| Erzeugung aus Reservierung | ❌ Packliste und Reservierung sind unverbundene Datenmodelle | — |
| Sortierung nach Lagerplatz | 🚫 `StorageLocation` ist über keine API befüllbar → alles unter „Nicht zugeordnet" | `domain/packlist.go:38`, `application/commands.go` |
| Teilrückgabe | ⚠️ Setzt den Status nicht, wenn nur ein Teil zurückkommt | `domain/packlist.go:142` |
| HTML-Ausgabe | ⚠️ String-Konkatenation ohne Escaping → XSS über Kunden-/Projektnamen | `project_service.go:583-653` |
| Sortieralgorithmus | ⚠️ Handgeschriebener O(n²)-Bubble-Sort | `project_service.go:402-408` |
| Update-Verhalten | ⚠️ Löscht alle Positionen und schreibt sie neu, ohne Optimistic Locking | `packlist_postgres.go:84-105` |

### Ausbauaufträge

**4.1 — Packliste aus Reservierungen generieren**
Endpunkt `POST /api/v1/projects/{id}/packlist/generate`. Erzeugt Positionen aus allen harten Reservierungen des Projekts, inkl. Menge. Bei erneutem Aufruf abgleichen statt duplizieren. Abhängig von 1.1.

**4.2 — StorageLocation befüllbar machen** 🚫
Feld in `AddPacklistItemCommand` und DTO aufnehmen, Handler durchreichen. Beim Generieren automatisch aus `equipment.location_id` ziehen. Ohne diesen Schritt ist die Sortierfunktion wirkungslos.

**4.3 — Teilrückgabe korrekt abbilden**
Zwischenstatus `partially_returned` einführen; Status erst bei vollständiger Rückgabe auf `returned` setzen.

**4.4 — HTML-Escaping**
`html/template` statt String-Konkatenation verwenden. Sicherheitsrelevant.

**4.5 — Rückgabeerfassung**
Fehlmengen und Schäden beim Rücklauf erfassen, Schaden erzeugt automatisch einen Werkstattauftrag (Kernfunktion 6) und ggf. einen Versicherungsfall (Kernfunktion 6/9).

---

## 5. Scannen: Ausgabe und Rücknahme 🚫

### Zweck

Der physische Warenfluss — schnell, offlinefest, auf PDA-Hardware.

### Fachregeln (SOLL)

1. **Check-Out gegen die Packliste:** Scan bucht auf das Projekt, Status wechselt, Restliste schrumpft sichtbar.
2. **Check-In mit Zustandserfassung:** Bewertung, Notiz, Foto. „Defekt" erzeugt Werkstattauftrag und sperrt das Gerät.
3. **RFID-Bulk:** Palette in einem Rutsch, serverseitig als **ein** Vorgang.
4. **Offlinefähig:** Queue auf dem Gerät, Sync mit **Idempotenzschlüssel** und echter Konfliktauflösung.
5. Unterschrift bei Übergabe (Dry Hire), Übergabeprotokoll als PDF.
6. Inventur: Zone scannen, Soll/Ist abgleichen, Differenzen dokumentieren.

### Ist-Zustand

Dies ist der **kritischste Bruch im ganzen System**. Die Android-App (v2.3.4, ~8.800 Zeilen Kotlin, echte CF-H906-Hardwareanbindung) ist substanziell gebaut, spricht aber nicht mit dem Backend.

| Bruch | Wirkung | Fundstelle |
|---|---|---|
| **Check-Out/Check-In ändern nichts** | Die App ruft `endSession`, aber nie `POST /api/v1/scanner/checkout` bzw. `/checkin` — die einzigen Endpunkte, die den Equipment-Status setzen würden. **Der Materialbestand wird durch Scannen nicht verändert.** | `CheckOutViewModel.kt:290-307` |
| **Response-Envelope fehlt** | App erwartet `{data:…}`, scanner-service schreibt das DTO roh. Gson liefert `data = null` → **jeder** Aufruf gilt als Fehlschlag. | `scanner-service/…/handlers.go:468-472` vs. `ApiResponse.kt` |
| **Falscher Feldname** | App sendet `{"type":…}`, Handler liest `{"context":…}` → immer `CONTEXT_REQUIRED`. | `handlers.go:719` |
| **Falsche Kontextwerte** | App sendet `out/in/inventory`, Domain kennt `check_out/check_in/lager_einraeumen/inventur`. Unbekannte Werte fallen auf `check_out` zurück → **Inventurscans werden als Check-Out klassifiziert**. | `session_service.go:118-125` |
| **Sync-Body falsch** | `SyncWorker` postet ein nacktes Array, Handler erwartet `{"scans":[…]}`. Sync kann strukturell nie funktionieren. | `SyncWorker.kt:35` vs. `handlers.go:152-165` |
| **Offline-Queue wird nie gefüllt** | Nur `scan()` hat einen Offline-Fallback, aber alle Workflows nutzen `sessionScan()` — ohne Fallback. | `ScannerRepository.kt:111-129` |
| **Drei App-Endpunkte fehlen serverseitig** | Standortwechsel und die gesamte RFID-Zuordnung sind nicht angebunden. | `ScannerApi.kt:49,58,61` |
| **RFID-Bulk ineffizient** | Jeder Tag wird einzeln per HTTP aufgelöst — 200 Tags = 200 Requests. | `CheckOutViewModel.kt:169 ff.` |
| **`user_id` leer** | App sendet `""`, Backend-Validierung fordert einen Wert → `VALIDATION_ERROR`. | `ScannerRepository.kt:69,118` |
| **Geräte-ID konstant** | `Build.SERIAL` liefert seit Android 10 `"unknown"`. | `ScannerRepository.kt:73,120,326` |
| **Demo-Modus im Release** | Sentinel-String `"demo-access-token"` liefert erfundene Erfolge — im Release-Build aktiv. | `ScannerRepository.kt:40,47,166,170` |
| **Logging im Release** | `HttpLoggingInterceptor.Level.BODY` unbedingt aktiv — Tokens landen im Logcat. | `di/AppModule.kt:50` |
| **Inventur wirkungslos** | Soll-Bestand wird nie befüllt → jeder Scan gilt als „unerwartet". | `warehouse-service/…/inventory_check_service.go` |
| **Web-Scanner umgeht scanner-service** | Die Browser-Seite spricht direkt mit inventory-service, ohne Sessions und Protokoll. | `frontend/…/ScannerPage.tsx` |
| **Traefik-Routen fehlen** | `/api/v1/warehouse/*` ist nicht geroutet — die beiden Scanner-Verträge sind von außen unerreichbar. | `docker-compose.yml:335` |

### Ausbauaufträge

**5.1 — API-Vertrag festschreiben** 🚫
Einen verbindlichen Kontrakt zwischen App und scanner-service definieren (OpenAPI-Datei im Repo). Dann beide Seiten daran ausrichten: Envelope, Feldnamen, Kontextwerte. **Blockiert alle weiteren Scanner-Aufträge.**

**5.2 — Check-Out/Check-In tatsächlich buchen** 🚫
App muss die Status-ändernden Endpunkte aufrufen. Serverseitig muss der Aufruf transaktional sein: Scan-Event + Equipment-Status + Packlistenposition in einem Vorgang. Abhängig von 5.1.

**5.3 — Fehlende Endpunkte ergänzen**
`PUT /api/v1/scan/equipment/{id}/location`, `PUT .../rfid`, `GET .../untagged` im scanner-service implementieren (Weiterleitung an inventory-service). Traefik-Regeln für `/api/v1/warehouse/*` ergänzen.

**5.4 — Offline-Queue reparieren**
Offline-Fallback in `sessionScan()` einbauen. Sync-Body auf das erwartete Format bringen. **Idempotenzschlüssel** je Scan einführen (Client-generierte UUID), serverseitig deduplizieren. Item-genaues Sync-Ergebnis statt „alles oder nichts".

**5.5 — RFID-Bulk als Batch**
Sammelendpunkt `POST /api/v1/scanner/bulk-resolve` mit Tag-Liste. App puffert Tags mit kurzem Debounce und löst gebündelt auf.

**5.6 — Inventur funktionsfähig machen** 🚫
Soll-Bestand beim Anlegen einer Inventur aus dem inventory-service befüllen (`InventoryCheck.AddItem` wird aktuell nie aufgerufen). Differenzen persistieren statt nur zu loggen. Abhängig von 2.1.

**5.7 — Release-Härtung der App**
Demo-Modus und Body-Logging auf Debug-Builds beschränken. Geräte-ID auf `Settings.Secure.ANDROID_ID` umstellen. `user_id` aus dem Token setzen.

---

## 6. Wartung, Prüfung, Werkstatt ⚠️

### Zweck

Betriebssicherheit und Rechtssicherheit. DGUV V3 ist Arbeitgeberpflicht, keine Kür.

### Fachregeln (SOLL)

1. **Prüffristen nach DGUV V3 / VDE 0701-0702**, abhängig von Einsatzbedingung und Gerätetyp — ortsveränderliche Betriebsmittel als Richtwert 6 Monate (bei Fehlerquote ≤ 2 % verlängerbar), ortsfeste Anlagen bis 4 Jahre. Nicht pauschal.
2. **Messwerte gegen Grenzwerte prüfen:** Isolationswiderstand, Schutzleiterwiderstand, Ableitstrom. Das Ergebnis muss berechnet, nicht vom Client diktiert werden.
3. **„Nicht bestanden" hat Konsequenzen:** Gerät sperren, aus laufenden Reservierungen entfernen, Fehlmenge auslösen, Benachrichtigung.
4. **Fälligkeiten aktiv überwachen:** Scheduler erzeugt aus fälligen Plänen Aufgaben und warnt vor.
5. **Prüfprotokoll revisionssicher archivieren.**
6. Reparaturauftrag mit Kosten, Ersatzteilen, Ausfallzeit.
7. Wartungszyklus schließt sich: nach Abschluss wird der Plan fortgeschrieben.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| Datenmodell | ✅ Vollständig, inkl. Messwertfelder und IZYTRON-Import | `maintenance-service/internal/domain/entities.go` |
| Prüffrist | ❌ Pauschal `AddDate(1,0,0)` hart codiert | `electrical_test_service.go:31,175` |
| CSV-Folgetermin | ❌ Aus der Prüfsoftware gelesen und **explizit verworfen** | `handlers.go:779-780` |
| Grenzwertprüfung | ❌ Werte werden gespeichert, nie validiert | `electrical_test_service.go` |
| Konsequenz bei `failed` | ❌ Keine | — |
| Scheduler | ❌ Nicht vorhanden; Status `overdue` wird nie gesetzt | — |
| Wartungszyklus | 🚫 `planRepo` injiziert, nie benutzt → Plan wird nach Abschluss nicht fortgeschrieben | `maintenance_task_service.go:15,22,28` |
| Checklistenergebnisse | 🚫 Werden beim Abschluss durch ein **leeres Array** ersetzt | `maintenance_task_service.go:183-187` |
| Intervalltypen | ⚠️ `after-use` und `hours-based` bekommen nie ein Fälligkeitsdatum | `maintenance_plan_service.go:36-40` |
| Equipment-Sperre | ⚠️ Client zeigt auf nicht existierenden Servicenamen; Fehler werden verschluckt | `cmd/server/main.go:55` |
| Enum-Mismatch Frontend | 🚫 Frontend sendet `interval`/`after_use`, Backend erwartet `interval-based`/`after-use` | `WorkshopPage.tsx:82,142` |
| Reparatur melden | ❌ Modal ohne State-Binding und ohne API-Call — reine Attrappe | `WorkshopPage.tsx:560-600` |
| Wartungsdetailseite | ❌ Zu 100 % Mock | `MaintenanceDetail.tsx:35,72-74` |

### Ausbauaufträge

**6.1 — Prüffristen regelbasiert** 🚫
Fristentabelle pro Gerätekategorie und Einsatzbedingung in `tenant_config`. Aus der CSV/XML importierten Folgetermin übernehmen, wenn vorhanden. Pauschale Jahresfrist entfernen.

**6.2 — Grenzwertprüfung**
VDE-Grenzwerte je Schutzklasse hinterlegen, `Result` serverseitig aus den Messwerten berechnen, Client-Angabe nur als Vorschlag akzeptieren.

**6.3 — Konsequenzen bei Nichtbestehen**
Bei `failed`: Equipment sperren, betroffene Reservierungen markieren, Fehlmenge erzeugen, Benachrichtigung auslösen. Abhängig von Kernfunktion 9 und 13.

**6.4 — Wartungszyklus schließen** 🔌
`planRepo` im `MaintenanceTaskService` benutzen: nach Abschluss `LastExecutedAt` und `NextDueAt` fortschreiben, `MaintenanceRecord` erzeugen.

**6.5 — Checklistenergebnisse persistieren** 🚫
Die drei Zeilen korrigieren, die das Ergebnis durch ein leeres Array ersetzen. Ohne das ist jedes Prüfprotokoll wertlos.

**6.6 — Scheduler**
Periodischer Job erzeugt aus fälligen Plänen Aufgaben, setzt `overdue`, versendet Vorwarnungen. Kann als Cron im Service laufen oder über Kernfunktion 13 (Workflows).

**6.7 — Enum-Mismatch beheben**
Frontend und Backend auf dieselben Werte bringen. Ein Testfall, der die Enums beider Seiten vergleicht, verhindert Rückfälle.

**6.8 — Reparaturworkflow bauen**
Defekterfassung → Werkstattauftrag → Kosten/Ersatzteile → Wiederfreigabe. Die Attrappe im Frontend durch echte Funktion ersetzen.

---

## 7. Personal und Zeiterfassung ⚠️

### Fachregeln (SOLL)

1. Verfügbarkeit und Doppelbelegungsschutz.
2. **Qualifikationen bei der Zuweisung prüfen** — ohne gültigen Führerschein keine Fahrerrolle.
3. **Ablaufüberwachung** für Zertifikate mit Vorwarnung.
4. **Zeiterfassung nach ArbZG:** Pflichtpausen (30 min ab 6 h, 45 min ab 9 h), 10-h-Höchstgrenze, 11 h Ruhezeit, Zuschläge Nacht/Sonn-/Feiertag.
5. Manuelle Nacherfassung und Korrektur mit Genehmigungsschritt.
6. Abwesenheiten (Urlaub, Krankheit) mit Auswirkung auf Verfügbarkeit.
7. Freelancer-Anfrage per Link mit Zu-/Absage.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| Verfügbarkeit/Doppelbelegung | ✅ Wird hart verhindert, kein Force-Override | `assignment_service.go:94,162` |
| Overlap-Grenzen | ⚠️ Inklusiv an beiden Rändern → 18:00-Ende kollidiert mit 18:00-Beginn | `domain/entities.go:140-160` |
| Qualifikationsprüfung | 🔌 `HasQualification`/`CanDrive` gebaut, bei Zuweisung nie aufgerufen | `entities.go:163-187` |
| Ablaufüberwachung | ❌ Status wird nur bei Anlage gesetzt → abgelaufene Zertifikate bleiben ewig „gültig" | `qualification_service.go:82-84` |
| Manuelle Nacherfassung | 🚫 Backend **ignoriert** übergebene Start-/Endzeiten, schreibt 2× „jetzt" → immer 0 Stunden | `time_record_handlers.go:112-165` |
| Frontend-Erfassung | 🚫 Sendet kein `crew_member_id` → HTTP 500 | `TimeTrackingPage.tsx:417-422` |
| Genehmigung | 🔌 Handler existiert, **keine Route** | `time_record_handlers.go:167` |
| ArbZG-Regeln | ❌ Nicht vorhanden; Überstunden werden vom Client diktiert | `time_record_service.go:81-109` |
| Parallele Stempeluhren | ⚠️ Keine Prüfung auf laufenden Datensatz | `time_record_service.go:34-78` |
| Abwesenheiten | ❌ Frontend-Tab vorhanden, Backend-Route und Modell fehlen komplett | `TimeTrackingPage.tsx` |
| Freelancer-Booking | ✅ Sicheres 32-Byte-Token, funktionierender Ablauf | `booking_service.go` |
| Booking-Link | ⚠️ Hart `http://localhost:3000/booking/` | `booking_service.go:129` |
| Token-Ablauf | ❌ Kein `expires_at` | `migrations/002_create_booking_requests.sql` |
| Projektname im Booking | ⚠️ `"Projekt " + ID` — project-service wird nie gefragt | `booking_service.go:126,175` |
| CalDAV | ❌ Existiert nicht. Nur ICS-**Download**, kein abonnierbarer Feed, im Frontend nirgends aufgerufen | `handlers.go:436-471` |
| Projektfilter Assignments | 🚫 `project_id`-Parameter wird ignoriert → Projektansicht zeigt alle Zuweisungen des Mandanten | `handlers.go:252-280` |

### Ausbauaufträge

**7.1 — Zeiterfassung reparieren** 🚫
Backend muss `start_time`/`end_time` aus der Payload übernehmen. Frontend muss `crew_member_id` mitsenden. Route für die Genehmigung ergänzen. Ohne diese drei Schritte ist die Zeiterfassung unbenutzbar.

**7.2 — ArbZG-Regelwerk**
Pflichtpausen automatisch abziehen, Höchstarbeitszeit und Ruhezeit prüfen und bei Verstoß warnen. Zuschlagssätze konfigurierbar. Überstunden berechnen statt entgegennehmen.

**7.3 — Qualifikationsprüfung aktivieren** 🔌
`CreateAssignment` prüft die für die Rolle nötige Qualifikation. Rollen↔Qualifikations-Mapping konfigurierbar.

**7.4 — Zertifikatsablauf überwachen**
Täglicher Job setzt abgelaufene Qualifikationen auf `expired` und warnt 30/14/7 Tage vorher.

**7.5 — Abwesenheitsverwaltung**
Modell, Migration, Endpunkte und Einrechnung in die Verfügbarkeit. Der Frontend-Tab existiert bereits.

**7.6 — Overlap-Grenzen korrigieren**
Halboffene Intervalle wie in Kernfunktion 1.

**7.7 — Booking härten**
`expires_at` ergänzen, Basis-URL konfigurierbar machen, Projektdaten beim project-service abrufen, Rate-Limiting auf die öffentlichen Endpunkte.

**7.8 — Projektfilter implementieren**
`project_id` in `ListAssignments` auswerten; `ConflictDTO` um `project_id` erweitern, damit das Frontend filtern kann.

**7.9 — Kalender-Abo statt Download**
ICS als tokengeschützter, abonnierbarer Feed mit `REFRESH-INTERVAL` und korrektem Content-Type, ohne `attachment`-Disposition. Im Frontend anbieten.

---

## 8. Transport ⚠️

### Fachregeln (SOLL)

1. Fahrzeuge mit Nutzlast/Volumen, Beladungsprüfung.
2. **Touren mit mehreren Stops, Adressen, Zeitfenstern**, Reihenfolge optimierbar.
3. **Lenk- und Ruhezeiten** nach VO (EG) 561/2006 bei Fahrzeugen über 3,5 t.
4. Lieferschein automatisch erzeugen.
5. Fahrzeugstatus bei Tourstart/-ende mitführen.
6. DGUV-Fristen für Fahrzeuge überwachen, überfällige Fahrzeuge sperren.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| Kapazitätsprüfung | ✅ Das einzige echte Regelwerk im Service | `application/services.go:248-293` |
| Gewichtsdaten | ⚠️ Kommen vom Client, nicht aus dem Artikelstamm | `commands.go` |
| Mehrere Stops/Routen | ❌ Eine Tour hat genau ein Projekt, kein Stop-Modell, keine Adressen | — |
| Statusübergänge | ⚠️ Beliebiger String akzeptiert; `loading`/`delivered` nie erreichbar | `services.go:181-199` |
| Fahrzeugstatus | ❌ Wird bei Tourstart/-ende nicht mitgeführt | — |
| Lenk-/Ruhezeiten | ❌ Felder existieren, werden nicht befüllt, keine Regeln | `services.go:348-377` |
| Lieferschein | 🔌 Implementiert, **keine Route**, nur Noop-Client verdrahtet | `services.go:400-445`, `main.go:47` |
| Fahrervalidierung | 🔌 `CrewClient` komplett gebaut, **nie instanziiert** | `internal/ports/crew_client.go` |
| DGUV Fahrzeuge | ❌ Reine Datenfelder ohne Logik | — |
| Traefik-Routing | 🚫 Regel lautet `/api/v1/vehicles`, tatsächliche Pfade sind `/api/v1/transport/vehicles` → **Service über Gateway unerreichbar** | `docker-compose.yml` |
| Frontend-Routen | ⚠️ `updateTour` und `deleteVehicle` rufen nicht existierende Routen | `services/api.ts` |
| Detailseite | ⚠️ Fällt auf `mockTourDetail` zurück | `TransportDetail.tsx:39-40` |

### Ausbauaufträge

**8.1 — Traefik-Routing korrigieren** 🚫
Ohne diesen Fix ist der gesamte Service in Produktion nicht erreichbar. Einzeiler.

**8.2 — Stop-Modell einführen**
`TourStop` mit Adresse, Zeitfenster, Reihenfolge, Typ (Abholung/Lieferung/Rückholung). Grundlage für Routenoptimierung.

**8.3 — Lieferschein freischalten** 🔌
Route ergänzen, echten `DocumentHTTPClient` statt Noop verdrahten, Längenprüfung vor `tourID[:8]` (Panic-Risiko).

**8.4 — Fahrervalidierung aktivieren** 🔌
`CrewClient` instanziieren, `TourService` erweitern, Fahrerqualifikation prüfen. Abhängig von 7.3.

**8.5 — Statusübergänge validieren**
Übergangsmatrix wie beim Projekt. Fahrzeugstatus mitführen.

**8.6 — Lenk- und Ruhezeiten**
Erfassung vervollständigen (`ActivityType`, Pausen, Orte), Regelprüfung nach VO 561/2006, Warnung bei Verstoß.

**8.7 — Gewichtsdaten aus dem Stamm ziehen**
Statt Client-Angaben das Gewicht vom inventory-service abrufen.

---

## 9. Fehlmengen und Zumietung ❌

### Zweck

Der Moment, in dem Geld verdient oder verloren wird — und das eigentliche Differenzierungsmerkmal des Produkts.

### Fachregeln (SOLL)

1. **Fehlmengen aus der Disposition ableiten:** je Zeitraum, Projekt und Artikel die Unterdeckung, mit Vorwarnzeit.
2. **Auflösungswege anbieten:** umdisponieren, Ersatzartikel, zumieten, Kunde informieren.
3. **Sub-Rental:** Anfrage an Partner → Angebot → Bestätigung. Zumietware wird als temporärer Bestand mit Kosten geführt und fließt in Verfügbarkeit **und** Marge ein.
4. **Federation:** Partner per gegenseitig authentifiziertem TLS verbinden, Katalog synchronisieren, Datenfreigabe nach Kategorien steuern.
5. Zumietung erzeugt Eingangsrechnung und Übergabedokument.

### Ist-Zustand

**Der größte Blindgänger des Projekts.** Es gibt kein Fehlmengen-Backend. Die Seite rechnet im Browser:

| Aspekt | Status | Fundstelle |
|---|---|---|
| Datenbasis | ❌ `equipmentApi.list({limit:100})` — hartes Limit 100 | `ShortagesPage.tsx:36-46` |
| Berechnung | ❌ Keine Zeiträume, keine Reservierungen; `required = inUse + 1` frei erfunden | `ShortagesPage.tsx:53-92` |
| **Zeigt immer null** | 🚫 Filtert auf Projektstatus `active`/`planning` — **existieren im Enum nicht** → Liste ist permanent leer | `ShortagesPage.tsx:66` |
| Betroffene Projekte | ❌ Einfach die ersten zwei Projektnamen | `ShortagesPage.tsx:82` |
| Partnerliste | ❌ Hartkodiertes 5-Elemente-Array | `ShortagesPage.tsx:19-25` |
| Anfrage senden | ❌ Sendet nichts, zeigt nur einen Toast | `ShortagesPage.tsx:116-121` |
| federation-service | 🔌 Existiert mit vollständigen Endpunkten, wird **nie aufgerufen** | `federation-service/…/router.go` |
| Frontend-Routen | ⚠️ `federationApi.subRentals()` zeigt auf nicht existierende Route | `services/api.ts:1449` |

**Federation selbst:**

| Aspekt | Status | Fundstelle |
|---|---|---|
| mTLS | ❌ **Existiert nicht.** Standard-`http.Client` ohne `tls.Config` und ohne Client-Zertifikate | `sharing_service.go:35,163-164` |
| Zertifikate | ⚠️ Selbstsigniert, alle mit `SerialNumber = 1`, **ohne `ExtKeyUsage`** → als mTLS-Client-Zertifikat unbrauchbar | `certificate_service.go:45` |
| Privater Schlüssel | 🚫 Wird über die API unter dem Feldnamen `ServerCertPem` ausgeliefert und **nirgends gespeichert** → nach einem Aufruf verloren | `certificate_service.go:69-77` |
| Kettenprüfung | ❌ Nur Fingerprint-Vergleich, keine Signaturprüfung | `certificate_service.go:141` |
| Rechnung bei Abschluss | ❌ `invoiceID := uuid.New()` — **frei erfundene ID** | `sharing_service.go:145` |
| Partnerstatus | ❌ `Status`/`TrustLevel` werden nirgends ausgewertet — gesperrte Partner nutzbar | gesamt |
| Gegenseitigkeit | 🚫 Konsumiert `/api/v1/federation/equipment`, stellt diesen Endpunkt selbst **nicht** bereit | `sharing_service.go:163` vs. `router.go` |

### Ausbauaufträge

**9.1 — Fehlmengen-Backend** 🚫
Neuer Endpunkt `GET /api/v1/shortages?from=&to=` im project-service oder inventory-service. Berechnet je Artikel und Zeitraum die Unterdeckung aus Reservierungen gegen Bestand. Frontend komplett auf diese Quelle umstellen. **Abhängig von 1.3.**

**9.2 — Auflösungswege**
Pro Fehlmenge Aktionen anbieten: Ersatzartikel vorschlagen (gleiche Kategorie, freier Bestand), umdisponieren, zumieten.

**9.3 — Frontend an federation-service anschließen**
Hartkodierte Partnerliste durch `GET /api/v1/federation/partners` ersetzen, Anfrage über `POST /api/v1/federation/requests` senden, Routen im API-Layer korrigieren.

**9.4 — mTLS tatsächlich implementieren** 🚫
Zertifikate mit korrektem `ExtKeyUsage` (ClientAuth/ServerAuth), eindeutigen Seriennummern und SANs erzeugen. Privaten Schlüssel verschlüsselt in `key_pem_encrypted` speichern (die Spalte existiert). `http.Client` mit `tls.Config{Certificates, RootCAs}` konfigurieren. Kettenprüfung in `ValidatePartnerCert`.

**9.5 — Gegenseitigkeit herstellen**
Den Endpunkt bereitstellen, den die eigene Synchronisierung erwartet — sonst kann keine zwei Instanzen miteinander sprechen.

**9.6 — Partnerstatus durchsetzen**
`pending`/`suspended`/`revoked` blockieren jede Anfrage und Synchronisierung.

**9.7 — Echte Rechnung bei Abschluss**
Statt erfundener UUID einen Aufruf an den invoice-service. Übergabedokument über den document-service erzeugen.

**9.8 — Zumietware als temporärer Bestand**
Bestätigte Sub-Rentals fließen in die Verfügbarkeitsrechnung ein und tragen ihre Kosten in die Projektmarge. Abhängig von 1.3.

---

## 10. Rechnung, Steuer, Buchhaltung 🚫

### Zweck

Rechtssichere Fakturierung und Übergabe an die Buchhaltung.

### Fachregeln (SOLL)

1. Rechnung aus Projekt inkl. Mietdauer, Rabatten, Zusatzkosten, Personal, Transport, Schäden.
2. **§14 UStG-Pflichtangaben:** leistender Unternehmer mit Anschrift, Steuernummer/USt-IdNr, Leistungszeitpunkt bzw. -zeitraum, fortlaufende Nummer, Entgelt nach Steuersätzen aufgeschlüsselt.
3. **E-Rechnung (XRechnung/ZUGFeRD)** — für B2B-Inlandsumsätze in Deutschland verpflichtend, Übergangsfristen laufen bis 2028 aus.
4. **Steuerfälle:** Kleinunternehmer §19, Reverse Charge §13b, innergemeinschaftliche Lieferung, ermäßigter Satz, Steuerbefreiungen.
5. **GoBD:** Unveränderbarkeit finalisierter Belege, lückenlose Nummernvergabe, Hash-Chain über **alle** relevanten Felder inkl. Positionen, revisionssicherer Audit-Trail, 10 Jahre Aufbewahrung, GDPdU-Datenträgerüberlassung.
6. **DATEV-Export im echten Format:** EXTF/DTVF-Kopfzeile mit Formatkennung, Version, Berater-/Mandantennummer und Wirtschaftsjahr; semikolongetrennt; Dezimalkomma; Windows-1252; BU-Schlüssel; Debitorenkonten je Kunde.
7. Mahnwesen mit Stufen, Fristen, Verzugszinsen §288 BGB, Mahn-PDF und Versand.
8. Zahlungsabgleich mit Kontoanbindung (FinTS/CAMT).

### Ist-Zustand

Das Produkt wirbt mit GoBD-Konformität. **Nichts davon ist wirksam implementiert.**

| Aspekt | Status | Fundstelle |
|---|---|---|
| **Rechnungslebenszyklus** | 🚫 DB-Constraint `immutable_finalized CHECK (status = 'draft')` — **keine Rechnung kann je `sent`, `paid`, `cancelled` oder `credited` werden** | `migrations/001_create_invoices.sql:41` |
| Auftragsbestätigung | 🚫 CHECK erlaubt `confirmed` nicht, Code und Route setzen es | `migrations/002_create_quotes.sql:25` |
| Gutschriften 19 % | 🚫 `tax_rate DECIMAL(5,4)` (max 9,9999), Code schreibt `19` → numeric overflow | `migrations/006_create_credit_notes.sql:12` |
| Kleinunternehmer-Flag | 🚫 Wird **nie persistiert**, geht aber in den Hash ein → jede §19-Rechnung scheitert nach Reload an der eigenen Integritätsprüfung | `invoice_postgres.go:22-66` |
| Teilzahlungen | 🚫 `paid_amount`/`remaining_amount` werden nicht geschrieben → nach Reload 0 | `invoice_postgres.go:240-259` |
| Hash-Chain | ❌ Kein `previous_hash`; Einzeldokument-Hash **ohne Positionen** → Positionen austauschbar | `domain/invoice.go:379-399` |
| Unveränderbarkeit | ❌ `Update` löscht alle Positionen und schreibt sie neu — für jede Rechnung | `invoice_postgres.go:266` |
| Nummernvergabe | ⚠️ `FOR UPDATE` ohne Transaktion → Lücken/Doppelnummern möglich; keine Jahresrücksetzung | `sequence_postgres.go:20-62` |
| Nummernformat | ⚠️ `EF-%d-%04d` hart codiert; Einstellungsseite hat keine Wirkung | `invoice_service.go:58-65` |
| §14-Pflichtangaben | ❌ Im PDF nicht vorhanden — Firmendaten sind nicht einmal Felder | `invoice_service.go:644-833` |
| PDF-Generator | 🚫 Verwirft das HTML, strippt Tags und rendert Rohtext | `chromedp_generator.go:58-79` |
| Schöner PDF-Generator | 🔌 Vollständig gebaut, **nie aufgerufen** | `chromedp_generator.go:83-266` |
| PDF-Templates | 🔌 Sechs Templates vorhanden, nur eines genutzt — über einen relativen Pfad, der im Container nicht existiert | `invoice_service.go:876` |
| E-Rechnung | ❌ Repo-weit kein Treffer | — |
| Reverse Charge | ❌ Existiert nicht, wird aber in `IMPLEMENTATION.md:46` behauptet | — |
| DATEV | ❌ Komma-CSV ohne Kopfzeile, **drei Buchungszeilen pro Rechnung** statt einer, Datum als MMTT statt TTMM, `paid`-Rechnungen übersprungen, Limit 1000 | `export_service.go:31-113` |
| Versandstatus | ⚠️ Ohne SMTP wird trotzdem `sent` gesetzt — Rechnung gilt als versandt ohne Versand | `invoice_service.go:546-606` |
| Mahnwesen | ⚠️ Gebühren hart codiert; Auto-Mahnlauf ohne Route; kein PDF, kein Versand, keine Verzugszinsen | `dunning.go:33-37`, `dunning_service.go:152-215` |
| Banking-Import | 🚫 `_ = tenantID` — importierte Transaktionen werden **nicht gespeichert** | `handlers.go:1030` |
| Zahlungsabgleich | ⚠️ Ignoriert `partially_paid` → teilbezahlte Rechnungen unsichtbar | `invoice_postgres.go:357` |
| Audit-Trail | 🚫 **Kein einziger Service schreibt Audit-Einträge** | repo-weit 0 Treffer |
| Audit-Verifikation | 🚫 Zeitstempel wird zweimal erzeugt → Prüfsumme weicht **immer** ab, Kette nie gültig | `audit_service.go:53,76` |
| Audit-Hash | ❌ Deckt `old_values`/`new_values`/`user_id` **nicht** ab — genau die Änderungsdaten sind ungeschützt | `audit_service.go:28-88` |
| Audit-Export | 🚫 ZIP wird im Speicher gebaut und verworfen; `file_path` bleibt leer, kein Download | `export_service.go:32-106` |
| GoBD-Prüfung im UI | ❌ `setTimeout(2500)` und dann fest „bestanden" — der echte Endpunkt wird nicht aufgerufen | `DocumentsPage.tsx:113-127` |
| Dokument-Hash-Chain | 🚫 `previous_checksum` wird bei `CreateDocument` nie gesetzt → Kette bricht ab Dokument 2 | `document_service.go:57-79` |
| Dokument-PDF | ⚠️ Handgeschriebener PDF-String mit konstanten `xref`-Offsets → vermutlich defekte Dateien | `pdf_generator.go:20-68` |
| Öffentliche Signatur | 🚫 Prüft **keinen Token** — wer eine Signatur-ID kennt, kann fremd unterschreiben; `Verified` wird bedingungslos `true` | `signature_service.go:185-225` |

### Ausbauaufträge

**10.1 — Rechnungslebenszyklus entsperren** 🚫
Migration, die `immutable_finalized` entfernt und durch eine korrekte Regel ersetzt: Änderungen nur im Status `draft` erlauben, per Trigger oder Rechteentzug. Gleichzeitig `confirmed` in den Quote-CHECK aufnehmen und `credit_notes.tax_rate` auf `DECIMAL(5,2)` erweitern. **Ohne diesen Schritt kann das System keine einzige Rechnung versenden.**

**10.2 — Repository vervollständigt schreiben** 🚫
`is_kleinunternehmer`, `kleinunternehmer_text`, `paid_amount`, `remaining_amount`, `invoice_items.tax_amount` in `Create`, `Update` und die Select-Spalten aufnehmen. Behebt zugleich den Hash-Fehlschlag bei §19-Rechnungen.

**10.3 — GoBD-Hash-Chain korrekt**
`previous_hash` in `invoices` ergänzen, Hash über alle wertrelevanten Felder **inklusive Positionen** bilden, Kette beim Finalisieren fortschreiben, Verifikationsendpunkt bereitstellen. Analog im document-service die Kette beim Anlegen befüllen.

**10.4 — Unveränderbarkeit durchsetzen**
DB-Trigger, der UPDATE/DELETE auf finalisierten Rechnungen und deren Positionen verhindert. Storno nur über Gutschrift.

**10.5 — Nummernvergabe absichern**
Atomares UPSERT in einer Transaktion (die korrekte Implementierung existiert bereits ungenutzt in `invoice_postgres.go:334-351`). Jahresrücksetzung. Präfix aus `tenant_config` statt Konstante.

**10.6 — Rechnungs-PDF neu bauen** 🔌
Den vorhandenen, ungenutzten Generator verdrahten oder auf HTML→PDF umstellen. Templates über einen absoluten, containertauglichen Pfad laden. Firmenstammdaten als Felder einführen und alle §14-Pflichtangaben ausgeben. Deutsche Beschriftungen. HTML-Escaping.

**10.7 — E-Rechnung**
XRechnung (UBL/CII) und ZUGFeRD/Factur-X als Ausgabeformat. Validierung gegen das offizielle Schema. Höchste regulatorische Priorität.

**10.8 — Steuerfälle**
Reverse Charge §13b, innergemeinschaftliche Lieferung, Steuerbefreiungen. Steuersätze aus `vat_classes` lesen statt aus dem Go-Typ. Rundung auf 2 Nachkommastellen mit `decimal` statt `float64`.

**10.9 — DATEV-Export neu schreiben**
Echtes EXTF-Format mit Kopfzeile, eine Buchungszeile pro Rechnung mit BU-Schlüssel, Debitorenkonten je Kunde, Dezimalkomma, Windows-1252, Datum als TTMM. Bezahlte Rechnungen einschließen. Paginierung statt Limit 1000. SKR04 zusätzlich zu SKR03.

**10.10 — Audit-Trail in Betrieb nehmen** 🚫
Zeitstempel-Bug beheben (einmal erzeugen, wiederverwenden). Nutzdaten in den Hash aufnehmen. Append-only per DB-Rechten durchsetzen. **Middleware in allen Services**, die schreibende Operationen protokolliert. Ohne Schreiber bleibt der Trail leer, egal wie korrekt er ist.

**10.11 — Audit-Export mit Download**
ZIP persistieren, `file_path` setzen, Download-Endpunkt, Datumsfilter auswerten, GDPdU-Struktur (INDEX.XML + DTD).

**10.12 — Mahnwesen vervollständigen**
Gebühren und Fristen aus `tenant_config`. Mahn-PDF aus den vorhandenen Templates. Versand. Verzugszinsen §288 BGB. Auto-Mahnlauf als Job.

**10.13 — Banking**
Transaktionen persistieren (Tabelle fehlt), Duplikaterkennung, `partially_paid` in den Abgleich aufnehmen. Perspektivisch FinTS/CAMT statt CSV.

**10.14 — Signatur-Endpunkt absichern** 🚫
Token prüfen, Ablauffrist einführen, `Verified` nicht bedingungslos setzen. Sicherheitsrelevant.

---

## 11. Ausgaben und Belege ⚠️

### Fachregeln (SOLL)

1. Beleg fotografieren → OCR erkennt Lieferant, Betrag, Steuer, Datum → Kontovorschlag SKR03/04 → Freigabe → Export.
2. **Vorsteuer korrekt:** aus Brutto `Betrag × Satz/(1+Satz)`.
3. **Bewirtungsbelege §4 Abs. 5 EStG:** 70/30-Split, Pflichtangaben (Teilnehmer, Anlass, Ort) validieren. Trinkgeld gehört zu den Aufwendungen.
4. Beleg revisionssicher archivieren, Prüfsumme bilden, kein Hartlöschen.
5. Budgetüberwachung je Kategorie/Projekt.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| KI-Belegerkennung | ✅ Handwerklich der sauberste Teil des Systems — deutscher Buchhaltungs-Prompt mit SKR03-Mapping, umfassendes Ergebnisschema | `ai-service/…/receipt_vision_service.go` |
| …aber lauffähig? | 🚫 Nein — kein AI-Provider wird je registriert (siehe Kernfunktion 13) | `ai-service/cmd/server/main.go:54` |
| **Vorsteuerberechnung** | 🚫 `TaxAmount = Amount * TaxRate` statt `Amount * r/(1+r)` → **jeder Vorsteuerbetrag ist falsch** | `domain/expense.go:124-127` |
| Einheit Prozent/Dezimal | 🚫 Frontend sendet `19`, Domain erwartet `0.19` — nirgends umgerechnet | `api.ts` vs. `expense.go:93` |
| Belegupload | 🚫 Frontend ruft `POST /api/v1/expenses/{id}/receipt` — **Route existiert nicht** | `ExpensesPage.tsx:200` |
| Prüfsumme | ❌ Feld existiert, wird nie berechnet | — |
| Freigabe-Workflow | 🚫 `Submit` (draft→pending) hat **keine Route** → nichts kann je genehmigt werden | `router.go` |
| Löschen | ❌ Hartlöschen, GoBD-widrig | `expense_service.go:241-251` |
| ID-Erzeugung | 🚫 Hash aus `tenant+vendor+date` → zwei Belege desselben Lieferanten am selben Tag kollidieren | `expense_service.go:30` |
| Budget | 🔌 `AddExpense` wird nie aufgerufen → `Spent` bleibt 0, `IsOverBudget()` immer false | `domain/budget.go` |
| Bewirtung | ⚠️ Trinkgeld wird komplett herausgerechnet statt zu 70 % abgezogen; Pflichtangaben nicht validiert | `expense.go:107-114` |
| Kategorien | ⚠️ Frontend hat 8 hartkodierte Kategorien ohne Bezug zur Backend-Tabelle mit SKR-Codes | `ExpensesPage.tsx:27-35` |
| Lieferanten-Lernen | 🔌 Tabelle `vendor_mappings` existiert, kein Go-Code greift darauf zu | `migrations/003` |
| Migrationen | 🚫 FKs auf `public.tenants` — **existiert nicht** → Tabellen werden vermutlich nie angelegt | `migrations/001:24, 002:11, 003:56` |

### Ausbauaufträge

**11.1 — Vorsteuerberechnung korrigieren** 🚫
Formel korrigieren, Einheit (Prozent vs. Dezimal) zwischen Frontend, DTO und Domain vereinheitlichen und mit Tests absichern. Betrifft jeden bereits erfassten Beleg — Datenmigration prüfen.

**11.2 — Migrationen reparieren** 🚫
FKs auf die nicht existierende Tabelle entfernen (siehe auch Teil C). Ohne das existieren die Tabellen nicht.

**11.3 — Belegupload und Archivierung**
Upload-Endpunkt, Speicherung über den `StorageAdapter`, SHA-256-Prüfsumme, Verknüpfung mit dem document-service.

**11.4 — Freigabe-Workflow**
`Submit`-Route ergänzen, damit die Kette draft→pending→approved durchlaufbar wird. Hartlöschen durch Storno ersetzen.

**11.5 — ID-Erzeugung auf UUID**
Betrifft alle Services mit dem djb2-Hash-Muster (siehe Teil C).

**11.6 — Budgetverbrauch aktivieren** 🔌
`AddExpense` bei Genehmigung aufrufen, Überschreitungswarnung.

**11.7 — Bewirtung korrigieren**
Trinkgeld in die 70/30-Rechnung einbeziehen, Pflichtangaben validieren.

**11.8 — Lieferanten-Lernen**
`vendor_mappings` nutzen: erkannte Zuordnungen speichern und beim nächsten Beleg desselben Lieferanten vorschlagen.

---

## 12. Auswertung ❌

### Fachregeln (SOLL)

**Auslastung pro Artikel** ist die Kennzahl, die über Investitionen entscheidet. Dazu: ROI je Gerät, Marge je Projekt, Umsatz je Kunde, Schadensquote, Zumietkosten, offene Posten.

Alle nur sinnvoll, wenn Kernfunktion 1 steht.

### Ist-Zustand

| Aspekt | Status | Fundstelle |
|---|---|---|
| KPI-Werte | 🚫 `Value` wird hart auf `nil` gesetzt — **keine Datenquelle angebunden** | `kpi_service.go:46` |
| Snapshot-Job | 🔌 `SnapshotKPIs` hat 0 Aufrufer, weder Route noch Scheduler | `kpi_service.go:27` |
| Berichtsgenerierung | 🚫 `GenerateReport` verwirft die Definition (`_ = def`), kein Run wird je `completed` | `report_service.go:86-121` |
| CSV/PDF-Export | 🚫 Statusprüfung macht beide Routen dauerhaft unerreichbar | `report_service.go:151-160,202-212` |
| Scheduler | ❌ Läuft mit `uuid.Nil` als Mandant; `ScheduleCron` wird nie geparst | `cmd/server/main.go:117` |
| Rollen-Dashboard | ⚠️ Rollen `gf`/`lager`/`buchhaltung` — entsprechen **nicht** dem RBAC-Modell des auth-service | `kpi_service.go:167-198` |
| Frontend | ⚠️ Umgeht den Service komplett, aggregiert clientseitig | `ReportsPage.tsx:86-104` |

### Ausbauaufträge

**12.1 — KPI-Datenquellen anbinden** 🚫
Pro KPI eine echte Abfrage: Umsatz aus invoice-service, Bestand aus inventory-service, aktive Projekte aus project-service, Auslastung aus der Verfügbarkeitsengine. Abhängig von 1.3.

**12.2 — Snapshot-Job starten** 🔌
Täglicher Ticker über alle Mandanten iterierend.

**12.3 — Berichtsgenerierung implementieren**
Worker, der `queued` → `running` → `completed` durchläuft, Daten sammelt und die Datei ablegt. Erst danach werden die Export-Routen erreichbar.

**12.4 — Rollen vereinheitlichen**
Auf das RBAC-Modell des auth-service umstellen.

**12.5 — Frontend anbinden**
Clientseitige Aggregation durch Serviceaufrufe ersetzen — sonst skaliert nichts über die ersten 100 Datensätze hinaus.

---

## 13. Plattform ⚠️

### 13.1 Authentifizierung und Rollen

**SOLL:** Anmeldung, Mandantentrennung, wirksames Rollenmodell, beendbare Sessions.

| Aspekt | Status | Fundstelle |
|---|---|---|
| Passwort-Hashing | ✅ **Argon2id** (64 MB, t=1, p=4), 12-Zeichen-Policy | `password_manager.go:27-36` |
| Legacy-Hashes | ⚠️ SHA-256-Fallback bleibt unbegrenzt gültig, keine Rehash-Migration | `password_manager.go:79-87` |
| Brute-Force-Schutz | ✅ Zwei Ebenen (IP eskalierend, Benutzer mit Auto-Unlock), fail-closed | `session_manager.go:189-258` |
| **RSA-Schlüsselpaar** | 🚫 Wird bei **jedem Start neu erzeugt** und nirgends gespeichert → jeder Neustart wirft alle Nutzer raus; horizontale Skalierung unmöglich | `cmd/server/main.go:53` |
| **RBAC** | 🚫 25 Permissions für 8 Rollen definiert — **0 Aufrufer**. Tatsächlich geprüft wird an **2** Endpunkten | `application/permissions.go:151-185` |
| Ungeschützte Endpunkte | 🚫 `AssignRole`, `DeleteUser`, `CreateTenant`, `InviteUser` und **alle Backup-Routen inkl. Restore** ohne Rollencheck | `router.go:81-111` |
| Logout | ❌ Löscht nur das Cookie; Refresh-Token bleibt 7 Tage einlösbar | `handlers.go:237-255` |
| Session-Manager | 🔌 Vollständig gebaut, **nur in Tests aufgerufen**; Tabelle `auth.sessions` tot | `session_manager.go` |
| QR-Login | 🚫 Tabelle `auth.qr_login_tokens` hat **keine Migration** → Laufzeitfehler | repo-weit 0 SQL-Treffer |
| Login-Lookup | ⚠️ Ohne Tenant-Filter → tenantübergreifend | `user_postgres.go:94-104` |
| Passwort-Reset-Link | ⚠️ Hart `http://localhost:3000/...` | `user_service.go:461` |
| Setup-Token-Vergleich | ⚠️ Nicht konstantzeitig; **Panic** bei Token < 8 Zeichen | `setup_service.go:179-180` |
| Token im Browser | ⚠️ `localStorage` (XSS-exponiert), kein Refresh-Flow → stündlicher Logout | `stores/authStore.ts` |

### 13.2 Mandantentrennung

🚫 **Nicht durchgesetzt.** Nur der auth-service leitet den Mandanten aus verifizierten JWT-Claims ab. Die Services **ai, workflow, federation, reporting** haben **gar keine Auth-Middleware** und lesen `X-Tenant-ID` roh aus dem Header. Der insurance-service liest den Mandanten aus einem **Query-Parameter** und prüft ihn bei Detailabfragen gar nicht. inventory-, warehouse-, invoice-, expense-, document-, audit-, crew-, transport-, maintenance-service vertrauen ebenfalls dem ungeprüften Header.

Die vorhandene Schutzfunktion `middleware.TenantMiddleware()` wird von **keinem** Service eingebunden.

### 13.3 Branchenprofile

| Aspekt | Status | Fundstelle |
|---|---|---|
| Profile | ✅ 19 Profile in 5 Gruppen, 28 Feature-Flags | `industry_profiles.go` |
| **Feature-Flags** | 🔌 `hasFeature()` hat **0 Aufrufer** → alle 28 Flags wirkungslos | `hooks/useIndustry.ts:42-43` |
| Beschriftungen | ✅ Wirken an 4 Stellen | `ProjectForm.tsx`, `ProjectList.tsx`, `ScannerPage.tsx`, `CompanyPage.tsx` |
| Scanner-Startkacheln | 🔌 Werden nie ausgelesen | — |
| Startkategorien | ⚠️ Nur Vorschautext im Setup, werden nie angelegt | `SetupWizard.tsx:452` |
| Definition | ⚠️ Doppelt gepflegt (Go 587 Z. / TS 547 Z.) und **bereits divergiert** | beide Dateien |
| **Setup-Wizard** | 🚫 Speichert Branche und Module **nie** — ruft nach dem Setup geschützte Endpunkte ohne Token auf, 401 wird verschluckt → jeder neue Mandant startet auf `event_tech` | `SetupWizard.tsx:171-186` |

### 13.4 Module

⚠️ Reines Ausblenden der Sidebar. Routen bleiben registriert und per Direkt-URL erreichbar, kein Backend prüft die Freischaltung. Die Modulliste ist dreifach dupliziert.

### 13.5 KI-Plattform

| Aspekt | Status | Fundstelle |
|---|---|---|
| Provider-Clients | ✅ Fünf vollständig implementiert (Claude, GPT-4o, Gemini, Mistral, Ollama) | `provider_service.go` |
| **Registrierung** | 🚫 Kein Provider wird je zur Laufzeit registriert → **alle** KI-Funktionen liefern „Provider nicht gefunden" | `cmd/server/main.go:54` |
| **Anonymisierung** | 🚫 Opt-in, wirkt nur auf ein Protokollfeld — der **gesendete Text läuft nie durch sie hindurch** | `ai_service.go:81` |
| Anonymisierungsqualität | ⚠️ Nur DE-Muster; Namens-Regex trifft jedes „Großwort Großwort" | `anonymization_service.go` |
| Request-Verarbeitung | ❌ `CreateAIRequest` speichert nur, ruft nie einen Provider; `pending` wird nie abgearbeitet | `ai_service.go:49-93` |
| Provider-Auswahl | ⚠️ Vier Funktionen hart auf `"mistral"`; `Priority` nie ausgewertet | `prediction_service.go:70,100,138,178` |
| Dashboard | ❌ Hartkodierte Fantasiewerte | `ai_service.go:271-288` |
| Gemini-Key | ⚠️ Steht in der Request-URL → landet in Logs | `provider_service.go:403` |
| Schema-Konflikt | 🚫 CHECK erlaubt `asset_creator`, Code sendet `smart_asset_creator` | `migrations/001:32` |

### 13.6 Workflows

| Aspekt | Status | Fundstelle |
|---|---|---|
| **Trigger-Dienst** | 🚫 Drei Methoden, die nur loggen und `nil` zurückgeben | `trigger_service.go:32-48` |
| **Aktionsausführung** | 🔌 416 Zeilen mit echtem SMTP und HTTP — **nie aufgerufen** | `action_executor.go` |
| Schritte anlegen | 🚫 Keine Route → Workflows bleiben leere Hüllen | `router.go` |
| Ausführungsschleife | ❌ Kein Step-Executor; Status wird nie `completed`/`failed` | `workflow_service.go:118-142` |
| Scheduler | ❌ Nicht vorhanden | — |
| Bedingungen | ⚠️ Können nicht auf Instanzdaten zugreifen | `action_executor.go:334-354` |
| Webhook | ⚠️ Keine SSRF-Absicherung | `action_executor.go:161-179` |

### 13.7 Benachrichtigungen

| Aspekt | Status | Fundstelle |
|---|---|---|
| **Versand** | 🚫 Legt nur einen DB-Satz an, ruft **keinen Kanal** auf | `notification_service.go:40-77` |
| **Digest-Dienst** | 🔌 Vollständig implementiert, **0 Aufrufer** | `digest_service.go:70` |
| SMTP-Treiber | ✅ Voll implementiert | `channel_drivers.go` |
| Web-Push | ❌ Ohne VAPID-JWT und ohne Verschlüsselung → wird abgelehnt | `channel_drivers.go:241-253` |
| In-App | ❌ Nur ein Logeintrag | `channel_drivers.go:293-307` |
| IMAP | ❌ Explizit nicht implementiert | `mail_service.go:328-333` |
| Postfach-Passwörter | ⚠️ Klartext | `migrations/002` |
| System-Mails | ✅ Einladung, Reset, Buchungsanfrage funktionieren direkt | — |

### Ausbauaufträge Plattform

**13.1 — RSA-Schlüssel persistieren** 🚫
Aus Datei oder Env laden, beim ersten Start erzeugen und ablegen. **Höchste Priorität** — ohne das ist kein Produktivbetrieb und keine Skalierung möglich.

**13.2 — RBAC durchsetzen** 🚫🔌
Das vollständige Permission-Modell existiert bereits. An alle schreibenden Endpunkte hängen, besonders Rollenzuweisung, Benutzerlöschung, Mandantenanlage und Backup-Wiederherstellung.

**13.3 — Auth-Middleware in allen Services**
`JWTAuthMiddleware` + `TenantMiddleware` überall einbinden. Mandant ausschließlich aus verifizierten Claims. Query-Parameter-Variante im insurance-service entfernen.

**13.4 — Session-Invalidierung** 🔌
Session-Manager produktiv einbinden, Logout serverseitig wirksam machen, Refresh-Token-Blacklist.

**13.5 — QR-Login-Migration** 🚫
Fehlende Tabelle anlegen.

**13.6 — Setup-Wizard reparieren** 🚫
Branche und Module als Teil des Setup-Requests übergeben und in derselben Transaktion speichern. Zusätzlich die sieben verworfenen Firmenfelder (USt-IdNr, Handelsregister, Geschäftsführer, Adresse) übernehmen — sie werden für §14-Pflichtangaben gebraucht.

**13.7 — Branchenprofile wirksam machen**
Eine Definitionsquelle (Backend), Frontend zieht sie über den vorhandenen Endpunkt. `hasFeature()` an den Stellen einsetzen, an denen Module und Felder branchenabhängig sind. Startkategorien beim Setup tatsächlich anlegen.

**13.8 — KI-Provider registrieren** 🔌
Beim Start aus der Datenbank in die Laufzeit-Map laden, API-Keys sicher speichern (Spalte fehlt), Fallback-Kette nach `Priority`. **Ein kleiner Fix, der das gesamte KI-Modul aktiviert.**

**13.9 — Anonymisierung in den Sendepfad** 🚫
Vor jedem Provider-Aufruf anwenden, nicht daneben. Für eine Software, die mit Datenschutzkontrolle wirbt, ist das die entscheidende Stelle.

**13.10 — Workflow-Engine bauen**
Step-Executor, der die vorhandene Aktionsausführung nutzt, Statusübergänge, Retry und Timeout auswertet. Trigger-Dienst implementieren. Routen für Schritte. Scheduler. SSRF-Schutz für Webhooks.

**13.11 — Benachrichtigungsversand anschließen** 🔌
`SendNotification` an die Treiber hängen, `ProcessScheduledNotifications` per Ticker starten. Web-Push mit VAPID-JWT und Payload-Verschlüsselung. In-App über WebSocket oder Polling.

**13.12 — Modulschaltung serverseitig**
Deaktivierte Module auch im Backend ablehnen, nicht nur ausblenden.

---

# Teil C — Querschnittsbefunde

## C.1 Event Sourcing existiert nicht

Die Architekturdokumentation beschreibt KurrentDB als Source of Truth mit CQRS-Projektionen. **Der Code benutzt es nirgends.**

Verifiziert über alle Services: `NewKurrentDBClient`, `NewEventStoreFromEnv`, `EventStoreAdapter`, `AppendToStream`, `ProjectionManager` — **je 0 Treffer**. Die Domain-Objekte betten `AggregateRoot` ein und erzeugen Events in ein In-Memory-Slice `Changes`, das **nie persistiert oder publiziert** wird. `docker-compose.yml` startet trotzdem einen KurrentDB-Container und injiziert Verbindungsvariablen in alle Services, die niemand liest.

**Persistenz ist zu 100 % direktes PostgreSQL.** Es gibt keinen Message-Bus, keine Outbox, keine service-übergreifenden Events. Kopplung läuft über wenige direkte HTTP-Aufrufe.

**Empfehlung:** Das ist kein Fehler, sondern eine bewusste Vereinfachung aus einer früheren Session (dokumentiert in `lessons_learned.md`: „erst CRUD, dann Event Sourcing"). Aber die Dokumentation muss das abbilden, sonst baut jede nachfolgende KI auf einer Architektur auf, die es nicht gibt. Entweder Event Sourcing nachrüsten oder die tote Infrastruktur samt Container entfernen. Für die Betriebsgröße ist Letzteres wahrscheinlich richtig.

## C.2 Migrationen sind unzuverlässig

`scripts/migrate-internal.sh:51` führt jede `.sql` mit `psql … || true` aus und filtert die Ausgabe. **Jeder Migrationsfehler wird stillschweigend verschluckt.** Es gibt keine Versionstabelle; alle Dateien laufen bei jedem Start erneut.

Konkrete Folgen:

| Problem | Betroffen |
|---|---|
| FK auf nicht existierende `tenants`-Tabelle (liegt in anderer DB) | expense, transport, maintenance — Tabellen werden vermutlich **gar nicht angelegt** |
| Zwei widersprüchliche Migrationsverzeichnisse | warehouse, scanner |
| Doppelte Migrationsnummern | auth (`005` ×2), expense (`003` ×2), project (`002` ×2) |
| Fehlende Migration | `auth.qr_login_tokens` |
| Constraint blockiert Fachlogik | invoice (`immutable_finalized`, Quote-Status, Gutschrift-Steuersatz) |
| Typ-Inkonsistenz | `tenant_id` mal `UUID`, mal `VARCHAR(255)`, mal `VARCHAR(100)` |

**Ausbauauftrag C.2:** Migrationswerkzeug mit Versionstabelle einführen (golang-migrate o. ä.), Fehler zum Abbruch führen lassen, alle Nummernkollisionen auflösen, DB-übergreifende FKs entfernen. **Das gehört in Phase 1** — solange Migrationen still scheitern, ist jede Fehlersuche Raten.

## C.3 Gateway-Routing ist lückenhaft

Es existieren zwei widersprüchliche Traefik-Konfigurationen (`docker-compose.yml`-Labels und `infra/traefik/dynamic/routers.yml`) mit unterschiedlichen Pfadpräfixen. Nicht geroutet und damit von außen unerreichbar:

- `/api/v1/transport/*` — der gesamte transport-service (Regel lautet `/api/v1/vehicles`)
- `/api/v1/warehouses`, `/api/v1/warehouse/*`, `/api/v1/zones`, `/api/v1/racks`, `/api/v1/bays`
- `/api/v1/equipment-types`, `/api/v1/flightcases`
- `/api/v1/reservations`, `/api/v1/customers`, `/api/v1/contacts`
- `/api/v1/insurance/dashboard`

Der Vite-Dev-Proxy ist korrekt konfiguriert und umgeht Traefik — deshalb fällt die Lücke in der Entwicklung nicht auf, sondern erst in Produktion.

**Ausbauauftrag C.3:** Auf eine Routing-Quelle konsolidieren, alle Pfade abdecken, einen Smoke-Test ergänzen, der jeden registrierten Endpunkt über das Gateway anspricht.

## C.4 ID-Erzeugung

Die meisten Services bilden IDs aus einem **djb2-Hash** deterministischer Eingaben (`equip_%d` aus `tenantID+barcode`, `exp_%d` aus `tenant+vendor+date`, Projekte aus `tenant+name`). Folgen: Zwei Projekte gleichen Namens kollidieren; zwei Belege desselben Lieferanten am selben Tag überschreiben sich; nach Löschen und Neuanlegen entsteht dieselbe ID. Der Kommentar im Code sagt selbst „in real scenario, use UUID".

**Ausbauauftrag C.4:** Durchgängig UUIDv7 (zeitsortierbar). Migration mit Mapping-Tabelle für Bestandsdaten.

## C.5 Frontend↔Backend-Vertragsbrüche

Über 20 verifizierte Fälle, in denen das Frontend Routen aufruft, die es nicht gibt, oder Felder erwartet, die das Backend nicht liefert. Besonders folgenreich:

- `invoiceApi.sendEmail` → `/send-email`, Backend hat `/send` → „Per E-Mail senden" ist tot
- `invoiceApi.update(id, {status:'paid'})` → Command hat kein `status`-Feld → „Als bezahlt markieren" wirkungslos
- `auditApi.verify()` erwartet `{valid}`, Backend liefert `{is_valid}` → Fallback zeigt **immer „gültig"**
- `expenseApi.uploadReceipt` → Route existiert nicht
- Der notification-service liefert PascalCase-JSON (fehlende JSON-Tags), das Frontend braucht einen Normalisierungs-Workaround

**Ausbauauftrag C.5:** OpenAPI-Spezifikation je Service generieren, Frontend-Client daraus erzeugen. Beendet diese Fehlerklasse dauerhaft.

## C.6 Testabdeckung

10 Go-Testdateien für 18 Services, 1 Frontend-Test, 1 E2E-Spezifikation — bei rund 120.000 Zeilen Code. Die Make-Targets existieren, dahinter steckt fast nichts.

**Ausbauauftrag C.6:** Für jede Kernfunktion, die in Phase 0–2 angefasst wird, mindestens einen Integrationstest, der den vollständigen Weg vom Endpunkt bis in die Datenbank prüft. Priorität: Verfügbarkeitsrechnung, Rechnungslebenszyklus, Scanner-Buchung.

---

# Teil D — Priorisierte Roadmap

## Phase 0 — Verdrahtung (Tage, nicht Wochen)

Alles hier ist bereits gebaut und nur nicht angeschlossen. Bester Aufwand-Nutzen-Schnitt im gesamten Projekt.

| # | Auftrag | Wirkung |
|---|---|---|
| 13.1 | RSA-Schlüssel persistieren | Anmeldung übersteht Neustarts, Skalierung möglich |
| 13.8 | KI-Provider registrieren | Gesamtes KI-Modul inkl. Belegerkennung wird lebendig |
| 13.11 | Benachrichtigungsversand anschließen | Benachrichtigungen kommen an |
| 13.6 | Setup-Wizard reparieren | Branche und Module werden gespeichert |
| 6.4 | Wartungszyklus schließen | Wartungspläne werden fortgeschrieben |
| 6.5 | Checklistenergebnisse persistieren | Prüfprotokolle sind nicht mehr wertlos |
| 2.4 | Zustandshistorie aktivieren | Gerätehistorie wird sichtbar |
| 8.1 | Traefik-Routing transport | Service wird überhaupt erreichbar |
| 11.6 | Budgetverbrauch aktivieren | Budgetwarnung funktioniert |

## Phase 1 — Sicherheit und Fundament

| # | Auftrag |
|---|---|
| 13.2 | RBAC durchsetzen — solange jeder Nutzer Backups einspielen kann, ist Mehrbenutzerbetrieb ausgeschlossen |
| 13.3 | Auth-Middleware in allen Services, Mandant nur aus Claims |
| 13.9 | Anonymisierung in den KI-Sendepfad |
| 10.14 | Signatur-Endpunkt absichern |
| C.2 | Migrationswerkzeug mit Versionstabelle, alle Blocker aus Teil E |
| 2.1 | Lagerstruktur-Konflikt auflösen |
| C.3 | Gateway-Routing konsolidieren |

## Phase 2 — Das Herzstück

| # | Auftrag |
|---|---|
| 1.1–1.5 | Verfügbarkeits- und Dispositionsengine |
| 5.1–5.2 | Scanner-Vertrag und echte Buchung |
| 5.6 | Inventur funktionsfähig |
| 4.1–4.2 | Packliste aus Reservierungen, Lagerplatzsortierung |
| 3.1–3.3 | Angebot → Auftrag → Projekt |

## Phase 3 — Kaufmännische Nutzbarkeit

| # | Auftrag |
|---|---|
| 10.1–10.2 | Rechnungslebenszyklus entsperren, Repository vervollständigen |
| 10.6 | Rechnungs-PDF mit §14-Pflichtangaben |
| 10.3–10.5 | GoBD: Hash-Chain, Unveränderbarkeit, Nummernvergabe |
| 10.10 | Audit-Trail in Betrieb nehmen |
| 11.1–11.2 | Vorsteuerberechnung und Migrationen der Ausgaben |

## Phase 4 — Differenzierung

| # | Auftrag |
|---|---|
| 9.1–9.3 | Fehlmengen echt berechnen, Frontend an federation-service |
| 9.4–9.8 | mTLS, Gegenseitigkeit, Zumietware als Bestand |
| 12.1–12.3 | KPIs mit echten Datenquellen |

## Phase 5 — Regulatorik und Compliance

| # | Auftrag |
|---|---|
| 10.7 | E-Rechnung (XRechnung/ZUGFeRD) — Frist läuft |
| 10.9 | DATEV-Export im echten Format |
| 6.1–6.3 | Prüffristen, Grenzwerte, Konsequenzen |
| 7.1–7.2 | Zeiterfassung und ArbZG |
| 8.6 | Lenk- und Ruhezeiten |

---

## Eine realistische Einordnung

Der beschriebene Umfang entspricht dem etablierter Systeme wie Rentman oder Easyjob, in denen Jahre an Entwicklung stecken. Die Codebasis hat eine gute Struktur und deckt erstaunlich viel Fläche ab — aber die Tiefe fehlt fast überall.

**Empfehlung:** Nicht alle 13 Bereiche parallel ausbauen. Phase 0 bis 2 vollständig fertigstellen, mit echten Daten im eigenen Betrieb testen, und erst dann weitergehen. Eine Software, die Verfügbarkeit, Scannen und Packliste zuverlässig kann, ist bereits nutzbar. Eine, die alles halb kann, ist es nicht.

**Vor dem Einsatz bei zahlenden Kunden zwingend:** Phase 1 (Sicherheit) und Phase 3 (GoBD, §14 UStG). Eine Rechnungssoftware, die keine rechtskonformen Rechnungen erzeugt, ist ein Haftungsrisiko.

---

# Teil E — Blocker-Anhang mit Fundstellen

Befunde, die eine Funktion vollständig blockieren. Nach Schweregrad sortiert.

## E.1 Betrieb unmöglich

| # | Befund | Fundstelle |
|---|---|---|
| E-1 | RSA-Keypair bei jedem Start neu erzeugt | `auth-service/cmd/server/main.go:53` |
| E-2 | `immutable_finalized CHECK (status='draft')` sperrt jede Rechnung | `invoice-service/migrations/001:41` |
| E-3 | Quote-CHECK ohne `confirmed`, obwohl Code und Route es setzen | `invoice-service/migrations/002:25` |
| E-4 | `credit_notes.tax_rate DECIMAL(5,4)` vs. Wert `19` → overflow | `invoice-service/migrations/006:12` |
| E-5 | FKs auf nicht existierende `public.tenants` | expense `001:24, 002:11, 003:56`; transport `001:16,37`; maintenance `001:17,29,48,66,78,103` |
| E-6 | Zwei widersprüchliche Migrationsverzeichnisse, Repos sprechen gegen das nicht angewendete Schema | `warehouse-service/migrations/` vs. `internal/infrastructure/migrations/` |
| E-7 | Migration für `auth.qr_login_tokens` fehlt komplett | repo-weit |
| E-8 | Traefik-Regel `/api/v1/vehicles` vs. tatsächlich `/api/v1/transport/*` | `docker-compose.yml` |
| E-9 | Migrationsfehler werden verschluckt (`|| true`) | `scripts/migrate-internal.sh:51` |

## E.2 Sicherheit

| # | Befund | Fundstelle |
|---|---|---|
| E-10 | RBAC-Modell ist Dead Code; Rollencheck an 2 von ~40 Endpunkten | `auth-service/…/permissions.go:151-185` |
| E-11 | Kein Rollencheck bei Rollenzuweisung, Benutzerlöschung, Backup-Restore | `auth-service/…/router.go:81-111` |
| E-12 | Vier Services ohne jede Auth, Mandant aus rohem Header | ai, workflow, federation, reporting |
| E-13 | insurance-service: Mandant aus Query-Parameter, bei Detailabfragen gar nicht geprüft | `insurance/…/handlers.go:64,160,334` |
| E-14 | Öffentlicher Signatur-Endpunkt prüft keinen Token | `document-service/…/signature_service.go:185-225` |
| E-15 | Refresh-Token 7 Tage beliebig oft einlösbar, Logout wirkungslos | `auth-service/…/user_service.go:183-240` |
| E-16 | Privater EC-Schlüssel über API ausgeliefert, nie gespeichert | `federation/…/certificate_service.go:69-77` |
| E-17 | Kein mTLS trotz Produktversprechen | `federation/…/sharing_service.go:35` |
| E-18 | Demo-Modus und Body-Logging im Release-Build der App | `ScannerRepository.kt:40`, `di/AppModule.kt:50` |
| E-19 | Webhook-Aktion ohne SSRF-Schutz | `workflow/…/action_executor.go:161-179` |
| E-20 | HTML-Injection in Packliste und Rechnungs-HTML (kein Escaping) | `project_service.go:583-653`, `invoice_service.go:644-833` |
| E-21 | Payload-Bau per `fmt.Sprintf` ohne Escaping | `scanner/…/inventory_http.go:154,182` |
| E-22 | Panic bei Setup-Token < 8 Zeichen | `auth-service/…/setup_service.go:180` |

## E.3 Funktion vorgetäuscht

| # | Befund | Fundstelle |
|---|---|---|
| E-23 | GoBD-Prüfung im UI ist `setTimeout` + festes „bestanden" | `DocumentsPage.tsx:113-127` |
| E-24 | Audit-Kette kann strukturell nie gültig sein (Zeitstempel doppelt erzeugt) | `audit_service.go:53,76` |
| E-25 | Audit-Verifikation im Frontend zeigt wegen Feldnamen-Mismatch **immer** „gültig" | `AuditPage.tsx:44` |
| E-26 | Kein Service schreibt Audit-Einträge | repo-weit |
| E-27 | Fehlmengenseite zeigt wegen Enum-Bug **immer** null | `ShortagesPage.tsx:66` |
| E-28 | „Zumietung anfragen" sendet nichts, zeigt nur einen Toast | `ShortagesPage.tsx:116-121` |
| E-29 | Federation erfindet Rechnungs-ID per `uuid.New()` | `sharing_service.go:145` |
| E-30 | Inventur-Soll wird nie befüllt → jeder Scan „unerwartet" | `warehouse/…/inventory_check_service.go:230` |
| E-31 | Scanner-Check-Out/Check-In ändert keinen Bestand | `CheckOutViewModel.kt:290-307` |
| E-32 | Verfügbarkeitsprüfung: Bestand hart auf 1, keine Zeiträume | `availability_service.go:62-64` |
| E-33 | KPI-Werte hart auf `nil` | `kpi_service.go:46` |
| E-34 | KI-Dashboard mit erfundenen Zahlen | `ai_service.go:271-288` |
| E-35 | Notification-Dashboard mit hartkodierter Zustellrate | `notification_service.go:121-133` |
| E-36 | Reparatur-Modal ohne Funktion | `WorkshopPage.tsx:560-600` |
| E-37 | Wartungsdetailseite zu 100 % Mock | `MaintenanceDetail.tsx:35` |
| E-38 | Banking-Import speichert nichts | `invoice/…/handlers.go:1030` |
| E-39 | Vorsteuerberechnung mathematisch falsch | `expense.go:124-127` |
| E-40 | Zeiterfassung ignoriert übergebene Zeiten → immer 0 Stunden | `time_record_handlers.go:112-165` |
| E-41 | Setup-Wizard speichert Branche und Module nie | `SetupWizard.tsx:171-186` |
| E-42 | Checklistenergebnisse werden durch leeres Array ersetzt | `maintenance_task_service.go:183-187` |
| E-43 | CSV-Prüftermin wird explizit verworfen | `maintenance/…/handlers.go:779-780` |
| E-44 | Claim-Workflow nicht durchlaufbar (`reported` ist Sackgasse) | `insurance/…/commands.go:281-298` |
| E-45 | Fotoupload im Schadensfall speichert nichts | `insurance/…/handlers.go:465-506` |
| E-46 | Audit-Export verwirft das erzeugte ZIP | `export_service.go:32-106` |
| E-47 | Berichtsgenerierung verwirft die Definition, wird nie fertig | `report_service.go:86-121` |
| E-48 | Workflow-Trigger sind drei leere Methoden | `trigger_service.go:32-48` |

## E.4 Fertig gebaut, nicht angeschlossen 🔌

| # | Komponente | Fundstelle |
|---|---|---|
| E-49 | Session-Manager (nur in Tests aufgerufen) | `auth-service/…/session_manager.go` |
| E-50 | Permission-Modell (0 Aufrufer) | `auth-service/…/permissions.go` |
| E-51 | KI-Provider-Registrierung (nie aufgerufen) | `ai-service/cmd/server/main.go:54` |
| E-52 | Workflow-ActionExecutor (416 Z., nie aufgerufen) | `workflow/…/action_executor.go` |
| E-53 | Digest-Versand (0 Aufrufer) | `notification/…/digest_service.go:70` |
| E-54 | Schöner Rechnungs-PDF-Generator (nie aufgerufen) | `chromedp_generator.go:83-266` |
| E-55 | Sechs PDF-Templates (nur eines genutzt, über ungültigen Pfad) | `invoice-service/…/pdf/templates/` |
| E-56 | Equipment-Historien-Repository (nie instanziiert) | `inventory/…/equipment_history_postgres.go` |
| E-57 | ZPL-Etikettendienst (kein Endpunkt) | `warehouse/…/zpl_service.go` |
| E-58 | Lieferschein-Erzeugung (keine Route, nur Noop-Client) | `transport/…/services.go:400-445` |
| E-59 | CrewClient für Fahrervalidierung (nie instanziiert) | `transport/…/crew_client.go` |
| E-60 | OCR-Prozessor (`_ = ocrProcessor`) | `document-service/cmd/server/main.go:59` |
| E-61 | `planRepo` im Wartungsdienst (injiziert, nie benutzt) | `maintenance_task_service.go:15,22,28` |
| E-62 | Qualifikationsprüfung (nur im Fahrer-Handler genutzt) | `crew/…/entities.go:163-187` |
| E-63 | Auto-Mahnlauf (keine Route, kein Scheduler) | `dunning_service.go:152-215` |
| E-64 | Atomare Nummernvergabe (korrekt implementiert, ungenutzt) | `invoice_postgres.go:334-351` |
| E-65 | `hasFeature()` für 28 Branchen-Flags (0 Aufrufer) | `frontend/…/useIndustry.ts:42-43` |
| E-66 | KPI-Snapshot-Job (0 Aufrufer) | `kpi_service.go:27` |
| E-67 | `pkg/common`: `response`, `errors`, `health`, `server` (0 Services nutzen sie) | `pkg/common/` |
| E-68 | `FindOverlapping` für Reservierungen (0 Aufrufer) | `reservation_postgres.go:164-199` |

---

## Nicht verifiziert

- Es wurde **nichts zur Laufzeit ausgeführt** — kein Build, keine Datenbank, keine Tests. Alle Aussagen über Laufzeitfehler sind statisch aus dem Quellcode abgeleitet.
- Ob das Projekt aktuell kompiliert, wurde nicht geprüft.
- Welche Traefik-Konfiguration produktiv greift, ließ sich aus dem Code nicht abschließend bestimmen.
- Deployment, Monitoring (Prometheus/Grafana/Loki) und CI wurden nur gestreift.
- Der Inhalt der RFID-Bibliothek `rfiddrive-release.aar` (Binärdatei) wurde nicht analysiert; Aussagen dazu stützen sich auf die Aufrufe im Kotlin-Code.
- Die vorhandenen Markdown-Dokumente in den Services (`IMPLEMENTATION.md`, `ARCHITECTURE.md`, `COMPLETION_CHECKLIST.md` u. a.) wurden bewusst **nicht** als Quelle verwendet — sie widersprechen dem Code an mehreren Stellen.
