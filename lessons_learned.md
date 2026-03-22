# Lessons Learned

## Regeln aus vergangenen Sessions
- [2026-03-20] Regel: Bei geklonten Repos IMMER alle Branches prüfen (`git branch -a`), nicht nur den Default-Branch. Besonders bei Repos die aus vorherigen Claude-Sessions stammen – dort liegen oft Feature-Branches mit umfangreicher Vorarbeit.
  - Grund: Branch `claude/event-tech-inventory-system-H42yD` enthielt 10.000 Zeilen detaillierte Planungsdokumente (Microservice-Architektur, DB-Schema, Module-Specs, Implementierungsphasen) die ich erst spät entdeckt habe. Dadurch habe ich eigene Architektur-Dokumente erstellt die teilweise redundant zur bestehenden Planung waren.

- [2026-03-20] Regel: Vor dem Erstellen neuer Architektur-Dokumente IMMER prüfen ob bereits Planungsdokumente existieren (in allen Branches, docs/, planning/, etc.). Bestehende Planung als Grundlage nehmen, nicht neu erfinden.
  - Grund: Meine Architektur (Monolith, Caddy) widersprach der bestehenden Planung (17 Microservices, KurrentDB, Traefik, CQRS). Die bestehende Planung war deutlich umfangreicher und durchdachter.

- [2026-03-20] Regel: Wenn eine bestehende Planung eine Microservice-Architektur mit Event Sourcing vorsieht, diese NICHT eigenmächtig zu einem Monolithen vereinfachen. Die Architektur-Entscheidungen wurden bewusst getroffen.
  - Grund: Ich hatte eigenständig eine einfachere Monolith-Architektur geplant, obwohl die bestehende Planung bewusst MSA + CQRS + KurrentDB gewählt hatte.

- [2026-03-20] Regel: Zielgruppen-Angaben des Nutzers immer sofort in ALLE relevanten Dokumente übernehmen (README, Planung, Architektur). Nicht nur intern merken.
  - Grund: Nutzer hat klargestellt dass die Software für VT-Firmen bis 100 Mitarbeiter sein soll, nicht nur 3-30. Das hat Auswirkungen auf Multi-Tenancy, Performance-Anforderungen und Rollenhierarchien.

## Session-Log
### 2026-03-20 - Architekturplanung Lagerverwaltung & Rechnungssoftware
- Aufgaben: Detaillierte Architektur erstellen (DB-Schema, API-Design, Frontend-Architektur, Docker-Setup, Federation-Protokoll, MVP-Phasenplan)
- Fehler:
  1. Nicht alle Branches im Repo geprüft → umfangreiche Vorarbeit auf Feature-Branch übersehen
  2. Eigene Architektur-Dokumente erstellt die teilweise im Widerspruch zur bestehenden Planung standen (Monolith vs. Microservices)
- Neue Regeln: Siehe oben (3 neue Regeln)
- Ergebnis: Bestehende Planung vom Feature-Branch als Grundlage übernommen. Eigene Dokumente (Frontend-Architektur, Federation-Protokoll, Deployment) ergänzen die bestehende Planung in Detailbereichen.

### 2026-03-21 – Autonome Nacht-Implementierung: 18 Microservices + Frontend

**Projekt:** RentFlow – Vollständige Implementierung der Microservice-Architektur
**Was passiert ist:** Komplette Implementierung von 18 Go-Microservices, shared pkg/common Library, React-Frontend, Docker-Infrastruktur und CI/CD-Pipelines in einer einzigen Session. 257 Go-Dateien (~36.235 LOC), 48 Frontend-Dateien (~2.532 LOC), 32 SQL-Migrationen, 41 Docker/CI/Config-Dateien.

**Erkenntnisse & Regeln:**

