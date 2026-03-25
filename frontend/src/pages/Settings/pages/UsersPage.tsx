import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi, api } from '../../../services/api'
import { UserPlus, X, Users, Mail, Clock, CheckCircle, ShieldOff, ShieldCheck, Trash2, AlertTriangle } from 'lucide-react'
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
  last_login_at: string | null
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
  { value: 'admin', label: 'Administrator', desc: 'Vollzugriff inkl. Servereinstellungen und Personal' },
  { value: 'projectlead', label: 'Projektleiter', desc: 'Alles ausser Servereinstellungen und Personal' },
  { value: 'warehouse', label: 'Lager', desc: 'Equipment, Scanner, Lager, Inventur, Wartung' },
  { value: 'technician', label: 'Techniker', desc: 'Equipment, Projekte, Scanner, Wartung' },
  { value: 'driver', label: 'Fahrer', desc: 'Transport, Scanner, zugewiesene Projekte' },
  { value: 'accounting', label: 'Buchhaltung', desc: 'Rechnungen, Belege, Kontakte, Reports' },
  { value: 'freelancer', label: 'Freelancer', desc: 'Nur zugewiesene Projekte und Zeiterfassung' },
  { value: 'readonly', label: 'Nur Lesen', desc: 'Kann alle Daten einsehen, aber nicht bearbeiten' },
]

