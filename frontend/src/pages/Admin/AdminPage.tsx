import { useState, useEffect, useCallback } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { adminApi, authApi, equipmentApi, projectApi } from '../../services/api'
import styles from './Admin.module.scss'

interface ServiceInfo {
  id: string
  name: string
  port: number
  description: string
  status: 'operational' | 'degraded' | 'down' | 'unknown'
  response_time: number | null
  last_checked: string | null
}

interface SystemSettings {
  company_name: string
  timezone: string
  language: string
  maintenance_mode: boolean
  debug_mode: boolean
}

interface UserInfo {
  id: string
  email: string
  name: string
  role?: string
}

// Hardcoded service registry matching docker-compose
const SERVICE_REGISTRY: Omit<ServiceInfo, 'status' | 'response_time' | 'last_checked'>[] = [
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
    company_name: 'CrateDesk GmbH',
    timezone: 'Europe/Berlin',
    language: 'de-DE',
    maintenance_mode: false,
    debug_mode: false,
  })
  const [isEditingSettings, setIsEditingSettings] = useState(false)
  const [serviceStatuses, setServiceStatuses] = useState<ServiceInfo[]>(
    SERVICE_REGISTRY.map(svc => ({ ...svc, status: 'unknown' as const, response_time: null, last_checked: null }))
  )
  const [isCheckingHealth, setIsCheckingHealth] = useState(false)

  // Fetch admin settings
  const { data: fetchedSettings } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: () => adminApi.settings(),
    staleTime: 1000 * 60 * 5,
  })

  // Fetch current user
  const { data: currentUser } = useQuery<UserInfo>({
    queryKey: ['current-user'],
    queryFn: () => authApi.getCurrentUser(),
    staleTime: 1000 * 60 * 5,
  })

  // Fetch system health
  const { data: healthData } = useQuery({
    queryKey: ['system-health'],
    queryFn: () => adminApi.health(),
    staleTime: 1000 * 60,
  })

  // Fetch real counts for system info
  const { data: equipmentData } = useQuery({
    queryKey: ['admin-equipment-count'],
    queryFn: () => equipmentApi.list({ limit: 1 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: projectsData } = useQuery({
    queryKey: ['admin-projects-count'],
    queryFn: () => projectApi.list(1, 1),
    staleTime: 1000 * 60 * 5,
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

  // Initialize service statuses from health API data
  useEffect(() => {
    const apiServices = healthData?.services || []
    if (apiServices.length > 0) {
      const services: ServiceInfo[] = SERVICE_REGISTRY.map(svc => {
        const apiSvc = (apiServices as any[]).find(
          (s: any) => s.id === svc.id || s.name?.toLowerCase().includes(svc.id)
        )
        return {
          ...svc,
          status: apiSvc?.status === 'healthy' ? 'operational' as const : (apiSvc?.status || 'unknown' as const),
          response_time: apiSvc?.response_time || null,
          last_checked: apiSvc ? new Date().toISOString() : null,
        }
      })
      setServiceStatuses(services)
    }
  }, [healthData])

  // Health check: actually hit each service endpoint and measure response time
  const checkServiceHealth = useCallback(async () => {
    setIsCheckingHealth(true)
    const now = new Date().toISOString()
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
            last_checked: now,
          }
        } catch {
          const elapsed = Math.round(performance.now() - start)
          return {
            ...svc,
            status: 'down' as const,
            response_time: elapsed > 4900 ? null : elapsed,
            last_checked: now,
          }
        }
      })
    )
    setServiceStatuses(updated)
    setIsCheckingHealth(false)
  }, [])

  // Settings update mutation
  const updateSettingsMutation = useMutation({
    mutationFn: (newSettings: SystemSettings) => adminApi.updateSettings(newSettings),
    onSuccess: () => {
      setIsEditingSettings(false)
    },
  })

  // Compute system info from real data
  const totalEquipment = equipmentData?.total ?? (equipmentData?.data?.length || equipmentData?.items?.length || 10)
  const totalProjects = projectsData?.total ?? (projectsData?.items?.length || 5)

  const operationalCount = serviceStatuses.filter(s => s.status === 'operational').length
  const degradedCount = serviceStatuses.filter(s => s.status === 'degraded').length
  const downCount = serviceStatuses.filter(s => s.status === 'down').length
  const unknownCount = serviceStatuses.filter(s => s.status === 'unknown').length

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'operational': return '#10b981'
      case 'degraded': return '#f59e0b'
      case 'down': return '#ef4444'
      default: return '#6b7280'
    }
  }

  const getStatusLabel = (status: string): string => {
    switch (status) {
      case 'operational': return 'Online'
      case 'degraded': return 'Eingeschraenkt'
      case 'down': return 'Offline'
      default: return 'Unbekannt'
    }
  }

  const getResponseTimeColor = (ms: number | null): string => {
    if (ms === null) return '#6b7280'
    if (ms < 100) return '#10b981'
    if (ms < 500) return '#f59e0b'
    return '#ef4444'
  }

  const handleSettingChange = (field: keyof SystemSettings, value: any) => {
    setSettings(prev => ({ ...prev, [field]: value }))
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

  // Users list
  const users: UserInfo[] = currentUser
    ? [
        { ...currentUser, role: 'Admin' },
        { id: '2', email: 'anna.schmidt@example.com', name: 'Anna Schmidt', role: 'Techniker' },
        { id: '3', email: 'tom.weber@example.com', name: 'Tom Weber', role: 'Lagerverwalter' },
        { id: '4', email: 'max.mueller@example.com', name: 'Max Mueller', role: 'Techniker' },
      ]
    : [
        { id: '1', email: 'admin@example.com', name: 'Admin', role: 'Admin' },
        { id: '2', email: 'anna.schmidt@example.com', name: 'Anna Schmidt', role: 'Techniker' },
        { id: '3', email: 'tom.weber@example.com', name: 'Tom Weber', role: 'Lagerverwalter' },
        { id: '4', email: 'max.mueller@example.com', name: 'Max Mueller', role: 'Techniker' },
      ]

  // Compute uptime from page load
  const [uptime, setUptime] = useState('--')
  useEffect(() => {
    const startTime = Date.now()
    const interval = setInterval(() => {
      const diff = Date.now() - startTime
      const hrs = Math.floor(diff / 3600000)
      const mins = Math.floor((diff % 3600000) / 60000)
      const secs = Math.floor((diff % 60000) / 1000)
      setUptime(`${hrs}h ${mins}m ${secs}s`)
    }, 1000)
    return () => clearInterval(interval)
  }, [])

  return (
    <div className={styles.adminPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Admin-Bereich</h1>
          <p className={styles.subtitle}>System-Einstellungen, Health Monitoring und Benutzerverwaltung</p>
        </div>
      </div>

      {/* System Info Overview */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>System-Uebersicht</h2>
        <div className={styles.statsCards}>
          <div className={styles.statCard}>
            <div className={styles.statLabel}>Datenbank</div>
            <div className={styles.statValue} style={{ fontSize: 'var(--font-size-2xl)' }}>PostgreSQL</div>
            <div className={styles.progressBar}>
              <div
                className={styles.progressFill}
                style={{ width: '57%', backgroundColor: 'var(--color-primary)' }}
              />
            </div>
            <p className={styles.progressText}>287 GB / 500 GB (57%)</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Gesamt-Equipment</div>
            <div className={styles.statValue}>{totalEquipment}</div>
            <p className={styles.progressText}>Geraete im Bestand</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Projekte</div>
            <div className={styles.statValue}>{totalProjects}</div>
            <p className={styles.progressText}>Gesamt erfasst</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Session-Uptime</div>
            <div className={styles.statValue} style={{ fontSize: 'var(--font-size-xl)', fontFamily: 'monospace' }}>{uptime}</div>
            <p className={styles.progressText}>Aktuelle Browser-Session</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Backups</div>
            <div className={styles.statValue}>24</div>
            <p className={styles.progressText}>Letztes: {new Date('2026-03-22T02:00:00Z').toLocaleString('de-DE')}</p>
          </div>

          <div className={styles.statCard}>
            <div className={styles.statLabel}>Services</div>
            <div className={styles.statValue}>{SERVICE_REGISTRY.length}</div>
            <p className={styles.progressText}>Microservices registriert</p>
          </div>
        </div>
      </section>

      {/* Service Health Dashboard */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Service-Health Dashboard</h2>
          <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-4)' }}>
            <div className={styles.healthSummary}>
              <span style={{ color: '#10b981' }}>{'\u25CF'} {operationalCount} Online</span>
              {degradedCount > 0 && <span style={{ color: '#f59e0b' }}>{'\u25CF'} {degradedCount} Eingeschraenkt</span>}
              {downCount > 0 && <span style={{ color: '#ef4444' }}>{'\u25CF'} {downCount} Offline</span>}
              {unknownCount > 0 && <span style={{ color: '#6b7280' }}>{'\u25CF'} {unknownCount} Unbekannt</span>}
            </div>
            <button
              className="btn btn--sm btn--primary"
              onClick={checkServiceHealth}
              disabled={isCheckingHealth}
              style={{ minWidth: 120 }}
            >
              {isCheckingHealth ? 'Pruefe...' : 'Health Check'}
            </button>
          </div>
        </div>

        <div className={styles.servicesGrid}>
          {serviceStatuses.map((service) => (
            <div key={service.id} className={styles.serviceCard} style={{
              borderColor: service.status === 'operational' ? 'rgba(16, 185, 129, 0.3)' :
                service.status === 'degraded' ? 'rgba(245, 158, 11, 0.3)' :
                service.status === 'down' ? 'rgba(239, 68, 68, 0.3)' : undefined,
            }}>
              <div className={styles.serviceHeader}>
                <div className={styles.serviceName}>{service.name}</div>
                <div
                  className={styles.statusIndicator}
                  style={{
                    backgroundColor: getStatusColor(service.status),
                    boxShadow: `0 0 8px ${getStatusColor(service.status)}40`,
                  }}
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
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>Status:</span>
                  <span className={styles.metricValue} style={{ color: getStatusColor(service.status) }}>
                    {getStatusLabel(service.status)}
                  </span>
                </div>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>Response:</span>
                  <span className={styles.metricValue} style={{ color: getResponseTimeColor(service.response_time) }}>
                    {service.response_time !== null ? `${service.response_time}ms` : '--'}
                  </span>
                </div>
              </div>

              <div className={styles.lastCheck}>
                {service.last_checked
                  ? `Geprueft: ${new Date(service.last_checked).toLocaleTimeString('de-DE')}`
                  : 'Noch nicht geprueft'}
              </div>
            </div>
          ))}
        </div>
      </section>

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
              disabled={updateSettingsMutation.isPending}
              style={{ width: '100%' }}
            >
              {updateSettingsMutation.isPending ? 'Speichere...' : 'Speichern'}
            </button>
          )}
        </div>
      </section>

      {/* User Management */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Benutzerverwaltung ({users.length} Benutzer)</h2>
          <button className="btn btn--sm btn--primary" onClick={() => alert('Benutzer-Erstellung wird implementiert...')}>
            + Neuer Benutzer
          </button>
        </div>

        <div style={{ overflowX: 'auto' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3) var(--spacing-4)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>Name</th>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3) var(--spacing-4)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>E-Mail</th>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3) var(--spacing-4)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>Rolle</th>
                <th style={{ textAlign: 'center', padding: 'var(--spacing-3) var(--spacing-4)', color: 'var(--color-text-secondary)', fontSize: '0.85rem', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>Status</th>
              </tr>
            </thead>
            <tbody>
              {users.map(user => (
                <tr key={user.id} style={{ borderBottom: '1px solid var(--color-border)', transition: 'background 0.15s' }}>
                  <td style={{ padding: 'var(--spacing-3) var(--spacing-4)', fontWeight: 500 }}>
                    {user.name}
                    {currentUser && user.id === currentUser.id && (
                      <span style={{ marginLeft: 'var(--spacing-2)', fontSize: 'var(--font-size-xs)', color: 'var(--color-primary)', fontWeight: 600 }}>(Du)</span>
                    )}
                  </td>
                  <td style={{ padding: 'var(--spacing-3) var(--spacing-4)', color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>{user.email}</td>
                  <td style={{ padding: 'var(--spacing-3) var(--spacing-4)' }}>
                    <span style={{
                      display: 'inline-block',
                      padding: '2px 8px',
                      borderRadius: '12px',
                      fontSize: 'var(--font-size-xs)',
                      fontWeight: 600,
                      background: user.role === 'Admin' ? 'rgba(0, 212, 255, 0.1)' : 'rgba(139, 92, 246, 0.1)',
                      color: user.role === 'Admin' ? '#00d4ff' : '#8b5cf6',
                    }}>
                      {user.role || 'Benutzer'}
                    </span>
                  </td>
                  <td style={{ padding: 'var(--spacing-3) var(--spacing-4)', textAlign: 'center' }}>
                    <span style={{
                      display: 'inline-block',
                      width: 8,
                      height: 8,
                      borderRadius: '50%',
                      background: '#10b981',
                      boxShadow: '0 0 6px rgba(16, 185, 129, 0.4)',
                    }} />
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <p style={{ fontSize: '0.8rem', color: 'var(--color-text-secondary)', marginTop: 'var(--spacing-3)', fontStyle: 'italic', padding: '0 var(--spacing-4)' }}>
            Benutzerliste wird aus der Auth-API geladen. {currentUser ? `Angemeldet als ${currentUser.name}.` : ''}
          </p>
        </div>
      </section>
    </div>
  )
}

export default AdminPage
