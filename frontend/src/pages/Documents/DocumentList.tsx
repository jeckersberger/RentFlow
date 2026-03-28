import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as documentApi from '@/services/document';
import type { DocumentTemplate, Document } from '@/types/document';
import './DocumentList.scss';

type Tab = 'templates' | 'documents';

const templateColumns: Column<DocumentTemplate>[] = [
  { key: 'name', label: 'Name' },
  { key: 'type', label: 'Typ' },
  {
    key: 'is_default',
    label: 'Standard',
    render: (row) => row.is_default ? 'Ja' : '-',
  },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'cancelled'} />,
  },
];

const documentColumns: Column<Document>[] = [
  { key: 'title', label: 'Titel' },
  { key: 'type', label: 'Typ' },
  { key: 'mime_type', label: 'Format' },
  {
    key: 'file_size',
    label: 'Groesse',
    render: (row) => row.file_size ? `${(row.file_size / 1024).toFixed(1)} KB` : '-',
  },
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
  {
    key: 'created_at',
    label: 'Erstellt',
    render: (row) =>
      new Date(row.created_at).toLocaleDateString('de-DE'),
  },
];

export default function DocumentList() {
  const [tab, setTab] = useState<Tab>('documents');

  const { data: templates, isLoading: templatesLoading } = useQuery({
    queryKey: ['document-templates'],
    queryFn: documentApi.listTemplates,
  });

  const { data: documents, isLoading: documentsLoading } = useQuery({
    queryKey: ['documents'],
    queryFn: documentApi.listDocuments,
  });

  const templateItems: DocumentTemplate[] = Array.isArray(templates) ? templates : [];
  const documentItems: Document[] = Array.isArray(documents) ? documents : [];

  return (
    <PageWrapper title="Dokumente">
      <div className="tabs">
        <button className={`tabs__btn ${tab === 'documents' ? 'tabs__btn--active' : ''}`} onClick={() => setTab('documents')}>Dokumente</button>
        <button className={`tabs__btn ${tab === 'templates' ? 'tabs__btn--active' : ''}`} onClick={() => setTab('templates')}>Templates</button>
      </div>

      {tab === 'documents' && (
        <DataTable<Document>
          columns={documentColumns}
          data={documentItems}
          loading={documentsLoading}
          emptyMessage="Keine Dokumente vorhanden"
        />
      )}

      {tab === 'templates' && (
        <DataTable<DocumentTemplate>
          columns={templateColumns}
          data={templateItems}
          loading={templatesLoading}
          emptyMessage="Keine Templates vorhanden"
        />
      )}
    </PageWrapper>
  );
}
