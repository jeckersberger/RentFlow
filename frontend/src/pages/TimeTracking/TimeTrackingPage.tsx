import { useState, useMemo } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { timeTrackingApi } from '../../services/api'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import './TimeTracking.scss'

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
  const [formDate, setFormDate] = useState(new Date().toISOString().split('T')[0])
  const [formStart, setFormStart] = useState('08:00')
  const [formEnd, setFormEnd] = useState('17:00')
  const [formNotes, setFormNotes] = useState('')
  const [isSaving, setIsSaving] = useState(false)
  const queryClient = useQueryClient()

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
      case 'Offen': return 'statusOpen'
      case 'Genehmigt': return 'statusApproved'
      case 'Ausstehend': return 'statusPending'
      case 'Abgelehnt': return 'statusRejected'
      default: return ''
    }
  }

  // Extract unique activities
  const activities: string[] = Array.from(new Set(timeEntries.map((e: TimeEntry) => e.activity))) as string[]

  return (
    <div className="page">
      <div className="header">
        <div>
          <h1 className="title">Zeiterfassung</h1>
          <p className="subtitle">Arbeitszeiten erfassen und auswerten</p>
        </div>
        <div className="headerActions">
          <button className="btnPrimary" onClick={() => setShowModal(true)}>
            + Zeit erfassen
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="statsGrid">
        <div className="statCard">
          <div className="statLabel">Stunden diese Woche</div>
          <div className="statValue">{Math.floor(totalWeekHours / 60)}:{String(totalWeekHours % 60).padStart(2, '0')}</div>
        </div>
        <div className={`statCard ${openCount > 0 ? 'statCardWarning' : ''}`}>
          <div className="statLabel">Offene Eintraege</div>
          <div className="statValue">{openCount}</div>
        </div>
        <div className="statCard">
          <div className="statLabel">Mitarbeiter aktiv</div>
          <div className="statValue">{[...new Set(timeEntries.map((e: TimeEntry) => e.employee))].length}</div>
        </div>
        <div className="statCard">
          <div className="statLabel">Projekte diese Woche</div>
          <div className="statValue">{[...new Set(timeEntries.map((e: TimeEntry) => e.project))].length}</div>
        </div>
      </div>

      {/* Weekly Summary Bar Chart */}
      <div className="weeklyCard">
        <h3 className="weeklyTitle">Wochenübersicht</h3>
        <div className="weeklyChart">
          {weeklySummary.map((day, idx) => (
            <div key={idx} className="weeklyBar">
              <div className="barContainer">
                <div
                  className="bar"
                  style={{ height: `${day.totalMinutes > 0 ? (day.totalMinutes / maxMinutes) * 100 : 0}%` }}
                />
              </div>
              <div className="barLabel">{day.label}</div>
              <div className="barValue">
                {day.hours > 0 ? `${day.hours}:${String(day.minutes).padStart(2, '0')}` : '-'}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Tabs */}
      <div className="tabBar">
        {tabs.map(tab => (
          <button
            key={tab.key}
            className={`tab ${activeTab === tab.key ? 'tabActive' : ''}`}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Filter */}
      <div className="filterBar">
        <input
          type="text"
          className="searchInput"
          placeholder="Mitarbeiter oder Projekt suchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content */}
      <div className="tableCard">
        {activeTab === 'hours' ? (
          isLoadingEntries ? (
            <SkeletonTable rows={5} columns={5} />
          ) : filteredEntries.length === 0 ? (
            <div className="emptyState">
              <div className="emptyIcon">{'⏰'}</div>
              <h3 className="emptyTitle">Keine Zeiteintraege</h3>
              <p className="emptyDescription">
                {searchQuery ? 'Keine Eintraege gefunden. Versuchen Sie eine andere Suche.' : 'Erfassen Sie Arbeitszeiten fuer Ihr Team.'}
              </p>
              {!searchQuery && <button className="btnPrimary" onClick={() => setShowModal(true)}>Zeit erfassen</button>}
            </div>
          ) : (
            <table className="table">
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
                    <td className="dateCell">{formatDate(entry.date)}</td>
                    <td className="employeeName">{entry.employee}</td>
                    <td><span className="projectTag">{entry.project}</span></td>
                    <td>{entry.activity}</td>
                    <td className="timeCell">{entry.start}</td>
                    <td className="timeCell">{entry.end}</td>
                    <td className="durationCell">{entry.duration}</td>
                    <td><span className={`badge ${getStatusClass(entry.status)}`}>{entry.status}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        ) : activeTab === 'activities' ? (
          isLoadingEntries ? (
            <SkeletonTable rows={5} columns={4} />
          ) : activities.length === 0 ? (
            <div className="emptyState">
              <div className="emptyIcon">{'📊'}</div>
              <h3 className="emptyTitle">Keine Aktivitaeten</h3>
              <p className="emptyDescription">Aktivitaeten werden angezeigt, sobald Zeiteintraege erfasst werden.</p>
            </div>
          ) : (
          <div className="activitiesList">
            {activities.map((act, idx) => {
              const actEntries = timeEntries.filter((e: TimeEntry) => e.activity === act)
              const totalMins = actEntries.reduce((sum: number, e: TimeEntry) => {
                const [h, m] = e.duration.split(':').map(Number)
                return sum + (isNaN(h) ? 0 : h) * 60 + (isNaN(m) ? 0 : m)
              }, 0)
              const maxMins = Math.max(...activities.map(a => timeEntries.filter((e: TimeEntry) => e.activity === a).reduce((s: number, e: TimeEntry) => { const [h, m] = e.duration.split(':').map(Number); return s + (isNaN(h) ? 0 : h) * 60 + (isNaN(m) ? 0 : m); }, 0)), 1)
              return (
                <div key={idx} className="activityItem">
                  <div className="activityName">{act}</div>
                  <div className="activityMeta">
                    {actEntries.length} Eintraege &bull; {Math.floor(totalMins / 60)}:{String(totalMins % 60).padStart(2, '0')} Stunden
                  </div>
                  <div className="activityBar">
                    <div
                      className="activityBarFill"
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
            <SkeletonTable rows={5} columns={4} />
          ) : absences.length === 0 ? (
            <div className="emptyState">
              <div className="emptyIcon">{'📅'}</div>
              <h3 className="emptyTitle">Keine Abwesenheiten</h3>
              <p className="emptyDescription">Abwesenheiten werden hier angezeigt, sobald sie erfasst werden.</p>
            </div>
          ) : (
            <table className="table">
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
                    <td className="employeeName">{abs.employee}</td>
                    <td>{abs.type}</td>
                    <td className="dateCell">{formatDate(abs.from)}</td>
                    <td className="dateCell">{formatDate(abs.to)}</td>
                    <td>{abs.days}</td>
                    <td><span className={`badge ${getStatusClass(abs.status)}`}>{abs.status}</span></td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        )}
      </div>

      {/* Time Entry Modal */}
      {showModal && (
        <div className="modalOverlay" onClick={() => setShowModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modalHeader">
              <h2 className="modalTitle">Zeit erfassen</h2>
              <button className="modalClose" onClick={() => setShowModal(false)}>&times;</button>
            </div>
            <div className="modalBody">
              <div className="formGroup">
                <label className="formLabel">Datum</label>
                <input type="date" className="formInput" value={formDate} onChange={e => setFormDate(e.target.value)} />
              </div>
              <div className="formRow">
                <div className="formGroup">
                  <label className="formLabel">Start</label>
                  <input type="time" className="formInput" value={formStart} onChange={e => setFormStart(e.target.value)} />
                </div>
                <div className="formGroup">
                  <label className="formLabel">Ende</label>
                  <input type="time" className="formInput" value={formEnd} onChange={e => setFormEnd(e.target.value)} />
                </div>
              </div>
              <div className="formGroup">
                <label className="formLabel">Projekt</label>
                <select className="formSelect">
                  <option value="">Projekt auswählen...</option>
                  <option>Stadtfest München 2026</option>
                  <option>Firmen-Gala TechCorp</option>
                  <option>Open Air Festival Bodensee</option>
                </select>
              </div>
              <div className="formGroup">
                <label className="formLabel">Aktivität</label>
                <select className="formSelect">
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
              <div className="formGroup">
                <label className="formLabel">Notizen</label>
                <textarea className="formTextarea" rows={2} placeholder="Optionale Notizen..." value={formNotes} onChange={e => setFormNotes(e.target.value)} />
              </div>
            </div>
            <div className="modalFooter">
              <button className="btnSecondary" onClick={() => setShowModal(false)}>Abbrechen</button>
              <button className="btnPrimary" disabled={isSaving} onClick={async () => {
                setIsSaving(true)
                try {
                  await timeTrackingApi.createEntry({
                    date: formDate,
                    start_time: formStart,
                    end_time: formEnd,
                    notes: formNotes,
                  })
                  queryClient.invalidateQueries({ queryKey: ['time-entries'] })
                  setShowModal(false)
                  setFormNotes('')
                } catch (err) {
                  console.error('Failed to save time entry', err)
                } finally {
                  setIsSaving(false)
                }
              }}>{isSaving ? 'Speichert...' : 'Speichern'}</button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default TimeTrackingPage
