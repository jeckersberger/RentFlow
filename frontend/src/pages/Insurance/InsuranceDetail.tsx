import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import type { InsuranceClaim, ClaimStatusTimeline } from '../../types/insurance'
import './Insurance.module.scss'

const mockClaims: Record<string, InsuranceClaim> = {
  'c1': {
    id: 'c1',
    claim_number: 'SCH-2026-001',
    policy_id: '1',
    status: 'approved',
    reported_date: '2026-03-15',
    incident_date: '2026-03-14',
    description: 'Beschädigung an Ausrüstung während Transport',
    total_claimed: 5000,
    total_approved: 4500,
    total_settled: 4500,
    items: [
      {
        id: 'i1',
        description: 'Beschädigte Stahlkonstruktion',
        claimed_amount: 3000,
        approved_amount: 3000,
        cost_category: 'equipment',
      },
      {
        id: 'i2',
        description: 'Reparaturbericht und Inspektionen',
        claimed_amount: 2000,
        approved_amount: 1500,
        cost_category: 'labor',
      },
    ],
    photos: [],
    assigned_adjuster: 'Hans Mueller',
  },
}

const mockTimeline: ClaimStatusTimeline[] = [
  { status: 'reported', timestamp: '2026-03-15T09:30:00Z', notes: 'Schadensfall gemeldet' },
  { status: 'documented', timestamp: '2026-03-15T14:00:00Z', notes: 'Dokumentation erhalten' },
  { status: 'submitted', timestamp: '2026-03-16T10:30:00Z', notes: 'Unterlagen eingereicht' },
  { status: 'approved', timestamp: '2026-03-20T15:45:00Z', notes: 'Anspruch genehmigt' },
]

function InsuranceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: claim } = useQuery({
    queryKey: ['claim', id],
    queryFn: async () => mockClaims[id || 'c1'] || mockClaims['c1'],
    staleTime: 1000 * 60 * 5,
  })

  if (!claim) {
    return <div className="empty-state">Schadensfall nicht gefunden</div>
  }

  const getStatusColor = (status: string): string => {
    const colors: Record<string, string> = {
      reported: 'var(--color-warning)',
      documented: 'var(--color-primary)',
      submitted: 'var(--color-primary)',
      approved: 'var(--color-success)',
      settled: 'var(--color-success)',
      rejected: 'var(--color-danger)',
    }
    return colors[status] || 'var(--color-text-secondary)'
  }

  const getStatusEmoji = (status: string): string => {
    const emojis: Record<string, string> = {
      reported: '📌',
      documented: '📋',
      submitted: '📤',
      approved: '✓',
      settled: '✓',
      rejected: '✕',
    }
    return emojis[status] || '•'
  }

  return (
    <div className="insurance-detail-page">
      <div className="page-header">
        <div>
          <button
            className="btn btn--sm btn--secondary"
            onClick={() => navigate('/insurance')}
            style={{ marginBottom: 'var(--spacing-3)' }}
          >
            ← Zurück
          </button>
          <h1 className="page-title">Schadensfall {claim.claim_number}</h1>
          <p className="page-subtitle">{claim.description}</p>
        </div>
      </div>

      <div className="detail-grid">
        {/* Main Content */}
        <div className="detail-column">
          {/* Claim Timeline */}
          <div className="detail-card">
            <h2 className="detail-card__title">📅 Bearbeitungs-Status</h2>
            <div className="claim-timeline">
              {mockTimeline.map((step, index) => (
                <div key={step.status} className="timeline-step">
                  <div className="timeline-step__dot" style={{ color: getStatusColor(step.status) }}>
                    {getStatusEmoji(step.status)}
                  </div>
                  <div className="timeline-step__content">
                    <h4 className="timeline-step__status">
                      {step.status.charAt(0).toUpperCase() + step.status.slice(1)}
                    </h4>
                    <p className="timeline-step__date">
                      {new Date(step.timestamp).toLocaleDateString('de-DE')} {new Date(step.timestamp).toLocaleTimeString('de-DE')}
                    </p>
                    {step.notes && (
                      <p className="timeline-step__notes">{step.notes}</p>
                    )}
                  </div>
                  {index < mockTimeline.length - 1 && (
                    <div className="timeline-step__connector"></div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Claim Items */}
          <div className="detail-card">
            <h2 className="detail-card__title">📋 Schadensposten</h2>
            <div className="claim-items-list">
              {claim.items.map(item => (
                <div key={item.id} className="claim-item">
                  <div className="claim-item__header">
                    <h4 className="claim-item__description">{item.description}</h4>
                    <span className="claim-item__category">{item.cost_category}</span>
                  </div>
                  <div className="claim-item__amounts">
                    <div className="amount-block">
                      <span className="amount-block__label">Gefordert:</span>
                      <span className="amount-block__value">€{item.claimed_amount.toLocaleString('de-DE')}</span>
                    </div>
                    {item.approved_amount && (
                      <div className="amount-block">
                        <span className="amount-block__label">Genehmigt:</span>
                        <span className="amount-block__value">€{item.approved_amount.toLocaleString('de-DE')}</span>
                      </div>
                    )}
                  </div>
                  {item.receipt_url && (
                    <div className="claim-item__receipt">
                      <a href={item.receipt_url} target="_blank" rel="noopener noreferrer">
                        📎 Quittung ansehen
                      </a>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Photo Gallery */}
          {claim.photos && claim.photos.length > 0 && (
            <div className="detail-card">
              <h2 className="detail-card__title">📸 Fotos</h2>
              <div className="photo-gallery">
                {claim.photos.map((photo, index) => (
                  <div key={index} className="photo-item">
                    <img src={photo} alt={`Schadensfall Foto ${index + 1}`} />
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* Sidebar */}
        <div className="detail-sidebar">
          {/* Claim Summary */}
          <div className="detail-card">
            <h2 className="detail-card__title">💼 Zusammenfassung</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <label className="detail-card__row-label">Schadennummer</label>
                <span className="detail-card__row-value">{claim.claim_number}</span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Status</label>
                <span className="detail-card__row-value">
                  <span style={{ color: getStatusColor(claim.status) }}>
                    {getStatusEmoji(claim.status)} {claim.status.charAt(0).toUpperCase() + claim.status.slice(1)}
                  </span>
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Schadendatum</label>
                <span className="detail-card__row-value">
                  {new Date(claim.incident_date).toLocaleDateString('de-DE')}
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Meldedatum</label>
                <span className="detail-card__row-value">
                  {new Date(claim.reported_date).toLocaleDateString('de-DE')}
                </span>
              </div>
              {claim.assigned_adjuster && (
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Sachbearbeiter</label>
                  <span className="detail-card__row-value">{claim.assigned_adjuster}</span>
                </div>
              )}
            </div>
          </div>

          {/* Settlement Tracking */}
          <div className="detail-card">
            <h2 className="detail-card__title">💰 Abwicklung</h2>
            <div className="settlement-tracking">
              <div className="settlement-row">
                <span className="settlement-label">Gefordert:</span>
                <span className="settlement-value">€{claim.total_claimed.toLocaleString('de-DE')}</span>
              </div>
              {claim.total_approved && (
                <div className="settlement-row">
                  <span className="settlement-label">Genehmigt:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-success)' }}>
                    €{claim.total_approved.toLocaleString('de-DE')}
                  </span>
                </div>
              )}
              {claim.total_settled && (
                <div className="settlement-row">
                  <span className="settlement-label">Abgewickelt:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-success)' }}>
                    €{claim.total_settled.toLocaleString('de-DE')}
                  </span>
                </div>
              )}

              {claim.total_approved && (
                <div className="settlement-progress">
                  <div className="progress-label">
                    <span>Genehmigungsquote</span>
                    <span>{Math.round((claim.total_approved / claim.total_claimed) * 100)}%</span>
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress-bar__fill"
                      style={{ width: `${(claim.total_approved / claim.total_claimed) * 100}%` }}
                    ></div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="detail-card">
            <div className="action-buttons">
              <button
                className="btn btn--secondary"
                onClick={() => alert('Dokumente werden heruntergeladen...')}
              >
                📥 Dokumente
              </button>
              <button
                className="btn btn--secondary"
                onClick={() => alert('E-Mail wird verfasst...')}
              >
                ✉️ Nachricht
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default InsuranceDetail
