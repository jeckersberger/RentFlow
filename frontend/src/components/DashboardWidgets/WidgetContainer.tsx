import { useState, useRef, useEffect } from 'react'
import './DashboardWidgets.scss'

interface WidgetContainerProps {
  title: string
  icon?: string
  collapsed: boolean
  onToggleCollapse: () => void
  onRemove: () => void
  onMoveUp: () => void
  onMoveDown: () => void
  isFirst: boolean
  isLast: boolean
  children: React.ReactNode
}

export function WidgetContainer({
  title,
  icon,
  collapsed,
  onToggleCollapse,
  onRemove,
  onMoveUp,
  onMoveDown,
  isFirst,
  isLast,
  children,
}: WidgetContainerProps) {
  const [menuOpen, setMenuOpen] = useState(false)
  const menuRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setMenuOpen(false)
      }
    }
    if (menuOpen) {
      document.addEventListener('mousedown', handleClickOutside)
    }
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [menuOpen])

  return (
    <div className={`widget-container ${collapsed ? 'widget-container--collapsed' : ''}`}>
      <div className="widget-container__header">
        <div className="widget-container__drag-handle" title="Verschieben">
          <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
            <circle cx="5" cy="3" r="1.5" />
            <circle cx="11" cy="3" r="1.5" />
            <circle cx="5" cy="8" r="1.5" />
            <circle cx="11" cy="8" r="1.5" />
            <circle cx="5" cy="13" r="1.5" />
            <circle cx="11" cy="13" r="1.5" />
          </svg>
        </div>

        <div className="widget-container__title-row">
          {icon && <span className="widget-container__icon">{icon}</span>}
          <h3 className="widget-container__title">{title}</h3>
        </div>

        <div className="widget-container__actions">
          <button
            className="widget-container__action-btn"
            onClick={onToggleCollapse}
            title={collapsed ? 'Aufklappen' : 'Einklappen'}
          >
            <svg
              width="16"
              height="16"
              viewBox="0 0 16 16"
              fill="currentColor"
              style={{ transform: collapsed ? 'rotate(-90deg)' : 'rotate(0deg)', transition: 'transform 0.2s ease' }}
            >
              <path d="M4 6l4 4 4-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
            </svg>
          </button>

          <div className="widget-container__menu-wrapper" ref={menuRef}>
            <button
              className="widget-container__action-btn"
              onClick={() => setMenuOpen(!menuOpen)}
              title="Mehr Optionen"
            >
              <svg width="16" height="16" viewBox="0 0 16 16" fill="currentColor">
                <circle cx="8" cy="3" r="1.5" />
                <circle cx="8" cy="8" r="1.5" />
                <circle cx="8" cy="13" r="1.5" />
              </svg>
            </button>

            {menuOpen && (
              <div className="widget-container__menu">
                {!isFirst && (
                  <button className="widget-container__menu-item" onClick={() => { onMoveUp(); setMenuOpen(false) }}>
                    <span>↑</span> Nach oben
                  </button>
                )}
                {!isLast && (
                  <button className="widget-container__menu-item" onClick={() => { onMoveDown(); setMenuOpen(false) }}>
                    <span>↓</span> Nach unten
                  </button>
                )}
                <button className="widget-container__menu-item widget-container__menu-item--danger" onClick={() => { onRemove(); setMenuOpen(false) }}>
                  <span>✕</span> Entfernen
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      {!collapsed && (
        <div className="widget-container__body">
          {children}
        </div>
      )}
    </div>
  )
}
