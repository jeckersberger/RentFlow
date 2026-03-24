import { useQuery } from '@tanstack/react-query'
import { invoiceApi, mailApi, maintenanceApi, bookingApi } from '../services/api'

export function useBadgeCounts() {
  // Overdue invoices count
  const { data: invoicesOverdue = 0 } = useQuery({
    queryKey: ['badge-invoices-overdue'],
    queryFn: () =>
      invoiceApi
        .list(1, 1000)
        .then((res) => {
          const items = res?.data || res?.items || []
          if (Array.isArray(items)) {
            return items.filter((i: any) => i.status === 'overdue').length
          }
          return 0
        })
        .catch(() => 0),
    staleTime: 60_000,
    retry: false,
    refetchOnWindowFocus: false,
  })

  // Unread mail count
  const { data: mailUnread = 0 } = useQuery({
    queryKey: ['badge-mail-unread'],
    queryFn: () =>
      mailApi
        .list({ folder: 'inbox', is_read: false, limit: 1000 })
        .then((res) => {
          if (typeof res?.total === 'number') return res.total
          if (Array.isArray(res)) return res.filter((m: any) => !m.is_read).length
          if (Array.isArray(res?.data)) return res.data.filter((m: any) => !m.is_read).length
          return 0
        })
        .catch(() => 0),
    staleTime: 60_000,
    retry: false,
    refetchOnWindowFocus: false,
  })

  // Pending crew bookings count
  const { data: crewPending = 0 } = useQuery({
    queryKey: ['badge-crew-pending'],
    queryFn: () =>
      bookingApi
        .list()
        .then((res) => {
          const items = Array.isArray(res) ? res : res?.data || []
          if (Array.isArray(items)) {
            return items.filter((b: any) => b.status === 'pending').length
          }
          return 0
        })
        .catch(() => 0),
    staleTime: 60_000,
    retry: false,
    refetchOnWindowFocus: false,
  })

  // Overdue maintenance tasks count
  const { data: maintenanceOverdue = 0 } = useQuery({
    queryKey: ['badge-maintenance-overdue'],
    queryFn: () =>
      maintenanceApi
        .getOverdueTasks()
        .then((res) => {
          if (Array.isArray(res)) return res.length
          if (Array.isArray(res?.items)) return res.items.length
          if (typeof res?.total === 'number') return res.total
          return 0
        })
        .catch(() => 0),
    staleTime: 60_000,
    retry: false,
    refetchOnWindowFocus: false,
  })

  return {
    invoicesOverdue,
    mailUnread,
    crewPending,
    maintenanceOverdue,
  }
}
