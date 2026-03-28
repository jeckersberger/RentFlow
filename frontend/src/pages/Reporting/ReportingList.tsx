import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as reportingApi from '@/services/reporting';
import type { ReportDefinition, ReportSnapshot } from '@/types/reporting';
import './ReportingList.scss';

type Tab = 'definitions' | 'snapshots';

const definitionColumns: Column<ReportDefinition>[] = [
  { key: 'name', label: 'Name' },
  { key: 'type', label: 'Typ' },
  { key: 'format', label: 'Format' },
  { key: 'schedule', label: 'Zeitplan', render: (row) => row.schedule || '-' },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'cancelled'} />,
  },
];

const snapshotColumns: Column<ReportSnapshot>[] = [
  { key: 'title', label: 'Titel' },
  { key: 'format', label: 'Format' },
  { key: 'period_start', label: 'Von', render: (row) => row.period_start || '-' },
  { key: 'period_end', label: 'Bis', render: (row) => row.period_end || '-' },
  {
    key: 'file_size',
    label: 'Groesse',
    render: (row) => row.file_size ? `${(row.file_size / 1024).toFixed(1)} KB` : '-',
  },
  {
    key: 'generated_at',
    label: 'Generiert',
    render: (row) =>
      new Date(row.generated_at).toLocaleDateString('de-DE'),
  },
];

export default function ReportingList() {
  const [tab, setTab] = useState<Tab>('definitions');

  const { data: definitions, isLoading: definitionsLoading } = useQuery({
    queryKey: ['report-definitions'],
    queryFn: reportingApi.listDefinitions,
  });

  const { data: snapshots, isLoading: snapshotsLoading } = useQuery({
    queryKey: ['report-snapshots'],
    queryFn: () => reportingApi.listSnapshots(),
  });

  const definitionItems: ReportDefinition[] = Array.isArray(definitions) ? definitions : [];
  const snapshotItems: ReportSnapshot[] = Array.isArray(snapshots) ? snapshots : [];

  return (
    <PageWrapper title="Reports">
      <div className="tabs">
        <button className={`tabs__btn ${tab === 'definitions' ? 'tabs__btn--active' : ''}`} onClick={() => setTab('definitions')}>Report-Definitionen</button>
        <button className={`tabs__btn ${tab === 'snapshots' ? 'tabs__btn--active' : ''}`} onClick={() => setTab('snapshots')}>Snapshots</button>
      </div>

      {tab === 'definitions' && (
        <DataTable<ReportDefinition>
          columns={definitionColumns}
          data={definitionItems}
          loading={definitionsLoading}
          emptyMessage="Keine Report-Definitionen vorhanden"
        />
      )}

      {tab === 'snapshots' && (
        <DataTable<ReportSnapshot>
          columns={snapshotColumns}
          data={snapshotItems}
          loading={snapshotsLoading}
          emptyMessage="Keine Snapshots vorhanden"
        />
      )}
    </PageWrapper>
  );
}
