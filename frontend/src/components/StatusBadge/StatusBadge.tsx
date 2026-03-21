import './StatusBadge.module.scss'

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
        return 'success'
      case 'reserved':
      case 'draft':
      case 'in_progress':
        return 'warning'
      case 'checked_out':
      case 'sent':
        return 'info'
      case 'maintenance':
      case 'overdue':
        return 'danger'
      case 'retired':
      case 'cancelled':
        return 'secondary'
      default:
        return 'secondary'
    }
  }

  const color = getStatusColor(status)
  const displayLabel = label || status.replace(/_/g, ' ')

  return (
    <span className={`status-badge status-badge--${color} status-badge--${size}`}>
      {displayLabel}
    </span>
  )
}
