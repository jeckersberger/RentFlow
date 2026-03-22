import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { crewApi } from '../../services/api'
import type { CrewMember } from '../../types/crew'
import './Crew.module.scss'

// Map backend CrewMemberDTO to frontend CrewMember type
function mapBackendCrewMember(dto: any): CrewMember {
  const statusMap: Record<string, string> = {
    active: 'available',
    busy: 'busy',
    on_leave: 'on_leave',
    inactive: 'on_leave',
  }
  return {
    id: dto.id,
    name: `${dto.first_name} ${dto.last_name}`,
    email: dto.email || '',
    phone: dto.phone || '',
    role: dto.role || 'technician',
    availability: (statusMap[dto.status] || 'available') as CrewMember['availability'],
    qualifications: [],
    hours_this_week: 0,
    current_assignment: undefined,
    joined_date: dto.created_at || '',
  }
}

function CrewPage() {
  const navigate = useNavigate()
  const [filterRole, setFilterRole] = useState<string>('all')
  const [filterAvailability, setFilterAvailability] = useState<string>('all')

  const { data: crewMembers = [], isLoading, error } = useQuery({
    queryKey: ['crew'],
    queryFn: async () => {
      const result = await crewApi.listMembers({ page: 1, per_page: 100 })
      const items = Array.isArray(result) ? result : (result.data || [])
      return items.map(mapBackendCrewMember)
    },
    staleTime: 1000 * 60 * 5,
  })

  const activeMembers = crewMembers.filter((m: any) => m.availability === 'available')
  const busyMembers = crewMembers.filter((m: any) => m.availability === 'busy')
  const totalHoursThisWeek = crewMembers.reduce((sum: number, m: any) => sum + m.hours_this_week, 0)

  const filteredMembers = crewMembers.filter((member: any) => {
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

  if (isLoading) {
    return (
      <div className="crew-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Crew Management</h1>
            <p className="page-subtitle">Daten werden geladen...</p>
          </div>
        </div>
        <div className="stats-grid">
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="stat-card">
              <div className="stat-card__label">Laden...</div>
              <div className="stat-card__value">--</div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="crew-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Crew Management</h1>
            <p className="page-subtitle">Fehler beim Laden der Daten</p>
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
          {filteredMembers.map((member: any) => {
            const expiredQualifications = member.qualifications.filter((q: any) => q.is_expired).length
            const expiringQualifications = member.qualifications.filter(
              (q: any) => !q.is_expired && new Date(q.expiry_date) < new Date(Date.now() + 30 * 24 * 60 * 60 * 1000)
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
