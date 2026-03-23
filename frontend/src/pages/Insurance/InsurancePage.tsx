import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import type { InsurancePolicy, InsuranceClaim } from '../../types/insurance'
import './Insurance.scss'

const mockPolicies: InsurancePolicy[] = [
  {
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
  {
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
  {
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
  {
    id: '4',
    policy_number: 'UV-2024-004',
    type: 'workers_comp',
    provider: 'HDI Versicherung',
    coverage_amount: 2000000,
    premium_annual: 5600,
    start_date: '2024-01-01',
    end_date: '2025-12-31',
    is_active: false,
    deductible: 250,
  },
]

const mockClaims: InsuranceClaim[] = [
  {
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
  {
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
  {
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
  {
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
  {
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
]

interface NewClaimFormData {
  policy_id: string
  incident_date: string
  description: string
  affected_equipment: string
  estimated_damage: string
}

function InsurancePage() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'policies' | 'open_claims' | 'settled'>('policies')
  const [showNewClaimModal, setShowNewClaimModal] = useState(false)
  const [newClaimForm, setNewClaimForm] = useState<NewClaimFormData>({
    policy_id: '',
    incident_date: '',
    description: '',
    affected_equipment: '',
    estimated_damage: '',
  })
  const [claimSubmitted, setClaimSubmitted] = useState(false)

  const { data: policies = [], isLoading: policiesLoading, error: policiesError } = useQuery({
    queryKey: ['insurance-policies'],
    queryFn: async () => {
      return mockPolicies
    },
    staleTime: 1000 * 60 * 5,
  })

  const { data: claims = [], isLoading: claimsLoading } = useQuery({
    queryKey: ['insurance-claims'],
    queryFn: async () => {
      return mockClaims
    },
    staleTime: 1000 * 60 * 5,
  })

  const isLoading = policiesLoading || claimsLoading

  const openClaims = claims.filter(c => ['reported', 'documented', 'submitted', 'approved'].includes(c.status))
  const settledClaims = claims.filter(c => ['settled', 'rejected'].includes(c.status))
  const currentYearClaims = claims.filter(c => {
    const year = new Date(c.reported_date).getFullYear()
    return year === new Date().getFullYear()
  })
  const totalPremiums = policies.reduce((sum, p) => sum + p.premium_annual, 0)

  const getPolicyTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      liability: 'Haftpflicht',
      equipment: 'Ausruestung',
      vehicle: 'Fahrzeuge',
      workers_comp: 'Unfallversicherung',
    }
    return labels[type] || type
  }

  const getPolicyTypeIcon = (type: string): string => {
    const icons: Record<string, string> = {
      liability: '\u{1F6E1}',
      equipment: '\u{1F3AC}',
      vehicle: '\u{1F69A}',
      workers_comp: '\u{1F9D1}\u200D\u{1F527}',
    }
    return icons[type] || '\u{1F4CB}'
  }

  const getClaimStatusLabel = (status: string): string => {
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

  const handleNewClaimSubmit = () => {
    if (!newClaimForm.policy_id || !newClaimForm.incident_date || !newClaimForm.description) {
      return
    }
    setClaimSubmitted(true)
    setTimeout(() => {
      setShowNewClaimModal(false)
      setClaimSubmitted(false)
      setNewClaimForm({
        policy_id: '',
        incident_date: '',
        description: '',
        affected_equipment: '',
        estimated_damage: '',
      })
    }, 2000)
  }

  const getDaysUntilExpiry = (endDate: string): number => {
    const end = new Date(endDate)
    const now = new Date()
    return Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  }

  if (isLoading) {
    return (
      <div className="insurance-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Versicherung</h1>
            <p className="page-subtitle">Daten werden geladen...</p>
          </div>
        </div>
        <div className="stats-grid">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="stat-card stat-card--loading">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (policiesError) {
    return (
      <div className="insurance-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Versicherung</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten</p>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">{'\u26A0'}</div>
          <h3 className="empty-state__title">Daten konnten nicht geladen werden</h3>
          <p className="empty-state__description">{String(policiesError)}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="insurance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Versicherung</h1>
          <p className="page-subtitle">Verwaltung von Versicherungspolicen und Schadenfaellen</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => setShowNewClaimModal(true)}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          + Neuer Schadensfall
        </button>
      </div>

      {/* KPI Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__icon">{'\u{1F6E1}'}</div>
          <div className="stat-card__label">Aktive Policen</div>
          <div className="stat-card__value">{policies.filter(p => p.is_active).length}</div>
          <div className="stat-card__subtext">von {policies.length} gesamt</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__icon">{'\u26A0'}</div>
          <div className="stat-card__label">Offene Schadenfaelle</div>
          <div className="stat-card__value" style={{ color: 'var(--color-warning)' }}>{openClaims.length}</div>
          <div className="stat-card__subtext">{openClaims.filter(c => c.status === 'reported').length} neu gemeldet</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__icon">{'\u{1F4C5}'}</div>
          <div className="stat-card__label">Schadenfaelle dieses Jahr</div>
          <div className="stat-card__value">{currentYearClaims.length}</div>
          <div className="stat-card__subtext">
            {'\u20AC'}{currentYearClaims.reduce((sum, c) => sum + c.total_claimed, 0).toLocaleString('de-DE')} Gesamtschaden
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card__icon">{'\u{1F4B0}'}</div>
          <div className="stat-card__label">Jahrespraemien</div>
          <div className="stat-card__value" style={{ fontSize: 'var(--font-size-2xl)' }}>
            {'\u20AC'}{totalPremiums.toLocaleString('de-DE')}
          </div>
          <div className="stat-card__subtext">
            {'\u20AC'}{Math.round(totalPremiums / 12).toLocaleString('de-DE')}/Monat
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="tabs">
        <button
          className={`tab-button ${activeTab === 'policies' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('policies')}
        >
          Versicherungspolicen ({policies.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'open_claims' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('open_claims')}
        >
          Offene Schadenfaelle ({openClaims.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'settled' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('settled')}
        >
          Abgewickelte Schadenfaelle ({settledClaims.length})
        </button>
      </div>

      {/* Active Policies */}
      {activeTab === 'policies' && (
        <div className="policies-grid">
          {policies.map(policy => {
            const daysLeft = getDaysUntilExpiry(policy.end_date)
            const isExpiringSoon = daysLeft > 0 && daysLeft <= 90
            const isExpired = daysLeft <= 0

            return (
              <div key={policy.id} className={`policy-card ${isExpired ? 'policy-card--expired' : ''} ${isExpiringSoon ? 'policy-card--expiring' : ''}`}>
                <div className="policy-card__header">
                  <div className="policy-card__provider">
                    <span className="policy-card__type-icon">{getPolicyTypeIcon(policy.type)}</span>
                    <h3 className="policy-card__title">{policy.provider}</h3>
                  </div>
                  <span className={`policy-status ${policy.is_active ? 'policy-status--active' : 'policy-status--inactive'}`}>
                    {policy.is_active ? 'Aktiv' : 'Inaktiv'}
                  </span>
                </div>

                <div className="policy-card__type">
                  <span className="policy-type-badge">{getPolicyTypeLabel(policy.type)}</span>
                  {isExpiringSoon && (
                    <span className="policy-type-badge policy-type-badge--warning">
                      {daysLeft} Tage verbleibend
                    </span>
                  )}
                  {isExpired && (
                    <span className="policy-type-badge policy-type-badge--danger">
                      Abgelaufen
                    </span>
                  )}
                </div>

                <div className="policy-card__details">
                  <div className="detail-row">
                    <span className="detail-label">Police Nr.:</span>
                    <span className="detail-value">{policy.policy_number}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">Deckungssumme:</span>
                    <span className="detail-value">{'\u20AC'}{policy.coverage_amount.toLocaleString('de-DE')}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">Jahrespraemie:</span>
                    <span className="detail-value">{'\u20AC'}{policy.premium_annual.toLocaleString('de-DE')}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">Selbstbeteiligung:</span>
                    <span className="detail-value">{'\u20AC'}{policy.deductible.toLocaleString('de-DE')}</span>
                  </div>
                  <div className="detail-row">
                    <span className="detail-label">Gueltig bis:</span>
                    <span className={`detail-value ${isExpiringSoon ? 'detail-value--warning' : ''} ${isExpired ? 'detail-value--danger' : ''}`}>
                      {new Date(policy.end_date).toLocaleDateString('de-DE')}
                    </span>
                  </div>
                </div>

                <div className="policy-card__footer">
                  <button
                    className="btn btn--sm btn--secondary"
                    onClick={() => setShowNewClaimModal(true)}
                  >
                    Schadensfall melden
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}

      {/* Open Claims */}
      {activeTab === 'open_claims' && (
        <div className="claims-grid">
          {openClaims.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">{'\u2713'}</div>
              <h3 className="empty-state__title">Keine offenen Schadenfaelle</h3>
              <p className="empty-state__description">Das ist eine gute Nachricht!</p>
            </div>
          ) : (
            openClaims.map(claim => {
              const policy = policies.find(p => p.id === claim.policy_id)
              return (
                <div
                  key={claim.id}
                  className="claim-card"
                  onClick={() => navigate(`/insurance/claims/${claim.id}`)}
                >
                  <div className="claim-card__header">
                    <h3 className="claim-card__number">{claim.claim_number}</h3>
                    <span className={`claim-status claim-status--${claim.status}`}>
                      {getClaimStatusLabel(claim.status)}
                    </span>
                  </div>

                  {policy && (
                    <div className="claim-card__policy">
                      <span className="claim-card__policy-badge">
                        {getPolicyTypeIcon(policy.type)} {policy.provider}
                      </span>
                    </div>
                  )}

                  <div className="claim-card__description">{claim.description}</div>

                  <div className="claim-card__amounts">
                    <div className="amount-item">
                      <span className="amount-label">Gefordert:</span>
                      <span className="amount-value">{'\u20AC'}{claim.total_claimed.toLocaleString('de-DE')}</span>
                    </div>
                    {claim.total_approved && (
                      <div className="amount-item">
                        <span className="amount-label">Genehmigt:</span>
                        <span className="amount-value" style={{ color: 'var(--color-success)' }}>
                          {'\u20AC'}{claim.total_approved.toLocaleString('de-DE')}
                        </span>
                      </div>
                    )}
                  </div>

                  {/* Mini Status Timeline */}
                  <div className="claim-card__timeline-mini">
                    {(['reported', 'submitted', 'approved', 'settled'] as const).map((step) => {
                      const stepOrder = ['reported', 'submitted', 'approved', 'settled']
                      const currentOrder = stepOrder.indexOf(claim.status)
                      const thisOrder = stepOrder.indexOf(step)
                      const isComplete = thisOrder <= currentOrder
                      const isCurrent = step === claim.status

                      return (
                        <div key={step} className={`timeline-mini-step ${isComplete ? 'timeline-mini-step--complete' : ''} ${isCurrent ? 'timeline-mini-step--current' : ''}`}>
                          <div className="timeline-mini-step__dot"></div>
                          <span className="timeline-mini-step__label">{getClaimStatusLabel(step)}</span>
                        </div>
                      )
                    })}
                  </div>

                  <div className="claim-card__meta">
                    <p>
                      <strong>Schadendatum:</strong> {new Date(claim.incident_date).toLocaleDateString('de-DE')}
                    </p>
                    {claim.assigned_adjuster && (
                      <p>
                        <strong>Sachbearbeiter:</strong> {claim.assigned_adjuster}
                      </p>
                    )}
                    <p>
                      <strong>Posten:</strong> {claim.items.length}
                    </p>
                  </div>

                  <button
                    className="btn btn--sm btn--primary"
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/insurance/claims/${claim.id}`)
                    }}
                    style={{ width: '100%', marginTop: 'var(--spacing-3)' }}
                  >
                    Details ansehen
                  </button>
                </div>
              )
            })
          )}
        </div>
      )}

      {/* Settled Claims */}
      {activeTab === 'settled' && (
        <div className="claims-grid">
          {settledClaims.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">{'\u{1F4CB}'}</div>
              <h3 className="empty-state__title">Keine abgewickelten Schadenfaelle</h3>
              <p className="empty-state__description">Abgewickelte und abgelehnte Schadenfaelle erscheinen hier.</p>
            </div>
          ) : (
            settledClaims.map(claim => (
              <div
                key={claim.id}
                className={`claim-card claim-card--settled ${claim.status === 'rejected' ? 'claim-card--rejected' : ''}`}
                onClick={() => navigate(`/insurance/claims/${claim.id}`)}
              >
                <div className="claim-card__header">
                  <h3 className="claim-card__number">{claim.claim_number}</h3>
                  <span className={`claim-status claim-status--${claim.status}`}>
                    {getClaimStatusLabel(claim.status)}
                  </span>
                </div>

                <div className="claim-card__description">{claim.description}</div>

                <div className="claim-card__amounts">
                  <div className="amount-item">
                    <span className="amount-label">Gefordert:</span>
                    <span className="amount-value">{'\u20AC'}{claim.total_claimed.toLocaleString('de-DE')}</span>
                  </div>
                  {claim.total_settled && (
                    <div className="amount-item">
                      <span className="amount-label">Abgewickelt:</span>
                      <span className="amount-value" style={{ color: 'var(--color-success)' }}>
                        {'\u20AC'}{claim.total_settled.toLocaleString('de-DE')}
                      </span>
                    </div>
                  )}
                  {claim.status === 'rejected' && (
                    <div className="amount-item">
                      <span className="amount-label">Ergebnis:</span>
                      <span className="amount-value" style={{ color: 'var(--color-danger)' }}>
                        Abgelehnt
                      </span>
                    </div>
                  )}
                </div>

                <div className="claim-card__meta">
                  <p>
                    <strong>Schadendatum:</strong> {new Date(claim.incident_date).toLocaleDateString('de-DE')}
                  </p>
                  {claim.assigned_adjuster && (
                    <p>
                      <strong>Sachbearbeiter:</strong> {claim.assigned_adjuster}
                    </p>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* New Claim Modal */}
      {showNewClaimModal && (
        <div className="modal-overlay" onClick={() => !claimSubmitted && setShowNewClaimModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal__header">
              <h2 className="modal__title">Neuer Schadensfall</h2>
              <button
                className="modal__close"
                onClick={() => setShowNewClaimModal(false)}
                disabled={claimSubmitted}
              >
                {'\u2715'}
              </button>
            </div>

            {claimSubmitted ? (
              <div className="modal__success">
                <div className="modal__success-icon">{'\u2713'}</div>
                <h3>Schadensfall erfolgreich gemeldet</h3>
                <p>Ihre Schadenmeldung wird bearbeitet. Sie erhalten in Kuerze eine Bestaetigung.</p>
              </div>
            ) : (
              <div className="modal__body">
                <div className="form-group">
                  <label className="form-label" htmlFor="claim-policy">Versicherungspolice *</label>
                  <select
                    id="claim-policy"
                    className="form-select"
                    value={newClaimForm.policy_id}
                    onChange={(e) => setNewClaimForm({ ...newClaimForm, policy_id: e.target.value })}
                  >
                    <option value="">Police auswaehlen...</option>
                    {policies.filter(p => p.is_active).map(p => (
                      <option key={p.id} value={p.id}>
                        {p.provider} - {getPolicyTypeLabel(p.type)} ({p.policy_number})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="form-group">
                  <label className="form-label" htmlFor="claim-date">Schadendatum *</label>
                  <input
                    id="claim-date"
                    type="date"
                    className="form-input"
                    value={newClaimForm.incident_date}
                    onChange={(e) => setNewClaimForm({ ...newClaimForm, incident_date: e.target.value })}
                    max={new Date().toISOString().split('T')[0]}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label" htmlFor="claim-description">Beschreibung des Schadensfalls *</label>
                  <textarea
                    id="claim-description"
                    className="form-textarea"
                    rows={4}
                    placeholder="Beschreiben Sie den Vorfall detailliert..."
                    value={newClaimForm.description}
                    onChange={(e) => setNewClaimForm({ ...newClaimForm, description: e.target.value })}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label" htmlFor="claim-equipment">Betroffenes Equipment</label>
                  <input
                    id="claim-equipment"
                    type="text"
                    className="form-input"
                    placeholder="z.B. Moving Head Robe T1, Mischpult..."
                    value={newClaimForm.affected_equipment}
                    onChange={(e) => setNewClaimForm({ ...newClaimForm, affected_equipment: e.target.value })}
                  />
                </div>

                <div className="form-group">
                  <label className="form-label" htmlFor="claim-damage">Geschaetzter Schaden ({'\u20AC'})</label>
                  <input
                    id="claim-damage"
                    type="number"
                    className="form-input"
                    placeholder="0,00"
                    min="0"
                    step="0.01"
                    value={newClaimForm.estimated_damage}
                    onChange={(e) => setNewClaimForm({ ...newClaimForm, estimated_damage: e.target.value })}
                  />
                </div>

                {newClaimForm.policy_id && (
                  <div className="form-info">
                    <div className="form-info__title">Selbstbeteiligung</div>
                    <div className="form-info__value">
                      {'\u20AC'}{policies.find(p => p.id === newClaimForm.policy_id)?.deductible.toLocaleString('de-DE') || '0'}
                    </div>
                  </div>
                )}

                <div className="modal__footer">
                  <button
                    className="btn btn--secondary"
                    onClick={() => setShowNewClaimModal(false)}
                  >
                    Abbrechen
                  </button>
                  <button
                    className="btn btn--primary"
                    onClick={handleNewClaimSubmit}
                    disabled={!newClaimForm.policy_id || !newClaimForm.incident_date || !newClaimForm.description}
                  >
                    Schadensfall melden
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default InsurancePage
