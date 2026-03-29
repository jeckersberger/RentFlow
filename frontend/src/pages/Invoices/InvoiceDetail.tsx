import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Pencil, FileDown, Mail } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as invoiceApi from '@/services/invoices';
import type { Invoice, InvoiceItem, Payment } from '@/types/invoice';
import './InvoiceDetail.scss';

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

export default function InvoiceDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: item, isLoading } = useQuery({
    queryKey: ['invoices', id],
    queryFn: () => invoiceApi.get(id!),
    enabled: !!id,
  });

  // Fetch items and payments separately (service provides these endpoints)
  const { data: invoiceItems } = useQuery({
    queryKey: ['invoices', id, 'items'],
    queryFn: () => invoiceApi.listItems(id!),
    enabled: !!id,
  });

  const { data: payments } = useQuery({
    queryKey: ['invoices', id, 'payments'],
    queryFn: () => invoiceApi.listPayments(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <PageWrapper title="Rechnung">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  if (!item) {
    return (
      <PageWrapper title="Rechnung">
        <p className="empty-state">Rechnung nicht gefunden.</p>
      </PageWrapper>
    );
  }

  const lines: InvoiceItem[] = Array.isArray(invoiceItems) ? invoiceItems : [];
  const pmts: Payment[] = Array.isArray(payments) ? payments : [];

  const handleDownloadPDF = async () => {
    const token = localStorage.getItem('cd_access_token') || '';
    const res = await fetch(`/api/v1/invoices/${id}/pdf`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (!res.ok) return;
    const blob = await res.blob();
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${item?.invoice_number || 'rechnung'}.pdf`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const actionButtons = (
    <div style={{ display: 'flex', gap: '8px' }}>
      <button className="btn btn--primary" onClick={handleDownloadPDF}>
        <FileDown size={16} />
        <span>PDF</span>
      </button>
      <button className="btn btn--secondary" onClick={() => navigate(`/invoices/${id}/edit`)}>
        <Pencil size={16} />
        <span>Bearbeiten</span>
      </button>
    </div>
  );

  const editButton = (
    <button className="btn btn--secondary">
      <Pencil size={16} />
      <span>Bearbeiten</span>
    </button>
  );

  return (
    <PageWrapper title={`Rechnung ${item.invoice_number}`} actions={actionButtons}>
      <button className="btn btn--ghost" onClick={() => navigate(-1)}>
        <ArrowLeft size={18} />
        <span>Zurueck</span>
      </button>

      {/* Invoice header info */}
      <motion.div
        className="detail-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <div className="detail-grid">
          <div className="detail-field">
            <span className="detail-field__label">Rechnungsnummer</span>
            <span className="detail-field__value detail-field__value--mono">
              {item.invoice_number}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Status</span>
            <StatusBadge status={item.status} />
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Rechnungsdatum</span>
            <span className="detail-field__value">
              {formatDate(item.invoice_date)}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Faelligkeitsdatum</span>
            <span className="detail-field__value">
              {formatDate(item.due_date)}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Kunde</span>
            <span className="detail-field__value">
              {item.customer_name || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Notizen</span>
            <span className="detail-field__value">
              {item.notes || '-'}
            </span>
          </div>
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

        <div className="invoice-totals">
          <div className="invoice-totals__row">
            <span>Netto</span>
            <span>{formatEur(item.total_net)}</span>
          </div>
          {!item.kleinunternehmer && (
            <div className="invoice-totals__row">
              <span>MwSt. ({(item.vat_rate / 100).toFixed(0)}%)</span>
              <span>{formatEur(item.total_vat)}</span>
            </div>
          )}
          {item.kleinunternehmer && (
            <div className="invoice-totals__row">
              <span>Kleinunternehmer gem. §19 UStG</span>
              <span>0,00 EUR</span>
            </div>
          )}
          <div className="invoice-totals__row invoice-totals__row--total">
            <span>Brutto</span>
            <span>{formatEur(item.total_gross)}</span>
          </div>
          {item.amount_paid > 0 && (
            <div className="invoice-totals__row">
              <span>Bereits bezahlt</span>
              <span>{formatEur(item.amount_paid)}</span>
            </div>
          )}
        </div>
      </motion.div>

      {/* Payments table */}
      <motion.div
        className="detail-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3, delay: 0.2 }}
      >
        <h3 className="detail-card__title">Zahlungen</h3>
        <div className="data-table-wrapper">
          <table className="data-table">
            <thead>
              <tr>
                <th>Datum</th>
                <th>Methode</th>
                <th>Referenz</th>
                <th className="data-table__th--right">Betrag</th>
              </tr>
            </thead>
            <tbody>
              {pmts.length === 0 ? (
                <tr>
                  <td colSpan={4} className="data-table__empty">
                    Keine Zahlungen
                  </td>
                </tr>
              ) : (
                pmts.map((pmt) => (
                  <tr key={pmt.id}>
                    <td>{formatDate(pmt.payment_date)}</td>
                    <td>{pmt.payment_method || '-'}</td>
                    <td className="data-table__cell--mono">
                      {pmt.reference || '-'}
                    </td>
                    <td className="data-table__cell--right">
                      {formatEur(pmt.amount)}
                    </td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
      </motion.div>
    </PageWrapper>
  );
}
