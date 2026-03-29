import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { Plus } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as quoteApi from '@/services/quotes';
import type { Quote } from '@/types/quote';

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

const columns: Column<Quote>[] = [
  { key: 'quote_number', label: 'Nummer' },
  { key: 'customer_name', label: 'Kunde', render: (row) => row.customer_name || '-' },
  { key: 'subject', label: 'Betreff', render: (row) => row.subject || '-' },
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
  {
    key: 'total_gross',
    label: 'Betrag',
    render: (row) => formatEur(row.total_gross),
  },
  {
    key: 'valid_until',
    label: 'Gueltig bis',
    render: (row) => formatDate(row.valid_until),
  },
];

export default function QuoteList() {
  const navigate = useNavigate();

  const { data, isLoading } = useQuery({
    queryKey: ['quotes'],
    queryFn: () => quoteApi.list({ page: 1, per_page: 50 }),
  });

  const items: Quote[] = Array.isArray(data) ? data : [];

  const addButton = (
    <button
      className="btn btn--primary"
      onClick={() => navigate('/quotes/new')}
    >
      <Plus size={16} />
      <span>Neues Angebot</span>
    </button>
  );

  return (
    <PageWrapper title="Angebote" actions={addButton}>
      <DataTable<Quote>
        columns={columns}
        data={items}
        loading={isLoading}
        onRowClick={(row) => navigate(`/quotes/${row.id}`)}
        emptyMessage="Keine Angebote vorhanden"
      />
    </PageWrapper>
  );
}
