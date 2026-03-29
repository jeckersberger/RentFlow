import { useParams, useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { ArrowLeft, Pencil } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as projectApi from '@/services/projects';
import type { Project } from '@/types/project';
import './ProjectDetail.scss';

function formatDate(dateStr?: string): string {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleDateString('de-DE');
}

function formatEur(cents: number | undefined): string {
  if (cents == null) return '-';
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
  }).format(cents / 100);
}

export default function ProjectDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: item, isLoading } = useQuery({
    queryKey: ['projects', id],
    queryFn: () => projectApi.get(id!),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <PageWrapper title="Projekt">
        <div className="detail-skeleton">
          <div className="skeleton skeleton--heading" />
          <div className="skeleton skeleton--block" />
        </div>
      </PageWrapper>
    );
  }

  if (!item) {
    return (
      <PageWrapper title="Projekt">
        <p className="empty-state">Projekt nicht gefunden.</p>
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
    <PageWrapper title={item.name} actions={editButton}>
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
          <div className="detail-field detail-field--full">
            <span className="detail-field__label">Beschreibung</span>
            <span className="detail-field__value">
              {item.description || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Status</span>
            <StatusBadge status={item.status} />
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Ansprechpartner</span>
            <span className="detail-field__value">
              {item.contact_name || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Startdatum</span>
            <span className="detail-field__value">
              {formatDate(item.start_date)}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Enddatum</span>
            <span className="detail-field__value">
              {formatDate(item.end_date)}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Veranstaltungsort</span>
            <span className="detail-field__value">
              {item.venue_name || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Adresse</span>
            <span className="detail-field__value">
              {item.venue_address || '-'}
            </span>
          </div>
          <div className="detail-field">
            <span className="detail-field__label">Gesamtbetrag</span>
            <span className="detail-field__value">
              {formatEur(item.budget)}
            </span>
          </div>
          <div className="detail-field detail-field--full">
            <span className="detail-field__label">Notizen</span>
            <span className="detail-field__value">
              {item.notes || '-'}
            </span>
          </div>
        </div>
      </motion.div>
    </PageWrapper>
  );
}
