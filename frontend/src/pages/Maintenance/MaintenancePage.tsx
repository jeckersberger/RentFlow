import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import './Maintenance.module.scss'

interface MaintenanceTask {
  id: string
  equipment_id: string
  equipment_name: string
  plan_id: string
  plan_name: string
  status: 'pending' | 'in_progress' | 'completed' | 'cancelled'
  priority: 'low' | 'medium' | 'high' | 'critical'
  due_date: string
  completed_date?: string
}

interface ECheckResult {
  id: string
  equipment_id: string
  equipment_name: string
  result_date: string
  status: 'pass' | 'warning' | 'fail'
  measurements: Array<{ name: string; value: string; status: 'ok' | 'warning' | 'fail' }>
}

interface MaintenancePlan {
  id: string
  equipment_id: string
  equipment_name: string
  frequency: string
  next_maintenance: string
  status: 'active' | 'completed' | 'paused'
}

// Mock data
const mockTasks: MaintenanceTask[] = [
  {
    id: '1',
    equipment_id: '1',
    equipment_name: 'JBL VTX A12',
    plan_id: '1',
    plan_name: 'Vierteljährliche Kontrolle',
    status: 'pending',
    priority: 'critical',
    due_date: '2026-03-15',
  },
  {
    id: '2',
    equipment_id: '3',
    equipment_name: 'MA Lighting grandMA3',
    plan_id: '2',
    plan_name: 'Monatliche Kontrolle',
    status: 'pending',
    priority: 'high',
    due_date: '2026-03-20',
  },
  {
    id: '3',
    equipment_id: '7',
    equipment_name: 'Yamaha CL5',
    plan_id: '3',
    plan_name: 'Nach-Einsatz-Kontrolle',
    status: 'in_progress',
    priority: 'high',
    due_date: '2026-03-25',
  },
  {
    id: '4',
    equipment_id: '2',
    equipment_name: 'Shure SM58',
    plan_id: '4',
    plan_name: 'Halbjährliche Überprüfung',
    status: 'completed',
    priority: 'medium',
    due_date: '2026-03-10',
    completed_date: '2026-03-10',
  },
  {
    id: '5',
    equipment_id: '5',
    equipment_name: 'Blackmagic ATEM Mini Extreme',
    plan_id: '5',
    plan_name: 'Vierteljährliche Kontrolle',
    status: 'pending',
    priority: 'medium',
    due_date: '2026-04-05',
  },
]

const mockEChecks: ECheckResult[] = [
  {
    id: '1',
    equipment_id: '2',
    equipment_name: 'Shure SM58',
    result_date: '2026-03-10T10:30:00Z',
    status: 'pass',
    measurements: [
      { name: 'Impedanz', value: '300 Ω', status: 'ok' },
      { name: 'Ausgangspegel', value: '-35 dBV', status: 'ok' },
      { name: 'Frequenzgang', value: '50-15000 Hz', status: 'ok' },
    ],
  },
  {
    id: '2',
    equipment_id: '10',
    equipment_name: 'Robe MegaPointe',
    result_date: '2026-03-08T14:15:00Z',
    status: 'warning',
    measurements: [
      { name: 'Lampenlebensdauer', value: '87%', status: 'warning' },
      { name: 'Lüfter', value: 'Normal', status: 'ok' },
      { name: 'Bewegung Pan/Tilt', value: 'Leichte Verzögerung', status: 'warning' },
    ],
  },
  {
    id: '3',
    equipment_id: '6',
    equipment_name: 'Prolyte X30V Truss 3m',
    result_date: '2026-03-05T09:00:00Z',
    status: 'pass',
    measurements: [
      { name: 'Inspektionszeichen', value: 'Gültig bis 2027', status: 'ok' },
      { name: 'Verschleiß', value: 'Minimal', status: 'ok' },
      { name: 'Struktur', value: 'Intact', status: 'ok' },
    ],
  },
]

const mockPlans: MaintenancePlan[] = [
  {
    id: '1',
    equipment_id: '1',
    equipment_name: 'JBL VTX A12',
    frequency: 'Vierteljährlich',
    next_maintenance: '2026-03-15',
    status: 'active',
  },
  {
    id: '2',
    equipment_id: '3',
    equipment_name: 'MA Lighting grandMA3',
    frequency: 'Monatlich',
    next_maintenance: '2026-03-20',
    status: 'active',
  },
  {
    id: '3',
    equipment_id: '7',
    equipment_name: 'Yamaha CL5',
    frequency: 'Nach Einsatz',
    next_maintenance: '2026-03-25',
    status: 'active',
  },
]

