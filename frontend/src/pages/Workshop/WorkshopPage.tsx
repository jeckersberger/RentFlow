import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { equipmentApi } from '../../services/api'
import styles from './Workshop.module.scss'

type TabKey = 'repairs' | 'inspections' | 'lost' | 'inventory'

interface Repair {
  id: string
  equipmentName: string
  defect: string
  priority: 'Hoch' | 'Mittel' | 'Niedrig'
  status: 'Offen' | 'In Arbeit' | 'Erledigt'
  assignedTo: string
  createdAt: string
}

interface Inspection {
  id: string
  equipmentName: string
  nextDate: string
  type: string
  status: 'Fällig' | 'Geplant' | 'Bestanden' | 'Überfällig'
}

interface LostItem {
  id: string
  equipmentName: string
  lastSeen: string
  project: string
  reportedBy: string
  reportedAt: string
}

function WorkshopPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('repairs')
  const [searchQuery, setSearchQuery] = useState('')
  const [showRepairModal, setShowRepairModal] = useState(false)

  const { data: equipmentData, isLoading } = useQuery({
    queryKey: ['equipment-workshop'],
    queryFn: () => equipmentApi.list({ limit: 100 }),
    staleTime: 1000 * 60 * 5,
  })

  const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])

  // Generate simulated repair data from equipment in maintenance
  const repairs: Repair[] = useMemo(() => {
    const inMaintenance = equipment.filter((e: any) => e.status === 'in_maintenance' || e.condition === 'damaged' || e.condition === 'fair')
    const simulated: Repair[] = inMaintenance.map((e: any, idx: number) => ({
      id: `rep-${e.id}`,
      equipmentName: e.name,
      defect: idx % 3 === 0 ? 'Defekter XLR-Anschluss' : idx % 3 === 1 ? 'Motorschaden Moving Head' : 'Kabelbruch Netzteil',
      priority: idx % 3 === 0 ? 'Hoch' : idx % 3 === 1 ? 'Mittel' : 'Niedrig',
      status: idx % 3 === 0 ? 'Offen' : idx % 3 === 1 ? 'In Arbeit' : 'Erledigt',
      assignedTo: idx % 2 === 0 ? 'Thomas Müller' : 'Sarah Schmidt',
      createdAt: new Date(Date.now() - idx * 86400000 * 3).toISOString(),
    }))

    // Always have at least demo data
    if (simulated.length === 0) {
      return [
        { id: 'rep-demo-1', equipmentName: 'Yamaha CL5', defect: 'Fader Kanal 12 reagiert nicht', priority: 'Hoch', status: 'Offen', assignedTo: 'Thomas Müller', createdAt: '2026-03-20T10:00:00Z' },
        { id: 'rep-demo-2', equipmentName: 'Martin MAC Aura XB', defect: 'Pan/Tilt Motor defekt', priority: 'Mittel', status: 'In Arbeit', assignedTo: 'Sarah Schmidt', createdAt: '2026-03-18T14:00:00Z' },
        { id: 'rep-demo-3', equipmentName: 'Shure SM58', defect: 'Kapsel locker', priority: 'Niedrig', status: 'Erledigt', assignedTo: 'Thomas Müller', createdAt: '2026-03-15T09:00:00Z' },
      ]
    }
    return simulated
  }, [equipment])

  const inspections: Inspection[] = useMemo(() => [
    { id: 'insp-1', equipmentName: 'Chainmaster BGV-D8+ 1t', nextDate: '2026-04-15', type: 'DGUV V3', status: 'Geplant' },
    { id: 'insp-2', equipmentName: 'JBL VTX A12', nextDate: '2026-04-01', type: 'BGV A3', status: 'Fällig' },
    { id: 'insp-3', equipmentName: 'Prolyte X30V Truss 3m', nextDate: '2026-03-10', type: 'DGUV V3', status: 'Überfällig' },
    { id: 'insp-4', equipmentName: 'MA Lighting grandMA3', nextDate: '2026-05-20', type: 'BGV A3', status: 'Geplant' },
    { id: 'insp-5', equipmentName: 'Robe MegaPointe', nextDate: '2026-06-01', type: 'DGUV V3', status: 'Bestanden' },
  ], [])

  const lostItems: LostItem[] = useMemo(() => [
    { id: 'lost-1', equipmentName: 'Shure SM58 (RF-MIC-001)', lastSeen: 'Marienplatz, München', project: 'Stadtfest München 2026', reportedBy: 'Max Huber', reportedAt: '2026-03-19T18:00:00Z' },
    { id: 'lost-2', equipmentName: 'XLR-Kabel 10m (x3)', lastSeen: 'Hilton Hotel, Frankfurt', project: 'Firmen-Gala TechCorp', reportedBy: 'Sarah Schmidt', reportedAt: '2026-03-21T09:00:00Z' },
  ], [])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return isNaN(date.getTime()) ? 'N/A' : date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' })
  }

  const tabs: { key: TabKey; label: string; count: number }[] = [
    { key: 'repairs', label: 'Reparaturen', count: repairs.filter(r => r.status !== 'Erledigt').length },
    { key: 'inspections', label: 'Prüfungen', count: inspections.filter(i => i.status === 'Fällig' || i.status === 'Überfällig').length },
    { key: 'lost', label: 'Verlorene Materialien', count: lostItems.length },
    { key: 'inventory', label: 'Bestandszählung', count: 0 },
  ]

  const getPriorityClass = (priority: string) => {
    switch (priority) {
      case 'Hoch': return styles.priorityHigh
      case 'Mittel': return styles.priorityMedium
      case 'Niedrig': return styles.priorityLow
      default: return styles.priorityMedium
    }
  }

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'Offen': return styles.statusOpen
      case 'In Arbeit': return styles.statusProgress
      case 'Erledigt': return styles.statusDone
      case 'Fällig': return styles.statusWarning
      case 'Überfällig': return styles.statusDanger
      case 'Geplant': return styles.statusPlanned
      case 'Bestanden': return styles.statusDone
      default: return ''
    }
  }

  const openRepairs = repairs.filter(r => r.status === 'Offen').length
  const inProgressRepairs = repairs.filter(r => r.status === 'In Arbeit').length
  const overdueInspections = inspections.filter(i => i.status === 'Überfällig').length

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Werkstatt</h1>
          <p className={styles.subtitle}>Reparaturen, Prüfungen und Bestandsverwaltung</p>
        </div>
        <div className={styles.headerActions}>
          <button className={styles.btnPrimary} onClick={() => setShowRepairModal(true)}>
            + Reparatur melden
          </button>
          <button className={styles.btnSecondary}>
            + Prüfung planen
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className={styles.statsGrid}>
        <div className={`${styles.statCard} ${openRepairs > 0 ? styles.statCardDanger : ''}`}>
          <div className={styles.statLabel}>Offene Reparaturen</div>
          <div className={styles.statValue}>{openRepairs}</div>
        </div>
        <div className={`${styles.statCard} ${inProgressRepairs > 0 ? styles.statCardWarning : ''}`}>
          <div className={styles.statLabel}>In Arbeit</div>
          <div className={styles.statValue}>{inProgressRepairs}</div>
        </div>
        <div className={`${styles.statCard} ${overdueInspections > 0 ? styles.statCardDanger : ''}`}>
          <div className={styles.statLabel}>Überfällige Prüfungen</div>
          <div className={styles.statValue}>{overdueInspections}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Verlorene Teile</div>
          <div className={styles.statValue}>{lostItems.length}</div>
        </div>
      </div>

      {/* Tabs */}
      <div className={styles.tabBar}>
        {tabs.map(tab => (
          <button
            key={tab.key}
            className={`${styles.tab} ${activeTab === tab.key ? styles.tabActive : ''}`}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
            {tab.count > 0 && <span className={styles.tabBadge}>{tab.count}</span>}
          </button>
        ))}
      </div>

      {/* Filter */}
      <div className={styles.filterBar}>
        <input
          type="text"
          className={styles.searchInput}
          placeholder="Suchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content */}
      <div className={styles.tableCard}>
        {isLoading ? (
          <div className={styles.loadingState}>
            <div className={styles.spinner} />
            <p>Daten werden geladen...</p>
          </div>
        ) : activeTab === 'repairs' ? (
          repairs.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>🔧</div>
              <h3 className={styles.emptyTitle}>Keine Reparaturen</h3>
              <p className={styles.emptyDescription}>Alle Geräte sind in einwandfreiem Zustand.</p>
              <button className={styles.btnPrimary} onClick={() => setShowRepairModal(true)}>Reparatur melden</button>
            </div>
          ) : (
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Equipment</th>
                  <th>Defektbeschreibung</th>
                  <th>Priorität</th>
                  <th>Status</th>
                  <th>Zugewiesen an</th>
                  <th>Erstellt</th>
                </tr>
              </thead>
              <tbody>
                {repairs.filter(r => r.equipmentName.toLowerCase().includes(searchQuery.toLowerCase())).map(repair => (
                  <tr key={repair.id}>
                    <td className={styles.equipmentName}>{repair.equipmentName}</td>
                    <td>{repair.defect}</td>
                    <td><span className={`${styles.badge} ${getPriorityClass(repair.priority)}`}>{repair.priority}</span></td>
                    <td><span className={`${styles.badge} ${getStatusClass(repair.status)}`}>{repair.status}</span></td>
                    <td className={styles.assignee}>{repair.assignedTo}</td>
                    <td className={styles.dateCell}>{formatDate(repair.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        ) : activeTab === 'inspections' ? (
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Equipment</th>
                <th>Nächste Prüfung</th>
                <th>Typ</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {inspections.filter(i => i.equipmentName.toLowerCase().includes(searchQuery.toLowerCase())).map(insp => (
                <tr key={insp.id}>
                  <td className={styles.equipmentName}>{insp.equipmentName}</td>
                  <td className={styles.dateCell}>{formatDate(insp.nextDate)}</td>
                  <td><span className={styles.typeBadge}>{insp.type}</span></td>
                  <td><span className={`${styles.badge} ${getStatusClass(insp.status)}`}>{insp.status}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : activeTab === 'lost' ? (
          lostItems.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📦</div>
              <h3 className={styles.emptyTitle}>Keine verlorenen Materialien</h3>
              <p className={styles.emptyDescription}>Alle Materialien sind erfasst.</p>
            </div>
          ) : (
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Equipment</th>
                  <th>Zuletzt gesehen</th>
                  <th>Projekt</th>
                  <th>Gemeldet von</th>
                  <th>Gemeldet am</th>
                </tr>
              </thead>
              <tbody>
                {lostItems.map(item => (
                  <tr key={item.id}>
                    <td className={styles.equipmentName}>{item.equipmentName}</td>
                    <td>{item.lastSeen}</td>
                    <td><span className={styles.projectTag}>{item.project}</span></td>
                    <td>{item.reportedBy}</td>
                    <td className={styles.dateCell}>{formatDate(item.reportedAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        ) : (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>📊</div>
            <h3 className={styles.emptyTitle}>Bestandszählung</h3>
            <p className={styles.emptyDescription}>
              Starten Sie eine neue Bestandszählung, um Ihren Lagerbestand abzugleichen.
            </p>
            <button className={styles.btnPrimary}>Zählung starten</button>
          </div>
        )}
      </div>

      {/* Repair Modal */}
      {showRepairModal && (
        <div className={styles.modalOverlay} onClick={() => setShowRepairModal(false)}>
          <div className={styles.modal} onClick={e => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Reparatur melden</h2>
              <button className={styles.modalClose} onClick={() => setShowRepairModal(false)}>&times;</button>
            </div>
            <div className={styles.modalBody}>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Equipment</label>
                <select className={styles.formSelect}>
                  <option value="">Equipment auswählen...</option>
                  {equipment.map((e: any) => (
                    <option key={e.id} value={e.id}>{e.name}</option>
                  ))}
                </select>
              </div>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Defektbeschreibung</label>
                <textarea className={styles.formTextarea} rows={3} placeholder="Beschreiben Sie den Defekt..." />
              </div>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Priorität</label>
                <select className={styles.formSelect}>
                  <option value="Niedrig">Niedrig</option>
                  <option value="Mittel">Mittel</option>
                  <option value="Hoch">Hoch</option>
                </select>
              </div>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Zuweisen an</label>
                <input type="text" className={styles.formInput} placeholder="Mitarbeitername..." />
              </div>
            </div>
            <div className={styles.modalFooter}>
              <button className={styles.btnSecondary} onClick={() => setShowRepairModal(false)}>Abbrechen</button>
              <button className={styles.btnPrimary} onClick={() => setShowRepairModal(false)}>Reparatur erstellen</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default WorkshopPage
