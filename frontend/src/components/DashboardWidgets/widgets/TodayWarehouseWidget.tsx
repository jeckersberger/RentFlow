import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { projectApi } from '../../../services/api'

interface UpcomingProject {
  id: string
  name: string
  start_date: string
  end_date: string
  status: string
  client?: string
}

export function TodayWarehouseWidget() {
  const navigate = useNavigate()

  const { data: todayWarehouseProjects } = useQuery<UpcomingProject[]>({
    queryKey: ['dashboard-today-warehouse'],
    queryFn: async () => {
      try {
        const res = await projectApi.list(1, 50)
        const projects = res?.items || []
        const today = new Date()
        const todayStr = today.toISOString().split('T')[0]
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        return projects.filter((p: any) => {
          return p.start_date <= todayStr && p.end_date >= todayStr && p.status !== 'completed'
        }).slice(0, 5)
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  if (!todayWarehouseProjects || todayWarehouseProjects.length === 0) {
    return (
      <div className="widget-empty">
        <span className="widget-empty__icon">✓</span>
        <p className="widget-empty__text">Keine Lagerbewegungen für heute geplant.</p>
      </div>
    )
  }

  return (
    <div className="widget-list">
      {todayWarehouseProjects.map((project) => (
        <div
          key={project.id}
          className="widget-list__item widget-list__item--clickable"
          onClick={() => navigate(`/projects/${project.id}`)}
        >
          <div className="widget-list__icon">📦</div>
          <div className="widget-list__content">
            <p className="widget-list__title">{project.name}</p>
            <p className="widget-list__meta">
              {project.client || 'Projekt'} &middot; {new Date(project.start_date).toLocaleDateString('de-DE')} - {new Date(project.end_date).toLocaleDateString('de-DE')}
            </p>
          </div>
          <span className={`widget-list__badge widget-list__badge--${project.status}`}>
            {project.status === 'active' ? 'Aktiv' : project.status === 'planning' ? 'Planung' : project.status}
          </span>
        </div>
      ))}
      <button className="widget-link" onClick={() => navigate('/warehouse')}>
        Zum Lager →
      </button>
    </div>
  )
}
