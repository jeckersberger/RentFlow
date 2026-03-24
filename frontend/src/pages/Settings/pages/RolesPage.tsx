import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { useNotificationStore } from '../../../stores/notificationStore'
import { Shield, Plus, X, Eye, Check, Minus, Save } from 'lucide-react'
import '../Settings.scss'

type PermissionLevel = 'full' | 'read' | 'none'

interface RolePermissions {
  [permission: string]: PermissionLevel
}

interface RoleDefinition {
  key: string
  label: string
  description: string
  color: string
  isCustom?: boolean
  permissions: RolePermissions
}

const PERMISSIONS = [
  { key: 'equipment', label: 'Equipment verwalten' },
  { key: 'projects', label: 'Projekte verwalten' },
  { key: 'invoices', label: 'Rechnungen erstellen' },
  { key: 'crew', label: 'Crew verwalten' },
  { key: 'scanner', label: 'Scanner nutzen' },
  { key: 'reports', label: 'Reports sehen' },
  { key: 'settings', label: 'Einstellungen aendern' },
  { key: 'users', label: 'Benutzer verwalten' },
]

const DEFAULT_ROLES: RoleDefinition[] = [
  {
    key: 'admin', label: 'Admin', description: 'Vollzugriff auf alle Bereiche', color: '#ef4444',
    permissions: { equipment: 'full', projects: 'full', invoices: 'full', crew: 'full', scanner: 'full', reports: 'full', settings: 'full', users: 'full' },
  },
  {
    key: 'manager', label: 'Manager', description: 'Projekte, Equipment, Rechnungen, Crew', color: '#f59e0b',
    permissions: { equipment: 'full', projects: 'full', invoices: 'full', crew: 'full', scanner: 'full', reports: 'full', settings: 'none', users: 'none' },
  },
  {
    key: 'warehouse', label: 'Lager', description: 'Equipment und Lagerverwaltung', color: '#10b981',
    permissions: { equipment: 'full', projects: 'none', invoices: 'none', crew: 'none', scanner: 'full', reports: 'none', settings: 'none', users: 'none' },
  },
  {
    key: 'driver', label: 'Fahrer', description: 'Transport und zugewiesene Fahrten', color: '#3b82f6',
    permissions: { equipment: 'none', projects: 'none', invoices: 'none', crew: 'none', scanner: 'full', reports: 'none', settings: 'none', users: 'none' },
  },
  {
    key: 'crew', label: 'Crew', description: 'Zugewiesene Projekte und Zeiterfassung', color: '#8b5cf6',
    permissions: { equipment: 'none', projects: 'none', invoices: 'none', crew: 'none', scanner: 'full', reports: 'none', settings: 'none', users: 'none' },
  },
  {
    key: 'freelancer', label: 'Freelancer', description: 'Externer Zugriff auf zugewiesene Bereiche', color: '#06b6d4',
    permissions: { equipment: 'none', projects: 'none', invoices: 'none', crew: 'none', scanner: 'full', reports: 'none', settings: 'none', users: 'none' },
  },
  {
    key: 'readonly', label: 'Nur Lesen', description: 'Kann alle Daten einsehen, nicht bearbeiten', color: '#6b7280',
    permissions: { equipment: 'read', projects: 'read', invoices: 'read', crew: 'none', scanner: 'none', reports: 'read', settings: 'none', users: 'none' },
  },
]

