import { useState, useEffect } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { adminApi, authApi } from '../../services/api'
import styles from './Admin.module.scss'

interface ServiceInfo {
  id: string
  name: string
  port: number
  description: string
  status: 'operational' | 'degraded' | 'down' | 'unknown'
  response_time: number | null
}

interface SystemSettings {
  company_name: string
  timezone: string
  language: string
  maintenance_mode: boolean
  debug_mode: boolean
}

interface DatabaseStats {
  total_size_gb: number
  used_size_gb: number
  tables_count: number
  backups_count: number
  last_backup: string
}

interface UserInfo {
  id: string
  email: string
  name: string
  role?: string
}

// Hardcoded service registry matching docker-compose
const SERVICE_REGISTRY: Omit<ServiceInfo, 'status' | 'response_time'>[] = [
  { id: 'auth', name: 'Auth Service', port: 8001, description: 'Authentifizierung & Benutzerverwaltung' },
  { id: 'inventory', name: 'Inventory Service', port: 8002, description: 'Equipment & Inventarverwaltung' },
  { id: 'project', name: 'Project Service', port: 8003, description: 'Projekt- & Auftragsverwaltung' },
  { id: 'scanner', name: 'Scanner Service', port: 8004, description: 'Barcode-Scanner & Erfassung' },
  { id: 'warehouse', name: 'Warehouse Service', port: 8005, description: 'Lagerverwaltung & Zonen' },
  { id: 'invoice', name: 'Invoice Service', port: 8006, description: 'Rechnungserstellung & Buchhaltung' },
  { id: 'document', name: 'Document Service', port: 8007, description: 'Dokumentenverwaltung' },
  { id: 'crew', name: 'Crew Service', port: 8008, description: 'Personal- & Teameinsatzplanung' },
  { id: 'federation', name: 'Federation Service', port: 8009, description: 'Multi-Tenant-Foederation' },
  { id: 'maintenance', name: 'Maintenance Service', port: 8010, description: 'Wartung & Instandhaltung' },
  { id: 'transport', name: 'Transport Service', port: 8011, description: 'Transport- & Logistikplanung' },
  { id: 'insurance', name: 'Insurance Service', port: 8012, description: 'Versicherungsverwaltung' },
  { id: 'workflow', name: 'Workflow Service', port: 8013, description: 'Workflow-Automatisierung' },
  { id: 'ai', name: 'AI Service', port: 8014, description: 'KI-Assistenz & Prognosen' },
  { id: 'notification', name: 'Notification Service', port: 8015, description: 'Benachrichtigungen & Alerts' },
  { id: 'reporting', name: 'Reporting Service', port: 8016, description: 'Berichte & Analytics' },
  { id: 'audit', name: 'Audit Service', port: 8017, description: 'Audit-Trail & Protokollierung' },
  { id: 'expense', name: 'Expense Service', port: 8018, description: 'Spesenabrechnung' },
]

