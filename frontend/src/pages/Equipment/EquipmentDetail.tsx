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
import './Equipment.scss'

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: 'available', label: 'Verfügbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Vermietet' },
  { value: 'in_maintenance', label: 'In Wartung' },
  { value: 'damaged', label: 'Beschädigt' },
  { value: 'retired', label: 'Ausgemustert' },
]

const CONDITION_LABELS: Record<string, string> = {
  new: 'Neu',
  excellent: 'Sehr gut',
  good: 'Gut',
  fair: 'Befriedigend',
  poor: 'Mangelhaft',
  defective: 'Defekt',
}

function EquipmentDetailPage() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [showStatusModal, setShowStatusModal] = useState(false)
  const [showDeleteModal, setShowDeleteModal] = useState(false)
  const [showQrModal, setShowQrModal] = useState(false)
  const [qrBlobUrl, setQrBlobUrl] = useState<string | null>(null)
  const [newStatus, setNewStatus] = useState<EquipmentStatus | ''>('')
  const [showRfidModal, setShowRfidModal] = useState(false)
  const [rfidTagInput, setRfidTagInput] = useState('')
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

  const { mutate: deleteEquipment, isPending: isDeletePending } = useMutation({
    mutationFn: async () => {
      return equipmentApi.delete(id!)
    },
    onSuccess: () => {
      addNotification('Ausrüstung erfolgreich gelöscht', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      queryClient.invalidateQueries({ queryKey: ['equipment-list'] })
      navigate('/equipment')
    },
    onError: () => {
      addNotification('Fehler beim Löschen der Ausrüstung', 'error', {
        title: 'Fehler',
        duration: 5000,
      })
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

  const { mutate: assignRfid, isPending: isAssigningRfid } = useMutation({
    mutationFn: async () => {
      return equipmentApi.assignRfidTag(id!, rfidTagInput)
    },
    onSuccess: () => {
      setShowRfidModal(false)
      setRfidTagInput('')
      addNotification('RFID Tag erfolgreich zugeordnet', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      queryClient.invalidateQueries({ queryKey: ['equipment', id] })
    },
    onError: () => {
      addNotification('Fehler beim Zuordnen des RFID Tags', 'error', {
        title: 'Fehler',
        duration: 5000,
      })
    },
  })

  const handleShowQRCode = async () => {
    try {
      const blob = await equipmentApi.getQRCode(id!)
      const url = window.URL.createObjectURL(blob)
      setQrBlobUrl(url)
      setShowQrModal(true)
    } catch (err) {
      console.error('Failed to load QR code:', err)
      addNotification('QR-Code konnte nicht geladen werden', 'error', {
        title: 'Fehler',
        duration: 5000,
      })
    }
  }

  const handleDownloadQRCode = () => {
    if (!qrBlobUrl) return
    const link = document.createElement('a')
    link.href = qrBlobUrl
    link.download = `${equipment?.name}-qr-code.png`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  }

  const getCategoryName = (categoryId: string) => {
    const cat = categories?.find((c: Category) => c.id === categoryId)
    return cat ? cat.name : categoryId || '—'
  }

  if (isLoading) {
    return (
      <div className="equipment-detail-page">
        <div className="detail-grid">
          <div className="detail-card">
            <div className="detail-card__content">
              {[1, 2, 3, 4, 5].map((i) => (
                <div key={i} className="detail-card__row">
                  <span className="detail-card__row-label" style={{ background: 'var(--color-bg-tertiary)', borderRadius: 'var(--radius-sm)', height: '1rem', width: '80px' }}>&nbsp;</span>
                  <span className="detail-card__row-value" style={{ background: 'var(--color-bg-tertiary)', borderRadius: 'var(--radius-sm)', height: '1rem', width: '150px' }}>&nbsp;</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="equipment-detail-page">
        <div className="error-message" role="alert">
          Fehler beim Laden der Ausrüstung. Bitte versuchen Sie es später erneut.
        </div>
        <button className="btn btn--secondary" onClick={() => navigate('/equipment')}>
          Zurück zur Liste
        </button>
      </div>
    )
  }

  if (!equipment) {
    return (
      <div className="equipment-detail-page">
        <div className="error-message" role="alert">
          Ausrüstung nicht gefunden
        </div>
        <button className="btn btn--secondary" onClick={() => navigate('/equipment')}>
          Zurück zur Liste
        </button>
      </div>
    )
  }

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
            onClick={handleShowQRCode}
          >
            QR-Code
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => {
              setRfidTagInput(equipment?.rfid_tag || '')
              setShowRfidModal(true)
            }}
          >
            RFID Tag zuordnen
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => navigate(`/equipment/${id}/edit`)}
          >
            Bearbeiten
          </button>
          <button
            className="btn btn--danger"
            onClick={() => setShowDeleteModal(true)}
          >
            Löschen
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
              <span className="detail-card__row-value">
                {CONDITION_LABELS[equipment.condition] || equipment.condition || '—'}
              </span>
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
              <span className="detail-card__row-label">RFID Tag</span>
              <span className="detail-card__row-value" style={{ fontFamily: equipment.rfid_tag ? 'monospace' : 'inherit' }}>
                {equipment.rfid_tag || '—'}
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

            {equipment.dimensions && (equipment.dimensions.length || equipment.dimensions.width || equipment.dimensions.height) && (
              <div className="detail-card__row">
                <span className="detail-card__row-label">Abmessungen</span>
                <span className="detail-card__row-value">
                  {[equipment.dimensions.length, equipment.dimensions.width, equipment.dimensions.height]
                    .filter(Boolean)
                    .join(' x ')}{' '}
                  {equipment.dimensions.unit || 'cm'}
                </span>
              </div>
            )}

            {equipment.purchase_price != null && equipment.purchase_price > 0 && (
              <div className="detail-card__row">
                <span className="detail-card__row-label">Einkaufspreis</span>
                <span className="detail-card__row-value">
                  {`€${equipment.purchase_price.toFixed(2)}`}
                </span>
              </div>
            )}
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
                <div style={{ padding: 'var(--spacing-3)', backgroundColor: 'rgba(0, 212, 255, 0.08)', borderRadius: 'var(--radius-md)', borderLeft: '3px solid var(--color-primary)' }}>
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
              {equipment.weight != null && equipment.weight > 0 && (
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
                  <span className="detail-card__row-value" style={{ display: 'flex', gap: 'var(--spacing-2)', flexWrap: 'wrap' }}>
                    {equipment.tags.map((tag: string) => (
                      <span
                        key={tag}
                        style={{
                          padding: '0.15rem 0.5rem',
                          backgroundColor: 'rgba(0, 212, 255, 0.1)',
                          border: '1px solid rgba(0, 212, 255, 0.2)',
                          borderRadius: 'var(--radius-full)',
                          fontSize: 'var(--font-size-xs)',
                          color: 'var(--color-primary)',
                        }}
                      >
                        {tag}
                      </span>
                    ))}
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

      {/* Status Change Modal */}
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

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Ausrüstung löschen"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowDeleteModal(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--danger"
              onClick={() => deleteEquipment()}
              disabled={isDeletePending}
            >
              {isDeletePending ? 'Wird gelöscht...' : 'Endgültig löschen'}
            </button>
          </div>
        }
      >
        <p style={{ margin: 0, color: 'var(--color-text-primary)' }}>
          Möchten Sie <strong>{equipment.name}</strong> wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.
        </p>
      </Modal>

      {/* QR Code Modal */}
      <Modal
        isOpen={showQrModal}
        onClose={() => {
          setShowQrModal(false)
          if (qrBlobUrl) {
            window.URL.revokeObjectURL(qrBlobUrl)
            setQrBlobUrl(null)
          }
        }}
        title="QR-Code"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => {
                setShowQrModal(false)
                if (qrBlobUrl) {
                  window.URL.revokeObjectURL(qrBlobUrl)
                  setQrBlobUrl(null)
                }
              }}
            >
              Schließen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleDownloadQRCode}
            >
              Herunterladen
            </button>
          </div>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 'var(--spacing-4)' }}>
          {qrBlobUrl && (
            <img
              src={qrBlobUrl}
              alt={`QR-Code für ${equipment.name}`}
              style={{
                width: '200px',
                height: '200px',
                borderRadius: 'var(--radius-md)',
                background: 'white',
                padding: 'var(--spacing-3)',
              }}
            />
          )}
          <p style={{ margin: 0, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', textAlign: 'center' }}>
            {equipment.name} ({equipment.sku})
          </p>
        </div>
      </Modal>

      {/* RFID Modal */}
      <Modal
        isOpen={showRfidModal}
        onClose={() => setShowRfidModal(false)}
        title="RFID Tag zuordnen"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowRfidModal(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={() => {
                if (rfidTagInput.trim()) {
                  assignRfid()
                }
              }}
              disabled={!rfidTagInput.trim() || isAssigningRfid}
            >
              {isAssigningRfid ? 'Wird zugeordnet...' : 'Zuordnen'}
            </button>
          </div>
        }
      >
        <Input
          label="RFID Tag (Hex)"
          value={rfidTagInput}
          onChange={(e) => setRfidTagInput(e.target.value)}
          placeholder="z.B. E28011606000020..."
        />
        <p style={{ margin: 'var(--spacing-2) 0 0 0', fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)' }}>
          Scannen Sie den RFID-Tag mit dem CF-H906 UHF PDA oder geben Sie die Tag-ID manuell ein.
        </p>
      </Modal>
    </div>
  )
}

export default EquipmentDetailPage
