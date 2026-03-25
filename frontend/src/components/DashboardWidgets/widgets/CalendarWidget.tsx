import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { projectApi } from '../../../services/api'

export function CalendarWidget() {
  const navigate = useNavigate()

  const { data: projects } = useQuery({
    queryKey: ['dashboard-upcoming-projects'],
    queryFn: async () => {
      try {
        const res = await projectApi.list(1, 10)
        return res?.items || []
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 10,
  })

  // Generate this week's days
  const today = new Date()
  const startOfWeek = new Date(today)
  const dayOfWeek = today.getDay()
  const diff = dayOfWeek === 0 ? -6 : 1 - dayOfWeek // Monday start
  startOfWeek.setDate(today.getDate() + diff)

  const weekDays = Array.from({ length: 7 }, (_, i) => {
    const date = new Date(startOfWeek)
    date.setDate(startOfWeek.getDate() + i)
    return date
  })

  const dayNames = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So']

  // Map projects to days
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const getProjectsForDay = (date: Date): any[] => {
    if (!projects) return []
    const dateStr = date.toISOString().split('T')[0]
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    return projects.filter((p: any) => {
      return p.start_date <= dateStr && p.end_date >= dateStr
    }).slice(0, 2)
  }

  const isToday = (date: Date) => {
    return date.toISOString().split('T')[0] === today.toISOString().split('T')[0]
  }

  return (
    <div className="calendar-widget">
      <div className="calendar-widget__week-header">
        <span className="calendar-widget__week-label">
          KW {getWeekNumber(today)} &middot; {startOfWeek.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' })}
        </span>
      </div>

      <div className="calendar-widget__grid">
        {weekDays.map((date, index) => {
          const dayProjects = getProjectsForDay(date)

          return (
            <div
              key={index}
              className={`calendar-widget__day ${isToday(date) ? 'calendar-widget__day--today' : ''} ${index >= 5 ? 'calendar-widget__day--weekend' : ''}`}
            >
              <div className="calendar-widget__day-header">
                <span className="calendar-widget__day-name">{dayNames[index]}</span>
                <span className="calendar-widget__day-number">{date.getDate()}</span>
              </div>
              <div className="calendar-widget__day-events">
                {dayProjects.map((project, pi) => (
                  <div
                    key={pi}
                    className="calendar-widget__event"
                    onClick={() => project.id && navigate(`/projects/${project.id}`)}
                    title={project.name}
                  >
                    {project.name}
                  </div>
                ))}
              </div>
            </div>
          )
        })}
      </div>

      <button className="widget-link" onClick={() => navigate('/calendar')}>
        Zum Kalender →
      </button>
    </div>
  )
}

function getWeekNumber(date: Date): number {
  const d = new Date(Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()))
  const dayNum = d.getUTCDay() || 7
  d.setUTCDate(d.getUTCDate() + 4 - dayNum)
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1))
  return Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7)
}
