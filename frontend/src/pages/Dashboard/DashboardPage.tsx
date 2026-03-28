import { Package, FolderKanban, FileText, Users } from 'lucide-react';
import { motion } from 'framer-motion';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import './DashboardPage.scss';

interface KpiCardProps {
  icon: React.ReactNode;
  label: string;
  value: string | number;
  color: string;
  delay?: number;
}

function KpiCard({ icon, label, value, color, delay = 0 }: KpiCardProps) {
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
        <span className="kpi-card__value">{value}</span>
        <span className="kpi-card__label">{label}</span>
      </div>
    </motion.div>
  );
}

export default function DashboardPage() {
  // Placeholder data -- will be connected to APIs later
  const kpis = [
    {
      icon: <Package size={24} />,
      label: 'Equipment',
      value: 248,
      color: '#00d4ff',
    },
    {
      icon: <FolderKanban size={24} />,
      label: 'Aktive Projekte',
      value: 12,
      color: '#22c55e',
    },
    {
      icon: <FileText size={24} />,
      label: 'Offene Rechnungen',
      value: '4.320 EUR',
      color: '#f59e0b',
    },
    {
      icon: <Users size={24} />,
      label: 'Kunden',
      value: 37,
      color: '#8b5cf6',
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
          <p className="bento-card__placeholder">
            Projektdaten werden geladen...
          </p>
        </motion.div>

        <motion.div
          className="bento-card"
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.35, delay: 0.42 }}
        >
          <h3 className="bento-card__title">Aktivitaet</h3>
          <p className="bento-card__placeholder">
            Letzte Aktivitaeten...
          </p>
        </motion.div>
      </div>
    </PageWrapper>
  );
}
