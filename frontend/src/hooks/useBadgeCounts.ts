import { useQuery, useQueryClient } from '@tanstack/react-query'
import {
  projectApi,
  equipmentApi,
  invoiceApi,
  mailApi,
  contactApi,
  crewApi,
  bookingApi,
  maintenanceApi,
  auditApi,
} from '../services/api'
import { useCallback } from 'react'

export interface BadgeCounts {
  projects: number
  equipment: number
  invoices: number
  invoicesOverdue: number
  contacts: number
  crew: number
  mailUnread: number
  maintenance: number
  maintenanceOverdue: number
  audit: number
}

const EMPTY_COUNTS: BadgeCounts = {
  projects: 0,
  equipment: 0,
  invoices: 0,
  invoicesOverdue: 0,
  contacts: 0,
  crew: 0,
  mailUnread: 0,
  maintenance: 0,
  maintenanceOverdue: 0,
  audit: 0,
}

function getLastSeen(section: string): string {
  return localStorage.getItem(`rentflow_lastSeen_${section}`) || '1970-01-01T00:00:00Z'
}

export function useMarkSectionSeen() {
  const queryClient = useQueryClient()
  return useCallback(
    (section: string) => {
      localStorage.setItem(`rentflow_lastSeen_${section}`, new Date().toISOString())
      queryClient.invalidateQueries({ queryKey: ['badge-counts'] })
    },
    [queryClient]
  )
}

/**
 * Count how many items in an array were created after a given ISO timestamp.
 * Looks for `created_at` or `createdAt` fields.
 */
function countNewSince(items: unknown[], lastSeen: string): number {
  if (!Array.isArray(items)) return 0
  const since = new Date(lastSeen).getTime()
  return items.filter((item: any) => {
    const created = new Date(item.created_at || item.createdAt || '').getTime()
    return !isNaN(created) && created > since
  }).length
}

/** Safely extract an array from various API response shapes */
function toArray(res: any): any[] {
  if (Array.isArray(res)) return res
  if (Array.isArray(res?.data)) return res.data
  if (Array.isArray(res?.items)) return res.items
  return []
}

export function useBadgeCounts(): BadgeCounts {
  const { data } = useQuery({
    queryKey: ['badge-counts'],
    queryFn: async () => {
      const lastSeenProjects = getLastSeen('projects')
      const lastSeenEquipment = getLastSeen('equipment')
      const lastSeenInvoices = getLastSeen('invoices')
      const lastSeenContacts = getLastSeen('contacts')
      const lastSeenCrew = getLastSeen('crew')
      const lastSeenMaintenance = getLastSeen('maintenance')
      const lastSeenAudit = getLastSeen('audit')

      // Fire all API requests in parallel
      const [
        projectsRes,
        equipmentRes,
        invoicesRes,
        contactsRes,
        crewRes,
        bookingsRes,
        tasksRes,
        overdueTasksRes,
        mailRes,
        auditRes,
      ] = await Promise.allSettled([
        projectApi.list(1, 1000).catch(() => []),
        equipmentApi.list({ limit: 1000 }).catch(() => []),
        invoiceApi.list(1, 1000).catch(() => []),
        contactApi.list({ limit: 1000 }).catch(() => []),
        crewApi.listMembers({ per_page: 1000 }).catch(() => []),
        bookingApi.list().catch(() => []),
        maintenanceApi.listTasks(1, 1000).catch(() => []),
        maintenanceApi.getOverdueTasks().catch(() => []),
        mailApi.list({ folder: 'inbox', is_read: false, limit: 1000 }).catch(() => []),
        auditApi.logs({ limit: 1000 }).catch(() => []),
      ])

      const projectsArr = projectsRes.status === 'fulfilled' ? toArray(projectsRes.value) : []
      const equipmentArr = equipmentRes.status === 'fulfilled' ? toArray(equipmentRes.value) : []
      const invoicesArr = invoicesRes.status === 'fulfilled' ? toArray(invoicesRes.value) : []
      const contactsArr = contactsRes.status === 'fulfilled' ? toArray(contactsRes.value) : []
      const crewArr = crewRes.status === 'fulfilled' ? toArray(crewRes.value) : []
      const bookingsArr = bookingsRes.status === 'fulfilled' ? toArray(bookingsRes.value) : []
      const tasksArr = tasksRes.status === 'fulfilled' ? toArray(tasksRes.value) : []
      const overdueTasksArr = overdueTasksRes.status === 'fulfilled' ? toArray(overdueTasksRes.value) : []
      const auditArr = auditRes.status === 'fulfilled' ? toArray(auditRes.value) : []

      // Overdue invoices (always shown, regardless of lastSeen)
      const invoicesOverdue = invoicesArr.filter(
        (i: any) => i.status === 'overdue' || i.status === 'überfällig'
      ).length

      // Pending bookings count toward crew badge
      const pendingBookings = bookingsArr.filter((b: any) => b.status === 'pending').length

      // Overdue maintenance (always shown)
      const maintenanceOverdue = overdueTasksArr.length

      // Unread mail count
      let mailUnread = 0
      if (mailRes.status === 'fulfilled') {
        const raw = mailRes.value
        if (typeof raw?.total === 'number') {
          mailUnread = raw.total
        } else {
          const arr = toArray(raw)
          mailUnread = arr.filter((m: any) => !m.is_read).length
        }
      }

      return {
        projects: countNewSince(projectsArr, lastSeenProjects),
        equipment: countNewSince(equipmentArr, lastSeenEquipment),
        invoices: countNewSince(invoicesArr, lastSeenInvoices),
        invoicesOverdue,
        contacts: countNewSince(contactsArr, lastSeenContacts),
        crew: countNewSince(crewArr, lastSeenCrew) + pendingBookings,
        mailUnread,
        maintenance: countNewSince(tasksArr, lastSeenMaintenance),
        maintenanceOverdue,
        audit: countNewSince(auditArr, lastSeenAudit),
      }
    },
    staleTime: 60_000,
    retry: false,
    refetchOnWindowFocus: false,
  })

  return data || EMPTY_COUNTS
}
