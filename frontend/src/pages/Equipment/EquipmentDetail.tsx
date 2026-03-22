import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { equipmentApi, categoryApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { Select } from '../../components/Form/Select'
import { Input } from '../../components/Form/Input'
import { EquipmentStatus, Category, PriceResult } from '../../types/equipment'
import './Equipment.module.scss'

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'available', label: 'Verfügbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Vermietet' },
  { value: 'in_maintenance', label: 'Wartung' },
  { value: 'damaged', label: 'Beschädigt' },
  { value: 'retired', label: 'Ausgemustert' },
]

function EquipmentDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [showStatusModal, setShowStatusModal] = useState(false)
  const [newStatus, setNewStatus] = useState<EquipmentStatus | ''>('')
  const [priceCalcDays, setPriceCalcDays] = useState('1')
  const [priceCalcDiscount, setPriceCalcDiscount] = useState('0')
  const [calculatedPrice, setCalculatedPrice] = useState<PriceResult | null>(null)

  const {
    data: equipment,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['equipment', id],
    queryFn: () => equipmentApi.getById(id!),
    enabled: !!id,
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const { mutate: changeStatus, isPending } = useMutation<unknown, unknown, string, { previous: unknown }>({
    mutationFn: async (status) => {
      return equipmentApi.changeStatus(id!, status)
    },
    onMutate: async (status) => {
      await queryClient.cancelQueries({ queryKey: ['equipment', id] })
      const previous = queryClient.getQueryData(['equipment', id])
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      queryClient.setQueryData(['equipment', id], (old: Record<string, unknown>) => ({
        ...(old || {}),
        status: status as EquipmentStatus,
      }))
      return { previous }
    },
    onSuccess: () => {
      setShowStatusModal(false)
      addNotification('Status erfolgreich geändert', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
    },
    onError: (_err, _vars, context) => {
      if (context?.previous) queryClient.setQueryData(['equipment', id], context.previous)
      addNotification(
        'Fehler beim Ändern des Status. Bitte versuchen Sie es später erneut.',
        'error',
        {
          title: 'Fehler',
          duration: 5000,
        }
      )
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment', id] })
    },
  })

  const { mutate: calculatePrice } = useMutation({
    mutationFn: async () => {
      // Backend expects discount as 0-1 range (e.g. 0.1 for 10%)
      const discountFraction = parseInt(priceCalcDiscount) / 100
      return equipmentApi.getPrice(id!, parseInt(priceCalcDays), discountFraction)
    },
    onSuccess: (data: PriceResult) => {
      setCalculatedPrice(data)
    },
  })

  const handleDownloadQRCode = async () => {
    try {
      const blob = await equipmentApi.getQRCode(id!)
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${equipment?.name}-qr-code.png`
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
    } catch (err) {
      console.error('Failed to download QR code:', err)
    }
  }

  const getCategoryName = (categoryId: string) => {
    const cat = categories?.find((c: Category) => c.id === categoryId)
    return cat ? cat.name : categoryId || '—'
  }

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
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <button
            className="btn btn--secondary"
            onClick={handleDownloadQRCode}
          >
            QR-Code
          </button>
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
              <span className="detail-card__row-label">Zustand</span>
              <span className="detail-card__row-value">{equipment.condition || '—'}</span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Kategorie</span>
              <span className="detail-card__row-value">{getCategoryName(equipment.category_id)}</span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Standort-ID</span>
              <span className="detail-card__row-value">
                {equipment.location_id || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Barcode</span>
              <span className="detail-card__row-value">
                {equipment.barcode || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Seriennummer</span>
              <span className="detail-card__row-value">
                {equipment.serial_number || '—'}
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
                  {equipment.rental_price_day ? `€${equipment.rental_price_day.toFixed(2)}` : '—'}
                </p>
              </div>

              <div className="status-box">
                <p className="status-box__label">Wochenpreis</p>
                <p className="status-box__value">
                  {equipment.rental_price_week ? `€${equipment.rental_price_week.toFixed(2)}` : '—'}
                </p>
              </div>
            </div>
          </div>

          <div className="detail-card" style={{ marginTop: 'var(--spacing-4)' }}>
            <h2 className="detail-card__title">Preisrechner</h2>
            <div className="detail-card__content" style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
              <Input
                type="number"
                label="Anzahl Tage"
                value={priceCalcDays}
                onChange={(e) => setPriceCalcDays(e.target.value)}
                min="1"
              />
              <Input
                type="number"
                label="Rabatt (%)"
                value={priceCalcDiscount}
                onChange={(e) => setPriceCalcDiscount(e.target.value)}
                min="0"
                max="100"
              />
              <button
                className="btn btn--primary"
                onClick={() => calculatePrice()}
                style={{ width: '100%' }}
              >
                Berechnen
              </button>
              {calculatedPrice && (
                <div style={{ padding: 'var(--spacing-3)', backgroundColor: 'var(--color-primary-50)', borderRadius: 'var(--radius-md)', borderLeft: '3px solid var(--color-primary)' }}>
                  <p style={{ margin: '0 0 var(--spacing-1) 0', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
                    Berechneter Preis (inkl. MwSt.)
                  </p>
                  <p style={{ margin: 0, fontSize: 'var(--font-size-2xl)', fontWeight: 'var(--font-weight-bold)', color: 'var(--color-primary)' }}>
                    €{calculatedPrice.total.toFixed(2)}
                  </p>
                  <p style={{ margin: 'var(--spacing-1) 0 0 0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                    ({calculatedPrice.days} Tage × €{calculatedPrice.daily_rate.toFixed(2)} = €{calculatedPrice.base_price.toFixed(2)})
                  </p>
                  {calculatedPrice.volume_discount > 0 && (
                    <p style={{ margin: 'var(--spacing-1) 0 0 0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                      Mengenrabatt: -€{calculatedPrice.volume_discount.toFixed(2)} ({(calculatedPrice.volume_discount_percent * 100).toFixed(0)}%)
                    </p>
                  )}
                  {calculatedPrice.custom_discount > 0 && (
                    <p style={{ margin: 'var(--spacing-1) 0 0 0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                      Kundenrabatt: -€{calculatedPrice.custom_discount.toFixed(2)}
                    </p>
                  )}
                  <p style={{ margin: 'var(--spacing-1) 0 0 0', fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)' }}>
                    Netto: €{calculatedPrice.subtotal.toFixed(2)} | MwSt. ({(calculatedPrice.vat_percent * 100).toFixed(0)}%): €{calculatedPrice.vat.toFixed(2)}
                  </p>
                </div>
              )}
            </div>
          </div>

          <div className="detail-card" style={{ marginTop: 'var(--spacing-4)' }}>
            <h2 className="detail-card__title">Sonstiges</h2>
            <div className="detail-card__content">
              {equipment.weight > 0 && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Gewicht</span>
                  <span className="detail-card__row-value">
                    {equipment.weight} kg
                  </span>
                </div>
              )}

              {equipment.tags && equipment.tags.length > 0 && (
                <div className="detail-card__row">
                  <span className="detail-card__row-label">Tags</span>
                  <span className="detail-card__row-value">
                    {equipment.tags.join(', ')}
                  </span>
                </div>
              )}

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
