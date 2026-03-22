import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
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

const mockServices: ServiceHealth[] = [
  { id: '1', name: 'Equipment Service', status: 'operational', uptime: 99.98, response_time: 45, last_check: '2026-03-22T14:35:00Z' },
  { id: '2', name: 'Project Service', status: 'operational', uptime: 99.95, response_time: 52, last_check: '2026-03-22T14:34:00Z' },
  { id: '3', name: 'Crew Service', status: 'operational', uptime: 99.99, response_time: 38, last_check: '2026-03-22T14:35:00Z' },
  { id: '4', name: 'Invoice Service', status: 'operational', uptime: 99.92, response_time: 67, last_check: '2026-03-22T14:34:00Z' },
  { id: '5', name: 'Transport Service', status: 'degraded', uptime: 99.45, response_time: 152, last_check: '2026-03-22T14:33:00Z' },
  { id: '6', name: 'Maintenance Service', status: 'operational', uptime: 99.97, response_time: 41, last_check: '2026-03-22T14:35:00Z' },
  { id: '7', name: 'Document Service', status: 'operational', uptime: 99.89, response_time: 58, last_check: '2026-03-22T14:34:00Z' },
  { id: '8', name: 'Insurance Service', status: 'operational', uptime: 99.93, response_time: 49, last_check: '2026-03-22T14:35:00Z' },
  { id: '9', name: 'Reporting Service', status: 'operational', uptime: 99.98, response_time: 102, last_check: '2026-03-22T14:34:00Z' },
  { id: '10', name: 'Authentication Service', status: 'operational', uptime: 100.0, response_time: 25, last_check: '2026-03-22T14:35:00Z' },
  { id: '11', name: 'Cache Service', status: 'operational', uptime: 99.99, response_time: 12, last_check: '2026-03-22T14:35:00Z' },
  { id: '12', name: 'Database Service', status: 'operational', uptime: 99.94, response_time: 78, last_check: '2026-03-22T14:34:00Z' },
  { id: '13', name: 'File Storage Service', status: 'operational', uptime: 99.96, response_time: 89, last_check: '2026-03-22T14:35:00Z' },
  { id: '14', name: 'Email Service', status: 'operational', uptime: 99.87, response_time: 234, last_check: '2026-03-22T14:33:00Z' },
  { id: '15', name: 'Notification Service', status: 'operational', uptime: 99.91, response_time: 56, last_check: '2026-03-22T14:34:00Z' },
  { id: '16', name: 'Analytics Service', status: 'operational', uptime: 99.99, response_time: 124, last_check: '2026-03-22T14:35:00Z' },
  { id: '17', name: 'Payment Service', status: 'operational', uptime: 99.99, response_time: 178, last_check: '2026-03-22T14:35:00Z' },
  { id: '18', name: 'Audit Service', status: 'operational', uptime: 99.98, response_time: 67, last_check: '2026-03-22T14:34:00Z' },
]

const mockSettings: SystemSettings = {
  company_name: 'RentFlow GmbH',
  timezone: 'Europe/Berlin',
  language: 'de-DE',
  maintenance_mode: false,
  debug_mode: false,
}

const mockDatabaseStats: DatabaseStats = {
  total_size_gb: 500,
  used_size_gb: 287,
  tables_count: 42,
  backups_count: 24,
  last_backup: '2026-03-22T02:00:00Z',
}

function AdminPage() {
  const [settings, setSettings] = useState<SystemSettings>(mockSettings)
  const [isEditingSettings, setIsEditingSettings] = useState(false)

  const { data: services = mockServices } = useQuery({
    queryKey: ['system-services'],
    queryFn: async () => mockServices,
    staleTime: 1000 * 60,
  })

  const { data: dbStats = mockDatabaseStats } = useQuery({
    queryKey: ['database-stats'],
    queryFn: async () => mockDatabaseStats,
    staleTime: 1000 * 60 * 5,
  })

  const operationalCount = services.filter(s => s.status === 'operational').length
  const degradedCount = services.filter(s => s.status === 'degraded').length
  const downCount = services.filter(s => s.status === 'down').length

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
    // Save settings logic
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
                setSettings(mockSettings)
              }
              setIsEditingSettings(!isEditingSettings)
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
            <span style={{ color: '#10b981' }}>● {operationalCount} Operational</span>
            <span style={{ color: '#f59e0b' }}>● {degradedCount} Degraded</span>
            <span style={{ color: '#ef4444' }}>● {downCount} Down</span>
          </div>
        </div>

        <div className={styles.servicesGrid}>
          {services.map(service => (
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