function UsersPage() {
  const queryClient = useQueryClient()
  const [showModal, setShowModal] = useState(false)
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('crew')
  const [inviteSuccess, setInviteSuccess] = useState('')
  const [deleteConfirm, setDeleteConfirm] = useState<UserData | null>(null)

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

  const deactivateMutation = useMutation({
    mutationFn: (userId: string) => userApi.deactivate(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const activateMutation = useMutation({
    mutationFn: (userId: string) => userApi.activate(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (userId: string) => userApi.delete(userId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setDeleteConfirm(null)
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

  const getUserStatus = (user: UserData): { badge: string; label: string } => {
    // Gesperrt (inactive)
    if (user.status === 'inactive' || user.status === 'disabled') {
      return { badge: 'badge badge--danger', label: 'Gesperrt' }
    }
    // Einladung offen: active status but never logged in
    if (user.status === 'active' && !user.last_login_at) {
      return { badge: 'badge badge--warning', label: 'Einladung offen' }
    }
    // Aktiv
    if (user.status === 'active') {
      return { badge: 'badge badge--success', label: 'Aktiv' }
    }
    // Locked
    if (user.status === 'locked') {
      return { badge: 'badge badge--danger', label: 'Gesperrt' }
    }
    return { badge: 'badge badge--secondary', label: user.status }
  }

  const isUserDeactivated = (user: UserData) => {
    return user.status === 'inactive' || user.status === 'disabled'
  }

  const isCurrentUser = (user: UserData) => {
    try {
      const stored = localStorage.getItem('rentflow_user')
      if (stored) {
        const currentUser = JSON.parse(stored)
        return currentUser.id === user.id
      }
    } catch { /* ignore */ }
    return false
  }

  const isAdmin = (user: UserData) => {
    const roles = Array.isArray(user.roles) ? user.roles : [user.roles]
    return roles.some(r => ['admin', 'superadmin', 'owner'].includes(r?.toLowerCase()))
  }

  const canModifyUser = (user: UserData) => {
    // Admins cannot be deactivated or deleted
    if (isAdmin(user)) return false
    // Cannot modify yourself
    if (isCurrentUser(user)) return false
    return true
  }

  // Einladung zurueckziehen
  const revokeInvitationMutation = useMutation({
    mutationFn: (invitationId: string) =>
      api.delete(`/api/v1/invitations/${invitationId}`).then(r => r.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['invitations'] })
    },
  })

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
                <button
                  className="sp-btn sp-btn--secondary"
                  style={{ padding: '2px 8px', fontSize: 'var(--font-size-xs)', color: 'var(--color-danger)', borderColor: 'var(--color-danger)' }}
                  onClick={() => revokeInvitationMutation.mutate(inv.id)}
                  disabled={revokeInvitationMutation.isPending}
                  title="Einladung zurueckziehen"
                >
                  <X size={12} style={{ marginRight: 2, verticalAlign: 'middle' }} />
                  Zurueckziehen
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Active Users */}
      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        {isLoading ? (
          <SkeletonTable rows={4} columns={5} />
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
                <th style={{ textAlign: 'right' }}>Aktionen</th>
              </tr>
            </thead>
            <tbody>
              {users.map((user) => {
                const status = getUserStatus(user)
                const isSelf = isCurrentUser(user)
                const isDeactivated = isUserDeactivated(user)
                return (
                  <tr key={user.id} style={isDeactivated ? { opacity: 0.6 } : undefined}>
                    <td style={{ fontWeight: 'var(--font-weight-medium)' }}>
                      {user.first_name} {user.last_name}
                    </td>
                    <td>{user.email}</td>
                    <td><span className={roleBadge(Array.isArray(user.roles) ? user.roles[0] : '')}>{roleLabel(user.roles)}</span></td>
                    <td><span className={status.badge}>{status.label}</span></td>
                    <td style={{ textAlign: 'right' }}>
                      {canModifyUser(user) && (
                        <div style={{ display: 'flex', gap: 'var(--spacing-2)', justifyContent: 'flex-end' }}>
                          {isDeactivated ? (
                            <button
                              className="sp-btn sp-btn--secondary"
                              style={{ padding: '4px 10px', fontSize: 'var(--font-size-xs)' }}
                              onClick={() => activateMutation.mutate(user.id)}
                              disabled={activateMutation.isPending}
                              title="Benutzer entsperren"
                            >
                              <ShieldCheck size={14} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                              Entsperren
                            </button>
                          ) : (
                            <button
                              className="sp-btn sp-btn--secondary"
                              style={{ padding: '4px 10px', fontSize: 'var(--font-size-xs)' }}
                              onClick={() => deactivateMutation.mutate(user.id)}
                              disabled={deactivateMutation.isPending}
                              title="Benutzer sperren"
                            >
                              <ShieldOff size={14} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                              Sperren
                            </button>
                          )}
                          <button
                            className="sp-btn sp-btn--secondary"
                            style={{
                              padding: '4px 10px', fontSize: 'var(--font-size-xs)',
                              color: 'var(--color-danger)', borderColor: 'var(--color-danger)',
                            }}
                            onClick={() => setDeleteConfirm(user)}
                            title="Benutzer loeschen"
                          >
                            <Trash2 size={14} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                            Loeschen
                          </button>
                        </div>
                      )}
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        )}
      </div>

      {/* Delete Confirmation Modal */}
      {deleteConfirm && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050, padding: 'var(--spacing-4)',
        }} onClick={() => setDeleteConfirm(null)}>
          <div
            className="sp-card"
            style={{
              width: 440, maxWidth: '100%', margin: 0, padding: 0, overflow: 'hidden',
            }}
            onClick={e => e.stopPropagation()}
          >
            <div style={{
              padding: 'var(--spacing-6)', textAlign: 'center',
            }}>
              <AlertTriangle size={48} style={{ color: 'var(--color-danger)', marginBottom: 'var(--spacing-3)' }} />
              <h3 style={{ margin: '0 0 var(--spacing-2) 0', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-lg)' }}>
                Benutzer loeschen?
              </h3>
              <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-2) 0' }}>
                Sind Sie sicher, dass Sie den Benutzer <strong>{deleteConfirm.first_name} {deleteConfirm.last_name}</strong> ({deleteConfirm.email}) loeschen moechten?
              </p>
              <p style={{ color: 'var(--color-text-muted)', fontSize: 'var(--font-size-xs)', margin: 0 }}>
                Der Benutzer wird deaktiviert und kann sich nicht mehr einloggen.
              </p>
            </div>
            <div style={{
              display: 'flex', justifyContent: 'flex-end', gap: 'var(--spacing-3)',
              padding: 'var(--spacing-4) var(--spacing-6)', borderTop: '1px solid var(--color-border)',
            }}>
              <button className="sp-btn sp-btn--secondary" onClick={() => setDeleteConfirm(null)}>
                Abbrechen
              </button>
              <button
                className="sp-btn"
                style={{
                  background: 'var(--color-danger)', color: '#fff', border: 'none',
                }}
                onClick={() => deleteMutation.mutate(deleteConfirm.id)}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? 'Wird geloescht...' : 'Endgueltig loeschen'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Invite Modal */}
      {showModal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050, padding: 'var(--spacing-4)',
        }} onClick={() => setShowModal(false)}>
          <div
            className="sp-card"
            style={{
              width: 500, maxWidth: '100%', margin: 0, padding: 0, overflow: 'hidden',
              maxHeight: '85vh', display: 'flex', flexDirection: 'column',
            }}
            onClick={e => e.stopPropagation()}
          >
            {/* Modal Header */}
            <div style={{
              display: 'flex', justifyContent: 'space-between', alignItems: 'center',
              padding: 'var(--spacing-5) var(--spacing-6)', borderBottom: '1px solid var(--color-border)',
              flexShrink: 0,
            }}>
              <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>
                Mitarbeiter einladen
              </h2>
              <button
                onClick={() => setShowModal(false)}
                style={{
                  background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)',
                  padding: 4, borderRadius: 'var(--radius-md)', display: 'flex', alignItems: 'center', justifyContent: 'center',
                  transition: 'all 0.15s ease',
                }}
                onMouseEnter={(e) => { e.currentTarget.style.background = 'rgba(255,255,255,0.08)'; e.currentTarget.style.color = 'var(--color-text-primary)' }}
                onMouseLeave={(e) => { e.currentTarget.style.background = 'none'; e.currentTarget.style.color = 'var(--color-text-muted)' }}
              >
                <X size={20} />
              </button>
            </div>

            {inviteSuccess ? (
              <div style={{ textAlign: 'center', padding: 'var(--spacing-8) var(--spacing-6)' }}>
                <CheckCircle size={48} style={{ color: 'var(--color-success)', marginBottom: 'var(--spacing-3)' }} />
                <p style={{ color: 'var(--color-success)', fontWeight: 'var(--font-weight-semibold)', margin: 0 }}>{inviteSuccess}</p>
              </div>
            ) : (
              <>
                {/* Modal Body - Scrollable */}
                <div style={{
                  padding: 'var(--spacing-5) var(--spacing-6)', overflowY: 'auto', flex: 1,
                }}>
                  <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-5) 0', lineHeight: 1.5 }}>
                    Der Mitarbeiter erhaelt einen Einladungslink per E-Mail und kann sich selbst registrieren.
                  </p>

                  <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-5)' }}>
                    <div className="sp-field">
                      <label className="sp-label">E-Mail-Adresse *</label>
                      <input
                        className="sp-input"
                        type="email"
                        placeholder="mitarbeiter@je-soundulight.de"
                        value={inviteEmail}
                        onChange={(e) => setInviteEmail(e.target.value)}
                        autoFocus
                        style={{ width: '100%' }}
                      />
                    </div>

                    <div className="sp-field">
                      <label className="sp-label" style={{ marginBottom: 'var(--spacing-2)' }}>Rolle</label>
                      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
                        {ROLES.map((role) => (
                          <label key={role.value} style={{
                            display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)',
                            padding: 'var(--spacing-3) var(--spacing-4)', borderRadius: 'var(--radius-md)',
                            border: `1.5px solid ${inviteRole === role.value ? 'var(--color-primary)' : 'var(--color-border)'}`,
                            background: inviteRole === role.value ? 'rgba(0, 212, 255, 0.06)' : 'transparent',
                            cursor: 'pointer', transition: 'all 0.15s ease',
                          }}>
                            <input
                              type="radio" name="role" value={role.value}
                              checked={inviteRole === role.value}
                              onChange={() => setInviteRole(role.value)}
                              style={{ accentColor: 'var(--color-primary)', flexShrink: 0, width: 16, height: 16 }}
                            />
                            <div style={{ minWidth: 0, flex: 1 }}>
                              <div style={{ fontWeight: 'var(--font-weight-medium)', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)' }}>
                                {role.label}
                              </div>
                              <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', lineHeight: 1.4 }}>
                                {role.desc}
                              </div>
                            </div>
                          </label>
                        ))}
                      </div>
                    </div>
                  </div>

                  {inviteMutation.isError && (
                    <p style={{ color: 'var(--color-danger)', fontSize: 'var(--font-size-sm)', marginTop: 'var(--spacing-3)', marginBottom: 0 }}>
                      {(inviteMutation.error as Error)?.message || 'Fehler beim Einladen'}
                    </p>
                  )}
                </div>

                {/* Modal Footer */}
                <div style={{
                  display: 'flex', justifyContent: 'flex-end', gap: 'var(--spacing-3)',
                  padding: 'var(--spacing-4) var(--spacing-6)', borderTop: '1px solid var(--color-border)',
                  flexShrink: 0,
                }}>
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
