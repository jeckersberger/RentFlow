import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

function LocalePage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [form, setForm] = useState({
    timezone: 'Europe/Berlin',
    date_format: 'DD.MM.YYYY',
    currency: 'EUR',
    currency_symbol: '\u20ac',
    language: 'de',
  })

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'locale'],
    queryFn: () => configApi.get('locale'),
  })

  useEffect(() => {
    if (config) {
      setForm({
        timezone: config.timezone || 'Europe/Berlin',
        date_format: config.date_format || 'DD.MM.YYYY',
        currency: config.currency || 'EUR',
        currency_symbol: config.currency_symbol || '\u20ac',
        language: config.language || 'de',
      })
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('locale', form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'locale'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const update = (field: string, value: string) =>
    setForm((prev) => ({ ...prev, [field]: value }))

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Regionale Einstellungen</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Sprache & Region</h1>
        <p>Passen Sie Zeitzone, Datumsformat und Währung an.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Regionale Einstellungen</h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Sprache</label>
            <select className="sp-select" value={form.language} onChange={(e) => update('language', e.target.value)}>
              <option value="de">Deutsch</option>
              <option value="en">English</option>
              <option value="fr">Français</option>
              <option value="nl">Nederlands</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">Zeitzone</label>
            <select className="sp-select" value={form.timezone} onChange={(e) => update('timezone', e.target.value)}>
              <option value="Europe/Berlin">Europe/Berlin (MEZ)</option>
              <option value="Europe/Vienna">Europe/Vienna</option>
              <option value="Europe/Zurich">Europe/Zurich</option>
              <option value="Europe/London">Europe/London (GMT)</option>
              <option value="America/New_York">America/New_York (EST)</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">Datumsformat</label>
            <select className="sp-select" value={form.date_format} onChange={(e) => update('date_format', e.target.value)}>
              <option value="DD.MM.YYYY">DD.MM.YYYY</option>
              <option value="YYYY-MM-DD">YYYY-MM-DD</option>
              <option value="MM/DD/YYYY">MM/DD/YYYY</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">Währung</label>
            <select className="sp-select" value={form.currency} onChange={(e) => update('currency', e.target.value)}>
              <option value="EUR">EUR - Euro</option>
              <option value="CHF">CHF - Schweizer Franken</option>
              <option value="USD">USD - US Dollar</option>
              <option value="GBP">GBP - Britisches Pfund</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">Währungssymbol</label>
            <input className="sp-input" value={form.currency_symbol} onChange={(e) => update('currency_symbol', e.target.value)} placeholder="€" style={{ width: 80 }} />
          </div>
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

export default LocalePage
