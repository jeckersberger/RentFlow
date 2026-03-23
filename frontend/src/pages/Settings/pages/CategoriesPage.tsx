import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { categoryApi } from '../../../services/api'
import { Plus, Tags } from 'lucide-react'
import { Category } from '../../../types/equipment'
import { SkeletonTable } from '../../../components/Skeleton/SkeletonLoader'
import EmptyState from '../../../components/EmptyState/EmptyState'
import '../Settings.module.scss'

function CategoriesPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [newName, setNewName] = useState('')
  const [newIcon, setNewIcon] = useState('')
  const [newColor, setNewColor] = useState('#6366f1')

  const { data: categoriesData, isLoading } = useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list(),
    staleTime: 1000 * 60 * 10,
  })

  const createMutation = useMutation({
    mutationFn: (data: { name: string; icon: string; color: string }) => categoryApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      setNewName('')
      setNewIcon('')
      setNewColor('#6366f1')
      setShowForm(false)
    },
  })

  const categories = Array.isArray(categoriesData) ? categoriesData : []

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Kategorien</h1>
        <p>Verwalten Sie die Ausrüstungskategorien.</p>
      </div>

      <div style={{ marginBottom: 'var(--spacing-4)' }}>
        <button className="sp-btn sp-btn--primary" onClick={() => setShowForm(!showForm)}>
          <Plus size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          Neue Kategorie
        </button>
      </div>

      {showForm && (
        <div className="sp-card">
          <h3 className="sp-card__title">Neue Kategorie</h3>
          <div className="sp-grid" style={{ gridTemplateColumns: '1fr 100px 60px' }}>
            <div className="sp-field">
              <label className="sp-label">Name</label>
              <input className="sp-input" value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="z.B. Beleuchtung" />
            </div>
            <div className="sp-field">
              <label className="sp-label">Icon</label>
              <input className="sp-input" value={newIcon} onChange={(e) => setNewIcon(e.target.value)} placeholder="💡" />
            </div>
            <div className="sp-field">
              <label className="sp-label">Farbe</label>
              <input type="color" value={newColor} onChange={(e) => setNewColor(e.target.value)}
                style={{ width: '100%', height: 38, border: 'none', borderRadius: 'var(--radius-md)', cursor: 'pointer' }} />
            </div>
          </div>
          <div className="sp-footer" style={{ marginTop: 'var(--spacing-3)' }}>
            <button className="sp-btn sp-btn--secondary" onClick={() => setShowForm(false)}>Abbrechen</button>
            <button className="sp-btn sp-btn--primary" onClick={() => createMutation.mutate({ name: newName.trim(), icon: newIcon.trim() || '📦', color: newColor })} disabled={createMutation.isPending || !newName.trim()}>
              {createMutation.isPending ? 'Erstellen...' : 'Erstellen'}
            </button>
          </div>
        </div>
      )}

      <div className="sp-card" style={{ padding: 0, overflow: 'hidden' }}>
        {isLoading ? (
          <SkeletonTable rows={4} columns={3} />
        ) : categories.length === 0 ? (
          <EmptyState
            icon={Tags}
            title="Keine Kategorien vorhanden"
            description="Erstellen Sie Ihre erste Kategorie, um Ausrüstung zu organisieren."
            action={{ label: 'Neue Kategorie', onClick: () => setShowForm(true) }}
            compact
          />
        ) : (
          <table className="sp-table">
            <thead>
              <tr>
                <th>Icon</th>
                <th>Name</th>
                <th>Farbe</th>
              </tr>
            </thead>
            <tbody>
              {categories.map((cat) => (
                <tr key={cat.id}>
                  <td style={{ fontSize: '1.25rem' }}>{cat.icon || '📦'}</td>
                  <td>{cat.name}</td>
                  <td>
                    {cat.color && (
                      <span style={{ display: 'inline-block', width: 16, height: 16, borderRadius: '50%', backgroundColor: cat.color, verticalAlign: 'middle' }} />
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  )
}

export default CategoriesPage
