import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { Assignment, TimeRecord } from '../../types/crew'
import './Crew.module.scss'

const mockAssignments: Assignment[] = [
  {
    id: 'a1',
    crew_member_id: 'current-user',
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    start_date: '2026-03-25',
    end_date: '2026-03-27',
    status: 'in_progress',
  },
  {
    id: 'a2',
    crew_member_id: 'current-user',
    project_id: 'p2',
    project_name: 'Corporate Event TechCorp',
    start_date: '2026-03-28',
    end_date: '2026-03-29',
    status: 'scheduled',
  },
  {
    id: 'a3',
    crew_member_id: 'current-user',
    project_id: 'p3',
    project_name: 'Open Air Festival Bodensee',
    start_date: '2026-04-05',
    end_date: '2026-04-07',
    status: 'scheduled',
  },
  {
    id: 'a4',
    crew_member_id: 'current-user',
    project_id: 'p4',
    project_name: 'Wedding Reception Setup',
    start_date: '2026-04-10',
    end_date: '2026-04-10',
    status: 'scheduled',
  },
]

const mockTimeRecords: TimeRecord[] = [
  {
    id: 'tr1',
    crew_member_id: 'current-user',
    start_time: '2026-03-22T08:00:00Z',
    end_time: undefined,
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    duration_hours: 0,
    status: 'active',
  },
  {
    id: 'tr2',
    crew_member_id: 'current-user',
    start_time: '2026-03-21T07:00:00Z',
    end_time: '2026-03-21T15:30:00Z',
    project_id: 'p1',
    project_name: 'Stadtfest München 2026',
    duration_hours: 8.5,
    status: 'completed',
  },
]

