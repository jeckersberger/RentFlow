import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ScanLine, ArrowUpFromLine, ArrowDownToLine, Wifi } from 'lucide-react';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as scannerApi from '@/services/scanner';
import type { ScanEvent, ScannerDevice } from '@/types/scanner';
import toast from 'react-hot-toast';
import './ScannerPage.scss';

const eventColumns: Column<ScanEvent>[] = [
  {
    key: 'action',
    label: 'Aktion',
    render: (row) => <StatusBadge status={row.action} />,
  },
  { key: 'barcode', label: 'Barcode', render: (row) => row.barcode || '-' },
  { key: 'rfid_tag', label: 'RFID', render: (row) => row.rfid_tag || '-' },
  { key: 'device_id', label: 'Geraet', render: (row) => row.device_id || '-' },
  {
    key: 'timestamp',
    label: 'Zeitpunkt',
    render: (row) =>
      new Date(row.timestamp).toLocaleString('de-DE', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      }),
  },
];

const deviceColumns: Column<ScannerDevice>[] = [
  { key: 'device_name', label: 'Name', render: (row) => row.device_name || row.device_id },
  { key: 'device_type', label: 'Typ' },
  {
    key: 'last_seen',
    label: 'Zuletzt gesehen',
    render: (row) =>
      row.last_seen
        ? new Date(row.last_seen).toLocaleString('de-DE')
        : 'Nie',
  },
  {
    key: 'ring_requested',
    label: 'Ring',
    render: (row) => (row.ring_requested ? 'Aktiv' : '-'),
  },
];

export default function ScannerPage() {
  const queryClient = useQueryClient();
  const [scanInput, setScanInput] = useState('');
  const [activeTab, setActiveTab] = useState<'scan' | 'events' | 'devices'>('scan');

  const { data: eventsData, isLoading: eventsLoading } = useQuery({
    queryKey: ['scanner-events'],
    queryFn: () => scannerApi.listEvents({ page: 1, per_page: 50 }),
    enabled: activeTab === 'events',
  });

  const { data: devicesData, isLoading: devicesLoading } = useQuery({
    queryKey: ['scanner-devices'],
    queryFn: () => scannerApi.listDevices(),
    enabled: activeTab === 'devices',
  });

  const scanMutation = useMutation({
    mutationFn: (barcode: string) => scannerApi.scan({ barcode }),
    onSuccess: () => {
      toast.success('Scan erfolgreich');
      setScanInput('');
      queryClient.invalidateQueries({ queryKey: ['scanner-events'] });
    },
    onError: () => toast.error('Scan fehlgeschlagen'),
  });

  const ringMutation = useMutation({
    mutationFn: (deviceId: string) => scannerApi.triggerRing(deviceId),
    onSuccess: () => toast.success('Ring ausgeloest'),
    onError: () => toast.error('Ring fehlgeschlagen'),
  });

  const events: ScanEvent[] = Array.isArray(eventsData) ? eventsData : [];
  const devices: ScannerDevice[] = Array.isArray(devicesData) ? devicesData : [];

  const handleScan = (e: React.FormEvent) => {
    e.preventDefault();
    if (scanInput.trim()) {
      scanMutation.mutate(scanInput.trim());
    }
  };

  return (
    <PageWrapper title="Scanner">
      <div className="scanner-tabs">
        <button
          className={`scanner-tabs__tab ${activeTab === 'scan' ? 'scanner-tabs__tab--active' : ''}`}
          onClick={() => setActiveTab('scan')}
        >
          <ScanLine size={16} />
          Scannen
        </button>
        <button
          className={`scanner-tabs__tab ${activeTab === 'events' ? 'scanner-tabs__tab--active' : ''}`}
          onClick={() => setActiveTab('events')}
        >
          <ArrowUpFromLine size={16} />
          Scan-Historie
        </button>
        <button
          className={`scanner-tabs__tab ${activeTab === 'devices' ? 'scanner-tabs__tab--active' : ''}`}
          onClick={() => setActiveTab('devices')}
        >
          <Wifi size={16} />
          Geraete
        </button>
      </div>

      {activeTab === 'scan' && (
        <div className="scanner-panel">
          <form onSubmit={handleScan} className="scanner-panel__form">
            <div className="scanner-panel__input-group">
              <ScanLine size={20} className="scanner-panel__icon" />
              <input
                type="text"
                value={scanInput}
                onChange={(e) => setScanInput(e.target.value)}
                placeholder="Barcode oder RFID-Tag eingeben..."
                className="scanner-panel__input"
                autoFocus
              />
            </div>
            <button
              type="submit"
              className="btn btn--primary"
              disabled={scanMutation.isPending}
            >
              Scannen
            </button>
          </form>

          <div className="scanner-panel__actions">
            <button className="btn btn--secondary">
              <ArrowUpFromLine size={16} />
              Check-Out
            </button>
            <button className="btn btn--secondary">
              <ArrowDownToLine size={16} />
              Check-In
            </button>
          </div>
        </div>
      )}

      {activeTab === 'events' && (
        <DataTable<ScanEvent>
          columns={eventColumns}
          data={events}
          loading={eventsLoading}
          emptyMessage="Keine Scan-Ereignisse vorhanden"
        />
      )}

      {activeTab === 'devices' && (
        <DataTable<ScannerDevice>
          columns={deviceColumns}
          data={devices}
          loading={devicesLoading}
          onRowClick={(row) => ringMutation.mutate(row.id)}
          emptyMessage="Keine Scanner-Geraete registriert"
        />
      )}
    </PageWrapper>
  );
}
