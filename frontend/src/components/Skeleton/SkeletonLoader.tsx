import './SkeletonLoader.scss'

/* ─── Primitive ─── */

interface SkeletonBaseProps {
  width?: string | number
  height?: string | number
  borderRadius?: string
  className?: string
  style?: React.CSSProperties
}

function toCSS(v: string | number | undefined, fallback: string): string {
  if (v === undefined) return fallback
  return typeof v === 'number' ? `${v}px` : v
}

export function SkeletonBlock({ width, height, borderRadius, className = '', style }: SkeletonBaseProps) {
  return (
    <div
      className={`skel-shimmer ${className}`}
      style={{
        width: toCSS(width, '100%'),
        height: toCSS(height, '20px'),
        borderRadius: borderRadius ?? 'var(--radius-md)',
        ...style,
      }}
    />
  )
}

/* ─── SkeletonText ─── */

interface SkeletonTextProps {
  lines?: number
  width?: string
  gap?: number
}

export function SkeletonText({ lines = 3, width = '100%', gap = 10 }: SkeletonTextProps) {
  return (
    <div className="skel-text" style={{ gap }}>
      {Array.from({ length: lines }).map((_, i) => (
        <SkeletonBlock
          key={i}
          width={i === lines - 1 ? '60%' : width}
          height={14}
          borderRadius="var(--radius-base)"
        />
      ))}
    </div>
  )
}

/* ─── SkeletonAvatar ─── */

interface SkeletonAvatarProps {
  size?: number
}

export function SkeletonAvatar({ size = 40 }: SkeletonAvatarProps) {
  return <SkeletonBlock width={size} height={size} borderRadius="var(--radius-full)" />
}

/* ─── SkeletonCard ─── */

interface SkeletonCardProps {
  count?: number
}

export function SkeletonCard({ count = 1 }: SkeletonCardProps) {
  return (
    <div className="skel-card-grid">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="skel-card">
          <SkeletonBlock height={18} width="55%" />
          <SkeletonBlock height={32} width="40%" style={{ marginTop: 12 }} />
          <SkeletonBlock height={12} width="70%" style={{ marginTop: 16 }} />
        </div>
      ))}
    </div>
  )
}

/* ─── SkeletonTable ─── */

interface SkeletonTableProps {
  rows?: number
  columns?: number
}

export function SkeletonTable({ rows = 5, columns = 5 }: SkeletonTableProps) {
  return (
    <div className="skel-table">
      {/* header */}
      <div className="skel-table__header">
        {Array.from({ length: columns }).map((_, c) => (
          <SkeletonBlock key={c} height={14} width="80%" />
        ))}
      </div>
      {/* rows */}
      {Array.from({ length: rows }).map((_, r) => (
        <div key={r} className="skel-table__row">
          {Array.from({ length: columns }).map((_, c) => (
            <SkeletonBlock key={c} height={14} width={`${55 + Math.round(Math.random() * 35)}%`} />
          ))}
        </div>
      ))}
    </div>
  )
}

/* ─── SkeletonKPI ─── */

interface SkeletonKPIProps {
  count?: number
}

export function SkeletonKPI({ count = 4 }: SkeletonKPIProps) {
  return (
    <div className="skel-kpi-grid">
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="skel-kpi">
          <SkeletonBlock height={14} width="50%" />
          <SkeletonBlock height={28} width="35%" style={{ marginTop: 10 }} />
        </div>
      ))}
    </div>
  )
}
