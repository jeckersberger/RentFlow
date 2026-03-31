import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { insuranceApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import './Insurance.scss'

// Backend response types (matching Go DTOs)
interface PolicyResponse {
  id: string
  tenant_id: string
  policy_number: string
  policy_type: string
  provider: string
  coverage_amount: number
  deductible: number
  premium_annual: number
  premium_monthly: number
  start_date: string
  end_date: string
  status: string
  notes?: string
  created_at: string
  updated_at: string
}

interface ClaimResponse {
  id: string
  tenant_id: string
  policy_id: string
  claim_number: string
  equipment_id?: string
  project_id?: string
  incident_date: string
  reported_date: string
  description: string
  damage_type: string
  status: string
  claimed_amount: number
  approved_amount?: number
  settled_amount?: number
  adjuster_notes?: string
  created_by: string
  created_at: string
  updated_at: string
}

// Adapter: map backend policy to a common shape for the UI
interface UIPolicy {
  id: string
  policy_number: string
  type: string
  provider: string
  coverage_amount: number
  premium_annual: number
  start_date: string
  end_date: string
  is_active: boolean
  deductible: number
}

interface UIClaim {
  id: string
  claim_number: string
  policy_id: string
  status: string
  reported_date: string
  incident_date: string
  description: string
  damage_type: string
  total_claimed: number
  total_approved?: number
  total_settled?: number
  adjuster_notes?: string
}

function mapPolicy(p: PolicyResponse): UIPolicy {
  return {
    id: p.id,
    policy_number: p.policy_number,
    type: p.policy_type,
    provider: p.provider,
    coverage_amount: p.coverage_amount,
    premium_annual: p.premium_annual,
    start_date: p.start_date,
    end_date: p.end_date,
    is_active: p.status === 'active',
    deductible: p.deductible,
  }
}

function mapClaim(c: ClaimResponse): UIClaim {
  return {
    id: c.id,
    claim_number: c.claim_number,
    policy_id: c.policy_id,
    status: c.status,
    reported_date: c.reported_date,
    incident_date: c.incident_date,
    description: c.description,
    damage_type: c.damage_type,
    total_claimed: c.claimed_amount,
    total_approved: c.approved_amount,
    total_settled: c.settled_amount,
    adjuster_notes: c.adjuster_notes,
  }
}

interface NewPolicyFormData {
  policy_number: string
  policy_type: string
  provider: string
  coverage_amount: string
  deductible: string
  premium_annual: string
  start_date: string
  end_date: string
  notes: string
}

interface NewClaimFormData {
  policy_id: string
  incident_date: string
  description: string
  damage_type: string
  claimed_amount: string
}

const emptyPolicyForm: NewPolicyFormData = {
  policy_number: '',
  policy_type: 'liability',
  provider: '',
  coverage_amount: '',
  deductible: '',
  premium_annual: '',
  start_date: '',
  end_date: '',
  notes: '',
}

const emptyClaimForm: NewClaimFormData = {
  policy_id: '',
  incident_date: '',
  description: '',
  damage_type: 'damage',
  claimed_amount: '',
}

function InsurancePage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [activeTab, setActiveTab] = useState<'policies' | 'open_claims' | 'settled'>('policies')
  const [showNewPolicyModal, setShowNewPolicyModal] = useState(false)
  const [showNewClaimModal, setShowNewClaimModal] = useState(false)
  const [policyForm, setPolicyForm] = useState<NewPolicyFormData>(emptyPolicyForm)
  const [claimForm, setClaimForm] = useState<NewClaimFormData>(emptyClaimForm)

  const { data: policiesRaw, isLoading: policiesLoading, error: policiesError } = useQuery({
    queryKey: ['insurance-policies'],
    queryFn: () => insuranceApi.listPolicies(),
    staleTime: 1000 * 60 * 5,
    retry: 1,
  })

  const { data: claimsRaw, isLoading: claimsLoading, error: claimsError } = useQuery({
    queryKey: ['insurance-claims'],
    queryFn: () => insuranceApi.listClaims(),
    staleTime: 1000 * 60 * 5,
    retry: 1,
  })

  const createPolicyMutation = useMutation({
    mutationFn: (data: object) => insuranceApi.createPolicy(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['insurance-policies'] })
      setShowNewPolicyModal(false)
      setPolicyForm(emptyPolicyForm)
      addNotification('Die Police wurde erfolgreich angelegt.', 'success', { title: 'Police erstellt' })
    },
    onError: (err: any) => {
      addNotification(`Police konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`, 'error', { title: 'Fehler' })
    },
  })

  const createClaimMutation = useMutation({
    mutationFn: (data: object) => insuranceApi.createClaim(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['insurance-claims'] })
      setShowNewClaimModal(false)
      setClaimForm(emptyClaimForm)
      addNotification('Der Schadensfall wurde erfolgreich gemeldet.', 'success', { title: 'Schadensfall erstellt' })
    },
    onError: (err: any) => {
      addNotification(`Schadensfall konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`, 'error', { title: 'Fehler' })
    },
  })

  // Normalize API responses to arrays
  const policiesArray: PolicyResponse[] = Array.isArray(policiesRaw) ? policiesRaw : (policiesRaw?.data || policiesRaw?.items || [])
  const claimsArray: ClaimResponse[] = Array.isArray(claimsRaw) ? claimsRaw : (claimsRaw?.data || claimsRaw?.items || [])

  const policies: UIPolicy[] = policiesArray.map(mapPolicy)
  const claims: UIClaim[] = claimsArray.map(mapClaim)

  const isLoading = (policiesLoading && !policiesError) || (claimsLoading && !claimsError)

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

  const getDaysUntilExpiry = (endDate: string): number => {
    const end = new Date(endDate)
    const now = new Date()
    return Math.ceil((end.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
  }

  const handleNewPolicySubmit = () => {
    if (!policyForm.policy_number.trim() || !policyForm.provider.trim() || !policyForm.start_date || !policyForm.end_date) {
      addNotification('Policennummer, Versicherer, Start- und Enddatum sind Pflichtfelder.', 'error', { title: 'Fehler' })
      return
    }
    createPolicyMutation.mutate({
      policy_number: policyForm.policy_number,
      policy_type: policyForm.policy_type,
      provider: policyForm.provider,
      coverage_amount: policyForm.coverage_amount ? parseFloat(policyForm.coverage_amount) : 0,
      deductible: policyForm.deductible ? parseFloat(policyForm.deductible) : 0,
      premium_annual: policyForm.premium_annual ? parseFloat(policyForm.premium_annual) : 0,
      premium_monthly: policyForm.premium_annual ? parseFloat(policyForm.premium_annual) / 12 : 0,
      start_date: policyForm.start_date,
      end_date: policyForm.end_date,
      notes: policyForm.notes || undefined,
    })
  }

  const handleNewClaimSubmit = () => {
    if (!claimForm.policy_id || !claimForm.incident_date || !claimForm.description) {
      addNotification('Police, Schadendatum und Beschreibung sind Pflichtfelder.', 'error', { title: 'Fehler' })
      return
    }
    createClaimMutation.mutate({
      policy_id: claimForm.policy_id,
      incident_date: claimForm.incident_date,
      description: claimForm.description,
      damage_type: claimForm.damage_type,
      claimed_amount: claimForm.claimed_amount ? parseFloat(claimForm.claimed_amount) : 0,
    })
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

  if (policiesError && claimsError) {
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
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--primary"
            onClick={() => setShowNewPolicyModal(true)}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            + Neue Police
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => setShowNewClaimModal(true)}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            + Neuer Schadensfall
          </button>
        </div>
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
        <>
          {policies.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">{'\u{1F6E1}'}</div>
              <h3 className="empty-state__title">Keine Policen vorhanden</h3>
              <p className="empty-state__description">Legen Sie Ihre erste Versicherungspolice an, um loszulegen.</p>
              <button className="btn btn--primary" onClick={() => setShowNewPolicyModal(true)}>
                + Neue Police
              </button>
            </div>
          ) : (
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
                        onClick={() => {
                          setClaimForm({ ...emptyClaimForm, policy_id: policy.id })
                          setShowNewClaimModal(true)
                        }}
                      >
                        Schadensfall melden
                      </button>
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </>
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
                    {claim.total_approved != null && (
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
                  {claim.total_settled != null && (
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
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* New Policy Modal */}
      {showNewPolicyModal && (
        <div className="modal-overlay" onClick={() => setShowNewPolicyModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal__header">
              <h2 className="modal__title">Neue Versicherungspolice</h2>
              <button
                className="modal__close"
                onClick={() => setShowNewPolicyModal(false)}
              >
                {'\u2715'}
              </button>
            </div>

            <div className="modal__body">
              <div className="form-group">
                <label className="form-label" htmlFor="policy-number">Policennummer *</label>
                <input
                  id="policy-number"
                  type="text"
                  className="form-input"
                  placeholder="z.B. LV-2026-001"
                  value={policyForm.policy_number}
                  onChange={(e) => setPolicyForm({ ...policyForm, policy_number: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-type">Typ</label>
                <select
                  id="policy-type"
                  className="form-select"
                  value={policyForm.policy_type}
                  onChange={(e) => setPolicyForm({ ...policyForm, policy_type: e.target.value })}
                >
                  <option value="liability">Haftpflicht</option>
                  <option value="equipment">Ausruestung</option>
                  <option value="vehicle">Fahrzeuge</option>
                  <option value="workers_comp">Unfallversicherung</option>
                </select>
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-provider">Versicherer *</label>
                <input
                  id="policy-provider"
                  type="text"
                  className="form-input"
                  placeholder="z.B. Allianz Versicherung"
                  value={policyForm.provider}
                  onChange={(e) => setPolicyForm({ ...policyForm, provider: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-coverage">Deckungssumme ({'\u20AC'})</label>
                <input
                  id="policy-coverage"
                  type="number"
                  className="form-input"
                  placeholder="0"
                  min="0"
                  value={policyForm.coverage_amount}
                  onChange={(e) => setPolicyForm({ ...policyForm, coverage_amount: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-deductible">Selbstbeteiligung ({'\u20AC'})</label>
                <input
                  id="policy-deductible"
                  type="number"
                  className="form-input"
                  placeholder="0"
                  min="0"
                  value={policyForm.deductible}
                  onChange={(e) => setPolicyForm({ ...policyForm, deductible: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-premium">Jahrespraemie ({'\u20AC'})</label>
                <input
                  id="policy-premium"
                  type="number"
                  className="form-input"
                  placeholder="0"
                  min="0"
                  value={policyForm.premium_annual}
                  onChange={(e) => setPolicyForm({ ...policyForm, premium_annual: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-start">Startdatum *</label>
                <input
                  id="policy-start"
                  type="date"
                  className="form-input"
                  value={policyForm.start_date}
                  onChange={(e) => setPolicyForm({ ...policyForm, start_date: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-end">Enddatum *</label>
                <input
                  id="policy-end"
                  type="date"
                  className="form-input"
                  value={policyForm.end_date}
                  onChange={(e) => setPolicyForm({ ...policyForm, end_date: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="policy-notes">Notizen</label>
                <textarea
                  id="policy-notes"
                  className="form-textarea"
                  rows={3}
                  placeholder="Optionale Notizen..."
                  value={policyForm.notes}
                  onChange={(e) => setPolicyForm({ ...policyForm, notes: e.target.value })}
                />
              </div>

              <div className="modal__footer">
                <button
                  className="btn btn--secondary"
                  onClick={() => { setShowNewPolicyModal(false); setPolicyForm(emptyPolicyForm) }}
                >
                  Abbrechen
                </button>
                <button
                  className="btn btn--primary"
                  onClick={handleNewPolicySubmit}
                  disabled={createPolicyMutation.isPending}
                >
                  {createPolicyMutation.isPending ? 'Speichern...' : 'Police anlegen'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* New Claim Modal */}
      {showNewClaimModal && (
        <div className="modal-overlay" onClick={() => setShowNewClaimModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal__header">
              <h2 className="modal__title">Neuer Schadensfall</h2>
              <button
                className="modal__close"
                onClick={() => setShowNewClaimModal(false)}
              >
                {'\u2715'}
              </button>
            </div>

            <div className="modal__body">
              <div className="form-group">
                <label className="form-label" htmlFor="claim-policy">Versicherungspolice *</label>
                <select
                  id="claim-policy"
                  className="form-select"
                  value={claimForm.policy_id}
                  onChange={(e) => setClaimForm({ ...claimForm, policy_id: e.target.value })}
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
                  value={claimForm.incident_date}
                  onChange={(e) => setClaimForm({ ...claimForm, incident_date: e.target.value })}
                  max={new Date().toISOString().split('T')[0]}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="claim-damage-type">Schadensart</label>
                <select
                  id="claim-damage-type"
                  className="form-select"
                  value={claimForm.damage_type}
                  onChange={(e) => setClaimForm({ ...claimForm, damage_type: e.target.value })}
                >
                  <option value="damage">Beschaedigung</option>
                  <option value="theft">Diebstahl</option>
                  <option value="loss">Verlust</option>
                  <option value="accident">Unfall</option>
                  <option value="water_damage">Wasserschaden</option>
                  <option value="fire">Brandschaden</option>
                  <option value="other">Sonstiges</option>
                </select>
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="claim-description">Beschreibung des Schadensfalls *</label>
                <textarea
                  id="claim-description"
                  className="form-textarea"
                  rows={4}
                  placeholder="Beschreiben Sie den Vorfall detailliert..."
                  value={claimForm.description}
                  onChange={(e) => setClaimForm({ ...claimForm, description: e.target.value })}
                />
              </div>

              <div className="form-group">
                <label className="form-label" htmlFor="claim-amount">Geschaetzter Schaden ({'\u20AC'})</label>
                <input
                  id="claim-amount"
                  type="number"
                  className="form-input"
                  placeholder="0,00"
                  min="0"
                  step="0.01"
                  value={claimForm.claimed_amount}
                  onChange={(e) => setClaimForm({ ...claimForm, claimed_amount: e.target.value })}
                />
              </div>

              {claimForm.policy_id && (
                <div className="form-info">
                  <div className="form-info__title">Selbstbeteiligung</div>
                  <div className="form-info__value">
                    {'\u20AC'}{policies.find(p => p.id === claimForm.policy_id)?.deductible.toLocaleString('de-DE') || '0'}
                  </div>
                </div>
              )}

              <div className="modal__footer">
                <button
                  className="btn btn--secondary"
                  onClick={() => { setShowNewClaimModal(false); setClaimForm(emptyClaimForm) }}
                >
                  Abbrechen
                </button>
                <button
                  className="btn btn--primary"
                  onClick={handleNewClaimSubmit}
                  disabled={!claimForm.policy_id || !claimForm.incident_date || !claimForm.description || createClaimMutation.isPending}
                >
                  {createClaimMutation.isPending ? 'Melden...' : 'Schadensfall melden'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default InsurancePage
