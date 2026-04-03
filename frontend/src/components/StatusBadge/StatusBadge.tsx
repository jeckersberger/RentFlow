import { getStatusLabel } from '../../utils/statusLabels'
import './StatusBadge.scss'

interface StatusBadgeProps {
  status: string
  label?: string
  size?: 'sm' | 'md' | 'lg'
}

export function StatusBadge({ status, label, size = 'md' }: StatusBadgeProps) {
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'available':
      case 'confirmed':
      case 'paid':
      case 'completed':
        return 'success'
      case 'reserved':
      case 'draft':
      case 'quoted':
      case 'partial':
        return 'warning'
      case 'in_progress':
      case 'active':
        return 'cyan'
      case 'checked_out':
      case 'rented':
      case 'sent':
      case 'planning':
      case 'invoiced':
      case 'in_use':
      case 'in_transit':
      case 'busy':
        return 'info'
      case 'maintenance':
      case 'in_maintenance':
      case 'loading':
        return 'orange'
      case 'overdue':
      case 'damaged':
      case 'sick':
      case 'lost':
        return 'danger'
      case 'retired':
      case 'cancelled':
        return 'secondary'
      case 'planned':
      case 'on_leave':
        return 'warning'
      case 'delivered':
        return 'success'
      default:
        return 'secondary'
    }
  }

  const safeStatus = status || 'draft'
  const color = getStatusColor(safeStatus)
  const displayLabel = label || getStatusLabel(safeStatus) || safeStatus.replace(/_/g, ' ')

  return (
    <span className={`status-badge status-badge--${color} status-badge--${size}`}>
      {displayLabel}
    </span>
  )
}
