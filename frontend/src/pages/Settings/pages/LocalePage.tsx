import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useTranslation } from 'react-i18next'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

function LocalePage() {
  const { t, i18n } = useTranslation()
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [form, setForm] = useState({
    timezone: 'Europe/Berlin',
    date_format: 'dd.MM.yyyy',
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
        date_format: config.date_format || 'dd.MM.yyyy',
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

      // Apply language change immediately
      const lang = form.language
      i18n.changeLanguage(lang)
      localStorage.setItem('rentflow_language', lang)

      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const update = (field: string, value: string) => {
    setForm((prev) => ({ ...prev, [field]: value }))

    // Apply language change immediately on selection (before save)
    if (field === 'language') {
      i18n.changeLanguage(value)
      localStorage.setItem('rentflow_language', value)
    }
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>{t('nav.settings', 'Regionale Einstellungen')}</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>{i18n.language === 'en' ? 'Language & Region' : 'Sprache & Region'}</h1>
        <p>{i18n.language === 'en'
          ? 'Adjust timezone, date format and currency.'
          : 'Passen Sie Zeitzone, Datumsformat und Währung an.'}</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">
          {i18n.language === 'en' ? 'Regional Settings' : 'Regionale Einstellungen'}
        </h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">{t('language', 'Sprache')}</label>
            <select className="sp-select" value={form.language} onChange={(e) => update('language', e.target.value)}>
              <option value="de">Deutsch</option>
              <option value="en">English</option>
              <option value="fr">Fran\u00e7ais</option>
              <option value="nl">Nederlands</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">
              {i18n.language === 'en' ? 'Timezone' : 'Zeitzone'}
            </label>
            <select className="sp-select" value={form.timezone} onChange={(e) => update('timezone', e.target.value)}>
              <option value="Europe/Berlin">Europe/Berlin (MEZ)</option>
              <option value="Europe/Vienna">Europe/Vienna</option>
              <option value="Europe/Zurich">Europe/Zurich</option>
              <option value="Europe/London">Europe/London (GMT)</option>
              <option value="America/New_York">America/New_York (EST)</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">
              {i18n.language === 'en' ? 'Date Format' : 'Datumsformat'}
            </label>
            <select className="sp-select" value={form.date_format} onChange={(e) => update('date_format', e.target.value)}>
              <option value="dd.MM.yyyy">dd.MM.yyyy</option>
              <option value="yyyy-MM-dd">yyyy-MM-dd</option>
              <option value="MM/dd/yyyy">MM/dd/yyyy</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">
              {i18n.language === 'en' ? 'Currency' : 'Währung'}
            </label>
            <select className="sp-select" value={form.currency} onChange={(e) => update('currency', e.target.value)}>
              <option value="EUR">EUR - Euro</option>
              <option value="CHF">CHF - {i18n.language === 'en' ? 'Swiss Franc' : 'Schweizer Franken'}</option>
              <option value="USD">USD - US Dollar</option>
              <option value="GBP">GBP - {i18n.language === 'en' ? 'British Pound' : 'Britisches Pfund'}</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">
              {i18n.language === 'en' ? 'Currency Symbol' : 'Währungssymbol'}
            </label>
            <input className="sp-input" value={form.currency_symbol} onChange={(e) => update('currency_symbol', e.target.value)} placeholder="\u20ac" style={{ width: 80 }} />
          </div>
        </div>

        <div className="sp-card__info" style={{ marginTop: 16, padding: '12px 16px', background: 'var(--color-surface-alt, #f5f5f5)', borderRadius: 8, fontSize: 'var(--font-size-sm)' }}>
          <strong>{i18n.language === 'en' ? 'Number format:' : 'Zahlenformat:'}</strong>{' '}
          {form.language === 'de' ? '1.234,56' : '1,234.56'} &middot;{' '}
          <strong>{i18n.language === 'en' ? 'Date:' : 'Datum:'}</strong>{' '}
          {form.date_format === 'dd.MM.yyyy' ? '24.03.2026' : form.date_format === 'MM/dd/yyyy' ? '03/24/2026' : '2026-03-24'}
        </div>
      </div>

      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">{i18n.language === 'en' ? 'Saved!' : 'Gespeichert!'}</span>}
        {saveMutation.isError && <span className="sp-msg--error">{i18n.language === 'en' ? 'Error saving' : 'Fehler beim Speichern'}</span>}
        <button className="sp-btn sp-btn--primary" onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}>
          {saveMutation.isPending ? (i18n.language === 'en' ? 'Saving...' : 'Speichern...') : t('save', 'Speichern')}
        </button>
      </div>
    </div>
  )
}

export default LocalePage
