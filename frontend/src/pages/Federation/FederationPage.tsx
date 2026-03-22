import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { Partner, EquipmentAvailability, SubRentalRequest } from '../../types/federation'
import styles from './Federation.module.scss'

const mockPartners: Partner[] = [
  {
    id: '1',
    name: 'TechRent GmbH',
    company: 'TechRent',
    email: 'contact@techrent.de',
    phone: '+49 30 123456',
    country: 'Germany',
    trust_level: 'verified',
    equipment_availability: 45,
    certificate_status: 'valid',
    certificate_expiry: '2027-06-15',
    joined_date: '2024-03-01',
    rating: 4.8,
  },
  {
    id: '2',
    name: 'EventPro Solutions',
    company: 'EventPro',
    email: 'info@eventpro.de',
    phone: '+49 40 234567',
    country: 'Germany',
    trust_level: 'trusted',
    equipment_availability: 32,
    certificate_status: 'valid',
    certificate_expiry: '2026-09-20',
    joined_date: '2023-11-15',
    rating: 4.5,
  },
  {
    id: '3',
    name: 'Ausrüstungs Zentrale',
    company: 'AZ Rental',
    email: 'support@az-rental.de',
    phone: '+49 69 345678',
    country: 'Germany',
    trust_level: 'verified',
    equipment_availability: 67,
    certificate_status: 'expiring',
    certificate_expiry: '2026-04-30',
    joined_date: '2023-06-20',
    rating: 4.7,
  },
  {
    id: '4',
    name: 'Studio Equipment Plus',
    company: 'StudioEQ',
    email: 'contact@studioeq.de',
    phone: '+49 80 456789',
    country: 'Germany',
    trust_level: 'provisional',
    equipment_availability: 18,
    certificate_status: 'pending',
    certificate_expiry: '2026-05-10',
    joined_date: '2026-01-10',
    rating: 3.9,
  },
]

const mockEquipmentAvailability: EquipmentAvailability[] = [
  {
    partner_id: '1',
    partner_name: 'TechRent GmbH',
    equipment_type: 'LED-Wände',
    quantity_available: 8,
    location: 'Berlin',
    delivery_days: 1,
    price_per_day: 250,
  },
  {
    partner_id: '1',
    partner_name: 'TechRent GmbH',
    equipment_type: 'Beamer',
    quantity_available: 12,
    location: 'Berlin',
    delivery_days: 1,
    price_per_day: 80,
  },
  {
    partner_id: '2',
    partner_name: 'EventPro Solutions',
    equipment_type: 'Tische & Stühle',
    quantity_available: 200,
    location: 'Hamburg',
    delivery_days: 2,
    price_per_day: 15,
  },
  {
    partner_id: '3',
    partner_name: 'Ausrüstungs Zentrale',
    equipment_type: 'Gerüste',
    quantity_available: 45,
    location: 'Frankfurt',
    delivery_days: 1,
    price_per_day: 120,
  },
  {
    partner_id: '3',
    partner_name: 'Ausrüstungs Zentrale',
    equipment_type: 'Stromgeneratoren',
    quantity_available: 15,
    location: 'Frankfurt',
    delivery_days: 2,
    price_per_day: 200,
  },
]

const mockSubRentalRequests: SubRentalRequest[] = [
  {
    id: '1',
    partner_id: '1',
    partner_name: 'TechRent GmbH',
    equipment: 'LED-Wände',
    quantity: 6,
    start_date: '2026-03-25',
    end_date: '2026-03-30',
    status: 'approved',
    requested_at: '2026-03-20T10:30:00Z',
  },
  {
    id: '2',
    partner_id: '2',
    partner_name: 'EventPro Solutions',
    equipment: 'Stühle',
    quantity: 150,
    start_date: '2026-03-24',
    end_date: '2026-03-25',
    status: 'pending',
    requested_at: '2026-03-22T09:15:00Z',
  },
  {
    id: '3',
    partner_id: '3',
    partner_name: 'Ausrüstungs Zentrale',
    equipment: 'Gerüste',
    quantity: 30,
    start_date: '2026-04-01',
    end_date: '2026-04-05',
    status: 'pending',
    requested_at: '2026-03-21T14:45:00Z',
  },
]

function FederationPage() {
  const [filterTrust, setFilterTrust] = useState<string>('all')

  const { data: partners = mockPartners } = useQuery({
    queryKey: ['federation-partners'],
    queryFn: async () => mockPartners,
    staleTime: 1000 * 60 * 5,
  })

  const { data: equipmentAvailability = mockEquipmentAvailability } = useQuery({
    queryKey: ['equipment-availability'],
    queryFn: async () => mockEquipmentAvailability,
    staleTime: 1000 * 60 * 5,
  })

  const { data: subRentalRequests = mockSubRentalRequests } = useQuery({
    queryKey: ['sub-rental-requests'],
    queryFn: async () => mockSubRentalRequests,
    staleTime: 1000 * 60 * 5,
  })

  const filteredPartners = filterTrust === 'all'
    ? partners
    : partners.filter(p => p.trust_level === filterTrust)

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

  const totalEquipmentCount = equipmentAvailability.reduce((sum, e) => sum + e.quantity_available, 0)

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
            {partners.filter(p => p.trust_level === 'verified').length}
          </div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Verfügbare Ausrüstung</div>
          <div className={styles.statValue}>{totalEquipmentCount}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Anfragen ausstehend</div>
          <div className={styles.statValue} style={{ color: '#f59e0b' }}>
            {subRentalRequests.filter(r => r.status === 'pending').length}
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

        <div className={styles.partnersGrid}>
          {filteredPartners.map(partner => (
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
      </section>

      <div className={styles.bottomGrid}>
        {/* Equipment Availability */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Verfügbare Ausrüstung</h2>

          <div className={styles.equipmentTable}>
            <div className={styles.tableHeader}>
              <div className={styles.headerCell}>Partner</div>
              <div className={styles.headerCell}>Ausrüstung</div>
              <div className={styles.headerCell}>Menge</div>
              <div className={styles.headerCell}>Standort</div>
              <div className={styles.headerCell}>Liefertage</div>
              <div className={styles.headerCell}>Preis/Tag</div>
            </div>

            {equipmentAvailability.map((item, idx) => (
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
        </section>

        {/* Sub-Rental Requests */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>Untermiete-Anfragen</h2>

          <div className={styles.requestsList}>
            {subRentalRequests.map(request => (
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
        </section>
      </div>
    </div>
  )
}

export default FederationPage