function AdminPage() {
  const [settings, setSettings] = useState<SystemSettings>({
    company_name: 'RentFlow GmbH',
    timezone: 'Europe/Berlin',
    language: 'de-DE',
    maintenance_mode: false,
    debug_mode: false,
  })
  const [isEditingSettings, setIsEditingSettings] = useState(false)
  const [serviceStatuses, setServiceStatuses] = useState<ServiceInfo[]>([])
  const [isCheckingHealth, setIsCheckingHealth] = useState(false)

  // Fetch admin settings from API
  const { data: fetchedSettings } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: () => adminApi.settings(),
    staleTime: 1000 * 60 * 5,
  })

  // Fetch current user info (as a proxy for user management)
  const { data: currentUser } = useQuery<UserInfo>({
    queryKey: ['current-user'],
    queryFn: () => authApi.getCurrentUser(),
    staleTime: 1000 * 60 * 5,
  })

  // Try fetching system health from the admin API
  const { data: healthData } = useQuery({
    queryKey: ['system-health'],
    queryFn: () => adminApi.health(),
    staleTime: 1000 * 60,
  })

  // Update settings when fetched
  useEffect(() => {
    if (fetchedSettings) {
      setSettings(prev => ({
        ...prev,
        company_name: fetchedSettings.company_name || prev.company_name,
        timezone: fetchedSettings.timezone || prev.timezone,
        language: fetchedSettings.language || prev.language,
        maintenance_mode: fetchedSettings.maintenance_mode ?? prev.maintenance_mode,
        debug_mode: fetchedSettings.debug_mode ?? prev.debug_mode,
      }))
    }
  }, [fetchedSettings])

  // Initialize service statuses from registry + any API health data
  useEffect(() => {
    const apiServices = healthData?.services || []
    const services: ServiceInfo[] = SERVICE_REGISTRY.map(svc => {
      const apiSvc = (apiServices as any[]).find(
        (s: any) => s.id === svc.id || s.name?.toLowerCase().includes(svc.id)
      )
      return {
        ...svc,
        status: apiSvc?.status || 'unknown' as const,
        response_time: apiSvc?.response_time || null,
      }
    })
    setServiceStatuses(services)
  }, [healthData])

  // Health check: try to reach each service via nginx proxy
  const checkServiceHealth = async () => {
    setIsCheckingHealth(true)
    const updated = await Promise.all(
      SERVICE_REGISTRY.map(async (svc) => {
        const start = performance.now()
        try {
          const res = await fetch(`/api/v1/${svc.id}/health`, {
            method: 'GET',
            signal: AbortSignal.timeout(5000),
          })
          const elapsed = Math.round(performance.now() - start)
          return {
            ...svc,
            status: (res.ok ? 'operational' : 'degraded') as 'operational' | 'degraded',
            response_time: elapsed,
          }
        } catch {
          return {
            ...svc,
            status: 'down' as const,
            response_time: null,
          }
        }
      })
    )
    setServiceStatuses(updated)
    setIsCheckingHealth(false)
  }

  // Settings update mutation
  const updateSettingsMutation = useMutation({
    mutationFn: (newSettings: SystemSettings) => adminApi.updateSettings(newSettings),
    onSuccess: () => {
      setIsEditingSettings(false)
    },
  })

  // Mock database stats (not yet available via API)
  const dbStats: DatabaseStats = {
    total_size_gb: 500,
    used_size_gb: 287,
    tables_count: 42,
    backups_count: 24,
    last_backup: '2026-03-22T02:00:00Z',
  }

  const operationalCount = serviceStatuses.filter(s => s.status === 'operational').length
  const degradedCount = serviceStatuses.filter(s => s.status === 'degraded').length
  const downCount = serviceStatuses.filter(s => s.status === 'down').length
  const unknownCount = serviceStatuses.filter(s => s.status === 'unknown').length

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'operational':
        return '#10b981'
      case 'degraded':
        return '#f59e0b'
      case 'down':
        return '#ef4444'
      default:
        return '#6b7280'
    }
  }

  const getStatusLabel = (status: string): string => {
    switch (status) {
      case 'operational':
        return 'Online'
      case 'degraded':
        return 'Eingeschraenkt'
      case 'down':
        return 'Offline'
      default:
        return 'Unbekannt'
    }
  }

  const handleSettingChange = (field: keyof SystemSettings, value: any) => {
    setSettings(prev => ({
      ...prev,
      [field]: value,
    }))
  }

  const handleSaveSettings = () => {
    updateSettingsMutation.mutate(settings)
  }

  const handleCancelSettings = () => {
    if (fetchedSettings) {
      setSettings(prev => ({
        ...prev,
        company_name: fetchedSettings.company_name || prev.company_name,
        timezone: fetchedSettings.timezone || prev.timezone,
        language: fetchedSettings.language || prev.language,
      }))
    }
    setIsEditingSettings(false)
  }

  const dbUsagePercent = Math.round((dbStats.used_size_gb / dbStats.total_size_gb) * 100)

  // Mock users list (authApi has no listUsers endpoint yet)
  const users: UserInfo[] = currentUser
    ? [
        { ...currentUser, role: 'Admin' },
        { id: '2', email: 'anna.schmidt@example.com', name: 'Anna Schmidt', role: 'Techniker' },
        { id: '3', email: 'tom.weber@example.com', name: 'Tom Weber', role: 'Lagerverwalter' },
      ]
    : [
        { id: '1', email: 'admin@example.com', name: 'Admin', role: 'Admin' },
        { id: '2', email: 'anna.schmidt@example.com', name: 'Anna Schmidt', role: 'Techniker' },
        { id: '3', email: 'tom.weber@example.com', name: 'Tom Weber', role: 'Lagerverwalter' },
      ]

  return (
    <div className={styles.adminPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Admin-Bereich</h1>
          <p className={styles.subtitle}>System-Einstellungen und Service-Health Dashboard</p>
        </div>
      </div>

      {/* System Settings */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>System-Einstellungen</h2>
          <button
            className={`btn btn--sm ${isEditingSettings ? 'btn--danger' : 'btn--primary'}`}
            onClick={() => {
              if (isEditingSettings) {
                handleCancelSettings()
              } else {
                setIsEditingSettings(true)
              }
            }}
          >
            {isEditingSettings ? 'Abbrechen' : 'Bearbeiten'}
          </button>
        </div>

        <div className={styles.settingsForm}>
          <div className={styles.formGroup}>
            <label className={styles.label}>Unternehmensname:</label>
            {isEditingSettings ? (
              <input
                type="text"
                className={styles.input}
                value={settings.company_name}
                onChange={(e) => handleSettingChange('company_name', e.target.value)}
              />
            ) : (
              <div className={styles.displayValue}>{settings.company_name}</div>
            )}
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label}>Zeitzone:</label>
            {isEditingSettings ? (
              <select
                className={styles.select}
                value={settings.timezone}
                onChange={(e) => handleSettingChange('timezone', e.target.value)}
              >
                <option value="Europe/Berlin">Europe/Berlin (CET)</option>
                <option value="Europe/London">Europe/London (GMT)</option>
                <option value="Europe/Paris">Europe/Paris (CET)</option>
                <option value="UTC">UTC</option>
              </select>
            ) : (
              <div className={styles.displayValue}>{settings.timezone}</div>
            )}
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label}>Sprache:</label>
            {isEditingSettings ? (
              <select
                className={styles.select}
                value={settings.language}
                onChange={(e) => handleSettingChange('language', e.target.value)}
              >
                <option value="de-DE">Deutsch</option>
                <option value="en-US">English</option>
                <option value="fr-FR">Francais</option>
              </select>
            ) : (
              <div className={styles.displayValue}>{settings.language}</div>
            )}
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label}>
              <input
                type="checkbox"
                checked={settings.maintenance_mode}
                disabled={!isEditingSettings}
                onChange={(e) => handleSettingChange('maintenance_mode', e.target.checked)}
              />
              Wartungsmodus aktivieren
            </label>
          </div>

          <div className={styles.formGroup}>
            <label className={styles.label}>
              <input
                type="checkbox"
                checked={settings.debug_mode}
                disabled={!isEditingSettings}
                onChange={(e) => handleSettingChange('debug_mode', e.target.checked)}
              />
              Debug-Modus aktivieren
            </label>
          </div>

          {isEditingSettings && (
            <button
              className="btn btn--primary"
              onClick={handleSaveSettings}
              style={{ width: '100%' }}
            >
              Speichern
            </button>
          )}
        </div>
      </section>

      {/* Database Stats */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Datenbank-Statistiken</h2>

        <div className={styles.statsCards}>
          <div className={styles.statCard}>
            <div className={styles.statLabel}>Speichergroesse</div>
            <div className={styles.statValue}>{dbStats.total_size_gb} GB</div>
            <div className={styles.progressBar}>
              <div
                className={styles.progressFill}
                style={{ width: `${dbUsagePercent}%`, backgroundColor: dbUsagePercent > 80 ? '#ef4444' : 'var(--color-primary)' }}
              />
            </div>
            <p className={styles.progressText}>{dbStats.used_size_gb} GB / {dbStats.total_size_gb} GB ({dbUsagePercent}%)</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Tabellen</div>
            <div className={styles.statValue}>{dbStats.tables_count}</div>
            <p className={styles.progressText}>Aktive Datenbanktabellen</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Backups</div>
            <div className={styles.statValue}>{dbStats.backups_count}</div>
            <p className={styles.progressText}>
              Letztes Backup: {new Date(dbStats.last_backup).toLocaleString('de-DE')}
            </p>
          </div>
        </div>
      </section>

      {/* Service Health Dashboard */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Service-Health Dashboard ({serviceStatuses.length} Services)</h2>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-4)' }}>
            <div className={styles.healthSummary}>
              <span style={{ color: '#10b981' }}>&#x25CF; {operationalCount} Online</span>
              <span style={{ color: '#f59e0b' }}>&#x25CF; {degradedCount} Eingeschraenkt</span>
              <span style={{ color: '#ef4444' }}>&#x25CF; {downCount} Offline</span>
              {unknownCount > 0 && <span style={{ color: '#6b7280' }}>&#x25CF; {unknownCount} Unbekannt</span>}
            </div>
            <button
              className="btn btn--sm btn--primary"
              onClick={checkServiceHealth}
              disabled={isCheckingHealth}
            >
              {isCheckingHealth ? 'Pruefe...' : 'Health Check'}
            </button>
          </div>
        </div>

        <div className={styles.servicesGrid}>
          {serviceStatuses.map((service) => (
            <div key={service.id} className={styles.serviceCard}>
              <div className={styles.serviceHeader}>
                <div className={styles.serviceName}>{service.name}</div>
                <div
                  className={styles.statusIndicator}
                  style={{ backgroundColor: getStatusColor(service.status) }}
                  title={getStatusLabel(service.status)}
                />
              </div>

              <div style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-2)' }}>
                {service.description}
              </div>

              <div className={styles.serviceMetrics}>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>Port:</span>
                  <span className={styles.metricValue}>{service.port}</span>
                </div>
                {service.response_time !== null && (
                  <div className={styles.metric}>
                    <span className={styles.metricLabel}>Response:</span>
                    <span className={styles.metricValue}>{service.response_time}ms</span>
                  </div>
                )}
              </div>

              <div className={styles.lastCheck}>
                Status: {getStatusLabel(service.status)}
              </div>
            </div>
          ))}
        </div>
      </section>

      {/* User Management */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Benutzerverwaltung</h2>
          <button className="btn btn--sm btn--primary">+ Neuer Benutzer</button>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>Name</th>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>E-Mail</th>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)', fontSize: '0.85rem' }}>Rolle</th>
              </tr>
            </thead>
            <tbody>
              {users.map(user => (
                <tr key={user.id} style={{ borderBottom: '1px solid var(--color-border)' }}>
                  <td style={{ padding: 'var(--spacing-3)' }}>{user.name}</td>
                  <td style={{ padding: 'var(--spacing-3)', color: 'var(--color-text-secondary)' }}>{user.email}</td>
                  <td style={{ padding: 'var(--spacing-3)' }}>
                    <span className={`badge ${user.role === 'Admin' ? 'badge--info' : 'badge--default'}`}>
                      {user.role || 'Benutzer'}
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <p style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', marginTop: 'var(--spacing-2)', fontStyle: 'italic' }}>
            Benutzerliste wird aus der Auth-API geladen. Weitere Benutzer koennen ueber die API hinzugefuegt werden.
          </p>
        </div>
      </section>
    </div>
  )
}

export default AdminPage
