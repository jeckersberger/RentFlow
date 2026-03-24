import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { aiApi } from '../../services/api'
import styles from './AI.module.scss'

// ============================================================================
// HELPER COMPONENTS
// ============================================================================

function ActivityIcon({ type }: { type: string }) {
  const icons: Record<string, string> = {
    price: '\u20AC',
    mail: '\u2709',
    maintenance: '\u2699',
    forecast: '\u2197',
    recognition: '\u{1F4F7}',
    chat: '\u{1F4AC}',
  }
  return <span className={styles.activityIcon}>{icons[type] || '\u2022'}</span>
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

function AIPage() {
  // Fetch providers from API
  const { data: providers = [], isLoading: isLoadingProviders } = useQuery({
    queryKey: ['ai-providers'],
    queryFn: async () => {
      const result = await aiApi.providers()
      return Array.isArray(result) ? result : []
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })

  // Fetch AI activity/history from API
  const { data: activityLog = [], isLoading: isLoadingActivity } = useQuery({
    queryKey: ['ai-history'],
    queryFn: async () => {
      const result = await aiApi.history()
      return Array.isArray(result) ? result : []
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })

  const hasProviders = providers.length > 0

  return (
    <div className={styles.dashboard}>
      {/* Header */}
      <div className={styles.dashboardHeader}>
        <div>
          <h1 className={styles.dashboardTitle}>KI-Dashboard</h1>
          <p className={styles.dashboardSubtitle}>
            Intelligente Analyse und Optimierung fuer Ihren Verleih
          </p>
        </div>
        <div className={styles.headerStats}>
          <div className={styles.headerStat}>
            <span className={styles.headerStatValue}>{providers.length}</span>
            <span className={styles.headerStatLabel}>Provider verfuegbar</span>
          </div>
          <div className={styles.headerStat}>
            <span className={styles.headerStatValue}>{activityLog.length}</span>
            <span className={styles.headerStatLabel}>KI-Anfragen</span>
          </div>
        </div>
      </div>

      {/* Section 1: KI-Status */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>KI-Status</h2>
        {isLoadingProviders ? (
          <div className={styles.providerGrid}>
            <div className={styles.providerCard} style={{ opacity: 0.5 }}>
              <div className={styles.providerHeader}>
                <span className={styles.providerName}>Laden...</span>
              </div>
            </div>
          </div>
        ) : hasProviders ? (
          <div className={styles.providerGrid}>
            {providers.map((provider: any, idx: number) => {
              const name = typeof provider === 'string' ? provider : provider.name
              const isActive = typeof provider === 'object' ? provider.status === 'active' : false
              const model = typeof provider === 'object' ? provider.model : provider
              return (
                <div
                  key={idx}
                  className={`${styles.providerCard} ${isActive ? styles.providerActive : styles.providerInactive}`}
                >
                  <div className={styles.providerHeader}>
                    <span className={`${styles.statusDot} ${isActive ? styles.statusDotActive : styles.statusDotInactive}`} />
                    <span className={styles.providerName}>{name}</span>
                  </div>
                  <div className={styles.providerDetails}>
                    <span className={styles.providerModel}>{model}</span>
                    {!isActive && (
                      <span className={styles.providerInactiveText}>Nicht konfiguriert</span>
                    )}
                  </div>
                  {!isActive && (
                    <Link to="/settings/integrations" className={styles.configureLink}>
                      Konfigurieren
                    </Link>
                  )}
                </div>
              )
            })}
          </div>
        ) : (
          <div style={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius-card)',
            padding: 'var(--spacing-8)',
            textAlign: 'center',
          }}>
            <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>{'\u2699'}</div>
            <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-2)' }}>
              KI-Assistent wird konfiguriert
            </h3>
            <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-4)' }}>
              Kein KI-Provider ist derzeit verbunden. Konfigurieren Sie einen Provider in den Einstellungen, um KI-Funktionen zu nutzen.
            </p>
            <Link to="/settings/integrations" className="btn btn--primary" style={{ display: 'inline-block', padding: 'var(--spacing-3) var(--spacing-5)' }}>
              Provider konfigurieren
            </Link>
          </div>
        )}
      </section>

      {/* Section 2: Preis-Optimierung */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Preis-Optimierung</h2>
        </div>
        <div style={{
          background: 'var(--glass-bg)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-card)',
          padding: 'var(--spacing-8)',
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>{'\u{1F4B0}'}</div>
          <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-2)' }}>
            Keine Preisvorschlaege vorhanden
          </h3>
          <p style={{ color: 'var(--color-text-secondary)' }}>
            KI-basierte Preisvorschlaege werden hier angezeigt, sobald genuegend Daten vorhanden sind und ein KI-Provider konfiguriert ist.
          </p>
        </div>
      </section>

      {/* Section 3: Demand Forecasting */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Nachfrageprognose</h2>
        </div>
        <div style={{
          background: 'var(--glass-bg)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-card)',
          padding: 'var(--spacing-8)',
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>{'\u{1F4C8}'}</div>
          <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-2)' }}>
            Keine Prognosedaten verfuegbar
          </h3>
          <p style={{ color: 'var(--color-text-secondary)' }}>
            Nachfrageprognosen werden automatisch erstellt, sobald ausreichend historische Buchungsdaten vorhanden sind.
          </p>
        </div>
      </section>

      {/* Section 4: Predictive Maintenance */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Vorausschauende Wartung</h2>
        <div style={{
          background: 'var(--glass-bg)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-card)',
          padding: 'var(--spacing-8)',
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>{'\u{1F527}'}</div>
          <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-2)' }}>
            Keine Wartungsvorhersagen
          </h3>
          <p style={{ color: 'var(--color-text-secondary)' }}>
            Vorausschauende Wartungshinweise werden angezeigt, sobald Equipment-Nutzungsdaten und Sensordaten erfasst werden.
          </p>
        </div>
      </section>

      {/* Section 5: KI-Aktivitaetslog */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>KI-Aktivitaetslog</h2>
        {isLoadingActivity ? (
          <div style={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius-card)',
            padding: 'var(--spacing-6)',
            textAlign: 'center',
            color: 'var(--color-text-secondary)',
          }}>
            Aktivitaeten werden geladen...
          </div>
        ) : activityLog.length === 0 ? (
          <div style={{
            background: 'var(--glass-bg)',
            border: '1px solid var(--color-border)',
            borderRadius: 'var(--radius-card)',
            padding: 'var(--spacing-8)',
            textAlign: 'center',
          }}>
            <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>{'\u{1F4DD}'}</div>
            <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-2)' }}>
              Keine KI-Aktivitaeten
            </h3>
            <p style={{ color: 'var(--color-text-secondary)' }}>
              KI-Aktivitaeten werden hier protokolliert, sobald KI-Funktionen genutzt werden.
            </p>
          </div>
        ) : (
          <div className={styles.activityList}>
            {activityLog.map((entry: any, idx: number) => (
              <div key={entry.id || idx} className={styles.activityItem}>
                <ActivityIcon type={entry.type || 'chat'} />
                <div className={styles.activityContent}>
                  <span className={styles.activityDescription}>
                    {entry.description || entry.message || entry.content || 'KI-Anfrage'}
                  </span>
                  <span className={styles.activityTime}>
                    {entry.timestamp || entry.created_at || ''}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}

export default AIPage
