import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { useModuleStore } from '../../../stores/moduleStore'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import {
  Package,
  FolderKanban,
  Receipt,
  Users,
  Inbox,
  Wrench,
  Bot,
  Globe,
  LayoutDashboard,
  Settings,
  BarChart3,
  Shield,
  type LucideIcon,
} from 'lucide-react'
import '../Settings.scss'

interface ModuleDefinition {
  id: string
  name: string
  description: string
  icon: LucideIcon
  navItems: string[]
  default: boolean
}

const MODULES: ModuleDefinition[] = [
  {
    id: 'warehouse',
    name: 'Lagerverwaltung',
    description: 'Equipment, Scanner, Lager, Labels',
    icon: Package,
    navItems: ['/', '/equipment', '/scanner', '/warehouse'],
    default: true,
  },
  {
    id: 'projects',
    name: 'Projektverwaltung',
    description: 'Projekte, Kalender, Engpässe',
    icon: FolderKanban,
    navItems: ['/projects', '/calendar', '/shortages'],
    default: true,
  },
  {
    id: 'finance',
    name: 'Finanzen',
    description: 'Rechnungen, Angebote, Kontakte, Kleinunternehmerregelung',
    icon: Receipt,
    navItems: ['/invoices', '/quotes', '/contacts'],
    default: true,
  },
  {
    id: 'team',
    name: 'Personalverwaltung',
    description: 'Crew, Zeiterfassung, Transport',
    icon: Users,
    navItems: ['/crew', '/time-tracking', '/transport'],
    default: false,
  },
  {
    id: 'communication',
    name: 'Kommunikation',
    description: 'E-Mail Posteingang & Ausgang, KI-Zuordnung',
    icon: Inbox,
    navItems: ['/mail/inbox', '/mail/sent'],
    default: false,
  },
  {
    id: 'workshop',
    name: 'Werkstatt',
    description: 'Reparaturen, Prüfungen, Bestandszählungen',
    icon: Wrench,
    navItems: ['/workshop'],
    default: false,
  },
  {
    id: 'ai',
    name: 'KI-Assistent',
    description: 'Preis-Optimierung, Demand Forecasting, Smart Features',
    icon: Bot,
    navItems: ['/ai'],
    default: false,
  },
  {
    id: 'federation',
    name: 'Federation',
    description: 'Firmen verbinden, Equipment teilen, Sub-Rental',
    icon: Globe,
    navItems: ['/federation'],
    default: false,
  },
]

const ALWAYS_ACTIVE = [
  { name: 'Dashboard', icon: LayoutDashboard, path: '/' },
  { name: 'Einstellungen', icon: Settings, path: '/settings' },
  { name: 'Reports', icon: BarChart3, path: '/reports' },
  { name: 'Audit-Log', icon: Shield, path: '/audit' },
]

