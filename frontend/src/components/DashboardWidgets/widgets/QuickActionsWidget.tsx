import { useNavigate } from 'react-router-dom'

export function QuickActionsWidget() {
  const navigate = useNavigate()

  const actions = [
    { key: 'equipment', icon: '➕', label: 'Ausrüstung hinzufügen', path: '/equipment/new' },
    { key: 'project', icon: '📋', label: 'Neues Projekt', path: '/projects/new' },
    { key: 'invoice', icon: '💳', label: 'Rechnung erstellen', path: '/invoices/new' },
    { key: 'scanner', icon: '📱', label: 'Scannen', path: '/scanner' },
    { key: 'warehouse', icon: '🏭', label: 'Lager öffnen', path: '/warehouse' },
    { key: 'calendar', icon: '📅', label: 'Kalender', path: '/calendar' },
    { key: 'fleet', icon: '🚛', label: 'Fuhrpark', path: '/fleet' },
    { key: 'maintenance', icon: '🔧', label: 'Wartungen', path: '/maintenance' },
  ]

  return (
    <div className="quick-actions quick-actions--wide">
      {actions.map((action) => (
        <button
          key={action.key}
          className="quick-action-btn"
          onClick={() => navigate(action.path)}
        >
          <span className="quick-action-icon">{action.icon}</span>
          <span className="quick-action-text">{action.label}</span>
        </button>
      ))}
    </div>
  )
}
