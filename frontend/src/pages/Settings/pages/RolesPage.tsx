import { Shield, Lock } from 'lucide-react'
import '../Settings.module.scss'

const roles = [
  {
    name: 'Owner',
    description: 'Vollzugriff auf alle Bereiche, kann Firma und Abrechnung verwalten.',
    permissions: ['Alles', 'Firma', 'Abrechnung', 'Benutzer', 'Rollen'],
    color: '#ef4444',
  },
  {
    name: 'Admin',
    description: 'Kann alle operativen Bereiche verwalten, aber keine Firmeneinstellungen ändern.',
    permissions: ['Equipment', 'Projekte', 'Rechnungen', 'Crew', 'Lager', 'Reports', 'Einstellungen'],
    color: '#f59e0b',
  },
  {
    name: 'Manager',
    description: 'Kann Projekte und Equipment verwalten, Rechnungen erstellen.',
    permissions: ['Equipment', 'Projekte', 'Rechnungen', 'Crew', 'Lager'],
    color: '#8b5cf6',
  },
  {
    name: 'Techniker',
    description: 'Kann Equipment ein-/auschecken, Scanner nutzen und Wartung durchführen.',
    permissions: ['Equipment (lesen)', 'Scanner', 'Wartung', 'Lager', 'Eigene Aufgaben'],
    color: '#06b6d4',
  },
  {
    name: 'Benutzer',
    description: 'Grundlegender Lesezugriff auf zugewiesene Projekte und Equipment.',
    permissions: ['Dashboard', 'Eigene Projekte', 'Equipment (lesen)'],
    color: '#6b7280',
  },
]

function RolesPage() {
  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Rollen & Berechtigungen</h1>
        <p>Definieren Sie, welche Berechtigungen jede Rolle hat.</p>
      </div>

      <div style={{
        display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)',
        padding: 'var(--spacing-4)', borderRadius: 'var(--radius-md)',
        background: 'rgba(99, 102, 241, 0.08)', border: '1px solid rgba(99, 102, 241, 0.2)',
        marginBottom: 'var(--spacing-5)', color: 'var(--color-text-secondary)',
        fontSize: 'var(--font-size-sm)',
      }}>
        <Lock size={16} style={{ flexShrink: 0, color: 'var(--color-primary)' }} />
        <span>Benutzerdefinierte Rollen und Berechtigungsmatrix sind <strong>bald verfügbar</strong>. Aktuell gelten die unten gezeigten Standard-Rollen.</span>
      </div>

      <div className="sp-role-cards">
        {roles.map((role) => (
          <div key={role.name} className="sp-role-card">
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-2)' }}>
              <Shield size={18} style={{ color: role.color }} />
              <h3 style={{ margin: 0 }}>{role.name}</h3>
            </div>
            <p>{role.description}</p>
            <div className="sp-role-card__permissions">
              {role.permissions.map((perm) => (
                <span key={perm} className="sp-role-card__perm">{perm}</span>
              ))}
            </div>
          </div>
        ))}
      </div>

      <div className="sp-footer">
        <button className="sp-btn sp-btn--primary" disabled style={{ opacity: 0.5, cursor: 'not-allowed' }}>
          Rollen bearbeiten (Bald verfügbar)
        </button>
      </div>
    </div>
  )
}

export default RolesPage
