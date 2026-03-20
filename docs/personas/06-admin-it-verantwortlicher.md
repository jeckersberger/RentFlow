# Der Admin / IT-Verantwortliche — "Marco mit SSH"

> "Ich will den Server einmal aufsetzen und dann soll es einfach laufen. Bitte keine Anrufe um 2 Uhr nachts."

---

## Steckbrief

| Feld | Details |
|------|---------|
| **Wer ist "der Admin"?** | Meistens der Geschäftsführer selbst (Marco), oder ein befreundeter IT-ler ("Kumpel der was mit Computern macht") |
| **Typisches Profil** | 30-45 Jahre, technikaffin, hat schon Docker-Container laufen, kann Linux-Grundlagen |
| **Technik-Affinität** | 4-5/5 — Docker, Linux, Netzwerk-Basics, hat schon mal einen Webserver aufgesetzt |
| **Zeitbudget für IT** | Minimal. Einmal aufsetzen, dann "nie wieder anfassen" (unrealistisch, aber das ist der Wunsch) |
| **Geräte** | Laptop/PC mit Terminal (SSH), Browser für Admin-Panel |

---

## Hintergrund — Warum diese Persona anders ist

In großen Firmen gibt es eine IT-Abteilung. In einer 5-Mann-VT-Firma gibt es: Marco, der mal einen Raspberry Pi aufgesetzt hat und seitdem "der Computer-Typ" ist. Oder seinen Kumpel Jens, der hauptberuflich Webentwickler ist und "mal am Wochenende drüberschaut."

Diese Persona ist keine Vollzeit-Rolle. Es ist eine Aufgabe die jemand nebenbei macht — mit begrenztem Zeitbudget, begrenzter Geduld, und der klaren Erwartung dass es "einfach funktioniert." Wenn die Installation länger als 30 Minuten dauert, wird der Admin nervös. Wenn das Update die Software kaputt macht, wird der Admin wütend. Und wenn er um 2 Uhr nachts eine E-Mail bekommt dass der Server down ist — während sein Kumpel auf dem Festival steht — dann macht er das nie wieder.

---

## Typische Szenarien

### Szenario 1: Erstinstallation

**Ausgangssituation:** Marco hat sich entschieden, die Software einzusetzen. Er hat einen kleinen Root-Server bei Hetzner (CX21, 4GB RAM, 40GB SSD, 5€/Monat). Docker ist installiert. Jetzt soll die Software drauf.

**Marcos Erwartung:**
```bash
git clone https://github.com/...
cd lagerverwaltung
cp .env.example .env
# .env editieren: Domain, DB-Passwort, Admin-E-Mail
docker compose up -d
```
→ Software läuft. Browser öffnen → Login-Screen → Admin-Account anlegen → fertig.

**Was Marco NICHT will:**
- 15 Konfigurationsdateien editieren bevor irgendwas startet
- Separate Installation von PostgreSQL, Redis, Nginx, Certbot...
- Docker-Images selbst bauen
- Port-Forwarding manuell konfigurieren
- SSL-Zertifikat manuell beantragen

**Ideale Installationszeit:** 15-30 Minuten (inkl. Server-Setup und Domain-Konfiguration)

### Szenario 2: Erstes Setup nach Installation

Marco öffnet die Software im Browser. Ein Setup-Wizard führt ihn durch:

1. **Firma anlegen:** Name, Adresse, Logo, Steuernummer, Bankverbindung
2. **Admin-Account:** E-Mail, Passwort (sein eigener Account)
3. **Erste Benutzer:** Lisa (Lager), Thomas (Buchhaltung), Kevin (Freelancer) — jeweils mit vordefinierter Rolle
4. **Lager-Struktur:** Hauptlager anlegen, optional Räume/Regale definieren
5. **Grundeinstellungen:** Sprache (DE), Währung (EUR), MwSt-Satz (19%), Zahlungsziel-Standard (14 Tage)
6. **Federation (optional):** "Möchtest du dich mit Partner-Firmen vernetzen?" → Kann auch später eingerichtet werden

**Dauer:** 10-15 Minuten.

### Szenario 3: Update einspielen

Marco sieht im Admin-Panel: "Update verfügbar: v1.3.0 → v1.4.0"

**Was er erwartet:**
- Changelog lesen: Was ist neu? Gibt es Breaking Changes? Braucht die DB eine Migration?
- Button: "Jetzt updaten"
- System macht automatisch: Backup → neues Docker-Image pull → Migration → Health-Check → fertig
- Wenn etwas schiefgeht: automatischer Rollback auf die vorherige Version
- Nach dem Update: "Was ist neu?"-Dialog beim nächsten Login

**Was Marco NICHT will:**
- SSH auf den Server um manuell `docker pull` auszuführen
- Migration-Skripte von Hand starten
- Downtime von mehr als 2-3 Minuten
- Update das die Datenbank zerstört (ohne Rollback)

