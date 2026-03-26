import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { federationApi } from '../../services/api'
import type { Partner, EquipmentAvailability, SubRentalRequest } from '../../types/federation'
import { SkeletonTable, SkeletonCard } from '../../components/Skeleton/SkeletonLoader'
import styles from './Federation.module.scss'


function FederationPage() {
  const [filterTrust, setFilterTrust] = useState<string>('all')
  const [activeTab, setActiveTab] = useState<'partners' | 'equipment' | 'subrental'>('partners')
  const [showConnectModal, setShowConnectModal] = useState(false)
  const [connectStep, setConnectStep] = useState(1)
  const [connectUrl, setConnectUrl] = useState('')
  const [connectStatus, setConnectStatus] = useState<'idle' | 'connecting' | 'success' | 'error'>('idle')
  const [selectedPartner, setSelectedPartner] = useState<Partner | null>(null)
  const [subRentalFilter, setSubRentalFilter] = useState<string>('all')

  const { data: rawPartners = [], isLoading: isLoadingPartners } = useQuery({
    queryKey: ['federation-partners'],
    queryFn: async () => {
      const result = await federationApi.partners()
      return Array.isArray(result) ? result : []
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })
  const partners = rawPartners as Partner[]

  const { data: rawEquipment = [], isLoading: isLoadingEquipment } = useQuery({
    queryKey: ['equipment-availability'],
    queryFn: async () => {
      const result = await federationApi.equipment('all')
      return Array.isArray(result) ? result : []
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })
  const equipmentAvailability = rawEquipment as EquipmentAvailability[]

  const { data: rawSubRentals = [], isLoading: isLoadingSubRentals } = useQuery({
    queryKey: ['sub-rental-requests'],
    queryFn: async () => {
      const result = await federationApi.subRentals()
      return Array.isArray(result) ? result : []
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })
  const subRentalRequests = rawSubRentals as SubRentalRequest[]

  const filteredPartners = filterTrust === 'all'
    ? partners
    : partners.filter(p => p.trust_level === filterTrust)

  const filteredSubRentals = subRentalFilter === 'all'
    ? subRentalRequests
    : subRentalRequests.filter(r => r.status === subRentalFilter)

  const getTrustColor = (level: string): string => {
    switch (level) {
      case 'verified': return '#10b981'
      case 'trusted': return '#3b82f6'
      case 'provisional': return '#f59e0b'
      case 'suspended': return '#ef4444'
      default: return '#00d4ff'
    }
  }

  const getTrustLabel = (level: string): string => {
    switch (level) {
      case 'verified': return 'Verifiziert'
      case 'trusted': return 'Vertrauenswuerdig'
      case 'provisional': return 'Provisorisch'
      case 'suspended': return 'Suspendiert'
      default: return level
    }
  }

  const getPartnerStatusLabel = (level: string): string => {
    switch (level) {
      case 'verified': return 'Aktiv'
      case 'trusted': return 'Aktiv'
      case 'provisional': return 'Ausstehend'
      case 'suspended': return 'Getrennt'
      default: return level
    }
  }

  const getPartnerStatusColor = (level: string): string => {
    switch (level) {
      case 'verified': return '#10b981'
      case 'trusted': return '#10b981'
      case 'provisional': return '#f59e0b'
      case 'suspended': return '#ef4444'
      default: return '#6b7280'
    }
  }

  const getCertificateLabel = (status: string): string => {
    switch (status) {
      case 'valid': return 'Gueltig'
      case 'expiring': return 'Laeuft ab'
      case 'expired': return 'Abgelaufen'
      case 'pending': return 'Ausstehend'
      default: return status
    }
  }

  const getCertificateColor = (status: string): string => {
    switch (status) {
      case 'valid': return '#10b981'
      case 'expiring': return '#f59e0b'
      case 'expired': return '#ef4444'
      case 'pending': return '#6b7280'
      default: return '#00d4ff'
    }
  }

  const getSubRentalStatusLabel = (status: string): string => {
    switch (status) {
      case 'pending': return 'Ausstehend'
      case 'approved': return 'Genehmigt'
      case 'rejected': return 'Abgelehnt'
      case 'completed': return 'Abgeschlossen'
      default: return status
    }
  }

  const getStatusBadgeColor = (status: string): string => {
    switch (status) {
      case 'approved': return '#10b981'
      case 'pending': return '#f59e0b'
      case 'rejected': return '#ef4444'
      case 'completed': return '#3b82f6'
      default: return '#00d4ff'
    }
  }

  const totalEquipmentCount = equipmentAvailability.reduce((sum, e) => sum + e.quantity_available, 0)

  const handleConnect = () => {
    if (connectStep === 1 && connectUrl.trim()) {
      setConnectStep(2)
      setConnectStatus('connecting')
      setTimeout(() => {
        setConnectStatus('success')
        setConnectStep(3)
      }, 2500)
    }
  }

  const resetConnectModal = () => {
    setShowConnectModal(false)
    setConnectStep(1)
    setConnectUrl('')
    setConnectStatus('idle')
  }

  return (
    <div className={styles.federationPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Federation Partner-Netzwerk</h1>
          <p className={styles.subtitle}>Verwalten Sie Ihre Partnerschaften und Equipment-Verfuegbarkeit</p>
        </div>
        <button
          className="btn btn--primary"
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          onClick={() => setShowConnectModal(true)}
        >
          + Partner verbinden
        </button>
      </div>

      {/* Stats Grid */}
      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'🌐'}</div>
          <div className={styles.statLabel}>Partner gesamt</div>
          <div className={styles.statValue}>{partners.length}</div>
          <div className={styles.statSubtext}>{partners.filter(p => p.trust_level !== 'suspended').length} aktiv verbunden</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'\u2713'}</div>
          <div className={styles.statLabel}>Verifizierte Partner</div>
          <div className={styles.statValue} style={{ color: '#10b981' }}>
            {partners.filter(p => p.trust_level === 'verified').length}
          </div>
          <div className={styles.statSubtext}>Hoechste Vertrauensstufe</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'📦'}</div>
          <div className={styles.statLabel}>Verfuegbare Ausruestung</div>
          <div className={styles.statValue}>{totalEquipmentCount}</div>
          <div className={styles.statSubtext}>{equipmentAvailability.length} verschiedene Typen</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statIcon}>{'📋'}</div>
          <div className={styles.statLabel}>Offene Anfragen</div>
          <div className={styles.statValue} style={{ color: '#f59e0b' }}>
            {subRentalRequests.filter(r => r.status === 'pending').length}
          </div>
          <div className={styles.statSubtext}>{subRentalRequests.filter(r => r.status === 'approved').length} genehmigt</div>
        </div>
      </div>

      {/* Tab Bar */}
      <div className={styles.tabBar}>
        <button
          className={`${styles.tabButton} ${activeTab === 'partners' ? styles['tabButton--active'] : ''}`}
          onClick={() => setActiveTab('partners')}
        >
          Partner ({partners.length})
        </button>
        <button
          className={`${styles.tabButton} ${activeTab === 'equipment' ? styles['tabButton--active'] : ''}`}
          onClick={() => setActiveTab('equipment')}
        >
          Equipment-Katalog ({equipmentAvailability.length})
        </button>
        <button
          className={`${styles.tabButton} ${activeTab === 'subrental' ? styles['tabButton--active'] : ''}`}
          onClick={() => setActiveTab('subrental')}
        >
          Sub-Rental ({subRentalRequests.length})
        </button>
      </div>

      {/* Partners Tab */}
      {activeTab === 'partners' && (
        <section className={styles.section}>
          <div className={styles.filterSection}>
            <label htmlFor="trust-filter" className={styles.filterLabel}>Vertrauensstufe:</label>
            <select
              id="trust-filter"
              className={styles.filterSelect}
              value={filterTrust}
              onChange={(e) => setFilterTrust(e.target.value)}
            >
              <option value="all">Alle Partner</option>
              <option value="verified">Verifiziert</option>
              <option value="trusted">Vertrauenswuerdig</option>
              <option value="provisional">Provisorisch</option>
              <option value="suspended">Suspendiert</option>
            </select>
          </div>

          {isLoadingPartners ? (
            <SkeletonTable rows={4} columns={4} />
          ) : filteredPartners.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'🌐'}</div>
              <h3 className={styles.emptyTitle}>Keine Partner gefunden</h3>
              <p className={styles.emptyText}>
                {partners.length === 0
                  ? 'Verbinden Sie sich mit anderen EquipFlow-Firmen um Equipment zu teilen und Ihre Reichweite zu erhoehen.'
                  : 'Keine Partner mit der gewaehlten Vertrauensstufe gefunden.'}
              </p>
              {partners.length === 0 && (
                <button className="btn btn--primary" onClick={() => setShowConnectModal(true)}>
                  Ersten Partner verbinden
                </button>
              )}
            </div>
          ) : (
            <div className={styles.partnersGrid}>
              {filteredPartners.map((partner) => (
                <div key={partner.id} className={styles.partnerCard}>
                  <div className={styles.partnerHeader}>
                    <div>
                      <h3 className={styles.partnerName}>{partner.name}</h3>
                      <p className={styles.partnerCompany}>{partner.company}</p>
                    </div>
                    <div className={styles.partnerHeaderRight}>
                      <div
                        className={styles.partnerStatusBadge}
                        style={{ backgroundColor: `${getPartnerStatusColor(partner.trust_level)}15`, color: getPartnerStatusColor(partner.trust_level), borderColor: getPartnerStatusColor(partner.trust_level) }}
                      >
                        {getPartnerStatusLabel(partner.trust_level)}
                      </div>
                      <div className={styles.ratingBadge}>
                        {'\u2605'} {partner.rating}
                      </div>
                    </div>
                  </div>

                  <div className={styles.trustBadge} style={{ borderColor: getTrustColor(partner.trust_level) }}>
                    <span style={{ color: getTrustColor(partner.trust_level) }}>{'\u25CF'}</span>
                    {getTrustLabel(partner.trust_level)}
                  </div>

                  <div className={styles.partnerMeta}>
                    <p><strong>Land:</strong> {partner.country}</p>
                    <p><strong>E-Mail:</strong> {partner.email}</p>
                    <p><strong>Telefon:</strong> {partner.phone}</p>
                    <p><strong>Equipment verfuegbar:</strong> {partner.equipment_availability} Geraete</p>
                    <p><strong>Beigetreten:</strong> {new Date(partner.joined_date).toLocaleDateString('de-DE')}</p>
                  </div>

                  <div className={styles.certificateSection}>
                    <div className={styles.certificateStatus} style={{ borderColor: getCertificateColor(partner.certificate_status) }}>
                      <span style={{ color: getCertificateColor(partner.certificate_status) }}>{'\u25CF'}</span>
                      Zertifikat: {getCertificateLabel(partner.certificate_status)}
                    </div>
                    <p className={styles.certificateExpiry}>
                      Gueltig bis: {new Date(partner.certificate_expiry).toLocaleDateString('de-DE')}
                    </p>
                  </div>

                  <div className={styles.partnerActions}>
                    <button className="btn btn--sm btn--primary" onClick={() => setSelectedPartner(partner)}>Details</button>
                    <button className="btn btn--sm" onClick={() => alert(`Nachricht an ${partner.name} wird vorbereitet...`)}>Nachricht</button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Equipment Tab */}
      {activeTab === 'equipment' && (
        <section className={styles.section}>
          {isLoadingEquipment ? (
            <SkeletonCard count={3} />
          ) : equipmentAvailability.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'📦'}</div>
              <h3 className={styles.emptyTitle}>Keine Ausruestung verfuegbar</h3>
              <p className={styles.emptyText}>Partner-Equipment wird hier angezeigt sobald Verbindungen bestehen.</p>
            </div>
          ) : (
            <div className={styles.equipmentTable}>
              <div className={styles.tableHeader}>
                <div className={styles.headerCell}>Partner</div>
                <div className={styles.headerCell}>Ausruestung</div>
                <div className={styles.headerCell}>Menge</div>
                <div className={styles.headerCell}>Standort</div>
                <div className={styles.headerCell}>Liefertage</div>
                <div className={styles.headerCell}>Preis/Tag</div>
              </div>
              {equipmentAvailability.map((item, idx) => (
                <div key={idx} className={styles.tableRow}>
                  <div className={styles.cell}>{item.partner_name}</div>
                  <div className={styles.cell}><strong>{item.equipment_type}</strong></div>
                  <div className={styles.cell}>{item.quantity_available}x</div>
                  <div className={styles.cell}>{item.location}</div>
                  <div className={styles.cell}>{item.delivery_days} Tag(e)</div>
                  <div className={styles.cell}><strong>{item.price_per_day}{'\u20AC'}</strong>/Tag</div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Sub-Rental Tab */}
      {activeTab === 'subrental' && (
        <section className={styles.section}>
          <div className={styles.filterSection}>
            <label htmlFor="subrental-filter" className={styles.filterLabel}>Status:</label>
            <select
              id="subrental-filter"
              className={styles.filterSelect}
              value={subRentalFilter}
              onChange={(e) => setSubRentalFilter(e.target.value)}
            >
              <option value="all">Alle Anfragen</option>
              <option value="pending">Ausstehend</option>
              <option value="approved">Genehmigt</option>
              <option value="rejected">Abgelehnt</option>
              <option value="completed">Abgeschlossen</option>
            </select>
          </div>

          {isLoadingSubRentals ? (
            <SkeletonTable rows={4} columns={4} />
          ) : filteredSubRentals.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'📋'}</div>
              <h3 className={styles.emptyTitle}>Keine Anfragen vorhanden</h3>
              <p className={styles.emptyText}>Sub-Rental Anfragen von und an Partner erscheinen hier.</p>
            </div>
          ) : (
            <div className={styles.requestsList}>
              {filteredSubRentals.map((request) => (
                <div key={request.id} className={styles.requestCard}>
                  <div className={styles.requestHeader}>
                    <div>
                      <h4 className={styles.requestTitle}>{request.equipment}</h4>
                      <p className={styles.requestPartner}>{request.partner_name}</p>
                    </div>
                    <div
                      className={styles.requestStatus}
                      style={{ backgroundColor: `${getStatusBadgeColor(request.status)}20`, borderColor: getStatusBadgeColor(request.status) }}
                    >
                      <span style={{ color: getStatusBadgeColor(request.status) }}>{'\u25CF'}</span>
                      {getSubRentalStatusLabel(request.status)}
                    </div>
                  </div>

                  <div className={styles.requestDetails}>
                    <span><strong>Menge:</strong> {request.quantity}x</span>
                    <span><strong>Von:</strong> {new Date(request.start_date).toLocaleDateString('de-DE')}</span>
                    <span><strong>Bis:</strong> {new Date(request.end_date).toLocaleDateString('de-DE')}</span>
                    <span><strong>Angefragt:</strong> {new Date(request.requested_at).toLocaleDateString('de-DE')}</span>
                  </div>

                  {/* Status workflow mini */}
                  <div className={styles.requestWorkflow}>
                    {(['pending', 'approved', 'completed'] as const).map((step) => {
                      const stepOrder = ['pending', 'approved', 'completed']
                      const currentOrder = stepOrder.indexOf(request.status === 'rejected' ? 'pending' : request.status)
                      const thisOrder = stepOrder.indexOf(step)
                      const isComplete = thisOrder <= currentOrder && request.status !== 'rejected'

                      return (
                        <div key={step} className={`${styles.workflowStep} ${isComplete ? styles['workflowStep--complete'] : ''}`}>
                          <div className={styles.workflowDot}></div>
                          <span className={styles.workflowLabel}>{getSubRentalStatusLabel(step)}</span>
                        </div>
                      )
                    })}
                    {request.status === 'rejected' && (
                      <div className={`${styles.workflowStep} ${styles['workflowStep--rejected']}`}>
                        <div className={styles.workflowDot}></div>
                        <span className={styles.workflowLabel}>Abgelehnt</span>
                      </div>
                    )}
                  </div>

                  <div className={styles.requestActions}>
                    {request.status === 'pending' && (
                      <>
                        <button className="btn btn--sm btn--primary" onClick={() => alert('Anfrage genehmigt!')}>Genehmigen</button>
                        <button className="btn btn--sm" onClick={() => alert('Anfrage abgelehnt!')}>Ablehnen</button>
                      </>
                    )}
                    {request.status === 'approved' && (
                      <button className="btn btn--sm btn--primary" onClick={() => alert('Als abgeschlossen markiert!')}>Abschliessen</button>
                    )}
                    {(request.status === 'completed' || request.status === 'rejected') && (
                      <button className="btn btn--sm">Details</button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </section>
      )}

      {/* Partner Detail Modal */}
      {selectedPartner && (
        <div className={styles.modalOverlay} onClick={() => setSelectedPartner(null)}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Partner-Details</h2>
              <button className={styles.modalClose} onClick={() => setSelectedPartner(null)}>{'\u2715'}</button>
            </div>
            <div className={styles.modalBody}>
              <div className={styles.partnerDetailHeader}>
                <h3>{selectedPartner.company}</h3>
                <p>{selectedPartner.name}</p>
                <div className={styles.trustBadge} style={{ borderColor: getTrustColor(selectedPartner.trust_level), marginTop: 'var(--spacing-2)' }}>
                  <span style={{ color: getTrustColor(selectedPartner.trust_level) }}>{'\u25CF'}</span>
                  {getTrustLabel(selectedPartner.trust_level)}
                </div>
              </div>

              <div className={styles.partnerDetailSection}>
                <h4>Kontakt</h4>
                <div className={styles.partnerDetailGrid}>
                  <div><strong>E-Mail:</strong> {selectedPartner.email}</div>
                  <div><strong>Telefon:</strong> {selectedPartner.phone}</div>
                  <div><strong>Land:</strong> {selectedPartner.country}</div>
                  <div><strong>Bewertung:</strong> {'\u2605'} {selectedPartner.rating}/5</div>
                </div>
              </div>

              <div className={styles.partnerDetailSection}>
                <h4>Equipment-Katalog</h4>
                {equipmentAvailability.filter(e => e.partner_id === selectedPartner.id).length > 0 ? (
                  <div className={styles.partnerEquipmentList}>
                    {equipmentAvailability.filter(e => e.partner_id === selectedPartner.id).map((eq, idx) => (
                      <div key={idx} className={styles.partnerEquipmentItem}>
                        <span className={styles.partnerEquipmentName}>{eq.equipment_type}</span>
                        <span className={styles.partnerEquipmentQty}>{eq.quantity_available}x verfuegbar</span>
                        <span className={styles.partnerEquipmentPrice}>{eq.price_per_day}{'\u20AC'}/Tag</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>Kein Equipment im Katalog hinterlegt.</p>
                )}
              </div>

              <div className={styles.partnerDetailSection}>
                <h4>Sub-Rental Anfragen</h4>
                {subRentalRequests.filter(r => r.partner_id === selectedPartner.id).length > 0 ? (
                  <div className={styles.partnerSubRentalList}>
                    {subRentalRequests.filter(r => r.partner_id === selectedPartner.id).map(r => (
                      <div key={r.id} className={styles.partnerSubRentalItem}>
                        <span>{r.equipment} ({r.quantity}x)</span>
                        <span
                          className={styles.partnerSubRentalStatus}
                          style={{ color: getStatusBadgeColor(r.status) }}
                        >
                          {getSubRentalStatusLabel(r.status)}
                        </span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>Keine Sub-Rental Anfragen mit diesem Partner.</p>
                )}
              </div>

              <div className={styles.modalFooter}>
                <button className="btn btn--secondary" onClick={() => setSelectedPartner(null)}>Schliessen</button>
                <button className="btn btn--primary" onClick={() => alert(`Equipment von ${selectedPartner.company} anfragen...`)}>Equipment anfragen</button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Connect Partner Modal (Wizard) */}
      {showConnectModal && (
        <div className={styles.modalOverlay} onClick={() => connectStatus !== 'connecting' && resetConnectModal()}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Partner verbinden</h2>
              <button className={styles.modalClose} onClick={resetConnectModal} disabled={connectStatus === 'connecting'}>{'\u2715'}</button>
            </div>
            <div className={styles.modalBody}>
              {/* Wizard Steps indicator */}
              <div className={styles.wizardSteps}>
                <div className={`${styles.wizardStep} ${connectStep >= 1 ? styles['wizardStep--active'] : ''}`}>
                  <div className={styles.wizardStepDot}>1</div>
                  <span>Partner-URL</span>
                </div>
                <div className={styles.wizardConnector}></div>
                <div className={`${styles.wizardStep} ${connectStep >= 2 ? styles['wizardStep--active'] : ''}`}>
                  <div className={styles.wizardStepDot}>2</div>
                  <span>Verbinden</span>
                </div>
                <div className={styles.wizardConnector}></div>
                <div className={`${styles.wizardStep} ${connectStep >= 3 ? styles['wizardStep--active'] : ''}`}>
                  <div className={styles.wizardStepDot}>3</div>
                  <span>Fertig</span>
                </div>
              </div>

              {connectStep === 1 && (
                <div className={styles.wizardContent}>
                  <h3>Partner-URL eingeben</h3>
                  <p>Geben Sie die EquipFlow-URL des Partners ein, mit dem Sie sich verbinden moechten.</p>
                  <div className={styles.formGroup}>
                    <label className={styles.formLabel}>Partner EquipFlow URL</label>
                    <input
                      type="url"
                      className={styles.formInput}
                      placeholder="https://partner.rentflow.de"
                      value={connectUrl}
                      onChange={(e) => setConnectUrl(e.target.value)}
                    />
                  </div>
                  <div className={styles.modalFooter}>
                    <button className="btn btn--secondary" onClick={resetConnectModal}>Abbrechen</button>
                    <button className="btn btn--primary" onClick={handleConnect} disabled={!connectUrl.trim()}>
                      Verbindung herstellen
                    </button>
                  </div>
                </div>
              )}

              {connectStep === 2 && (
                <div className={styles.wizardContent} style={{ textAlign: 'center' }}>
                  <div className={styles.connectingSpinner}></div>
                  <h3>Verbindung wird hergestellt...</h3>
                  <p>Wir kontaktieren den Partner-Server und tauschen Zertifikate aus.</p>
                  <div className={styles.connectSteps}>
                    <div className={`${styles.connectStepItem} ${styles['connectStepItem--complete']}`}>DNS-Aufloesung...</div>
                    <div className={`${styles.connectStepItem} ${styles['connectStepItem--active']}`}>Zertifikataustausch...</div>
                    <div className={styles.connectStepItem}>Trust-Level festlegen...</div>
                  </div>
                </div>
              )}

              {connectStep === 3 && (
                <div className={styles.wizardContent} style={{ textAlign: 'center' }}>
                  <div className={styles.successIcon}>{'\u2713'}</div>
                  <h3>Verbindung erfolgreich!</h3>
                  <p>Der Partner wurde erfolgreich verbunden. Der Status ist "Provisorisch" bis die Verifizierung abgeschlossen ist.</p>
                  <div className={styles.modalFooter} style={{ justifyContent: 'center' }}>
                    <button className="btn btn--primary" onClick={resetConnectModal}>Fertig</button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default FederationPage
