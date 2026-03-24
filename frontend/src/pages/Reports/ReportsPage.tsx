import { useState, useMemo, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  PieChart, Pie, Cell, Legend,
} from 'recharts'
import type { ReportPeriod } from '../../types/reporting'
import { equipmentApi, projectApi, maintenanceApi, invoiceApi } from '../../services/api'
import './Reports.scss'

const CATEGORY_LABELS: Record<string, string> = {
  'cat-audio': 'Audio',
  'cat-lighting': 'Licht',
  'cat-video': 'Video',
  'cat-stage': 'Buehne/Rigging',
}

const DONUT_COLORS = ['#00d4ff', '#f59e0b', '#10b981', '#8b5cf6', '#ef4444', '#ec4899']

/** Compute start/end dates for a given period preset */
function getDateRange(period: ReportPeriod): { from: Date; to: Date } {
  const now = new Date()
  const to = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999)
  let from: Date

  switch (period) {
    case 'week': {
      const day = now.getDay() || 7
      from = new Date(now)
      from.setDate(now.getDate() - day + 1)
      from.setHours(0, 0, 0, 0)
      break
    }
    case 'month':
      from = new Date(now.getFullYear(), now.getMonth(), 1)
      break
    case 'quarter': {
      const qMonth = Math.floor(now.getMonth() / 3) * 3
      from = new Date(now.getFullYear(), qMonth, 1)
      break
    }
    case 'year':
      from = new Date(now.getFullYear(), 0, 1)
      break
    default:
      from = new Date(now.getFullYear(), 0, 1)
  }
  return { from, to }
}

function toDateString(d: Date): string {
  return d.toISOString().split('T')[0]
}

/** Check if a date string or Date falls within from..to range */
function isInRange(dateValue: string | undefined, from: Date, to: Date): boolean {
  if (!dateValue) return false
  const d = new Date(dateValue)
  if (isNaN(d.getTime())) return false
  return d >= from && d <= to
}

function extractArray(data: any): any[] {
  return data?.data || data?.items || (Array.isArray(data) ? data : [])
}

