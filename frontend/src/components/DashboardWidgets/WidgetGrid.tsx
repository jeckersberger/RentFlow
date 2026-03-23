import { useState, useCallback } from 'react'
import { WidgetContainer } from './WidgetContainer'
import { WidgetPicker } from './WidgetPicker'
import { WIDGET_REGISTRY, getWidgetDefinition } from './widgetRegistry'
import { WidgetInstance, STORAGE_KEY } from './types'
import './DashboardWidgets.scss'

const DEFAULT_WIDGETS: WidgetInstance[] = [
  { id: 'w-kpi', type: 'kpi', collapsed: false },
  { id: 'w-revenue-chart', type: 'revenue-chart', collapsed: false },
  { id: 'w-quick-actions', type: 'quick-actions', collapsed: false },
  { id: 'w-today-warehouse', type: 'today-warehouse', collapsed: false },
  { id: 'w-overdue-invoices', type: 'overdue-invoices', collapsed: false },
  { id: 'w-recent-activity', type: 'recent-activity', collapsed: false },
]

function loadLayout(): WidgetInstance[] {
  try {
    const saved = localStorage.getItem(STORAGE_KEY)
    if (saved) {
      const parsed = JSON.parse(saved)
      if (Array.isArray(parsed) && parsed.length > 0) {
        return parsed
      }
    }
  } catch {
    // ignore
  }
  return DEFAULT_WIDGETS
}

function saveLayout(widgets: WidgetInstance[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(widgets))
  } catch {
    // ignore
  }
}

export function WidgetGrid() {
  const [widgets, setWidgets] = useState<WidgetInstance[]>(loadLayout)
  const [pickerOpen, setPickerOpen] = useState(false)
  const [hasChanges, setHasChanges] = useState(false)

  const updateWidgets = useCallback((updater: (prev: WidgetInstance[]) => WidgetInstance[]) => {
    setWidgets(prev => {
      const next = updater(prev)
      setHasChanges(true)
      return next
    })
  }, [])

  const handleToggleCollapse = useCallback((widgetId: string) => {
    updateWidgets(prev => prev.map(w => w.id === widgetId ? { ...w, collapsed: !w.collapsed } : w))
  }, [updateWidgets])

  const handleRemove = useCallback((widgetId: string) => {
    updateWidgets(prev => prev.filter(w => w.id !== widgetId))
  }, [updateWidgets])

  const handleMoveUp = useCallback((index: number) => {
    if (index <= 0) return
    updateWidgets(prev => {
      const next = [...prev]
      ;[next[index - 1], next[index]] = [next[index], next[index - 1]]
      return next
    })
  }, [updateWidgets])

  const handleMoveDown = useCallback((index: number) => {
    updateWidgets(prev => {
      if (index >= prev.length - 1) return prev
      const next = [...prev]
      ;[next[index], next[index + 1]] = [next[index + 1], next[index]]
      return next
    })
  }, [updateWidgets])

  const handleAddWidget = useCallback((type: string) => {
    const def = getWidgetDefinition(type)
    if (!def) return
    const id = `w-${type}-${Date.now()}`
    updateWidgets(prev => [...prev, { id, type, collapsed: false }])
  }, [updateWidgets])

  const handleSaveLayout = useCallback(() => {
    saveLayout(widgets)
    setHasChanges(false)
  }, [widgets])

  const handleResetLayout = useCallback(() => {
    setWidgets(DEFAULT_WIDGETS)
    saveLayout(DEFAULT_WIDGETS)
    setHasChanges(false)
  }, [])

  const activeWidgetTypes = widgets.map(w => w.type)

  return (
    <div className="widget-grid-wrapper">
      <div className="widget-grid-toolbar">
        <button className="widget-grid-toolbar__btn widget-grid-toolbar__btn--primary" onClick={() => setPickerOpen(true)}>
          <span>+</span> Widget hinzufügen
        </button>
        <button
          className={`widget-grid-toolbar__btn ${hasChanges ? 'widget-grid-toolbar__btn--accent' : ''}`}
          onClick={handleSaveLayout}
          disabled={!hasChanges}
        >
          💾 Layout speichern
        </button>
        <button className="widget-grid-toolbar__btn" onClick={handleResetLayout}>
          ↺ Layout zurücksetzen
        </button>
      </div>

      <div className="widget-grid">
        {widgets.map((widget, index) => {
          const def = getWidgetDefinition(widget.type)
          if (!def) return null
          const WidgetComponent = def.component
          const isFullWidth = def.defaultSize === '2x1'

          return (
            <div
              key={widget.id}
              className={`widget-grid__cell ${isFullWidth ? 'widget-grid__cell--full' : ''}`}
            >
              <WidgetContainer
                title={def.title}
                icon={def.icon}
                collapsed={widget.collapsed}
                onToggleCollapse={() => handleToggleCollapse(widget.id)}
                onRemove={() => handleRemove(widget.id)}
                onMoveUp={() => handleMoveUp(index)}
                onMoveDown={() => handleMoveDown(index)}
                isFirst={index === 0}
                isLast={index === widgets.length - 1}
              >
                <WidgetComponent widgetId={widget.id} />
              </WidgetContainer>
            </div>
          )
        })}
      </div>

      {widgets.length === 0 && (
        <div className="widget-grid-empty">
          <div className="widget-grid-empty__icon">📋</div>
          <h3 className="widget-grid-empty__title">Keine Widgets aktiv</h3>
          <p className="widget-grid-empty__text">
            Klicken Sie auf "Widget hinzufügen" um Ihr Dashboard zu gestalten.
          </p>
          <button
            className="widget-grid-toolbar__btn widget-grid-toolbar__btn--primary"
            onClick={() => setPickerOpen(true)}
          >
            <span>+</span> Widget hinzufügen
          </button>
        </div>
      )}

      <WidgetPicker
        isOpen={pickerOpen}
        onClose={() => setPickerOpen(false)}
        availableWidgets={WIDGET_REGISTRY}
        activeWidgetTypes={activeWidgetTypes}
        onAddWidget={handleAddWidget}
      />
    </div>
  )
}
