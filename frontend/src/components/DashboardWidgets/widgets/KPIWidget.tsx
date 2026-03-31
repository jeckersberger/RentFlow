import { useQuery } from '@tanstack/react-query'
import { api, projectApi } from '../../../services/api'
import { invoiceApi } from '../../../services/api'

export function KPIWidget() {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      try {
        const [equipmentRes, projectRes] = await Promise.allSettled([
          api.get('/api/v1/equipment'),
          projectApi.list(1, 1),
        ])

        let totalEquipment = 0
        let availableEquipment = 0
        if (equipmentRes.status === 'fulfilled') {
          const equipment = equipmentRes.value.data?.data || equipmentRes.value.data?.items || equipmentRes.value.data || []
          if (Array.isArray(equipment)) {
            totalEquipment = equipment.length
            availableEquipment = equipment.filter((e: Record<string, string>) => e.status === 'available').length
          }
        }

        let activeProjects = 0
        if (projectRes.status === 'fulfilled') {
          activeProjects = projectRes.value?.total || 0
        }

        return {
          total_equipment: totalEquipment,
          available_equipment: availableEquipment,
          active_projects: activeProjects,
          pending_invoices: 0,
        }
      } catch {
        return { total_equipment: 0, available_equipment: 0, active_projects: 0, pending_invoices: 0 }
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  const { data: overdueInvoiceData } = useQuery<{ count: number; total: number }>({
    queryKey: ['dashboard-overdue-invoices'],
    queryFn: async () => {
      try {
        const res = await invoiceApi.list(1, 100)
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const invoices: any[] = res?.data || res?.items || []
        if (!Array.isArray(invoices)) return { count: 0, total: 0 }
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const overdue = invoices.filter((inv: any) => inv.status === 'overdue')
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        const total = overdue.reduce((sum: number, inv: any) => sum + (inv.total || 0), 0)
        return { count: overdue.length, total }
      } catch {
        return { count: 0, total: 0 }
      }
    },
    staleTime: 1000 * 60 * 5,
  })

  const kpis = [
    { title: 'Ausrüstung gesamt', value: stats?.total_equipment || 0, icon: '📦', trend: null },
    { title: 'Verfügbar', value: stats?.available_equipment || 0, icon: '✅', trend: null },
    { title: 'Aktive Projekte', value: stats?.active_projects || 0, icon: '📋', trend: null },
    { title: 'Überfällige Rechnungen', value: overdueInvoiceData?.count || 0, icon: '⚠️', trend: null },
  ]

  return (
    <div className="kpi-widget">
      {kpis.map((kpi, index) => (
        <div key={index} className="kpi-widget__card">
          <div className="kpi-widget__card-top">
            <span className="kpi-widget__label">{kpi.title}</span>
            <span className="kpi-widget__icon">{kpi.icon}</span>
          </div>
          <div className="kpi-widget__value">
            {isLoading ? '—' : kpi.value}
          </div>
          <div className="kpi-widget__sparkline">
            <svg viewBox="0 0 80 20" className="kpi-widget__sparkline-svg">
              <polyline
                fill="none"
                stroke="currentColor"
                strokeWidth="1.5"
                points="0,15 10,12 20,14 30,8 40,10 50,6 60,9 70,4 80,7"
              />
            </svg>
          </div>
        </div>
      ))}
    </div>
  )
}
