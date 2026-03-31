export interface WidgetDefinition {
  id: string
  type: string
  title: string
  description: string
  icon: string
  defaultSize: '1x1' | '2x1'
  component: React.ComponentType<WidgetProps>
}

export interface WidgetInstance {
  id: string
  type: string
  collapsed: boolean
}

export interface WidgetProps {
  widgetId: string
}

export interface DashboardLayout {
  widgets: WidgetInstance[]
  columns: 2 | 3
}

export const STORAGE_KEY = 'rentflow-dashboard-layout'
