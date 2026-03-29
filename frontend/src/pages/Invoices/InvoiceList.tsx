import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as invoiceApi from '@/services/invoices';
import type { Invoice } from '@/types/invoice';
import './InvoiceList.scss';

function formatEur(cents: number): string {
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('de-DE');
}

const columns: Column<Invoice>[] = [
  { key: 'invoice_number', label: 'Nummer' },
  {
    key: 'customer_name',
    label: 'Kunde',
    render: (row) => row.customer_name || '-',
  },
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
  {
    key: 'total_net',
    label: 'Netto',
    render: (row) => formatEur(row.total_net),
  },
  {
    key: 'total_gross',
    label: 'Brutto',
    render: (row) => formatEur(row.total_gross),
  },
  {
    key: 'invoice_date',
    label: 'Datum',
    render: (row) => formatDate(row.invoice_date),
  },
];

export default function InvoiceList() {
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['invoices'],
    queryFn: () => invoiceApi.list({ page: 1, per_page: 50 }),
  });

  const items: Invoice[] = Array.isArray(data) ? data : [];

  const addButton = (
    <button
      className="btn btn--primary"
      onClick={() => navigate('/invoices/new')}
    >
      <Plus size={16} />
      <span>Neu</span>
    </button>
  );

  return (
    <PageWrapper title="Rechnungen" actions={addButton}>
      <DataTable<Invoice>
        columns={columns}
        data={items}
        loading={isLoading}
        onRowClick={(row) => navigate(`/invoices/${row.id}`)}
        emptyMessage="Keine Rechnungen vorhanden"
      />
    </PageWrapper>
  );
}