- [2026-03-21] Regel: Bei großen Monorepo-Implementierungen mit vielen Services den gleichen hexagonalen Architektur-Boilerplate als Template nutzen. Jeder Service folgt exakt dem gleichen Muster: `cmd/server/main.go` → `internal/domain/` → `internal/application/` → `internal/adapters/http/` → `internal/infrastructure/repositories/` → `migrations/`. Das spart enorm Zeit und reduziert Fehler.
  - Grund: Durch konsistentes Template konnte ich 18 Services in einer Session bauen, ohne bei jedem Service die Struktur neu überlegen zu müssen.

- [2026-03-21] Regel: Go-Packages mit identischen Namen wie stdlib-Packages (z.B. `internal/adapters/http`) verursachen Import-Konflikte mit `net/http`. Entweder den Package-Namen ändern (z.B. `httphandler`) oder den stdlib-Import aliasieren (`nethttp "net/http"`).
  - Grund: In `main.go` importieren wir sowohl `net/http` für den Server als auch unser eigenes `internal/adapters/http` Package. Das Go-Typsystem löst das nicht automatisch auf.

- [2026-03-21] Regel: Bei Event-Sourcing-Architektur die Services zuerst mit direktem PostgreSQL-Zugriff bauen (CRUD), dann in einer zweiten Phase Event Sourcing (KurrentDB → Projektion) darüberlegen. Beides gleichzeitig zu implementieren verdoppelt die Komplexität.
  - Grund: Die pkg/common Events/KurrentDB-Abstraktionen sind implementiert, aber die Services nutzen noch direktes PostgreSQL. Das ist der richtige Ansatz für Phase 0 – Event Sourcing kommt in Phase 1.

- [2026-03-21] Regel: Bei Go-Microservices `go.work` für lokale Entwicklung nutzen, aber in Docker-Builds jedes Modul einzeln bauen. `go.work` wird NICHT in den Container kopiert – jeder Service hat sein eigenes `go.mod`.
  - Grund: Docker-Multi-Stage-Builds funktionieren pro Service. `go.work` ist nur ein Dev-Convenience-Tool.

- [2026-03-21] Regel: Parallele Sub-Agents für unabhängige Services nutzen. Services wie auth, inventory, invoice, project haben keine Build-Abhängigkeiten untereinander und können gleichzeitig implementiert werden.
  - Grund: Durch Parallelisierung konnten 18 Services in einer Nacht gebaut werden statt sequenziell Tage zu brauchen.

- [2026-03-21] Regel: Bei GoBD-konformer Rechnungssoftware die Hash-Chain (SHA-256) und lückenlose Nummernvergabe von Anfang an in der Datenbankstruktur vorsehen – nicht nachträglich hinzufügen. Felder: `hash`, `previous_hash`, `invoice_number` (UNIQUE, sequentiell).
  - Grund: GoBD-Konformität ist eine Grundvoraussetzung für deutsche Rechnungssoftware. Nachträgliches Einfügen einer Hash-Chain in bestehende Daten ist problematisch.

- [2026-03-21] Regel: SHA-256 + Salt ist KEINE sichere Passwort-Hashing-Methode. Bei der nächsten Iteration MUSS auf bcrypt oder argon2id gewechselt werden.
  - Grund: Der Auth-Service verwendet aktuell SHA-256 + Salt im PasswordManager. Das ist für einen MVP akzeptabel, aber für Produktion unsicher (zu schnell zu bruteforcen).

- [2026-03-21] Regel: Bei Frontend-Implementierung mit TypeScript immer zuerst die Type-Definitionen erstellen, dann die API-Service-Layer, dann die Komponenten. Types → Services → Components → Pages.
  - Grund: Spart Refactoring, weil Typen von Anfang an konsistent sind.

- [2026-03-21] Regel: Docker-Compose mit 20+ Containern braucht Ressourcen-Limits und Health-Checks. Ohne `mem_limit` und `healthcheck` kann ein einzelner Container den ganzen Host (Hetzner CX21, 4GB RAM) zum Absturz bringen.
  - Grund: 18 Services + PostgreSQL + KurrentDB + Redis + Traefik = 22 Container auf 4GB RAM. Jeder Container braucht ein hartes Limit.

