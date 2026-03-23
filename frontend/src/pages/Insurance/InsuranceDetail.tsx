import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import type { InsuranceClaim, ClaimStatusTimeline, InsurancePolicy } from '../../types/insurance'
import './Insurance.module.scss'

const mockPolicies: Record<string, InsurancePolicy> = {
  '1': {
    id: '1',
    policy_number: 'LV-2024-001',
    type: 'liability',
    provider: 'Allianz Versicherung',
    coverage_amount: 1000000,
    premium_annual: 2500,
    start_date: '2024-01-01',
    end_date: '2026-12-31',
    is_active: true,
    deductible: 500,
  },
  '2': {
    id: '2',
    policy_number: 'EQ-2024-002',
    type: 'equipment',
    provider: 'AXA Versicherung',
    coverage_amount: 500000,
    premium_annual: 3800,
    start_date: '2024-03-01',
    end_date: '2027-02-28',
    is_active: true,
    deductible: 1000,
  },
  '3': {
    id: '3',
    policy_number: 'VH-2024-003',
    type: 'vehicle',
    provider: 'Ergo Versicherung',
    coverage_amount: 750000,
    premium_annual: 4200,
    start_date: '2024-06-01',
    end_date: '2026-05-31',
    is_active: true,
    deductible: 750,
  },
}

