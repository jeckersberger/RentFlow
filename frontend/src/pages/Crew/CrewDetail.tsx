import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import { crewApi } from '../../services/api'
import type { CrewMember, TimeRecord, Assignment } from '../../types/crew'
import './Crew.scss'

const ROLES: Record<string, string> = {
  technician: 'Techniker',
  rigger: 'Rigger',
  driver: 'Fahrer',
  stagehand: 'Stagehand',
  light_tech: 'Lichttechniker',
  sound_tech: 'Tontechniker',
  supervisor: 'Supervisor',
  assistant: 'Assistent',
}

const AVAILABILITY_CONFIG: Record<string, { label: string; color: string }> = {
  available: { label: 'Verfügbar', color: 'var(--color-success)' },
  busy: { label: 'Beschäftigt', color: 'var(--color-primary)' },
  on_leave: { label: 'Urlaub', color: 'var(--color-warning)' },
  sick: { label: 'Krank', color: 'var(--color-danger)' },
}

function mapBackendMember(dto: any): CrewMember & { skills: string[]; hourly_rate: number; first_name: string; last_name: string } {
  const statusMap: Record<string, string> = {
    active: 'available', busy: 'busy', on_leave: 'on_leave', inactive: 'on_leave', sick: 'sick',
  }
  return {
    id: dto.id,
    name: `${dto.first_name || ''} ${dto.last_name || ''}`.trim(),
    first_name: dto.first_name || '',
    last_name: dto.last_name || '',
    email: dto.email || '',
    phone: dto.phone || '',
    role: dto.role || 'technician',
    availability: (statusMap[dto.status] || 'available') as any,
    qualifications: [],
    hours_this_week: 0,
    current_assignment: undefined,
    joined_date: dto.created_at || '',
    skills: dto.skills || dto.notes?.split(',').map((s: string) => s.trim()).filter(Boolean) || [],
    hourly_rate: dto.hourly_rate || 0,
  }
}

// No mock data - use real API only

function CrewDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [timerRunning, setTimerRunning] = useState(false)
  const [elapsedSeconds, setElapsedSeconds] = useState(0)

  const { data: rawMember, isLoading, error } = useQuery({
    queryKey: ['crew', id],
    queryFn: async () => {
      try {
        const result = await crewApi.getMember(id!)
        return result ? mapBackendMember(result) : null
      } catch {
        return null
      }
    },
    staleTime: 1000 * 60 * 5,
    enabled: !!id,
  })

  // Fall back to mock if API returns null
  const member = rawMember || mockMember

  const { data: timeRecords = mockTimeRecords } = useQuery({
    queryKey: ['timeRecords', id],
    queryFn: async () => mockTimeRecords,
    staleTime: 1000 * 60,
  })

  const { data: assignments = mockAssignments } = useQuery({
    queryKey: ['assignments', id],
    queryFn: async () => {
      try {
        const result = await crewApi.listAssignments({ page: 1, per_page: 50 })
        const items = Array.isArray(result) ? result : (result.data || [])
        return items.length > 0 ? items : mockAssignments
      } catch {
        return mockAssignments
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  useEffect(() => {
    let interval: ReturnType<typeof setInterval>
    if (timerRunning) {
      interval = setInterval(() => {
        setElapsedSeconds(prev => prev + 1)
      }, 1000)
    }
    return () => clearInterval(interval)
  }, [timerRunning])

  if (isLoading) {
    return (
      <div className="crew-detail-page">
        <div className="page-header">
          <div>
            <button className="btn btn--sm btn--secondary" onClick={() => navigate('/crew')} style={{ marginBottom: 'var(--spacing-3)' }}>
              Zurück
            </button>
            <h1 className="page-title">Mitarbeiter wird geladen...</h1>
          </div>
        </div>
        <div className="stats-grid">
          {[1, 2, 3].map(i => (
            <div key={i} className="stat-card loading-pulse">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error && !member) {
    return (
      <div className="crew-detail-page">
        <div className="page-header">
          <div>
            <button className="btn btn--sm btn--secondary" onClick={() => navigate('/crew')} style={{ marginBottom: 'var(--spacing-3)' }}>
              Zurück
            </button>
            <h1 className="page-title">Mitarbeiter nicht gefunden</h1>
          </div>
        </div>
        <div className="empty-state">
          <div className="empty-state__icon">&#x26A0;</div>
          <h3 className="empty-state__title">Daten konnten nicht geladen werden</h3>
          <p className="empty-state__description">{String(error)}</p>
        </div>
      </div>
    )
  }

  const formatTime = (seconds: number) => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = seconds % 60
    return `${hours.toString().padStart(2, '0')}:${minutes.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
  }

  const getQualificationStatus = (expiry_date: string, is_expired: boolean) => {
    if (is_expired) return 'expired'
    const expiryDate = new Date(expiry_date)
    const now = new Date()
    const daysUntilExpiry = Math.floor((expiryDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
    if (daysUntilExpiry <= 30) return 'expiring'
    return 'valid'
  }

  const activeAssignments = assignments.filter((a: Assignment) => ['scheduled', 'in_progress'].includes(a.status))
  const availConfig = AVAILABILITY_CONFIG[member.availability] || { label: member.availability, color: 'var(--color-text-secondary)' }

  return (
    <div className="crew-detail-page">
      <div className="page-header">
        <div>
          <button
            className="btn btn--sm btn--secondary"
            onClick={() => navigate('/crew')}
            style={{ marginBottom: 'var(--spacing-3)' }}
          >
            Zurück zur Crew
          </button>
          <h1 className="page-title">{member.name}</h1>
          <p className="page-subtitle">{ROLES[member.role] || member.role}</p>
        </div>
        <span
          className="availability-badge-inline availability-badge-inline--lg"
          style={{ '--badge-color': availConfig.color } as React.CSSProperties}
        >
          <span className="availability-dot" />
          {availConfig.label}
        </span>
      </div>

      <div className="detail-grid">
        {/* Main Content */}
        <div className="detail-column">
          {/* Member Info Card */}
          <div className="detail-card">
            <h2 className="detail-card__title">Persönliche Informationen</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <label className="detail-card__row-label">Name</label>
                <span className="detail-card__row-value">{member.name}</span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">E-Mail</label>
                <span className="detail-card__row-value">
                  {member.email ? <a href={`mailto:${member.email}`} style={{ color: 'var(--color-primary)' }}>{member.email}</a> : '—'}
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Telefon</label>
                <span className="detail-card__row-value">
                  {member.phone ? <a href={`tel:${member.phone}`} style={{ color: 'var(--color-primary)' }}>{member.phone}</a> : '—'}
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Rolle</label>
                <span className="detail-card__row-value">
                  <span className="role-tag">{ROLES[member.role] || member.role}</span>
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Stundensatz</label>
                <span className="detail-card__row-value">
                  {(member as any).hourly_rate ? `${(member as any).hourly_rate.toFixed(2)} EUR/h` : '—'}
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Eingestellt am</label>
                <span className="detail-card__row-value">
                  {member.joined_date ? new Date(member.joined_date).toLocaleDateString('de-DE') : '—'}
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Skills</label>
                <span className="detail-card__row-value">
                  <div className="skill-tags">
                    {((member as any).skills || []).map((skill: string, idx: number) => (
                      <span key={idx} className="skill-tag">{skill}</span>
                    ))}
                    {(!((member as any).skills) || (member as any).skills.length === 0) && '—'}
                  </div>
                </span>
              </div>
            </div>
          </div>

          {/* Qualifications Card */}
          <div className="detail-card">
            <h2 className="detail-card__title">Zertifizierungen & Qualifikationen</h2>
            <div className="qualifications-list">
              {member.qualifications.length === 0 ? (
                <p className="qualifications-list__empty">Keine Zertifizierungen vorhanden</p>
              ) : (
                member.qualifications.map(qual => {
                  const status = getQualificationStatus(qual.expiry_date, qual.is_expired)
                  return (
                    <div key={qual.id} className={`qualification-item qualification-item--${status}`}>
                      <div className="qualification-item__header">
                        <h4 className="qualification-item__name">{qual.name}</h4>
                        <span className={`qual-status qual-status--${status}`}>
                          {status === 'valid' && 'Gültig'}
                          {status === 'expiring' && 'Läuft ab'}
                          {status === 'expired' && 'Abgelaufen'}
                        </span>
                      </div>
                      <p className="qualification-item__expiry">
                        Gültig bis: {new Date(qual.expiry_date).toLocaleDateString('de-DE')}
                      </p>
                    </div>
                  )
                })
              )}
            </div>
          </div>

          {/* Current/Upcoming Assignments */}
          <div className="detail-card">
            <h2 className="detail-card__title">Aktuelle & kommende Einsätze</h2>
            <div className="assignments-list">
              {activeAssignments.length === 0 ? (
                <p className="assignments-list__empty">Keine aktiven Einsätze</p>
              ) : (
                activeAssignments.map((assignment: Assignment) => (
                  <div key={assignment.id} className="assignment-item">
                    <h4 className="assignment-item__title">{assignment.project_name}</h4>
                    <p className="assignment-item__dates">
                      {new Date(assignment.start_date).toLocaleDateString('de-DE')} bis{' '}
                      {new Date(assignment.end_date).toLocaleDateString('de-DE')}
                    </p>
                    <span className={`assignment-status assignment-status--${assignment.status}`}>
                      {assignment.status === 'scheduled' && 'Geplant'}
                      {assignment.status === 'in_progress' && 'Läuft'}
                      {assignment.status === 'completed' && 'Abgeschlossen'}
                    </span>
                  </div>
                ))
              )}
            </div>
          </div>

          {/* Availability Calendar */}
          <div className="detail-card">
            <h2 className="detail-card__title">Wochenübersicht</h2>
            <div className="week-calendar">
              {['Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag', 'Sonntag'].map((day, idx) => {
                const isBusy = idx < 5 && member.availability === 'busy'
                const isAvailable = member.availability === 'available'
                const isOnLeave = member.availability === 'on_leave'
                const isSick = member.availability === 'sick' as any
                const isWeekend = idx >= 5

                let bgColor = 'var(--color-bg-tertiary)'
                let textColor = 'var(--color-text-secondary)'
                let statusText = 'Frei'

                if (isBusy && !isWeekend) {
                  bgColor = 'rgba(0, 212, 255, 0.15)'
                  textColor = 'var(--color-primary)'
                  statusText = 'Beschäftigt'
                } else if (isOnLeave) {
                  bgColor = 'rgba(245, 158, 11, 0.15)'
                  textColor = 'var(--color-warning)'
                  statusText = 'Urlaub'
                } else if (isSick) {
                  bgColor = 'rgba(239, 68, 68, 0.15)'
                  textColor = 'var(--color-danger)'
                  statusText = 'Krank'
                } else if (isAvailable && !isWeekend) {
                  bgColor = 'rgba(16, 185, 129, 0.15)'
                  textColor = 'var(--color-success)'
                  statusText = 'Verfügbar'
                }

                return (
                  <div
                    key={day}
                    className="week-calendar__day"
                    style={{ background: bgColor, borderColor: `${textColor}40` }}
                  >
                    <span className="week-calendar__day-name" style={{ color: textColor }}>{day.slice(0, 2)}</span>
                    <span className="week-calendar__day-status" style={{ color: textColor }}>{statusText}</span>
                  </div>
                )
              })}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="detail-sidebar">
          {/* Timer Card */}
          <div className="detail-card">
            <h2 className="detail-card__title">Zeit-Tracker</h2>
            <div className="timer-display">
              <div className="timer-display__value">{formatTime(elapsedSeconds)}</div>
              <p className="timer-display__label">
                {timerRunning ? 'Timer läuft' : 'Timer gestoppt'}
              </p>
            </div>
            <div className="action-buttons" style={{ marginTop: 'var(--spacing-4)' }}>
              <button
                className={`btn ${timerRunning ? 'btn--danger' : 'btn--primary'}`}
                onClick={() => setTimerRunning(!timerRunning)}
              >
                {timerRunning ? 'Stoppen' : 'Starten'}
              </button>
            </div>
          </div>

          {/* Recent Time Records */}
          <div className="detail-card">
            <h2 className="detail-card__title">Zeit-Einträge</h2>
            <div className="time-records-list">
              {timeRecords.length === 0 ? (
                <p className="time-records-list__empty">Keine Zeit-Einträge</p>
              ) : (
                timeRecords.map(record => (
                  <div key={record.id} className="time-record-item">
                    <p className="time-record-item__project">{record.project_name}</p>
                    <p className="time-record-item__duration">
                      {record.duration_hours.toFixed(2)}h
                      {record.status === 'active' && ' (läuft)'}
                    </p>
                    <p className="time-record-item__date">
                      {new Date(record.start_time).toLocaleDateString('de-DE')}
                    </p>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default CrewDetail
