import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

function IntegrationsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [form, setForm] = useState({
    ai_provider: 'openai',
    ai_api_key: '',
    ai_model: 'gpt-4',
    webhook_url: '',
  })

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'integrations'],
    queryFn: () => configApi.get('integrations'),
  })

  useEffect(() => {
    if (config) {
      setForm({
        ai_provider: config.ai_provider || 'openai',
        ai_api_key: config.ai_api_key || '',
        ai_model: config.ai_model || 'gpt-4',
        webhook_url: config.webhook_url || '',
      })
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('integrations', form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'integrations'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const update = (field: string, value: string) =>
    setForm((prev) => ({ ...prev, [field]: value }))

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Integrationen</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Integrationen</h1>
        <p>Konfigurieren Sie AI-Provider, API-Keys und Webhooks.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">KI-Konfiguration</h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">AI Provider</label>
            <select className="sp-select" value={form.ai_provider} onChange={(e) => update('ai_provider', e.target.value)}>
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic</option>
              <option value="local">Lokal (Ollama)</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">Modell</label>
            <input className="sp-input" value={form.ai_model} onChange={(e) => update('ai_model', e.target.value)} />
          </div>
          <div className="sp-field sp-full">
            <label className="sp-label">API Key</label>
            <input className="sp-input" type="password" value={form.ai_api_key} onChange={(e) => update('ai_api_key', e.target.value)} placeholder="sk-..." />
          </div>
        </div>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Webhooks</h3>
        <div className="sp-field">
          <label className="sp-label">Webhook URL</label>
          <input className="sp-input" value={form.webhook_url} onChange={(e) => update('webhook_url', e.target.value)} placeholder="https://..." />
        </div>
      </div>

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
