import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Boxes, Package, Pencil, Trash2, Plus } from 'lucide-react'
import { equipmentTypeApi, categoryApi } from '../../services/api'
import { Category } from '../../types/equipment'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { Modal } from '../../components/Modal/Modal'
import { useNotificationStore } from '../../stores/notificationStore'
import EquipmentTypeForm from './EquipmentTypeForm'
import './EquipmentTypes.scss'

interface EquipmentType {
  id: string
  name: string
  manufacturer?: string
  model?: string
  category_id?: string
  sku_prefix?: string
  rental_price_day?: number
  rental_price_week?: number
  replacement_value?: number
  weight?: number
  dimensions?: string
  image_url?: string
  tags?: string[]
  item_count?: number
  status_summary?: Record<string, number>
  created_at?: string
  updated_at?: string
}

function EquipmentTypesPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()

  const [page, setPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [formOpen, setFormOpen] = useState(false)
  const [editingType, setEditingType] = useState<EquipmentType | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<EquipmentType | null>(null)
  const limit = 30

  const { data: categoriesRaw } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list(),
    staleTime: 1000 * 60 * 10,
  })

  const categories: Category[] = Array.isArray(categoriesRaw) ? categoriesRaw : Array.isArray(categoriesRaw?.data) ? categoriesRaw.data : []

  const categoryOptions = categories.map((cat: Category) => ({
    value: cat.id,
    label: cat.name,
  }))

  const categoryMap = new Map(categories.map((c: Category) => [c.id, c.name]))

  const offset = (page - 1) * limit

  const { data: typesData, isLoading, error } = useQuery({
    queryKey: ['equipment-types', page, selectedCategory],
    queryFn: () => {
      const params: Record<string, unknown> = { limit, offset }
      if (selectedCategory) params.category_id = selectedCategory
      return equipmentTypeApi.list(params as { limit?: number; offset?: number; category_id?: string })
    },
    staleTime: 1000 * 60 * 5,
  })

  const types: EquipmentType[] = Array.isArray(typesData?.data) ? typesData.data : Array.isArray(typesData) ? typesData : []
  const total: number = typesData?.total || types.length

  // Client-side search filter
  const filteredTypes = searchQuery
    ? types.filter(t =>
        t.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        (t.manufacturer || '').toLowerCase().includes(searchQuery.toLowerCase()) ||
        (t.model || '').toLowerCase().includes(searchQuery.toLowerCase())
      )
    : types

  const { mutate: deleteType, isPending: isDeleting } = useMutation({
    mutationFn: (id: string) => equipmentTypeApi.delete(id),
    onSuccess: () => {
      setDeleteTarget(null)
      addNotification('Typ erfolgreich geloescht', 'success', { title: 'Erfolg', duration: 3000 })
      queryClient.invalidateQueries({ queryKey: ['equipment-types'] })
    },
    onError: (err: unknown) => {
      const message = (err as { response?: { data?: { message?: string } } })?.response?.data?.message
        || 'Fehler beim Loeschen. Moeglicherweise existieren noch Einzelartikel.'
      addNotification(message, 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const handleCardClick = (type: EquipmentType) => {
    navigate(`/equipment/items?type_id=${type.id}&type_name=${encodeURIComponent(type.name)}`)
  }

  const handleEdit = (e: React.MouseEvent, type: EquipmentType) => {
    e.stopPropagation()
    setEditingType(type)
    setFormOpen(true)
  }

  const handleDelete = (e: React.MouseEvent, type: EquipmentType) => {
    e.stopPropagation()
    setDeleteTarget(type)
  }

  const handleFormClose = () => {
    setFormOpen(false)
    setEditingType(null)
  }

  const handleFormSuccess = () => {
    setFormOpen(false)
    setEditingType(null)
    queryClient.invalidateQueries({ queryKey: ['equipment-types'] })
  }

  const formatPrice = (price?: number) => {
    if (price == null || price === 0) return '--'
    return `${price.toFixed(2).replace('.', ',')} EUR`
  }

  const renderSkeleton = () => (
    <div className="et-skeleton">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="et-skeleton__card">
          <div className="et-skeleton__image" />
          <div className="et-skeleton__body">
            <div className="et-skeleton__line et-skeleton__line--title" />
            <div className="et-skeleton__line et-skeleton__line--short" />
            <div className="et-skeleton__line et-skeleton__line--medium" />
          </div>
        </div>
      ))}
    </div>
  )

  return (
    <div className="equipment-types-page">
      {/* Header */}
      <div className="et-header">
        <div>
          <h1 className="et-header__title">Equipment</h1>
          <p className="et-header__subtitle">
            Legen Sie Equipment-Typen an und erstellen Sie daraus Einzelartikel.
          </p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => { setEditingType(null); setFormOpen(true) }}
        >
          <Plus size={16} />
          Typ erstellen
        </button>
      </div>

      {/* Filters */}
      <div className="et-filters">
        <Input
          type="text"
          placeholder="Nach Name, Hersteller oder Modell suchen..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value)
            setPage(1)
          }}
        />
        <Select
          options={categoryOptions}
          value={selectedCategory}
          onChange={(e) => {
            setSelectedCategory(e.target.value)
            setPage(1)
          }}
          placeholder="Alle Kategorien"
        />
      </div>

      {/* Content */}
      {error ? (
        <div className="error-message">
          Fehler beim Laden der Equipment-Typen. Bitte versuchen Sie es erneut.
        </div>
      ) : isLoading ? (
        renderSkeleton()
      ) : filteredTypes.length === 0 ? (
        <div className="et-empty">
          <div className="et-empty__icon">
            <Boxes size={64} />
          </div>
          <h2 className="et-empty__title">
            {searchQuery || selectedCategory ? 'Keine Ergebnisse' : 'Noch keine Equipment-Typen'}
          </h2>
          <p className="et-empty__text">
            {searchQuery || selectedCategory
              ? 'Versuchen Sie andere Suchbegriffe oder Filter.'
              : 'Erstellen Sie Ihren ersten Equipment-Typ, um Einzelartikel daraus zu generieren.'}
          </p>
          {!searchQuery && !selectedCategory && (
            <button
              className="btn btn--primary"
              onClick={() => { setEditingType(null); setFormOpen(true) }}
            >
              <Plus size={16} />
              Ersten Typ erstellen
            </button>
          )}
        </div>
      ) : (
        <>
          <div className="et-grid">
            {filteredTypes.map((type) => (
              <div
                key={type.id}
                className="et-card"
                onClick={() => handleCardClick(type)}
              >
                {/* Actions overlay */}
                <div className="et-card__actions">
                  <button
                    className="et-card__action-btn"
                    title="Bearbeiten"
                    onClick={(e) => handleEdit(e, type)}
                  >
                    <Pencil size={14} />
                  </button>
                  <button
                    className="et-card__action-btn et-card__action-btn--danger"
                    title="Loeschen"
                    onClick={(e) => handleDelete(e, type)}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>

                {/* Image */}
                {type.image_url ? (
                  <img
                    src={type.image_url}
                    alt={type.name}
                    className="et-card__image"
                    onError={(e) => {
                      (e.target as HTMLImageElement).style.display = 'none'
                      ;(e.target as HTMLImageElement).nextElementSibling
                        ?.classList.remove('et-card__placeholder--hidden')
                    }}
                  />
                ) : (
                  <div className="et-card__placeholder">
                    <Package size={48} />
                  </div>
                )}

                {/* Body */}
                <div className="et-card__body">
                  <div className="et-card__header">
                    <div>
                      <h3 className="et-card__name">{type.name}</h3>
                      {(type.manufacturer || type.model) && (
                        <p className="et-card__manufacturer">
                          {[type.manufacturer, type.model].filter(Boolean).join(' ')}
                        </p>
                      )}
                    </div>
                    {type.category_id && categoryMap.get(type.category_id) && (
                      <span className="et-card__category-badge">
                        {categoryMap.get(type.category_id)}
                      </span>
                    )}
                  </div>

                  <div className="et-card__meta">
                    <div className="et-card__prices">
                      <div className="et-card__price">
                        <span className="et-card__price-value">{formatPrice(type.rental_price_day)}</span>
                        <span className="et-card__price-label">pro Tag</span>
                      </div>
                      {type.rental_price_week != null && type.rental_price_week > 0 && (
                        <div className="et-card__price">
                          <span className="et-card__price-value" style={{ fontSize: 'var(--font-size-base, 1rem)' }}>
                            {formatPrice(type.rental_price_week)}
                          </span>
                          <span className="et-card__price-label">pro Woche</span>
                        </div>
                      )}
                    </div>
                  </div>

                  <div className="et-card__footer">
                    <span className="et-card__count">
                      <Package size={14} />
                      {type.item_count ?? 0} Stueck
                    </span>
                    {type.status_summary && Object.keys(type.status_summary).length > 0 && (
                      <div className="et-card__status-dots">
                        {Object.entries(type.status_summary).map(([status, count]) => (
                          <span key={status} className="et-status-dot">
                            <span className={`et-status-dot__circle et-status-dot__circle--${status}`} />
                            {count}
                          </span>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Simple pagination */}
          {total > limit && (
            <div style={{ display: 'flex', justifyContent: 'center', gap: 'var(--spacing-3)', paddingTop: 'var(--spacing-4)' }}>
              <button
                className="btn btn--secondary btn--sm"
                disabled={page <= 1}
                onClick={() => setPage(p => p - 1)}
              >
                Zurueck
              </button>
              <span style={{ color: 'var(--color-text-secondary)', alignSelf: 'center', fontSize: 'var(--font-size-sm)' }}>
                Seite {page} von {Math.ceil(total / limit)}
              </span>
              <button
                className="btn btn--secondary btn--sm"
                disabled={page >= Math.ceil(total / limit)}
                onClick={() => setPage(p => p + 1)}
              >
                Weiter
              </button>
            </div>
          )}
        </>
      )}

      {/* Type Form Modal */}
      {formOpen && (
        <EquipmentTypeForm
          type={editingType}
          onClose={handleFormClose}
          onSuccess={handleFormSuccess}
        />
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="Equipment-Typ loeschen"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setDeleteTarget(null)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--danger"
              onClick={() => {
                if (deleteTarget) deleteType(deleteTarget.id)
              }}
              disabled={isDeleting}
            >
              {isDeleting ? 'Wird geloescht...' : 'Loeschen'}
            </button>
          </div>
        }
      >
        <p style={{ margin: 0, color: 'var(--color-text-primary)' }}>
          Moechten Sie <strong>{deleteTarget?.name}</strong> wirklich loeschen?
          {deleteTarget?.item_count && deleteTarget.item_count > 0
            ? ` Dieser Typ hat ${deleteTarget.item_count} Einzelartikel und kann erst geloescht werden, wenn alle Artikel entfernt wurden.`
            : ' Diese Aktion kann nicht rueckgaengig gemacht werden.'}
        </p>
      </Modal>
    </div>
  )
}

export default EquipmentTypesPage
