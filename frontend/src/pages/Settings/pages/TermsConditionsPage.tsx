import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.module.scss'

function TermsConditionsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [agbText, setAgbText] = useState('')

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'terms-conditions'],
    queryFn: () => configApi.get('terms-conditions'),
  })

  useEffect(() => {
    if (config?.text) setAgbText(config.text)
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('terms-conditions', { text: agbText }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'terms-conditions'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>AGB / Mietbedingungen</h1></div>
      <SkeletonCard count={1} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Allgemeine Geschäftsbedingungen</h1>
        <p>Bearbeiten Sie die AGB, die auf Angeboten und Rechnungen erscheinen.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">AGB-Text</h3>
        <textarea
          className="sp-textarea"
          value={agbText}
          onChange={(e) => setAgbText(e.target.value)}
          placeholder="Geben Sie hier Ihre Allgemeinen Geschäftsbedingungen ein..."
          style={{ minHeight: 300 }}
        />
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

export default TermsConditionsPage
