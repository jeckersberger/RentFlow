import { useParams, useNavigate } from 'react-router-dom';
import { useQuery, useMutation } from '@tanstack/react-query';
import toast from 'react-hot-toast';
import { ArrowLeft, Pencil, FileText } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as quoteApi from '@/services/quotes';
import type { Quote, QuoteItem } from '@/types/quote';
import './QuoteDetail.scss';

function formatEur(cents: number | undefined): string {
  if (cents == null) return '-';
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('de-DE');
}

export default function QuoteDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: item, isLoading } = useQuery({
    queryKey: ['quotes', id],
    queryFn: () => quoteApi.get(id!),
    enabled: !!id,
  });

  const { data: quoteItems } = useQuery({
    queryKey: ['quotes', id, 'items'],
    queryFn: () => quoteApi.listItems(id!),
    enabled: !!id,
  });

  const convertMutation = useMutation({
    mutationFn: () => quoteApi.convertToInvoice(id!),
    onSuccess: (data) => {
      toast.success('Angebot in Rechnung umgewandelt');
      navigate(`/invoices/${data.invoice_id}`);
    },
    onError: () => {
      toast.error('Fehler beim Umwandeln in Rechnung');
    },
  });

  if (isLoading) {
    return (
      <PageWrapper title="Angebot">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  if (!item) {
    return (
      <PageWrapper title="Angebot">
        <p className="empty-state">Angebot nicht gefunden.</p>
      </PageWrapper>
    );
  }

  const lines: QuoteItem[] = Array.isArray(quoteItems) ? quoteItems : [];

  const actionButtons = (
    <div style={{ display: 'flex', gap: '8px' }}>
      <button
        className="btn btn--primary"
        onClick={() => convertMutation.mutate()}
        disabled={convertMutation.isPending}
        title="Angebot in Rechnung umwandeln"
      >
        <FileText size={16} />
        <span>
          {convertMutation.isPending
            ? 'Wird umgewandelt...'
            : 'In Rechnung umwandeln'}
        </span>
      </button>
      <button
        className="btn btn--ghost"
        onClick={() => navigate(`/quotes/${id}/edit`)}
      >
        <Pencil size={16} />
        <span>Bearbeiten</span>
      </button>
    </div>
  );

  return (
    <PageWrapper title={`Angebot ${item.quote_number}`} actions={actionButtons}>
      <button className="btn btn--ghost" onClick={() => navigate(-1)}>
        <ArrowLeft size={18} />
        <span>Zurueck</span>
      </button>

      {/* Quote header info */}
      <motion.div
        className="detail-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <div className="detail-grid">
          <div className="detail-field">
            <span className="detail-field__label">Angebotsnummer</span>
            <span className="detail-field__value detail-field__value--mono">
              {item.quote_number}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Status</span>
            <StatusBadge status={item.status} />
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Kunde</span>
            <span className="detail-field__value">
              {item.customer_name || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">E-Mail</span>
            <span className="detail-field__value">
              {item.customer_email || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Angebotsdatum</span>
            <span className="detail-field__value">
              {formatDate(item.quote_date)}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Gueltig bis</span>
            <span className="detail-field__value">
              {formatDate(item.valid_until)}
            </span>
          </div>
          {item.subject && (
            <div className="detail-field detail-field--full">
              <span className="detail-field__label">Betreff</span>
              <span className="detail-field__value">{item.subject}</span>
            </div>
          )}
          {item.intro_text && (
            <div className="detail-field detail-field--full">
              <span className="detail-field__label">Einleitungstext</span>
              <span className="detail-field__value">{item.intro_text}</span>
            </div>
          )}
          {item.notes && (
            <div className="detail-field detail-field--full">
              <span className="detail-field__label">Notizen</span>
              <span className="detail-field__value">{item.notes}</span>
            </div>
          )}
        </div>
      </motion.div>

      {/* Items table */}
      <motion.div
        className="detail-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.1 }}
      >
        <h3 className="detail-card__title">Positionen</h3>
        <div className="data-table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Beschreibung</th>
                <th className="data-table__th--right">Menge</th>
                <th className="data-table__th--right">Einheit</th>
                <th className="data-table__th--right">Einzelpreis</th>
                <th className="data-table__th--right">Gesamt</th>
              </tr>
            </thead>
            <tbody>
              {lines.length === 0 ? (
                <tr>
                  <td colSpan={5} className="data-table__empty">
                    Keine Positionen
                  </td>
                </tr>
              ) : (
                lines.map((line) => (
                  <tr key={line.id}>
                    <td>{line.description}</td>
                    <td className="data-table__cell--right">{line.quantity}</td>
                    <td className="data-table__cell--right">{line.unit}</td>
                    <td className="data-table__cell--right">
                      {formatEur(line.unit_price)}
                    </td>
                    <td className="data-table__cell--right">
                      {formatEur(line.quantity * line.unit_price)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>

        <div className="quote-totals">
          <div className="quote-totals__row">
            <span>Netto</span>
            <span>{formatEur(item.total_net)}</span>
          </div>
          {item.kleinunternehmer && (
            <div className="quote-totals__row">
              <span>Kleinunternehmer gem. &sect;19 UStG</span>
              <span>0,00 EUR</span>
            </div>
          )}
          {!item.kleinunternehmer && item.vat_rate > 0 && (
            <div className="quote-totals__row">
              <span>MwSt. ({(item.vat_rate / 100).toFixed(0)}%)</span>
              <span>{formatEur(item.total_vat)}</span>
            </div>
          )}
          <div className="quote-totals__row quote-totals__row--total">
            <span>Brutto</span>
            <span>{formatEur(item.total_gross)}</span>
          </div>
        </div>
      </motion.div>

      {/* Outro text */}
      {item.outro_text && (
        <motion.div
          className="detail-card"
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.3, delay: 0.2 }}
        >
          <h3 className="detail-card__title">Schlusstext</h3>
          <p className="quote-outro-text">{item.outro_text}</p>
        </motion.div>
      )}
    </PageWrapper>
  );
}
