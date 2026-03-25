import { Link } from 'react-router-dom'
import './Legal.scss'

// TODO: Diese Platzhalter durch echte Tenant-Config-Daten ersetzen
const COMPANY = {
  name: '[Firmenname]',
  street: '[Strasse Nr.]',
  zip: '[PLZ]',
  city: '[Stadt]',
  country: 'Deutschland',
  ceo: '[Vorname Nachname]',
  email: '[email@example.com]',
  phone: '[+49 XXX XXXXXXX]',
}

function DatenschutzPage() {
  return (
    <div className="legal-page">
      <div className="legal-card">
        <div className="legal-card__header">
          <Link to="/login" className="legal-card__back">Zurueck zur Anmeldung</Link>
          <h1 className="legal-card__title">Datenschutzerklaerung</h1>
        </div>

        <div className="legal-card__content">
          {/* 1. Verantwortlicher */}
          <section className="legal-section">
            <h2>1. Verantwortlicher</h2>
            <p>
              {COMPANY.name}<br />
              {COMPANY.street}<br />
              {COMPANY.zip} {COMPANY.city}<br />
              {COMPANY.country}
            </p>
            <p>
              Telefon: {COMPANY.phone}<br />
              E-Mail: {COMPANY.email}
            </p>
          </section>

          {/* 2. Uebersicht */}
          <section className="legal-section">
            <h2>2. Uebersicht der Verarbeitungen</h2>
            <p>
              Die nachfolgende Uebersicht fasst die Arten der verarbeiteten
              Daten und die Zwecke ihrer Verarbeitung zusammen und verweist auf
              die betroffenen Personen.
            </p>
            <h3>Arten der verarbeiteten Daten</h3>
            <ul>
              <li>Bestandsdaten (z.B. Namen, Adressen)</li>
              <li>Kontaktdaten (z.B. E-Mail, Telefonnummern)</li>
              <li>Inhaltsdaten (z.B. Rechnungen, Belege, Projektdaten)</li>
              <li>Vertragsdaten (z.B. Vertragsgegenstand, Laufzeit)</li>
              <li>Nutzungsdaten (z.B. besuchte Seiten, Zugriffszeiten)</li>
              <li>Meta-/Kommunikationsdaten (z.B. IP-Adressen, Geraeteinformationen)</li>
              <li>Beschaeftigtendaten (z.B. Personalstammdaten, Zeiterfassung)</li>
            </ul>
          </section>

          {/* 3. Rechtsgrundlagen */}
          <section className="legal-section">
            <h2>3. Rechtsgrundlagen</h2>
            <p>
              Die Verarbeitung personenbezogener Daten erfolgt auf Grundlage
              folgender Rechtsgrundlagen:
            </p>
            <ul>
              <li>
                <strong>Vertragserfullung (Art. 6 Abs. 1 lit. b DSGVO):</strong>{' '}
                Verarbeitung zur Erfuellung vertraglicher Pflichten oder
                vorvertraglicher Massnahmen.
              </li>
              <li>
                <strong>Berechtigte Interessen (Art. 6 Abs. 1 lit. f DSGVO):</strong>{' '}
                Verarbeitung zur Wahrung berechtigter Interessen des
                Verantwortlichen, z.B. Sicherheit der Anwendung.
              </li>
              <li>
                <strong>Gesetzliche Pflichten (Art. 6 Abs. 1 lit. c DSGVO):</strong>{' '}
                Verarbeitung zur Erfuellung gesetzlicher Aufbewahrungspflichten
                (z.B. Steuerrecht, Handelsrecht).
              </li>
            </ul>
          </section>

          {/* 4. Hosting */}
          <section className="legal-section">
            <h2>4. Hosting</h2>
            <p>
              Diese Anwendung wird auf Servern der{' '}
              <strong>Hetzner Online GmbH</strong> (Industriestr. 25, 91710
              Gunzenhausen, Deutschland) gehostet. Die Server befinden sich in
              Deutschland.
            </p>
            <p>
              Fuer die Auslieferung und den DDoS-Schutz nutzen wir{' '}
              <strong>Cloudflare, Inc.</strong> (101 Townsend St, San Francisco,
              CA 94107, USA) als Content Delivery Network (CDN). Cloudflare
              verarbeitet dabei kurzfristig Meta- und Kommunikationsdaten (z.B.
              IP-Adressen). Die Verarbeitung erfolgt auf Basis unserer
              berechtigten Interessen an einer sicheren und effizienten
              Bereitstellung. Es bestehen EU-Standardvertragsklauseln.
            </p>
          </section>

          {/* 5. Cookies */}
          <section className="legal-section">
            <h2>5. Cookies und lokale Speicherung</h2>
            <p>
              Diese Anwendung verwendet ausschliesslich technisch notwendige
              Cookies und lokale Speichermechanismen (localStorage):
            </p>
            <ul>
              <li>
                <strong>Session-/Authentifizierungs-Token:</strong> Fuer die
                sichere Anmeldung und Sitzungsverwaltung (JWT-basiert). Diese
                sind fuer den Betrieb der Anwendung zwingend erforderlich.
              </li>
              <li>
                <strong>Benutzereinstellungen:</strong> Speicherung von
                Theme-Praeferenzen (Hell-/Dunkelmodus) und UI-Einstellungen.
              </li>
            </ul>
            <p>
              Es werden keine Tracking-Cookies, Analyse-Cookies oder
              Marketing-Cookies eingesetzt. Eine Einwilligung ist daher nicht
              erforderlich (Art. 5 Abs. 3 ePrivacy-Richtlinie).
            </p>
          </section>

          {/* 6. Datenverarbeitung */}
          <section className="legal-section">
            <h2>6. Datenverarbeitung im Rahmen der Anwendung</h2>

            <h3>6.1 Benutzerverwaltung</h3>
            <p>
              Bei der Registrierung und Nutzung werden folgende Daten
              verarbeitet: Name, E-Mail-Adresse, Passwort (verschluesselt
              gespeichert mittels Argon2id), Rolle und Berechtigungen. Die
              Verarbeitung erfolgt zur Vertragserfullung.
            </p>

            <h3>6.2 Rechnungen und Belege</h3>
            <p>
              Im Rahmen der Rechnungsverwaltung werden Kundendaten (Name,
              Adresse, Kontaktdaten), Rechnungspositionen und Zahlungsdaten
              verarbeitet. Die Aufbewahrung erfolgt gemaess gesetzlicher
              Aufbewahrungsfristen (10 Jahre gemaess &sect; 147 AO, &sect; 257 HGB).
            </p>

            <h3>6.3 Personalverwaltung / Crew</h3>
            <p>
              Fuer die Personalplanung werden Stammdaten (Name, Kontaktdaten,
              Qualifikationen), Einsatzplanung und ggf. Zeiterfassungsdaten
              verarbeitet. Rechtsgrundlage ist die Vertragserfullung im
              Beschaeftigungsverhaeltnis (&sect; 26 BDSG).
            </p>

            <h3>6.4 Projektdaten</h3>
            <p>
              Projektbezogene Daten wie Kundenzuordnung, Equipment-Reservierungen,
              Transportplanung und Dokumentation werden zur Vertragserfuellung
              verarbeitet.
            </p>
          </section>

          {/* 7. SSL */}
          <section className="legal-section">
            <h2>7. SSL-/TLS-Verschluesselung</h2>
            <p>
              Diese Anwendung nutzt aus Sicherheitsgruenden und zum Schutz der
              Uebertragung vertraulicher Inhalte eine SSL-/TLS-Verschluesselung.
              Eine verschluesselte Verbindung erkennen Sie daran, dass die
              Adresszeile des Browsers von &quot;http://&quot; auf &quot;https://&quot; wechselt
              und an dem Schloss-Symbol in Ihrer Browserzeile.
            </p>
          </section>

          {/* 8. Rechte der Betroffenen */}
          <section className="legal-section">
            <h2>8. Rechte der betroffenen Personen</h2>
            <p>
              Ihnen stehen als betroffene Person folgende Rechte gemaess der
              DSGVO zu:
            </p>
            <ul>
              <li>
                <strong>Auskunftsrecht (Art. 15 DSGVO):</strong> Sie haben das
                Recht, Auskunft ueber Ihre bei uns gespeicherten
                personenbezogenen Daten zu erhalten.
              </li>
              <li>
                <strong>Berichtigungsrecht (Art. 16 DSGVO):</strong> Sie koennen
                die Berichtigung unrichtiger oder die Vervollstaendigung
                unvollstaendiger Daten verlangen.
              </li>
              <li>
                <strong>Loeschungsrecht (Art. 17 DSGVO):</strong> Sie koennen
                die Loeschung Ihrer personenbezogenen Daten verlangen, sofern
                keine gesetzlichen Aufbewahrungspflichten entgegenstehen.
              </li>
              <li>
                <strong>Einschraenkung der Verarbeitung (Art. 18 DSGVO):</strong>{' '}
                Sie koennen die Einschraenkung der Verarbeitung Ihrer Daten
                verlangen.
              </li>
              <li>
                <strong>Datenuebertragbarkeit (Art. 20 DSGVO):</strong> Sie
                haben das Recht, Ihre Daten in einem strukturierten, gaengigen
                und maschinenlesbaren Format zu erhalten.
              </li>
              <li>
                <strong>Widerspruchsrecht (Art. 21 DSGVO):</strong> Sie koennen
                jederzeit gegen die Verarbeitung Ihrer personenbezogenen Daten
                Widerspruch einlegen, sofern die Verarbeitung auf berechtigten
                Interessen basiert.
              </li>
              <li>
                <strong>Beschwerderecht (Art. 77 DSGVO):</strong> Sie haben das
                Recht, sich bei einer Datenschutz-Aufsichtsbehoerde ueber die
                Verarbeitung Ihrer personenbezogenen Daten zu beschweren.
              </li>
            </ul>
          </section>

          {/* 9. Datensicherheit */}
          <section className="legal-section">
            <h2>9. Datensicherheit</h2>
            <p>
              Wir setzen technische und organisatorische Sicherheitsmassnahmen
              ein, um Ihre Daten gegen zufaellige oder vorsaetzliche
              Manipulationen, Verlust, Zerstoerung oder den Zugriff
              unberechtigter Personen zu schuetzen. Dazu gehoeren unter anderem:
            </p>
            <ul>
              <li>Verschluesselte Datenuebertragung (TLS/SSL)</li>
              <li>Verschluesselte Passwortspeicherung (Argon2id)</li>
              <li>Rollenbasierte Zugriffskontrolle (RBAC)</li>
              <li>Brute-Force-Schutz bei der Anmeldung</li>
              <li>Regelmaessige Sicherheitsupdates</li>
            </ul>
          </section>

          {/* 10. Aenderungen */}
          <section className="legal-section">
            <h2>10. Aenderung dieser Datenschutzerklaerung</h2>
            <p>
              Wir behalten uns vor, diese Datenschutzerklaerung anzupassen, damit
              sie stets den aktuellen rechtlichen Anforderungen entspricht oder
              um Aenderungen unserer Leistungen umzusetzen. Fuer Ihren erneuten
              Besuch gilt dann die neue Datenschutzerklaerung.
            </p>
            <p>
              <em>Stand: Maerz 2026</em>
            </p>
          </section>
        </div>

        <div className="legal-card__footer">
          <Link to="/impressum" className="legal-card__link">
            Impressum
          </Link>
        </div>
      </div>
    </div>
  )
}

export default DatenschutzPage
