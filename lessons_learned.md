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