function ReportsPage() {
  const [period, setPeriod] = useState<ReportPeriod>('year')
  const [customFrom, setCustomFrom] = useState('')
  const [customTo, setCustomTo] = useState('')

  // Effective date range: custom overrides period preset
  const dateRange = useMemo(() => {
    if (customFrom && customTo) {
      return {
        from: new Date(customFrom + 'T00:00:00'),
        to: new Date(customTo + 'T23:59:59.999'),
      }
    }
    return getDateRange(period)
  }, [period, customFrom, customTo])

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
    queryFn: () => invoiceApi.list(1, 500),
    staleTime: 1000 * 60 * 5,
  })

  const isLoading = equipLoading || projLoading

  // Filter invoices by date range
  const filteredInvoices = useMemo(() => {
    const invoices = extractArray(invoicesData)
    return invoices.filter((inv: any) => {
      const dateField = inv.issue_date || inv.created_at || inv.paid_date
      return isInRange(dateField, dateRange.from, dateRange.to)
    })
  }, [invoicesData, dateRange])

  // Filter projects by date range
  const filteredProjects = useMemo(() => {
    const projects = extractArray(projectsData)
    return projects.filter((p: any) => {
      const dateField = p.start_date || p.created_at
      return isInRange(dateField, dateRange.from, dateRange.to)
    })
  }, [projectsData, dateRange])

  // Compute real stats from filtered data
  const stats = useMemo(() => {
    const equipment = extractArray(equipmentData)

    const totalEquipment = equipment.length
    const checkedOut = equipment.filter((e: any) => e.status === 'checked_out' || e.status === 'reserved').length
    const inMaintenance = equipment.filter((e: any) => e.status === 'maintenance' || e.status === 'in_maintenance').length
    const available = totalEquipment - checkedOut - inMaintenance
    const utilization = totalEquipment > 0 ? Math.round((checkedOut / totalEquipment) * 100) : 0

    const activeProjects = filteredProjects.filter((p: any) =>
      p.status === 'active' || p.status === 'in_progress' || p.status === 'confirmed'
    ).length

    const openInvoices = filteredInvoices.filter((i: any) =>
      i.status === 'sent' || i.status === 'overdue' || i.status === 'draft' || i.status === 'partial'
    )
    const openInvoiceCount = openInvoices.length
    const openInvoiceAmount = openInvoices.reduce((sum: number, i: any) => sum + (i.total || 0), 0)

    const totalRevenue = filteredInvoices
      .filter((i: any) => i.status === 'paid')
      .reduce((sum: number, i: any) => sum + (i.total || 0), 0)

    const maintenanceStats = maintenanceData?.stats
    const maintenanceCompliance = maintenanceStats
      ? Math.round(((maintenanceStats.total_plans - (maintenanceStats.tasks_overdue || 0)) / Math.max(maintenanceStats.total_plans, 1)) * 100)
      : 0

    const overdueCount = filteredInvoices.filter((i: any) => i.status === 'overdue').length

    return {
      total_revenue: totalRevenue,
      equipment_utilization: utilization,
      active_projects: activeProjects,
      open_invoice_count: openInvoiceCount,
      open_invoice_amount: openInvoiceAmount,
      maintenance_compliance: maintenanceCompliance,
      total_equipment: totalEquipment,
      available_equipment: available,
      in_maintenance: inMaintenance,
      checked_out: checkedOut,
      overdue_count: overdueCount,
    }
  }, [equipmentData, filteredInvoices, filteredProjects, maintenanceData])

  // Revenue by month chart data (from filtered invoices)
  const monthlyRevenueData = useMemo(() => {
    const monthNames = ['Jan', 'Feb', 'Maer', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez']

    const monthMap: Record<string, number> = {}

    // Use all invoices with totals (paid + sent + overdue = recognized revenue)
    filteredInvoices.forEach((inv: any) => {
      const date = new Date(inv.issue_date || inv.created_at || inv.paid_date)
      if (!isNaN(date.getTime())) {
        const key = monthNames[date.getMonth()]
        monthMap[key] = (monthMap[key] || 0) + (inv.total || 0)
      }
    })

    return monthNames
      .map(month => ({ month, revenue: monthMap[month] || 0 }))
      .filter(d => d.revenue > 0)
  }, [filteredInvoices])

  // Equipment utilization breakdown (pie/donut showing status distribution)
  const utilizationData = useMemo(() => {
    const equipment = extractArray(equipmentData)
    const statusMap: Record<string, number> = {}

    equipment.forEach((e: any) => {
      const status = e.status || 'unbekannt'
      const label =
        status === 'available' ? 'Verfuegbar' :
        status === 'checked_out' ? 'Vermietet' :
        status === 'reserved' ? 'Reserviert' :
        status === 'maintenance' || status === 'in_maintenance' ? 'Wartung' :
        status === 'retired' ? 'Ausgemustert' :
        status
      statusMap[label] = (statusMap[label] || 0) + 1
    })

    return Object.entries(statusMap).map(([name, value]) => ({ name, value }))
  }, [equipmentData])

  // Revenue by category (donut chart)
  const categoryRevenueData = useMemo(() => {
    const equipment = extractArray(equipmentData)
    const catMap: Record<string, number> = {}

    equipment.forEach((e: any) => {
      const cat = e.category_id || 'sonstige'
      const label = CATEGORY_LABELS[cat] || cat.replace('cat-', '').charAt(0).toUpperCase() + cat.replace('cat-', '').slice(1)
      const dailyRevenue = e.rental_price_day || 0
      catMap[label] = (catMap[label] || 0) + dailyRevenue
    })

    return Object.entries(catMap).map(([name, value]) => ({ name, value }))
  }, [equipmentData])

  // --- CSV Export: build from currently displayed data ---
  const handleExportCSV = useCallback(() => {
    const rows: string[][] = []

    // Header
    rows.push(['Bereich', 'Kennzahl', 'Wert'])

    // KPIs
    rows.push(['KPI', 'Gesamtumsatz (EUR)', stats.total_revenue.toFixed(2)])
    rows.push(['KPI', 'Equipment-Auslastung (%)', stats.equipment_utilization.toString()])
    rows.push(['KPI', 'Offene Rechnungen', stats.open_invoice_count.toString()])
    rows.push(['KPI', 'Offene Rechnungen Betrag (EUR)', stats.open_invoice_amount.toFixed(2)])
    rows.push(['KPI', 'Aktive Projekte', stats.active_projects.toString()])
    rows.push(['KPI', 'Wartungs-Compliance (%)', stats.maintenance_compliance.toString()])
    rows.push(['KPI', 'Gesamt-Equipment', stats.total_equipment.toString()])
    rows.push(['KPI', 'Ueberfaellige Rechnungen', stats.overdue_count.toString()])
    rows.push([])

    // Monthly revenue
    rows.push(['Monat', 'Umsatz (EUR)'])
    monthlyRevenueData.forEach(d => {
      rows.push([d.month, d.revenue.toFixed(2)])
    })
    rows.push([])

    // Equipment status
    rows.push(['Equipment-Status', 'Anzahl'])
    utilizationData.forEach(d => {
      rows.push([d.name, d.value.toString()])
    })
    rows.push([])

    // Category revenue
    rows.push(['Kategorie', 'Tagesumsatz (EUR)'])
    categoryRevenueData.forEach(d => {
      rows.push([d.name, d.value.toFixed(2)])
    })
    rows.push([])

    // Invoices detail
    rows.push(['Rechnungsnr.', 'Status', 'Datum', 'Betrag (EUR)'])
    filteredInvoices.forEach((inv: any) => {
      rows.push([
        inv.number || inv.invoice_number || inv.id || '',
        inv.status || '',
        inv.issue_date || inv.created_at || '',
        (inv.total || 0).toFixed(2),
      ])
    })

    // Build CSV string with BOM for Excel compatibility
    const csvContent = '\uFEFF' + rows.map(row =>
      row.map(cell => {
        const str = String(cell ?? '')
        return str.includes(',') || str.includes('"') || str.includes('\n')
          ? `"${str.replace(/"/g, '""')}"`
          : str
      }).join(';')
    ).join('\r\n')

    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `report-${new Date().toISOString().split('T')[0]}.csv`
    document.body.appendChild(a)
    a.click()
    window.URL.revokeObjectURL(url)
    document.body.removeChild(a)
  }, [stats, monthlyRevenueData, utilizationData, categoryRevenueData, filteredInvoices])

  // --- PDF Export: use window.print() ---
  const handleExportPDF = useCallback(() => {
    window.print()
  }, [])

  const getPeriodLabel = (p: ReportPeriod): string => {
    const labels: Record<ReportPeriod, string> = {
      week: 'Diese Woche',
      month: 'Dieser Monat',
      quarter: 'Dieses Quartal',
      year: 'Dieses Jahr',
    }
    return labels[p]
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

  const STATUS_COLORS: Record<string, string> = {
    'Verfuegbar': '#10b981',
    'Vermietet': '#00d4ff',
    'Reserviert': '#f59e0b',
    'Wartung': '#ef4444',
    'Ausgemustert': '#6b7280',
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
    <div className="reports-page reports-print-area">
      {/* Header */}
      <div className="page-header">
        <div>
          <h1 className="page-title">Reports & Analysen</h1>
          <p className="page-subtitle">
            Geschaeftskennzahlen {toDateString(dateRange.from)} bis {toDateString(dateRange.to)}
          </p>
        </div>
        <div className="page-header__actions" style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button
            className="btn btn--secondary no-print"
            onClick={handleExportPDF}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            PDF Export
          </button>
          <button
            className="btn btn--secondary no-print"
            onClick={handleExportCSV}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            CSV Export
          </button>
        </div>
      </div>

      {/* Period Selector + Date Range */}
      <div className="period-selector no-print">
        <label className="period-label">Zeitraum:</label>
        <div className="period-buttons">
          {(['week', 'month', 'quarter', 'year'] as ReportPeriod[]).map(p => (
            <button
              key={p}
              className={`period-button ${period === p && !customFrom && !customTo ? 'period-button--active' : ''}`}
              onClick={() => { setPeriod(p); setCustomFrom(''); setCustomTo('') }}
            >
              {getPeriodLabel(p)}
            </button>
          ))}
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', alignItems: 'center', marginLeft: 'auto' }}>
          <label style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>Von:</label>
          <input
            type="date"
            value={customFrom}
            onChange={e => setCustomFrom(e.target.value)}
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
            value={customTo}
            onChange={e => setCustomTo(e.target.value)}
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
          {(customFrom || customTo) && (
            <button
              className="btn btn--secondary"
              style={{ padding: 'var(--spacing-2) var(--spacing-3)', fontSize: 'var(--font-size-xs)' }}
              onClick={() => { setCustomFrom(''); setCustomTo('') }}
            >
              Zuruecksetzen
            </button>
          )}
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
            {monthlyRevenueData.length > 0 ? (
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
            ) : (
              <div style={{ textAlign: 'center', padding: 'var(--spacing-8)', color: 'var(--color-text-secondary)' }}>
                Keine Umsatzdaten im gewaehlten Zeitraum
              </div>
            )}
          </div>
        </div>

        {/* Equipment Utilization Donut Chart */}
        <div className="chart-section">
          <h2 className="section-title">Equipment-Auslastung</h2>
          <div className="chart-container">
            {utilizationData.length > 0 ? (
              <ResponsiveContainer width="100%" height={300}>
                <PieChart>
                  <Pie
                    data={utilizationData}
                    cx="50%"
                    cy="50%"
                    innerRadius={70}
                    outerRadius={110}
                    paddingAngle={4}
                    dataKey="value"
                    label={({ name, value }) => `${name} (${value})`}
                  >
                    {utilizationData.map((entry, index) => (
                      <Cell
                        key={`cell-util-${index}`}
                        fill={STATUS_COLORS[entry.name] || DONUT_COLORS[index % DONUT_COLORS.length]}
                      />
                    ))}
                  </Pie>
                  <Tooltip
                    formatter={(value: any, _name: any, props: any) => [
                      `${value} Geraete (${stats.total_equipment > 0 ? Math.round((value as number) / stats.total_equipment * 100) : 0}%)`,
                      props.payload.name,
                    ]}
                    contentStyle={{
                      background: 'rgba(15, 23, 42, 0.95)',
                      border: '1px solid rgba(255,255,255,0.1)',
                      borderRadius: '8px',
                      color: '#fff',
                    }}
                  />
                  <Legend wrapperStyle={{ color: 'rgba(255,255,255,0.7)', fontSize: '12px' }} />
                </PieChart>
              </ResponsiveContainer>
            ) : (
              <div style={{ textAlign: 'center', padding: 'var(--spacing-8)', color: 'var(--color-text-secondary)' }}>
                Keine Equipment-Daten verfuegbar
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Revenue by Category Donut + Additional KPIs */}
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
              <div className="kpi-card__label">Vermietet / Gesamt</div>
              <div className="kpi-card__value" style={{ color: '#00d4ff' }}>
                {stats.checked_out} / {stats.total_equipment}
              </div>
              <div className="kpi-card__target">Equipment aktuell vermietet</div>
            </div>

            <div className="kpi-card">
              <div className="kpi-card__label">Gesamt-Equipment</div>
              <div className="kpi-card__value" style={{ color: '#00d4ff' }}>
                {stats.total_equipment}
              </div>
              <div className="kpi-card__target">Geraete im Bestand</div>
            </div>

            <div className="kpi-card">
              <div className="kpi-card__label">Ueberfaellige Rechnungen</div>
              <div className="kpi-card__value" style={{
                color: stats.overdue_count > 0 ? '#ef4444' : '#10b981'
              }}>
                {stats.overdue_count}
              </div>
              <div className="kpi-card__target">Sofortige Aufmerksamkeit</div>
            </div>
          </div>
        </div>
      </div>

      {/* Invoice Detail Table */}
      <div className="recent-runs-section">
        <h2 className="section-title">Rechnungen im Zeitraum ({filteredInvoices.length})</h2>
        <div className="runs-table">
          <div className="table-header">
            <div className="table-header__cell table-header__cell--name">Rechnungsnr.</div>
            <div className="table-header__cell table-header__cell--period">Status</div>
            <div className="table-header__cell table-header__cell--date">Datum</div>
            <div className="table-header__cell table-header__cell--user">Betrag</div>
            <div className="table-header__cell table-header__cell--action"></div>
          </div>
          <div className="table-body">
            {filteredInvoices.length === 0 && (
              <div className="table-row" style={{ justifyContent: 'center', textAlign: 'center', color: 'var(--color-text-secondary)' }}>
                <div className="table-cell" style={{ gridColumn: '1 / -1', justifyContent: 'center' }}>
                  Keine Rechnungen im gewaehlten Zeitraum
                </div>
              </div>
            )}
            {filteredInvoices.slice(0, 20).map((inv: any) => (
              <div key={inv.id} className="table-row">
                <div className="table-cell table-cell--name">
                  {inv.number || inv.invoice_number || inv.id?.slice(0, 8) || '--'}
                </div>
                <div className="table-cell table-cell--period">
                  <span
                    className="period-badge"
                    style={{
                      color:
                        inv.status === 'paid' ? '#10b981' :
                        inv.status === 'overdue' ? '#ef4444' :
                        inv.status === 'sent' ? '#f59e0b' :
                        'var(--color-text-secondary)',
                    }}
                  >
                    {inv.status === 'paid' ? 'Bezahlt' :
                     inv.status === 'overdue' ? 'Ueberfaellig' :
                     inv.status === 'sent' ? 'Gesendet' :
                     inv.status === 'draft' ? 'Entwurf' :
                     inv.status === 'partial' ? 'Teilzahlung' :
                     inv.status || '--'}
                  </span>
                </div>
                <div className="table-cell table-cell--date">
                  {inv.issue_date ? new Date(inv.issue_date).toLocaleDateString('de-DE') :
                   inv.created_at ? new Date(inv.created_at).toLocaleDateString('de-DE') : '--'}
                </div>
                <div className="table-cell table-cell--user">
                  {'\u20AC'}{(inv.total || 0).toLocaleString('de-DE', { minimumFractionDigits: 2 })}
                </div>
                <div className="table-cell table-cell--action">
                  {/* Placeholder for future actions */}
                </div>
              </div>
            ))}
            {filteredInvoices.length > 20 && (
              <div className="table-row" style={{ justifyContent: 'center', color: 'var(--color-text-secondary)' }}>
                <div className="table-cell" style={{ gridColumn: '1 / -1', justifyContent: 'center', fontSize: 'var(--font-size-xs)' }}>
                  ... und {filteredInvoices.length - 20} weitere Rechnungen
                </div>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default ReportsPage
