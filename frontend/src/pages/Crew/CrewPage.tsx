import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import type { CrewMember } from '../../types/crew'
import './Crew.module.scss'

const mockCrewMembers: CrewMember[] = [
  {
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
  {
    id: '2',
    name: 'Maria Schmidt',
    email: 'maria.schmidt@rentflow.de',
    phone: '+49 89 234567',
    role: 'rigger',
    availability: 'busy',
    qualifications: [
      { id: 'q4', type: 'IPAF', name: 'IPAF 1a/1b/3a/3b', expiry_date: '2026-08-20', is_expired: false },
      { id: 'q5', type: 'rope_access', name: 'Rope Access Level 1', expiry_date: '2027-01-10', is_expired: false },
    ],
    hours_this_week: 40,
    current_assignment: 'Open Air Festival Bodensee',
    joined_date: '2019-03-22',
  },
  {
    id: '3',
    name: 'Peter Weber',
    email: 'peter.weber@rentflow.de',
    phone: '+49 89 345678',
    role: 'driver',
    availability: 'available',
    qualifications: [
      { id: 'q6', type: 'forklift', name: 'Forklift Operator', expiry_date: '2026-11-05', is_expired: false },
    ],
    hours_this_week: 28,
    joined_date: '2021-05-10',
  },
  {
    id: '4',
    name: 'Anna Fischer',
    email: 'anna.fischer@rentflow.de',
    phone: '+49 89 456789',
    role: 'supervisor',
    availability: 'on_leave',
    qualifications: [
      { id: 'q7', type: 'IPAF', name: 'IPAF 1a/1b/3a/3b', expiry_date: '2025-09-14', is_expired: false },
      { id: 'q8', type: 'electrical_cert', name: 'Electrical Safety', expiry_date: '2026-07-22', is_expired: false },
      { id: 'q9', type: 'first_aid', name: 'First Aid', expiry_date: '2026-05-30', is_expired: false },
    ],
    hours_this_week: 0,
    joined_date: '2018-02-01',
  },
  {
    id: '5',
    name: 'Klaus Neumann',
    email: 'klaus.neumann@rentflow.de',
    phone: '+49 89 567890',
    role: 'technician',
    availability: 'available',
    qualifications: [
      { id: 'q10', type: 'IPAF', name: 'IPAF 1a/1b/3a/3b', expiry_date: '2025-04-28', is_expired: true },
      { id: 'q11', type: 'first_aid', name: 'First Aid', expiry_date: '2026-09-10', is_expired: false },
    ],
    hours_this_week: 35,
    current_assignment: 'Firmen-Gala TechCorp',
    joined_date: '2022-08-15',
  },
  {
    id: '6',
    name: 'Sarah Johnson',
    email: 'sarah.johnson@rentflow.de',
    phone: '+49 89 678901',
    role: 'assistant',
    availability: 'available',
    qualifications: [
      { id: 'q12', type: 'first_aid', name: 'First Aid', expiry_date: '2026-02-14', is_expired: true },
    ],
    hours_this_week: 20,
    joined_date: '2023-06-01',
  },
]

function CrewPage() {
  const navigate = useNavigate()
  const [filterRole, setFilterRole] = useState<string>('all')
  const [filterAvailability, setFilterAvailability] = useState<string>('all')

  const { data: crewMembers = mockCrewMembers } = useQuery({
    queryKey: ['crew'],
    queryFn: async () => mockCrewMembers,
    staleTime: 1000 * 60 * 5,
  })

  const activeMembers = crewMembers.filter(m => m.availability === 'available')
  const busyMembers = crewMembers.filter(m => m.availability === 'busy')
  const totalHoursThisWeek = crewMembers.reduce((sum, m) => sum + m.hours_this_week, 0)

  const filteredMembers = crewMembers.filter(member => {
    const roleMatch = filterRole === 'all' || member.role === filterRole
    const availabilityMatch = filterAvailability === 'all' || member.availability === filterAvailability
    return roleMatch && availabilityMatch
  })

  const getRoleLabel = (role: string): string => {
    const labels: Record<string, string> = {
      technician: 'Techniker',
      rigger: 'Rigger',
      driver: 'Fahrer',
      supervisor: 'Supervisor',
      assistant: 'Assistent',
    }
    return labels[role] || role
  }

  const getAvailabilityLabel = (status: string): string => {
    const labels: Record<string, string> = {
      available: 'Verfügbar',
      busy: 'Beschäftigt',
      on_leave: 'Urlaub',
    }
    return labels[status] || status
  }

  return (
    <div className="crew-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Crew Management</h1>
          <p className="page-subtitle">Verwaltung von Mitarbeitern und Einsätzen</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => navigate('/crew/new')}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          ➕ Neuer Mitarbeiter
        </button>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-card__label">Gesamt Mitarbeiter</div>
          <div className="stat-card__value">{crewMembers.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Verfügbar</div>
          <div className="stat-card__value" style={{ color: 'var(--color-success)' }}>{activeMembers.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Einsätze</div>
          <div className="stat-card__value" style={{ color: 'var(--color-warning)' }}>{busyMembers.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Stunden diese Woche</div>
          <div className="stat-card__value">{totalHoursThisWeek}</div>
        </div>
      </div>

      {/* Filters */}
      <div className="filters">
        <div className="filter-group">
          <label htmlFor="role-filter" className="filter-label">Rolle:</label>
          <select
            id="role-filter"
            className="filter-select"
            value={filterRole}
            onChange={(e) => setFilterRole(e.target.value)}
          >
            <option value="all">Alle Rollen</option>
            <option value="technician">Techniker</option>
            <option value="rigger">Rigger</option>
            <option value="driver">Fahrer</option>
            <option value="supervisor">Supervisor</option>
            <option value="assistant">Assistent</option>
          </select>
        </div>
        <div className="filter-group">
          <label htmlFor="availability-filter" className="filter-label">Verfügbarkeit:</label>
          <select
            id="availability-filter"
            className="filter-select"
            value={filterAvailability}
            onChange={(e) => setFilterAvailability(e.target.value)}
          >
            <option value="all">Alle</option>
            <option value="available">Verfügbar</option>
            <option value="busy">Beschäftigt</option>
            <option value="on_leave">Urlaub</option>
          </select>
        </div>
      </div>

      {/* Crew Members Grid */}
      {filteredMembers.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state__icon">👥</div>
          <h3 className="empty-state__title">Keine Mitarbeiter gefunden</h3>
          <p className="empty-state__description">Versuchen Sie, die Filter zu ändern.</p>
        </div>
      ) : (
        <div className="crew-grid">
          {filteredMembers.map((member) => {
            const expiredQualifications = member.qualifications.filter(q => q.is_expired).length
            const expiringQualifications = member.qualifications.filter(
              q => !q.is_expired && new Date(q.expiry_date) < new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
            ).length

            return (
              <div
                key={member.id}
                className="crew-card"
                onClick={() => navigate(`/crew/${member.id}`)}
              >
                <div className="crew-card__header">
                  <div>
                    <h3 className="crew-card__name">{member.name}</h3>
                    <p className="crew-card__role">{getRoleLabel(member.role)}</p>
                  </div>
                  <StatusBadge status={member.availability as any} />
                </div>

                <div className="crew-card__availability">
                  <span className={`availability-badge availability-badge--${member.availability}`}>
                    <span className="availability-indicator" style={{
                      display: 'inline-block',
                      width: '8px',
                      height: '8px',
                      borderRadius: '50%',
                      backgroundColor: member.availability === 'available' ? 'var(--color-success)' :
                                        member.availability === 'busy' ? 'var(--color-warning)' : 'var(--color-primary)',
                    }}></span>
                    {getAvailabilityLabel(member.availability)}
                  </span>
                </div>

                <div className="crew-card__meta">
                  <p><strong>E-Mail:</strong> {member.email}</p>
                  <p><strong>Telefon:</strong> {member.phone}</p>
                  <p><strong>Stunden diese Woche:</strong> {member.hours_this_week}h</p>
                  {member.current_assignment && (
                    <p><strong>Einsatz:</strong> {member.current_assignment}</p>
                  )}
                </div>

                {/* Weekly Calendar Indicator */}
                <div style={{
                  display: 'flex',
                  gap: 'var(--spacing-1)',
                  paddingTop: 'var(--spacing-2)',
                  borderTop: 'var(--card-border-width) solid var(--color-border)',
                }}>
                  {['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'].map((day, idx) => {
                    const isBusy = idx >= 0 && idx <= 4 && member.hours_this_week > 0 && member.availability === 'busy'
                    const isAvailable = member.availability === 'available'
                    const isOnLeave = member.availability === 'on_leave'
                    return (
                      <div
                        key={day}
                        title={`${day} - ${isBusy ? 'Beschäftigt' : isOnLeave ? 'Urlaub' : 'Frei'}`}
                        style={{
                          flex: 1,
                          aspectRatio: '1',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          borderRadius: 'var(--radius-md)',
                          fontSize: 'var(--font-size-2xs)',
                          fontWeight: 'var(--font-weight-semibold)',
                          backgroundColor: isBusy ? 'rgba(245, 158, 11, 0.2)' :
                                          isOnLeave ? 'rgba(0, 212, 255, 0.2)' :
                                          isAvailable ? 'rgba(16, 185, 129, 0.2)' : 'var(--color-bg-tertiary)',
                          color: isBusy ? 'var(--color-warning)' :
                                 isOnLeave ? 'var(--color-primary)' :
                                 isAvailable ? 'var(--color-success)' : 'var(--color-text-secondary)',
                          border: '1px solid',
                          borderColor: isBusy ? 'rgba(245, 158, 11, 0.4)' :
                                      isOnLeave ? 'rgba(0, 212, 255, 0.4)' :
                                      isAvailable ? 'rgba(16, 185, 129, 0.4)' : 'var(--color-border)',
                          cursor: 'default',
                        }}
                      >
                        {day}
                      </div>
                    )
                  })}
                </div>

                <div className="crew-card__qualifications">
                  <div className="qualifications-summary">
                    <span className="qual-badge qual-badge--valid">✓ {member.qualifications.length - expiredQualifications - expiringQualifications}</span>
                    {expiringQualifications > 0 && (
                      <span className="qual-badge qual-badge--expiring">⚠ {expiringQualifications}</span>
                    )}
                    {expiredQualifications > 0 && (
                      <span className="qual-badge qual-badge--expired">✕ {expiredQualifications}</span>
                    )}
                  </div>
                </div>

                <div className="crew-card__footer">
                  <button
                    className="btn btn--sm btn--primary"
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/crew/${member.id}`)
                    }}
                    style={{ flex: 'none' }}
                  >
                    Details
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default CrewPage