function ModulesPage() {
  const queryClient = useQueryClient()
  const setStoreModules = useModuleStore((s) => s.setEnabledModules)
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [enabledIds, setEnabledIds] = useState<string[]>(['warehouse', 'projects', 'finance'])

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'modules.enabled'],
    queryFn: () => configApi.get('modules.enabled'),
  })

  useEffect(() => {
    if (config && Array.isArray(config)) {
      setEnabledIds(config)
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('modules.enabled', enabledIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'modules.enabled'] })
      setStoreModules(enabledIds)
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const toggleModule = (moduleId: string) => {
    const newIds = enabledIds.includes(moduleId)
      ? enabledIds.filter((id) => id !== moduleId)
      : [...enabledIds, moduleId]
    setEnabledIds(newIds)
    // Auto-save immediately on toggle
    configApi.set('modules.enabled', newIds).then(() => {
      setStoreModules(newIds)
      queryClient.invalidateQueries({ queryKey: ['config', 'modules.enabled'] })
    })
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Module</h1></div>
      <SkeletonCard count={4} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Module</h1>
        <p>Aktivieren oder deaktivieren Sie Funktionsbereiche von CrateDesk.</p>
      </div>

      {/* Always active modules */}
      <div className="sp-card">
        <h3 className="sp-card__title">Immer aktiv</h3>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(180px, 1fr))',
          gap: 'var(--spacing-3)',
        }}>
          {ALWAYS_ACTIVE.map((item) => {
            const IconComponent = item.icon
            return (
              <div
                key={item.path}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 'var(--spacing-3)',
                  padding: 'var(--spacing-3) var(--spacing-4)',
                  background: 'var(--color-bg-secondary)',
                  borderRadius: 'var(--radius-md)',
                  border: '1px solid var(--color-border)',
                  opacity: 0.7,
                }}
              >
                <IconComponent size={18} style={{ color: 'var(--color-accent)', flexShrink: 0 }} />
                <span style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                  {item.name}
                </span>
              </div>
            )
          })}
        </div>
      </div>

      {/* Toggleable modules */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: 'var(--spacing-4)',
      }}>
        {MODULES.map((mod) => {
          const isEnabled = enabledIds.includes(mod.id)
          const IconComponent = mod.icon
          return (
            <div
              key={mod.id}
              style={{
                background: 'var(--color-bg-primary)',
                border: isEnabled
                  ? '1px solid var(--color-accent)'
                  : '1px solid var(--color-border)',
                borderRadius: 'var(--radius-card)',
                padding: 'var(--spacing-5)',
                boxShadow: isEnabled
                  ? '0 0 20px rgba(0, 200, 255, 0.1), var(--shadow-card)'
                  : 'var(--shadow-card)',
                opacity: isEnabled ? 1 : 0.55,
                transition: 'all 0.2s ease',
                cursor: 'pointer',
                display: 'flex',
                flexDirection: 'column' as const,
                gap: 'var(--spacing-3)',
              }}
              onClick={() => toggleModule(mod.id)}
            >
              <div style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
              }}>
                <div style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 'var(--spacing-3)',
                }}>
                  <div style={{
                    width: 40,
                    height: 40,
                    borderRadius: 'var(--radius-md)',
                    background: isEnabled
                      ? 'rgba(0, 200, 255, 0.1)'
                      : 'var(--color-bg-secondary)',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    flexShrink: 0,
                  }}>
                    <IconComponent
                      size={20}
                      style={{
                        color: isEnabled ? 'var(--color-accent)' : 'var(--color-text-muted)',
                      }}
                    />
                  </div>
                  <div>
                    <div style={{
                      fontWeight: 600,
                      fontSize: 'var(--font-size-md)',
                      color: isEnabled ? 'var(--color-text-primary)' : 'var(--color-text-muted)',
                    }}>
                      {mod.name}
                    </div>
                  </div>
                </div>

                {/* Toggle switch */}
                <div
                  style={{
                    width: 44,
                    height: 24,
                    borderRadius: 12,
                    background: isEnabled
                      ? 'var(--color-accent)'
                      : 'var(--color-bg-tertiary, rgba(255,255,255,0.1))',
                    position: 'relative',
                    transition: 'background 0.2s ease',
                    flexShrink: 0,
                    cursor: 'pointer',
                  }}
                  onClick={(e) => {
                    e.stopPropagation()
                    toggleModule(mod.id)
                  }}
                >
                  <div style={{
                    width: 18,
                    height: 18,
                    borderRadius: '50%',
                    background: '#fff',
                    position: 'absolute',
                    top: 3,
                    left: isEnabled ? 23 : 3,
                    transition: 'left 0.2s ease',
                    boxShadow: '0 1px 3px rgba(0,0,0,0.3)',
                  }} />
                </div>
              </div>

              <div style={{
                fontSize: 'var(--font-size-sm)',
                color: 'var(--color-text-secondary)',
                lineHeight: 1.5,
              }}>
                {mod.description}
              </div>
            </div>
          )
        })}
      </div>

      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Module gespeichert!</span>}
        {saveMutation.isError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button
          className="sp-btn sp-btn--primary"
          onClick={() => saveMutation.mutate()}
          disabled={saveMutation.isPending}
        >
          {saveMutation.isPending ? 'Speichern...' : 'Speichern'}
        </button>
      </div>
    </div>
  )
}

export default ModulesPage
