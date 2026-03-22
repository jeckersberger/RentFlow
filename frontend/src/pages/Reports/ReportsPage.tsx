import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import type { ReportPeriod, ReportDefinition, ReportRun, ReportingStats } from '../../types/reporting'
import './Reports.module.scss'

const mockStats: ReportingStats = {
  total_revenue: 1250000,
  equipment_utilization: 87,
  active_projects: 12,
  overdue_invoices: 3,
  avg_rental_days: 8.5,
  maintenance_compliance: 94,
}

const mockReports: ReportDefinition[] = [
  {
    id: 'r1',
    name: 'Monatsbericht Umsatz',
    description: 'Übersicht über Einnahmen und Rentals',
    type: 'revenue',
    created_date: '2026-01-15',
    last_run: '2026-03-20',
    run_count: 12,
  },
  {
    id: 'r2',
    name: 'Auslastungsbericht Ausrüstung',
    description: 'Verfügbarkeit und Nutzung der Ausrüstung',
    type: 'equipment',
    created_date: '2026-02-01',
    last_run: '2026-03-21',
    run_count: 8,
  },
  {
    id: 'r3',
    name: 'Projektfortschritt Analyse',
    description: 'Status aktiver Projekte und Meilensteine',
    type: 'projects',
    created_date: '2026-01-20',
    last_run: '2026-03-22',
    run_count: 10,
  },
  {
    id: 'r4',
    name: 'Wartungs- und Compliance-Bericht',
    description: 'Wartungsarbeiten und Zertifizierungen',
    type: 'maintenance',
    created_date: '2026-02-10',
    last_run: '2026-03-18',
    run_count: 6,
  },
]

const mockReportRuns: ReportRun[] = [
  {
    id: 'run1',
    report_id: 'r1',
    report_name: 'Monatsbericht Umsatz',
    period: 'month',
    generated_date: '2026-03-20T14:30:00Z',
    generated_by: 'admin',
  },
  {
    id: 'run2',
    report_id: 'r2',
    report_name: 'Auslastungsbericht Ausrüstung',
    period: 'week',
    generated_date: '2026-03-21T09:15:00Z',
    generated_by: 'user1',
  },
  {
    id: 'run3',
    report_id: 'r3',
    report_name: 'Projektfortschritt Analyse',
    period: 'month',
    generated_date: '2026-03-22T11:00:00Z',
    generated_by: 'admin',
  },
]

const mockMonthlyRevenueData = [
  { month: 'Januar', revenue: 125000 },
  { month: 'Februar', revenue: 148000 },
  { month: 'März', revenue: 167000 },
  { month: 'April', revenue: 142000 },
  { month: 'Mai', revenue: 189000 },
  { month: 'Juni', revenue: 201000 },
]

