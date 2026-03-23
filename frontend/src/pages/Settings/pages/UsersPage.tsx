import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { userApi } from '../../../services/api'
import { UserPlus, X, Users } from 'lucide-react'
import { SkeletonTable } from '../../../components/Skeleton/SkeletonLoader'
import EmptyState from '../../../components/EmptyState/EmptyState'
import '../Settings.scss'

interface UserData {
  id: string
  name: string
  email: string
  role: string
  status: string
}

function UsersPage() {
  const queryClient = useQueryClient()
  const [showModal, setShowModal] = useState(false)
  const [newUser, setNewUser] = useState({ name: '', email: '', password: '', role: 'user' })

  const { data: usersData, isLoading } = useQuery<UserData[]>({
    queryKey: ['users'],
    queryFn: () => userApi.list(),
  })

  const createMutation = useMutation({
    mutationFn: (data: typeof newUser) => userApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] })
      setShowModal(false)
      setNewUser({ name: '', email: '', password: '', role: 'user' })
    },
  })

  const roleBadge = (role: string) => {
    if (['admin', 'owner'].includes(role?.toLowerCase())) return 'badge badge--primary'
    return 'badge badge--secondary'
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

  const users = Array.isArray(usersData) ? usersData : []

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Benutzer</h1>
        <p>Verwalten Sie die Benutzerkonten und Zugriffsrechte.</p>
      </div>

      <div style={{ marginBottom: 'var(--spacing-4)' }}>
        <button className="sp-btn sp-btn--primary" onClick={() => setShowModal(true)}>
          <UserPlus size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          Benutzer einladen
        </button>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        {isLoading ? (
          <SkeletonTable rows={4} columns={4} />
        ) : users.length === 0 ? (
          <EmptyState
            icon={Users}
            title="Keine Benutzer vorhanden"
            description="Laden Sie Ihr erstes Teammitglied ein."
            action={{ label: 'Benutzer einladen', onClick: () => setShowModal(true) }}
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
                  <td>{user.name}</td>
                  <td>{user.email}</td>
                  <td><span className={roleBadge(user.role)}>{user.role}</span></td>
                  <td><span className={statusBadge(user.status)}>{statusLabel(user.status)}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {showModal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050,
        }}>
          <div className="sp-card" style={{ width: 480, maxWidth: '90vw', margin: 0 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-4)' }}>
              <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>Neuen Benutzer einladen</h2>
              <button onClick={() => setShowModal(false)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)' }}>
                <X size={20} />
              </button>
            </div>
            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
              <div className="sp-field">
                <label className="sp-label">Name</label>
                <input className="sp-input" value={newUser.name} onChange={(e) => setNewUser({ ...newUser, name: e.target.value })} />
              </div>
              <div className="sp-field">
                <label className="sp-label">E-Mail</label>
                <input className="sp-input" type="email" value={newUser.email} onChange={(e) => setNewUser({ ...newUser, email: e.target.value })} />
              </div>
              <div className="sp-field">
                <label className="sp-label">Passwort</label>
                <input className="sp-input" type="password" value={newUser.password} onChange={(e) => setNewUser({ ...newUser, password: e.target.value })} />
              </div>
              <div className="sp-field">
                <label className="sp-label">Rolle</label>
                <select className="sp-select" value={newUser.role} onChange={(e) => setNewUser({ ...newUser, role: e.target.value })}>
                  <option value="user">Benutzer</option>
                  <option value="admin">Admin</option>
                  <option value="manager">Manager</option>
                  <option value="technician">Techniker</option>
                </select>
              </div>
            </div>
            <div className="sp-footer">
              <button className="sp-btn sp-btn--secondary" onClick={() => setShowModal(false)}>Abbrechen</button>
              <button className="sp-btn sp-btn--primary" onClick={() => createMutation.mutate(newUser)} disabled={createMutation.isPending || !newUser.email || !newUser.name}>
                {createMutation.isPending ? 'Erstellen...' : 'Erstellen'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default UsersPage