- [2026-03-21] Regel: `gh` CLI ist im Container nicht verfügbar. Git-Operationen mit PAT-Token direkt über HTTPS machen: `git remote set-url origin https://<PAT>@github.com/user/repo.git`.
  - Grund: In vorheriger Session gescheitert, diesmal direkt PAT-basiert gearbeitet.

**Offene Punkte für nächste Session:**
1. ~~`go build` für alle 18 Services testen~~ ✅ Erledigt (21.03.2026)
2. ~~`go mod tidy` für alle Module~~ ✅ Erledigt (21.03.2026)
3. ~~Frontend `npm install` + Dev-Server testen~~ ✅ Erledigt (21.03.2026)
4. Unit-Tests für Core-Services schreiben (auth, inventory, invoice)
5. KurrentDB Event Sourcing tatsächlich verdrahten (aktuell nur PostgreSQL direkt)
6. Passwort-Hashing auf bcrypt/argon2id upgraden
7. ~~Import-Konflikte `net/http` vs `internal/adapters/http` auflösen~~ ✅ Erledigt (21.03.2026)
8. ~~Aufräumen: generierte Markdown-Dateien im Root~~ ✅ Erledigt (21.03.2026)

**Kategorie:** Architektur | Implementierung | DevOps

### 2026-03-21 – Fehlerüberprüfung: Alle 18 Services kompilieren

**Was passiert ist:** Systematische Kompilierungsprüfung aller Go-Services und des Frontends. 151 Dateien gefixt, 5 Build-Runden bis 18/18 Services fehlerfrei kompilieren.

**Behobene Fehler-Kategorien:**
1. `*logger.Logger` pointer-to-interface statt `logger.Logger` (63 Dateien)
2. Import-Pfade `github.com/rentflow/` → `github.com/jeckersberger/rentflow/` (5 Dateien in pkg/common)
3. `cfg.DatabaseURL` → `cfg.ConnectionString()` (18 main.go Dateien)
4. Repository-Konstruktoren: `*pgxpool.Pool` / `*sql.DB` → `*database.PostgresPool` (14 Dateien)
5. `QueryContext`/`ExecContext` → `Query`/`Exec` (PostgresPool-API hat kein Context-Suffix)
6. Invoice TaxRate: Custom Type brauchte `float64()` Casts
7. Auth-Service: `net/http` Package-Namenskonflikt mit eigenem `internal/adapters/http`
8. Frontend: `jsx: "react-jsx"` fehlte in tsconfig.json, `terser` als DevDependency

**Neue Regeln:**
- [2026-03-21] Regel: Nach jeder Code-Generierung sofort `go build ./...` laufen lassen. Niemals Code committen der nicht kompiliert.
- [2026-03-21] Regel: Interface-Typen in Go NIEMALS als Pointer übergeben (`*logger.Logger` ist falsch, `logger.Logger` ist richtig). Interfaces sind bereits Referenztypen.
- [2026-03-21] Regel: Bei eigenem Package namens `http` in `internal/adapters/http/`: In `main.go` IMMER Aliases verwenden (`nethttp "net/http"`, `svchttp "...internal/adapters/http"`).

**Offene Punkte:**
1. Unit-Tests für Core-Services schreiben
2. KurrentDB Event Sourcing verdrahten
3. Passwort-Hashing auf bcrypt/argon2id upgraden
4. TypeScript strict-mode Fehler fixen (unused imports in einigen Pages)

**Kategorie:** Qualitätssicherung | Bugfix

### 2026-03-21 – Phase 1 Implementierung & Verifikation