const mockClaims: Record<string, InsuranceClaim> = {
  'c1': {
    id: 'c1',
    claim_number: 'SCH-2026-001',
    policy_id: '1',
    status: 'approved',
    reported_date: '2026-03-15',
    incident_date: '2026-03-14',
    description: 'Beschaedigung an Ausruestung waehrend Transport',
    total_claimed: 5000,
    total_approved: 4500,
    total_settled: 4500,
    items: [
      {
        id: 'i1',
        description: 'Beschaedigte Stahlkonstruktion',
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
  'c2': {
    id: 'c2',
    claim_number: 'SCH-2026-002',
    policy_id: '2',
    status: 'submitted',
    reported_date: '2026-03-20',
    incident_date: '2026-03-19',
    description: 'Diebstahl von Ausruestung von der Baustelle',
    total_claimed: 8500,
    items: [
      {
        id: 'i3',
        description: 'Hochwertige Beleuchtungsanlage',
        claimed_amount: 5000,
        cost_category: 'equipment',
      },
      {
        id: 'i4',
        description: 'Diverse Kleinteile und Zubehoer',
        claimed_amount: 3500,
        cost_category: 'equipment',
      },
    ],
    photos: [],
    assigned_adjuster: 'Sarah Wagner',
  },
  'c3': {
    id: 'c3',
    claim_number: 'SCH-2025-045',
    policy_id: '3',
    status: 'settled',
    reported_date: '2025-11-10',
    incident_date: '2025-11-09',
    description: 'Fahrzeugkollision bei Anlieferung',
    total_claimed: 12000,
    total_approved: 11500,
    total_settled: 11500,
    items: [
      {
        id: 'i5',
        description: 'Fahrzeugschadensatz und Reparatur',
        claimed_amount: 12000,
        approved_amount: 11500,
        cost_category: 'vehicle',
      },
    ],
    photos: [],
    assigned_adjuster: 'Klaus Schmidt',
  },
  'c4': {
    id: 'c4',
    claim_number: 'SCH-2025-038',
    policy_id: '1',
    status: 'rejected',
    reported_date: '2025-09-05',
    incident_date: '2025-09-04',
    description: 'Wasserschaden an Mischpult - unsachgemaesse Lagerung',
    total_claimed: 3200,
    items: [
      {
        id: 'i6',
        description: 'Digitales Mischpult Yamaha CL5',
        claimed_amount: 3200,
        cost_category: 'equipment',
      },
    ],
    photos: [],
    assigned_adjuster: 'Hans Mueller',
  },
  'c5': {
    id: 'c5',
    claim_number: 'SCH-2026-003',
    policy_id: '2',
    status: 'reported',
    reported_date: '2026-03-22',
    incident_date: '2026-03-21',
    description: 'Sturz eines Scheinwerfers waehrend Aufbau',
    total_claimed: 2800,
    items: [
      {
        id: 'i7',
        description: 'Moving Head Robe T1 Profile',
        claimed_amount: 2800,
        cost_category: 'equipment',
      },
    ],
    photos: [],
  },
}

const mockTimelines: Record<string, ClaimStatusTimeline[]> = {
  'c1': [
    { status: 'reported', timestamp: '2026-03-15T09:30:00Z', notes: 'Schadensfall gemeldet durch Projektleiter' },
    { status: 'documented', timestamp: '2026-03-15T14:00:00Z', notes: 'Fotos und Dokumentation erhalten' },
    { status: 'submitted', timestamp: '2026-03-16T10:30:00Z', notes: 'Unterlagen bei Allianz eingereicht' },
    { status: 'approved', timestamp: '2026-03-20T15:45:00Z', notes: 'Anspruch genehmigt - 4.500 EUR' },
  ],
  'c2': [
    { status: 'reported', timestamp: '2026-03-20T08:15:00Z', notes: 'Diebstahl gemeldet, Polizeibericht erstellt' },
    { status: 'documented', timestamp: '2026-03-20T11:00:00Z', notes: 'Inventarliste und Polizeiaktenzeichen dokumentiert' },
    { status: 'submitted', timestamp: '2026-03-21T09:00:00Z', notes: 'Unterlagen bei AXA eingereicht, Aktenzeichen: AXA-2026-4782' },
  ],
  'c3': [
    { status: 'reported', timestamp: '2025-11-10T07:45:00Z', notes: 'Unfall bei Anlieferung gemeldet' },
    { status: 'documented', timestamp: '2025-11-10T12:30:00Z', notes: 'Gutachter bestellt, Fotos erstellt' },
    { status: 'submitted', timestamp: '2025-11-11T10:00:00Z', notes: 'Schadenmeldung bei Ergo eingereicht' },
    { status: 'approved', timestamp: '2025-11-25T14:20:00Z', notes: 'Anspruch genehmigt - 11.500 EUR' },
    { status: 'settled', timestamp: '2025-12-05T09:00:00Z', notes: 'Zahlung eingegangen - 11.500 EUR' },
  ],
  'c4': [
    { status: 'reported', timestamp: '2025-09-05T10:00:00Z', notes: 'Wasserschaden gemeldet' },
    { status: 'documented', timestamp: '2025-09-05T15:30:00Z', notes: 'Dokumentation erstellt' },
    { status: 'submitted', timestamp: '2025-09-06T09:00:00Z', notes: 'Bei Allianz eingereicht' },
    { status: 'rejected', timestamp: '2025-09-20T11:00:00Z', notes: 'Abgelehnt - unsachgemaesse Lagerung, kein Versicherungsschutz' },
  ],
  'c5': [
    { status: 'reported', timestamp: '2026-03-22T16:00:00Z', notes: 'Schaden waehrend Aufbau gemeldet' },
  ],
}

function InsuranceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: claim, isLoading } = useQuery({
    queryKey: ['claim', id],
    queryFn: async () => mockClaims[id || 'c1'] || null,
    staleTime: 1000 * 60 * 5,
  })

  const timeline = mockTimelines[id || 'c1'] || []
  const policy = claim ? mockPolicies[claim.policy_id] : null

  const getStatusLabel = (status: string): string => {
    const labels: Record<string, string> = {
      reported: 'Gemeldet',
      documented: 'Dokumentiert',
      submitted: 'Eingereicht',
      approved: 'Genehmigt',
      settled: 'Abgewickelt',
      rejected: 'Abgelehnt',
    }
    return labels[status] || status
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

  const getCategoryLabel = (category: string): string => {
    const labels: Record<string, string> = {
      equipment: 'Ausruestung',
      labor: 'Arbeitsleistung',
      vehicle: 'Fahrzeug',
      material: 'Material',
      other: 'Sonstiges',
    }
    return labels[category] || category
  }

  const getPolicyTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      liability: 'Haftpflicht',
      equipment: 'Ausruestung',
      vehicle: 'Fahrzeuge',
      workers_comp: 'Unfallversicherung',
    }
    return labels[type] || type
  }

  if (isLoading) {
    return (
      <div className="insurance-detail-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Schadensfall laden...</h1>
          </div>
        </div>
      </div>
    )
  }

  if (!claim) {
    return (
      <div className="insurance-detail-page">
        <div className="page-header">
          <div>
            <button
              className="btn btn--sm btn--secondary"
              onClick={() => navigate('/insurance')}
              style={{ marginBottom: 'var(--spacing-3)' }}
            >
              {'\u2190'} Zurueck zur Uebersicht
            </button>
            <h1 className="page-title">Schadensfall nicht gefunden</h1>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">{'\u{1F50D}'}</div>
          <h3 className="empty-state__title">Schadensfall nicht gefunden</h3>
          <p className="empty-state__description">Der angeforderte Schadensfall existiert nicht oder wurde geloescht.</p>
        </div>
      </div>
    )
  }

  // Determine which steps in the full flow are completed
  const allSteps = ['reported', 'documented', 'submitted', 'approved', 'settled']
  const rejectedAtIndex = claim.status === 'rejected'
    ? timeline.findIndex(t => t.status === 'rejected')
    : -1

  return (
    <div className="insurance-detail-page">
      <div className="page-header">
        <div>
          <button
            className="btn btn--sm btn--secondary"
            onClick={() => navigate('/insurance')}
            style={{ marginBottom: 'var(--spacing-3)' }}
          >
            {'\u2190'} Zurueck zur Uebersicht
          </button>
          <h1 className="page-title">Schadensfall {claim.claim_number}</h1>
          <p className="page-subtitle">{claim.description}</p>
        </div>
        <span className={`claim-status claim-status--${claim.status}`} style={{ fontSize: 'var(--font-size-sm)', padding: 'var(--spacing-2) var(--spacing-4)' }}>
          {getStatusLabel(claim.status)}
        </span>
      </div>

      <div className="detail-grid">
        {/* Main Content */}
        <div className="detail-column">
          {/* Status Progress Bar */}
          <div className="detail-card">
            <h2 className="detail-card__title">Bearbeitungsfortschritt</h2>
            <div className="status-progress">
              {allSteps.map((step, idx) => {
                const timelineEntry = timeline.find(t => t.status === step)
                const isCompleted = !!timelineEntry
                const isCurrent = step === claim.status
                const isRejected = claim.status === 'rejected' && !timelineEntry && idx > (rejectedAtIndex >= 0 ? rejectedAtIndex : timeline.length)

                return (
                  <div key={step} className={`status-progress__step ${isCompleted ? 'status-progress__step--completed' : ''} ${isCurrent ? 'status-progress__step--current' : ''} ${isRejected ? 'status-progress__step--disabled' : ''}`}>
                    <div className="status-progress__dot">
                      {isCompleted ? '\u2713' : (idx + 1)}
                    </div>
                    <span className="status-progress__label">{getStatusLabel(step)}</span>
                    {idx < allSteps.length - 1 && (
                      <div className={`status-progress__connector ${isCompleted && timeline.find(t => t.status === allSteps[idx + 1]) ? 'status-progress__connector--active' : ''}`}></div>
                    )}
                  </div>
                )
              })}
              {claim.status === 'rejected' && (
                <div className="status-progress__step status-progress__step--rejected">
                  <div className="status-progress__dot">{'\u2715'}</div>
                  <span className="status-progress__label">Abgelehnt</span>
                </div>
              )}
            </div>
          </div>

          {/* Claim Timeline */}
          <div className="detail-card">
            <h2 className="detail-card__title">Bearbeitungs-Verlauf</h2>
            <div className="claim-timeline">
              {timeline.map((step, index) => (
                <div key={step.status + index} className={`timeline-step ${step.status === claim.status ? 'timeline-step--current' : ''}`}>
                  <div className="timeline-step__dot" style={{ color: getStatusColor(step.status), borderColor: getStatusColor(step.status) }}>
                    {step.status === 'rejected' ? '\u2715' : '\u2713'}
                  </div>
                  <div className="timeline-step__content">
                    <h4 className="timeline-step__status" style={{ color: getStatusColor(step.status) }}>
                      {getStatusLabel(step.status)}
                    </h4>
                    <p className="timeline-step__date">
                      {new Date(step.timestamp).toLocaleDateString('de-DE')} um {new Date(step.timestamp).toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })}
                    </p>
                    {step.notes && (
                      <p className="timeline-step__notes">{step.notes}</p>
                    )}
                  </div>
                  {index < timeline.length - 1 && (
                    <div className="timeline-step__connector" style={{ backgroundColor: getStatusColor(step.status) }}></div>
                  )}
                </div>
              ))}
            </div>
          </div>

          {/* Claim Items */}
          <div className="detail-card">
            <h2 className="detail-card__title">Schadensposten ({claim.items.length})</h2>
            <div className="claim-items-list">
              {claim.items.map(item => (
                <div key={item.id} className="claim-item">
                  <div className="claim-item__header">
                    <h4 className="claim-item__description">{item.description}</h4>
                    <span className="claim-item__category">{getCategoryLabel(item.cost_category)}</span>
                  </div>
                  <div className="claim-item__amounts">
                    <div className="amount-block">
                      <span className="amount-block__label">Gefordert:</span>
                      <span className="amount-block__value">{'\u20AC'}{item.claimed_amount.toLocaleString('de-DE')}</span>
                    </div>
                    {item.approved_amount !== undefined && (
                      <div className="amount-block">
                        <span className="amount-block__label">Genehmigt:</span>
                        <span className="amount-block__value" style={{ color: 'var(--color-success)' }}>
                          {'\u20AC'}{item.approved_amount.toLocaleString('de-DE')}
                        </span>
                      </div>
                    )}
                    {item.approved_amount !== undefined && item.approved_amount < item.claimed_amount && (
                      <div className="amount-block">
                        <span className="amount-block__label">Differenz:</span>
                        <span className="amount-block__value" style={{ color: 'var(--color-warning)' }}>
                          -{'\u20AC'}{(item.claimed_amount - item.approved_amount).toLocaleString('de-DE')}
                        </span>
                      </div>
                    )}
                  </div>
                  {item.receipt_url && (
                    <div className="claim-item__receipt">
                      <a href={item.receipt_url} target="_blank" rel="noopener noreferrer">
                        Quittung ansehen
                      </a>
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* Items Summary */}
            <div className="claim-items-summary">
              <div className="claim-items-summary__row">
                <span>Gesamtforderung:</span>
                <strong>{'\u20AC'}{claim.items.reduce((sum, i) => sum + i.claimed_amount, 0).toLocaleString('de-DE')}</strong>
              </div>
              {claim.total_approved !== undefined && (
                <div className="claim-items-summary__row">
                  <span>Gesamt genehmigt:</span>
                  <strong style={{ color: 'var(--color-success)' }}>{'\u20AC'}{claim.total_approved.toLocaleString('de-DE')}</strong>
                </div>
              )}
            </div>
          </div>

          {/* Photo Gallery / Upload placeholder */}
          <div className="detail-card">
            <h2 className="detail-card__title">Fotos und Dokumente</h2>
            {claim.photos && claim.photos.length > 0 ? (
              <div className="photo-gallery">
                {claim.photos.map((photo, index) => (
                  <div key={index} className="photo-item">
                    <img src={photo} alt={`Schadensfall Foto ${index + 1}`} />
                  </div>
                ))}
              </div>
            ) : (
              <div className="photo-upload-placeholder">
                <div className="photo-upload-placeholder__icon">{'\u{1F4F7}'}</div>
                <p className="photo-upload-placeholder__text">Noch keine Fotos hochgeladen</p>
                <button className="btn btn--sm btn--secondary">Fotos hochladen</button>
              </div>
            )}
          </div>
        </div>

        {/* Sidebar */}
        <div className="detail-sidebar">
          {/* Claim Summary */}
          <div className="detail-card">
            <h2 className="detail-card__title">Zusammenfassung</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <label className="detail-card__row-label">Schadennummer</label>
                <span className="detail-card__row-value">{claim.claim_number}</span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Status</label>
                <span className="detail-card__row-value">
                  <span className={`claim-status claim-status--${claim.status}`}>
                    {getStatusLabel(claim.status)}
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
              <div className="detail-card__row">
                <label className="detail-card__row-label">Posten</label>
                <span className="detail-card__row-value">{claim.items.length}</span>
              </div>
            </div>
          </div>

          {/* Policy Info */}
          {policy && (
            <div className="detail-card">
              <h2 className="detail-card__title">Versicherungspolice</h2>
              <div className="detail-card__content">
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Versicherer</label>
                  <span className="detail-card__row-value">{policy.provider}</span>
                </div>
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Police Nr.</label>
                  <span className="detail-card__row-value">{policy.policy_number}</span>
                </div>
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Typ</label>
                  <span className="detail-card__row-value">{getPolicyTypeLabel(policy.type)}</span>
                </div>
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Deckung</label>
                  <span className="detail-card__row-value">{'\u20AC'}{policy.coverage_amount.toLocaleString('de-DE')}</span>
                </div>
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Selbstbeteiligung</label>
                  <span className="detail-card__row-value">{'\u20AC'}{policy.deductible.toLocaleString('de-DE')}</span>
                </div>
              </div>
            </div>
          )}

          {/* Settlement Tracking */}
          <div className="detail-card">
            <h2 className="detail-card__title">Abwicklung</h2>
            <div className="settlement-tracking">
              <div className="settlement-row">
                <span className="settlement-label">Gefordert:</span>
                <span className="settlement-value">{'\u20AC'}{claim.total_claimed.toLocaleString('de-DE')}</span>
              </div>
              {claim.total_approved !== undefined && (
                <div className="settlement-row">
                  <span className="settlement-label">Genehmigt:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-success)' }}>
                    {'\u20AC'}{claim.total_approved.toLocaleString('de-DE')}
                  </span>
                </div>
              )}
              {claim.total_settled !== undefined && (
                <div className="settlement-row">
                  <span className="settlement-label">Abgewickelt:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-success)' }}>
                    {'\u20AC'}{claim.total_settled.toLocaleString('de-DE')}
                  </span>
                </div>
              )}
              {claim.status === 'rejected' && (
                <div className="settlement-row">
                  <span className="settlement-label">Ergebnis:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-danger)' }}>
                    Abgelehnt
                  </span>
                </div>
              )}

              {claim.total_approved !== undefined && (
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

              {claim.total_settled !== undefined && claim.total_approved !== undefined && (
                <div className="settlement-progress">
                  <div className="progress-label">
                    <span>Auszahlungsquote</span>
                    <span>{Math.round((claim.total_settled / claim.total_approved) * 100)}%</span>
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress-bar__fill"
                      style={{ width: `${(claim.total_settled / claim.total_approved) * 100}%` }}
                    ></div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Action Buttons */}
          <div className="detail-card">
            <div className="action-buttons-vertical">
              <button
                className="btn btn--secondary"
                onClick={() => alert('Dokumente werden als PDF exportiert...')}
              >
                PDF exportieren
              </button>
              <button
                className="btn btn--secondary"
                onClick={() => alert('E-Mail an Versicherer wird vorbereitet...')}
              >
                Versicherer kontaktieren
              </button>
              {claim.status === 'reported' && (
                <button
                  className="btn btn--primary"
                  onClick={() => alert('Dokumentation wird vorbereitet...')}
                >
                  Dokumentation einreichen
                </button>
              )}
              {claim.status === 'approved' && !claim.total_settled && (
                <button
                  className="btn btn--primary"
                  onClick={() => alert('Zahlungseingang wird vermerkt...')}
                >
                  Zahlung vermerken
                </button>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default InsuranceDetail
