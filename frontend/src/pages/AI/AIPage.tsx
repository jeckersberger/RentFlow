import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { aiApi, configApi } from '../../services/api'
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
// CHAT MESSAGE TYPE
// ============================================================================

interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
  provider?: string
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

function AIPage() {
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([])
  const [chatInput, setChatInput] = useState('')
  const [selectedProvider, setSelectedProvider] = useState('')

  // Fetch provider config to check if anything is configured
  const { data: providerConfig } = useQuery({
    queryKey: ['config', 'ai.providers'],
    queryFn: async () => {
      try {
        return await configApi.get('ai.providers')
      } catch {
        return null
      }
    },
    retry: 1,
    staleTime: 1000 * 60 * 5,
  })

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

  // Derive which providers are configured from the config
  const configuredProviders: Array<{ key: string; label: string; model: string; active: boolean }> = []
  if (providerConfig && typeof providerConfig === 'object') {
    const providerLabels: Record<string, string> = {
      claude: 'Claude (Anthropic)',
      openai: 'GPT-4 (OpenAI)',
      ollama: 'Ollama (Lokal)',
    }
    for (const [key, conf] of Object.entries(providerConfig)) {
      const c = conf as any
      if (c && c.enabled) {
        configuredProviders.push({
          key,
          label: providerLabels[key] || key,
          model: c.model || '',
          active: true,
        })
      }
    }
  }

  // Also merge in providers from the AI service API
  const mergedProviders = configuredProviders.length > 0
    ? configuredProviders
    : providers.map((p: any) => ({
        key: typeof p === 'string' ? p.toLowerCase() : (p.name || p.id || ''),
        label: typeof p === 'string' ? p : (p.name || p.id || 'Provider'),
        model: typeof p === 'object' ? (p.model_name || p.model || '') : '',
        active: typeof p === 'object' ? (p.is_active || p.status === 'active') : false,
      }))

  const hasConfiguredProviders = mergedProviders.some((p: any) => p.active)

  // Set default selected provider
  if (!selectedProvider && hasConfiguredProviders) {
    const first = mergedProviders.find((p: any) => p.active)
    if (first) setSelectedProvider(first.key)
  }

  // Chat mutation
  const chatMutation = useMutation({
    mutationFn: async (message: string) => {
      return aiApi.chat(message, selectedProvider)
    },
    onSuccess: (data: any) => {
      const assistantMsg: ChatMessage = {
        id: data?.id || String(Date.now()),
        role: 'assistant',
        content: data?.output_text || data?.content || data?.text || data?.response || JSON.stringify(data),
        timestamp: new Date().toISOString(),
        provider: selectedProvider,
      }
      setChatMessages((prev) => [...prev, assistantMsg])
    },
    onError: (err: any) => {
      const errorMsg: ChatMessage = {
        id: String(Date.now()),
        role: 'assistant',
        content: `Fehler: ${err?.response?.data?.error || err?.message || 'Unbekannter Fehler'}`,
        timestamp: new Date().toISOString(),
        provider: selectedProvider,
      }
      setChatMessages((prev) => [...prev, errorMsg])
    },
  })

  const sendMessage = () => {
    const text = chatInput.trim()
    if (!text || !selectedProvider) return

    const userMsg: ChatMessage = {
      id: String(Date.now()),
      role: 'user',
      content: text,
      timestamp: new Date().toISOString(),
    }
    setChatMessages((prev) => [...prev, userMsg])
    setChatInput('')
    chatMutation.mutate(text)
  }

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
            <span className={styles.headerStatValue}>{mergedProviders.filter((p: any) => p.active).length}</span>
            <span className={styles.headerStatLabel}>Provider aktiv</span>
          </div>
          <div className={styles.headerStat}>
            <span className={styles.headerStatValue}>{activityLog.length}</span>
            <span className={styles.headerStatLabel}>KI-Anfragen</span>
          </div>
        </div>
      </div>

      {/* Not configured banner */}
      {!hasConfiguredProviders && (
        <div style={{
          background: 'linear-gradient(135deg, rgba(239, 68, 68, 0.08), rgba(245, 158, 11, 0.08))',
          border: '1px solid rgba(239, 68, 68, 0.25)',
          borderRadius: 'var(--radius-card)',
          padding: 'var(--spacing-6) var(--spacing-8)',
          textAlign: 'center',
        }}>
          <div style={{ fontSize: '2.5rem', marginBottom: 'var(--spacing-3)' }}>{'\u26A0'}</div>
          <h3 style={{ color: 'var(--color-text-primary)', marginBottom: 'var(--spacing-2)', fontSize: 'var(--font-size-xl)' }}>
            KI nicht konfiguriert
          </h3>
          <p style={{ color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-5)', maxWidth: 500, margin: '0 auto var(--spacing-5)' }}>
            Es ist kein KI-Provider aktiv. Konfigurieren Sie mindestens einen Provider (Claude, GPT-4 oder Ollama) in den Einstellungen, um alle KI-Funktionen zu nutzen.
          </p>
          <Link
            to="/settings/integrations"
            className="btn btn--primary"
            style={{ display: 'inline-block', padding: 'var(--spacing-3) var(--spacing-6)', fontSize: 'var(--font-size-base)' }}
          >
            Jetzt konfigurieren
          </Link>
        </div>
      )}

      {/* Section 1: KI-Status / Provider Overview */}
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
        ) : mergedProviders.length > 0 ? (
          <div className={styles.providerGrid}>
            {mergedProviders.map((provider: any, idx: number) => (
              <div
                key={idx}
                className={`${styles.providerCard} ${provider.active ? styles.providerActive : styles.providerInactive}`}
              >
                <div className={styles.providerHeader}>
                  <span className={`${styles.statusDot} ${provider.active ? styles.statusDotActive : styles.statusDotInactive}`} />
                  <span className={styles.providerName}>{provider.label}</span>
                </div>
                <div className={styles.providerDetails}>
                  <span className={styles.providerModel}>{provider.model}</span>
                  {!provider.active && (
                    <span className={styles.providerInactiveText}>Nicht konfiguriert</span>
                  )}
                </div>
                {!provider.active && (
                  <Link to="/settings/integrations" className={styles.configureLink}>
                    Konfigurieren
                  </Link>
                )}
              </div>
            ))}
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

      {/* Section 2: KI-Chat */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>KI-Chat</h2>
        <div style={{
          background: 'var(--glass-bg)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-card)',
          overflow: 'hidden',
        }}>
          {/* Chat messages area */}
          <div style={{
            minHeight: 300,
            maxHeight: 500,
            overflowY: 'auto',
            padding: 'var(--spacing-4)',
            display: 'flex',
            flexDirection: 'column',
            gap: 'var(--spacing-3)',
          }}>
            {chatMessages.length === 0 && (
              <div style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'var(--color-text-muted)',
                padding: 'var(--spacing-8)',
                textAlign: 'center',
              }}>
                <div style={{ fontSize: '2rem', marginBottom: 'var(--spacing-3)' }}>{'\u{1F4AC}'}</div>
                <p style={{ margin: 0 }}>
                  {hasConfiguredProviders
                    ? 'Stellen Sie eine Frage an den KI-Assistenten...'
                    : 'Konfigurieren Sie zuerst einen KI-Provider in den Einstellungen.'}
                </p>
              </div>
            )}
            {chatMessages.map((msg) => (
              <div
                key={msg.id}
                style={{
                  alignSelf: msg.role === 'user' ? 'flex-end' : 'flex-start',
                  maxWidth: '75%',
                  background: msg.role === 'user'
                    ? 'linear-gradient(135deg, rgba(0, 212, 255, 0.15), rgba(0, 212, 255, 0.08))'
                    : 'rgba(255, 255, 255, 0.05)',
                  border: `1px solid ${msg.role === 'user' ? 'rgba(0, 212, 255, 0.2)' : 'var(--color-border)'}`,
                  borderRadius: 'var(--radius-card)',
                  padding: 'var(--spacing-3) var(--spacing-4)',
                }}
              >
                <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginBottom: 'var(--spacing-1)' }}>
                  {msg.role === 'user' ? 'Sie' : (msg.provider || 'KI-Assistent')}
                  {' \u00B7 '}
                  {new Date(msg.timestamp).toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })}
                </div>
                <div style={{ color: 'var(--color-text-primary)', whiteSpace: 'pre-wrap', lineHeight: 1.5 }}>
                  {msg.content}
                </div>
              </div>
            ))}
            {chatMutation.isPending && (
              <div style={{
                alignSelf: 'flex-start',
                background: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius-card)',
                padding: 'var(--spacing-3) var(--spacing-4)',
                color: 'var(--color-text-muted)',
              }}>
                Denke nach...
              </div>
            )}
          </div>

          {/* Chat input */}
          <div style={{
            borderTop: '1px solid var(--color-border)',
            padding: 'var(--spacing-3) var(--spacing-4)',
            display: 'flex',
            gap: 'var(--spacing-3)',
            alignItems: 'center',
          }}>
            {/* Provider selector */}
            <select
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value)}
              disabled={!hasConfiguredProviders}
              style={{
                background: 'var(--glass-bg)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius-sm)',
                color: 'var(--color-text-primary)',
                padding: 'var(--spacing-2) var(--spacing-3)',
                fontSize: 'var(--font-size-sm)',
                minWidth: 140,
              }}
            >
              {!hasConfiguredProviders && <option value="">Kein Provider</option>}
              {mergedProviders.filter((p: any) => p.active).map((p: any) => (
                <option key={p.key} value={p.key}>{p.label}</option>
              ))}
            </select>

            <input
              value={chatInput}
              onChange={(e) => setChatInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && !e.shiftKey && sendMessage()}
              disabled={!hasConfiguredProviders || chatMutation.isPending}
              placeholder={hasConfiguredProviders ? 'Nachricht eingeben...' : 'Kein Provider konfiguriert'}
              style={{
                flex: 1,
                background: 'rgba(255, 255, 255, 0.05)',
                border: '1px solid var(--color-border)',
                borderRadius: 'var(--radius-sm)',
                color: 'var(--color-text-primary)',
                padding: 'var(--spacing-2) var(--spacing-3)',
                fontSize: 'var(--font-size-base)',
                outline: 'none',
              }}
            />

            <button
              onClick={sendMessage}
              disabled={!hasConfiguredProviders || !chatInput.trim() || chatMutation.isPending}
              className="btn btn--primary"
              style={{ padding: 'var(--spacing-2) var(--spacing-4)', whiteSpace: 'nowrap' }}
            >
              Senden
            </button>
          </div>
        </div>
      </section>

      {/* Section 3: Preis-Optimierung */}
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

      {/* Section 4: Demand Forecasting */}
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

      {/* Section 5: Predictive Maintenance */}
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

      {/* Section 6: KI-Aktivitaetslog */}
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
                <ActivityIcon type={entry.type || entry.request_type || 'chat'} />
                <div className={styles.activityContent}>
                  <span className={styles.activityDescription}>
                    {entry.description || entry.message || entry.input_text || entry.content || 'KI-Anfrage'}
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
