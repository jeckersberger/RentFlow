import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { Plus, Trash2 } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

interface VatScheme {
  id: string
  name: string
  rate: number
  is_default: boolean
}

const defaultSchemes: VatScheme[] = [
  { id: '1', name: 'Standard', rate: 19, is_default: true },
  { id: '2', name: 'Ermäßigt', rate: 7, is_default: false },
  { id: '3', name: 'Steuerfrei', rate: 0, is_default: false },
]

function VatSchemesPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [schemes, setSchemes] = useState<VatScheme[]>(defaultSchemes)
  const [newName, setNewName] = useState('')
  const [newRate, setNewRate] = useState('')

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'finance.vat'],
    queryFn: () => configApi.get('finance.vat'),
  })

  useEffect(() => {
    if (config?.schemes) setSchemes(config.schemes)
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('finance.vat', { schemes }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'finance.vat'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const addScheme = () => {
    if (!newName.trim()) return
    setSchemes((prev) => [...prev, { id: crypto.randomUUID(), name: newName.trim(), rate: parseFloat(newRate) || 0, is_default: false }])
    setNewName('')
    setNewRate('')
  }

  const removeScheme = (id: string) => setSchemes((prev) => prev.filter((s) => s.id !== id))

  const setDefault = (id: string) => {
    setSchemes((prev) => prev.map((s) => ({ ...s, is_default: s.id === id })))
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Steuersätze</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Steuersätze</h1>
        <p>Verwalten Sie die verfügbaren Mehrwertsteuersätze.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Neuen Steuersatz hinzufügen</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', alignItems: 'flex-end' }}>
          <div className="sp-field" style={{ flex: 1 }}>
            <label className="sp-label">Bezeichnung</label>
            <input className="sp-input" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="z.B. Innergemeinschaftlich" />
          </div>
          <div className="sp-field" style={{ width: 120 }}>
            <label className="sp-label">Satz (%)</label>
            <input className="sp-input" type="number" value={newRate} onChange={(e) => setNewRate(e.target.value)} placeholder="0" />
          </div>
          <button className="sp-btn sp-btn--primary" onClick={addScheme} style={{ height: 38 }}><Plus size={16} /></button>
        </div>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="sp-table">
          <thead>
            <tr>
              <th>Bezeichnung</th>
              <th>Satz</th>
              <th>Standard</th>
              <th style={{ width: 40 }}></th>
            </tr>
          </thead>
          <tbody>
            {schemes.map((s) => (
              <tr key={s.id}>
                <td>{s.name}</td>
                <td>{s.rate}%</td>
                <td>
                  <input type="radio" name="default-vat" checked={s.is_default} onChange={() => setDefault(s.id)}
                    style={{ accentColor: 'var(--color-primary)', cursor: 'pointer' }} />
                </td>
                <td>
                  <button onClick={() => removeScheme(s.id)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: 4 }}>
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

export default VatSchemesPage