**Projekt:** RentFlow – Phase 1: Auth-Service Security, Inventory-Service Prüfung, Frontend Verifikation
**Was passiert ist:** Auth-Service um Redis-basiertes Session-Management, Rate-Limiting und Brute-Force-Schutz erweitert. Inventory-Service als bereits vollständig für Phase 1 identifiziert (35+ Endpoints, ILIKE-Suche, Price Engine, QR-Codes, CSV-Import). Frontend war ebenfalls bereits komplett implementiert (Login, Dashboard, Equipment CRUD, Projects, Invoices, Scanner, Warehouse, Settings).

**Erkenntnisse & Regeln:**

- [2026-03-21] Regel: Bei Vitest auf Windows mit React IMMER `define: { 'process.env.NODE_ENV': '"test"' }` in vitest.config.ts setzen. Ohne diese Definition wird React im Production-Build geladen und `act(...)` schlägt fehl mit "act(...) is not supported in production builds of React."
  - Grund: Windows setzt NODE_ENV nicht automatisch auf "test" bei Vitest-Runs. React's Production-Build hat act() deaktiviert.

- [2026-03-21] Regel: Vor dem Implementieren neuer Features IMMER den bestehenden Code vollständig lesen. Oft ist mehr implementiert als erwartet – besonders wenn vorherige Sessions parallel mit Sub-Agents gearbeitet haben.
  - Grund: Frontend war bereits mit 48+ Dateien vollständig implementiert (Login, Dashboard, Equipment CRUD, Projects, Invoices, Scanner, etc.). Hätte ich blind angefangen "Frontend Core Pages" zu implementieren, wäre alles doppelt gebaut worden.

- [2026-03-21] Regel: Cache-Interface in Go IMMER respektieren. Wenn das Interface nur Get/Set/Delete/Exists hat, KEINE Methoden verwenden die nur auf dem konkreten Struct existieren (z.B. IncrementInt, AppendToList). Stattdessen Get→Deserialize→Modify→Serialize→Set Pattern verwenden.
  - Grund: SessionManager musste Brute-Force-Counter und User-Session-Listen über das Cache-Interface implementieren, obwohl RedisCache zusätzliche Methoden hat. Das Interface könnte durch eine andere Implementierung ersetzt werden.

- [2026-03-21] Regel: GitHub Actions Runner-Zuweisung kann wiederholt fehlschlagen (runner_id=0, steps=[]). Das ist ein GitHub-Infrastruktur-Problem, kein Code-Problem. Bei wiederholtem Auftreten: Code lokal verifizieren (go build, npm test, tsc) und Runner-Problem dokumentieren. Nicht stundenlang auf grüne CI warten.
  - Grund: Alle 4 Workflows scheiterten zweimal hintereinander an Runner-Zuweisung, obwohl der Code lokal fehlerfrei kompiliert und alle Tests besteht.

- [2026-03-21] Regel: Windows CMD und PowerShell haben massive Quoting-Probleme bei Shell-Skript-Strings. Für komplexe Shell-Befehle (for-Schleifen, etc.) entweder einzelne Befehle pro Service ausführen oder Skript-Dateien mit PowerShell's Set-Content schreiben – nicht mit CMD echo.
  - Grund: Docker-basierte Build-Loops scheiterten wiederholt an Windows CMD Quoting (Unterminated quoted string). Einzelne docker run-Aufrufe pro Service funktionierten sofort.

**Offene Punkte für nächste Session:**
1. GitHub Actions Runner-Problem beobachten – wenn es persistiert, alternative CI (z.B. Self-Hosted Runner auf NAS) evaluieren
2. customer-service ist noch ein leeres Verzeichnis – implementieren wenn benötigt
3. Unit-Tests für Auth-Service SessionManager schreiben
4. Unit-Tests für Inventory-Service Core-Funktionen schreiben
5. KurrentDB Event Sourcing verdrahten (aktuell nur PostgreSQL direkt)
6. ~~Passwort-Hashing auf bcrypt/argon2id upgraden~~ ✅ Erledigt (21.03.2026) – Argon2id implementiert

**Kategorie:** Implementierung | Testing | DevOps

