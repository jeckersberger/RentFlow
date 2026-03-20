# Personas — Lagerverwaltung Veranstaltungstechnik

> **Status:** Entwurf — wird noch vom Auftraggeber gegengelesen und verfeinert.

---

## Persona 1: Marco Berger, 38 — Geschäftsführer & Projektleiter

> "Ich muss jederzeit wissen, was wo ist, was rausgeht und was reinkommt — am besten auf einen Blick."

### Hintergrund
Marco hat vor 8 Jahren seine eigene VT-Firma gegründet, nachdem er 10 Jahre als Freelance-Tontechniker unterwegs war. Angefangen mit einem Sprinter und einem PA-System, hat er sich mittlerweile ein solides Equipment-Lager aufgebaut. Seine Firma beschäftigt 3 Festangestellte und arbeitet regelmäßig mit 5-8 Freelancern. Fullservice: Licht, Ton, Video und Rigging — Schwerpunkt Firmenevents und Festivals.

### Typischer Tag
- **07:30** — Kaffee, Laptop auf: Dashboard checken. Was geht heute raus? Gibt es Engpässe?
- **08:00** — Kurz ins Lager: LKW-Beladung kontrollieren, mit Lisa Packlisten abstimmen
- **09:00** — Büro: Angebote schreiben, Anfragen beantworten, Projekte kalkulieren
- **11:00** — Telefonate mit Kunden: Änderungswünsche, Technik-Beratung
- **13:00** — Unterwegs zum Kunden: Besichtigung Location, technische Planung
- **15:00** — Per Handy: Sub-Rental bei SoundPro anfragen (4x K2 für Samstag)
- **17:00** — Rechnungen freigeben, offene Posten mit Thomas besprechen
- **19:00** — Nochmal Dashboard: Morgen ist Festival-Aufbau, alles ready?

### Ziele & Motivation
- Eine einzige Software statt 5 verschiedene Tools (Excel, WhatsApp, Papier, E-Mail, separates Rechnungsprogramm)
- Kompletter Fluss: Angebot → Planung → Packliste → Lieferschein → Rechnung
- Jederzeit wissen: Welches Equipment ist wo? Was ist frei? Was ist defekt?
- Partner-Firmen unkompliziert Equipment ausleihen/verleihen
- Sein Team soll selbstständig arbeiten können ohne ständig nachfragen zu müssen

### Frustrationen
- **Doppelbuchungen** — Equipment wird für 2 Jobs gleichzeitig eingeplant, fällt erst auf wenn es zu spät ist
- **Rechnungen vergessen** — Job ist vorbei, Equipment zurück, aber keiner hat die Rechnung geschrieben
- **Sub-Rental-Chaos** — "Hab ich die K2 von SoundPro eigentlich zurückgegeben?" → WhatsApp durchsuchen
- **Keine Verbindung Lager↔Buchhaltung** — Thomas tippt Rechnungen aus Angeboten ab, Fehler passieren
- **Kein Überblick** — Muss Lisa anrufen um zu wissen ob das Mischpult im Lager oder auf einer Baustelle ist

### Software-Nutzung
- **Geräte:** Laptop (70%), Handy (25%), Tablet auf Baustelle (5%)
- **Häufigkeit:** Täglich, mehrmals am Tag
- **Kernfeatures:** Dashboard, Projektplanung, Angebote, Rechnungsfreigabe, Federation/Sub-Rental
- **Rolle im System:** Admin + Projektleiter

### No-Gos
- Zu kompliziert — muss ohne Schulung bedienbar sein
- Zu langsam — Dashboard muss in <2 Sekunden laden
- Cloud-Zwang — will eigenen Server, eigene Daten
- Vendor Lock-in — keine Abhängigkeit von Anbieter der plötzlich Preise erhöht

---

## Persona 2: Lisa Huber, 27 — Lagerverwalterin & Technikerin

> "Wenn ich im Lager stehe, muss ich in 2 Sekunden wissen ob etwas da ist, und in 5 Sekunden ob es funktioniert."

### Hintergrund
Lisa ist seit 3 Jahren bei Marcos Firma. Gelernte Veranstaltungstechnikerin, die sich schnell als Lager-Chefin etabliert hat. Sie kennt jedes Gerät im Lager, weiß wo es steht und ob es funktioniert. Neben der Lagerverwaltung fährt sie auch regelmäßig auf Baustellen zum Auf-/Abbau.

