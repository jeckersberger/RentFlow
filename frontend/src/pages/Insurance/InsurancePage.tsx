import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import type { InsurancePolicy, InsuranceClaim } from '../../types/insurance'
import './Insurance.module.scss'

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
]

const mockClaims: InsuranceClaim[] = [
  {
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
  {
    id: 'c2',
    claim_number: 'SCH-2026-002',
    policy_id: '2',
    status: 'submitted',
    reported_date: '2026-03-20',
    incident_date: '2026-03-19',
    description: 'Diebstahl von Ausrüstung von der Baustelle',
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
        description: 'Diverse Kleinteile und Zubehör',
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
        description: 'Fahzeugschadensatz und Reparatur',
        claimed_amount: 12000,
        approved_amount: 11500,
        cost_category: 'vehicle',
      },
    ],
    photos: [],
    assigned_adjuster: 'Klaus Schmidt',
  },
]

function InsurancePage() {
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState<'policies' | 'open_claims' | 'settled'>('policies')

  const { data: policies = mockPolicies } = useQuery({
    queryKey: ['insurance-policies'],
    queryFn: async () => mockPolicies,
    staleTime: 1000 * 60 * 5,
  })

  const { data: claims = mockClaims } = useQuery({
    queryKey: ['insurance-claims'],
    queryFn: async () => mockClaims,
    staleTime: 1000 * 60 * 5,
  })

  const openClaims = claims.filter(c => ['reported', 'documented', 'submitted', 'approved'].includes(c.status))
  const settledClaims = claims.filter(c => ['settled', 'rejected'].includes(c.status))
  const totalPremiums = policies.reduce((sum, p) => sum + p.premium_annual, 0)

  const getPolicyTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      liability: 'Haftung',
      equipment: 'Ausrüstung',
      vehicle: 'Fahrzeuge',
      workers_comp: 'Unfallversicherung',
    }
    return labels[type] || type
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

  return (
    <div className="insurance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Versicherung</h1>
          <p className="page-subtitle">Verwaltung von Versicherungspolicen und Schadenfällen</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => alert('Neuer Schadensfall wird eingeleitet...')}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          ➕ Neuer Schadensfall
        </button>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Aktive Policen</div>
          <div className="stat-card__value">{policies.filter(p => p.is_active).length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Offene Schadenfälle</div>
          <div className="stat-card__value" style={{ color: 'var(--color-warning)' }}>{openClaims.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Schadenfälle Jahr</div>
          <div className="stat-card__value">{claims.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Prämien pro Jahr</div>
          <div className="stat-card__value" style={{ fontSize: 'var(--font-size-2xl)' }}>
            €{totalPremiums.toLocaleString('de-DE')}
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
          Offene Schadenfälle ({openClaims.length})
        </button>
        <button
          className={`tab-button ${activeTab === 'settled' ? 'tab-button--active' : ''}`}
          onClick={() => setActiveTab('settled')}
        >
          Abgewickelte Schadenfälle ({settledClaims.length})
        </button>
      </div>

      {/* Active Policies */}
      {activeTab === 'policies' && (
        <div className="policies-grid">
          {policies.map(policy => (
            <div key={policy.id} className="policy-card">
              <div className="policy-card__header">
                <h3 className="policy-card__title">{policy.provider}</h3>
                <span className={`policy-status ${policy.is_active ? 'policy-status--active' : 'policy-status--inactive'}`}>
                  {policy.is_active ? '✓ Aktiv' : 'Inaktiv'}
                </span>
              </div>

              <div className="policy-card__type">
                <span className="policy-type-badge">{getPolicyTypeLabel(policy.type)}</span>
              </div>

              <div className="policy-card__details">
                <div className="detail-row">
                  <span className="detail-label">Police Nr.:</span>
                  <span className="detail-value">{policy.policy_number}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Deckung:</span>
                  <span className="detail-value">€{policy.coverage_amount.toLocaleString('de-DE')}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Prämie/Jahr:</span>
                  <span className="detail-value">€{policy.premium_annual.toLocaleString('de-DE')}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Selbstbeteiligung:</span>
                  <span className="detail-value">€{policy.deductible.toLocaleString('de-DE')}</span>
                </div>
                <div className="detail-row">
                  <span className="detail-label">Gültig bis:</span>
                  <span className="detail-value">{new Date(policy.end_date).toLocaleDateString('de-DE')}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Open Claims */}
      {activeTab === 'open_claims' && (
        <div className="claims-grid">
          {openClaims.length === 0 ? (
            <div className="empty-state">
              <div className="empty-state__icon">✓</div>
              <h3 className="empty-state__title">Keine offenen Schadenfälle</h3>
              <p className="empty-state__description">Das ist eine gute Nachricht!</p>
            </div>
          ) : (
            openClaims.map(claim => (
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

                <div className="claim-card__description">{claim.description}</div>

                <div className="claim-card__amounts">
                  <div className="amount-item">
                    <span className="amount-label">Gefordert:</span>
                    <span className="amount-value">€{claim.total_claimed.toLocaleString('de-DE')}</span>
                  </div>
                  {claim.total_approved && (
                    <div className="amount-item">
                      <span className="amount-label">Genehmigt:</span>
                      <span className="amount-value" style={{ color: 'var(--color-success)' }}>
                        €{claim.total_approved.toLocaleString('de-DE')}
                      </span>
                    </div>
                  )}
                </div>

                <div className="claim-card__meta">
                  <p>
                    <strong>Datum:</strong> {new Date(claim.incident_date).toLocaleDateString('de-DE')}
                  </p>
                  {claim.assigned_adjuster && (
                    <p>
                      <strong>Sachbearbeiter:</strong> {claim.assigned_adjuster}
                    </p>
                  )}
                </div>

                <button
                  className="btn btn--sm btn--primary"
                  onClick={(e) => {
                    e.stopPropagation()
                    navigate(`/insurance/claims/${claim.id}`)
                  }}
                  style={{ width: '100%', marginTop: 'var(--spacing-3)' }}
                >
                  Details Ansehen
                </button>
              </div>
            ))
          )}
        </div>
      )}

      {/* Settled Claims */}
      {activeTab === 'settled' && (
        <div className="claims-grid">
          {settledClaims.map(claim => (
            <div
              key={claim.id}
              className="claim-card claim-card--settled"
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
                  <span className="amount-value">€{claim.total_claimed.toLocaleString('de-DE')}</span>
                </div>
                {claim.total_settled && (
                  <div className="amount-item">
                    <span className="amount-label">Abgewickelt:</span>
                    <span className="amount-value" style={{ color: 'var(--color-success)' }}>
                      €{claim.total_settled.toLocaleString('de-DE')}
                    </span>
                  </div>
                )}
              </div>

              <div className="claim-card__meta">
                <p>
                  <strong>Datum:</strong> {new Date(claim.incident_date).toLocaleDateString('de-DE')}
                </p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default InsurancePage
