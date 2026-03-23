import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi, api } from '../../../services/api'
import { UserPlus, X, Users, Mail, Clock, CheckCircle } from 'lucide-react'
import { SkeletonTable } from '../../../components/Skeleton/SkeletonLoader'
import EmptyState from '../../../components/EmptyState/EmptyState'
import '../Settings.scss'

interface UserData {
  id: string
  first_name: string
  last_name: string
  email: string
  roles: string[]
  status: string
  created_at: string
}

interface InvitationData {
  id: string
  email: string
  role: string
  status: string
  expires_at: string
  created_at: string
}

const ROLES = [
  { value: 'admin', label: 'Administrator', desc: 'Voller Zugriff auf alle Funktionen' },
  { value: 'manager', label: 'Projektleiter', desc: 'Projekte, Equipment, Rechnungen verwalten' },
  { value: 'warehouse', label: 'Lagerist', desc: 'Equipment, Lager, Scanner' },
  { value: 'crew', label: 'Techniker/Crew', desc: 'Zugewiesene Projekte, Zeiterfassung' },
  { value: 'driver', label: 'Fahrer', desc: 'Transport, Zugewiesene Fahrten' },
  { value: 'readonly', label: 'Nur Lesen', desc: 'Kann alle Daten einsehen, aber nicht bearbeiten' },
]