### Typischer Tag
- **07:00** — Lager öffnen. Scanner einschalten. Packlisten für heute auf dem Handheld checken.
- **07:15** — Kommissionierung Job 1: Packliste abarbeiten, jedes Teil scannen, auf Europalette
- **08:30** — LKW beladen, Lieferschein digital unterschreiben lassen (Fahrer)
- **09:00** — Kommissionierung Job 2: Festival morgen, große Packliste (120 Positionen)
- **11:00** — Rückgabe vom Wochenend-Job: Equipment vom LKW scannen, Zustand prüfen
- **11:30** — 2 Kabel defekt → Defekt melden mit Foto + Notiz, Equipment auf "Reparatur" setzen
- **12:00** — Mittagspause
- **13:00** — Equipment einlagern: scannen, Lagerplatz zuweisen
- **14:00** — Inventur-Check für Regal B3 (Stecker/Adapter)
- **15:00** — Sub-Rental von SoundPro angekommen: 4x K2 einchecken, Zustand dokumentieren
- **16:00** — Packlisten für morgen vorbereiten, Fehlteile an Marco melden

### Ziele & Motivation
- Schnell arbeiten: Scannen → fertig. Keine 10 Klicks für eine einfache Aktion
- Fehler vermeiden: Nie wieder falsches Equipment einpacken
- Defekte sofort dokumentieren: Nicht erst 3 Tage später wenn keiner mehr weiß wer's war
- Wissen wo alles steht: "Wo ist der Adapter-Koffer?" → 2 Sekunden Antwort

### Frustrationen
- **Papier-Packlisten** — Fehler, Sachen vergessen, kann nichts abhaken was andere sehen
- **Scanner-Software langsam** — 10 Klicks für "Equipment rein"
- **Kein Defekt-Tracking** — War das Gerät defekt als es rausging? Keine Ahnung
- **Lagerplatz-Suche** — "Wo hab ich die 16er DMX-Kabel hingestellt?" → 15 Minuten suchen
- **Kein Echtzeit-Status** — Marco fragt "Ist das Mischpult da?" → muss physisch nachschauen

### Software-Nutzung
- **Geräte:** Zebra Android-Handheld (70%), Handy auf Baustelle (25%), PC selten (5%)
- **Häufigkeit:** Ganzen Tag, durchgehend
- **Kernfeatures:** Scan In/Out, Packlisten, Defekt-Meldung, Lagerplatz-Suche, Rückgabe-Prüfung
- **Rolle im System:** Lager-Mitarbeiterin (kein Zugang zu Preisen/Rechnungen)

### UX-Anforderungen
- Große Buttons, Scanner-optimiert
- Maximal 2 Taps für Standardaktionen
- Sofortiges Feedback nach Scan (grün = OK, rot = Problem)
- Offline-fähig (Lager hat manchmal schlechtes WLAN)

---

## Persona 3: Thomas Berger, 52 — Büro & Buchhaltung

> "Ich brauche saubere Zahlen, pünktliche Rechnungen und einen Export den mein Steuerberater versteht."

### Hintergrund
Thomas ist Marcos Schwager und macht seit der Firmengründung die Buchhaltung — Teilzeit, 3 Tage die Woche. Eigentlich gelernter Industriekaufmann, hat sich in die Veranstaltungsbranche eingearbeitet. Versteht das Geschäft, ist aber kein Techniker. Kämpft täglich mit der Lücke zwischen "was wurde geliefert" und "was wurde berechnet".

### Typischer Tag
- **08:00** — PC hochfahren, Kaffee, Mails checken
- **08:30** — Offene Posten durchgehen: Wer hat noch nicht bezahlt? 3 Rechnungen überfällig
- **09:00** — Zahlungserinnerungen schicken (aktuell per Hand: Word-Vorlage, PDF, E-Mail)
- **10:00** — Neue Rechnungen erstellen: Festival vom Wochenende → Marco sagt "war alles wie im Angebot, kannst Rechnung schreiben"
- **10:30** — Problem: Im Angebot standen 8 Moving Heads, geliefert wurden 10 → Rechnung anpassen
- **11:00** — Eingangsrechnungen: SoundPro hat Sub-Rental-Rechnung geschickt → verbuchen
- **12:00** — Teilrechnung für Großprojekt erstellen (Anzahlung 30%)
- **14:00** — DATEV-Export vorbereiten für Steuerberater → aktuell: alles händisch in Excel
- **15:00** — Monatsauswertung: Umsatz, offene Posten, MwSt-Zusammenfassung

### Ziele & Motivation
- Rechnung aus Projekt/Lieferschein auf Knopfdruck generieren
- Keine Tippfehler mehr durch manuelles Abtippen
- Automatische Zahlungserinnerungen
- DATEV-Export der einfach funktioniert
- Klare Übersicht: Was ist offen, was ist bezahlt, was ist überfällig

