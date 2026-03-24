import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { timeTrackingApi } from '../../services/api'
import styles from './TimeTracking.module.scss'

type TabKey = 'hours' | 'activities' | 'absence'

interface TimeEntry {
  id: string
  date: string
  employee: string
  project: string
  activity: string
  start: string
  end: string
  duration: string
  status: 'Offen' | 'Genehmigt'
}

interface AbsenceEntry {
  id: string
  employee: string
  type: string
  from: string
  to: string
  days: number
  status: 'Genehmigt' | 'Ausstehend' | 'Abgelehnt'
}

const WEEKDAYS = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So']

function mapTimeEntry(dto: any): TimeEntry {
  const start = dto.start_time || dto.start || ''
  const end = dto.end_time || dto.end || ''
  const duration = dto.duration || ''
  const statusMap: Record<string, TimeEntry['status']> = {
    open: 'Offen',
    pending: 'Offen',
    approved: 'Genehmigt',
  }
  return {
    id: dto.id,
    date: dto.date || (dto.start_time ? dto.start_time.split('T')[0] : ''),
    employee: dto.employee_name || dto.employee || '',
    project: dto.project_name || dto.project || '',
    activity: dto.activity || dto.description || '',
    start: start.includes('T') ? start.split('T')[1]?.substring(0, 5) : start,
    end: end.includes('T') ? end.split('T')[1]?.substring(0, 5) : end,
    duration,
    status: statusMap[dto.status] || 'Offen',
  }
}

function mapAbsenceEntry(dto: any): AbsenceEntry {
  const statusMap: Record<string, AbsenceEntry['status']> = {
    approved: 'Genehmigt',
    pending: 'Ausstehend',
    rejected: 'Abgelehnt',
  }
  return {
    id: dto.id,
    employee: dto.employee_name || dto.employee || '',
    type: dto.type || dto.absence_type || '',
    from: dto.start_date || dto.from || '',
    to: dto.end_date || dto.to || '',
    days: dto.days || dto.duration_days || 0,
    status: statusMap[dto.status] || 'Ausstehend',
  }
}

function TimeTrackingPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('hours')
  const [searchQuery, setSearchQuery] = useState('')
  const [showModal, setShowModal] = useState(false)

  const { data: timeEntries = [], isLoading: isLoadingEntries } = useQuery({
    queryKey: ['time-entries'],
    queryFn: async () => {
      const result = await timeTrackingApi.listEntries()
      const items = Array.isArray(result) ? result : (result.data || [])
      return items.map(mapTimeEntry)
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })

  const { data: absences = [], isLoading: isLoadingAbsences } = useQuery({
    queryKey: ['absences'],
    queryFn: async () => {
      const result = await timeTrackingApi.listAbsences()
      const items = Array.isArray(result) ? result : (result.data || [])
      return items.map(mapAbsenceEntry)
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })

  const filteredEntries = timeEntries.filter((e: TimeEntry) =>
    e.employee.toLowerCase().includes(searchQuery.toLowerCase()) ||
    e.project.toLowerCase().includes(searchQuery.toLowerCase())
  )

  // Weekly summary: compute hours per day of current week
  const weeklySummary = useMemo(() => {
    const today = new Date()
    const monday = new Date(today)
    monday.setDate(today.getDate() - ((today.getDay() + 6) % 7))

    return WEEKDAYS.map((label, idx) => {
      const day = new Date(monday)
      day.setDate(monday.getDate() + idx)
      const dateStr = day.toISOString().split('T')[0]

      const dayEntries = timeEntries.filter((e: TimeEntry) => e.date === dateStr)
      const totalMinutes = dayEntries.reduce((sum: number, e: TimeEntry) => {
        const [h, m] = e.duration.split(':').map(Number)
        return sum + (isNaN(h) ? 0 : h) * 60 + (isNaN(m) ? 0 : m)
      }, 0)

      return {
        label,
        date: day.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }),
        hours: Math.floor(totalMinutes / 60),
        minutes: totalMinutes % 60,
        totalMinutes,
      }
    })
  }, [timeEntries])

  const maxMinutes = Math.max(...weeklySummary.map(d => d.totalMinutes), 1)

  const totalWeekHours = weeklySummary.reduce((s, d) => s + d.totalMinutes, 0)
  const openCount = timeEntries.filter((e: TimeEntry) => e.status === 'Offen').length

  const tabs: { key: TabKey; label: string }[] = [
    { key: 'hours', label: 'Stundenerfassung' },
    { key: 'activities', label: 'Aktivitäten' },
    { key: 'absence', label: 'Abwesenheit' },
  ]

  const formatDate = (d: string) => {
    const date = new Date(d)
    return isNaN(date.getTime()) ? d : date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' })
  }

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'Offen': return styles.statusOpen
      case 'Genehmigt': return styles.statusApproved
      case 'Ausstehend': return styles.statusPending
      case 'Abgelehnt': return styles.statusRejected
      default: return ''
    }
  }

  // Extract unique activities
  const activities: string[] = Array.from(new Set(timeEntries.map((e: TimeEntry) => e.activity))) as string[]

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Zeiterfassung</h1>
          <p className={styles.subtitle}>Arbeitszeiten erfassen und auswerten</p>
        </div>
        <div className={styles.headerActions}>
          <button className={styles.btnPrimary} onClick={() => setShowModal(true)}>
            + Zeit erfassen
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className={styles.statsGrid}>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Stunden diese Woche</div>
          <div className={styles.statValue}>{Math.floor(totalWeekHours / 60)}:{String(totalWeekHours % 60).padStart(2, '0')}</div>
        </div>
        <div className={`${styles.statCard} ${openCount > 0 ? styles.statCardWarning : ''}`}>
          <div className={styles.statLabel}>Offene Eintraege</div>
          <div className={styles.statValue}>{openCount}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Mitarbeiter aktiv</div>
          <div className={styles.statValue}>{[...new Set(timeEntries.map((e: TimeEntry) => e.employee))].length}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Projekte diese Woche</div>
          <div className={styles.statValue}>{[...new Set(timeEntries.map((e: TimeEntry) => e.project))].length}</div>
        </div>
      </div>

      {/* Weekly Summary Bar Chart */}
      <div className={styles.weeklyCard}>
        <h3 className={styles.weeklyTitle}>Wochenübersicht</h3>
        <div className={styles.weeklyChart}>
          {weeklySummary.map((day, idx) => (
            <div key={idx} className={styles.weeklyBar}>
              <div className={styles.barContainer}>
                <div
                  className={styles.bar}
                  style={{ height: `${day.totalMinutes > 0 ? (day.totalMinutes / maxMinutes) * 100 : 0}%` }}
                />
              </div>
              <div className={styles.barLabel}>{day.label}</div>
              <div className={styles.barValue}>
                {day.hours > 0 ? `${day.hours}:${String(day.minutes).padStart(2, '0')}` : '-'}
              </div>
            </div>
          ))}
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
          placeholder="Mitarbeiter oder Projekt suchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content */}
      <div className={styles.tableCard}>
        {activeTab === 'hours' ? (
          isLoadingEntries ? (
            <div className={styles.emptyState}>
              <h3 className={styles.emptyTitle}>Laden...</h3>
              <p className={styles.emptyDescription}>Zeiteintraege werden geladen.</p>
            </div>
          ) : filteredEntries.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'⏰'}</div>
              <h3 className={styles.emptyTitle}>Keine Zeiteintraege</h3>
              <p className={styles.emptyDescription}>
                {searchQuery ? 'Keine Eintraege gefunden. Versuchen Sie eine andere Suche.' : 'Erfassen Sie Arbeitszeiten fuer Ihr Team.'}
              </p>
              {!searchQuery && <button className={styles.btnPrimary} onClick={() => setShowModal(true)}>Zeit erfassen</button>}
            </div>
          ) : (
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Datum</th>
                  <th>Mitarbeiter</th>
                  <th>Projekt</th>
                  <th>Aktivität</th>
                  <th>Start</th>
                  <th>Ende</th>
                  <th>Dauer</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {filteredEntries.map((entry: TimeEntry) => (
                  <tr key={entry.id}>
                    <td className={styles.dateCell}>{formatDate(entry.date)}</td>
                    <td className={styles.employeeName}>{entry.employee}</td>
                    <td><span className={styles.projectTag}>{entry.project}</span></td>
                    <td>{entry.activity}</td>
                    <td className={styles.timeCell}>{entry.start}</td>
                    <td className={styles.timeCell}>{entry.end}</td>
                    <td className={styles.durationCell}>{entry.duration}</td>
                    <td><span className={`${styles.badge} ${getStatusClass(entry.status)}`}>{entry.status}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        ) : activeTab === 'activities' ? (
          isLoadingEntries ? (
            <div className={styles.emptyState}>
              <h3 className={styles.emptyTitle}>Laden...</h3>
              <p className={styles.emptyDescription}>Aktivitaeten werden geladen.</p>
            </div>
          ) : activities.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'📊'}</div>
              <h3 className={styles.emptyTitle}>Keine Aktivitaeten</h3>
              <p className={styles.emptyDescription}>Aktivitaeten werden angezeigt, sobald Zeiteintraege erfasst werden.</p>
            </div>
          ) : (
          <div className={styles.activitiesList}>
            {activities.map((act, idx) => {
              const actEntries = timeEntries.filter((e: TimeEntry) => e.activity === act)
              const totalMins = actEntries.reduce((sum: number, e: TimeEntry) => {
                const [h, m] = e.duration.split(':').map(Number)
                return sum + (isNaN(h) ? 0 : h) * 60 + (isNaN(m) ? 0 : m)
              }, 0)
              const maxMins = Math.max(...activities.map(a => timeEntries.filter((e: TimeEntry) => e.activity === a).reduce((s: number, e: TimeEntry) => { const [h, m] = e.duration.split(':').map(Number); return s + (isNaN(h) ? 0 : h) * 60 + (isNaN(m) ? 0 : m); }, 0)), 1)
              return (
                <div key={idx} className={styles.activityItem}>
                  <div className={styles.activityName}>{act}</div>
                  <div className={styles.activityMeta}>
                    {actEntries.length} Eintraege &bull; {Math.floor(totalMins / 60)}:{String(totalMins % 60).padStart(2, '0')} Stunden
                  </div>
                  <div className={styles.activityBar}>
                    <div
                      className={styles.activityBarFill}
                      style={{ width: `${(totalMins / maxMins) * 100}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
          )
        ) : (
          isLoadingAbsences ? (
            <div className={styles.emptyState}>
              <h3 className={styles.emptyTitle}>Laden...</h3>
              <p className={styles.emptyDescription}>Abwesenheiten werden geladen.</p>
            </div>
          ) : absences.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'📅'}</div>
              <h3 className={styles.emptyTitle}>Keine Abwesenheiten</h3>
              <p className={styles.emptyDescription}>Abwesenheiten werden hier angezeigt, sobald sie erfasst werden.</p>
            </div>
          ) : (
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Mitarbeiter</th>
                  <th>Typ</th>
                  <th>Von</th>
                  <th>Bis</th>
                  <th>Tage</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {absences.map((abs: AbsenceEntry) => (
                  <tr key={abs.id}>
                    <td className={styles.employeeName}>{abs.employee}</td>
                    <td>{abs.type}</td>
                    <td className={styles.dateCell}>{formatDate(abs.from)}</td>
                    <td className={styles.dateCell}>{formatDate(abs.to)}</td>
                    <td>{abs.days}</td>
                    <td><span className={`${styles.badge} ${getStatusClass(abs.status)}`}>{abs.status}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}
      </div>

      {/* Time Entry Modal */}
      {showModal && (
        <div className={styles.modalOverlay} onClick={() => setShowModal(false)}>
          <div className={styles.modal} onClick={e => e.stopPropagation()}>
            <div className={styles.modalHeader}>
              <h2 className={styles.modalTitle}>Zeit erfassen</h2>
              <button className={styles.modalClose} onClick={() => setShowModal(false)}>&times;</button>
            </div>
            <div className={styles.modalBody}>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Datum</label>
                <input type="date" className={styles.formInput} defaultValue={new Date().toISOString().split('T')[0]} />
              </div>
              <div className={styles.formRow}>
                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Start</label>
                  <input type="time" className={styles.formInput} defaultValue="08:00" />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.formLabel}>Ende</label>
                  <input type="time" className={styles.formInput} defaultValue="17:00" />
                </div>
              </div>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Projekt</label>
                <select className={styles.formSelect}>
                  <option value="">Projekt auswählen...</option>
                  <option>Stadtfest München 2026</option>
                  <option>Firmen-Gala TechCorp</option>
                  <option>Open Air Festival Bodensee</option>
                </select>
              </div>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Aktivität</label>
                <select className={styles.formSelect}>
                  <option value="">Aktivität auswählen...</option>
                  <option>Aufbau</option>
                  <option>Abbau</option>
                  <option>Transport</option>
                  <option>Planung</option>
                  <option>Licht-Programmierung</option>
                  <option>Ton-Check</option>
                  <option>Vorbereitung</option>
                </select>
              </div>
              <div className={styles.formGroup}>
                <label className={styles.formLabel}>Notizen</label>
                <textarea className={styles.formTextarea} rows={2} placeholder="Optionale Notizen..." />
              </div>
            </div>
            <div className={styles.modalFooter}>
              <button className={styles.btnSecondary} onClick={() => setShowModal(false)}>Abbrechen</button>
              <button className={styles.btnPrimary} onClick={() => setShowModal(false)}>Speichern</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default TimeTrackingPage
