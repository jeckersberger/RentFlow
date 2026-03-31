import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

function TaxExemptionPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [form, setForm] = useState({
    enabled: false,
    threshold_previous_year: 25000,
    threshold_current_year: 100000,
  })

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'finance.kleinunternehmer'],
    queryFn: () => configApi.get('finance.kleinunternehmer'),
  })

  useEffect(() => {
    if (config) {
      setForm({
        enabled: config.enabled ?? false,
        threshold_previous_year: config.threshold_previous_year ?? 25000,
        threshold_current_year: config.threshold_current_year ?? 100000,
      })
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('finance.kleinunternehmer', form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'finance.kleinunternehmer'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Steuerbefreiung</h1></div>
      <SkeletonCard count={1} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Kleinunternehmerregelung</h1>
        <p>Konfigurieren Sie die Umsatzsteuerbefreiung nach &sect;19 UStG.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Status</h3>
        <label className="sp-toggle">
          <div className="sp-toggle__label">
            <strong>Kleinunternehmerregelung aktiv</strong>
            <span>Rechnungen werden ohne Umsatzsteuer ausgestellt</span>
          </div>
          <input type="checkbox" checked={form.enabled} onChange={(e) => setForm((prev) => ({ ...prev, enabled: e.target.checked }))} />
        </label>
      </div>

      {form.enabled && (
        <>
          <div className="sp-card">
            <h3 className="sp-card__title">Grenzwerte</h3>
            <div className="sp-grid">
              <div className="sp-field">
                <label className="sp-label">Vorjahresumsatz max. (EUR)</label>
                <input
                  className="sp-input"
                  type="number"
                  min={0}
                  value={form.threshold_previous_year}
                  onChange={(e) => setForm((prev) => ({ ...prev, threshold_previous_year: parseInt(e.target.value) || 0 }))}
                />
              </div>
              <div className="sp-field">
                <label className="sp-label">Laufendes Jahr max. (EUR)</label>
                <input
                  className="sp-input"
                  type="number"
                  min={0}
                  value={form.threshold_current_year}
                  onChange={(e) => setForm((prev) => ({ ...prev, threshold_current_year: parseInt(e.target.value) || 0 }))}
                />
              </div>
            </div>
          </div>

          <div className="sp-card">
            <h3 className="sp-card__title">Hinweistext auf Rechnungen</h3>
            <div style={{
              padding: 'var(--spacing-4)',
              background: 'var(--color-bg-primary)',
              borderRadius: 'var(--radius-md)',
              border: '1px solid var(--color-border)',
              color: 'var(--color-text-secondary)',
              fontSize: 'var(--font-size-sm)',
              lineHeight: 1.6,
              fontStyle: 'italic',
            }}>
              Gemäß &sect;19 UStG wird keine Umsatzsteuer berechnet.
              <br />
              (Kleinunternehmerregelung)
            </div>
          </div>
        </>
      )}

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

export default TaxExemptionPage
