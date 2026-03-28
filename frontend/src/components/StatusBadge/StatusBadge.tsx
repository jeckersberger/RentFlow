import './StatusBadge.scss';

type Variant = 'success' | 'warning' | 'error' | 'info' | 'default';

interface StatusBadgeProps {
  status: string;
  variant?: Variant;
}

const autoVariant: Record<string, Variant> = {
  available: 'success',
  active: 'success',
  paid: 'success',
  reserved: 'info',
  confirmed: 'info',
  partial_paid: 'info',
  draft: 'default',
  pending: 'default',
  damaged: 'error',
  cancelled: 'error',
  overdue: 'error',
  in_maintenance: 'warning',
};

function formatLabel(status: string): string {
  return status
    .replace(/_/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase());
}

export function StatusBadge({ status, variant }: StatusBadgeProps) {
  const resolved = variant ?? autoVariant[status] ?? 'default';

  return (
    <span className={`status-badge status-badge--${resolved}`}>
      {formatLabel(status)}
    </span>
  );
}
