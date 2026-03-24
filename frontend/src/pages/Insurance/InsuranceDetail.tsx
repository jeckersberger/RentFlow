import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { insuranceApi } from '../../services/api'
import './Insurance.scss'

function InsuranceDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: claimRaw, isLoading: claimLoading, error: claimError } = useQuery({
    queryKey: ['claim', id],
    queryFn: () => insuranceApi.getClaim(id!),
    enabled: !!id,
    retry: 1,
  })

  // The backend returns the claim directly (no wrapper)
  const claim = claimRaw

  // Fetch the policy for this claim
  const { data: policyRaw } = useQuery({
    queryKey: ['policy', claim?.policy_id],
    queryFn: () => insuranceApi.getPolicy(claim!.policy_id),
    enabled: !!claim?.policy_id,
    retry: 1,
  })

  const policy = policyRaw

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

  const getPolicyTypeLabel = (type: string): string => {
    const labels: Record<string, string> = {
      liability: 'Haftpflicht',
      equipment: 'Ausruestung',
      vehicle: 'Fahrzeuge',
      workers_comp: 'Unfallversicherung',
    }
    return labels[type] || type
  }

  if (claimLoading) {
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

  if (claimError || !claim) {
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
          <p className="empty-state__description">
            {claimError ? String(claimError) : 'Der angeforderte Schadensfall existiert nicht oder wurde geloescht.'}
          </p>
        </div>
      </div>
    )
  }

  // Map backend fields to UI-friendly names
  const claimStatus = claim.status || 'reported'
  const claimNumber = claim.claim_number || `SCH-${claim.id?.substring(0, 8)}`
  const claimedAmount = claim.claimed_amount || 0
  const approvedAmount = claim.approved_amount
  const settledAmount = claim.settled_amount
  const description = claim.description || ''
  const incidentDate = claim.incident_date || ''
  const reportedDate = claim.reported_date || ''

  // Build a simple status timeline from the claim state
  const allSteps = ['reported', 'submitted', 'approved', 'settled']
  const statusOrder = allSteps.indexOf(claimStatus)

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
          <h1 className="page-title">Schadensfall {claimNumber}</h1>
          <p className="page-subtitle">{description}</p>
        </div>
        <span className={`claim-status claim-status--${claimStatus}`} style={{ fontSize: 'var(--font-size-sm)', padding: 'var(--spacing-2) var(--spacing-4)' }}>
          {getStatusLabel(claimStatus)}
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
                const isCompleted = idx <= statusOrder && claimStatus !== 'rejected'
                const isCurrent = step === claimStatus

                return (
                  <div key={step} className={`status-progress__step ${isCompleted ? 'status-progress__step--completed' : ''} ${isCurrent ? 'status-progress__step--current' : ''}`}>
                    <div className="status-progress__dot">
                      {isCompleted ? '\u2713' : (idx + 1)}
                    </div>
                    <span className="status-progress__label">{getStatusLabel(step)}</span>
                    {idx < allSteps.length - 1 && (
                      <div className={`status-progress__connector ${isCompleted && idx < statusOrder ? 'status-progress__connector--active' : ''}`}></div>
                    )}
                  </div>
                )
              })}
              {claimStatus === 'rejected' && (
                <div className="status-progress__step status-progress__step--rejected">
                  <div className="status-progress__dot">{'\u2715'}</div>
                  <span className="status-progress__label">Abgelehnt</span>
                </div>
              )}
            </div>
          </div>

          {/* Claim Details */}
          <div className="detail-card">
            <h2 className="detail-card__title">Schadensdetails</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <label className="detail-card__row-label">Beschreibung</label>
                <span className="detail-card__row-value">{description}</span>
              </div>
              {claim.damage_type && (
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Schadensart</label>
                  <span className="detail-card__row-value">{claim.damage_type}</span>
                </div>
              )}
              {claim.adjuster_notes && (
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Sachbearbeiter-Notizen</label>
                  <span className="detail-card__row-value">{claim.adjuster_notes}</span>
                </div>
              )}
            </div>
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
                <span className="detail-card__row-value">{claimNumber}</span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Status</label>
                <span className="detail-card__row-value">
                  <span className={`claim-status claim-status--${claimStatus}`}>
                    {getStatusLabel(claimStatus)}
                  </span>
                </span>
              </div>
              {incidentDate && (
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Schadendatum</label>
                  <span className="detail-card__row-value">
                    {new Date(incidentDate).toLocaleDateString('de-DE')}
                  </span>
                </div>
              )}
              {reportedDate && (
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Meldedatum</label>
                  <span className="detail-card__row-value">
                    {new Date(reportedDate).toLocaleDateString('de-DE')}
                  </span>
                </div>
              )}
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
                  <span className="detail-card__row-value">{getPolicyTypeLabel(policy.policy_type)}</span>
                </div>
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Deckung</label>
                  <span className="detail-card__row-value">{'\u20AC'}{policy.coverage_amount?.toLocaleString('de-DE')}</span>
                </div>
                <div className="detail-card__row">
                  <label className="detail-card__row-label">Selbstbeteiligung</label>
                  <span className="detail-card__row-value">{'\u20AC'}{policy.deductible?.toLocaleString('de-DE')}</span>
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
                <span className="settlement-value">{'\u20AC'}{claimedAmount.toLocaleString('de-DE')}</span>
              </div>
              {approvedAmount != null && (
                <div className="settlement-row">
                  <span className="settlement-label">Genehmigt:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-success)' }}>
                    {'\u20AC'}{approvedAmount.toLocaleString('de-DE')}
                  </span>
                </div>
              )}
              {settledAmount != null && (
                <div className="settlement-row">
                  <span className="settlement-label">Abgewickelt:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-success)' }}>
                    {'\u20AC'}{settledAmount.toLocaleString('de-DE')}
                  </span>
                </div>
              )}
              {claimStatus === 'rejected' && (
                <div className="settlement-row">
                  <span className="settlement-label">Ergebnis:</span>
                  <span className="settlement-value" style={{ color: 'var(--color-danger)' }}>
                    Abgelehnt
                  </span>
                </div>
              )}

              {approvedAmount != null && claimedAmount > 0 && (
                <div className="settlement-progress">
                  <div className="progress-label">
                    <span>Genehmigungsquote</span>
                    <span>{Math.round((approvedAmount / claimedAmount) * 100)}%</span>
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress-bar__fill"
                      style={{ width: `${(approvedAmount / claimedAmount) * 100}%` }}
                    ></div>
                  </div>
                </div>
              )}

              {settledAmount != null && approvedAmount != null && approvedAmount > 0 && (
                <div className="settlement-progress">
                  <div className="progress-label">
                    <span>Auszahlungsquote</span>
                    <span>{Math.round((settledAmount / approvedAmount) * 100)}%</span>
                  </div>
                  <div className="progress-bar">
                    <div
                      className="progress-bar__fill"
                      style={{ width: `${(settledAmount / approvedAmount) * 100}%` }}
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
                onClick={() => navigate('/insurance')}
              >
                Zurueck zur Uebersicht
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default InsuranceDetail
