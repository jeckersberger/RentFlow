import { useState, useMemo } from 'react'
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

const DEMO_ENTRIES: TimeEntry[] = [
  { id: '1', date: '2026-03-23', employee: 'Thomas Müller', project: 'Stadtfest München 2026', activity: 'Aufbau', start: '07:00', end: '16:00', duration: '9:00', status: 'Offen' },
  { id: '2', date: '2026-03-23', employee: 'Sarah Schmidt', project: 'Stadtfest München 2026', activity: 'Licht-Programmierung', start: '08:00', end: '17:30', duration: '9:30', status: 'Offen' },
  { id: '3', date: '2026-03-22', employee: 'Thomas Müller', project: 'Firmen-Gala TechCorp', activity: 'Planung', start: '09:00', end: '13:00', duration: '4:00', status: 'Genehmigt' },
  { id: '4', date: '2026-03-22', employee: 'Max Huber', project: 'Stadtfest München 2026', activity: 'Transport', start: '06:00', end: '14:00', duration: '8:00', status: 'Genehmigt' },
  { id: '5', date: '2026-03-21', employee: 'Sarah Schmidt', project: 'Open Air Festival Bodensee', activity: 'Vorbereitung', start: '10:00', end: '18:00', duration: '8:00', status: 'Genehmigt' },
  { id: '6', date: '2026-03-21', employee: 'Max Huber', project: 'Stadtfest München 2026', activity: 'Aufbau', start: '07:30', end: '16:30', duration: '9:00', status: 'Genehmigt' },
  { id: '7', date: '2026-03-20', employee: 'Thomas Müller', project: 'Stadtfest München 2026', activity: 'Aufbau', start: '07:00', end: '17:00', duration: '10:00', status: 'Genehmigt' },
]

const DEMO_ABSENCES: AbsenceEntry[] = [
  { id: 'abs-1', employee: 'Max Huber', type: 'Urlaub', from: '2026-04-01', to: '2026-04-10', days: 8, status: 'Genehmigt' },
  { id: 'abs-2', employee: 'Sarah Schmidt', type: 'Krank', from: '2026-03-25', to: '2026-03-26', days: 2, status: 'Genehmigt' },
  { id: 'abs-3', employee: 'Thomas Müller', type: 'Fortbildung', from: '2026-04-15', to: '2026-04-17', days: 3, status: 'Ausstehend' },
]

const WEEKDAYS = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So']

function TimeTrackingPage() {
  const [activeTab, setActiveTab] = useState<TabKey>('hours')
  const [searchQuery, setSearchQuery] = useState('')
  const [showModal, setShowModal] = useState(false)

  const filteredEntries = DEMO_ENTRIES.filter(e =>
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

      const dayEntries = DEMO_ENTRIES.filter(e => e.date === dateStr)
      const totalMinutes = dayEntries.reduce((sum, e) => {
        const [h, m] = e.duration.split(':').map(Number)
        return sum + h * 60 + m
      }, 0)

      return {
        label,
        date: day.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' }),
        hours: Math.floor(totalMinutes / 60),
        minutes: totalMinutes % 60,
        totalMinutes,
      }
    })
  }, [])

  const maxMinutes = Math.max(...weeklySummary.map(d => d.totalMinutes), 1)

  const totalWeekHours = weeklySummary.reduce((s, d) => s + d.totalMinutes, 0)
  const openCount = DEMO_ENTRIES.filter(e => e.status === 'Offen').length

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
  const activities = [...new Set(DEMO_ENTRIES.map(e => e.activity))]

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
          <div className={styles.statValue}>{[...new Set(DEMO_ENTRIES.map(e => e.employee))].length}</div>
        </div>
        <div className={styles.statCard}>
          <div className={styles.statLabel}>Projekte diese Woche</div>
          <div className={styles.statValue}>{[...new Set(DEMO_ENTRIES.map(e => e.project))].length}</div>
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
          filteredEntries.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>⏰</div>
              <h3 className={styles.emptyTitle}>Keine Zeiteinträge</h3>
              <p className={styles.emptyDescription}>Erfassen Sie Arbeitszeiten für Ihr Team.</p>
              <button className={styles.btnPrimary} onClick={() => setShowModal(true)}>Zeit erfassen</button>
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
                {filteredEntries.map(entry => (
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
          <div className={styles.activitiesList}>
            {activities.map((act, idx) => {
              const actEntries = DEMO_ENTRIES.filter(e => e.activity === act)
              const totalMins = actEntries.reduce((sum, e) => {
                const [h, m] = e.duration.split(':').map(Number)
                return sum + h * 60 + m
              }, 0)
              return (
                <div key={idx} className={styles.activityItem}>
                  <div className={styles.activityName}>{act}</div>
                  <div className={styles.activityMeta}>
                    {actEntries.length} Eintraege &bull; {Math.floor(totalMins / 60)}:{String(totalMins % 60).padStart(2, '0')} Stunden
                  </div>
                  <div className={styles.activityBar}>
                    <div
                      className={styles.activityBarFill}
                      style={{ width: `${(totalMins / Math.max(...activities.map(a => DEMO_ENTRIES.filter(e => e.activity === a).reduce((s, e) => { const [h, m] = e.duration.split(':').map(Number); return s + h * 60 + m; }, 0)))) * 100}%` }}
                    />
                  </div>
                </div>
              )
            })}
          </div>
        ) : (
          DEMO_ABSENCES.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📅</div>
              <h3 className={styles.emptyTitle}>Keine Abwesenheiten</h3>
              <p className={styles.emptyDescription}>Abwesenheiten werden hier angezeigt.</p>
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
                {DEMO_ABSENCES.map(abs => (
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
