import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { invoiceApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Input } from '../../components/Form/Input'
import { Invoice, InvoiceStatus } from '../../types/invoice'
import '../Equipment/Equipment.module.scss'

const STATUS_TABS: Array<{ value: InvoiceStatus | ''; label: string }> = [
  { value: '', label: 'Alle' },
  { value: 'draft', label: 'Entwurf' },
  { value: 'sent', label: 'Gesendet' },
  { value: 'paid', label: 'Bezahlt' },
  { value: 'overdue', label: 'Überfällig' },
]

function InvoiceListPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState<InvoiceStatus | ''>('')
  const limit = 20

  const { data: invoiceData, isLoading, error } = useQuery({
    queryKey: ['invoice-list', page, searchQuery, selectedStatus],
    queryFn: () => invoiceApi.list(page, limit),
    staleTime: 1000 * 60 * 5,
  })

  const filteredData = (invoiceData?.data || []).filter((invoice: Invoice) => {
    const matchesSearch =
      !searchQuery ||
      invoice.number.toLowerCase().includes(searchQuery.toLowerCase()) ||
      invoice.client_name?.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = !selectedStatus || invoice.status === selectedStatus

    return matchesSearch && matchesStatus
  })

  const isOverdue = (invoice: Invoice) => {
    return invoice.status === 'overdue' || (
      new Date(invoice.due_date) < new Date() &&
      invoice.status !== 'paid'
    )
  }

  const columns: Column<Invoice>[] = [
    {
      key: 'number',
      label: 'Rechnungsnummer',
      sortable: true,
    },
    {
      key: 'client_name',
      label: 'Kunde',
    },
    {
      key: 'status',
      label: 'Status',
      render: (status: InvoiceStatus) => (
        <StatusBadge status={status} />
      ),
    },
    {
      key: 'total',
      label: 'Betrag',
      render: (total: number) => `€${total.toFixed(2)}`,
    },
    {
      key: 'due_date',
      label: 'Fälligkeitsdatum',
      render: (date: string, row: Invoice) => (
        <span style={{ color: isOverdue(row) ? 'var(--color-danger)' : 'inherit' }}>
          {new Date(date).toLocaleDateString('de-DE')}
        </span>
      ),
    },
  ]

  return (
    <div className="invoice-list-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Rechnungen</h1>
          <p className="page-subtitle">Verwalten Sie alle Rechnungen und Zahlungen</p>
        </div>
        <button
          className="btn btn--primary"
          onClick={() => navigate('/invoices/new')}
        >
          + Neue Rechnung
        </button>
      </div>

      <div style={{ marginBottom: 'var(--spacing-6)' }}>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)', marginBottom: 'var(--spacing-4)', flexWrap: 'wrap' }}>
          {STATUS_TABS.map((tab) => (
            <button
              key={tab.value}
              className={`btn ${selectedStatus === tab.value ? 'btn--primary' : 'btn--secondary'}`}
              onClick={() => {
                setSelectedStatus(tab.value)
                setPage(1)
              }}
            >
              {tab.label}
            </button>
          ))}
        </div>

        <Input
          type="text"
          placeholder="Nach Rechnungsnummer oder Kundenname suchen..."
          value={searchQuery}
          onChange={(e) => {
            setSearchQuery(e.target.value)
            setPage(1)
          }}
        />
      </div>

      {error && (
        <div className="error-message" role="alert">
          Fehler beim Laden der Rechnungen. Bitte versuchen Sie es später erneut.
        </div>
      )}

      <DataTable<Invoice>
        columns={columns}
        data={filteredData}
        rowKey="id"
        loading={isLoading}
        onRowClick={(invoice) => navigate(`/invoices/${invoice.id}`)}
        pagination={{
          page,
          total: invoiceData?.total || 0,
          limit,
          onPageChange: setPage,
        }}
      />
    </div>
  )
}

export default InvoiceListPage
