import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { Plus, Trash2 } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

interface ProjectType {
  id: string
  name: string
  color: string
}

const defaultTypes: ProjectType[] = [
  { id: '1', name: 'Dryhire', color: '#00d4ff' },
  { id: '2', name: 'Band / Konzert', color: '#8b5cf6' },
  { id: '3', name: 'Produktion', color: '#10b981' },
  { id: '4', name: 'Firmen-Event', color: '#f59e0b' },
  { id: '5', name: 'Hochzeit', color: '#ec4899' },
  { id: '6', name: 'Messe', color: '#06b6d4' },
]

function ProjectTypesPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [types, setTypes] = useState<ProjectType[]>(defaultTypes)
  const [newName, setNewName] = useState('')
  const [newColor, setNewColor] = useState('#6366f1')

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'project-types'],
    queryFn: () => configApi.get('project-types'),
  })

  useEffect(() => {
    if (config?.types) setTypes(config.types)
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('project-types', { types }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'project-types'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const addType = () => {
    if (!newName.trim()) return
    setTypes((prev) => [...prev, { id: crypto.randomUUID(), name: newName.trim(), color: newColor }])
    setNewName('')
    setNewColor('#6366f1')
  }

  const removeType = (t: ProjectType) => {
    if (window.confirm(`Projekttyp '${t.name}' wirklich löschen?`)) {
      setTypes((prev) => prev.filter((item) => item.id !== t.id))
    }
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Projekttypen</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Projekttypen</h1>
        <p>Definieren Sie die verfügbaren Projekttypen für Ihre Organisation.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Neuen Typ hinzufügen</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', alignItems: 'flex-end' }}>
          <div className="sp-field" style={{ flex: 1 }}>
            <label className="sp-label">Name</label>
            <input className="sp-input" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="z.B. Festival" />
          </div>
          <div className="sp-field">
            <label className="sp-label">Farbe</label>
            <input type="color" value={newColor} onChange={(e) => setNewColor(e.target.value)}
              style={{ width: 42, height: 38, border: 'none', borderRadius: 'var(--radius-md)', cursor: 'pointer' }} />
          </div>
          <button className="sp-btn sp-btn--primary" onClick={addType} style={{ height: 38 }}>
            <Plus size={16} />
          </button>
        </div>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        <table className="sp-table">
          <thead>
            <tr>
              <th>Farbe</th>
              <th>Name</th>
              <th style={{ width: 60 }}></th>
            </tr>
          </thead>
          <tbody>
            {types.map((t) => (
              <tr key={t.id}>
                <td>
                  <span style={{ display: 'inline-block', width: 16, height: 16, borderRadius: '50%', backgroundColor: t.color, verticalAlign: 'middle' }} />
                </td>
                <td>{t.name}</td>
                <td>
                  <button onClick={() => removeType(t)} title={`Projekttyp '${t.name}' löschen`} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-danger, #ef4444)', padding: 4 }}>
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

export default ProjectTypesPage