### Szenario 4: Benutzer verwalten

Neuer Freelancer fängt an → Marco will ihm einen Account geben.

**Idealer Ablauf:**
1. Admin-Panel → Benutzer → "Neuer Benutzer"
2. Name, E-Mail eingeben
3. Rolle auswählen: "Freelancer" (vordefinierte Rolle mit eingeschränkten Rechten)
4. Einladungslink wird per E-Mail verschickt
5. Freelancer klickt auf Link, setzt Passwort, fertig

**Rollen (vordefiniert):**
- **Admin** — Voller Zugriff, System-Einstellungen, Benutzer-Verwaltung
- **Projektleiter** — Projekte, Angebote, Equipment-Planung, Rechnungsfreigabe
- **Lager** — Scan In/Out, Packlisten, Defekt-Meldung, keine Preise/Finanzen
- **Buchhaltung** — Rechnungen, Mahnungen, DATEV, keine Lager-Operationen
- **Freelancer** — Nur eigene Jobs, Packlisten, Zeiterfassung, keine Preise/Finanzen
- **Benutzerdefiniert** — Eigene Rollen konfigurierbar (für Spezialfälle)

### Szenario 5: Federation einrichten

Marco will SoundPro als Partner verknüpfen.

**Idealer Ablauf:**
1. Admin-Panel → Federation → "Partner hinzufügen"
2. Marco gibt SoundPros Server-Adresse ein: `soundpro.example.com`
3. System generiert einen Einladungs-Token
4. Marco schickt Stefan den Token (per E-Mail/WhatsApp/persönlich)
5. Stefan gibt den Token in seiner Instanz ein
6. Beide Systeme tauschen Zertifikate aus (automatisch, mTLS)
7. Verbindung hergestellt. Stefan konfiguriert: welche Equipment-Kategorien geteilt werden, welche Preise Marco sieht
8. Marco sieht SoundPro als "verbundener Partner" in seiner Instanz

**Was der Admin NICHT manuell machen will:**
- Zertifikate generieren und austauschen
- Firewall-Regeln anpassen
- API-Keys manuell konfigurieren
- JSON/YAML-Dateien für die Federation editieren

### Szenario 6: Etwas geht schief

Montag morgen, 08:00. Lisa ruft an: "Die App lädt nicht." Marco setzt sich an den Laptop.

**Was er braucht:**
1. Admin-Panel → System-Status: Grün/Gelb/Rot für jede Komponente (Backend, DB, Frontend)
2. Wenn rot: Fehlermeldung in Klartext, nicht "Error 500 — Internal Server Error"
3. Logs einsehbar im Admin-Panel (die letzten 100 Zeilen, filterbar nach Schwere)
4. Quick-Actions: "Backend neustarten", "Cache leeren", "DB-Verbindung testen"
5. Wenn nichts hilft: SSH und `docker compose logs` — aber das sollte die Ausnahme sein, nicht die Regel

---

## Ziele & Motivation

### Was den Admin antreibt
- **"Set and forget":** Einmal aufsetzen, dann soll es laufen. Keine wöchentliche Wartung, keine täglichen Checks.
- **Sicherheit:** Die Daten der Firma (Equipment, Kunden, Rechnungen) liegen auf dem Server. Wenn der gehackt wird, ist das eine Katastrophe. SSL, sichere Passwörter, regelmäßige Backups — das muss die Software unterstützen, nicht der Admin manuell einrichten.
- **Kontrolle:** Der Admin will wissen ob alles läuft. Nicht jede Minute — aber ein Dashboard das ihm auf einen Blick zeigt: "Alles grün" oder "Achtung, Disk ist zu 90% voll."
- **Dokumentation:** Wenn Marcos Kumpel Jens das Setup macht, muss Jens eine Doku lesen können die in 10 Minuten alles erklärt. Nicht ein 50-seitiges Handbuch.

### Konkrete Wünsche
1. **`docker compose up -d` → läuft** (inklusive DB, Backend, Frontend, Reverse-Proxy, SSL)
2. **Admin-Panel im Browser** für Benutzer, Rollen, Federation, Backups, System-Health
3. **Auto-Updates über UI** mit Changelog, Rollback, null Downtime (idealerweise)
4. **Automatische Backups** (täglich, konfigurierbar, Retention Policy)
5. **Monitoring-Dashboard:** CPU, RAM, Disk, DB-Größe, aktive User, letzte Fehler
6. **Gute Doku:** README mit Copy-Paste-Befehlen, FAQ, Troubleshooting-Guide
7. **Health-Check-Endpoint:** `/api/health` der von externem Monitoring (Uptime Robot, etc.) abgefragt werden kann

---

## Frustrationen — was den Admin wahnsinnig macht

