import { Package, FolderKanban, FileText, Users } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { fetchCount } from '@/services/api';
import * as projectApi from '@/services/projects';
import type { Project } from '@/types/project';
import './DashboardPage.scss';

interface KpiCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color: string;
  delay?: number;
  loading?: boolean;
}

function KpiCard({ icon, label, value, color, delay = 0, loading }: KpiCardProps) {
  return (
    <motion.div
      className="kpi-card"
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35, delay }}
    >
      <div className="kpi-card__icon" style={{ color }}>
        {icon}
      </div>
      <div className="kpi-card__content">
        <span className="kpi-card__value">{loading ? '...' : value}</span>
        <span className="kpi-card__label">{label}</span>
      </div>
    </motion.div>
  );
}

function formatEur(cents: number): string {
  return new Intl.NumberFormat('de-DE', {
    style: 'currency',
    currency: 'EUR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(cents / 100);
}

export default function DashboardPage() {
  const { data: equipmentCount = 0, isLoading: loadingEq } = useQuery({
    queryKey: ['dashboard', 'equipment-count'],
    queryFn: () => fetchCount('/api/v1/equipment'),
  });

  const { data: projectCount = 0, isLoading: loadingPr } = useQuery({
    queryKey: ['dashboard', 'project-count'],
    queryFn: () => fetchCount('/api/v1/projects', { status: 'active' }),
  });

  const { data: invoiceTotal = 0, isLoading: loadingInv } = useQuery({
    queryKey: ['dashboard', 'invoice-total'],
    queryFn: () => fetchCount('/api/v1/invoices', { status: 'sent' }),
  });

  const { data: customerCount = 0, isLoading: loadingCu } = useQuery({
    queryKey: ['dashboard', 'customer-count'],
    queryFn: () => fetchCount('/api/v1/customers'),
  });

  const { data: recentProjects } = useQuery({
    queryKey: ['dashboard', 'recent-projects'],
    queryFn: () => projectApi.list({ page: 1, per_page: 5 }),
  });

  const projects: Project[] = Array.isArray(recentProjects) ? recentProjects : [];

  const kpis = [
    {
      icon: <Package size={24} />,
      label: 'Equipment',
      value: equipmentCount,
      color: '#00d4ff',
      loading: loadingEq,
    },
    {
      icon: <FolderKanban size={24} />,
      label: 'Aktive Projekte',
      value: projectCount,
      color: '#22c55e',
      loading: loadingPr,
    },
    {
      icon: <FileText size={24} />,
      label: 'Offene Rechnungen',
      value: invoiceTotal,
      color: '#f59e0b',
      loading: loadingInv,
    },
    {
      icon: <Users size={24} />,
      label: 'Kunden',
      value: customerCount,
      color: '#8b5cf6',
      loading: loadingCu,
    },
  ];

  return (
    <PageWrapper title="Dashboard">
      <div className="bento-grid">
        {kpis.map((kpi, i) => (
          <KpiCard key={kpi.label} {...kpi} delay={i * 0.08} />
        ))}
      </div>

      <div className="bento-grid bento-grid--wide">
        <motion.div
          className="bento-card bento-card--span-2"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35, delay: 0.35 }}
        >
          <h3 className="bento-card__title">Aktuelle Projekte</h3>
          {projects.length > 0 ? (
            <ul className="bento-card__list">
              {projects.map((p) => (
                <li key={p.id} className="bento-card__list-item">
                  <span>{p.name}</span>
                  <span className="bento-card__meta">{p.status}</span>
                </li>
              ))}
            </ul>
          ) : (
            <p className="bento-card__placeholder">Keine Projekte vorhanden</p>
          )}
        </motion.div>

        <motion.div
          className="bento-card"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35, delay: 0.42 }}
        >
          <h3 className="bento-card__title">Kurzinfo</h3>
          <ul className="bento-card__list">
            <li className="bento-card__list-item">
              <span>Equipment gesamt</span>
              <span className="bento-card__meta">{equipmentCount}</span>
            </li>
            <li className="bento-card__list-item">
              <span>Kunden gesamt</span>
              <span className="bento-card__meta">{customerCount}</span>
            </li>
            <li className="bento-card__list-item">
              <span>Offene Rechnungen</span>
              <span className="bento-card__meta">{invoiceTotal}</span>
            </li>
          </ul>
        </motion.div>
      </div>
    </PageWrapper>
  );
}
