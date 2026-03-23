import { WidgetDefinition } from './types'
import { KPIWidget } from './widgets/KPIWidget'
import { RevenueChartWidget } from './widgets/RevenueChartWidget'
import { TodayWarehouseWidget } from './widgets/TodayWarehouseWidget'
import { OverdueInvoicesWidget } from './widgets/OverdueInvoicesWidget'
import { UpcomingMaintenanceWidget } from './widgets/UpcomingMaintenanceWidget'
import { RecentActivityWidget } from './widgets/RecentActivityWidget'
import { QuickActionsWidget } from './widgets/QuickActionsWidget'
import { ProjectStatusWidget } from './widgets/ProjectStatusWidget'
import { EquipmentUtilizationWidget } from './widgets/EquipmentUtilizationWidget'
import { AIInsightsWidget } from './widgets/AIInsightsWidget'
import { WelcomeWidget } from './widgets/WelcomeWidget'
import { CalendarWidget } from './widgets/CalendarWidget'

export const WIDGET_REGISTRY: WidgetDefinition[] = [
  {
    id: 'kpi',
    type: 'kpi',
    title: 'KPI-Übersicht',
    description: 'Wichtige Kennzahlen auf einen Blick: Equipment, Projekte, Rechnungen',
    icon: '📊',
    defaultSize: '2x1',
    component: KPIWidget,
  },
  {
    id: 'revenue-chart',
    type: 'revenue-chart',
    title: 'Umsatzentwicklung',
    description: 'Monatlicher Umsatz als Balkendiagramm',
    icon: '📈',
    defaultSize: '1x1',
    component: RevenueChartWidget,
  },
  {
    id: 'quick-actions',
    type: 'quick-actions',
    title: 'Schnellzugriff',
    description: 'Schnelle Aktionen: Projekt erstellen, Equipment anlegen, Scannen',
    icon: '⚡',
    defaultSize: '1x1',
    component: QuickActionsWidget,
  },
  {
    id: 'today-warehouse',
    type: 'today-warehouse',
    title: 'Heute im Lager',
    description: 'Projekte mit Lagerbewegungen für heute',
    icon: '📦',
    defaultSize: '1x1',
    component: TodayWarehouseWidget,
  },
  {
    id: 'overdue-invoices',
    type: 'overdue-invoices',
    title: 'Überfällige Rechnungen',
    description: 'Anzahl und Summe überfälliger Rechnungen',
    icon: '⚠️',
    defaultSize: '1x1',
    component: OverdueInvoicesWidget,
  },
  {
    id: 'upcoming-maintenance',
    type: 'upcoming-maintenance',
    title: 'Nächste Wartungen',
    description: 'Anstehende Equipment-Wartungen nach Priorität',
    icon: '🔧',
    defaultSize: '1x1',
    component: UpcomingMaintenanceWidget,
  },
  {
    id: 'recent-activity',
    type: 'recent-activity',
    title: 'Letzte Aktivitäten',
    description: 'Die letzten 10 Aktionen im System',
    icon: '🕐',
    defaultSize: '2x1',
    component: RecentActivityWidget,
  },
  {
    id: 'project-status',
    type: 'project-status',
    title: 'Projektstatus',
    description: 'Verteilung der Projektstatus als Donut-Diagramm',
    icon: '🎯',
    defaultSize: '1x1',
    component: ProjectStatusWidget,
  },
  {
    id: 'equipment-utilization',
    type: 'equipment-utilization',
    title: 'Equipment-Auslastung',
    description: 'Auslastung pro Kategorie mit Gesamtübersicht',
    icon: '📊',
    defaultSize: '1x1',
    component: EquipmentUtilizationWidget,
  },
  {
    id: 'ai-insights',
    type: 'ai-insights',
    title: 'KI-Empfehlungen',
    description: 'Intelligente Vorschläge und Warnungen basierend auf Ihren Daten',
    icon: '✨',
    defaultSize: '2x1',
    component: AIInsightsWidget,
  },
  {
    id: 'welcome',
    type: 'welcome',
    title: 'Willkommen',
    description: 'Onboarding-Schritte für neue Benutzer',
    icon: '👋',
    defaultSize: '2x1',
    component: WelcomeWidget,
  },
  {
    id: 'calendar',
    type: 'calendar',
    title: 'Wochenkalender',
    description: 'Mini-Kalender mit Projekten dieser Woche',
    icon: '📅',
    defaultSize: '2x1',
    component: CalendarWidget,
  },
]

export function getWidgetDefinition(type: string): WidgetDefinition | undefined {
  return WIDGET_REGISTRY.find(w => w.type === type)
}
