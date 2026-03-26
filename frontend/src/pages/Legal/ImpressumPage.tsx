import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { tenantApi, configApi } from '../../services/api'
import { useAuthStore } from '../../stores/authStore'
import './Legal.scss'

const FALLBACK = {
  name: 'JE-Sound&Light',
  street: 'Buehlstrasse 2',
  zip: '90610',
  city: 'Winkelhaid',
  country: 'Deutschland',
  ceo: 'Janis Eckersberger',
  email: 'j.eckersberger@je-soundulight.de',
  phone: '+49 1523 7858522',
  ustIdNr: '',
  taxNumber: '',
  handelsregister: '',
}

function ImpressumPage() {
  const tenantId = useAuthStore((s) => s.tenantId)

  const { data: tenantData } = useQuery({
    queryKey: ['tenant', tenantId],
    queryFn: () => tenantApi.getById(tenantId!),
    enabled: !!tenantId,
    retry: false,
  })

  const { data: companyConfig } = useQuery({
    queryKey: ['config', 'company.details'],
    queryFn: () => configApi.get('company.details'),
    enabled: !!tenantId,
    retry: false,
  })

  const company = {
    name: tenantData?.name || FALLBACK.name,
    street: tenantData?.address_street || FALLBACK.street,
    zip: tenantData?.address_zip || FALLBACK.zip,
    city: tenantData?.address_city || FALLBACK.city,
    country: tenantData?.address_country || FALLBACK.country,
    ceo: companyConfig?.managing_director || tenantData?.managing_director || FALLBACK.ceo,
    email: tenantData?.email || FALLBACK.email,
    phone: tenantData?.phone || FALLBACK.phone,
    ustIdNr: companyConfig?.vat_id || tenantData?.vat_id || FALLBACK.ustIdNr,
    taxNumber: companyConfig?.tax_number || FALLBACK.taxNumber,
    handelsregister: companyConfig?.trade_register || tenantData?.trade_register || FALLBACK.handelsregister,
  }

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
              {company.name}<br />
              {company.street}<br />
              {company.zip} {company.city}<br />
              {company.country}
            </p>
          </section>

          <section className="legal-section">
            <h2>Vertreten durch</h2>
            <p>Geschaeftsfuehrer: {company.ceo}</p>
          </section>

          <section className="legal-section">
            <h2>Kontakt</h2>
            <p>
              Telefon: {company.phone}<br />
              E-Mail: {company.email}
            </p>
          </section>

          {company.handelsregister && (
            <section className="legal-section">
              <h2>Handelsregister</h2>
              <p>{company.handelsregister}</p>
            </section>
          )}

          {company.taxNumber && (
            <section className="legal-section">
              <h2>Steuernummer</h2>
              <p>{company.taxNumber}</p>
            </section>
          )}

          {company.ustIdNr && (
            <section className="legal-section">
              <h2>Umsatzsteuer-Identifikationsnummer</h2>
              <p>
                Umsatzsteuer-Identifikationsnummer gemaess &sect; 27a
                Umsatzsteuergesetz: {company.ustIdNr}
              </p>
            </section>
          )}

          <section className="legal-section">
            <h2>Verantwortlich fuer den Inhalt nach &sect; 55 Abs. 2 RStV</h2>
            <p>
              {company.ceo}<br />
              {company.street}<br />
              {company.zip} {company.city}
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