### 2026-03-21 – Phase 1 Abschluss: Code-Audit, PDF, Email, Scanner-Kamera

**Projekt:** RentFlow – Phase 1 Abschluss durch Code-basiertes Audit statt README-Vertrauen
**Was passiert ist:** Nutzer stellte fest, dass ich mich auf README-Checkboxen verlassen hatte statt den tatsächlichen Code zu prüfen. Daraufhin systematisches Audit aller 6 Milestones (M1.1-M1.6) mit parallelen Agents durchgeführt. 3 kritische Lücken identifiziert und geschlossen: PDF-Generierung, Email-Versand, Scanner-Kamera-Integration.

**Erkenntnisse & Regeln:**

- [2026-03-21] Regel: README-Checkboxen NIEMALS als Wahrheitsquelle für den Implementierungsstand verwenden. IMMER den tatsächlichen Code lesen und prüfen ob die Funktionalität wirklich implementiert ist (nicht nur ein Stub/Placeholder).
  - Grund: README zeigte Phase 1 als "fertig" an, aber der Code hatte leere Stubs für PDF-Generierung (nur HTML-Template ohne Rendering), keinen Email-Sender, und der Scanner nutzte keine echte Kamera. Das hätte ich blind als erledigt abgehakt.

- [2026-03-21] Regel: `chromedp` (Chrome DevTools Protocol) benötigt ab v0.15.0 mindestens Go 1.26. Bei Go 1.24 Projekten stattdessen `go-pdf/fpdf` v0.9.0 verwenden – reine Go-Bibliothek, kein Chrome/Chromium nötig, läuft überall.
  - Grund: `go get github.com/chromedp/chromedp` schlug fehl mit "requires go >= 1.26 (running go 1.24.13; GOTOOLCHAIN=local)". fpdf als Alternative ist leichtgewichtiger und hat keine externen Abhängigkeiten.

- [2026-03-21] Regel: Wenn Windows CMD zum Erstellen von Shell-Skripten für Docker verwendet wird, NIEMALS `echo` benutzen – es fügt `\r` (Carriage Return) ein, was Go-Tools als `malformed module path "\r"` interpretieren. Stattdessen die Linux-VM nutzen (`printf` von Bash) um Dateien mit Unix-Zeilenenden zu erzeugen.
  - Grund: `echo` in CMD schreibt Windows-Zeilenenden (\r\n). Docker-Container (Linux) interpretieren \r als Teil des Strings. `printf` auf der Linux-VM schreibt korrekte Unix-Zeilenenden.

- [2026-03-21] Regel: `npm config get omit` auf Windows prüfen! Wenn der Wert `dev` ist, werden DevDependencies (TypeScript, Vite, Vitest, ESLint) NICHT installiert. Fix: `npm install --omit=none` oder `npm config set omit ""`.
  - Grund: Frontend-Build und Tests schlugen fehl, weil `tsc` und `vite` fehlten. Die globale npm-Konfiguration hatte `omit=dev` gesetzt, was alle devDependencies übersprang. Das ist kein offensichtlicher Fehler – `npm install` läuft durch, aber die Tools fehlen.

- [2026-03-21] Regel: Bei Phase-Abschluss-Audits parallele Agents für unabhängige Milestones verwenden. Jeder Agent prüft einen Milestone (Code lesen, Stubs identifizieren, fehlende Features listen). Das beschleunigt das Audit um Faktor 3-4x.
  - Grund: 6 Milestones parallel geprüft statt sequenziell. Jeder Agent hatte einen klaren Fokus und konnte unabhängig arbeiten. Die konsolidierten Ergebnisse zeigten sofort die 3 kritischen Lücken.

- [2026-03-21] Regel: Passwort-Hashing wurde erfolgreich von SHA-256 auf Argon2id migriert (auth-service). Die Parameter (time=1, memory=64MB, threads=4, keyLength=32, saltLength=16) entsprechen OWASP-Empfehlungen. Argon2id ist der aktuelle Gold-Standard.
  - Grund: SHA-256 war nur für den MVP akzeptabel. Argon2id wurde im Rahmen der Phase-1-Fertigstellung implementiert.

