import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { AlertTriangle, Send } from 'lucide-react';
import toast from 'react-hot-toast';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import api from '@/services/api';
import './DunningOverview.scss';

interface OverdueInvoice {
  id: string;
  invoice_number: string;
  customer_name: string;
  customer_email: string;
  total_gross: number;
  amount_paid: number;
  due_date: string;
  days_overdue: number;
  suggested_level: string;
  dunning_count: number;
}

function formatEur(cents: number): string {
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

function formatDate(d: string): string {
  if (!d) return '-';
  return new Date(d).toLocaleDateString('de-DE');
}

function levelLabel(level: string): string {
  switch (level) {
    case 'reminder': return 'Zahlungserinnerung';
    case 'dunning_1': return '1. Mahnung';
    case 'dunning_2': return '2. Mahnung';
    case 'dunning_3': return '3. Mahnung';
    default: return level;
  }
}

export default function DunningOverview() {
  const queryClient = useQueryClient();

  const { data, isLoading } = useQuery({
    queryKey: ['dunning', 'overdue'],
    queryFn: async () => {
      const res = await api.get('/api/v1/dunning/overdue');
      return res as unknown as OverdueInvoice[];
    },
  });

  const sendMutation = useMutation({
    mutationFn: async (params: { invoice_id: string; level: string }) => {
      return api.post('/api/v1/dunning/send-reminder', params);
    },
    onSuccess: () => {
      toast.success('Mahnung versendet');
      queryClient.invalidateQueries({ queryKey: ['dunning'] });
    },
    onError: () => toast.error('Versand fehlgeschlagen'),
  });

  const items = Array.isArray(data) ? data : [];

  return (
    <PageWrapper title="Mahnwesen">
      {isLoading ? (
        <div className="detail-skeleton">
          <div className="skeleton skeleton--block" />
        </div>
      ) : items.length === 0 ? (
        <motion.div
          className="dunning-empty"
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <AlertTriangle size={48} />
          <p>Keine ueberfaelligen Rechnungen</p>
        </motion.div>
      ) : (
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
        >
          <div className="dunning-stats">
            <div className="dunning-stat">
              <span className="dunning-stat__value">{items.length}</span>
              <span className="dunning-stat__label">Ueberfaellig</span>
            </div>
            <div className="dunning-stat">
              <span className="dunning-stat__value">
                {formatEur(items.reduce((sum, i) => sum + (i.total_gross - i.amount_paid), 0))}
              </span>
              <span className="dunning-stat__label">Offener Betrag</span>
            </div>
          </div>

          <div className="data-table-wrapper">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Rechnung</th>
                  <th>Kunde</th>
                  <th className="data-table__th--right">Offen</th>
                  <th>Faellig seit</th>
                  <th>Tage</th>
                  <th>Mahnstufe</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items.map((inv) => (
                  <tr key={inv.id}>
                    <td className="data-table__cell--mono">{inv.invoice_number}</td>
                    <td>{inv.customer_name}</td>
                    <td className="data-table__cell--right">
                      {formatEur(inv.total_gross - inv.amount_paid)}
                    </td>
                    <td>{formatDate(inv.due_date)}</td>
                    <td>
                      <StatusBadge
                        status={inv.days_overdue > 28 ? 'cancelled' : inv.days_overdue > 14 ? 'overdue' : 'sent'}
                      />
                      <span style={{ marginLeft: 4 }}>{inv.days_overdue}d</span>
                    </td>
                    <td>{levelLabel(inv.suggested_level)}</td>
                    <td>
                      <button
                        className="btn btn--small btn--primary"
                        onClick={() => sendMutation.mutate({
                          invoice_id: inv.id,
                          level: inv.suggested_level,
                        })}
                        disabled={sendMutation.isPending || !inv.customer_email}
                        title={!inv.customer_email ? 'Keine E-Mail-Adresse' : 'Mahnung senden'}
                      >
                        <Send size={14} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>
      )}
    </PageWrapper>
  );
}
