import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { equipmentApi } from '../../services/api'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { Select } from '../../components/Form/Select'
import { Equipment, EquipmentStatus } from '../../types/equipment'
import './Equipment.module.scss'

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'available', label: 'Verfügbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Vermietet' },
  { value: 'maintenance', label: 'Wartung' },
  { value: 'retired', label: 'Ausgemustert' },
]

function EquipmentDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [showStatusModal, setShowStatusModal] = useState(false)
  const [newStatus, setNewStatus] = useState<EquipmentStatus | ''>('')

  const {
    data: equipment,
    isLoading,
    error,
    refetch,
  } = useQuery({
    queryKey: ['equipment', id],
    queryFn: () => equipmentApi.getById(id!),
    enabled: !!id,
  })

  const { mutate: changeStatus, isPending } = useMutation({
    mutationFn: async (status: string) => {
      return equipmentApi.update(id!, { status: status as EquipmentStatus })
    },
    onSuccess: () => {
      setShowStatusModal(false)
      refetch()
    },
  })

  if (isLoading) return <div className="equipment-detail-page">Wird geladen...</div>
  if (error) return <div className="error-message">Fehler beim Laden der Ausrüstung</div>
  if (!equipment) return <div className="error-message">Ausrüstung nicht gefunden</div>

  return (
    <div className="equipment-detail-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">{equipment.name}</h1>
          <p className="page-subtitle">SKU: {equipment.sku}</p>
        </div>
        <div>
          <button
            className="btn btn--secondary"
            onClick={() => navigate(`/equipment/${id}/edit`)}
          >
            Bearbeiten
          </button>
          <button
            className="btn btn--primary"
            onClick={() => setShowStatusModal(true)}
          >
            Status ändern
          </button>
        </div>
      </div>

      <div className="detail-grid">
        <div className="detail-card">
          <h2 className="detail-card__title">Ausrüstungsdetails</h2>
          <div className="detail-card__content">
            <div className="detail-card__row">
              <span className="detail-card__row-label">Status</span>
              <span className="detail-card__row-value">
                <StatusBadge status={equipment.status} />
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Kategorie</span>
              <span className="detail-card__row-value">{equipment.category}</span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Standort</span>
              <span className="detail-card__row-value">
                {equipment.location || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Barcode</span>
              <span className="detail-card__row-value">
                {equipment.barcode || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Beschreibung</span>
              <span className="detail-card__row-value">
                {equipment.description || '—'}
              </span>
            </div>
          </div>
        </div>

        <div>
          <div className="detail-card">
            <h2 className="detail-card__title">Preisgestaltung</h2>
            <div className="detail-card__content">
              <div className="status-box">
                <p className="status-box__label">Tagespreis</p>
                <p className="status-box__value">
                  €{equipment.price_daily?.toFixed(2) || '—'}
                </p>
              </div>

              <div className="status-box">
                <p className="status-box__label">Wochenpreis</p>
                <p className="status-box__value">
                  €{equipment.price_weekly?.toFixed(2) || '—'}
                </p>
              </div>

              <div className="status-box">
                <p className="status-box__label">Monatspreis</p>
                <p className="status-box__value">
                  €{equipment.price_monthly?.toFixed(2) || '—'}
                </p>
              </div>
            </div>
          </div>

          <div className="detail-card" style={{ marginTop: 'var(--spacing-4)' }}>
            <h2 className="detail-card__title">Sonstiges</h2>
            <div className="detail-card__content">
              <div className="detail-card__row">
                <span className="detail-card__row-label">Menge</span>
                <span className="detail-card__row-value">
                  {equipment.quantity || 1}
                </span>
              </div>

              <div className="detail-card__row">
                <span className="detail-card__row-label">Erstellt am</span>
                <span className="detail-card__row-value">
                  {new Date(equipment.created_at).toLocaleDateString('de-DE')}
                </span>
              </div>

              <div className="detail-card__row">
                <span className="detail-card__row-label">Aktualisiert am</span>
                <span className="detail-card__row-value">
                  {new Date(equipment.updated_at).toLocaleDateString('de-DE')}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <Modal
        isOpen={showStatusModal}
        onClose={() => setShowStatusModal(false)}
        title="Status ändern"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowStatusModal(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={() => {
                if (newStatus) {
                  changeStatus(newStatus)
                }
              }}
              disabled={!newStatus || isPending}
            >
              {isPending ? 'Wird geändert...' : 'Speichern'}
            </button>
          </div>
        }
      >
        <Select
          label="Neuer Status"
          options={STATUS_OPTIONS}
          value={newStatus}
          onChange={(e) => setNewStatus(e.target.value as EquipmentStatus | '')}
          placeholder="Status auswählen"
        />
      </Modal>
    </div>
  )
}

export default EquipmentDetailPage
