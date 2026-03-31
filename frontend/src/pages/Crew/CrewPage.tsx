import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { crewApi } from '../../services/api'
import { Modal } from '../../components/Modal/Modal'
import { Input } from '../../components/Form/Input'
import { useNotificationStore } from '../../stores/notificationStore'
import type { CrewMember } from '../../types/crew'
import { SkeletonKPI, SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import './Crew.scss'

// Map backend CrewMemberDTO to frontend CrewMember type
function mapBackendCrewMember(dto: any): CrewMember & { skills: string[]; hourly_rate: number; first_name: string; last_name: string } {
  const statusMap: Record<string, string> = {
    active: 'available',
    busy: 'busy',
    on_leave: 'on_leave',
    inactive: 'on_leave',
    sick: 'sick',
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

const ROLES = [
  { value: 'technician', label: 'Techniker' },
  { value: 'rigger', label: 'Rigger' },
  { value: 'driver', label: 'Fahrer' },
  { value: 'sound_engineer', label: 'Tontechniker' },
  { value: 'lighting_tech', label: 'Lichttechniker' },
  { value: 'freelancer', label: 'Freelancer' },
]

const AVAILABILITY_OPTIONS = [
  { value: 'available', label: 'Verfügbar', color: 'var(--color-success)' },
  { value: 'busy', label: 'Beschäftigt', color: 'var(--color-primary)' },
  { value: 'on_leave', label: 'Urlaub', color: 'var(--color-warning)' },
  { value: 'sick', label: 'Krank', color: 'var(--color-danger)' },
]

interface NewMemberForm {
  first_name: string
  last_name: string
  email: string
  phone: string
  role: string
  hourly_rate: string
  skills: string[]
}

const emptyForm: NewMemberForm = {
  first_name: '',
  last_name: '',
  email: '',
  phone: '',
  role: 'technician',
  hourly_rate: '',
  skills: [],
}

function CrewPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [filterRole, setFilterRole] = useState<string>('all')
  const [filterAvailability, setFilterAvailability] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [formData, setFormData] = useState<NewMemberForm>(emptyForm)
  const [skillInput, setSkillInput] = useState('')

  const { data: crewMembers = [], isLoading, error } = useQuery({
    queryKey: ['crew'],
    queryFn: async () => {
      const result = await crewApi.listMembers({ page: 1, per_page: 100 })
      if (!result) return []
      const items = Array.isArray(result) ? result : (result.data || [])
      if (!Array.isArray(items) || items.length === 0) return []
      return items.map(mapBackendCrewMember)
    },
    staleTime: 1000 * 60 * 5,
    retry: 1,
  })

  const createMutation = useMutation({
    mutationFn: (data: any) => crewApi.createMember(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['crew'] })
      setShowCreateModal(false)
      setFormData(emptyForm)
      addNotification('Der neue Mitarbeiter wurde erfolgreich angelegt.', 'success', { title: 'Mitarbeiter erstellt' })
    },
    onError: (err: any) => {
      addNotification(`Mitarbeiter konnte nicht erstellt werden: ${err.message || 'Unbekannter Fehler'}`, 'error', { title: 'Fehler' })
    },
  })

  const activeMembers = crewMembers.filter((m: any) => m.availability === 'available')
  const busyMembers = crewMembers.filter((m: any) => m.availability === 'busy')
  const onLeaveMembers = crewMembers.filter((m: any) => m.availability === 'on_leave' || m.availability === 'sick')

  const filteredMembers = crewMembers.filter((member: any) => {
    const roleMatch = filterRole === 'all' || member.role === filterRole
    const availabilityMatch = filterAvailability === 'all' || member.availability === filterAvailability
    const searchMatch = !searchQuery || member.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      member.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
      member.phone.includes(searchQuery)
    return roleMatch && availabilityMatch && searchMatch
  })

  const getRoleLabel = (role: string): string => {
    const found = ROLES.find(r => r.value === role)
    return found?.label || role
  }

  const getAvailabilityConfig = (status: string) => {
    const found = AVAILABILITY_OPTIONS.find(a => a.value === status)
    return found || { label: status, color: 'var(--color-text-secondary)' }
  }

  const handleAddSkill = () => {
    const trimmed = skillInput.trim()
    if (trimmed && !formData.skills.includes(trimmed)) {
      setFormData(prev => ({ ...prev, skills: [...prev.skills, trimmed] }))
      setSkillInput('')
    }
  }

  const handleRemoveSkill = (skill: string) => {
    setFormData(prev => ({ ...prev, skills: prev.skills.filter(s => s !== skill) }))
  }

  const handleSkillKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault()
      handleAddSkill()
    }
  }

  const handleSubmit = () => {
    if (!formData.first_name.trim() || !formData.last_name.trim()) {
      addNotification('Vorname und Nachname sind Pflichtfelder.', 'error', { title: 'Fehler' })
      return
    }
    createMutation.mutate({
      first_name: formData.first_name,
      last_name: formData.last_name,
      email: formData.email,
      phone: formData.phone,
      role: formData.role,
      hourly_rate: formData.hourly_rate ? parseFloat(formData.hourly_rate) : undefined,
      notes: formData.skills.join(', '),
      status: 'active',
    })
  }

  if (isLoading) {
    return (
      <div className="crew-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Crew Management</h1>
          </div>
        </div>
        <SkeletonKPI count={4} />
        <SkeletonTable rows={5} columns={5} />
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
          <p className="page-subtitle">Verwaltung von {crewMembers.length} Mitarbeitern und Einsätzen</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => setShowCreateModal(true)}
          style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
        >
          + Neuer Mitarbeiter
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
          <div className="stat-card__label">Im Einsatz</div>
          <div className="stat-card__value" style={{ color: 'var(--color-primary)' }}>{busyMembers.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">Abwesend</div>
          <div className="stat-card__value" style={{ color: 'var(--color-warning)' }}>{onLeaveMembers.length}</div>
        </div>
      </div>

      {/* Filters */}
      <div className="filters">
        <div className="filter-group" style={{ flex: '1 1 200px' }}>
          <input
            type="text"
            className="filter-select"
            placeholder="Mitarbeiter suchen..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{ width: '100%' }}
          />
        </div>
        <div className="filter-group">
          <label htmlFor="role-filter" className="filter-label">Rolle:</label>
          <select
            id="role-filter"
            className="filter-select"
            value={filterRole}
            onChange={(e) => setFilterRole(e.target.value)}
          >
            <option value="all">Alle Rollen</option>
            {ROLES.map(r => (
              <option key={r.value} value={r.value}>{r.label}</option>
            ))}
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
            {AVAILABILITY_OPTIONS.map(a => (
              <option key={a.value} value={a.value}>{a.label}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Crew Table */}
      {filteredMembers.length === 0 ? (
        <div className="empty-state">
          <div className="empty-state__icon">&#x1F465;</div>
          <h3 className="empty-state__title">Keine Mitarbeiter gefunden</h3>
          <p className="empty-state__description">
            {crewMembers.length === 0
              ? 'Legen Sie Ihren ersten Mitarbeiter an, um loszulegen.'
              : 'Versuchen Sie, die Filter zu ändern.'}
          </p>
          {crewMembers.length === 0 && (
            <button className="btn btn--primary" onClick={() => setShowCreateModal(true)}>
              + Neuer Mitarbeiter
            </button>
          )}
        </div>
      ) : (
        <div className="crew-table-wrapper">
          <table className="crew-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Rolle</th>
                <th>Verfügbarkeit</th>
                <th>Telefon</th>
                <th>E-Mail</th>
                <th>Skills</th>
              </tr>
            </thead>
            <tbody>
              {filteredMembers.map((member: any) => {
                const avail = getAvailabilityConfig(member.availability)
                return (
                  <tr
                    key={member.id}
                    onClick={() => navigate(`/crew/${member.id}`)}
                    className="crew-table__row"
                  >
                    <td>
                      <div className="crew-table__name-cell">
                        <div className="crew-table__avatar">
                          {member.name.split(' ').map((n: string) => n[0]).slice(0, 2).join('').toUpperCase()}
                        </div>
                        <span className="crew-table__name">{member.name}</span>
                      </div>
                    </td>
                    <td>
                      <span className="role-tag">{getRoleLabel(member.role)}</span>
                    </td>
                    <td>
                      <span
                        className="availability-badge-inline"
                        style={{
                          '--badge-color': avail.color,
                        } as React.CSSProperties}
                      >
                        <span className="availability-dot" />
                        {avail.label}
                      </span>
                    </td>
                    <td className="crew-table__contact">{member.phone || '—'}</td>
                    <td className="crew-table__contact">{member.email || '—'}</td>
                    <td>
                      <div className="skill-tags">
                        {(member.skills || []).slice(0, 3).map((skill: string, idx: number) => (
                          <span key={idx} className="skill-tag">{skill}</span>
                        ))}
                        {(member.skills || []).length > 3 && (
                          <span className="skill-tag skill-tag--more">+{member.skills.length - 3}</span>
                        )}
                        {(!member.skills || member.skills.length === 0) && (
                          <span style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-xs)' }}>—</span>
                        )}
                      </div>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Create Crew Member Modal */}
      <Modal
        isOpen={showCreateModal}
        onClose={() => { setShowCreateModal(false); setFormData(emptyForm) }}
        title="Neuer Mitarbeiter"
        size="lg"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end' }}>
            <button
              className="btn btn--secondary"
              onClick={() => { setShowCreateModal(false); setFormData(emptyForm) }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleSubmit}
              disabled={createMutation.isPending}
            >
              {createMutation.isPending ? 'Speichern...' : 'Mitarbeiter anlegen'}
            </button>
          </div>
        }
      >
        <div className="form-grid">
          <Input
            label="Vorname *"
            value={formData.first_name}
            onChange={(e) => setFormData(prev => ({ ...prev, first_name: e.target.value }))}
            placeholder="Vorname eingeben"
          />
          <Input
            label="Nachname *"
            value={formData.last_name}
            onChange={(e) => setFormData(prev => ({ ...prev, last_name: e.target.value }))}
            placeholder="Nachname eingeben"
          />
          <Input
            label="E-Mail"
            type="email"
            value={formData.email}
            onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
            placeholder="email@beispiel.de"
          />
          <Input
            label="Telefon"
            type="tel"
            value={formData.phone}
            onChange={(e) => setFormData(prev => ({ ...prev, phone: e.target.value }))}
            placeholder="+49 ..."
          />
          <div className="form-group">
            <label className="form-label">Rolle</label>
            <select
              className="form-input"
              value={formData.role}
              onChange={(e) => setFormData(prev => ({ ...prev, role: e.target.value }))}
            >
              {ROLES.map(r => (
                <option key={r.value} value={r.value}>{r.label}</option>
              ))}
            </select>
          </div>
          <Input
            label="Stundensatz (EUR)"
            type="number"
            value={formData.hourly_rate}
            onChange={(e) => setFormData(prev => ({ ...prev, hourly_rate: e.target.value }))}
            placeholder="0.00"
            min="0"
            step="0.50"
          />
          <div className="form-group form-group--full">
            <label className="form-label">Skills</label>
            <div className="skill-input-wrapper">
              <div className="skill-tags-input">
                {formData.skills.map((skill, idx) => (
                  <span key={idx} className="skill-tag skill-tag--removable">
                    {skill}
                    <button
                      type="button"
                      className="skill-tag__remove"
                      onClick={() => handleRemoveSkill(skill)}
                    >
                      x
                    </button>
                  </span>
                ))}
                <input
                  type="text"
                  className="skill-input-field"
                  value={skillInput}
                  onChange={(e) => setSkillInput(e.target.value)}
                  onKeyDown={handleSkillKeyDown}
                  onBlur={handleAddSkill}
                  placeholder={formData.skills.length === 0 ? 'Skill eingeben und Enter drücken...' : ''}
                />
              </div>
            </div>
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default CrewPage
