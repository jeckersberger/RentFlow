import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

interface Sequence {
  label: string
  key: string
  prefix: string
  next_number: number
}

const defaultSequences: Sequence[] = [
  { label: 'Rechnungen', key: 'invoice', prefix: 'RF', next_number: 1 },
  { label: 'Angebote', key: 'quote', prefix: 'AG', next_number: 1 },
  { label: 'Projekte', key: 'project', prefix: 'PRJ', next_number: 1 },
  { label: 'Gutschriften', key: 'credit_note', prefix: 'GS', next_number: 1 },
]

function NumberSequencesPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [sequences, setSequences] = useState<Sequence[]>(defaultSequences)

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'number_sequences'],
    queryFn: () => configApi.get('number_sequences'),
  })

  useEffect(() => {
    if (config?.sequences) {
      setSequences(config.sequences)
    }
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('number_sequences', { sequences }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'number_sequences'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const updateSeq = (index: number, field: keyof Sequence, value: string | number) => {
    setSequences((prev) => prev.map((s, i) => (i === index ? { ...s, [field]: value } : s)))
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Nummernkreise</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Nummernkreise</h1>
        <p>Konfigurieren Sie Präfixe und die nächste laufende Nummer für Dokumente.</p>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="sp-table">
          <thead>
            <tr>
              <th>Dokument</th>
              <th>Präfix</th>
              <th>Nächste Nr.</th>
              <th>Vorschau</th>
            </tr>
          </thead>
          <tbody>
            {sequences.map((seq, i) => (
              <tr key={seq.key}>
                <td>{seq.label}</td>
                <td>
                  <input
                    className="sp-input"
                    style={{ width: 80 }}
                    value={seq.prefix}
                    onChange={(e) => updateSeq(i, 'prefix', e.target.value)}
                  />
                </td>
                <td>
                  <input
                    className="sp-input"
                    style={{ width: 100 }}
                    type="number"
                    min={1}
                    value={seq.next_number}
                    onChange={(e) => updateSeq(i, 'next_number', parseInt(e.target.value) || 1)}
                  />
                </td>
                <td style={{ color: 'var(--color-primary)', fontFamily: 'var(--font-mono)', fontSize: 'var(--font-size-sm)' }}>
                  {seq.prefix}-2026-{String(seq.next_number).padStart(3, '0')}
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

export default NumberSequencesPage
