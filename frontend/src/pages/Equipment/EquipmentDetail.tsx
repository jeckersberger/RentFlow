import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Pencil } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as equipmentApi from '@/services/equipment';
import type { Equipment } from '@/types/equipment';
import './EquipmentDetail.scss';

function formatEur(cents: number | undefined): string {
  if (cents == null) return '-';
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

export default function EquipmentDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: item, isLoading } = useQuery({
    queryKey: ['equipment', id],
    queryFn: () => equipmentApi.get(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <PageWrapper title="Equipment">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  if (!item) {
    return (
      <PageWrapper title="Equipment">
        <p className="empty-state">Equipment nicht gefunden.</p>
      </PageWrapper>
    );
  }

  const backButton = (
    <button className="btn btn--ghost" onClick={() => navigate(-1)}>
      <ArrowLeft size={18} />
      <span>Zurueck</span>
    </button>
  );

  const editButton = (
    <button className="btn btn--secondary">
      <Pencil size={16} />
      <span>Bearbeiten</span>
    </button>
  );

  return (
    <PageWrapper title={item.name} actions={editButton}>
      {backButton}
      <motion.div
        className="detail-card"
        initial={{ opacity: 0, y: 12 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        <div className="detail-grid">
          <DetailField label="Beschreibung" value={item.description || '-'} />
          <DetailField label="Kategorie" value={item.category_id || '-'} />
          <DetailField label="Barcode" value={item.barcode || '-'} mono />
          <DetailField label="Seriennummer" value={item.serial_number || '-'} mono />
          <div className="detail-field">
            <span className="detail-field__label">Status</span>
            <StatusBadge status={item.status} />
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Zustand</span>
            <StatusBadge status={item.condition} />
          </div>
          <DetailField label="Tagessatz" value={formatEur(item.rental_price_day)} />
          <DetailField label="Kaufpreis" value={formatEur(item.purchase_price)} />
          <DetailField label="Gewicht" value={item.weight_grams ? `${(item.weight_grams / 1000).toFixed(1)} kg` : '-'} />
        </div>
      </motion.div>
    </PageWrapper>
  );
}

function DetailField({
  label,
  value,
  mono = false,
}: {
  label: string;
  value: string;
  mono?: boolean;
}) {
  return (
    <div className="detail-field">
      <span className="detail-field__label">{label}</span>
      <span className={`detail-field__value ${mono ? 'detail-field__value--mono' : ''}`}>
        {value}
      </span>
    </div>
  );
}