function MaintenancePage() {
  const navigate = useNavigate()
  const [expandedTask, setExpandedTask] = useState<string | null>(null)

  const { data: tasks = mockTasks } = useQuery({
    queryKey: ['maintenance-tasks'],
    queryFn: async () => mockTasks,
    staleTime: 1000 * 60 * 5,
  })

  const { data: eChecks = mockEChecks } = useQuery({
    queryKey: ['echeck-results'],
    queryFn: async () => mockEChecks,
    staleTime: 1000 * 60 * 5,
  })

  const { data: plans = mockPlans } = useQuery({
    queryKey: ['maintenance-plans'],
    queryFn: async () => mockPlans,
    staleTime: 1000 * 60 * 5,
  })

  const now = new Date()
  const overdueTasks = tasks.filter(t => new Date(t.due_date) < now && t.status === 'pending')
  const dueTasks = tasks.filter(t => {
    const dueDate = new Date(t.due_date)
    return dueDate >= now && dueDate <= new Date(now.getTime() + 7 * 24 * 60 * 60 * 1000) && t.status === 'pending'
  })
  const activePlans = plans.filter(p => p.status === 'active')

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'critical':
        return 'critical'
      case 'high':
        return 'high'
      case 'medium':
        return 'medium'
      case 'low':
        return 'low'
      default:
        return 'medium'
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' })
  }

  return (
    <div className="maintenance-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Wartung & Inspektion</h1>
          <p className="page-subtitle">Verwaltung von Wartungsplänen und Inspektionen</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--primary"
            onClick={() => navigate('/maintenance/new-task')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            ➕ Neue Aufgabe
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/maintenance/new-echeck')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            ✓ E-Check durchführen
          </button>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="stats-grid">
        <div className={`stat-card ${activePlans.length > 0 ? '' : 'stat-card--danger'}`}>
          <div className="stat-card__label">Aktive Pläne</div>
          <div className="stat-card__value">{activePlans.length}</div>
        </div>
        <div className={`stat-card ${tasks.filter(t => t.status === 'pending').length === 0 ? '' : 'stat-card--warning'}`}>
          <div className="stat-card__label">Offene Aufgaben</div>
          <div className="stat-card__value">{tasks.filter(t => t.status === 'pending').length}</div>
        </div>
        <div className={`stat-card ${overdueTasks.length === 0 ? '' : 'stat-card--danger'}`}>
          <div className="stat-card__label">Überfällige Aufgaben</div>
          <div className="stat-card__value">{overdueTasks.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-card__label">E-Checks diesen Monat</div>
          <div className="stat-card__value">{eChecks.length}</div>
        </div>
      </div>

      {/* Overdue Tasks Alert */}
      {overdueTasks.length > 0 && (
        <div className="alert-section">
          <h3 className="alert-section__title">⚠️ Überfällige Aufgaben</h3>
          <div className="alert-section__items">
            {overdueTasks.map(task => (
              <div key={task.id} className="alert-item">
                <div className="alert-item__content">
                  <p className="alert-item__name">{task.equipment_name}</p>
                  <p className="alert-item__meta">
                    Fällig seit: {formatDate(task.due_date)} • {task.plan_name}
                  </p>
                </div>
                <button
                  className="btn btn--sm btn--primary"
                  onClick={() => navigate(`/maintenance/tasks/${task.id}`)}
                >
                  Öffnen
                </button>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Due Tasks Section */}
      {dueTasks.length > 0 && (
        <div className="section-card">
          <h2 className="section-card__title">📋 Nächste Aufgaben (nächste 7 Tage)</h2>
          <div className="section-card__content">
            {dueTasks.map(task => (
              <div
                key={task.id}
                className="task-item"
                onClick={() => navigate(`/maintenance/tasks/${task.id}`)}
              >
                <div className="task-item__header">
                  <div className="task-item__checkbox">
                    {task.status === 'completed' && '✓'}
                  </div>
                  <div className="task-item__info">
                    <p className="task-item__title">{task.equipment_name}</p>
                    <p className="task-item__meta">
                      {task.plan_name} • Fällig: {formatDate(task.due_date)}
                    </p>
                  </div>
                </div>
                <div className="task-item__badges">
                  <span className={`priority-badge priority-badge--${getPriorityColor(task.priority)}`}>
                    {task.priority.toUpperCase()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* E-Check Results */}
      <div className="section-card">
        <h2 className="section-card__title">✓ Aktuelle E-Check Ergebnisse</h2>
        <div className="section-card__content">
          {eChecks.length === 0 ? (
            <div className="empty-state" style={{ minHeight: '200px' }}>
              <div className="empty-state__icon">📊</div>
              <h3 className="empty-state__title">Keine E-Check Ergebnisse</h3>
              <p className="empty-state__description">Führen Sie einen E-Check durch, um die Ergebnisse hier zu sehen.</p>
            </div>
          ) : (
            eChecks.slice(0, 3).map(eCheck => (
              <div key={eCheck.id} style={{ paddingBottom: 'var(--spacing-3)', borderBottom: 'var(--card-border-width) solid var(--color-border)', marginBottom: 'var(--spacing-3)' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-2)' }}>
                  <div>
                    <p style={{ margin: 0, fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)', fontSize: 'var(--font-size-sm)' }}>
                      {eCheck.equipment_name}
                    </p>
                    <p style={{ margin: '0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                      {new Date(eCheck.result_date).toLocaleString('de-DE')}
                    </p>
                  </div>
                  <StatusBadge status={eCheck.status} />
                </div>
                <div className="echeck-results">
                  {eCheck.measurements.map((measurement, idx) => (
                    <div key={idx} className="echeck-item">
                      <div className="echeck-item__label">{measurement.name}</div>
                      <div className="echeck-item__value">
                        <div className="echeck-item__measurement">{measurement.value}</div>
                        <div className={`echeck-item__status echeck-item__status--${measurement.status}`}>
                          {measurement.status === 'ok' ? '✓ OK' : measurement.status === 'warning' ? '⚠ Warnung' : '✗ Fehler'}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            ))
          )}
        </div>
      </div>

      {/* Maintenance Plans Overview */}
      <div className="section-card">
        <h2 className="section-card__title">📅 Aktive Wartungspläne</h2>
        <div className="section-card__content">
          {activePlans.length === 0 ? (
            <div className="empty-state" style={{ minHeight: '200px' }}>
              <div className="empty-state__icon">📋</div>
              <h3 className="empty-state__title">Keine Wartungspläne</h3>
              <p className="empty-state__description">Erstellen Sie einen Wartungsplan, um Inspektionen zu planen.</p>
            </div>
          ) : (
            <div className="calendar-list">
              {activePlans
                .sort((a, b) => new Date(a.next_maintenance).getTime() - new Date(b.next_maintenance).getTime())
                .map(plan => {
                  const nextDate = new Date(plan.next_maintenance)
                  const daysUntil = Math.ceil((nextDate.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))

                  return (
                    <div
                      key={plan.id}
                      className="calendar-item"
                      onClick={() => navigate(`/maintenance/plans/${plan.id}`)}
                    >
                      <div className="calendar-item__date">
                        <div className="calendar-item__date-day">{nextDate.getDate()}</div>
                        <div className="calendar-item__date-month">
                          {nextDate.toLocaleDateString('de-DE', { month: 'short' })}
                        </div>
                      </div>
                      <div className="calendar-item__content">
                        <p className="calendar-item__title" style={{ margin: 0, fontWeight: 'var(--font-weight-semibold)', color: 'var(--color-text-primary)' }}>
                          {plan.equipment_name}
                        </p>
                        <p className="calendar-item__equipment" style={{ margin: '0 0 var(--spacing-1) 0' }}>
                          {plan.frequency}
                        </p>
                        <p style={{ margin: 0, fontSize: 'var(--font-size-xs)', color: daysUntil <= 3 ? 'var(--color-danger)' : 'var(--color-text-secondary)' }}>
                          {daysUntil <= 0 ? 'Heute fällig' : daysUntil === 1 ? 'Morgen' : `In ${daysUntil} Tagen`}
                        </p>
                      </div>
                    </div>
                  )
                })}
            </div>
          )}
        </div>
      </div>

      {/* All Tasks */}
      <div className="section-card">
        <h2 className="section-card__title">📊 Alle Aufgaben ({tasks.length})</h2>
        <div className="section-card__content">
          {tasks.length === 0 ? (
            <div className="empty-state" style={{ minHeight: '200px' }}>
              <div className="empty-state__icon">✓</div>
              <h3 className="empty-state__title">Keine Aufgaben</h3>
              <p className="empty-state__description">Alle Wartungsaufgaben sind abgeschlossen.</p>
            </div>
          ) : (
            tasks.map(task => (
              <div
                key={task.id}
                className="task-item"
                onClick={() => setExpandedTask(expandedTask === task.id ? null : task.id)}
              >
                <div className="task-item__header">
                  <div className={`task-item__checkbox ${task.status === 'completed' ? 'task-item__checkbox--checked' : ''}`}>
                    {task.status === 'completed' && '✓'}
                  </div>
                  <div className="task-item__info">
                    <p className="task-item__title">{task.equipment_name}</p>
                    <p className="task-item__meta">
                      {task.plan_name} • Fällig: {formatDate(task.due_date)}
                    </p>
                  </div>
                </div>
                <div className="task-item__badges">
                  <span className={`priority-badge priority-badge--${getPriorityColor(task.priority)}`}>
                    {task.priority.charAt(0).toUpperCase() + task.priority.slice(1)}
                  </span>
                  <StatusBadge status={task.status} />
                  <span style={{ marginLeft: 'var(--spacing-2)' }}>
                    {expandedTask === task.id ? '▼' : '▶'}
                  </span>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  )
}

export default MaintenancePage
