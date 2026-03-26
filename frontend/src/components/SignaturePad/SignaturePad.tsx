import { useRef, useEffect, useState, useCallback } from 'react'

interface SignaturePadProps {
  onSave: (dataUrl: string) => void
  onCancel?: () => void
  width?: number
  height?: number
  title?: string
}

export function SignaturePad({
  onSave,
  onCancel,
  width = 400,
  height = 200,
  title = 'Unterschrift',
}: SignaturePadProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const [isDrawing, setIsDrawing] = useState(false)
  const [hasDrawn, setHasDrawn] = useState(false)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    // High-DPI support
    const dpr = window.devicePixelRatio || 1
    canvas.width = width * dpr
    canvas.height = height * dpr
    canvas.style.width = `${width}px`
    canvas.style.height = `${height}px`
    ctx.scale(dpr, dpr)

    // White background
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, width, height)

    // Signature line
    ctx.strokeStyle = '#d1d5db'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(20, height - 30)
    ctx.lineTo(width - 20, height - 30)
    ctx.stroke()

    // "X" marker
    ctx.fillStyle = '#9ca3af'
    ctx.font = '14px Arial'
    ctx.fillText('✕', 20, height - 35)

    // Drawing settings
    ctx.strokeStyle = '#1a1a2e'
    ctx.lineWidth = 2
    ctx.lineCap = 'round'
    ctx.lineJoin = 'round'
  }, [width, height])

  const getPos = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    const canvas = canvasRef.current
    if (!canvas) return { x: 0, y: 0 }
    const rect = canvas.getBoundingClientRect()
    if ('touches' in e) {
      return {
        x: e.touches[0].clientX - rect.left,
        y: e.touches[0].clientY - rect.top,
      }
    }
    return {
      x: e.clientX - rect.left,
      y: e.clientY - rect.top,
    }
  }, [])

  const startDraw = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    e.preventDefault()
    const ctx = canvasRef.current?.getContext('2d')
    if (!ctx) return
    setIsDrawing(true)
    setHasDrawn(true)
    const pos = getPos(e)
    ctx.beginPath()
    ctx.moveTo(pos.x, pos.y)
  }, [getPos])

  const draw = useCallback((e: React.MouseEvent | React.TouchEvent) => {
    if (!isDrawing) return
    e.preventDefault()
    const ctx = canvasRef.current?.getContext('2d')
    if (!ctx) return
    const pos = getPos(e)
    ctx.lineTo(pos.x, pos.y)
    ctx.stroke()
  }, [isDrawing, getPos])

  const endDraw = useCallback(() => {
    setIsDrawing(false)
  }, [])

  const clear = () => {
    const canvas = canvasRef.current
    if (!canvas) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return
    const dpr = window.devicePixelRatio || 1
    ctx.setTransform(1, 0, 0, 1, 0, 0)
    ctx.clearRect(0, 0, canvas.width, canvas.height)
    ctx.scale(dpr, dpr)
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, width, height)
    ctx.strokeStyle = '#d1d5db'
    ctx.lineWidth = 1
    ctx.beginPath()
    ctx.moveTo(20, height - 30)
    ctx.lineTo(width - 20, height - 30)
    ctx.stroke()
    ctx.fillStyle = '#9ca3af'
    ctx.font = '14px Arial'
    ctx.fillText('✕', 20, height - 35)
    ctx.strokeStyle = '#1a1a2e'
    ctx.lineWidth = 2
    ctx.lineCap = 'round'
    ctx.lineJoin = 'round'
    setHasDrawn(false)
  }

  const save = () => {
    if (!canvasRef.current || !hasDrawn) return
    const dataUrl = canvasRef.current.toDataURL('image/png')
    onSave(dataUrl)
  }

  return (
    <div style={{
      background: 'var(--color-bg-card, #111827)',
      borderRadius: '12px',
      padding: '20px',
      border: '1px solid var(--color-border, #1e293b)',
    }}>
      <h3 style={{ margin: '0 0 12px 0', fontSize: '16px', color: 'var(--color-text, #e2e8f0)' }}>
        {title}
      </h3>
      <canvas
        ref={canvasRef}
        style={{
          border: '2px solid var(--color-border, #1e293b)',
          borderRadius: '8px',
          cursor: 'crosshair',
          touchAction: 'none',
          display: 'block',
          width: '100%',
          maxWidth: `${width}px`,
        }}
        onMouseDown={startDraw}
        onMouseMove={draw}
        onMouseUp={endDraw}
        onMouseLeave={endDraw}
        onTouchStart={startDraw}
        onTouchMove={draw}
        onTouchEnd={endDraw}
      />
      <div style={{ display: 'flex', gap: '8px', marginTop: '12px', justifyContent: 'flex-end' }}>
        <button
          onClick={clear}
          style={{
            padding: '8px 16px', borderRadius: '8px', border: '1px solid var(--color-border)',
            background: 'transparent', color: 'var(--color-text-secondary)', cursor: 'pointer',
          }}
        >
          Löschen
        </button>
        {onCancel && (
          <button
            onClick={onCancel}
            style={{
              padding: '8px 16px', borderRadius: '8px', border: '1px solid var(--color-border)',
              background: 'transparent', color: 'var(--color-text-secondary)', cursor: 'pointer',
            }}
          >
            Abbrechen
          </button>
        )}
        <button
          onClick={save}
          disabled={!hasDrawn}
          style={{
            padding: '8px 16px', borderRadius: '8px', border: 'none',
            background: hasDrawn ? '#00d4ff' : 'rgba(0,212,255,0.3)',
            color: hasDrawn ? '#0a0f1a' : 'rgba(10,15,26,0.5)',
            cursor: hasDrawn ? 'pointer' : 'not-allowed', fontWeight: 600,
          }}
        >
          Unterschrift bestätigen
        </button>
      </div>
    </div>
  )
}
