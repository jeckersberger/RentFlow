import { useNavigate } from 'react-router-dom'

export function UpcomingMaintenanceWidget() {
  const navigate = useNavigate()

  const maintenanceSoon = [
    { id: 'm1', equipment: 'Yamaha CL5', task: 'Wartung nach Nutzung (alle 100h)', due: '2026-03-25', priority: 'high' },
    { id: 'm2', equipment: 'MA Lighting grandMA3', task: 'Software-Update und Kalibrierung', due: '2026-03-15', priority: 'critical' },
    { id: 'm3', equipment: 'JBL VTX A12', task: 'Reinigung und Kontrolle', due: '2026-05-15', priority: 'medium' },
  ]

  const priorityColor = (p: string) => {
    switch (p) {
      case 'critical': return '#ef4444'
      case 'high': return '#f59e0b'
      case 'medium': return '#3b82f6'
      default: return '#6b7280'
    }
  }

  const priorityLabel = (p: string) => {
    switch (p) {
      case 'critical': return 'Kritisch'
      case 'high': return 'Hoch'
      case 'medium': return 'Mittel'
      default: return 'Normal'
    }
  }

  return (
    <div className="widget-list">
      {maintenanceSoon.map((item) => (
        <div key={item.id} className="widget-list__item">
          <div className="widget-list__icon">🔧</div>
          <div className="widget-list__content">
            <p className="widget-list__title">{item.equipment}</p>
            <p className="widget-list__meta">{item.task}</p>
          </div>
          <div className="widget-list__right">
            <span className="widget-list__priority" style={{ color: priorityColor(item.priority) }}>
              {priorityLabel(item.priority)}
            </span>
            <span className="widget-list__date">
              {new Date(item.due).toLocaleDateString('de-DE')}
            </span>
          </div>
        </div>
      ))}
      <button className="widget-link" onClick={() => navigate('/maintenance')}>
        Alle Wartungen →
      </button>
    </div>
  )
}
