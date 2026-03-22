import { useState, useEffect } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { adminApi } from '../../services/api'
import styles from './Admin.module.scss'

interface ServiceHealth {
  id: string
  name: string
  status: 'operational' | 'degraded' | 'down'
  uptime: number
  response_time: number
  last_check: string
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

function AdminPage() {
  const [settings, setSettings] = useState<SystemSettings>({
    company_name: 'RentFlow GmbH',
    timezone: 'Europe/Berlin',
    language: 'de-DE',
    maintenance_mode: false,
    debug_mode: false,
  })
  const [isEditingSettings, setIsEditingSettings] = useState(false)

  // Fetch system health/services
  const { data: healthData, isLoading: isLoadingHealth } = useQuery({
    queryKey: ['system-health'],
    queryFn: () => adminApi.health(),
    staleTime: 1000 * 60,
  })
  const services = healthData?.services || []

  // Fetch admin settings
  const { data: fetchedSettings } = useQuery({
    queryKey: ['admin-settings'],
    queryFn: () => adminApi.settings(),
    staleTime: 1000 * 60 * 5,
  })

  // Update settings when fetched
  useEffect(() => {
    if (fetchedSettings) {
      setSettings(fetchedSettings)
    }
  }, [fetchedSettings])

  // Settings update mutation
  const updateSettingsMutation = useMutation({
    mutationFn: (newSettings: SystemSettings) => adminApi.updateSettings(newSettings),
    onSuccess: () => {
      setIsEditingSettings(false)
    },
  })

  // Mock database stats (not yet in Phase 4)
  const dbStats: DatabaseStats = {
    total_size_gb: 500,
    used_size_gb: 287,
    tables_count: 42,
    backups_count: 24,
    last_backup: '2026-03-22T02:00:00Z',
  }

  const operationalCount = (services as ServiceHealth[]).filter(s => s.status === 'operational').length
  const degradedCount = (services as ServiceHealth[]).filter(s => s.status === 'degraded').length
  const downCount = (services as ServiceHealth[]).filter(s => s.status === 'down').length

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'operational':
        return '#10b981'
      case 'degraded':
        return '#f59e0b'
      case 'down':
        return '#ef4444'
      default:
        return '#00d4ff'
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
      setSettings(fetchedSettings)
    }
    setIsEditingSettings(false)
  }

  const dbUsagePercent = Math.round((dbStats.used_size_gb / dbStats.total_size_gb) * 100)

  return (
    <div className={styles.adminPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>⚙️ Admin-Bereich</h1>
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
            {isEditingSettings ? '✕ Abbrechen' : '✎ Bearbeiten'}
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
                <option value="fr-FR">Français</option>
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
              💾 Speichern
            </button>
          )}
        </div>
      </section>

      {/* Database Stats */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Datenbank-Statistiken</h2>

        <div className={styles.statsCards}>
          <div className={styles.statCard}>
            <div className={styles.statLabel}>Speichergröße</div>
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
          <h2 className={styles.sectionTitle}>Service-Health Dashboard</h2>
          <div className={styles.healthSummary}>
            <span style={{ color: '#10b981' }}>● {isLoadingHealth ? '⏳' : operationalCount} Operational</span>
            <span style={{ color: '#f59e0b' }}>● {isLoadingHealth ? '⏳' : degradedCount} Degraded</span>
            <span style={{ color: '#ef4444' }}>● {isLoadingHealth ? '⏳' : downCount} Down</span>
          </div>
        </div>

        {isLoadingHealth ? (
          <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
            Laden...
          </div>
        ) : services.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '40px', color: 'var(--color-text-secondary)' }}>
            Keine Services verfügbar
          </div>
        ) : (
          <div className={styles.servicesGrid}>
            {(services as ServiceHealth[]).map((service: ServiceHealth) => (
            <div key={service.id} className={styles.serviceCard}>
              <div className={styles.serviceHeader}>
                <div className={styles.serviceName}>{service.name}</div>
                <div
                  className={styles.statusIndicator}
                  style={{ backgroundColor: getStatusColor(service.status) }}
                  title={service.status}
                />
              </div>

              <div className={styles.serviceMetrics}>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>Uptime:</span>
                  <span className={styles.metricValue}>{service.uptime.toFixed(2)}%</span>
                </div>
                <div className={styles.metric}>
                  <span className={styles.metricLabel}>Response:</span>
                  <span className={styles.metricValue}>{service.response_time}ms</span>
                </div>
              </div>

              <div className={styles.lastCheck}>
                Last check: {new Date(service.last_check).toLocaleTimeString('de-DE', {
                  hour: '2-digit',
                  minute: '2-digit',
                  second: '2-digit',
                })}
              </div>

              <button className="btn btn--sm" style={{ width: '100%', marginTop: 'var(--spacing-3)' }}>
                Details
              </button>
            </div>
            ))}
          </div>
        )}
      </section>

      {/* User Management */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Benutzerverwaltung</h2>
          <button className="btn btn--sm btn--primary">➕ Neuer Benutzer</button>
        </div>

        <div className={styles.userManagementPlaceholder}>
          <div className={styles.placeholderContent}>
            <p className={styles.placeholderText}>Benutzerverwaltungs-Oberfläche wird hier angezeigt</p>
            <button className="btn btn--sm">Zur Benutzerverwaltung</button>
          </div>
        </div>
      </section>
    </div>
  )
}

export default AdminPage
