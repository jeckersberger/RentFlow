import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { Plus, Trash2 } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

interface PaymentTerm {
  id: string
  name: string
  days: number
  is_default: boolean
}

const defaultTerms: PaymentTerm[] = [
  { id: '1', name: 'Sofort fällig', days: 0, is_default: false },
  { id: '2', name: '14 Tage netto', days: 14, is_default: true },
  { id: '3', name: '30 Tage netto', days: 30, is_default: false },
]

function PaymentTermsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [terms, setTerms] = useState<PaymentTerm[]>(defaultTerms)
  const [newName, setNewName] = useState('')
  const [newDays, setNewDays] = useState('')

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'finance.payment_terms'],
    queryFn: () => configApi.get('finance.payment_terms'),
  })

  useEffect(() => {
    if (config?.terms) setTerms(config.terms)
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('finance.payment_terms', { terms }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'finance.payment_terms'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const addTerm = () => {
    if (!newName.trim()) return
    setTerms((prev) => [...prev, { id: crypto.randomUUID(), name: newName.trim(), days: parseInt(newDays) || 0, is_default: false }])
    setNewName('')
    setNewDays('')
  }

  const removeTerm = (id: string) => setTerms((prev) => prev.filter((t) => t.id !== id))

  const setDefault = (id: string) => {
    setTerms((prev) => prev.map((t) => ({ ...t, is_default: t.id === id })))
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Zahlungsbedingungen</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Zahlungsbedingungen</h1>
        <p>Definieren Sie die Zahlungsfristen für Ihre Rechnungen.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Neue Zahlungsbedingung</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', alignItems: 'flex-end' }}>
          <div className="sp-field" style={{ flex: 1 }}>
            <label className="sp-label">Bezeichnung</label>
            <input className="sp-input" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="z.B. 60 Tage netto" />
          </div>
          <div className="sp-field" style={{ width: 120 }}>
            <label className="sp-label">Tage</label>
            <input className="sp-input" type="number" min={0} value={newDays} onChange={(e) => setNewDays(e.target.value)} placeholder="0" />
          </div>
          <button className="sp-btn sp-btn--primary" onClick={addTerm} style={{ height: 38 }}><Plus size={16} /></button>
        </div>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="sp-table">
          <thead>
            <tr>
              <th>Bezeichnung</th>
              <th>Tage</th>
              <th>Standard</th>
              <th style={{ width: 40 }}></th>
            </tr>
          </thead>
          <tbody>
            {terms.map((t) => (
              <tr key={t.id}>
                <td>{t.name}</td>
                <td>{t.days === 0 ? 'Sofort' : `${t.days} Tage`}</td>
                <td>
                  <input type="radio" name="default-term" checked={t.is_default} onChange={() => setDefault(t.id)}
                    style={{ accentColor: 'var(--color-primary)', cursor: 'pointer' }} />
                </td>
                <td>
                  <button onClick={() => removeTerm(t.id)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: 4 }}>
                    <Trash2 size={14} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
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

export default PaymentTermsPage
