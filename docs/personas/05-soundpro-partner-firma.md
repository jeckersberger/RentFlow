# SoundPro GmbH (Stefan Keller) — Partner-Firma

> "Wir kennen uns seit 15 Jahren. Wenn Marco anruft und K2 braucht, sag ich ja. Aber manchmal vergessen wir beide die Rechnung."

---

## Steckbrief

| Feld | Details |
|------|---------|
| **Firmenname** | SoundPro GmbH |
| **Inhaber** | Stefan Keller, 42 |
| **Firmengröße** | 5 Mitarbeiter (3 fest, 2 Freelancer) |
| **Spezialisierung** | Beschallung: PA-Systeme, Monitoring, Recording, Streaming |
| **Lager** | 250m² in einem Gewerbegebiet, 15 Minuten von Marcos Lager entfernt |
| **Equipment-Schwerpunkt** | L-Acoustics (K2, A15, SB28), Yamaha (CL/QL-Serie), Shure (Wireless) |
| **Umsatz** | Ca. 600.000€/Jahr |
| **Beziehung zu Marco** | Langjährige Partnerschaft, gegenseitiger Sub-Rental, gelegentlich Co-Produktion |

---

## Hintergrund & Geschichte

Stefan und Marco kennen sich aus der Freelancer-Zeit — vor 15 Jahren haben sie zusammen auf einem Festival gearbeitet. Stefan am FOH, Marco als Stagehand. Über die Jahre haben sie sich immer wieder auf Baustellen getroffen, und als beide ihre eigenen Firmen gegründet haben, war die Zusammenarbeit eine natürliche Entwicklung.

SoundPro ist auf Ton spezialisiert. Stefan hat sich bewusst gegen Fullservice entschieden — lieber in einem Bereich richtig gut sein als in drei Bereichen mittelmäßig. Sein Equipment-Lager ist kleiner als Marcos, aber im Bereich PA/Beschallung deutlich besser bestückt: L-Acoustics K2 und A15 Arrays, Yamaha CL5 und QL1 Mischpulte, Shure Axient Digital Funkmikrofone, Lake Processing, und alles was man für eine professionelle Beschallung braucht.

Die Zusammenarbeit funktioniert in beide Richtungen:
- **Marco braucht PA:** Sein eigenes d&b E-System reicht für Firmenkunden. Aber wenn ein Festival oder Open-Air ansteht das größeres PA braucht, leiht er bei SoundPro — meistens K2 oder A15.
- **Stefan braucht Licht:** SoundPro hat kein Licht-Equipment. Wenn ein Kunde "Ton + Licht" will, leiht Stefan bei Marco — Moving Heads, PAR, Traversen.
- **Co-Produktion:** Bei großen Festivals stellen beide Firmen Equipment UND Crew. Marco macht Licht + Video, SoundPro macht Ton. Abrechnung: jeder rechnet seinen Teil separat mit dem Kunden ab, oder einer rechnet alles und der andere stellt eine Sub-Rechnung.

Das alles funktioniert — menschlich. Technisch ist es ein Desaster.

---

## Die aktuelle Zusammenarbeit — ein typisches Szenario

### Dienstag, 15 Uhr
Marco plant das Festival am Wochenende. Sein eigenes PA (d&b E12) reicht nicht für die Main Stage (3.000 Leute). Er braucht 4x L-Acoustics K2 Tops + 2x SB28 Subs von SoundPro.

**Was heute passiert:**
1. Marco schickt Stefan eine WhatsApp: "Servus Stefan, brauch 4x K2 + 2x SB28, Sa-So. Hast du die frei?"
2. Stefan liest die Nachricht 2 Stunden später (er war auf einer Baustelle). Checkt seinen Google Calendar — hat er an dem Wochenende einen Job der das Equipment braucht? Er glaubt nicht. Checkt sicherheitshalber seine Excel-Tabelle. Die ist vom letzten Monat. Er ruft seinen Lagerist an: "Peter, stehen die K2 im Lager?" Peter schaut nach: "Ja, 6 von 8 sind da. 2 sind bei der Hochzeit am Samstag." Stefan: "Ok, 4 sind frei."
3. Stefan schreibt Marco zurück: "Geht klar. 4x K2 + 2x SB28, Sa-So. Wann holst du ab?"
4. Marco: "Freitag 16 Uhr?" Stefan: "Passt."
5. Freitag 16 Uhr: Marco (oder Lisa) fährt zu SoundPro, lädt die Cases. Stefan oder Peter sind da, machen einen kurzen Zustand-Check: "Die K2 sind alle top, der eine Sub hat einen kleinen Kratzer am Grill, war schon." Händischer Lieferschein, zwei Unterschriften.
6. Festival Sa-So: Equipment im Einsatz. Alles läuft gut.
7. Sonntag Nacht: Marco bringt die Cases zurück. Stefan ist nicht da (2 Uhr nachts), also stellt Marco die Cases vor die Lagertür und schickt eine WhatsApp: "K2 + Subs stehen vor deiner Tür."
8. Montag: Stefan/Peter checken die Rückgabe. Alles da, alles heile.
9. Stefan schickt Marco eine Rechnung... oder auch nicht. Manchmal vergisst Stefan es (er hat auch genug zu tun). Manchmal kommt die Rechnung 3 Wochen später. Und manchmal stimmt der Preis nicht ("Ich dachte wir hatten 400€/Tag ausgemacht?" — "Nee, 480€, wie letztes Mal.").

