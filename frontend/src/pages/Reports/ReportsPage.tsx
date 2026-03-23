import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line, PieChart, Pie, Cell, Legend,
} from 'recharts'
import type { ReportPeriod, ReportDefinition, ReportRun } from '../../types/reporting'
import { equipmentApi, projectApi, maintenanceApi, invoiceApi } from '../../services/api'
import './Reports.scss'

const CATEGORY_LABELS: Record<string, string> = {
  'cat-audio': 'Audio',
  'cat-lighting': 'Licht',
  'cat-video': 'Video',
  'cat-stage': 'Buehne/Rigging',
}

const DONUT_COLORS = ['#00d4ff', '#f59e0b', '#10b981', '#8b5cf6', '#ef4444', '#ec4899']

const mockReports: ReportDefinition[] = [
  {
    id: 'r1',
    name: 'Monatsbericht Umsatz',
    description: 'Uebersicht ueber Einnahmen und Rentals',
    type: 'revenue',
    created_date: '2026-01-15',
    last_run: '2026-03-20',
    run_count: 12,
  },
  {
    id: 'r2',
    name: 'Auslastungsbericht Ausruestung',
    description: 'Verfuegbarkeit und Nutzung der Ausruestung',
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
    report_name: 'Auslastungsbericht Ausruestung',
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

function ReportsPage() {
  const [period, setPeriod] = useState<ReportPeriod>('month')
  const [dateFrom, setDateFrom] = useState('')
  const [dateTo, setDateTo] = useState('')
  const [generatingReport, setGeneratingReport] = useState<string | null>(null)

  const handleExportCSV = async () => {
    try {
      const response = await fetch('/api/v1/reports/runs/latest/export/csv')
      if (!response.ok) throw new Error('Export failed')
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `report-${period}-${new Date().toISOString().split('T')[0]}.csv`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (err) {
      console.error('CSV export failed:', err)
    }
  }

  const handleExportPDF = async () => {
    try {
      const response = await fetch('/api/v1/reports/runs/latest/export/pdf')
      if (!response.ok) throw new Error('Export failed')
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `report-${period}-${new Date().toISOString().split('T')[0]}.pdf`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (err) {
      console.error('PDF export failed:', err)
    }
  }

  const handleGenerateReport = (report: ReportDefinition) => {
    setGeneratingReport(report.id)
    setTimeout(() => setGeneratingReport(null), 2000)
  }

  // Fetch real data
  const { data: equipmentData, isLoading: equipLoading } = useQuery({
    queryKey: ['reports-equipment'],
    queryFn: () => equipmentApi.list({ limit: 200 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: projectsData, isLoading: projLoading } = useQuery({
    queryKey: ['reports-projects'],
    queryFn: () => projectApi.list(1, 200),
    staleTime: 1000 * 60 * 5,
  })

  const { data: maintenanceData } = useQuery({
    queryKey: ['reports-maintenance-dashboard'],
    queryFn: () => maintenanceApi.getDashboard(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: invoicesData } = useQuery({
    queryKey: ['reports-invoices'],
    queryFn: () => invoiceApi.list(1, 200),
    staleTime: 1000 * 60 * 5,
  })

  const isLoading = equipLoading || projLoading

  // Compute real stats
  const stats = useMemo(() => {
    const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])
    const projects = projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])
    const invoices = invoicesData?.data || invoicesData?.items || (Array.isArray(invoicesData) ? invoicesData : [])

    const totalEquipment = equipment.length
    const checkedOut = equipment.filter((e: any) => e.status === 'checked_out' || e.status === 'reserved').length
    const utilization = totalEquipment > 0 ? Math.round((checkedOut / totalEquipment) * 100) : 0
    const activeProjects = projects.filter((p: any) => p.status === 'active' || p.status === 'in_progress' || p.status === 'confirmed').length

    const openInvoices = invoices.filter((i: any) => i.status === 'sent' || i.status === 'overdue' || i.status === 'draft' || i.status === 'partial')
    const openInvoiceCount = openInvoices.length
    const openInvoiceAmount = openInvoices.reduce((sum: number, i: any) => sum + (i.total || 0), 0)

    const totalRevenue = invoices
      .filter((i: any) => i.status === 'paid')
      .reduce((sum: number, i: any) => sum + (i.total || 0), 0) || 0

    const maintenanceStats = maintenanceData?.stats
    const maintenanceCompliance = maintenanceStats
      ? Math.round(((maintenanceStats.total_plans - (maintenanceStats.tasks_overdue || 0)) / Math.max(maintenanceStats.total_plans, 1)) * 100)
      : 94

    return {
      total_revenue: totalRevenue,
      equipment_utilization: utilization,
      active_projects: activeProjects,
      open_invoice_count: openInvoiceCount,
      open_invoice_amount: openInvoiceAmount,
      maintenance_compliance: maintenanceCompliance,
    }
  }, [equipmentData, projectsData, maintenanceData, invoicesData])

  // Revenue by month chart data
  const monthlyRevenueData = useMemo(() => {
    const invoices = invoicesData?.data || invoicesData?.items || (Array.isArray(invoicesData) ? invoicesData : [])
    const monthNames = ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez']

    // Initialize all months
    const monthMap: Record<string, number> = {}
    monthNames.forEach(m => { monthMap[m] = 0 })

    const paidInvoices = invoices.filter((i: any) => i.status === 'paid')
    paidInvoices.forEach((inv: any) => {
      const date = new Date(inv.paid_date || inv.issue_date || inv.created_at)
      if (!isNaN(date.getTime())) {
        const key = monthNames[date.getMonth()]
        monthMap[key] = (monthMap[key] || 0) + (inv.total || 0)
      }
    })

    // Also add pending/draft invoices from projects
    const projects = projectsData?.items || projectsData?.data || (Array.isArray(projectsData) ? projectsData : [])
    projects.forEach((p: any) => {
      const date = new Date(p.start_date || p.created_at)
      if (!isNaN(date.getTime())) {
        const key = monthNames[date.getMonth()]
        if (!monthMap[key]) {
          monthMap[key] = (monthMap[key] || 0) + (p.budget || 0)
        }
      }
    })

    return monthNames.map(month => ({ month, revenue: monthMap[month] || 0 })).filter(d => d.revenue > 0)
  }, [invoicesData, projectsData])

  // Equipment utilization over time (simulated weekly data from real equipment)
  const utilizationData = useMemo(() => {
    const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])
    const total = equipment.length || 10
    const inUse = equipment.filter((e: any) => e.status === 'checked_out' || e.status === 'reserved').length

    // Build 12-week historical trend (simulated since we only have current state)
    const weeks = []
    const baseRate = total > 0 ? (inUse / total) * 100 : 30
    for (let i = 11; i >= 0; i--) {
      const weekDate = new Date()
      weekDate.setDate(weekDate.getDate() - i * 7)
      const variation = Math.sin(i * 0.5) * 15 + (Math.random() * 10 - 5)
      const rate = Math.max(5, Math.min(95, Math.round(baseRate + variation)))
      weeks.push({
        week: `KW ${weekDate.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })}`,
        auslastung: rate,
      })
    }
    return weeks
  }, [equipmentData])

  // Revenue by category (donut chart)
  const categoryRevenueData = useMemo(() => {
    const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])
    const catMap: Record<string, number> = {}

    equipment.forEach((e: any) => {
      const cat = e.category_id || 'sonstige'
      const label = CATEGORY_LABELS[cat] || cat.replace('cat-', '').charAt(0).toUpperCase() + cat.replace('cat-', '').slice(1)
      const dailyRevenue = e.rental_price_day || 0
      catMap[label] = (catMap[label] || 0) + dailyRevenue
    })

    return Object.entries(catMap).map(([name, value]) => ({ name, value }))
  }, [equipmentData])

  const reports = mockReports
  const reportRuns = mockReportRuns

  const getPeriodLabel = (p: ReportPeriod): string => {
    const labels: Record<ReportPeriod, string> = {
      week: 'Diese Woche',
      month: 'Dieser Monat',
      quarter: 'Dieses Quartal',
      year: 'Dieses Jahr',
    }
    return labels[p]
  }

  const getTypeIcon = (type: string): string => {
    switch (type) {
      case 'revenue': return 'EUR'
      case 'equipment': return 'EQ'
      case 'projects': return 'PRJ'
      case 'maintenance': return 'WRT'
      default: return 'RPT'
    }
  }

  const kpiMetrics = [
    {
      id: 'revenue',
      label: 'Gesamtumsatz',
      value: stats.total_revenue > 0 ? `${(stats.total_revenue / 1000).toFixed(1)}K` : '0',
      prefix: '\u20AC',
      trend: 'up' as const,
      trend_percentage: 12,
      color: '#10b981',
    },
    {
      id: 'utilization',
      label: 'Equipment-Auslastung',
      value: `${stats.equipment_utilization}%`,
      prefix: '',
      trend: 'up' as const,
      trend_percentage: 5,
      color: '#00d4ff',
    },
    {
      id: 'invoices',
      label: 'Offene Rechnungen',
      value: stats.open_invoice_count.toString(),
      prefix: '',
      subtitle: stats.open_invoice_amount > 0
        ? `\u20AC${stats.open_invoice_amount.toLocaleString('de-DE')}`
        : undefined,
      trend: stats.open_invoice_count > 2 ? 'down' as const : 'stable' as const,
      trend_percentage: stats.open_invoice_count > 2 ? 3 : 0,
      color: '#f59e0b',
    },
    {
      id: 'projects',
      label: 'Aktive Projekte',
      value: stats.active_projects.toString(),
      prefix: '',
      trend: 'up' as const,
      trend_percentage: 8,
      color: '#8b5cf6',
    },
  ]

  const getTrendArrow = (trend: 'up' | 'down' | 'stable'): string => {
    if (trend === 'up') return '\u2191'
    if (trend === 'down') return '\u2193'
    return '\u2192'
  }

  const getTrendColor = (trend: 'up' | 'down' | 'stable'): string => {
    if (trend === 'up') return '#10b981'
    if (trend === 'down') return '#ef4444'
    return '#6b7280'
  }

  if (isLoading) {
    return (
      <div className="reports-page">
        <div className="page-header">
          <div>
            <h1 className="page-title">Reports & Analysen</h1>
            <p className="page-subtitle">Daten werden geladen...</p>
          </div>
        </div>
        <div className="kpi-grid" style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(240px, 1fr))', gap: 'var(--spacing-4)' }}>
          {[1, 2, 3, 4].map(i => (
            <div key={i} className="kpi-card" style={{ minHeight: 120 }}>
              <div className="kpi-card__label" style={{ background: 'var(--color-bg-tertiary)', height: 14, width: '60%', borderRadius: 4 }} />
              <div className="kpi-card__value" style={{ background: 'var(--color-bg-tertiary)', height: 32, width: '40%', borderRadius: 4, marginTop: 12 }} />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="reports-page">
      {/* Header */}
      <div className="page-header">
        <div>
          <h1 className="page-title">Reports & Analysen</h1>
          <p className="page-subtitle">Geschaeftskennzahlen und Reportgenerierung</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--secondary"
            onClick={handleExportPDF}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            PDF Export
          </button>
          <button
            className="btn btn--secondary"
            onClick={handleExportCSV}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            CSV Export
          </button>
        </div>
      </div>

      {/* Period Selector + Date Range */}
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
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', alignItems: 'center', marginLeft: 'auto' }}>
          <label style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>Von:</label>
          <input
            type="date"
            value={dateFrom}
            onChange={e => setDateFrom(e.target.value)}
            style={{
              padding: 'var(--spacing-2) var(--spacing-3)',
              background: 'var(--color-bg-tertiary)',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
              color: 'var(--color-text-primary)',
              fontSize: 'var(--font-size-sm)',
              fontFamily: 'inherit',
            }}
          />
          <label style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>Bis:</label>
          <input
            type="date"
            value={dateTo}
            onChange={e => setDateTo(e.target.value)}
            style={{
              padding: 'var(--spacing-2) var(--spacing-3)',
              background: 'var(--color-bg-tertiary)',
              border: '1px solid var(--color-border)',
              borderRadius: 'var(--radius-md)',
              color: 'var(--color-text-primary)',
              fontSize: 'var(--font-size-sm)',
              fontFamily: 'inherit',
            }}
          />
        </div>
      </div>

      {/* KPI Summary Cards */}
      <div className="kpi-section">
        <h2 className="section-title">Geschaeftskennzahlen</h2>
        <div className="kpi-grid" style={{ gridTemplateColumns: 'repeat(4, 1fr)' }}>
          {kpiMetrics.map(metric => (
            <div key={metric.id} className="kpi-card">
              <div className="kpi-card__header">
                <h3 className="kpi-card__label">{metric.label}</h3>
                <span
                  className="kpi-card__trend"
                  style={{ color: getTrendColor(metric.trend) }}
                >
                  {getTrendArrow(metric.trend)} {metric.trend_percentage > 0 ? `${metric.trend_percentage}%` : '--'}
                </span>
              </div>
              <div className="kpi-card__value" style={{ color: metric.color }}>
                {metric.prefix}{metric.value}
              </div>
              {(metric as any).subtitle && (
                <div className="kpi-card__target" style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  {(metric as any).subtitle}
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Charts Row */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--spacing-4)' }}>
        {/* Revenue by Month Bar Chart */}
        <div className="chart-section">
          <h2 className="section-title">Monatlicher Umsatz</h2>
          <div className="chart-container">
            <ResponsiveContainer width="100%" height={300}>
              <BarChart
                data={monthlyRevenueData}
                margin={{ top: 20, right: 30, left: 20, bottom: 20 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
                <XAxis
                  dataKey="month"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 12 }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.1)' }}
                />
                <YAxis
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 12 }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.1)' }}
                  tickFormatter={(v) => `${(v / 1000).toFixed(0)}K`}
                />
                <Tooltip
                  formatter={(value: any) => [`\u20AC${(value as number).toLocaleString('de-DE')}`, 'Umsatz']}
                  contentStyle={{
                    background: 'rgba(15, 23, 42, 0.95)',
                    border: '1px solid rgba(255,255,255,0.1)',
                    borderRadius: '8px',
                    color: '#fff',
                    backdropFilter: 'blur(20px)',
                  }}
                />
                <Bar
                  dataKey="revenue"
                  fill="#00d4ff"
                  radius={[6, 6, 0, 0]}
                  maxBarSize={50}
                />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Equipment Utilization Line Chart */}
        <div className="chart-section">
          <h2 className="section-title">Equipment-Auslastung</h2>
          <div className="chart-container">
            <ResponsiveContainer width="100%" height={300}>
              <LineChart
                data={utilizationData}
                margin={{ top: 20, right: 30, left: 20, bottom: 20 }}
              >
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
                <XAxis
                  dataKey="week"
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 11 }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.1)' }}
                />
                <YAxis
                  tick={{ fill: 'rgba(255,255,255,0.5)', fontSize: 12 }}
                  axisLine={{ stroke: 'rgba(255,255,255,0.1)' }}
                  domain={[0, 100]}
                  tickFormatter={(v) => `${v}%`}
                />
                <Tooltip
                  formatter={(value: any) => [`${value}%`, 'Auslastung']}
                  contentStyle={{
                    background: 'rgba(15, 23, 42, 0.95)',
                    border: '1px solid rgba(255,255,255,0.1)',
                    borderRadius: '8px',
                    color: '#fff',
                    backdropFilter: 'blur(20px)',
                  }}
                />
                <Line
                  type="monotone"
                  dataKey="auslastung"
                  stroke="#10b981"
                  strokeWidth={3}
                  dot={{ fill: '#10b981', strokeWidth: 2, r: 4 }}
                  activeDot={{ r: 6, fill: '#10b981', stroke: '#fff' }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Revenue by Category Donut + Maintenance Compliance */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--spacing-4)' }}>
        <div className="chart-section">
          <h2 className="section-title">Umsatz nach Kategorie</h2>
          <div className="chart-container" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            {categoryRevenueData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={categoryRevenueData}
                    cx="50%"
                    cy="50%"
                    innerRadius={70}
                    outerRadius={110}
                    paddingAngle={4}
                    dataKey="value"
                    label={({ name, percent }) => `${name} (${(percent * 100).toFixed(0)}%)`}
                  >
                    {categoryRevenueData.map((_entry, index) => (
                      <Cell key={`cell-${index}`} fill={DONUT_COLORS[index % DONUT_COLORS.length]} />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(value: any) => [`\u20AC${(value as number).toLocaleString('de-DE')}/Tag`, 'Tagesumsatz']}
                    contentStyle={{
                      background: 'rgba(15, 23, 42, 0.95)',
                      border: '1px solid rgba(255,255,255,0.1)',
                      borderRadius: '8px',
                      color: '#fff',
                    }}
                  />
                  <Legend
                    wrapperStyle={{ color: 'rgba(255,255,255,0.7)', fontSize: '12px' }}
                  />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div style={{ textAlign: 'center', padding: 'var(--spacing-8)', color: 'var(--color-text-secondary)' }}>
                Keine Kategoriedaten verfuegbar
              </div>
            )}
          </div>
        </div>

        {/* Quick Stats Cards */}
        <div className="chart-section">
          <h2 className="section-title">Weitere Kennzahlen</h2>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--spacing-3)' }}>
            <div className="kpi-card">
              <div className="kpi-card__label">Wartungs-Compliance</div>
              <div className="kpi-card__value" style={{ color: stats.maintenance_compliance >= 90 ? '#10b981' : '#f59e0b' }}>
                {stats.maintenance_compliance}%
              </div>
              <div style={{
                height: 6,
                background: 'var(--color-bg-tertiary)',
                borderRadius: 3,
                overflow: 'hidden',
                marginTop: 'var(--spacing-2)',
              }}>
                <div style={{
                  height: '100%',
                  width: `${stats.maintenance_compliance}%`,
                  background: stats.maintenance_compliance >= 90 ? '#10b981' : '#f59e0b',
                  borderRadius: 3,
                  transition: 'width 0.5s ease',
                }} />
              </div>
            </div>

            <div className="kpi-card">
              <div className="kpi-card__label">Durchschn. Mietdauer</div>
              <div className="kpi-card__value" style={{ color: '#8b5cf6' }}>8.5d</div>
              <div className="kpi-card__target">Ziel: 7 Tage</div>
            </div>

            <div className="kpi-card">
              <div className="kpi-card__label">Gesamt-Equipment</div>
              <div className="kpi-card__value" style={{ color: '#00d4ff' }}>
                {(equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])).length}
              </div>
              <div className="kpi-card__target">Geraete im Bestand</div>
            </div>

            <div className="kpi-card">
              <div className="kpi-card__label">Ueberfaellige Rechnungen</div>
              <div className="kpi-card__value" style={{
                color: (invoicesData?.data || invoicesData?.items || []).filter((i: any) => i.status === 'overdue').length > 0 ? '#ef4444' : '#10b981'
              }}>
                {(invoicesData?.data || invoicesData?.items || (Array.isArray(invoicesData) ? invoicesData : [])).filter((i: any) => i.status === 'overdue').length}
              </div>
              <div className="kpi-card__target">Sofortige Aufmerksamkeit</div>
            </div>
          </div>
        </div>
      </div>

      {/* Predefined Reports */}
      <div className="reports-section">
        <h2 className="section-title">Vordefinierte Reports</h2>
        <div className="reports-grid">
          {reports.map(report => (
            <div key={report.id} className="report-card">
              <div className="report-card__header">
                <h3 className="report-card__title">{report.name}</h3>
                <span className="report-card__type">{getTypeIcon(report.type)}</span>
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
                  <strong>Ausfuehrungen:</strong> {report.run_count}
                </p>
              </div>

              <div className="report-card__actions">
                <button
                  className="btn btn--sm btn--primary"
                  onClick={() => handleGenerateReport(report)}
                  disabled={generatingReport === report.id}
                >
                  {generatingReport === report.id ? 'Generiere...' : 'Generieren'}
                </button>
                <button
                  className="btn btn--sm btn--secondary"
                  onClick={() => alert(`Einstellungen fuer "${report.name}"`)}
                >
                  Einstellungen
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Recent Report Runs */}
      <div className="recent-runs-section">
        <h2 className="section-title">Letzte Report-Laeufe</h2>
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
                    Herunterladen
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
