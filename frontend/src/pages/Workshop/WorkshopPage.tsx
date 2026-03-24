import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { equipmentApi, maintenanceApi } from '../../services/api'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
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

const PRIORITY_MAP: Record<string, 'Hoch' | 'Mittel' | 'Niedrig'> = {
  critical: 'Hoch',
  high: 'Hoch',
  medium: 'Mittel',
  low: 'Niedrig',
}

const STATUS_MAP: Record<string, 'Offen' | 'In Arbeit' | 'Erledigt'> = {
  pending: 'Offen',
  planned: 'Offen',
  overdue: 'Offen',
  in_progress: 'In Arbeit',
  completed: 'Erledigt',
  cancelled: 'Erledigt',
}

function WorkshopPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('repairs')
  const [searchQuery, setSearchQuery] = useState('')
  const [showRepairModal, setShowRepairModal] = useState(false)

  const { data: equipmentData, isLoading: eqLoading } = useQuery({
    queryKey: ['equipment-workshop'],
    queryFn: () => equipmentApi.list({ limit: 100 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['maintenance-tasks-workshop'],
    queryFn: () => maintenanceApi.listTasks(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: plansData, isLoading: plansLoading } = useQuery({
    queryKey: ['maintenance-plans-workshop'],
    queryFn: () => maintenanceApi.listPlans(),
    staleTime: 1000 * 60 * 5,
  })

  const isLoading = eqLoading || tasksLoading || plansLoading

  const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])
  const tasks = tasksData?.items || tasksData?.data || (Array.isArray(tasksData) ? tasksData : [])
  const plans = plansData?.items || plansData?.data || (Array.isArray(plansData) ? plansData : [])

  // Map maintenance tasks to repairs
  const repairs: Repair[] = useMemo(() => {
    return tasks.map((t: any) => ({
      id: t.id,
      equipmentName: t.equipment_name || 'Unbekannt',
      defect: t.plan_name || t.notes || 'Wartungsaufgabe',
      priority: PRIORITY_MAP[t.priority] || 'Mittel',
      status: STATUS_MAP[t.status] || 'Offen',
      assignedTo: t.assigned_to || '',
      createdAt: t.created_at || t.scheduled_at || '',
    }))
  }, [tasks])

  // Map maintenance plans to inspections
  const inspections: Inspection[] = useMemo(() => {
    const now = new Date()
    return plans.map((p: any) => {
      const nextDate = p.next_due_at || p.next_maintenance || ''
      const isActive = p.is_active ?? (p.status === 'active')
      let status: Inspection['status'] = 'Geplant'
      if (nextDate) {
        const due = new Date(nextDate)
        const daysUntil = Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
        if (daysUntil < 0) status = 'Überfällig'
        else if (daysUntil <= 7) status = 'Fällig'
        else status = 'Geplant'
      }
      if (!isActive && p.status === 'completed') status = 'Bestanden'

      return {
        id: p.id,
        equipmentName: p.equipment_name || p.name || 'Unbekannt',
        nextDate: nextDate,
        type: p.plan_type || p.frequency || 'Wartung',
        status,
      }
    })
  }, [plans])

  // Lost items from equipment with status 'lost' or 'missing'
  const lostItems: LostItem[] = useMemo(() => {
    return equipment
      .filter((e: any) => e.status === 'lost' || e.status === 'missing')
      .map((e: any) => ({
        id: e.id,
        equipmentName: e.name,
        lastSeen: e.location || '',
        project: e.project_name || '',
        reportedBy: '',
        reportedAt: e.updated_at || e.created_at || '',
      }))
  }, [equipment])

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
          <SkeletonTable rows={5} columns={5} />
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