### Frustrationen
- **Manuelles Abtippen** — Angebot → Rechnung: Positionen abtippen, Fehler passieren
- **Vergessene Rechnungen** — Job ist 2 Wochen her, Rechnung wurde nie geschrieben
- **Teilrechnungen** — 30% Anzahlung, 50% nach Aufbau, 20% nach Abbau → Hölle in Excel
- **DATEV-Export** — Steuerberater braucht bestimmtes Format, aktuell alles händisch
- **Keine Verbindung** — "Was wurde geliefert?" muss er Lisa/Marco fragen
- **Gutschriften/Stornos** — Kompliziert, fehleranfällig, keine Vorlage

### Software-Nutzung
- **Geräte:** Desktop-PC (95%), selten Laptop (5%)
- **Häufigkeit:** 3 Tage/Woche, jeweils 6-7 Stunden
- **Kernfeatures:** Rechnungserstellung, Mahnwesen, DATEV-Export, offene Posten, Berichte
- **Rolle im System:** Buchhaltung (kein Zugang zu Lager-Operationen)

### UX-Anforderungen
- Große Schrift, klare Tabellen
- Tastatur-Navigation (Tab durch Felder)
- Vertraute Konzepte (ähnlich Excel-Logik)
- Wenig Überraschungen — "Was passiert wenn ich hier klicke?"

---

## Persona 4: Kevin Radic, 24 — Freelance-Techniker

> "Sag mir einfach wann, wo, was — und zeig mir die Packliste. Mehr brauch ich nicht."

### Hintergrund
Kevin ist seit 2 Jahren als Freelancer in der Veranstaltungstechnik unterwegs. Arbeitet für 3-4 verschiedene Firmen, darunter Marcos. Digital Native — kommt mit jeder App klar, wenn sie logisch aufgebaut ist. Ist meistens auf der Baustelle, selten im Lager. Hat kein eigenes Equipment.

### Typische Woche
- **Montag** — Frei (Wochenend-Job war bis Sonntag 3 Uhr)
- **Dienstag** — 07:00 bei Marco im Lager: Festival-Aufbau. Packliste auf dem Handy checken, Equipment abholen, scannen, LKW beladen
- **Mittwoch** — Ganzer Tag auf der Baustelle: Aufbau PA, Licht programmieren, Soundcheck
- **Donnerstag** — Show-Tag: Betreuung, dann Abbau bis 2 Uhr nachts. Rückgabe-Zustand dokumentieren
- **Freitag** — Anderer Auftraggeber: Firmen-Gala, kleine PA + Licht
- **Samstag** — Marco: Corporate Event, Tagesveranstaltung

### Ziele & Motivation
- Alle Job-Infos an einem Ort: Wann, wo, was mitnehmen, wer ist Ansprechpartner
- Packliste interaktiv abhaken (nicht PDF)
- Wissen welches Equipment er mitnehmen soll BEVOR er im Lager steht
- Arbeitszeiten einfach erfassen

### Frustrationen
- **WhatsApp-Chaos** — Job-Infos kommen als Nachricht, gehen unter
- **PDF-Packliste** — Nicht interaktiv, kann nichts abhaken
- **Keine Übersicht** — "Welche Firma hat mich nächste Woche gebucht?" → 3 verschiedene Kalender
- **Equipement-Unklarheit** — Steht im Lager und weiß nicht welche Kisten er braucht

### Software-Nutzung
- **Geräte:** iPhone (100%)
- **Häufigkeit:** Täglich, kurze Sessions (5-10 Minuten)
- **Kernfeatures:** Job-Übersicht, Packliste, Equipment-Scan, Zustand-Dokumentation, Zeiterfassung
- **Rolle im System:** Freelancer (stark eingeschränkt)

### Berechtigungen — darf sehen
- Seine zugewiesenen Jobs (Datum, Ort, Zeitplan)
- Packlisten für seine Jobs
- Ansprechpartner + Kontaktdaten (nur für den jeweiligen Job)
- Equipment-Details (Name, Foto, Kategorie)

### Berechtigungen — darf NICHT sehen
- Preise (Miet-, Verkaufs-, Einkaufspreise)
- Rechnungen
- Kundendaten (Firmennamen, Adressen, Umsätze)
- Andere Jobs (die nicht seine sind)
- Lagerbestände / Lagerwert
- Firmen-Finanzen

---

## Persona 5: SoundPro GmbH — Partner-Firma

> "Wir leihen uns gegenseitig Equipment aus wenn einer von uns ausgebucht ist. Das muss schnell und unkompliziert gehen."

