import { useState, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { Package, ArrowLeft, Plus } from 'lucide-react'
import { equipmentApi, equipmentTypeApi, categoryApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import EmptyState from '../../components/EmptyState/EmptyState'
import ErrorState from '../../components/ErrorState/ErrorState'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import { useNotificationStore } from '../../stores/notificationStore'
import { Equipment, EquipmentStatus, Category } from '../../types/equipment'
import { generateCSV, downloadCSV, formatDateForExport } from '../../utils/csvExport'
import './Equipment.scss'
import './EquipmentTypes.scss'

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: '', label: 'Alle Status' },
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

function EquipmentListPage() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [page, setPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')
  const [importMessage, setImportMessage] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<Equipment | null>(null)
  const [createItemsCount, setCreateItemsCount] = useState(1)
  const [showCreateItems, setShowCreateItems] = useState(false)
  const limit = 20

  // Type filter from URL params
  const typeId = searchParams.get('type_id') || ''
  const typeName = searchParams.get('type_name') || ''

  const { mutate: createItemsFromType, isPending: isCreatingItems } = useMutation({
    mutationFn: () => equipmentTypeApi.createItems(typeId, createItemsCount),
    onSuccess: () => {
      addNotification(`${createItemsCount} Einzelartikel erstellt`, 'success', { title: 'Erfolg', duration: 3000 })
      setShowCreateItems(false)
      setCreateItemsCount(1)
      queryClient.invalidateQueries({ queryKey: ['equipment-list'] })
    },
    onError: () => {
      addNotification('Fehler beim Erstellen der Einzelartikel', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const categoryOptions = (categories || []).map((cat: Category) => ({
    value: cat.id,
    label: cat.name,
  }))

  const { mutate: importCSV } = useMutation({
    mutationFn: async (file: File) => {
      return equipmentApi.importCSV(file)
    },
    onSuccess: () => {
      setImportMessage('Datei erfolgreich importiert!')
      if (fileInputRef.current) fileInputRef.current.value = ''
      queryClient.invalidateQueries({ queryKey: ['equipment-list'] })
      setTimeout(() => setImportMessage(''), 3000)
    },
    onError: () => {
      setImportMessage('Fehler beim Import. Bitte überprüfen Sie die CSV-Datei.')
      setTimeout(() => setImportMessage(''), 3000)
    },
  })

  const exportEquipmentCSV = () => {
    const EQUIPMENT_HEADERS = [
      { key: 'name', label: 'Name' },
      { key: 'sku', label: 'SKU' },
      { key: 'barcode', label: 'Barcode' },
      { key: 'category_name', label: 'Kategorie' },
      { key: 'status', label: 'Status' },
      { key: 'condition', label: 'Zustand' },
      { key: 'rental_price_day', label: 'Tagespreis' },
      { key: 'rental_price_week', label: 'Wochenpreis' },
      { key: 'purchase_price', label: 'Einkaufspreis' },
      { key: 'weight', label: 'Gewicht (kg)' },
      { key: 'description', label: 'Beschreibung' },
    ]

    const STATUS_LABELS: Record<string, string> = {
      available: 'Verfügbar',
      reserved: 'Reserviert',
      checked_out: 'Vermietet',
      rented: 'Vermietet',
      in_maintenance: 'In Wartung',
      maintenance: 'In Wartung',
      damaged: 'Beschädigt',
      retired: 'Ausgemustert',
      lost: 'Verloren',
    }

    const rows = items.map((eq) => {
      const cat = categories?.find((c: Category) => c.id === eq.category_id)
      return {
        name: eq.name || '',
        sku: eq.sku || '',
        barcode: eq.barcode || '',
        category_name: cat?.name || '',
        status: STATUS_LABELS[eq.status] || eq.status || '',
        condition: CONDITION_LABELS[eq.condition] || eq.condition || '',
        rental_price_day: eq.rental_price_day != null ? String(eq.rental_price_day) : '',
        rental_price_week: eq.rental_price_week != null ? String(eq.rental_price_week) : '',
        purchase_price: eq.purchase_price != null ? String(eq.purchase_price) : '',
        weight: eq.weight != null ? String(eq.weight) : '',
        description: eq.description || '',
      }
    })

    const csv = generateCSV(EQUIPMENT_HEADERS, rows)
    downloadCSV(csv, `equipment_export_${formatDateForExport()}.csv`)
    addNotification(`${rows.length} Einträge exportiert`, 'success', { title: 'Export erfolgreich', duration: 3000 })
  }

  const { mutate: deleteEquipment, isPending: isDeleting } = useMutation({
    mutationFn: async (id: string) => {
      return equipmentApi.delete(id)
    },
    onSuccess: () => {
      setDeleteTarget(null)
      addNotification('Ausrüstung erfolgreich gelöscht', 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
      queryClient.invalidateQueries({ queryKey: ['equipment-list'] })
    },
    onError: () => {
      addNotification('Fehler beim Löschen der Ausrüstung', 'error', {
        title: 'Fehler',
        duration: 5000,
      })
    },
  })

  // Use search endpoint when there's a search query, otherwise use list endpoint
  const offset = (page - 1) * limit

  const { data: equipmentData, isLoading, error } = useQuery({
    queryKey: ['equipment-list', page, searchQuery, selectedCategory, selectedStatus, typeId],
    queryFn: async () => {
      if (searchQuery) {
        return equipmentApi.search(searchQuery, limit, offset)
      }
      const params: Record<string, unknown> = { limit, offset }
      if (selectedCategory) params.category_id = selectedCategory
      if (selectedStatus) params.status = selectedStatus
      if (typeId) params.type_id = typeId
      return equipmentApi.list(params as { limit?: number; offset?: number; status?: string; category_id?: string })
    },
    staleTime: 1000 * 60 * 5,
  })

  const items: Equipment[] = equipmentData?.data || []
  const total: number = equipmentData?.total || 0

  const columns: Column<Equipment>[] = [
    {
      key: 'name',
      label: 'Name',
      sortable: true,
    },
    {
      key: 'sku',
      label: 'SKU',
    },
    {
      key: 'category_id',
      label: 'Kategorie',
      render: (categoryId: unknown) => {
        const cat = categories?.find((c: Category) => c.id === categoryId)
        return cat ? cat.name : (categoryId as string) || '—'
      },
    },
    {
      key: 'status',
      label: 'Status',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (status: any) => (
        <StatusBadge status={status as EquipmentStatus} />
      ),
    },
    {
      key: 'condition',
      label: 'Zustand',
      render: (condition: unknown) => {
        const condStr = condition as string
        return CONDITION_LABELS[condStr] || condStr || '—'
      },
    },
    {
      key: 'rental_price_day',
      label: 'Tagespreis',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (price: any) => (price ? `€${Number(price).toFixed(2)}` : '—'),
    },
    {
      key: 'actions',
      label: '',
      width: '50px',
      render: (_value: unknown, row: Equipment) => (
        <div className="table-actions">
          <button
            className="icon-btn"
            title="Löschen"
            onClick={(e) => {
              e.stopPropagation()
              setDeleteTarget(row)
            }}
          >
            🗑
          </button>
        </div>
      ),
    },
  ]

  const handleFileSelect = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (file) {
      importCSV(file)
    }
  }

  return (
    <div className="equipment-list-page">
      {/* Breadcrumb when filtered by type */}
      {typeId && (
        <div className="et-breadcrumb" style={{ marginBottom: 'var(--spacing-2)' }}>
          <Link to="/equipment" className="et-breadcrumb__link" style={{ display: 'inline-flex', alignItems: 'center', gap: '4px' }}>
            <ArrowLeft size={14} />
            Equipment
          </Link>
          <span className="et-breadcrumb__separator">/</span>
          <span className="et-breadcrumb__current">{typeName || 'Typ'}</span>
          <span className="et-breadcrumb__separator">/</span>
          <span className="et-breadcrumb__current">Einzelartikel</span>
        </div>
      )}

      <div className="page-header">
        <div>
          <h1 className="page-title">
            {typeId ? `${typeName || 'Typ'} — Einzelartikel` : 'Alle Einzelartikel'}
          </h1>
          <p className="page-subtitle">
            {typeId
              ? 'Einzelartikel dieses Typs verwalten'
              : 'Alle Equipment-Einzelartikel auf einen Blick'}
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          {typeId && (
            <button
              className="btn btn--secondary"
              onClick={() => navigate('/equipment')}
            >
              <ArrowLeft size={14} />
              Zurück
            </button>
          )}
          {typeId && (
            <button
              className="btn btn--primary"
              onClick={() => setShowCreateItems(true)}
            >
              <Plus size={14} />
              Einzelartikel erstellen
            </button>
          )}
          {!typeId && (
            <>
              <button
                className="btn btn--secondary"
                onClick={() => navigate('/equipment/timeline')}
                style={{ background: 'linear-gradient(135deg, rgba(0, 212, 255, 0.1) 0%, rgba(139, 92, 246, 0.1) 100%)', borderColor: 'rgba(0, 212, 255, 0.3)' }}
              >
                Zeitleiste öffnen
              </button>
              <button
                className="btn btn--secondary"
                onClick={() => navigate('/equipment/labels')}
              >
                Labels drucken
              </button>
              <button
                className="btn btn--secondary"
                onClick={exportEquipmentCSV}
                disabled={items.length === 0}
              >
                Exportieren
              </button>
              <button
                className="btn btn--secondary"
                onClick={() => navigate('/equipment/import')}
              >
                CSV importieren
              </button>
              <input
                ref={fileInputRef}
                type="file"
                accept=".csv"
                style={{ display: 'none' }}
                onChange={handleFileSelect}
              />
              <button
                className="btn btn--primary"
                onClick={() => navigate('/equipment')}
              >
                + Neuer Equipment-Typ
              </button>
            </>
          )}
        </div>
      </div>

      {importMessage && (
        <div style={{ padding: 'var(--spacing-3)', marginBottom: 'var(--spacing-4)', backgroundColor: importMessage.includes('erfolgreich') ? 'rgba(16, 185, 129, 0.1)' : 'rgba(239, 68, 68, 0.1)', border: `1px solid ${importMessage.includes('erfolgreich') ? 'var(--color-success)' : 'var(--color-danger)'}`, borderRadius: 'var(--radius-card)', color: importMessage.includes('erfolgreich') ? 'var(--color-success)' : 'var(--color-danger)' }}>
          {importMessage}
        </div>
      )}

      <div className="filters-section">
        <Input
          type="text"
          placeholder="Nach Name oder SKU suchen..."
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
          placeholder="Kategorie auswählen"
        />

        <Select
          options={STATUS_OPTIONS}
          value={selectedStatus}
          onChange={(e) => {
            setSelectedStatus(e.target.value)
            setPage(1)
          }}
        />
      </div>

      {error ? (
        <ErrorState
          variant="generic"
          title="Fehler beim Laden der Ausrüstung"
          description="Die Daten konnten nicht geladen werden. Bitte versuchen Sie es erneut."
          onRetry={() => window.location.reload()}
          compact
        />
      ) : isLoading ? (
        <SkeletonTable rows={6} columns={6} />
      ) : items.length === 0 ? (
        <EmptyState
          icon={Package}
          title="Noch keine Ausrüstung"
          description="Legen Sie Ihr erstes Equipment an, um loszulegen."
          action={{ label: 'Ausrüstung hinzufügen', href: '/equipment/new' }}
        />
      ) : (
        <DataTable<Equipment>
          columns={columns}
          data={items}
          rowKey="id"
          loading={false}
          onRowClick={(equipment) => navigate(`/equipment/${equipment.id}`)}
          pagination={{
            page,
            total,
            limit,
            onPageChange: setPage,
          }}
        />
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        isOpen={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="Ausrüstung löschen"
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
                if (deleteTarget) {
                  deleteEquipment(deleteTarget.id)
                }
              }}
              disabled={isDeleting}
            >
              {isDeleting ? 'Wird gelöscht...' : 'Löschen'}
            </button>
          </div>
        }
      >
        <p style={{ margin: 0, color: 'var(--color-text-primary)' }}>
          Möchten Sie <strong>{deleteTarget?.name}</strong> wirklich löschen? Diese Aktion kann nicht rückgängig gemacht werden.
        </p>
      </Modal>

      {/* Create Items from Type Modal */}
      <Modal
        isOpen={showCreateItems}
        onClose={() => setShowCreateItems(false)}
        title="Einzelartikel erstellen"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowCreateItems(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={() => createItemsFromType()}
              disabled={isCreatingItems || createItemsCount < 1}
            >
              {isCreatingItems ? 'Wird erstellt...' : `${createItemsCount} Artikel erstellen`}
            </button>
          </div>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-4)' }}>
          <p style={{ margin: 0, color: 'var(--color-text-primary)' }}>
            Wie viele Einzelartikel vom Typ <strong>{typeName}</strong> sollen erstellt werden?
          </p>
          <Input
            label="Anzahl"
            type="number"
            min="1"
            max="100"
            value={String(createItemsCount)}
            onChange={(e) => setCreateItemsCount(Math.max(1, parseInt(e.target.value) || 1))}
          />
        </div>
      </Modal>
    </div>
  )
}

export default EquipmentListPage
