import { useState, useMemo, useRef } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { equipmentApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Input } from '../../components/Form/Input'
import { Select } from '../../components/Form/Select'
import { Equipment, EquipmentStatus } from '../../types/equipment'
import './Equipment.module.scss'

const CATEGORIES = [
  { value: 'lighting', label: 'Beleuchtung' },
  { value: 'sound', label: 'Ton' },
  { value: 'staging', label: 'Bühne' },
  { value: 'projection', label: 'Projektion' },
  { value: 'decoration', label: 'Dekoration' },
]

const STATUS_OPTIONS: Array<{ value: string; label: string }> = [
  { value: '', label: 'Alle Status' },
  { value: 'available', label: 'Verfügbar' },
  { value: 'reserved', label: 'Reserviert' },
  { value: 'checked_out', label: 'Vermietet' },
  { value: 'maintenance', label: 'Wartung' },
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

  const { data: equipmentData, isLoading: _isLoading, error } = useQuery({
    queryKey: ['equipment-list', page, searchQuery, selectedCategory, selectedStatus],
    queryFn: async () => {
      const params: Record<string, unknown> = { page, limit }
      if (searchQuery) params.search = searchQuery
      if (selectedCategory) params.category = selectedCategory
      if (selectedStatus) params.status = selectedStatus
      return equipmentApi.list(page, limit)
    },
    staleTime: 1000 * 60 * 5,
  })

  const _filteredData = useMemo(() => {
    if (!equipmentData?.data) return []
    return equipmentData.data.filter((item: Equipment) => {
      const matchesSearch =
        !searchQuery ||
        item.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.sku.toLowerCase().includes(searchQuery.toLowerCase())

      const matchesCategory = !selectedCategory || item.category === selectedCategory
      const matchesStatus = !selectedStatus || item.status === selectedStatus

      return matchesSearch && matchesCategory && matchesStatus
    })
  }, [equipmentData, searchQuery, selectedCategory, selectedStatus])

  const _columns: Column<Equipment>[] = [
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
      key: 'category',
      label: 'Kategorie',
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
      key: 'location',
      label: 'Standort',
    },
    {
      key: 'price_daily',
      label: 'Tagespreis',
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      render: (price: any) => (price ? `€${price.toFixed(2)}` : '—'),
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
          options={CATEGORIES}
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
        columns={_columns}
        data={_filteredData}
        rowKey="id"
        loading={_isLoading}
        onRowClick={(equipment) => navigate(`/equipment/${equipment.id}`)}
        pagination={{
          page,
          total: equipmentData?.total || 0,
          limit,
          onPageChange: setPage,
        }}
      />
    </div>
  )
}

export default EquipmentListPage
