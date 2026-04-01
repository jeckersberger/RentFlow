import { Project } from '../../../types/project'
import '../ProjectDetail.scss'

interface OverviewTabProps {
  project: Project
}

const STATUS_LABELS: Record<string, string> = {
  draft: 'Entwurf',
  quoted: 'Angebot',
  confirmed: 'Bestätigt',
  in_progress: 'In Bearbeitung',
  completed: 'Abgeschlossen',
  cancelled: 'Storniert',
  invoiced: 'Fakturiert',
}

export function OverviewTab({ project }: OverviewTabProps) {
  const venueAddress = project.venue_address
  const venueString = venueAddress
    ? [venueAddress.street, venueAddress.postal_code, venueAddress.city, venueAddress.country]
        .filter(Boolean)
        .join(', ')
    : project.location || null

  const clientAddress = project.client_address
  const clientString = clientAddress
    ? [clientAddress.street, clientAddress.postal_code, clientAddress.city, clientAddress.country]
        .filter(Boolean)
        .join(', ')
    : null

  const mapsUrl = (address: string) =>
    `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(address)}`

  const formatDate = (date: string) => new Date(date).toLocaleDateString('de-DE', {
    weekday: 'short',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  })

  // Progress items based on what data exists
  const progressItems = [
    { label: 'Material', icon: '\u{1F4E6}', done: false },
    { label: 'Crew', icon: '\u{1F465}', done: false },
    { label: 'Angebot', icon: '\u{1F4CB}', done: project.status !== 'draft' },
    { label: 'Rechnung', icon: '\u{1F4B0}', done: project.status === 'invoiced' },
    { label: 'Tasks', icon: '\u{2705}', done: project.status === 'completed' },
  ]

  return (
    <div>
      <div className="overviewGrid">
        {/* Project Info Card */}
        <div className="glassCard">
          <div className="glassCardTitle">
            <span className="glassCardTitleIcon">i</span>
            Projektinformationen
          </div>
          <div className="infoRow">
            <span className="infoLabel">Name</span>
            <span className="infoValue">{project.name}</span>
          </div>
          <div className="infoRow">
            <span className="infoLabel">Status</span>
            <span className="infoValue">
              {STATUS_LABELS[project.status] || project.status}
            </span>
          </div>
          {project.project_manager && (
            <div className="infoRow">
              <span className="infoLabel">Projektleiter</span>
              <span className="infoValue">{project.project_manager}</span>
            </div>
          )}
          {project.budget != null && project.budget > 0 && (
            <div className="infoRow">
              <span className="infoLabel">Budget</span>
              <span className="infoValue">
                {project.budget.toLocaleString('de-DE', { style: 'currency', currency: project.currency || 'EUR' })}
              </span>
            </div>
          )}
          <div className="infoRow">
            <span className="infoLabel">Planungszeitraum</span>
            <span className="infoValue">
              {formatDate(project.start_date)} &ndash; {formatDate(project.end_date)}
            </span>
          </div>
          {project.setup_date && (
            <div className="infoRow">
              <span className="infoLabel">Aufbau</span>
              <span className="infoValue">{formatDate(project.setup_date)}</span>
            </div>
          )}
          {project.teardown_date && (
            <div className="infoRow">
              <span className="infoLabel">Abbau</span>
              <span className="infoValue">{formatDate(project.teardown_date)}</span>
            </div>
          )}
        </div>

        {/* Client Card */}
        <div className="glassCard">
          <div className="glassCardTitle">
            <span className="glassCardTitleIcon">C</span>
            Auftraggeber
          </div>
          {project.client_name ? (
            <>
              <div className="infoRow">
                <span className="infoLabel">Name</span>
                <span className="infoValue">{project.client_name}</span>
              </div>
              {project.client_email && (
                <div className="infoRow">
                  <span className="infoLabel">E-Mail</span>
                  <a href={`mailto:${project.client_email}`} className="infoLink">
                    {project.client_email}
                  </a>
                </div>
              )}
              {project.client_phone && (
                <div className="infoRow">
                  <span className="infoLabel">Telefon</span>
                  <a href={`tel:${project.client_phone}`} className="infoLink">
                    {project.client_phone}
                  </a>
                </div>
              )}
              {clientString && (
                <div className="infoRow">
                  <span className="infoLabel">Adresse</span>
                  <span className="infoValue">
                    {clientString}
                    <br />
                    <a
                      href={mapsUrl(clientString)}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="infoLink"
                    >
                      Google Maps
                    </a>
                  </span>
                </div>
              )}
            </>
          ) : (
            <p style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-sm)', margin: 0 }}>
              Kein Auftraggeber zugewiesen.
            </p>
          )}
        </div>

        {/* Location Card */}
        <div className="glassCard">
          <div className="glassCardTitle">
            <span className="glassCardTitleIcon">L</span>
            Standort / Venue
          </div>
          {venueString ? (
            <>
              <div className="infoRow">
                <span className="infoLabel">Adresse</span>
                <span className="infoValue">
                  {venueString}
                  <br />
                  <a
                    href={mapsUrl(venueString)}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="infoLink"
                  >
                    Google Maps
                  </a>
                </span>
              </div>
            </>
          ) : (
            <p style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-sm)', margin: 0 }}>
              Kein Standort angegeben.
            </p>
          )}
        </div>

        {/* Progress Card */}
        <div className="glassCard">
          <div className="glassCardTitle">
            <span className="glassCardTitleIcon">P</span>
            Projektfortschritt
          </div>
          <div className="progressIcons">
            {progressItems.map((item) => (
              <div key={item.label} className="progressItem">
                <div
                  className={`progressIcon ${
                    item.done ? 'progressIconDone' : 'progressIconPending'
                  }`}
                >
                  {item.icon}
                </div>
                <span className="progressLabel">{item.label}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Tags */}
        {project.tags && project.tags.length > 0 && (
          <div className="glassCard overviewFullWidth">
            <div className="glassCardTitle">Tags</div>
            <div className="tags">
              {project.tags.map((tag) => (
                <span key={tag} className="tag">{tag}</span>
              ))}
            </div>
          </div>
        )}

        {/* Notes */}
        <div className="glassCard overviewFullWidth">
          <div className="glassCardTitle">Notizen</div>
          {project.notes ? (
            <p className="notesText">{project.notes}</p>
          ) : (
            <p style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-sm)', margin: 0 }}>
              Keine Notizen vorhanden.
            </p>
          )}
        </div>

        {/* Quick Stats */}
        <div className="glassCard overviewFullWidth">
          <div className="glassCardTitle">Statistiken</div>
          <div className="quickStats">
            <div className="statItem">
              <div className="statValue">&mdash;</div>
              <div className="statLabel">Gewicht (kg)</div>
            </div>
            <div className="statItem">
              <div className="statValue">&mdash;</div>
              <div className="statLabel">Volumen (m&sup3;)</div>
            </div>
            <div className="statItem">
              <div className="statValue">&mdash;</div>
              <div className="statLabel">Leistung (kW)</div>
            </div>
          </div>
        </div>
      </div>

      {/* Description */}
      {project.description && (
        <div className="glassCard" style={{ marginTop: 'var(--spacing-5)' }}>
          <div className="glassCardTitle">Beschreibung</div>
          <p className="notesText">{project.description}</p>
        </div>
      )}
    </div>
  )
}
