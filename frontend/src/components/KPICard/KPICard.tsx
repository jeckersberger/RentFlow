import './KPICard.module.scss'

interface KPICardProps {
  title: string
  value: string | number
  icon?: string
  change?: {
    value: number
    isPositive: boolean
  }
  isLoading?: boolean
}

export function KPICard({ title, value, icon, change, isLoading }: KPICardProps) {
  return (
    <div className="kpi-card">
      {icon && <div className="kpi-card__icon">{icon}</div>}
      <div className="kpi-card__content">
        <p className="kpi-card__title">{title}</p>
        <p className="kpi-card__value">
          {isLoading ? '—' : value}
        </p>
        {change && (
          <p className={`kpi-card__change kpi-card__change--${change.isPositive ? 'positive' : 'negative'}`}>
            {change.isPositive ? '↑' : '↓'} {Math.abs(change.value)}%
          </p>
        )}
      </div>
    </div>
  )
}
