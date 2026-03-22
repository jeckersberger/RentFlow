import './Skeleton.scss'

interface SkeletonProps {
  width?: string | number
  height?: string | number
  variant?: 'text' | 'circular' | 'rectangular'
  count?: number
}

export function Skeleton({
  width = '100%',
  height = '20px',
  variant = 'text',
  count = 1,
}: SkeletonProps) {
  const skeletons = Array.from({ length: count })

  return (
    <>
      {skeletons.map((_, index) => (
        <div
          key={index}
          className={`skeleton skeleton--${variant}`}
          style={{
            width: typeof width === 'number' ? `${width}px` : width,
            height: typeof height === 'number' ? `${height}px` : height,
          }}
        />
      ))}
    </>
  )
}