### "20 Config-Dateien und nichts funktioniert"
Jedes Mal wenn Marco eine neue Software ausprobiert: `.env`, `config.yml`, `docker-compose.yml`, `nginx.conf`, `postgres.conf`... und jede Datei hat 50 Optionen von denen er 3 braucht. Dann eine vergessene Option → Error beim Start → 2 Stunden Debugging. Er will EINE `.env`-Datei mit den 5 Werten die wirklich nötig sind, und alles andere hat vernünftige Defaults.

### "Das Update hat alles kaputt gemacht"
Marcos größte Angst: Er klickt auf "Update" und danach funktioniert nichts mehr. Kein Rollback, keine Fehlermeldung, nur ein weißer Screen. Das ist ihm bei einer anderen Software passiert. Seitdem macht er vor jedem Update manuell ein DB-Dump — was 20 Minuten dauert und die er oft vergisst.

### "Der Server ist down und keiner merkt es"
Lisa arbeitet im Lager, scannt Equipment — und plötzlich lädt die App nicht mehr. Sie ruft Marco an. Marco ist auf einer Baustelle. Er kann nicht SSH-en weil sein Handy kein Terminal hat (bzw. er hat keine SSH-App installiert). Die Software ist 4 Stunden down bis Marco abends zu Hause ist. Das darf nicht passieren.

### "Wie war nochmal das Admin-Passwort?"
Marco hat die Software vor 6 Monaten aufgesetzt. Seitdem hat er sich nie ins Admin-Panel eingeloggt (weil alles lief). Jetzt muss er einen neuen Benutzer anlegen — und hat das Passwort vergessen. Passwort-Reset per E-Mail? Geht nicht, weil er den E-Mail-Server nie konfiguriert hat. Also: SSH, Datenbank, manueller Reset. Das sollte einfacher gehen.

---

## Technische Erwartungen

### Server-Anforderungen (minimal)
- **OS:** Ubuntu 22.04+ / Debian 12+ (oder jedes Linux mit Docker)
- **RAM:** 2GB minimum, 4GB empfohlen
- **Disk:** 20GB minimum (abhängig von Fotos/Dokumenten)
- **CPU:** 1 vCPU minimum, 2 empfohlen
- **Kosten:** Hetzner CX21 für 5€/Monat reicht für eine kleine Firma

### Docker Compose Stack
```yaml
# Erwartete Einfachheit:
services:
  app:        # Backend + Frontend (single container oder 2)
  db:         # PostgreSQL
  # Optional:
  backup:     # Automatisches Backup
```

Kein Redis, kein RabbitMQ, kein Elasticsearch, kein separater Worker — es sei denn absolut notwendig. Jeder zusätzliche Container erhöht die Komplexität und das Ausfallrisiko.

### SSL/TLS
- Automatisch via Let's Encrypt (eingebaut oder via Caddy/Traefik)
- Einzige Konfiguration: Domain-Name in `.env`
- Kein manuelles Zertifikat-Management

### Backup-Strategie
- Automatisches DB-Backup (pg_dump), täglich um 3 Uhr nachts
- Konfigurierbare Retention (7 Tage, 4 Wochen, 3 Monate)
- Backup-Status im Admin-Panel sichtbar ("Letztes Backup: heute 03:00, Größe: 45MB")
- Optional: Backup auf externen Storage (S3, SFTP)

---

## Software-Nutzung

| Aspekt | Details |
|--------|---------|
| **Gerät** | Laptop/PC mit Terminal + Browser |
| **Häufigkeit** | Einmal aufsetzen, dann alle 1-2 Monate (Update, neuer User) |
| **Session-Länge** | 30 Minuten bei Installation, 5-10 Minuten bei Updates/User-Verwaltung |
| **Kernfeatures** | Admin-Panel, Update-Manager, Backup-Status, Benutzer-Verwaltung, System-Logs, Federation-Setup |
| **Rolle im System** | System-Admin (voller Zugriff auf alles) |

---

## Zitate

- *"Docker Compose, .env-Datei mit 5 Variablen, `docker compose up -d`, Browser auf, Login. Wenn das so funktioniert, bin ich glücklich."*
- *"Das letzte Mal als ich eine Software geupdated habe, hat es die Datenbank zerstört. Seitdem mache ich vor jedem Update ein manuelles Backup. Automatisches Backup + Rollback = ich schlafe besser."*
- *"Ich bin kein Sysadmin. Ich bin Tontechniker der zufällig Docker kennt. Die Software muss für MICH gebaut sein, nicht für jemanden mit einem CCNA."*
- *"Wenn ich einen neuen Freelancer anlegen will, sollte das 2 Minuten dauern. Nicht 20."*
- *"Health-Check-Endpoint, den ich in Uptime Robot eintrage. Wenn der rot wird, krieg ICH die E-Mail — nicht Lisa die im Lager steht."*