function RolesPage() {
  const queryClient = useQueryClient()
  const addNotification = useNotificationStore((s) => s.addNotification)
  const [roles, setRoles] = useState<RoleDefinition[]>(DEFAULT_ROLES)
  const [hasChanges, setHasChanges] = useState(false)
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [showNewRoleModal, setShowNewRoleModal] = useState(false)
  const [newRoleName, setNewRoleName] = useState('')
  const [newRoleDescription, setNewRoleDescription] = useState('')
  const [newRolePermissions, setNewRolePermissions] = useState<RolePermissions>(
    Object.fromEntries(PERMISSIONS.map((p) => [p.key, 'none' as PermissionLevel]))
  )

  const { data: rolesConfig } = useQuery({
    queryKey: ['config', 'roles.permissions'],
    queryFn: () => configApi.get('roles.permissions'),
  })

  useEffect(() => {
    if (rolesConfig?.value && Array.isArray(rolesConfig.value)) {
      setRoles(rolesConfig.value)
    }
  }, [rolesConfig])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('roles.permissions', { value: roles }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'roles.permissions'] })
      setHasChanges(false)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
    onError: () => {
      addNotification('Fehler beim Speichern der Berechtigungen', 'error')
    },
  })

  const cyclePermission = (roleIndex: number, permKey: string) => {
    setRoles((prev) => {
      const updated = prev.map((r, i) => {
        if (i !== roleIndex) return r
        const current = r.permissions[permKey] || 'none'
        let next: PermissionLevel = 'none'
        if (current === 'none') next = 'read'
        else if (current === 'read') next = 'full'
        else next = 'none'
        return { ...r, permissions: { ...r.permissions, [permKey]: next } }
      })
      return updated
    })
    setHasChanges(true)
  }

  const deleteCustomRole = (roleKey: string) => {
    setRoles((prev) => prev.filter((r) => r.key !== roleKey))
    setHasChanges(true)
  }

  const addCustomRole = () => {
    if (!newRoleName.trim()) return
    const key = newRoleName.toLowerCase().replace(/[^a-z0-9]/g, '_')
    if (roles.some((r) => r.key === key)) {
      addNotification('Eine Rolle mit diesem Namen existiert bereits', 'error')
      return
    }
    const newRole: RoleDefinition = {
      key,
      label: newRoleName.trim(),
      description: newRoleDescription.trim(),
      color: '#' + Math.floor(Math.random() * 0xffffff).toString(16).padStart(6, '0'),
      isCustom: true,
      permissions: { ...newRolePermissions },
    }
    setRoles((prev) => [...prev, newRole])
    setHasChanges(true)
    setShowNewRoleModal(false)
    setNewRoleName('')
    setNewRoleDescription('')
    setNewRolePermissions(Object.fromEntries(PERMISSIONS.map((p) => [p.key, 'none' as PermissionLevel])))
  }

  const PermissionCell = ({ level, onClick }: { level: PermissionLevel; onClick: () => void }) => {
    const style: React.CSSProperties = {
      width: 32, height: 32, display: 'flex', alignItems: 'center', justifyContent: 'center',
      borderRadius: 'var(--radius-md)', cursor: 'pointer', transition: 'all 0.15s ease',
      border: '1px solid transparent', margin: '0 auto',
    }

    if (level === 'full') {
      return (
        <button onClick={onClick} style={{ ...style, background: 'rgba(16, 185, 129, 0.15)', color: 'var(--color-success)', borderColor: 'rgba(16, 185, 129, 0.3)' }} title="Vollzugriff (Klicken zum Aendern)">
          <Check size={16} />
        </button>
      )
    }
    if (level === 'read') {
      return (
        <button onClick={onClick} style={{ ...style, background: 'rgba(99, 102, 241, 0.12)', color: '#818cf8', borderColor: 'rgba(99, 102, 241, 0.25)' }} title="Nur Lesen (Klicken zum Aendern)">
          <Eye size={16} />
        </button>
      )
    }
    return (
      <button onClick={onClick} style={{ ...style, background: 'transparent', color: 'var(--color-text-muted)' }} title="Kein Zugriff (Klicken zum Aendern)">
        <Minus size={16} />
      </button>
    )
  }

  return (
    <div className="sp-page" style={{ maxWidth: 1100 }}>
      <div className="sp-header">
        <h1>Rollen & Berechtigungen</h1>
        <p>Definieren Sie, welche Berechtigungen jede Rolle hat. Klicken Sie auf eine Zelle, um die Berechtigung zu aendern.</p>
      </div>

      {/* Permission Matrix */}
      <div className="sp-card" style={{ overflowX: 'auto' }}>
        <h3 className="sp-card__title">
          <Shield size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Berechtigungsmatrix
        </h3>

        <div style={{ display: 'flex', gap: 'var(--spacing-3)', marginBottom: 'var(--spacing-4)', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: 20, height: 20, borderRadius: 4, background: 'rgba(16, 185, 129, 0.15)', color: 'var(--color-success)' }}><Check size={12} /></span>
            Vollzugriff
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: 20, height: 20, borderRadius: 4, background: 'rgba(99, 102, 241, 0.12)', color: '#818cf8' }}><Eye size={12} /></span>
            Nur Lesen
          </span>
          <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ display: 'inline-flex', alignItems: 'center', justifyContent: 'center', width: 20, height: 20, borderRadius: 4, color: 'var(--color-text-muted)' }}><Minus size={12} /></span>
            Kein Zugriff
          </span>
        </div>

        <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 'var(--font-size-sm)' }}>
          <thead>
            <tr style={{ borderBottom: '2px solid var(--color-border)' }}>
              <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 600, fontSize: 'var(--font-size-xs)', minWidth: 180 }}>
                Berechtigung
              </th>
              {roles.map((role) => (
                <th key={role.key} style={{ textAlign: 'center', padding: 'var(--spacing-2) var(--spacing-2)', fontWeight: 600, fontSize: 'var(--font-size-xs)', minWidth: 80 }}>
                  <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 2 }}>
                    <span style={{ color: role.color }}>{role.label}</span>
                    {role.isCustom && (
                      <button
                        onClick={() => deleteCustomRole(role.key)}
                        style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: 0, fontSize: 'var(--font-size-xs)' }}
                        title="Rolle entfernen"
                      >
                        <X size={12} />
                      </button>
                    )}
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {PERMISSIONS.map((perm) => (
              <tr key={perm.key} style={{ borderBottom: '1px solid var(--color-border)' }}>
                <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)', fontWeight: 500 }}>
                  {perm.label}
                </td>
                {roles.map((role, roleIndex) => (
                  <td key={role.key} style={{ padding: 'var(--spacing-1)', textAlign: 'center' }}>
                    <PermissionCell
                      level={role.permissions[perm.key] || 'none'}
                      onClick={() => cyclePermission(roleIndex, perm.key)}
                    />
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Save + New Role */}
      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Berechtigungen gespeichert!</span>}
        {saveMutation.isError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button
          className="sp-btn sp-btn--secondary"
          onClick={() => setShowNewRoleModal(true)}
        >
          <Plus size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          Neue Rolle erstellen
        </button>
        <button
          className="sp-btn sp-btn--primary"
          onClick={() => saveMutation.mutate()}
          disabled={saveMutation.isPending || !hasChanges}
        >
          <Save size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          {saveMutation.isPending ? 'Speichern...' : 'Speichern'}
        </button>
      </div>

      {/* New Role Modal */}
      {showNewRoleModal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050,
        }} onClick={() => setShowNewRoleModal(false)}>
          <div
            className="sp-card"
            style={{ width: 520, maxWidth: '90vw', margin: 0, maxHeight: '85vh', overflowY: 'auto' }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-5)' }}>
              <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>
                Neue Rolle erstellen
              </h2>
              <button onClick={() => setShowNewRoleModal(false)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: 4 }}>
                <X size={20} />
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-4)' }}>
              <div className="sp-field">
                <label className="sp-label">Name *</label>
                <input
                  className="sp-input"
                  type="text"
                  placeholder="z.B. Projektassistent"
                  value={newRoleName}
                  onChange={(e) => setNewRoleName(e.target.value)}
                  autoFocus
                />
              </div>

              <div className="sp-field">
                <label className="sp-label">Beschreibung</label>
                <input
                  className="sp-input"
                  type="text"
                  placeholder="Kurze Beschreibung der Rolle"
                  value={newRoleDescription}
                  onChange={(e) => setNewRoleDescription(e.target.value)}
                />
              </div>

              <div className="sp-field">
                <label className="sp-label">Berechtigungen</label>
                <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)', marginTop: 'var(--spacing-2)' }}>
                  {PERMISSIONS.map((perm) => {
                    const current = newRolePermissions[perm.key] || 'none'
                    return (
                      <div key={perm.key} style={{
                        display: 'flex', alignItems: 'center', justifyContent: 'space-between',
                        padding: 'var(--spacing-2) var(--spacing-3)',
                        background: 'var(--color-bg-secondary)', borderRadius: 'var(--radius-md)',
                        border: '1px solid var(--color-border)',
                      }}>
                        <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-primary)' }}>{perm.label}</span>
                        <div style={{ display: 'flex', gap: 'var(--spacing-1)' }}>
                          {(['none', 'read', 'full'] as PermissionLevel[]).map((level) => (
                            <button
                              key={level}
                              onClick={() => setNewRolePermissions((prev) => ({ ...prev, [perm.key]: level }))}
                              style={{
                                padding: '4px 10px', borderRadius: 'var(--radius-md)', border: 'none',
                                fontSize: 'var(--font-size-xs)', fontWeight: 500, cursor: 'pointer',
                                transition: 'all 0.15s ease',
                                background: current === level
                                  ? level === 'full' ? 'rgba(16, 185, 129, 0.2)' : level === 'read' ? 'rgba(99, 102, 241, 0.15)' : 'rgba(107, 114, 128, 0.15)'
                                  : 'transparent',
                                color: current === level
                                  ? level === 'full' ? 'var(--color-success)' : level === 'read' ? '#818cf8' : 'var(--color-text-muted)'
                                  : 'var(--color-text-muted)',
                              }}
                            >
                              {level === 'none' ? 'Kein' : level === 'read' ? 'Lesen' : 'Voll'}
                            </button>
                          ))}
                        </div>
                      </div>
                    )
                  })}
                </div>
              </div>
            </div>

            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--spacing-3)', marginTop: 'var(--spacing-5)' }}>
              <button className="sp-btn sp-btn--secondary" onClick={() => setShowNewRoleModal(false)}>
                Abbrechen
              </button>
              <button
                className="sp-btn sp-btn--primary"
                onClick={addCustomRole}
                disabled={!newRoleName.trim()}
              >
                Rolle erstellen
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default RolesPage
