import { motion } from 'framer-motion';
import { FileText, Download, CalendarCheck } from 'lucide-react';
import './PortalPage.scss';

interface FeatureCardProps {
  icon: React.ReactNode;
  title: string;
  description: string;
  delay: number;
}

function FeatureCard({ icon, title, description, delay }: FeatureCardProps) {
  return (
    <motion.div
      className="portal-feature"
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.4, delay }}
    >
      <div className="portal-feature__icon">{icon}</div>
      <h3 className="portal-feature__title">{title}</h3>
      <p className="portal-feature__description">{description}</p>
    </motion.div>
  );
}

const features: Omit<FeatureCardProps, 'delay'>[] = [
  {
    icon: <FileText size={28} />,
    title: 'Angebote einsehen',
    description:
      'Sehen Sie Ihre aktuellen und vergangenen Angebote ein und akzeptieren Sie diese direkt online.',
  },
  {
    icon: <Download size={28} />,
    title: 'Rechnungen herunterladen',
    description:
      'Laden Sie Ihre Rechnungen als PDF herunter und behalten Sie den Ueberblick ueber alle Zahlungen.',
  },
  {
    icon: <CalendarCheck size={28} />,
    title: 'Verfuegbarkeit pruefen',
    description:
      'Pruefen Sie in Echtzeit, ob das gewuenschte Equipment zu Ihrem Wunschtermin verfuegbar ist.',
  },
];

export default function PortalPage() {
  return (
    <div className="portal-page">
      <motion.div
        className="portal-hero"
        initial={{ opacity: 0, y: 24 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, ease: 'easeOut' }}
      >
        <div className="portal-hero__logo">CD</div>
        <h1 className="portal-hero__title">Kunden-Portal</h1>
        <p className="portal-hero__subtitle">
          Dieses Feature wird bald verfuegbar sein
        </p>
      </motion.div>

      <div className="portal-features">
        {features.map((feature, i) => (
          <FeatureCard key={feature.title} {...feature} delay={0.3 + i * 0.12} />
        ))}
      </div>

      <motion.p
        className="portal-footer"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ duration: 0.5, delay: 0.8 }}
      >
        Powered by CrateDesk
      </motion.p>
    </div>
  );
}