function ReportsPage() {
  const [period, setPeriod] = useState<ReportPeriod>('month')

  const { data: stats = mockStats } = useQuery({
    queryKey: ['reports-stats', period],
    queryFn: async () => mockStats,
    staleTime: 1000 * 60 * 5,
  })

  const { data: reports = mockReports } = useQuery({
    queryKey: ['reports'],
    queryFn: async () => mockReports,
    staleTime: 1000 * 60 * 5,
  })

  const { data: reportRuns = mockReportRuns } = useQuery({
    queryKey: ['report-runs', period],
    queryFn: async () => mockReportRuns,
    staleTime: 1000 * 60 * 5,
  })

  const getPeriodLabel = (p: ReportPeriod): string => {
    const labels: Record<ReportPeriod, string> = {
      week: 'Diese Woche',
      month: 'Dieser Monat',
      quarter: 'Dieses Quartal',
      year: 'Dieses Jahr',
    }
    return labels[p]
  }

  const getTrendIcon = (trend: 'up' | 'down' | 'stable'): string => {
    if (trend === 'up') return '📈'
    if (trend === 'down') return '📉'
    return '➡️'
  }

  const getTrendColor = (trend: 'up' | 'down' | 'stable'): string => {
    if (trend === 'up') return 'var(--color-success)'
    if (trend === 'down') return 'var(--color-danger)'
    return 'var(--color-text-secondary)'
  }

  const kpiMetrics = [
    {
      id: 'revenue',
      label: 'Gesamtumsatz',
      value: `€${(stats.total_revenue / 1000).toFixed(0)}K`,
      unit: '',
      trend: 'up' as const,
      trend_percentage: 12,
      target: 1100000,
    },
    {
      id: 'utilization',
      label: 'Ausrüstungsauslastung',
      value: `${stats.equipment_utilization}%`,
      unit: '',
      trend: 'up' as const,
      trend_percentage: 5,
      target: 85,
    },
    {
      id: 'projects',
      label: 'Aktive Projekte',
      value: stats.active_projects.toString(),
      unit: '',
      trend: 'up' as const,
      trend_percentage: 8,
      target: 10,
    },
    {
      id: 'overdue',
      label: 'Überfällige Rechnungen',
      value: stats.overdue_invoices.toString(),
      unit: '',
      trend: 'down' as const,
      trend_percentage: 2,
      target: 0,
    },
    {
      id: 'rental_days',
      label: 'Durchschnittliche Mietdauer',
      value: `${stats.avg_rental_days}d`,
      unit: '',
      trend: 'stable' as const,
      trend_percentage: 0,
      target: 7,
    },
    {
      id: 'compliance',
      label: 'Wartungs-Compliance',
      value: `${stats.maintenance_compliance}%`,
      unit: '',
      trend: 'up' as const,
      trend_percentage: 3,
      target: 95,
    },
  ]

  return (
    <div className="reports-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Reports & Analysen</h1>
          <p className="page-subtitle">Geschäftskennzahlen und Reportgenerierung</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--secondary"
            onClick={() => alert('PDF wird exportiert...')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            📥 PDF Exportieren
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => alert('CSV wird exportiert...')}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            📊 CSV Exportieren
          </button>
        </div>
      </div>

      {/* Period Selector */}
      <div className="period-selector">
        <label className="period-label">Zeitraum:</label>
        <div className="period-buttons">
          {(['week', 'month', 'quarter', 'year'] as ReportPeriod[]).map(p => (
            <button
              key={p}
              className={`period-button ${period === p ? 'period-button--active' : ''}`}
              onClick={() => setPeriod(p)}
            >
              {getPeriodLabel(p)}
            </button>
          ))}
        </div>
      </div>

      {/* KPI Dashboard */}
      <div className="kpi-section">
        <h2 className="section-title">📊 Geschäftskennzahlen</h2>
        <div className="kpi-grid">
          {kpiMetrics.map(metric => (
            <div key={metric.id} className="kpi-card">
              <div className="kpi-card__header">
                <h3 className="kpi-card__label">{metric.label}</h3>
                <span
                  className="kpi-card__trend"
                  style={{ color: getTrendColor(metric.trend) }}
                >
                  {getTrendIcon(metric.trend)} {metric.trend_percentage}%
                </span>
              </div>

              <div className="kpi-card__value">{metric.value}</div>

              <div className="kpi-card__target">
                {metric.target && (
                  <span>Ziel: {typeof metric.target === 'number' ? (
                    metric.label.includes('%') || metric.label.includes('Auslastung') || metric.label.includes('Compliance')
                      ? `${metric.target}%`
                      : `€${metric.target.toLocaleString('de-DE')}`
                  ) : metric.target}</span>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Monthly Revenue Chart */}
      <div className="chart-section">
        <h2 className="section-title">📈 Monatlicher Umsatz</h2>
        <div className="chart-container">
          <ResponsiveContainer width="100%" height={300}>
            <BarChart
              data={mockMonthlyRevenueData}
              margin={{ top: 20, right: 30, left: 0, bottom: 20 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" vertical={false} />
              <XAxis
                dataKey="month"
                tick={{ fill: 'var(--color-text-secondary)', fontSize: 12 }}
              />
              <YAxis
                tick={{ fill: 'var(--color-text-secondary)', fontSize: 12 }}
                label={{ value: 'Umsatz (€)', angle: -90, position: 'insideLeft' }}
              />
              <Tooltip
                formatter={(value: any) => `€${(value as number).toLocaleString('de-DE')}`}
                contentStyle={{
                  background: 'var(--color-bg-secondary)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-md)',
                  color: 'var(--color-text-primary)',
                }}
              />
              <Bar
                dataKey="revenue"
                fill="var(--color-primary)"
                radius={[8, 8, 0, 0]}
              />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Reports Definitions */}
      <div className="reports-section">
        <h2 className="section-title">📋 Vordefinierte Reports</h2>
        <div className="reports-grid">
          {reports.map(report => (
            <div key={report.id} className="report-card">
              <div className="report-card__header">
                <h3 className="report-card__title">{report.name}</h3>
                <span className="report-card__type">{report.type}</span>
              </div>

              <p className="report-card__description">{report.description}</p>

              <div className="report-card__meta">
                <p>
                  <strong>Erstellt:</strong> {new Date(report.created_date).toLocaleDateString('de-DE')}
                </p>
                <p>
                  <strong>Letzter Lauf:</strong> {report.last_run ? new Date(report.last_run).toLocaleDateString('de-DE') : 'Nie'}
                </p>
                <p>
                  <strong>Ausführungen:</strong> {report.run_count}
                </p>
              </div>

              <div className="report-card__actions">
                <button
                  className="btn btn--sm btn--primary"
                  onClick={() => alert(`Report "${report.name}" wird generiert...`)}
                >
                  ⚙️ Generieren
                </button>
                <button
                  className="btn btn--sm btn--secondary"
                  onClick={() => alert(`Einstellungen für "${report.name}" werden angezeigt...`)}
                >
                  ⚙️ Einstellungen
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Recent Report Runs */}
      <div className="recent-runs-section">
        <h2 className="section-title">📅 Letzte Report-Läufe</h2>
        <div className="runs-table">
          <div className="table-header">
            <div className="table-header__cell table-header__cell--name">Name</div>
            <div className="table-header__cell table-header__cell--period">Zeitraum</div>
            <div className="table-header__cell table-header__cell--date">Generiert am</div>
            <div className="table-header__cell table-header__cell--user">Durch</div>
            <div className="table-header__cell table-header__cell--action"></div>
          </div>
          <div className="table-body">
            {reportRuns.map(run => (
              <div key={run.id} className="table-row">
                <div className="table-cell table-cell--name">{run.report_name}</div>
                <div className="table-cell table-cell--period">
                  <span className="period-badge">{getPeriodLabel(run.period)}</span>
                </div>
                <div className="table-cell table-cell--date">
                  {new Date(run.generated_date).toLocaleDateString('de-DE')}
                </div>
                <div className="table-cell table-cell--user">{run.generated_by}</div>
                <div className="table-cell table-cell--action">
                  <button
                    className="btn btn--sm btn--secondary"
                    onClick={() => alert(`Report wird heruntergeladen...`)}
                  >
                    📥 Herunterladen
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ReportsPage
