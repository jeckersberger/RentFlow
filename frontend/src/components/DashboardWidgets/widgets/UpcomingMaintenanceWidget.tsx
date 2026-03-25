import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../../services/api'

export function UpcomingMaintenanceWidget() {
  const navigate = useNavigate()

  const { data: maintenanceData, isLoading } = useQuery({
    queryKey: ['dashboard-upcoming-maintenance'],
    queryFn: async () => {
      try {
        const res = await api.get('/api/v1/maintenance/tasks', { params: { page: 1, per_page: 10, sort: 'due_date' } })
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const tasks: any[] = res.data?.data || res.data?.items || res.data || []
        if (!Array.isArray(tasks)) return []
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return tasks.filter((t: any) => t.status !== 'completed' && t.status !== 'cancelled').slice(0, 5).map((t: any) => ({
          id: t.id,
          equipment: t.equipment_name || t.title || 'Unbekannt',
          task: t.description || t.task_type || '',
          due: t.due_date || t.next_due || '',
          priority: t.priority || 'medium',
        }))
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  const maintenanceSoon = maintenanceData || []

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

  if (isLoading) {
    return (
      <div className="widget-list">
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '1rem' }}>Laden...</p>
      </div>
    )
  }

  if (maintenanceSoon.length === 0) {
    return (
      <div className="widget-list">
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '1rem' }}>
          Keine anstehenden Wartungen.
        </p>
        <button className="widget-link" onClick={() => navigate('/maintenance')}>
          Alle Wartungen →
        </button>
      </div>
    )
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