**Geschlossene Lücken:**
1. **PDF-Generierung** (M1.4): `go-pdf/fpdf` implementiert – professionelles A4-Layout mit Header, Empfängerdaten, Positionstabelle, Summenbereich, Footer
2. **Email-Versand** (M1.4): SMTP-Sender mit PDF-Anhang, Multipart-MIME, Base64-kodiert, RFC-2045-konform
3. **Scanner-Kamera** (M1.6): `html5-qrcode` mit echter Kamera-Integration, QR-Code-Parsing (`rentflow://equipment/{uuid}`), Vibrations-Feedback, Batch-Modus

**Verbleibende Nice-to-haves (nicht Phase-1-blockierend):**
- ~~Command Palette (⌘K) für Frontend-Navigation~~ ✅ Implementiert
- ~~Dark Mode Toggle~~ ✅ Bereits vorhanden (themeStore + CSS Variables)
- Offline/Service Worker für Scanner (vite-plugin-pwa bereits konfiguriert)
- Equipment-History Audit Trail (Tabelle existiert, Handler gibt leer zurück)
- pg_trgm Suchoptimierung (ILIKE reicht für MVP)

**Kategorie:** Qualitätssicherung | Implementierung | Debugging

---

### 2026-03-21 – Session 3: Phase 1 Final – Frontend-Tests, Credit Notes, Optimistic Updates

**Projekt:** RentFlow Phase 1 Fertigstellung
**Was passiert ist:** Frontend-Tests brachen nach Hinzufügen von CommandPalette und ToastContainer, weil App.test.tsx die react-router-dom Mock nicht vollständig hatte (fehlte useNavigate, useLocation). Zusätzlich mussten Toast und CommandPalette als ganze Komponenten gemockt werden.

**Erkenntnisse:**

- [2026-03-21] Regel: Wenn neue Komponenten zu App.tsx hinzugefügt werden die eigene Hooks nutzen (useNavigate, useNotificationStore etc.), IMMER sofort die Mocks in App.test.tsx aktualisieren. Am sichersten: Komplexe Komponenten komplett mocken (`vi.mock('../components/Toast/Toast', () => ({ ToastContainer: () => null }))`).
  - Grund: CommandPalette nutzt useNavigate, ToastContainer nutzt useNotificationStore – beides nicht im bestehenden Mock. Statt einzelne Hooks zu mocken, ist es robuster die Komponenten selbst zu mocken.

- [2026-03-21] Regel: Bei Windows-Git-Repos die per Mount in Linux-VM bearbeitet werden: `git commit` IMMER auf dem Windows-Host ausführen (via Desktop Commander), nicht in der Linux-VM. Die VM hat keine Git-Identity konfiguriert und temp-Dateien verursachen "Operation not permitted"-Fehler.
  - Grund: Linux-VM kennt keine Git-Identity, und .git/objects temp-files haben Cross-Filesystem-Permissions-Probleme.

- [2026-03-21] Regel: CMD.exe auf Windows unterstützt KEINE `-m "text"` Git-Commits mit Leerzeichen/Sonderzeichen zuverlässig. Stattdessen commit message in Datei schreiben und `git commit -F commitmsg.txt` verwenden.
  - Grund: PowerShell kennt kein `&&`, CMD zerlegt `-m` Strings an Leerzeichen trotz Anführungszeichen.

- [2026-03-21] Regel: `@rollup/rollup-linux-x64-gnu` wird nur in der Linux-VM benötigt, NICHT in package.json committen. Es ist ein plattformspezifisches optionales Binary.
  - Grund: node_modules wurden auf Windows installiert (x64-msvc), Linux braucht x64-gnu. Nach Installation sofort aus package.json entfernen.

**Kategorie:** Testing | DevOps | Debugging
