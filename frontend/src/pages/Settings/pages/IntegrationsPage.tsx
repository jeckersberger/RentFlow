import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi, aiApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface ProviderConfig {
  enabled: boolean
  api_key: string
  model: string
  endpoint?: string
}

interface AIProvidersConfig {
  claude: ProviderConfig
  openai: ProviderConfig
  ollama: ProviderConfig
}

interface TestResult {
  provider: string
  success: boolean
  message: string
  loading: boolean
}

// ---------------------------------------------------------------------------
// Defaults
// ---------------------------------------------------------------------------

const DEFAULT_PROVIDERS: AIProvidersConfig = {
  claude: { enabled: false, api_key: '', model: 'claude-sonnet-4-20250514' },
  openai: { enabled: false, api_key: '', model: 'gpt-4o' },
  ollama: { enabled: false, api_key: '', model: 'mistral', endpoint: 'http://localhost:11434' },
}

const MODEL_OPTIONS: Record<string, string[]> = {
  claude: ['claude-sonnet-4-20250514', 'claude-opus-4-20250514', 'claude-haiku-4-20250414'],
  openai: ['gpt-4o', 'gpt-4o-mini', 'gpt-4-turbo', 'o1', 'o1-mini'],
  ollama: ['mistral', 'llama3', 'codellama', 'deepseek-coder', 'phi3', 'gemma2'],
}

const PROVIDER_META: Record<string, { label: string; description: string; icon: string }> = {
  claude: { label: 'Claude (Anthropic)', description: 'Anthropic Claude API', icon: '\u2728' },
  openai: { label: 'GPT-4 (OpenAI)', description: 'OpenAI ChatGPT API', icon: '\u{1F916}' },
  ollama: { label: 'Ollama (Lokal)', description: 'Lokales KI-Modell via Ollama', icon: '\u{1F4BB}' },
}