**Was das kostet:**
- 4 WhatsApp-Nachrichten, 1 Anruf, 1 interne Nachfrage — für eine einfache Verfügbarkeitsanfrage
- 30 Minuten Stefans Zeit für die Koordination
- Keine digitale Zustandsdokumentation (wenn ein K2 beschädigt zurückkommt, steht Aussage gegen Aussage)
- Vergessene oder verspätete Rechnungen
- Unklare Preise (war es 400 oder 480 pro Tag?)

---

## Was mit Federation passieren würde

### Das gleiche Szenario — mit Software

1. Marco öffnet die Software, Modul "Sub-Rental". Er sieht seine verknüpften Partner: SoundPro GmbH (Status: aktiv, verbunden). Er klickt auf "Verfügbarkeit prüfen".
2. Das System fragt SoundPros Instanz via Federation-API ab: "L-Acoustics K2, Samstag-Sonntag, Menge: 4". SoundPros System antwortet automatisch: "4x K2 verfügbar. Preis: 480€/Tag/Stück (Partner-Kondition). 2x SB28 verfügbar. Preis: 320€/Tag/Stück."
3. Marco sieht die Verfügbarkeit sofort (kein Warten auf WhatsApp-Antwort), erstellt eine Sub-Rental-Anfrage mit einem Klick.
4. Stefan bekommt in seiner Instanz eine Benachrichtigung: "Neue Sub-Rental-Anfrage von MB Veranstaltungstechnik: 4x K2 + 2x SB28, Sa-So." Er prüft, bestätigt mit einem Klick.
5. Beide Systeme aktualisieren den Status: Equipment ist als "Sub-Rental" markiert, in Marcos System als "incoming", in Stefans System als "outgoing". Doppelbuchungen ausgeschlossen.
6. Lisa holt das Equipment ab, scannt bei SoundPro die QR-Codes. Zustand wird in beiden Systemen dokumentiert (Fotos, Zustandsnotizen). Digitale Unterschrift statt Papier.
7. Nach dem Event: Lisa bringt das Equipment zurück, scannt es zurück. SoundPros System bestätigt: alles da, alles heile.
8. Automatisch: SoundPros System generiert eine Ausgangsrechnung an Marco. Marcos System zeigt sie als Eingangsrechnung für Thomas. Betrag, Positionen, Zeitraum — alles korrekt, weil es auf den bestätigten Sub-Rental-Daten basiert.

**Zeitersparnis:** 4 WhatsApp-Nachrichten + 1 Anruf + 30 Minuten Koordination → 3 Klicks + 2 Minuten.

---

## Was SoundPro über Federation teilt

### Freigegebene Informationen (Stefan konfiguriert das selbst)

| Information | Geteilt? | Details |
|-------------|----------|---------|
| **Equipment-Kategorien** | Ja, selektiv | Nur PA-Systeme und Subs. Nicht: Mischpulte, Mikrofone, Processing (die verleiht Stefan ungern) |
| **Verfügbarkeit** | Ja | Zeitraum-basiert: "4x K2 frei von Sa-So" |
| **Partner-Mietpreise** | Ja | Spezielle Konditionen für Marco (15% unter Listenpreis) |
| **Equipment-Zustand** | Ja | Bei Übergabe: Fotos, Notizen |
| **Sub-Rental-Status** | Ja | Angefragt → Bestätigt → Ausgegeben → Zurückgegeben |

### NICHT freigegebene Informationen

| Information | Warum nicht |
|-------------|------------|
| **Endkunden-Preise** | Das sind Stefans Kalkulationen, geht Marco nichts an |
| **Kundenliste** | Wettbewerbsrelevant, vertraulich |
| **Lagerwert / Gesamtbestand** | Geschäftsgeheimnis |
| **Andere Sub-Rental-Partner** | Stefan arbeitet auch mit anderen Firmen — Marco muss das nicht wissen |
| **Finanzen, Umsatz, Margen** | Offensichtlich vertraulich |
| **Mitarbeiter, Gehälter** | Datenschutz |
| **Mischpulte & Mikrofone** | Stefan hat sich entschieden diese Kategorien nicht über Federation zu teilen — er verleiht sie nur auf persönliche Anfrage |

### Federation-Konfiguration bei SoundPro

