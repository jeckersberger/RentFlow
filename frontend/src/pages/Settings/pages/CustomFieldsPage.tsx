import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { Plus, Trash2 } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

interface CustomField {
  id: string
  entity: string
  name: string
  type: string
  required: boolean
}

function CustomFieldsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [fields, setFields] = useState<CustomField[]>([])
  const [newField, setNewField] = useState({ entity: 'equipment', name: '', type: 'text', required: false })

  const { data: config, isLoading } = useQuery({
    queryKey: ['config', 'custom-fields'],
    queryFn: () => configApi.get('custom-fields'),
  })

  useEffect(() => {
    if (config?.fields) setFields(config.fields)
  }, [config])

  const saveMutation = useMutation({
    mutationFn: () => configApi.set('custom-fields', { fields }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'custom-fields'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const addField = () => {
    if (!newField.name.trim()) return
    setFields((prev) => [...prev, { ...newField, id: crypto.randomUUID(), name: newField.name.trim() }])
    setNewField({ entity: 'equipment', name: '', type: 'text', required: false })
  }

  const removeField = (id: string) => {
    if (!confirm('Eigenes Feld wirklich loeschen?')) return
    const updated = fields.filter((f) => f.id !== id)
    setFields(updated)
    // Sofort speichern
    configApi.set('custom-fields', { fields: updated }).then(() => {
      queryClient.invalidateQueries({ queryKey: ['config', 'custom-fields'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    })
  }

  const entityLabels: Record<string, string> = {
    equipment: 'Equipment',
    project: 'Projekte',
    contact: 'Kontakte',
    invoice: 'Rechnungen',
  }

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Eigene Felder</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Eigene Felder</h1>
        <p>Definieren Sie zusätzliche Felder für Equipment, Projekte und Kontakte.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Neues Feld hinzufügen</h3>
        <div className="sp-grid" style={{ gridTemplateColumns: '1fr 1fr 1fr auto auto' }}>
          <div className="sp-field">
            <label className="sp-label">Entität</label>
            <select className="sp-select" value={newField.entity} onChange={(e) => setNewField({ ...newField, entity: e.target.value })}>
              <option value="equipment">Equipment</option>
              <option value="project">Projekte</option>
              <option value="contact">Kontakte</option>
              <option value="invoice">Rechnungen</option>
            </select>
          </div>
          <div className="sp-field">
            <label className="sp-label">Feldname</label>
            <input className="sp-input" value={newField.name} onChange={(e) => setNewField({ ...newField, name: e.target.value })} placeholder="z.B. Seriennummer" />
          </div>
          <div className="sp-field">
            <label className="sp-label">Typ</label>
            <select className="sp-select" value={newField.type} onChange={(e) => setNewField({ ...newField, type: e.target.value })}>
              <option value="text">Text</option>
              <option value="number">Zahl</option>
              <option value="date">Datum</option>
              <option value="select">Auswahl</option>
              <option value="checkbox">Checkbox</option>
            </select>
          </div>
          <div className="sp-field" style={{ display: 'flex', alignItems: 'flex-end' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: 6, cursor: 'pointer', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
              <input type="checkbox" checked={newField.required} onChange={(e) => setNewField({ ...newField, required: e.target.checked })} />
              Pflicht
            </label>
          </div>
          <div className="sp-field" style={{ display: 'flex', alignItems: 'flex-end' }}>
            <button className="sp-btn sp-btn--primary" onClick={addField} style={{ height: 38 }}><Plus size={16} /></button>
          </div>
        </div>
      </div>

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        {fields.length === 0 ? (
          <div className="sp-empty">Keine eigenen Felder definiert.</div>
        ) : (
          <table className="sp-table">
            <thead>
              <tr>
                <th>Entität</th>
                <th>Feldname</th>
                <th>Typ</th>
                <th>Pflicht</th>
                <th style={{ width: 40 }}></th>
              </tr>
            </thead>
            <tbody>
              {[...fields].sort((a, b) => a.entity.localeCompare(b.entity) || a.name.localeCompare(b.name)).map((f) => (
                <tr key={f.id}>
                  <td><span className="badge badge--secondary">{entityLabels[f.entity] || f.entity}</span></td>
                  <td>{f.name}</td>
                  <td>{f.type}</td>
                  <td>{f.required ? 'Ja' : 'Nein'}</td>
                  <td>
                    <button onClick={() => removeField(f.id)} style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: 4 }}>
                      <Trash2 size={14} />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
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

export default CustomFieldsPage
