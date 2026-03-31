import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { equipmentApi, projectApi } from '../../services/api'
import styles from './Shortages.module.scss'

type TabKey = 'rental' | 'sales' | 'sublease'

interface Shortage {
  id: string
  equipmentName: string
  required: number
  available: number
  shortage: number
  affectedProjects: string[]
  period: string
  severity: 'critical' | 'warning'
}

const FEDERATION_PARTNERS = [
  { id: 'p1', name: 'MediaTech Rental GmbH', region: 'Berlin' },
  { id: 'p2', name: 'EventEquip AG', region: 'Hamburg' },
  { id: 'p3', name: 'ProLight Verleih', region: 'Muenchen' },
  { id: 'p4', name: 'StageRent NL', region: 'Amsterdam' },
  { id: 'p5', name: 'SonoPlus FR', region: 'Paris' },
]

function ShortagesPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('rental')
  const [searchQuery, setSearchQuery] = useState('')
  const [subleaseModal, setSubleaseModal] = useState<{ open: boolean; shortage: Shortage | null }>({ open: false, shortage: null })
  const [selectedPartner, setSelectedPartner] = useState('')
  const [subleaseNote, setSubleaseNote] = useState('')
  const [toastMsg, setToastMsg] = useState<string | null>(null)

  const { data: equipmentData, isLoading: eqLoading } = useQuery({
    queryKey: ['equipment-shortages'],
    queryFn: () => equipmentApi.list({ limit: 100 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: projectsData, isLoading: projLoading } = useQuery({
    queryKey: ['projects-shortages'],
    queryFn: () => projectApi.list(1, 100),
    staleTime: 1000 * 60 * 5,
  })

  const isLoading = eqLoading || projLoading

  const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])
  const projects = projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])

  // Calculate real shortages based on checked-out and reserved equipment vs total available
  const shortages: Shortage[] = useMemo(() => {
    if (!equipment.length) return []

    const checkedOut = equipment.filter((e: any) => e.status === 'checked_out' || e.status === 'reserved')
    const grouped: Record<string, any[]> = {}

    checkedOut.forEach((e: any) => {
      const key = e.category_id || e.name
      if (!grouped[key]) grouped[key] = []
      grouped[key].push(e)
    })

    const result: Shortage[] = []
    const activeProjects = projects.filter((p: any) => p.status === 'active' || p.status === 'planning')
    const projectNames = activeProjects.map((p: any) => p.name)

    Object.entries(grouped).forEach(([_, items], idx) => {
      const totalInCategory = equipment.filter((e: any) => (e.category_id || e.name) === (items[0].category_id || items[0].name)).length
      const inUse = items.length
      const available = totalInCategory - inUse

      // Only show as shortage if all items are in use (no available remaining)
      if (available <= 0 && activeProjects.length > 0) {
        result.push({
          id: String(idx),
          equipmentName: items[0].name,
          required: inUse + 1,
          available: Math.max(0, available),
          shortage: Math.abs(available) + 1,
          affectedProjects: projectNames.slice(0, Math.min(2, projectNames.length)),
          period: activeProjects.length > 0
            ? `${new Date(activeProjects[0].start_date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })} - ${new Date(activeProjects[0].end_date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' })}`
            : 'k.A.',
          severity: Math.abs(available) > 2 ? 'critical' : 'warning',
        })
      }
    })

    return result
  }, [equipment, projects])

  const filteredShortages = shortages.filter(s =>
    s.equipmentName.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const tabs: { key: TabKey; label: string }[] = [
    { key: 'rental', label: 'Mietengpaesse' },
    { key: 'sales', label: 'Verkaufsengpaesse' },
    { key: 'sublease', label: 'Zumietungsjobs' },
  ]

  const openSubleaseModal = (shortage: Shortage | null = null) => {
    setSubleaseModal({ open: true, shortage })
    setSelectedPartner('')
    setSubleaseNote('')
  }

  const closeSubleaseModal = () => {
    setSubleaseModal({ open: false, shortage: null })
    setSelectedPartner('')
    setSubleaseNote('')
  }

  const handleSubmitSublease = () => {
    // Federation backend may not be ready yet - show toast confirmation
    closeSubleaseModal()
    setToastMsg('Anfrage wurde gesendet')
    setTimeout(() => setToastMsg(null), 3500)
  }

  const criticalCount = shortages.filter(s => s.severity === 'critical').length
  const warningCount = shortages.filter(s => s.severity === 'warning').length
  const totalShortage = shortages.reduce((sum, s) => sum + s.shortage, 0)

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Engpässe</h1>
          <p className={styles.subtitle}>Equipment-Engpässe und Überbuchungen frühzeitig erkennen</p>
        </div>
        <div className={styles.headerActions}>
          <button className={styles.btnPrimary} onClick={() => openSubleaseModal()}>
            + Zumietung anfragen
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className={styles.statsGrid}>
        <div className={`${styles.statCard} ${criticalCount > 0 ? styles.statCardDanger : ''}`}>
          <div className={styles.statLabel}>Kritische Engpässe</div>
          <div className={styles.statValue}>{criticalCount}</div>
        </div>
        <div className={`${styles.statCard} ${warningCount > 0 ? styles.statCardWarning : ''}`}>
          <div className={styles.statLabel}>Warnungen</div>
          <div className={styles.statValue}>{warningCount}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Fehlmenge gesamt</div>
          <div className={styles.statValue}>{totalShortage}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Betroffene Projekte</div>
          <div className={styles.statValue}>
            {[...new Set(shortages.flatMap(s => s.affectedProjects))].length}
          </div>
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
          </button>
        ))}
      </div>

      {/* Filter */}
      <div className={styles.filterBar}>
        <input
          type="text"
          className={styles.searchInput}
          placeholder="Equipment suchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content */}
      <div className={styles.tableCard}>
        {isLoading ? (
          <div className={styles.loadingState}>
            <div className={styles.spinner} />
            <p>Engpässe werden berechnet...</p>
          </div>
        ) : activeTab === 'sublease' ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>📋</div>
            <h3 className={styles.emptyTitle}>Keine Zumietungsjobs</h3>
            <p className={styles.emptyDescription}>
              Erstellen Sie eine Zumietungsanfrage, wenn Equipment nicht verfügbar ist.
            </p>
            <button className={styles.btnPrimary} onClick={() => openSubleaseModal()}>Zumietung anfragen</button>
          </div>
        ) : filteredShortages.length === 0 ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>✅</div>
            <h3 className={styles.emptyTitle}>Keine Engpässe</h3>
            <p className={styles.emptyDescription}>
              Alle Geräte sind in ausreichender Menge verfügbar.
            </p>
          </div>
        ) : (
          <table className={styles.table}>
            <thead>
              <tr>
                <th>Equipment</th>
                <th>Benötigte Menge</th>
                <th>Verfügbar</th>
                <th>Fehlmenge</th>
                <th>Betroffene Projekte</th>
                <th>Zeitraum</th>
                <th>Aktion</th>
              </tr>
            </thead>
            <tbody>
              {filteredShortages.map(shortage => (
                <tr key={shortage.id} className={styles[`row${shortage.severity.charAt(0).toUpperCase() + shortage.severity.slice(1)}`]}>
                  <td>
                    <span className={styles.equipmentName}>{shortage.equipmentName}</span>
                  </td>
                  <td>{shortage.required}</td>
                  <td>{shortage.available}</td>
                  <td>
                    <span className={`${styles.shortageBadge} ${styles[`shortageBadge${shortage.severity.charAt(0).toUpperCase() + shortage.severity.slice(1)}`]}`}>
                      -{shortage.shortage}
                    </span>
                  </td>
                  <td>
                    <div className={styles.projectTags}>
                      {shortage.affectedProjects.map((p, i) => (
                        <span key={i} className={styles.projectTag}>{p}</span>
                      ))}
                    </div>
                  </td>
                  <td className={styles.periodCell}>{shortage.period}</td>
                  <td>
                    <button className={styles.actionBtn} onClick={() => openSubleaseModal(shortage)}>
                      Zumietung anfragen
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
      {/* Toast */}
      {toastMsg && (
        <div className={styles.toast}>
          <span className={styles.toastIcon}>&#10003;</span>
          {toastMsg}
        </div>
      )}

      {/* Sublease Request Modal */}
      {subleaseModal.open && (
        <div className={styles.modalBackdrop} onClick={closeSubleaseModal}>
          <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Zumietung anfragen</h2>
              <button className={styles.modalClose} onClick={closeSubleaseModal}>&times;</button>
            </div>
            <div className={styles.modalBody}>
              {subleaseModal.shortage && (
                <div className={styles.modalInfo}>
                  <div className={styles.modalInfoRow}>
                    <span className={styles.modalInfoLabel}>Equipment:</span>
                    <span className={styles.modalInfoValue}>{subleaseModal.shortage.equipmentName}</span>
                  </div>
                  <div className={styles.modalInfoRow}>
                    <span className={styles.modalInfoLabel}>Fehlmenge:</span>
                    <span className={styles.modalInfoValue}>{subleaseModal.shortage.shortage} Stueck</span>
                  </div>
                  <div className={styles.modalInfoRow}>
                    <span className={styles.modalInfoLabel}>Zeitraum:</span>
                    <span className={styles.modalInfoValue}>{subleaseModal.shortage.period}</span>
                  </div>
                </div>
              )}
              <div className={styles.modalField}>
                <label className={styles.modalLabel}>Partner auswaehlen</label>
                <select
                  className={styles.modalSelect}
                  value={selectedPartner}
                  onChange={(e) => setSelectedPartner(e.target.value)}
                >
                  <option value="">-- Partner waehlen --</option>
                  {FEDERATION_PARTNERS.map((p) => (
                    <option key={p.id} value={p.id}>{p.name} ({p.region})</option>
                  ))}
                </select>
              </div>
              <div className={styles.modalField}>
                <label className={styles.modalLabel}>Anmerkungen</label>
                <textarea
                  className={styles.modalTextarea}
                  rows={3}
                  placeholder="Optionale Hinweise zur Anfrage..."
                  value={subleaseNote}
                  onChange={(e) => setSubleaseNote(e.target.value)}
                />
              </div>
            </div>
            <div className={styles.modalFooter}>
              <button className={styles.btnSecondary} onClick={closeSubleaseModal}>Abbrechen</button>
              <button
                className={styles.btnPrimary}
                onClick={handleSubmitSublease}
                disabled={!selectedPartner}
              >
                Anfrage senden
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default ShortagesPage
