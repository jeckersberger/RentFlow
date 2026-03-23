import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

function BankDetailsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [form, setForm] = useState({
    account_holder: '',
    iban: '',
    bic: '',
    bank_name: '',
  })

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'finance.bank_details'],
    queryFn: () => configApi.get('finance.bank_details'),
  })

  useEffect(() => {
    if (config) {
      setForm({
        account_holder: config.account_holder || '',
        iban: config.iban || '',
        bic: config.bic || '',
        bank_name: config.bank_name || '',
      })
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('finance.bank_details', form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'finance.bank_details'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const update = (field: string, value: string) =>
    setForm((prev) => ({ ...prev, [field]: value }))

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Bankverbindung</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Bankdaten</h1>
        <p>Diese Bankdaten erscheinen auf Ihren Rechnungen.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Bankverbindung</h3>
        <div className="sp-grid">
          <div className="sp-field sp-full">
            <label className="sp-label">Kontoinhaber</label>
            <input className="sp-input" value={form.account_holder} onChange={(e) => update('account_holder', e.target.value)} />
          </div>
          <div className="sp-field sp-full">
            <label className="sp-label">IBAN</label>
            <input className="sp-input" value={form.iban} onChange={(e) => update('iban', e.target.value)} placeholder="DE89 3704 0044 0532 0130 00" style={{ fontFamily: 'var(--font-mono)', letterSpacing: '0.05em' }} />
          </div>
          <div className="sp-field">
            <label className="sp-label">BIC / SWIFT</label>
            <input className="sp-input" value={form.bic} onChange={(e) => update('bic', e.target.value)} placeholder="COBADEFFXXX" style={{ fontFamily: 'var(--font-mono)' }} />
          </div>
          <div className="sp-field">
            <label className="sp-label">Bank</label>
            <input className="sp-input" value={form.bank_name} onChange={(e) => update('bank_name', e.target.value)} placeholder="Commerzbank AG" />
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

export default BankDetailsPage
