import { Link } from 'react-router-dom'
import './Legal.scss'

const COMPANY = {
  name: 'JE-Sound&Light',
  street: 'Buehlstrasse 2',
  zip: '90610',
  city: 'Winkelhaid',
  country: 'Deutschland',
  ceo: 'Janis Eckersberger',
  email: 'j.eckersberger@je-soundulight.de',
  phone: '+49 1523 7858522',
  ustIdNr: '', // Kleinunternehmer — keine USt-IdNr
  handelsregister: '', // Einzelunternehmen — kein HR-Eintrag
}

function ImpressumPage() {
  return (
    <div className="legal-page">
      <div className="legal-card">
        <div className="legal-card__header">
          <Link to="/login" className="legal-card__back">Zurueck zur Anmeldung</Link>
          <h1 className="legal-card__title">Impressum</h1>
        </div>

        <div className="legal-card__content">
          <section className="legal-section">
            <h2>Angaben gemaess &sect; 5 TMG</h2>
            <p>
              {COMPANY.name}<br />
              {COMPANY.street}<br />
              {COMPANY.zip} {COMPANY.city}<br />
              {COMPANY.country}
            </p>
          </section>

          <section className="legal-section">
            <h2>Vertreten durch</h2>
            <p>Geschaeftsfuehrer: {COMPANY.ceo}</p>
          </section>

          <section className="legal-section">
            <h2>Kontakt</h2>
            <p>
              Telefon: {COMPANY.phone}<br />
              E-Mail: {COMPANY.email}
            </p>
          </section>

          {COMPANY.handelsregister && (
            <section className="legal-section">
              <h2>Handelsregister</h2>
              <p>{COMPANY.handelsregister}</p>
            </section>
          )}

          {COMPANY.ustIdNr && (
            <section className="legal-section">
              <h2>Umsatzsteuer-Identifikationsnummer</h2>
              <p>
                Umsatzsteuer-Identifikationsnummer gemaess &sect; 27a
                Umsatzsteuergesetz: {COMPANY.ustIdNr}
              </p>
            </section>
          )}

          <section className="legal-section">
            <h2>Verantwortlich fuer den Inhalt nach &sect; 55 Abs. 2 RStV</h2>
            <p>
              {COMPANY.ceo}<br />
              {COMPANY.street}<br />
              {COMPANY.zip} {COMPANY.city}
            </p>
          </section>

          <section className="legal-section">
            <h2>Streitschlichtung</h2>
            <p>
              Die Europaeische Kommission stellt eine Plattform zur
              Online-Streitbeilegung (OS) bereit:{' '}
              <a
                href="https://ec.europa.eu/consumers/odr/"
                target="_blank"
                rel="noopener noreferrer"
              >
                https://ec.europa.eu/consumers/odr/
              </a>
            </p>
            <p>
              Wir sind nicht bereit oder verpflichtet, an
              Streitbeilegungsverfahren vor einer Verbraucherschlichtungsstelle
              teilzunehmen.
            </p>
          </section>

          <section className="legal-section">
            <h2>Haftung fuer Inhalte</h2>
            <p>
              Als Diensteanbieter sind wir gemaess &sect; 7 Abs. 1 TMG fuer eigene
              Inhalte auf diesen Seiten nach den allgemeinen Gesetzen
              verantwortlich. Nach &sect;&sect; 8 bis 10 TMG sind wir als
              Diensteanbieter jedoch nicht verpflichtet, uebermittelte oder
              gespeicherte fremde Informationen zu ueberwachen oder nach
              Umstaenden zu forschen, die auf eine rechtswidrige Taetigkeit
              hinweisen.
            </p>
          </section>
        </div>

        <div className="legal-card__footer">
          <Link to="/datenschutz" className="legal-card__link">
            Datenschutzerklaerung
          </Link>
        </div>
      </div>
    </div>
  )
}

export default ImpressumPage
