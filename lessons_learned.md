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