function MyAssignments() {
  const [filterStatus, setFilterStatus] = useState<string>('all')

  const { data: assignments = mockAssignments } = useQuery({
    queryKey: ['my-assignments'],
    queryFn: async () => mockAssignments,
    staleTime: 1000 * 60 * 5,
  })

  const { data: timeRecords = mockTimeRecords } = useQuery({
    queryKey: ['my-time-records'],
    queryFn: async () => mockTimeRecords,
    staleTime: 1000 * 60 * 1,
  })

  const activeTimeRecord = timeRecords.find(r => r.status === 'active')
  const completedToday = timeRecords.filter(
    r => r.status === 'completed' &&
    new Date(r.start_time).toDateString() === new Date().toDateString()
  ).length

  const totalHoursToday = timeRecords
    .filter(r => new Date(r.start_time).toDateString() === new Date().toDateString())
    .reduce((sum, r) => sum + r.duration_hours, 0)

  const filteredAssignments = assignments.filter(a =>
    filterStatus === 'all' || a.status === filterStatus
  )

  const getStatusLabel = (status: string): string => {
    const labels: Record<string, string> = {
      scheduled: 'Geplant',
      in_progress: 'Laufend',
      completed: 'Abgeschlossen',
    }
    return labels[status] || status
  }

  const getStatusColor = (status: string): string => {
    const colors: Record<string, string> = {
      scheduled: 'var(--color-primary)',
      in_progress: 'var(--color-success)',
      completed: 'var(--color-text-secondary)',
    }
    return colors[status] || 'var(--color-text-secondary)'
  }

  const formatDate = (dateStr: string): string => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('de-DE', { weekday: 'short', month: 'short', day: 'numeric' })
  }

  const formatTime = (dateStr: string): string => {
    const date = new Date(dateStr)
    return date.toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })
  }

  const getElapsedTime = (startTime: string): string => {
    const now = new Date()
    const start = new Date(startTime)
    const diffMs = now.getTime() - start.getTime()
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60))
    const diffMins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
    return `${diffHours}h ${diffMins}m`
  }

  return (
    <div className="crew-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Meine Einsätze</h1>
          <p className="page-subtitle">Deine anstehenden Aufgaben und Zeiten</p>
        </div>
      </div>

      {/* Time Record Status Card */}
      {activeTimeRecord && (
        <div style={{
          background: 'rgba(16, 185, 129, 0.1)',
          border: '2px solid var(--color-success)',
          borderRadius: 'var(--radius-card)',
          padding: 'var(--spacing-4)',
          marginBottom: 'var(--spacing-4)',
        }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            gap: 'var(--spacing-3)',
          }}>
            <div>
              <div style={{
                fontSize: 'var(--font-size-xs)',
                fontWeight: 'var(--font-weight-semibold)',
                color: 'var(--color-success)',
                textTransform: 'uppercase',
                marginBottom: 'var(--spacing-1)',
              }}>⏱️ Aktive Zeiterfassung</div>
              <div style={{
                fontSize: 'var(--font-size-base)',
                fontWeight: 'var(--font-weight-semibold)',
                color: 'var(--color-text-primary)',
              }}>
                {activeTimeRecord.project_name}
              </div>
              <div style={{
                fontSize: 'var(--font-size-sm)',
                color: 'var(--color-text-secondary)',
                marginTop: 'var(--spacing-1)',
              }}>
                Gestartet: {formatTime(activeTimeRecord.start_time)} ({getElapsedTime(activeTimeRecord.start_time)})
              </div>
            </div>
            <button
              className="btn btn--primary"
              onClick={() => alert('Zeit stoppen...')}
              style={{ padding: 'var(--spacing-2) var(--spacing-4)' }}
            >
              ⏹️ Stopp
            </button>
          </div>
        </div>
      )}

      {/* Daily Stats */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Heute erfasst</div>
          <div className="stat-card__value">{totalHoursToday.toFixed(1)}h</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Zeiteinträge heute</div>
          <div className="stat-card__value">{completedToday}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Nächster Einsatz</div>
          <div className="stat-card__value" style={{ fontSize: 'var(--font-size-base)' }}>
            {assignments.length > 0 ? formatDate(assignments[0].start_date) : '—'}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Aktive Einsätze</div>
          <div className="stat-card__value">
            {assignments.filter(a => a.status === 'in_progress').length}
          </div>
        </div>
      </div>

      {/* Filter */}
      <div className="filters">
        <div className="filter-group">
          <label htmlFor="status-filter" className="filter-label">Status:</label>
          <select
            id="status-filter"
            className="filter-select"
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
          >
            <option value="all">Alle</option>
            <option value="scheduled">Geplant</option>
            <option value="in_progress">Laufend</option>
            <option value="completed">Abgeschlossen</option>
          </select>
        </div>
      </div>

      {/* Assignments List */}
      {filteredAssignments.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state__icon">📅</div>
          <h3 className="empty-state__title">Keine Einsätze gefunden</h3>
          <p className="empty-state__description">Du hast derzeit keine Einsätze mit diesem Status.</p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
          {filteredAssignments.map((assignment) => (
            <div
              key={assignment.id}
              style={{
                background: 'var(--color-bg-secondary)',
                border: `2px solid ${assignment.status === 'in_progress' ? 'var(--color-success)' : 'var(--color-border)'}`,
                borderRadius: 'var(--radius-card)',
                padding: 'var(--spacing-4)',
                display: 'flex',
                flexDirection: 'column',
                gap: 'var(--spacing-3)',
                transition: 'all var(--transition-normal)',
                cursor: 'pointer',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.boxShadow = 'var(--shadow-card-hover)'
                e.currentTarget.style.borderColor = 'var(--color-border-strong)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.boxShadow = 'none'
                e.currentTarget.style.borderColor = assignment.status === 'in_progress' ? 'var(--color-success)' : 'var(--color-border)'
              }}
            >
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'flex-start',
                gap: 'var(--spacing-3)',
              }}>
                <div>
                  <h3 style={{
                    fontSize: 'var(--font-size-base)',
                    fontWeight: 'var(--font-weight-semibold)',
                    color: 'var(--color-text-primary)',
                    margin: '0 0 var(--spacing-1) 0',
                  }}>
                    {assignment.project_name}
                  </h3>
                  <p style={{
                    fontSize: 'var(--font-size-sm)',
                    color: 'var(--color-text-secondary)',
                    margin: 0,
                  }}>
                    📅 {formatDate(assignment.start_date)} bis {formatDate(assignment.end_date)}
                  </p>
                </div>
                <span style={{
                  display: 'inline-flex',
                  alignItems: 'center',
                  gap: 'var(--spacing-2)',
                  padding: 'var(--spacing-2) var(--spacing-3)',
                  borderRadius: 'var(--radius-full)',
                  fontSize: 'var(--font-size-xs)',
                  fontWeight: 'var(--font-weight-semibold)',
                  backgroundColor: `${getStatusColor(assignment.status)}20`,
                  color: getStatusColor(assignment.status),
                  textTransform: 'uppercase',
                  whiteSpace: 'nowrap',
                }}>
                  {assignment.status === 'in_progress' && '🔴'}
                  {assignment.status === 'scheduled' && '⏰'}
                  {assignment.status === 'completed' && '✅'}
                  {getStatusLabel(assignment.status)}
                </span>
              </div>
              <button
                className="btn btn--sm btn--primary"
                onClick={() => alert(`Einsatzdetails für ${assignment.project_name} werden angezeigt...`)}
              >
                📋 Details ansehen
              </button>
            </div>
          ))}
        </div>
      )}

      {/* Time Records Section */}
      <div style={{ marginTop: 'var(--spacing-6)' }}>
        <h2 style={{
          fontSize: 'var(--font-size-lg)',
          fontWeight: 'var(--font-weight-semibold)',
          color: 'var(--color-text-primary)',
          margin: '0 0 var(--spacing-3) 0',
        }}>
          ⏱️ Letzte Zeiteinträge
        </h2>
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          {timeRecords.slice(0, 5).map((record) => (
            <div
              key={record.id}
              style={{
                background: 'var(--color-bg-secondary)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius-md)',
                padding: 'var(--spacing-3)',
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                gap: 'var(--spacing-3)',
              }}
            >
              <div style={{ flex: 1 }}>
                <p style={{
                  fontSize: 'var(--font-size-sm)',
                  fontWeight: 'var(--font-weight-medium)',
                  color: 'var(--color-text-primary)',
                  margin: '0 0 var(--spacing-1) 0',
                }}>
                  {record.project_name}
                </p>
                <p style={{
                  fontSize: 'var(--font-size-xs)',
                  color: 'var(--color-text-secondary)',
                  margin: 0,
                }}>
                  {formatTime(record.start_time)} - {record.end_time ? formatTime(record.end_time) : 'Laufend'}
                </p>
              </div>
              <div style={{
                fontSize: 'var(--font-size-sm)',
                fontWeight: 'var(--font-weight-semibold)',
                color: record.status === 'active' ? 'var(--color-success)' : 'var(--color-text-primary)',
              }}>
                {record.status === 'active' ? '⏱️ Aktiv' : `${record.duration_hours}h`}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export default MyAssignments
