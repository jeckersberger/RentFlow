import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Pencil } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as customerApi from '@/services/customers';
import type { Customer } from '@/types/customer';
import './CustomerDetail.scss';

export default function CustomerDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: item, isLoading } = useQuery({
    queryKey: ['customers', id],
    queryFn: () => customerApi.get(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <PageWrapper title="Kunde">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  if (!item) {
    return (
      <PageWrapper title="Kunde">
        <p className="empty-state">Kunde nicht gefunden.</p>
      </PageWrapper>
    );
  }

  const editButton = (
    <button className="btn btn--secondary">
      <Pencil size={16} />
      <span>Bearbeiten</span>
    </button>
  );

  return (
    <PageWrapper title={item.company_name} actions={editButton}>
      <button className="btn btn--ghost" onClick={() => navigate(-1)}>
        <ArrowLeft size={18} />
        <span>Zurueck</span>
      </button>

      <motion.div
        className="detail-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <div className="detail-grid">
          <div className="detail-field">
            <span className="detail-field__label">Firma</span>
            <span className="detail-field__value">{item.company_name}</span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Kundennummer</span>
            <span className="detail-field__value detail-field__value--mono">
              {item.customer_number || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">E-Mail</span>
            <span className="detail-field__value">{item.email || '-'}</span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Telefon</span>
            <span className="detail-field__value">{item.phone || '-'}</span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Steuernummer</span>
            <span className="detail-field__value">{item.tax_id || '-'}</span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Status</span>
            <StatusBadge status={item.is_active ? 'active' : 'cancelled'} />
          </div>
          <div className="detail-field detail-field--full">
            <span className="detail-field__label">Rechnungsadresse</span>
            <span className="detail-field__value">
              {item.billing_address_street ? (
                <>
                  {item.billing_address_street}<br />
                  {item.billing_address_zip} {item.billing_address_city}
                  {item.billing_address_country && <><br />{item.billing_address_country}</>}
                </>
              ) : '-'}
            </span>
          </div>
          <div className="detail-field detail-field--full">
            <span className="detail-field__label">Notizen</span>
            <span className="detail-field__value">{item.notes || '-'}</span>
          </div>
        </div>
      </motion.div>
    </PageWrapper>
  );
}