### Hintergrund
SoundPro ist eine 5-Mann-Firma, spezialisiert auf Beschallung (PA-Systeme, Monitoring, Recording). Inhaber Stefan Keller kennt Marco seit 15 Jahren aus der Freelancer-Zeit. Die Firmen arbeiten regelmäßig zusammen: Sub-Rental (gegenseitig), Dry-Hire, gelegentlich Co-Produktion bei großen Festivals.

### Typische Zusammenarbeit
- **2-4x pro Monat**, in der Festivalsaison (Mai-September) deutlich mehr
- **Sub-Rental:** Marco braucht PA-Equipment das er nicht hat → leiht von SoundPro
- **Gegenseitig:** SoundPro braucht Licht-Equipment → leiht von Marco
- **Dry-Hire:** Nur Equipment ohne Techniker
- **Co-Produktion:** Beide Firmen stellen Equipment + Crew für großes Festival

### Typischer Ablauf (aktuell)
1. Marco ruft Stefan an: "Hast du 4x K2 frei am Samstag?"
2. Stefan schaut in seinem System, ruft zurück: "Ja, hab ich"
3. WhatsApp: "OK, brauche ich. Wann kann ich abholen?"
4. Equipment wird abgeholt, Lieferschein auf Papier
5. Nach dem Event: Rückgabe, Zustand prüfen
6. Stefan schickt Rechnung per Mail (2 Wochen später, manchmal vergessen)
7. Thomas verbucht die Rechnung

### Typischer Ablauf (mit Federation)
1. Marco öffnet die Software: "Sub-Rental anfragen"
2. Sieht SoundPros verfügbares PA-Equipment (gefiltert: nur freigegebene Kategorien)
3. Wählt 4x K2, Datum Samstag → sendet Anfrage
4. Stefan bekommt Benachrichtigung in seiner Instanz → bestätigt
5. Equipment wird in beiden Systemen als "Sub-Rental" markiert
6. Lisa scannt die K2 bei Abholung ein → Status in beiden Systemen aktualisiert
7. Nach Rückgabe: Lisa scannt zurück → Rechnung wird automatisch generiert
8. Thomas sieht die Eingangsrechnung, Thomas' Gegenstück bei SoundPro sieht die Ausgangsrechnung

### Geteilte Infos via Federation
- Equipment-Verfügbarkeit (Kategorie, Menge, Zeitraum)
- Partner-Mietpreise (spezielle Konditionen)
- Abhol-/Lieferzeiten
- Equipment-Zustand
- Sub-Rental-Status (angefragt, bestätigt, ausgegeben, zurück)

### NICHT geteilte Infos
- Endkunden-Preise und Kalkulationen
- Kundenliste und Kundendaten
- Interne Kalkulation und Margen
- Mitarbeiter-Daten und Gehälter
- Finanzdaten und Umsätze
- Lagerwert und Gesamtbestand

### Technisch
- Eigene Instanz auf eigenem Server
- Verbindung via Federation-API (mTLS)
- Jede Firma kontrolliert: welche Equipment-Kategorien geteilt werden, welche Preise Partner sehen, ob automatische Bestätigung oder manuell

---

## Persona 6: Der Admin / IT-Verantwortliche

> "Ich will den Server einmal aufsetzen und dann soll es einfach laufen."

### Hintergrund
In kleinen VT-Firmen ist "der Admin" meistens der Geschäftsführer selbst (Marco), oder ein befreundeter IT-ler der "mal drüberschaut". Kann Docker, kennt Linux-Basics, hat schon mal einen Webserver aufgesetzt. Will sich aber nicht jede Woche mit der Software beschäftigen — es soll einfach laufen.

### Aufgaben
- **Einmalig:** Server aufsetzen (Docker Compose), Domain + SSL einrichten, erste Benutzer anlegen
- **Selten:** Updates einspielen, Federation-Verbindungen einrichten, neue Benutzer anlegen
- **Automatisch (hoffentlich):** Backups, SSL-Erneuerung, Health-Monitoring

### Ziele & Motivation
- `docker compose up -d` → Software läuft
- Admin-Panel im UI: Benutzer verwalten, Federation-Setup, Backup-Status, System-Health
- Updates per Klick im UI (nicht per SSH)
- Gute Dokumentation mit Copy-Paste-Befehlen

### Frustrationen
- 20 Konfigurationsdateien bevor irgendwas läuft
- Keine klare Installationsanleitung
- Updates die Dinge kaputt machen (ohne Rollback)
- Kein Monitoring — System ist down und keiner merkt es

### Software-Nutzung
- **Geräte:** PC/Laptop mit SSH und Browser
- **Häufigkeit:** Einmal aufsetzen, dann selten (Updates, User-Verwaltung)
- **Kernfeatures:** Admin-Panel, Update-Manager, Backup-Status, Logs
- **Rolle im System:** System-Admin (voller Zugriff)
