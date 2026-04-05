import { Equipment } from '../../types/equipment'
import { Project } from '../../types/project'

// ============================================================================
// Types
// ============================================================================

export type ViewMode = 'week' | 'month'

export interface Allocation {
  id: string
  equipmentId: string
  projectId: string
  projectName: string
  clientName: string
  startDate: Date
  endDate: Date
  color: string
}

export interface Conflict {
  equipmentId: string
  allocations: Allocation[]
}

export interface TooltipData {
  allocation: Allocation
  conflict?: Allocation[]
  x: number
  y: number
}

// ============================================================================
// Constants
// ============================================================================

export const DAY_NAMES_DE = ['So', 'Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa']
export const MONTH_NAMES_DE = [
  'Januar', 'Februar', 'Maerz', 'April', 'Mai', 'Juni',
  'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember',
]
export const MONTH_SHORT_DE = [
  'Jan', 'Feb', 'Maer', 'Apr', 'Mai', 'Jun',
  'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez',
]

const PROJECT_COLORS = [
  '#3b82f6', '#8b5cf6', '#10b981', '#f59e0b',
  '#ef4444', '#06b6d4', '#ec4899', '#14b8a6',
  '#f97316', '#6366f1', '#84cc16', '#a855f7',
]

// ============================================================================
// Date helpers
// ============================================================================

export function startOfDay(d: Date): Date {
  return new Date(d.getFullYear(), d.getMonth(), d.getDate())
}

export function addDays(d: Date, n: number): Date {
  const r = new Date(d)
  r.setDate(r.getDate() + n)
  return r
}

export function isSameDay(a: Date, b: Date): boolean {
  return a.getFullYear() === b.getFullYear() && a.getMonth() === b.getMonth() && a.getDate() === b.getDate()
}

export function isWeekend(d: Date): boolean {
  const day = d.getDay()
  return day === 0 || day === 6
}

export function getMonday(d: Date): Date {
  const result = new Date(d)
  const day = result.getDay()
  const diff = day === 0 ? -6 : 1 - day
  result.setDate(result.getDate() + diff)
  return startOfDay(result)
}

export function getColorForProject(projectId: string): string {
  let hash = 0
  for (let i = 0; i < projectId.length; i++) {
    hash = ((hash << 5) - hash + projectId.charCodeAt(i)) | 0
  }
  return PROJECT_COLORS[Math.abs(hash) % PROJECT_COLORS.length]
}

// ============================================================================
// Build allocations from projects + equipment
// ============================================================================

export function buildAllocations(equipment: Equipment[], projects: Project[]): Allocation[] {
  const allocations: Allocation[] = []
  const activeProjects = projects.filter(p =>
    p.status !== 'cancelled' && p.start_date && p.end_date
  )

  // Simulate equipment-to-project assignments deterministically
  // In production, this would come from the reservations API
  activeProjects.forEach((project, pIdx) => {
    const color = getColorForProject(project.id)
    const eqStart = (pIdx * 3) % equipment.length
    const eqCount = 2 + (pIdx % 4)

    for (let i = 0; i < eqCount && i < equipment.length; i++) {
      const eq = equipment[(eqStart + i) % equipment.length]
      const start = project.setup_date
        ? new Date(project.setup_date)
        : addDays(new Date(project.start_date), -1)
      const end = project.teardown_date
        ? new Date(project.teardown_date)
        : addDays(new Date(project.end_date), 1)

      allocations.push({
        id: `alloc-${project.id}-${eq.id}`,
        equipmentId: eq.id,
        projectId: project.id,
        projectName: project.name,
        clientName: project.client_name || '',
        startDate: startOfDay(start),
        endDate: startOfDay(end),
        color,
      })
    }
  })

  return allocations
}

// ============================================================================
// Detect conflicts (overlapping allocations on same equipment)
// ============================================================================

export function detectConflicts(allocations: Allocation[]): Map<string, Conflict> {
  const byEquipment = new Map<string, Allocation[]>()
  allocations.forEach(a => {
    if (!byEquipment.has(a.equipmentId)) byEquipment.set(a.equipmentId, [])
    byEquipment.get(a.equipmentId)!.push(a)
  })

  const conflicts = new Map<string, Conflict>()

  byEquipment.forEach((eqAllocations, eqId) => {
    for (let i = 0; i < eqAllocations.length; i++) {
      for (let j = i + 1; j < eqAllocations.length; j++) {
        const a = eqAllocations[i]
        const b = eqAllocations[j]
        if (a.startDate < b.endDate && b.startDate < a.endDate) {
          if (!conflicts.has(eqId)) {
            conflicts.set(eqId, { equipmentId: eqId, allocations: [] })
          }
          const c = conflicts.get(eqId)!
          if (!c.allocations.find(x => x.id === a.id)) c.allocations.push(a)
          if (!c.allocations.find(x => x.id === b.id)) c.allocations.push(b)
        }
      }
    }
  })

  return conflicts
}
