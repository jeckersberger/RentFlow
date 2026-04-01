import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { projectApi } from '../../services/api'
import './Calendar.scss'

interface Project {
  id: string
  name: string
  client?: string
  status: string
  start_date: string
  end_date: string
  location?: string
}

const WEEKDAYS = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So']

const PROJECT_COLORS = [
  'rgba(0, 212, 255, 0.7)',
  'rgba(139, 92, 246, 0.7)',
  'rgba(16, 185, 129, 0.7)',
  'rgba(245, 158, 11, 0.7)',
  'rgba(239, 68, 68, 0.7)',
  'rgba(6, 182, 212, 0.7)',
  'rgba(168, 85, 247, 0.7)',
  'rgba(34, 197, 94, 0.7)',
]

function getColorForProject(index: number) {
  return PROJECT_COLORS[index % PROJECT_COLORS.length]
}

function CalendarPage() {
  const navigate = useNavigate()
  const [currentDate, setCurrentDate] = useState(new Date())

  const year = currentDate.getFullYear()
  const month = currentDate.getMonth()

  const { data: projectsData, isLoading } = useQuery({
    queryKey: ['projects-calendar'],
    queryFn: () => projectApi.list(1, 100),
    staleTime: 1000 * 60 * 5,
  })

  const projects: Project[] = useMemo(() => {
    const raw = projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])
    return raw.filter((p: Project) => p.start_date && p.end_date)
  }, [projectsData])

  const firstDayOfMonth = new Date(year, month, 1)
  // Monday = 0, Sunday = 6
  const startDayOfWeek = (firstDayOfMonth.getDay() + 6) % 7
  const daysInMonth = new Date(year, month + 1, 0).getDate()

  const calendarDays = useMemo(() => {
    const days: (number | null)[] = []
    for (let i = 0; i < startDayOfWeek; i++) {
      days.push(null)
    }
    for (let d = 1; d <= daysInMonth; d++) {
      days.push(d)
    }
    // Fill remaining to complete the grid
    while (days.length % 7 !== 0) {
      days.push(null)
    }
    return days
  }, [startDayOfWeek, daysInMonth])

  function getProjectsForDay(day: number): { project: Project; color: string }[] {
    const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`
    return projects
      .map((p, idx) => ({ project: p, color: getColorForProject(idx) }))
      .filter(({ project }) => {
        return dateStr >= project.start_date && dateStr <= project.end_date
      })
  }

  function goToToday() {
    setCurrentDate(new Date())
  }

  function goPrev() {
    setCurrentDate(new Date(year, month - 1, 1))
  }

  function goNext() {
    setCurrentDate(new Date(year, month + 1, 1))
  }

  const monthName = currentDate.toLocaleDateString('de-DE', { month: 'long', year: 'numeric' })
  const today = new Date()
  const isToday = (day: number) =>
    day === today.getDate() && month === today.getMonth() && year === today.getFullYear()

  return (
    <div className="page">
      <div className="header">
        <div>
          <h1 className="title">Kalender</h1>
          <p className="subtitle">Projektplanung und Terminübersicht in der Monatsansicht</p>
        </div>
        <div className="headerActions">
          <button className="btnSecondary" onClick={() => navigate('/projects/new')}>
            + Projekt erstellen
          </button>
        </div>
      </div>

      <div className="navigation">
        <button className="navBtn" onClick={goPrev}>&larr;</button>
        <button className="todayBtn" onClick={goToToday}>Heute</button>
        <h2 className="monthLabel">{monthName}</h2>
        <button className="navBtn" onClick={goNext}>&rarr;</button>
      </div>

      {isLoading ? (
        <div className="loadingState">
          <div className="spinner" />
          <p>Projekte werden geladen...</p>
        </div>
      ) : (
        <div className="calendarCard">
          <div className="weekdayHeader">
            {WEEKDAYS.map(day => (
              <div key={day} className="weekdayCell">{day}</div>
            ))}
          </div>
          <div className="grid">
            {calendarDays.map((day, idx) => {
              const dayProjects = day ? getProjectsForDay(day) : []
              return (
                <div
                  key={idx}
                  className={`dayCell ${!day ? 'dayCellEmpty' : ''} ${day && isToday(day) ? 'dayCellToday' : ''}`}
                >
                  {day && (
                    <>
                      <span className="dayNumber">{day}</span>
                      <div className="dayProjects">
                        {dayProjects.slice(0, 3).map(({ project, color }) => (
                          <div
                            key={project.id}
                            className="projectBar"
                            style={{ backgroundColor: color }}
                            onClick={() => navigate(`/projects/${project.id}`)}
                            title={`${project.name} (${project.client || ''})`}
                          >
                            {project.name}
                          </div>
                        ))}
                        {dayProjects.length > 3 && (
                          <div className="moreIndicator">
                            +{dayProjects.length - 3} weitere
                          </div>
                        )}
                      </div>
                    </>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Legend */}
      {projects.length > 0 && (
        <div className="legend">
          <h3 className="legendTitle">Projekte</h3>
          <div className="legendItems">
            {projects.map((p, idx) => (
              <div
                key={p.id}
                className="legendItem"
                onClick={() => navigate(`/projects/${p.id}`)}
              >
                <span
                  className="legendDot"
                  style={{ backgroundColor: getColorForProject(idx) }}
                />
                <span className="legendName">{p.name}</span>
                <span className="legendDate">
                  {new Date(p.start_date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })}
                  {' - '}
                  {new Date(p.end_date).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}

      {!isLoading && projects.length === 0 && (
        <div className="emptyState">
          <div className="emptyIcon">📅</div>
          <h3 className="emptyTitle">Keine Projekte vorhanden</h3>
          <p className="emptyDescription">
            Erstellen Sie ein Projekt, um es im Kalender zu sehen.
          </p>
          <button className="btnPrimary" onClick={() => navigate('/projects/new')}>
            Projekt erstellen
          </button>
        </div>
      )}
    </div>
  )
}

export default CalendarPage
