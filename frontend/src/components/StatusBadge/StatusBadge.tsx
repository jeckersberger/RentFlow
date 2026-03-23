import './StatusBadge.module.scss'

interface StatusBadgeProps {
  status: string
  label?: string
  size?: 'sm' | 'md' | 'lg'
}

const STATUS_LABELS: Record<string, string> = {
  available: 'Verfügbar',
  reserved: 'Reserviert',
  checked_out: 'Vermietet',
  in_maintenance: 'In Wartung',
  maintenance: 'In Wartung',
  damaged: 'Beschädigt',
  retired: 'Ausgemustert',
  confirmed: 'Bestätigt',
  paid: 'Bezahlt',
  draft: 'Entwurf',
  quoted: 'Angebot',
  in_progress: 'In Bearbeitung',
  completed: 'Abgeschlossen',
  invoiced: 'Abgerechnet',
  sent: 'Gesendet',
  overdue: 'Überfällig',
  cancelled: 'Storniert',
  partial: 'Teilweise bezahlt',
  active: 'Aktiv',
  planning: 'In Planung',
  // Transport statuses
  in_use: 'Im Einsatz',
  planned: 'Geplant',
  loading: 'Beladung',
  in_transit: 'Unterwegs',
  delivered: 'Geliefert',
  // Crew statuses
  busy: 'Beschäftigt',
  on_leave: 'Urlaub',
  sick: 'Krank',
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

  const color = getStatusColor(status)
  const displayLabel = label || STATUS_LABELS[status] || status.replace(/_/g, ' ')

  return (
    <span className={`status-badge status-badge--${color} status-badge--${size}`}>
      {displayLabel}
    </span>
  )
}
