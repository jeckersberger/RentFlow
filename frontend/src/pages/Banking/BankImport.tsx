import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { Upload, Check, X, RefreshCw, Download } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import api from '@/services/api';
import './BankImport.scss';

interface BankTransaction {
  id: string;
  booking_date: string;
  amount: number;
  reference: string;
  counterparty_name: string;
  counterparty_iban: string;
  matched_invoice_id: string | null;
  match_confidence: string;
}

function formatEur(cents: number): string {
  return new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(cents / 100);
}

function formatDate(d: string): string {
  if (!d) return '-';
  return new Date(d).toLocaleDateString('de-DE');
}

export default function BankImport() {
  const queryClient = useQueryClient();
  const [tab, setTab] = useState<'import' | 'unmatched' | 'matched'>('unmatched');

  const { data: unmatched } = useQuery({
    queryKey: ['banking', 'unmatched'],
    queryFn: () => api.get('/api/v1/banking/transactions?matched=false') as unknown as BankTransaction[],
  });

  const { data: matched } = useQuery({
    queryKey: ['banking', 'matched'],
    queryFn: () => api.get('/api/v1/banking/transactions?matched=true') as unknown as BankTransaction[],
  });

  const importMutation = useMutation({
    mutationFn: async (file: File) => {
      const text = await file.text();
      return api.post('/api/v1/banking/import', { csv_data: text });
    },
    onSuccess: () => {
      toast.success('Kontoauszug importiert');
      queryClient.invalidateQueries({ queryKey: ['banking'] });
    },
    onError: () => toast.error('Import fehlgeschlagen'),
  });

  const autoMatchMutation = useMutation({
    mutationFn: () => api.post('/api/v1/banking/auto-match') as Promise<unknown>,
    onSuccess: () => {
      toast.success('Automatischer Abgleich durchgefuehrt');
      queryClient.invalidateQueries({ queryKey: ['banking'] });
    },
    onError: () => toast.error('Abgleich fehlgeschlagen'),
  });

  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) importMutation.mutate(file);
  };

  const handleDatevExport = async () => {
    const token = localStorage.getItem('cd_access_token') || '';
    const res = await fetch('/api/v1/export/datev?format=skr03', {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) { toast.error('Export fehlgeschlagen'); return; }
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'datev-export.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  const unmatchedList = Array.isArray(unmatched) ? unmatched : [];
  const matchedList = Array.isArray(matched) ? matched : [];

  const actions = (
    <div style={{ display: 'flex', gap: '8px' }}>
      <label className="btn btn--primary" style={{ cursor: 'pointer' }}>
        <Upload size={16} />
        <span>CSV Import</span>
        <input type="file" accept=".csv" onChange={handleFileUpload} style={{ display: 'none' }} />
      </label>
      <button className="btn btn--secondary" onClick={() => autoMatchMutation.mutate()} disabled={autoMatchMutation.isPending}>
        <RefreshCw size={16} />
        <span>Auto-Abgleich</span>
      </button>
      <button className="btn btn--ghost" onClick={handleDatevExport}>
        <Download size={16} />
        <span>DATEV</span>
      </button>
    </div>
  );

  return (
    <PageWrapper title="Bank & Buchhaltung" actions={actions}>
      <div className="bank-tabs">
        <button className={`bank-tab ${tab === 'unmatched' ? 'bank-tab--active' : ''}`} onClick={() => setTab('unmatched')}>
          Offen ({unmatchedList.length})
        </button>
        <button className={`bank-tab ${tab === 'matched' ? 'bank-tab--active' : ''}`} onClick={() => setTab('matched')}>
          Zugeordnet ({matchedList.length})
        </button>
      </div>

      <motion.div initial={{ opacity: 0, y: 12 }} animate={{ opacity: 1, y: 0 }}>
        <div className="data-table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Datum</th>
                <th>Auftraggeber</th>
                <th>Verwendungszweck</th>
                <th className="data-table__th--right">Betrag</th>
                <th>Status</th>
              </tr>
            </thead>
            <tbody>
              {(tab === 'unmatched' ? unmatchedList : matchedList).map((tx) => (
                <tr key={tx.id}>
                  <td>{formatDate(tx.booking_date)}</td>
                  <td>{tx.counterparty_name || '-'}</td>
                  <td style={{ maxWidth: 300, overflow: 'hidden', textOverflow: 'ellipsis' }}>{tx.reference || '-'}</td>
                  <td className={`data-table__cell--right ${tx.amount >= 0 ? 'text-success' : 'text-error'}`}>
                    {formatEur(tx.amount)}
                  </td>
                  <td>
                    {tx.matched_invoice_id ? (
                      <StatusBadge status={tx.match_confidence === 'high' ? 'paid' : 'sent'} />
                    ) : (
                      <StatusBadge status="draft" />
                    )}
                  </td>
                </tr>
              ))}
              {(tab === 'unmatched' ? unmatchedList : matchedList).length === 0 && (
                <tr><td colSpan={5} className="data-table__empty">Keine Transaktionen</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </motion.div>
    </PageWrapper>
  );
}
