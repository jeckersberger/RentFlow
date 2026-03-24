import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { FileText } from 'lucide-react'
import { invoiceApi, configApi, projectApi } from '../../services/api'
import { DataTable, Column } from '../../components/DataTable/DataTable'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Input } from '../../components/Form/Input'
import { Modal } from '../../components/Modal/Modal'
import EmptyState from '../../components/EmptyState/EmptyState'
import ErrorState from '../../components/ErrorState/ErrorState'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import { Invoice, InvoiceStatus } from '../../types/invoice'
import { generateCSV, downloadCSV, formatDateForExport } from '../../utils/csvExport'
import '../Equipment/Equipment.scss'

const formatCurrency = (value: number) =>
  new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(value)

const STATUS_TABS: Array<{ value: InvoiceStatus | ''; label: string }> = [
  { value: '', label: 'Alle' },
  { value: 'draft', label: 'Entwurf' },
  { value: 'sent', label: 'Gesendet' },
  { value: 'paid', label: 'Bezahlt' },
  { value: 'overdue', label: 'Überfällig' },
  { value: 'cancelled', label: 'Storniert' },
  { value: 'partial', label: 'Teilweise bezahlt' },
]

function InvoiceListPage() {
  const navigate = useNavigate()
  const [page, setPage] = useState(1)
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedStatus, setSelectedStatus] = useState<InvoiceStatus | ''>('')
  const [showProjectModal, setShowProjectModal] = useState(false)
  const [selectedProjectId, setSelectedProjectId] = useState('')
  const limit = 20

  // Load Kleinunternehmer config
  const { data: kuConfig } = useQuery({
    queryKey: ['config', 'finance.kleinunternehmer'],
    queryFn: () => configApi.get('finance.kleinunternehmer'),
  })
  const isKleinunternehmer = kuConfig?.kleinunternehmer ?? kuConfig?.value ?? false

  // Load projects for the "Aus Projekt erstellen" modal
  const { data: projectsData } = useQuery({
    queryKey: ['projects-for-invoice'],
    queryFn: () => projectApi.list(1, 100),
    enabled: showProjectModal,
  })
  const projects = projectsData?.items || projectsData?.data || []

  const { mutate: createFromProject, isPending: isCreating } = useMutation({
    mutationFn: async (projectId: string) => {
      return invoiceApi.createFromProject(projectId)
    },
    onSuccess: (data) => {
      setShowProjectModal(false)
      setSelectedProjectId('')
      if (data?.id) {
        navigate(`/invoices/${data.id}`)
      } else {
        navigate(`/invoices/new?project=${selectedProjectId}`)
      }
    },
    onError: () => {
      // Fallback: navigate to form with project pre-filled
      setShowProjectModal(false)
      navigate(`/invoices/new?project=${selectedProjectId}`)
    },
  })

  const { data: invoiceData, isLoading, error } = useQuery({
    queryKey: ['invoice-list', page, limit],
    queryFn: () => invoiceApi.list(page, limit),
    staleTime: 1000 * 60 * 5,
  })

  const filteredData = (invoiceData?.data || []).filter((invoice: Invoice) => {
    const matchesSearch =
      !searchQuery ||
      invoice.number?.toLowerCase().includes(searchQuery.toLowerCase()) ||
      invoice.client_name?.toLowerCase().includes(searchQuery.toLowerCase())

    const matchesStatus = !selectedStatus || invoice.status === selectedStatus

    return matchesSearch && matchesStatus
  })

  const exportInvoicesCSV = () => {
    const INVOICE_HEADERS = [
      { key: 'number', label: 'Rechnungsnummer' },
      { key: 'client_name', label: 'Kunde' },
      { key: 'status', label: 'Status' },
      { key: 'issue_date', label: 'Rechnungsdatum' },
      { key: 'due_date', label: 'Fälligkeitsdatum' },
      { key: 'subtotal', label: 'Nettobetrag' },
      { key: 'tax_total', label: 'MwSt' },
      { key: 'total', label: 'Bruttobetrag' },
      { key: 'notes', label: 'Notizen' },
    ]
    const STATUS_LABELS: Record<string, string> = {
      draft: 'Entwurf', sent: 'Gesendet', paid: 'Bezahlt',
      overdue: 'Überfällig', cancelled: 'Storniert', partial: 'Teilweise bezahlt',
    }
    const rows = filteredData.map((inv: Invoice) => ({
      number: inv.number || '',
      client_name: inv.client_name || '',
      status: STATUS_LABELS[inv.status] || inv.status || '',
      issue_date: inv.issue_date ? new Date(inv.issue_date).toLocaleDateString('de-DE') : '',
      due_date: inv.due_date ? new Date(inv.due_date).toLocaleDateString('de-DE') : '',
      subtotal: inv.subtotal != null ? String(inv.subtotal) : '',
      tax_total: inv.tax_total != null ? String(inv.tax_total) : '',
      total: inv.total != null ? String(inv.total) : '',
      notes: inv.notes || '',
    }))
    const csv = generateCSV(INVOICE_HEADERS, rows)
    downloadCSV(csv, `rechnungen_export_${formatDateForExport()}.csv`)
  }

  const exportDatevCSV = () => {
    // DATEV SKR03 CSV Export
    // Columns: Umsatz (Betrag);Soll/Haben;Konto;Gegenkonto;Belegdatum;Buchungstext;Belegnummer
    const BOM = '\uFEFF'
    const header = 'Umsatz (Betrag);Soll/Haben;Konto;Gegenkonto;Belegdatum;Buchungstext;Belegnummer'

    const rows = filteredData
      .filter((inv: Invoice) => inv.status !== 'draft' && inv.status !== 'cancelled')
      .flatMap((inv: Invoice) => {
        const issueDate = inv.issue_date ? new Date(inv.issue_date) : new Date()
        const belegdatum = `${issueDate.getDate().toString().padStart(2, '0')}${(issueDate.getMonth() + 1).toString().padStart(2, '0')}`
        const buchungstext = `RE ${inv.number || ''} ${inv.client_name || ''}`.trim()
        const belegnummer = inv.number || ''
        const total = inv.total ?? 0
        const taxTotal = inv.tax_total ?? 0
        const netto = inv.subtotal ?? (total - taxTotal)

        const lines: string[] = []

        // Booking line 1: Forderungen (1400) an Erlöse 19% (8400) - Nettobetrag
        if (netto > 0) {
          lines.push(
            `${netto.toFixed(2).replace('.', ',')};S;1400;8400;${belegdatum};${buchungstext};${belegnummer}`
          )
        }

        // Booking line 2: Forderungen (1400) an Umsatzsteuer 19% (1776) - MwSt-Betrag
        if (taxTotal > 0) {
          lines.push(
            `${taxTotal.toFixed(2).replace('.', ',')};S;1400;1776;${belegdatum};${buchungstext} USt;${belegnummer}`
          )
        }

        return lines
      })

    const csv = BOM + [header, ...rows].join('\r\n')
    const now = new Date()
    const filename = `datev-export-${now.getFullYear()}-${(now.getMonth() + 1).toString().padStart(2, '0')}.csv`
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
  }

  const isOverdue = (invoice: Invoice) => {
    return invoice.status === 'overdue' || (
      new Date(invoice.due_date) < new Date() &&
      invoice.status !== 'paid' &&
      invoice.status !== 'cancelled' &&
      invoice.status !== 'draft'
    )
  }

  const columns: Column<Invoice>[] = [
    {
      key: 'number',
      label: 'Nummer',
      sortable: true,
      render: (value: string) => (
        <span style={{ fontWeight: 600, color: 'var(--color-primary)' }}>{value}</span>
      ),
    },
    {
      key: 'client_name',
      label: 'Kunde',
      render: (value: string) => value || '—',
    },
    {
      key: 'issue_date',
      label: 'Datum',
      render: (date: string) => date ? new Date(date).toLocaleDateString('de-DE') : '—',
    },
    {
      key: 'due_date',
      label: 'Fällig am',
      render: (date: string, row: Invoice) => (
        <span style={{ color: isOverdue(row) ? 'var(--color-danger)' : 'inherit', fontWeight: isOverdue(row) ? 600 : 400 }}>
          {date ? new Date(date).toLocaleDateString('de-DE') : '—'}
        </span>
      ),
    },
    {
      key: 'subtotal',
      label: 'Betrag (netto)',
      render: (value: number) => (
        <span style={{ fontVariantNumeric: 'tabular-nums' }}>
          {formatCurrency(value ?? 0)}
        </span>
      ),
    },
    {
      key: 'total',
      label: 'Betrag (brutto)',
      render: (value: number) => (
        <span style={{ fontWeight: 600, fontVariantNumeric: 'tabular-nums' }}>
          {formatCurrency(value ?? 0)}
        </span>
      ),
    },
    {
      key: 'status',
      label: 'Status',
      render: (_status: string, row: Invoice) => (
        <StatusBadge
          status={isOverdue(row) && row.status !== 'overdue' ? 'overdue' : row.status}
        />
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
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <button
            className="btn btn--secondary"
            onClick={exportInvoicesCSV}
            disabled={filteredData.length === 0}
          >
            Exportieren
          </button>
          <button
            className="btn btn--secondary"
            onClick={exportDatevCSV}
            disabled={filteredData.length === 0}
          >
            DATEV Export
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => setShowProjectModal(true)}
          >
            Aus Projekt erstellen
          </button>
          <button
            className="btn btn--primary"
            onClick={() => navigate('/invoices/new')}
          >
            + Neue Rechnung
          </button>
        </div>
      </div>

      {isKleinunternehmer && (
        <div
          style={{
            padding: 'var(--spacing-3) var(--spacing-4)',
            background: 'rgba(59, 130, 246, 0.1)',
            border: '1px solid rgba(59, 130, 246, 0.25)',
            borderRadius: 'var(--radius-md)',
            color: '#60a5fa',
            fontSize: '0.9rem',
            marginBottom: 'var(--spacing-4)',
          }}
        >
          <strong>Kleinunternehmerregelung aktiv</strong> — Rechnungen werden ohne MwSt. ausgestellt (§19 UStG)
        </div>
      )}

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
              style={{ fontSize: '0.85rem', padding: 'var(--spacing-2) var(--spacing-3)' }}
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

      {error ? (
        <ErrorState
          variant="generic"
          title="Fehler beim Laden der Rechnungen"
          description="Die Rechnungsdaten konnten nicht geladen werden. Bitte versuchen Sie es erneut."
          onRetry={() => window.location.reload()}
          compact
        />
      ) : isLoading ? (
        <SkeletonTable rows={6} columns={7} />
      ) : filteredData.length === 0 ? (
        <EmptyState
          icon={FileText}
          title={selectedStatus ? 'Keine Rechnungen mit diesem Status' : 'Noch keine Rechnungen'}
          description={selectedStatus ? 'Versuchen Sie einen anderen Filter.' : 'Erstellen Sie Ihre erste Rechnung, um loszulegen.'}
          action={selectedStatus ? undefined : { label: 'Neue Rechnung', href: '/invoices/new' }}
        />
      ) : (
        <DataTable<Invoice>
          columns={columns}
          data={filteredData}
          rowKey="id"
          loading={false}
          onRowClick={(invoice) => navigate(`/invoices/${invoice.id}`)}
          pagination={{
            page,
            total: invoiceData?.total || 0,
            limit,
            onPageChange: setPage,
          }}
        />
      )}

      {/* Modal: Aus Projekt erstellen */}
      <Modal
        isOpen={showProjectModal}
        onClose={() => {
          setShowProjectModal(false)
          setSelectedProjectId('')
        }}
        title="Rechnung aus Projekt erstellen"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => {
                setShowProjectModal(false)
                setSelectedProjectId('')
              }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={() => {
                if (selectedProjectId) {
                  createFromProject(selectedProjectId)
                }
              }}
              disabled={!selectedProjectId || isCreating}
            >
              {isCreating ? 'Wird erstellt...' : 'Rechnung erstellen'}
            </button>
          </div>
        }
      >
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-3)' }}>
          <label style={{ color: 'var(--color-text-secondary)', fontSize: '0.9rem' }}>
            Projekt auswählen
          </label>
          <select
            className="form-input"
            value={selectedProjectId}
            onChange={(e) => setSelectedProjectId(e.target.value)}
            style={{
              width: '100%',
              padding: 'var(--spacing-3)',
              backgroundColor: 'var(--glass-bg-input-strong)',
              border: '1px solid var(--color-border-strong)',
              borderRadius: 'var(--radius-input)',
              color: 'var(--color-text-primary)',
              fontSize: '0.95rem',
            }}
          >
            <option value="">— Projekt wählen —</option>
            {projects.map((p: { id: string; name?: string; title?: string }) => (
              <option key={p.id} value={p.id}>
                {p.name || p.title || p.id}
              </option>
            ))}
          </select>
          <p style={{ color: 'var(--color-text-secondary)', fontSize: '0.8rem', margin: 0 }}>
            Die Rechnungspositionen werden aus den Projektdaten übernommen.
          </p>
        </div>
      </Modal>
    </div>
  )
}

export default InvoiceListPage
