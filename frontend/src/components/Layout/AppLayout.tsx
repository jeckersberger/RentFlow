import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { AnimatePresence, motion } from 'framer-motion';
import { useSidebarStore } from '@/stores/sidebarStore';
import './AppLayout.scss';

const navItems = [
  { to: '/', icon: '⊞', label: 'Dashboard' },
  { to: '/equipment', icon: '⚙', label: 'Equipment' },
  { to: '/projects', icon: '📋', label: 'Projekte' },
  { to: '/customers', icon: '👤', label: 'Kunden' },
  { to: '/quotes', icon: '📝', label: 'Angebote' },
  { to: '/invoices', icon: '💰', label: 'Rechnungen' },
  { to: '/dunning', icon: '⚠', label: 'Mahnwesen' },
  { to: '/scanner', icon: '📷', label: 'Scanner' },
  { to: '/warehouse', icon: '🏭', label: 'Lager' },
  { to: '/crew', icon: '👥', label: 'Crew' },
  { to: '/documents', icon: '📄', label: 'Dokumente' },
  { to: '/expenses', icon: '💳', label: 'Ausgaben' },
  { to: '/maintenance', icon: '🔧', label: 'Wartung' },
  { to: '/transport', icon: '🚛', label: 'Transport' },
  { to: '/reporting', icon: '📊', label: 'Reports' },
];

export default function AppLayout() {
  const navigate = useNavigate();
  const { isExpanded, isPinned, expand, collapse, pin } = useSidebarStore();

  const handleLogout = () => {
    localStorage.removeItem('cd_access_token');
    localStorage.removeItem('cd_user');
    navigate('/login', { replace: true });
  };

  return (
    <div className="app-layout">
      <aside
        className={`sidebar ${isExpanded ? 'sidebar--expanded' : ''} ${isPinned ? 'sidebar--pinned' : ''}`}
        onMouseEnter={() => !isPinned && expand()}
        onMouseLeave={() => !isPinned && collapse()}
      >
        <div className="sidebar__logo">
          <span className="sidebar__logo-icon">CD</span>
          {isExpanded && <span className="sidebar__logo-text">CrateDesk</span>}
        </div>

        <nav className="sidebar__nav">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === '/'}
              className={({ isActive }) =>
                `sidebar__link ${isActive ? 'sidebar__link--active' : ''}`
              }
            >
              <span className="sidebar__link-icon">{item.icon}</span>
              {isExpanded && <span className="sidebar__link-label">{item.label}</span>}
            </NavLink>
          ))}
        </nav>

        <div className="sidebar__footer">
          <button className="sidebar__pin-btn" onClick={pin} title={isPinned ? 'Loslassen' : 'Anpinnen'}>
            {isPinned ? '📌' : '📍'}
          </button>
          {isExpanded && (
            <button className="sidebar__logout-btn" onClick={handleLogout}>
              Abmelden
            </button>
          )}
        </div>
      </aside>

      <div className="app-layout__main">
        <header className="app-header">
          <div className="app-header__left">
            <h2 className="app-header__title">CrateDesk</h2>
          </div>
        </header>

        <main className="app-layout__content">
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, y: 8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              transition={{ duration: 0.15 }}
            >
              <Outlet />
            </motion.div>
          </AnimatePresence>
        </main>
      </div>
    </div>
  );
}
