import { useState, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import { fetchCount } from '@/services/api';
import * as reportingApi from '@/services/reporting';
import type { ReportDefinition, ReportSnapshot } from '@/types/reporting';
import './ReportingList.scss';

type Tab = 'dashboard' | 'definitions' | 'snapshots';

interface KpiData {
  equipment: number;
  projects: number;
  invoices: number;
  customers: number;
}

const BAR_COLORS = ['#00d4ff', '#8b5cf6', '#22c55e', '#f59e0b'];

function KpiCard({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="kpi-card">
      <div className="kpi-card__value" style={{ color }}>{value}</div>
      <div className="kpi-card__label">{label}</div>
    </div>
  );
}

function KpiBarChart({ kpis }: { kpis: KpiData }) {
  const entries = [
    { label: 'Equipment', value: kpis.equipment, color: BAR_COLORS[0] },
    { label: 'Projekte', value: kpis.projects, color: BAR_COLORS[1] },
    { label: 'Rechnungen', value: kpis.invoices, color: BAR_COLORS[2] },
    { label: 'Kunden', value: kpis.customers, color: BAR_COLORS[3] },
  ];

  const maxValue = Math.max(...entries.map((e) => e.value), 1);

  return (
    <div className="kpi-bar-chart">
      <h3 className="kpi-bar-chart__title">Uebersicht</h3>
      <div className="kpi-bar-chart__bars">
        {entries.map((entry) => (
          <div key={entry.label} className="kpi-bar-chart__row">
            <span className="kpi-bar-chart__label">{entry.label}</span>
            <div className="kpi-bar-chart__track">
              <div
                className="kpi-bar-chart__fill"
                style={{
                  width: `${Math.max((entry.value / maxValue) * 100, 2)}%`,
                  backgroundColor: entry.color,
                }}
              />
            </div>
            <span className="kpi-bar-chart__value">{entry.value}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

function KpiDashboard() {
  const [kpis, setKpis] = useState<KpiData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function loadKpis() {
      try {
        const [equipment, projects, invoices, customers] = await Promise.all([
          fetchCount('/api/v1/equipment').catch(() => 0),
          fetchCount('/api/v1/projects').catch(() => 0),
          fetchCount('/api/v1/invoices').catch(() => 0),
          fetchCount('/api/v1/customers').catch(() => 0),
        ]);
        if (!cancelled) {
          setKpis({ equipment, projects, invoices, customers });
        }
      } catch {
        if (!cancelled) setError(true);
      } finally {
        if (!cancelled) setLoading(false);
      }
    }

    loadKpis();
    return () => { cancelled = true; };
  }, []);

  if (loading) {
    return <div className="kpi-loading">KPIs werden geladen...</div>;
  }

  if (error || !kpis) {
    return (
      <div className="kpi-empty">
        <div className="kpi-empty__icon">&#128202;</div>
        <p className="kpi-empty__text">Daten werden gesammelt</p>
        <p className="kpi-empty__sub">
          Sobald Daten in den Services vorhanden sind, erscheinen hier die KPIs.
        </p>
      </div>
    );
  }

  const total = kpis.equipment + kpis.projects + kpis.invoices + kpis.customers;
  if (total === 0) {
    return (
      <div className="kpi-empty">
        <div className="kpi-empty__icon">&#128202;</div>
        <p className="kpi-empty__text">Daten werden gesammelt</p>
        <p className="kpi-empty__sub">
          Legen Sie Equipment, Projekte, Rechnungen oder Kunden an, um KPIs zu sehen.
        </p>
      </div>
    );
  }

  return (
    <div className="kpi-dashboard">
      <div className="kpi-dashboard__cards">
        <KpiCard label="Equipment gesamt" value={kpis.equipment} color={BAR_COLORS[0]} />
        <KpiCard label="Aktive Projekte" value={kpis.projects} color={BAR_COLORS[1]} />
        <KpiCard label="Offene Rechnungen" value={kpis.invoices} color={BAR_COLORS[2]} />
        <KpiCard label="Kunden" value={kpis.customers} color={BAR_COLORS[3]} />
      </div>
      <KpiBarChart kpis={kpis} />
    </div>
  );
}

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
  const [tab, setTab] = useState<Tab>('dashboard');

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
        <button
          className={`tabs__btn ${tab === 'dashboard' ? 'tabs__btn--active' : ''}`}
          onClick={() => setTab('dashboard')}
        >
          KPI Dashboard
        </button>
        <button
          className={`tabs__btn ${tab === 'definitions' ? 'tabs__btn--active' : ''}`}
          onClick={() => setTab('definitions')}
        >
          Report-Definitionen
        </button>
        <button
          className={`tabs__btn ${tab === 'snapshots' ? 'tabs__btn--active' : ''}`}
          onClick={() => setTab('snapshots')}
        >
          Snapshots
        </button>
      </div>

      {tab === 'dashboard' && <KpiDashboard />}

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
