import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams, useNavigate } from 'react-router-dom'
import type { CrewMember, TimeRecord, Assignment } from '../../types/crew'
import './Crew.module.scss'

const mockCrewMembers: Record<string, CrewMember> = {
  '1': {
    id: '1',
    name: 'Thomas Müller',
    email: 'thomas.mueller@rentflow.de',
    phone: '+49 89 123456',
    role: 'technician',
    availability: 'available',
    qualifications: [
      { id: 'q1', type: 'IPAF', name: 'IPAF 1a/1b/3a/3b', expiry_date: '2026-12-31', is_expired: false },
      { id: 'q2', type: 'electrical_cert', name: 'Electrical Safety', expiry_date: '2025-06-30', is_expired: false },
      { id: 'q3', type: 'first_aid', name: 'First Aid', expiry_date: '2025-03-15', is_expired: true },
    ],
    hours_this_week: 32,
    current_assignment: 'Stadtfest München 2026',
    joined_date: '2020-01-15',
  },
}

const mockTimeRecords: TimeRecord[] = [
  {
    id: 'tr1',
    crew_member_id: '1',
    start_time: '2026-03-22T08:00:00Z',
    end_time: '2026-03-22T16:30:00Z',
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    duration_hours: 8.5,
    status: 'completed',
  },
  {
    id: 'tr2',
    crew_member_id: '1',
    start_time: '2026-03-23T07:00:00Z',
    end_time: undefined,
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    duration_hours: 2.25,
    status: 'active',
  },
]

const mockAssignments: Assignment[] = [
  {
    id: 'a1',
    crew_member_id: '1',
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    start_date: '2026-03-22',
    end_date: '2026-03-24',
    status: 'in_progress',
  },
  {
    id: 'a2',
    crew_member_id: '1',
    project_id: 'p2',
    project_name: 'Firmen-Gala TechCorp',
    start_date: '2026-04-10',
    end_date: '2026-04-10',
    status: 'scheduled',
  },
]

function CrewDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [timerRunning, setTimerRunning] = useState(false)
  const [elapsedSeconds, setElapsedSeconds] = useState(0)

  const { data: member } = useQuery({
    queryKey: ['crew', id],
    queryFn: async () => mockCrewMembers[id || '1'] || mockCrewMembers['1'],
    staleTime: 1000 * 60 * 5,
  })

  const { data: timeRecords = mockTimeRecords } = useQuery({
    queryKey: ['timeRecords', id],
    queryFn: async () => mockTimeRecords,
    staleTime: 1000 * 60,
  })

  const { data: assignments = mockAssignments } = useQuery({
    queryKey: ['assignments', id],
    queryFn: async () => mockAssignments,
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

  if (!member) {
    return <div className="empty-state">Mitarbeiter nicht gefunden</div>
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

  const activeAssignments = assignments.filter(a => ['scheduled', 'in_progress'].includes(a.status))
  const _completedAssignments = assignments.filter(a => a.status === 'completed')
  void _completedAssignments // reserved for completed assignments tab

  return (
    <div className="crew-detail-page">
      <div className="page-header">
        <div>
          <button
            className="btn btn--sm btn--secondary"
            onClick={() => navigate('/crew')}
            style={{ marginBottom: 'var(--spacing-3)' }}
          >
            ← Zurück
          </button>
          <h1 className="page-title">{member.name}</h1>
          <p className="page-subtitle">{member.role.replace(/_/g, ' ')}</p>
        </div>
      </div>

      <div className="detail-grid">
        {/* Main Content */}
        <div className="detail-column">
          {/* Member Info Card */}
          <div className="detail-card">
            <h2 className="detail-card__title">👤 Persönliche Informationen</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <label className="detail-card__row-label">E-Mail</label>
                <span className="detail-card__row-value">{member.email}</span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Telefon</label>
                <span className="detail-card__row-value">{member.phone}</span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Eingestellt am</label>
                <span className="detail-card__row-value">
                  {new Date(member.joined_date).toLocaleDateString('de-DE')}
                </span>
              </div>
              <div className="detail-card__row">
                <label className="detail-card__row-label">Verfügbarkeit</label>
                <span className="detail-card__row-value">
                  <span className={`availability-badge availability-badge--${member.availability}`}>
                    {member.availability === 'available' && '✓ Verfügbar'}
                    {member.availability === 'busy' && '🔄 Beschäftigt'}
                    {member.availability === 'on_leave' && '🏖️ Urlaub'}
                  </span>
                </span>
              </div>
            </div>
          </div>

          {/* Qualifications Card */}
          <div className="detail-card">
            <h2 className="detail-card__title">📜 Zertifizierungen</h2>
            <div className="qualifications-list">
              {member.qualifications.length === 0 ? (
                <p className="qualifications-list__empty">Keine Zertifizierungen vorhanden</p>
              ) : (
                member.qualifications.map(qual => {
                  const status = getQualificationStatus(qual.expiry_date, qual.is_expired)
                  return (
                    <div key={qual.id} className={`qualification-item qualification-item--${status}`}>
                      <div className="qualification-item__header">
                        <h4 className="qualification-item__name">
                          {status === 'valid' && '✓ '}
                          {status === 'expiring' && '⚠️ '}
                          {status === 'expired' && '✕ '}
                          {qual.name}
                        </h4>
                        <span className={`qual-status qual-status--${status}`}>
                          {status === 'valid' && 'Gültig'}
                          {status === 'expiring' && 'Läuft ab'}
                          {status === 'expired' && 'Abgelaufen'}
                        </span>
                      </div>
                      <p className="qualification-item__expiry">
                        Verfällt: {new Date(qual.expiry_date).toLocaleDateString('de-DE')}
                      </p>
                    </div>
                  )
                })
              )}
            </div>
          </div>

          {/* Current/Upcoming Assignments */}
          <div className="detail-card">
            <h2 className="detail-card__title">📅 Aktuelle & kommende Einsätze</h2>
            <div className="assignments-list">
              {activeAssignments.length === 0 ? (
                <p className="assignments-list__empty">Keine aktiven Einsätze</p>
              ) : (
                activeAssignments.map(assignment => (
                  <div key={assignment.id} className="assignment-item">
                    <h4 className="assignment-item__title">{assignment.project_name}</h4>
                    <p className="assignment-item__dates">
                      {new Date(assignment.start_date).toLocaleDateString('de-DE')} bis{' '}
                      {new Date(assignment.end_date).toLocaleDateString('de-DE')}
                    </p>
                    <span className={`assignment-status assignment-status--${assignment.status}`}>
                      {assignment.status === 'scheduled' && '📌 Geplant'}
                      {assignment.status === 'in_progress' && '🔄 Läuft'}
                    </span>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="detail-sidebar">
          {/* Timer Card */}
          <div className="detail-card">
            <h2 className="detail-card__title">⏱️ Zeit-Tracker</h2>
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
                {timerRunning ? '⏹️ Stoppen' : '▶️ Starten'}
              </button>
            </div>
          </div>

          {/* Recent Time Records */}
          <div className="detail-card">
            <h2 className="detail-card__title">📊 Zeit-Einträge</h2>
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