Stefan hat in seiner Instanz konfiguriert:
- **Partner "MB Veranstaltungstechnik":** Verknüpft, aktiv
- **Freigegebene Kategorien:** PA-Tops, PA-Subs (alles andere: nicht freigegeben)
- **Preismodell:** Partner-Kondition (-15% vom Listenpreis)
- **Bestätigungsmodus:** Manuell (Stefan will jede Anfrage selbst bestätigen, keine automatische Zusage)
- **Zeitfenster:** Verfügbarkeit nur 30 Tage im Voraus sichtbar (nicht langfristiger)

---

## Typische Interaktionsszenarien

### Szenario 1: Standard Sub-Rental
Marco braucht PA → schaut in der Software → SoundPro hat's → Anfrage → Stefan bestätigt → Lisa holt ab → Event → Lisa bringt zurück → Rechnung automatisch.

### Szenario 2: Engpass bei beiden
Marco UND Stefan haben am gleichen Wochenende große Jobs. Keiner hat genug K2. Aber Stefan kennt noch "AudioRent Süd" die auch eine Instanz betreiben. Stefan hat auch mit AudioRent eine Federation-Verbindung. Das System könnte Marco zeigen: "SoundPro hat keine K2 frei, aber AudioRent Süd hat 4 Stück." (Nur wenn AudioRent das freigibt und Stefan die Weitervermittlung erlaubt.)

### Szenario 3: Gegenseitiges Sub-Rental
Stefan braucht Licht für ein Firmen-Event. Er schaut in seiner Software: MB Veranstaltungstechnik hat 8x Claypaky B-Eye frei. Anfrage an Marco. Marco bestätigt. Stefans Leute holen die B-Eyes bei Marco ab. Nach dem Event: Rückgabe, automatische Rechnung von Marco an Stefan.

### Szenario 4: Co-Produktion Festival
Großes Festival: Ton (SoundPro) + Licht (Marco). Beide liefern Equipment und Crew. In der Software: ein gemeinsames Projekt, aber jede Firma sieht nur ihre eigenen Positionen (und Preise). Am Ende: jeder rechnet seinen Teil ab. Sub-Rental-Positionen werden automatisch verrechnet.

### Szenario 5: Streitfall — beschädigtes Equipment
Marco gibt eine K2 zurück die einen Kratzer hat. Stefans System zeigt: "Bei Ausgabe: kein Kratzer (Foto von Freitag 16:05). Bei Rückgabe: Kratzer rechts unten (Foto von Montag 09:30)." Klare Beweislage. Kein Streit, keine Diskussion. Marco übernimmt die Reparaturkosten.

---

## Was SoundPro von der Software erwartet

### Must-Haves
1. **Verfügbarkeitsanfragen ohne Telefon/WhatsApp** — Einfach in der Software nachschauen
2. **Klare Partner-Konditionen** — Preise einmal konfigurieren, nicht jedes Mal neu verhandeln
3. **Automatische Rechnungen** — Sub-Rental abgeschlossen → Rechnung wird generiert
4. **Zustandsdokumentation** — Fotos bei Aus- und Rückgabe, digital archiviert
5. **Kontrolle über geteilte Daten** — Stefan entscheidet was Marco sehen darf, nicht die Software

### Nice-to-Haves
- Digitaler Rahmenvertrag/AGB zwischen Partnern (einmalig hinterlegen)
- Statistiken: "Wie viel haben wir dieses Jahr an/von Marco vermietet?"
- Automatische Versicherungsbestätigung bei Sub-Rental
- Gemeinsame Equipment-Kategorie-Standards (damit "K2" bei beiden das Gleiche bedeutet)

### Technische Anforderungen
- **Eigene Instanz** auf eigenem Server (SoundPro hostet selbst, separate DB)
- **Federation-API via mTLS** — sichere, verschlüsselte Verbindung
- **Kein zentraler Server** — Peer-to-Peer, keine Abhängigkeit von einem Dritten
- **Offline-tolerant** — wenn Marcos Server down ist, funktioniert Stefans System trotzdem normal
- **Datensouveränität** — Stefans Daten bleiben auf Stefans Server. Marcos System fragt nur ab, speichert nicht dauerhaft.

---

## Zitate (Stefan Keller)

- *"Marco und ich, das läuft seit Jahren. Aber letztes Jahr haben wir 3 Rechnungen vergessen. Beide Seiten. Das ist unprofessionell."*
- *"Ich will wissen wann Marcos Team meine K2 abholt, wie die zurückkommen, und dass die Rechnung automatisch rausgeht. Ist das zu viel verlangt?"*
- *"Bitte keine zentrale Plattform wo wir uns beide anmelden müssen. Ich will MEINEN Server mit MEINEN Daten. Und Marco hat seinen. Und die reden miteinander — das ist Federation."*
- *"Die Preise die Marco sieht sind nicht die Preise die meine anderen Kunden sehen. Partner-Konditionen. Das muss die Software verstehen."*
- *"Wenn ich 'nicht freigeben' sage, dann darf Marcos System das auch nicht sehen. Punkt. Keine Hintertür, keine Ausnahme."*