function UsersPage() {
  const queryClient = useQueryClient()
  const [showModal, setShowModal] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('crew')
  const [inviteSuccess, setInviteSuccess] = useState('')

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const { data: usersRaw, isLoading } = useQuery<any>({
    queryKey: ['users'],
    queryFn: () => userApi.list(),
  })
  // Handle paginated response: { data: [...], page, total } or just [...]
  const usersData: UserData[] = Array.isArray(usersRaw) ? usersRaw : (usersRaw?.data || [])

  const { data: invitationsData } = useQuery<InvitationData[]>({
    queryKey: ['invitations'],
    queryFn: () => api.get('/api/v1/invitations').then(r => r.data),
  })

  const inviteMutation = useMutation({
    mutationFn: (data: { email: string; role: string }) =>
      api.post('/api/v1/users/invite', data).then(r => r.data),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({ queryKey: ['invitations'] })
      setInviteSuccess(`Einladung an ${variables.email} gesendet!`)
      setInviteEmail('')
      setInviteRole('crew')
      setTimeout(() => {
        setInviteSuccess('')
        setShowModal(false)
      }, 2000)
    },
  })

  const roleBadge = (role: string) => {
    if (['admin', 'owner', 'superadmin'].includes(role?.toLowerCase())) return 'badge badge--primary'
    if (['manager'].includes(role?.toLowerCase())) return 'badge badge--success'
    return 'badge badge--secondary'
  }

  const roleLabel = (roles: string[] | string) => {
    const role = Array.isArray(roles) ? roles[0] : roles
    const found = ROLES.find(r => r.value === role)
    return found?.label || role
  }

  const statusBadge = (status: string) => {
    if (status === 'active') return 'badge badge--success'
    if (['invited', 'pending'].includes(status)) return 'badge badge--warning'
    return 'badge badge--secondary'
  }

  const statusLabel = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active': return 'Aktiv'
      case 'invited': case 'pending': return 'Eingeladen'
      case 'inactive': case 'disabled': return 'Deaktiviert'
      default: return status
    }
  }

  const users = usersData
  const invitations = Array.isArray(invitationsData) ? invitationsData.filter(i => i.status === 'pending') : []

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Benutzer & Team</h1>
        <p>Verwalten Sie Ihr Team. Laden Sie Mitarbeiter per E-Mail ein — sie erhalten einen Link zum Registrieren.</p>
      </div>

      <div style={{ marginBottom: 'var(--spacing-4)' }}>
        <button className="sp-btn sp-btn--primary" onClick={() => setShowModal(true)}>
          <UserPlus size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          Mitarbeiter einladen
        </button>
      </div>

      {/* Pending Invitations */}
      {invitations.length > 0 && (
        <div className="sp-card" style={{ marginBottom: 'var(--spacing-6)' }}>
          <h3 className="sp-card__title">Offene Einladungen</h3>
          {invitations.map((inv) => (
            <div key={inv.id} style={{
              display: 'flex', alignItems: 'center', justifyContent: 'space-between',
              padding: 'var(--spacing-3) 0', borderBottom: '1px solid var(--color-border)',
            }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)' }}>
                <Mail size={16} style={{ color: 'var(--color-warning)', flexShrink: 0 }} />
                <div>
                  <div style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--color-text-primary)' }}>{inv.email}</div>
                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
                    {roleLabel(inv.role)} — Eingeladen am {new Date(inv.created_at).toLocaleDateString('de-DE')}
                  </div>
                </div>
              </div>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)' }}>
                <Clock size={14} style={{ color: 'var(--color-text-muted)' }} />
                <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
                  Gueltig bis {new Date(inv.expires_at).toLocaleDateString('de-DE')}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Active Users */}
      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        {isLoading ? (
          <SkeletonTable rows={4} columns={4} />
        ) : users.length === 0 ? (
          <EmptyState
            icon={Users}
            title="Noch keine Teammitglieder"
            description="Laden Sie Ihr erstes Teammitglied per E-Mail ein."
            action={{ label: 'Mitarbeiter einladen', onClick: () => setShowModal(true) }}
            compact
          />
        ) : (
          <table className="sp-table">
            <thead>
              <tr>
                <th>Name</th>
                <th>E-Mail</th>
                <th>Rolle</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => (
                <tr key={user.id}>
                  <td style={{ fontWeight: 'var(--font-weight-medium)' }}>
                    {user.first_name} {user.last_name}
                  </td>
                  <td>{user.email}</td>
                  <td><span className={roleBadge(Array.isArray(user.roles) ? user.roles[0] : '')}>{roleLabel(user.roles)}</span></td>
                  <td><span className={statusBadge(user.status)}>{statusLabel(user.status)}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Invite Modal */}
      {showModal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050,
        }} onClick={() => setShowModal(false)}>
          <div className="sp-card" style={{ width: 520, maxWidth: '90vw', margin: 0 }} onClick={e => e.stopPropagation()}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-5)' }}>
              <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>
                Mitarbeiter einladen
              </h2>
              <button onClick={() => setShowModal(false)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)' }}>
                <X size={20} />
              </button>
            </div>

            {inviteSuccess ? (
              <div style={{ textAlign: 'center', padding: 'var(--spacing-6)' }}>
                <CheckCircle size={48} style={{ color: 'var(--color-success)', marginBottom: 'var(--spacing-3)' }} />
                <p style={{ color: 'var(--color-success)', fontWeight: 'var(--font-weight-semibold)' }}>{inviteSuccess}</p>
              </div>
            ) : (
              <>
                <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-4) 0' }}>
                  Der Mitarbeiter erhaelt einen Einladungslink per E-Mail und kann sich selbst registrieren.
                </p>

                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-4)' }}>
                  <div className="sp-field">
                    <label className="sp-label">E-Mail-Adresse *</label>
                    <input
                      className="sp-input"
                      type="email"
                      placeholder="mitarbeiter@je-soundulight.de"
                      value={inviteEmail}
                      onChange={(e) => setInviteEmail(e.target.value)}
                      autoFocus
                    />
                  </div>

                  <div className="sp-field">
                    <label className="sp-label">Rolle</label>
                    <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
                      {ROLES.map((role) => (
                        <label key={role.value} style={{
                          display: 'flex', alignItems: 'flex-start', gap: 'var(--spacing-3)',
                          padding: 'var(--spacing-3)', borderRadius: 'var(--radius-md)',
                          border: `1px solid ${inviteRole === role.value ? 'var(--color-primary)' : 'var(--color-border)'}`,
                          background: inviteRole === role.value ? 'rgba(0, 212, 255, 0.06)' : 'transparent',
                          cursor: 'pointer', transition: 'all 0.15s ease',
                        }}>
                          <input
                            type="radio" name="role" value={role.value}
                            checked={inviteRole === role.value}
                            onChange={() => setInviteRole(role.value)}
                            style={{ marginTop: 2 }}
                          />
                          <div>
                            <div style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)' }}>
                              {role.label}
                            </div>
                            <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
                              {role.desc}
                            </div>
                          </div>
                        </label>
                      ))}
                    </div>
                  </div>
                </div>

                {inviteMutation.isError && (
                  <p style={{ color: 'var(--color-danger)', fontSize: 'var(--font-size-sm)', marginTop: 'var(--spacing-3)' }}>
                    {(inviteMutation.error as Error)?.message || 'Fehler beim Einladen'}
                  </p>
                )}

                <div className="sp-footer">
                  <button className="sp-btn sp-btn--secondary" onClick={() => setShowModal(false)}>Abbrechen</button>
                  <button
                    className="sp-btn sp-btn--primary"
                    onClick={() => inviteMutation.mutate({ email: inviteEmail, role: inviteRole })}
                    disabled={inviteMutation.isPending || !inviteEmail}
                  >
                    {inviteMutation.isPending ? 'Wird eingeladen...' : 'Einladung senden'}
                  </button>
                </div>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

export default UsersPage