// ---------------------------------------------------------------------------
// Helper: mask API key for display
// ---------------------------------------------------------------------------
function maskKey(key: string): string {
  if (!key) return ''
  if (key.length <= 8) return '\u2022'.repeat(key.length)
  return key.slice(0, 4) + '\u2022'.repeat(Math.min(20, key.length - 8)) + key.slice(-4)
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function IntegrationsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [providers, setProviders] = useState<AIProvidersConfig>(DEFAULT_PROVIDERS)
  const [testResults, setTestResults] = useState<Record<string, TestResult>>({})
  const [webhookUrl, setWebhookUrl] = useState('')
  // Track which keys are being edited (so we can show plaintext vs masked)
  const [editingKeys, setEditingKeys] = useState<Record<string, boolean>>({})

  // Fetch stored config
  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'ai.providers'],
    queryFn: async () => {
      try {
        const result = await configApi.get('ai.providers')
        return result
      } catch {
        return null
      }
    },
  })

  const { data: webhookConfig } = useQuery({
    queryKey: ['config', 'integrations'],
    queryFn: async () => {
      try {
        return await configApi.get('integrations')
      } catch {
        return null
      }
    },
  })

  useEffect(() => {
    if (config && typeof config === 'object') {
      setProviders((prev) => ({
        claude: { ...prev.claude, ...config.claude },
        openai: { ...prev.openai, ...config.openai },
        ollama: { ...prev.ollama, ...config.ollama },
      }))
    }
  }, [config])

  useEffect(() => {
    if (webhookConfig) {
      setWebhookUrl(webhookConfig.webhook_url || '')
    }
  }, [webhookConfig])

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: async () => {
      // Save AI provider config
      await configApi.set('ai.providers', providers)
      // Save webhook separately under integrations
      await configApi.set('integrations', {
        webhook_url: webhookUrl,
        ai_provider: Object.entries(providers).find(([, v]) => v.enabled)?.[0] || '',
        ai_api_key: '', // Don't store key in legacy location
        ai_model: '',
      })
      // Register enabled providers with the AI service
      for (const [key, prov] of Object.entries(providers)) {
        if (prov.enabled) {
          try {
            await aiApi.registerProvider({
              name: key,
              model_name: prov.model,
              api_endpoint: prov.endpoint,
              is_active: true,
              priority: key === 'claude' ? 1 : key === 'openai' ? 2 : 3,
              config: { model: prov.model, ...(prov.endpoint ? { endpoint: prov.endpoint } : {}) },
            })
          } catch {
            // Provider registration to ai-service may fail if service is down; config is still saved
          }
        }
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'ai.providers'] })
      queryClient.invalidateQueries({ queryKey: ['config', 'integrations'] })
      queryClient.invalidateQueries({ queryKey: ['ai-providers'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  // Update a provider field
  const updateProvider = (provider: keyof AIProvidersConfig, field: keyof ProviderConfig, value: string | boolean) => {
    setProviders((prev) => ({
      ...prev,
      [provider]: { ...prev[provider], [field]: value },
    }))
  }

  // Test provider connection
  const testProvider = async (key: string) => {
    const prov = providers[key as keyof AIProvidersConfig]
    setTestResults((prev) => ({
      ...prev,
      [key]: { provider: key, success: false, message: '', loading: true },
    }))

    try {
      const result = await aiApi.testProvider(key, prov.api_key, prov.model, prov.endpoint)
      setTestResults((prev) => ({
        ...prev,
        [key]: { provider: key, success: result.success, message: result.message, loading: false },
      }))
    } catch (err: any) {
      setTestResults((prev) => ({
        ...prev,
        [key]: { provider: key, success: false, message: err?.message || 'Test fehlgeschlagen', loading: false },
      }))
    }
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Integrationen</h1></div>
      <SkeletonCard count={3} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Integrationen</h1>
        <p>Konfigurieren Sie KI-Provider, API-Keys und Webhooks.</p>
      </div>

      {/* Provider Cards */}
      {(Object.keys(PROVIDER_META) as Array<keyof AIProvidersConfig>).map((key) => {
        const meta = PROVIDER_META[key]
        const prov = providers[key]
        const test = testResults[key]
        const isEditing = editingKeys[key]

        return (
          <div className="sp-card" key={key} style={{ opacity: prov.enabled ? 1 : 0.7 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-4)' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)' }}>
                <span style={{ fontSize: '1.5rem' }}>{meta.icon}</span>
                <div>
                  <h3 className="sp-card__title" style={{ margin: 0 }}>{meta.label}</h3>
                  <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)' }}>{meta.description}</span>
                </div>
              </div>
              <label style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)', cursor: 'pointer' }}>
                <span style={{ fontSize: 'var(--font-size-sm)', color: prov.enabled ? 'var(--color-success)' : 'var(--color-text-muted)' }}>
                  {prov.enabled ? 'Aktiv' : 'Inaktiv'}
                </span>
                <div
                  onClick={() => updateProvider(key, 'enabled', !prov.enabled)}
                  style={{
                    width: 44, height: 24, borderRadius: 12,
                    background: prov.enabled ? 'var(--color-primary)' : 'rgba(107, 114, 128, 0.3)',
                    position: 'relative', cursor: 'pointer', transition: 'background 0.2s',
                  }}
                >
                  <div style={{
                    width: 18, height: 18, borderRadius: '50%',
                    background: '#fff', position: 'absolute', top: 3,
                    left: prov.enabled ? 23 : 3, transition: 'left 0.2s',
                  }} />
                </div>
              </label>
            </div>

            {prov.enabled && (
              <div className="sp-grid">
                {/* API Key (not for Ollama) */}
                {key !== 'ollama' && (
                  <div className="sp-field sp-full">
                    <label className="sp-label">API Key</label>
                    <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
                      <input
                        className="sp-input"
                        type={isEditing ? 'text' : 'password'}
                        value={prov.api_key}
                        onChange={(e) => updateProvider(key, 'api_key', e.target.value)}
                        onFocus={() => setEditingKeys((prev) => ({ ...prev, [key]: true }))}
                        onBlur={() => setEditingKeys((prev) => ({ ...prev, [key]: false }))}
                        placeholder={key === 'claude' ? 'sk-ant-...' : 'sk-...'}
                        style={{ flex: 1 }}
                      />
                    </div>
                    {prov.api_key && !isEditing && (
                      <span style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-muted)', marginTop: 'var(--spacing-1)', display: 'block' }}>
                        Gespeichert: {maskKey(prov.api_key)}
                      </span>
                    )}
                  </div>
                )}

                {/* Ollama Server URL */}
                {key === 'ollama' && (
                  <div className="sp-field sp-full">
                    <label className="sp-label">Server URL</label>
                    <input
                      className="sp-input"
                      value={prov.endpoint || ''}
                      onChange={(e) => updateProvider(key, 'endpoint', e.target.value)}
                      placeholder="http://localhost:11434"
                    />
                  </div>
                )}

                {/* Model selection */}
                <div className="sp-field">
                  <label className="sp-label">Modell</label>
                  <select className="sp-select" value={prov.model} onChange={(e) => updateProvider(key, 'model', e.target.value)}>
                    {MODEL_OPTIONS[key].map((m) => (
                      <option key={m} value={m}>{m}</option>
                    ))}
                  </select>
                </div>

                {/* Test button */}
                <div className="sp-field" style={{ display: 'flex', alignItems: 'flex-end' }}>
                  <button
                    className="sp-btn sp-btn--secondary"
                    onClick={() => testProvider(key)}
                    disabled={test?.loading || (key !== 'ollama' && !prov.api_key)}
                    style={{ whiteSpace: 'nowrap' }}
                  >
                    {test?.loading ? 'Teste...' : 'Testen'}
                  </button>
                </div>

                {/* Test result */}
                {test && !test.loading && (
                  <div className="sp-field sp-full">
                    <span style={{
                      fontSize: 'var(--font-size-sm)',
                      color: test.success ? 'var(--color-success)' : 'var(--color-danger)',
                      display: 'flex', alignItems: 'center', gap: 'var(--spacing-2)',
                    }}>
                      {test.success ? '\u2713' : '\u2717'} {test.message}
                    </span>
                  </div>
                )}
              </div>
            )}
          </div>
        )
      })}

      {/* Webhooks */}
      <div className="sp-card">
        <h3 className="sp-card__title">Webhooks</h3>
        <div className="sp-field">
          <label className="sp-label">Webhook URL</label>
          <input className="sp-input" value={webhookUrl} onChange={(e) => setWebhookUrl(e.target.value)} placeholder="https://..." />
        </div>
      </div>

      {/* Save */}
      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Gespeichert!</span>}
        {saveMutation.isError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button className="sp-btn sp-btn--primary" onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}>
          {saveMutation.isPending ? 'Speichern...' : 'Speichern'}
        </button>
      </div>
    </div>
  )
}

export default IntegrationsPage
