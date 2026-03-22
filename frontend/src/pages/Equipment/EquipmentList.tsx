import { useState, useRef } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { equipmentApi, categoryApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { Equipment, EquipmentStatus, Category } from '../../types/equipment'
import './Equipment.module.scss'

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: '', label: 'Alle Status' },
  { value: 'available', label: 'Verfügbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Vermietet' },
  { value: 'in_maintenance', label: 'Wartung' },
  { value: 'damaged', label: 'Beschädigt' },
  { value: 'retired', label: 'Ausgemustert' },
]

function EquipmentListPage() {
  const navigate = useNavigate()
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [page, setPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState('')
  const [selectedStatus, setSelectedStatus] = useState('')
  const [importMessage, setImportMessage] = useState('')
  const limit = 20

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const categoryOptions = (categories || []).map((cat: Category) => ({
    value: cat.id,
    label: cat.name,
  }))

  const { mutate: importCSV, isPending: isImporting } = useMutation({
    mutationFn: async (file: File) => {
      return equipmentApi.importCSV(file)
    },
    onSuccess: () => {
      setImportMessage('Datei erfolgreich importiert!')
      if (fileInputRef.current) fileInputRef.current.value = ''
      setTimeout(() => setImportMessage(''), 3000)
    },
    onError: () => {
      setImportMessage('Fehler beim Import. Bitte überprüfen Sie die CSV-Datei.')
      setTimeout(() => setImportMessage(''), 3000)
    },
  })

  // Use search endpoint when there's a search query, otherwise use list endpoint
  const offset = (page - 1) * limit

  const { data: equipmentData, isLoading, error } = useQuery({
    queryKey: ['equipment-list', page, searchQuery, selectedCategory, selectedStatus],
    queryFn: async () => {
      if (searchQuery) {
        return equipmentApi.search(searchQuery, limit, offset)
      }
      const params: Record<string, unknown> = { limit, offset }
      if (selectedCategory) params.category_id = selectedCategory
      if (selectedStatus) params.status = selectedStatus
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
    },
    {
      key: 'rental_price_day',
      label: 'Tagespreis',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (price: any) => (price ? `€${Number(price).toFixed(2)}` : '—'),
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
      <div className="page-header">
        <div>
          <h1 className="page-title">Ausrüstungsverwaltung</h1>
          <p className="page-subtitle">Verwalten Sie Ihre Ausrüstung und verfügbaren Ressourcen</p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <button
            className="btn btn--secondary"
            onClick={() => navigate('/equipment/labels')}
          >
            Labels drucken
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => fileInputRef.current?.click()}
            disabled={isImporting}
          >
            {isImporting ? 'Wird importiert...' : '📥 CSV importieren'}
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
            onClick={() => navigate('/equipment/new')}
          >
            + Neue Ausrüstung
          </button>
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

      {error && (
        <div className="error-message" role="alert">
          Fehler beim Laden der Ausrüstung. Bitte versuchen Sie es später erneut.
        </div>
      )}

      <DataTable<Equipment>
        columns={columns}
        data={items}
        rowKey="id"
        loading={isLoading}
        onRowClick={(equipment) => navigate(`/equipment/${equipment.id}`)}
        pagination={{
          page,
          total,
          limit,
          onPageChange: setPage,
        }}
      />
    </div>
  )
}

export default EquipmentListPage
