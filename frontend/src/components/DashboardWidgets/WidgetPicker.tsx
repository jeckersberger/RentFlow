import { WidgetDefinition } from './types'
import './DashboardWidgets.scss'

interface WidgetPickerProps {
  isOpen: boolean
  onClose: () => void
  availableWidgets: WidgetDefinition[]
  activeWidgetTypes: string[]
  onAddWidget: (type: string) => void
}

export function WidgetPicker({ isOpen, onClose, availableWidgets, activeWidgetTypes, onAddWidget }: WidgetPickerProps) {
  if (!isOpen) return null

  return (
    <div className="widget-picker-backdrop" onClick={onClose}>
      <div className="widget-picker" onClick={(e) => e.stopPropagation()}>
        <div className="widget-picker__header">
          <h2 className="widget-picker__title">Widget hinzufügen</h2>
          <button className="widget-picker__close" onClick={onClose}>✕</button>
        </div>

        <div className="widget-picker__grid">
          {availableWidgets.map((widget) => {
            const isActive = activeWidgetTypes.includes(widget.type)
            return (
              <button
                key={widget.type}
                className={`widget-picker__card ${isActive ? 'widget-picker__card--active' : ''}`}
                onClick={() => {
                  if (!isActive) {
                    onAddWidget(widget.type)
                    onClose()
                  }
                }}
                disabled={isActive}
              >
                <div className="widget-picker__card-icon">{widget.icon}</div>
                <div className="widget-picker__card-info">
                  <h3 className="widget-picker__card-name">{widget.title}</h3>
                  <p className="widget-picker__card-desc">{widget.description}</p>
                </div>
                <div className="widget-picker__card-size">
                  {widget.defaultSize === '2x1' ? 'Volle Breite' : 'Standard'}
                </div>
                {isActive && <div className="widget-picker__card-badge">Aktiv</div>}
              </button>
            )
          })}
        </div>
      </div>
    </div>
  )
}
