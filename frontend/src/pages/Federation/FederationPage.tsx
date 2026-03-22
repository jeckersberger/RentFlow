import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { federationApi } from '../../services/api'
import type { Partner, EquipmentAvailability, SubRentalRequest } from '../../types/federation'
import styles from './Federation.module.scss'

function FederationPage() {
  const [filterTrust, setFilterTrust] = useState<string>('all')

  // Fetch federation partners
  const { data: partners = [], isLoading: isLoadingPartners } = useQuery({
    queryKey: ['federation-partners'],
    queryFn: () => federationApi.partners(),
    staleTime: 1000 * 60 * 5,
  })

  // Fetch equipment availability
  const { data: equipmentAvailability = [], isLoading: isLoadingEquipment } = useQuery({
    queryKey: ['equipment-availability'],
    queryFn: () => federationApi.equipment('all'),
    staleTime: 1000 * 60 * 5,
  })

  // Fetch sub-rental requests
  const { data: subRentalRequests = [], isLoading: isLoadingSubRentals } = useQuery({
    queryKey: ['sub-rental-requests'],
    queryFn: () => federationApi.subRentals(),
    staleTime: 1000 * 60 * 5,
  })

  const filteredPartners = filterTrust === 'all'
    ? partners
    : (partners as Partner[]).filter(p => p.trust_level === filterTrust)

  const getTrustColor = (level: string): string => {
    switch (level) {
      case 'verified':
        return '#10b981'
      case 'trusted':
        return '#3b82f6'
      case 'provisional':
        return '#f59e0b'
      case 'suspended':
        return '#ef4444'
      default:
        return '#00d4ff'
    }
  }

  const getTrustLabel = (level: string): string => {
    switch (level) {
      case 'verified':
        return 'Verifiziert'
      case 'trusted':
        return 'Vertrauenswürdig'
      case 'provisional':
        return 'Provisorisch'
      case 'suspended':
        return 'Suspendiert'
      default:
        return level
    }
  }

  const getCertificateColor = (status: string): string => {
    switch (status) {
      case 'valid':
        return '#10b981'
      case 'expiring':
        return '#f59e0b'
      case 'expired':
        return '#ef4444'
      case 'pending':
        return '#6b7280'
      default:
        return '#00d4ff'
    }
  }

  const getStatusBadgeColor = (status: string): string => {
    switch (status) {
      case 'approved':
        return '#10b981'
      case 'pending':
        return '#f59e0b'
      case 'rejected':
        return '#ef4444'
      case 'completed':
        return '#3b82f6'
      default:
        return '#00d4ff'
    }
  }

  const totalEquipmentCount = (equipmentAvailability as EquipmentAvailability[]).reduce((sum, e) => sum + e.quantity_available, 0)

  return (
    <div className={styles.federationPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>🌐 Federation Partner-Netzwerk</h1>
          <p className={styles.subtitle}>Verwalten Sie Ihre Partnerschaft und Ausrüstungsverfügbarkeit</p>
        </div>
        <button className="btn btn--primary" style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}>
          ➕ Partner hinzufügen
        </button>
      </div>

      {/* Stats Grid */}
      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Partner gesamt</div>
          <div className={styles.statValue}>{partners.length}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Verifizierte Partner</div>
          <div className={styles.statValue} style={{ color: '#10b981' }}>
            {(partners as Partner[]).filter((p: Partner) => p.trust_level === 'verified').length}
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Verfügbare Ausrüstung</div>
          <div className={styles.statValue}>{totalEquipmentCount}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Anfragen ausstehend</div>
          <div className={styles.statValue} style={{ color: '#f59e0b' }}>
            {(subRentalRequests as SubRentalRequest[]).filter((r: SubRentalRequest) => r.status === 'pending').length}
          </div>
        </div>
      </div>

      {/* Filter */}
      <div className={styles.filterSection}>
        <label htmlFor="trust-filter" className={styles.filterLabel}>Trust Level filtern:</label>
        <select
          id="trust-filter"
          className={styles.filterSelect}
          value={filterTrust}
          onChange={(e) => setFilterTrust(e.target.value)}
        >
          <option value="all">Alle Partner</option>
          <option value="verified">Verifiziert</option>
          <option value="trusted">Vertrauenswürdig</option>
          <option value="provisional">Provisorisch</option>
          <option value="suspended">Suspendiert</option>
        </select>
      </div>

      {/* Partner Network */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Partner-Übersicht</h2>

        {isLoadingPartners ? (
          <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
            Laden...
          </div>
        ) : filteredPartners.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
            Keine Partner gefunden
          </div>
        ) : (
          <div className={styles.partnersGrid}>
            {filteredPartners.map((partner: Partner) => (
            <div key={partner.id} className={styles.partnerCard}>
              <div className={styles.partnerHeader}>
                <div>
                  <h3 className={styles.partnerName}>{partner.name}</h3>
                  <p className={styles.partnerCompany}>{partner.company}</p>
                </div>
                <div className={styles.ratingBadge}>
                  ⭐ {partner.rating}
                </div>
              </div>

              <div className={styles.trustBadge} style={{ borderColor: getTrustColor(partner.trust_level) }}>
                <span style={{ color: getTrustColor(partner.trust_level) }}>●</span>
                {getTrustLabel(partner.trust_level)}
              </div>

              <div className={styles.partnerMeta}>
                <p><strong>Land:</strong> {partner.country}</p>
                <p><strong>E-Mail:</strong> {partner.email}</p>
                <p><strong>Telefon:</strong> {partner.phone}</p>
                <p><strong>Ausrüstung verfügbar:</strong> {partner.equipment_availability}</p>
              </div>

              <div className={styles.certificateSection}>
                <div className={styles.certificateStatus} style={{ borderColor: getCertificateColor(partner.certificate_status) }}>
                  <span style={{ color: getCertificateColor(partner.certificate_status) }}>●</span>
                  {partner.certificate_status}
                </div>
                <p className={styles.certificateExpiry}>
                  Gültig bis: {new Date(partner.certificate_expiry).toLocaleDateString('de-DE')}
                </p>
              </div>

              <div className={styles.partnerActions}>
                <button className="btn btn--sm btn--primary">Details</button>
                <button className="btn btn--sm">Nachricht</button>
              </div>
            </div>
            ))}
          </div>
        )}
      </section>

      <div className={styles.bottomGrid}>
        {/* Equipment Availability */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Verfügbare Ausrüstung</h2>

          {isLoadingEquipment ? (
            <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
              Laden...
            </div>
          ) : equipmentAvailability.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
              Keine Ausrüstung verfügbar
            </div>
          ) : (
            <div className={styles.equipmentTable}>
              <div className={styles.tableHeader}>
                <div className={styles.headerCell}>Partner</div>
                <div className={styles.headerCell}>Ausrüstung</div>
                <div className={styles.headerCell}>Menge</div>
                <div className={styles.headerCell}>Standort</div>
                <div className={styles.headerCell}>Liefertage</div>
                <div className={styles.headerCell}>Preis/Tag</div>
              </div>

              {(equipmentAvailability as EquipmentAvailability[]).map((item, idx) => (
              <div key={idx} className={styles.tableRow}>
                <div className={styles.cell}>{item.partner_name}</div>
                <div className={styles.cell}>{item.equipment_type}</div>
                <div className={styles.cell}>{item.quantity_available}</div>
                <div className={styles.cell}>{item.location}</div>
                <div className={styles.cell}>{item.delivery_days}d</div>
                <div className={styles.cell}>{item.price_per_day}€</div>
              </div>
            ))}
            </div>
          )}
        </section>

        {/* Sub-Rental Requests */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Untermiete-Anfragen</h2>

          {isLoadingSubRentals ? (
            <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
              Laden...
            </div>
          ) : subRentalRequests.length === 0 ? (
            <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
              Keine Anfragen vorhanden
            </div>
          ) : (
            <div className={styles.requestsList}>
              {(subRentalRequests as SubRentalRequest[]).map((request: SubRentalRequest) => (
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
                    <span style={{ color: getStatusBadgeColor(request.status) }}>●</span>
                    {request.status}
                  </div>
                </div>

                <div className={styles.requestDetails}>
                  <span><strong>Menge:</strong> {request.quantity}</span>
                  <span><strong>Von:</strong> {new Date(request.start_date).toLocaleDateString('de-DE')}</span>
                  <span><strong>Bis:</strong> {new Date(request.end_date).toLocaleDateString('de-DE')}</span>
                </div>

                <div className={styles.requestActions}>
                  {request.status === 'pending' && (
                    <>
                      <button className="btn btn--sm btn--primary">Genehmigen</button>
                      <button className="btn btn--sm">Ablehnen</button>
                    </>
                  )}
                  {request.status !== 'pending' && (
                    <button className="btn btn--sm">Details</button>
                  )}
                </div>
              </div>
            ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}

export default FederationPage
