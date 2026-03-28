import { ReactNode } from 'react';
import { motion } from 'framer-motion';
import './PageWrapper.scss';

interface PageWrapperProps {
  title: string;
  children: ReactNode;
  actions?: ReactNode;
}

const pageVariants = {
  initial: { opacity: 0, y: 12 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
};

export function PageWrapper({ title, children, actions }: PageWrapperProps) {
  return (
    <motion.div
      className="page-wrapper"
      variants={pageVariants}
      initial="initial"
      animate="animate"
      exit="exit"
      transition={{ duration: 0.25, ease: 'easeOut' }}
    >
      <div className="page-wrapper__header">
        <h2 className="page-wrapper__title">{title}</h2>
        {actions && <div className="page-wrapper__actions">{actions}</div>}
      </div>
      <div className="page-wrapper__body">{children}</div>
    </motion.div>
  );
}
