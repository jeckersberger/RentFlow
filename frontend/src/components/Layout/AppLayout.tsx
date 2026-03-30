import { Outlet, NavLink, useNavigate } from 'react-router-dom';
import { AnimatePresence, motion } from 'framer-motion';
import { useQuery } from '@tanstack/react-query';
import { useSidebarStore } from '@/stores/sidebarStore';
import { fetchCount } from '@/services/api';
import { useWebSocket } from '@/hooks/useWebSocket';
import { NotificationBell } from '@/components/NotificationBell/NotificationBell';
import './AppLayout.scss';

const navItems = [
  { to: '/', icon: '⊞', label: 'Dashboard', badgeKey: null },
  { to: '/calendar', icon: '📅', label: 'Kalender', badgeKey: null },
  { to: '/equipment', icon: '⚙', label: 'Equipment', badgeKey: 'equipment' },
  { to: '/projects', icon: '📋', label: 'Projekte', badgeKey: 'projects_active' },
  { to: '/customers', icon: '👤', label: 'Kunden', badgeKey: null },
  { to: '/quotes', icon: '📝', label: 'Angebote', badgeKey: 'quotes_open' },
  { to: '/invoices', icon: '💰', label: 'Rechnungen', badgeKey: 'invoices_open' },
  { to: '/dunning', icon: '⚠', label: 'Mahnwesen', badgeKey: 'dunning' },
  { to: '/banking', icon: '🏦', label: 'Bank', badgeKey: null },
  { to: '/scanner', icon: '📷', label: 'Scanner', badgeKey: null },
  { to: '/warehouse', icon: '🏭', label: 'Lager', badgeKey: null },
  { to: '/crew', icon: '👥', label: 'Crew', badgeKey: null },
  { to: '/documents', icon: '📄', label: 'Dokumente', badgeKey: null },
  { to: '/expenses', icon: '💳', label: 'Ausgaben', badgeKey: null },
  { to: '/maintenance', icon: '🔧', label: 'Wartung', badgeKey: null },
  { to: '/transport', icon: '🚛', label: 'Transport', badgeKey: null },
  { to: '/workflows', icon: '🔀', label: 'Workflows', badgeKey: null },
  { to: '/reporting', icon: '📊', label: 'Reports', badgeKey: null },
  { to: '/settings', icon: '⚙️', label: 'Einstellungen', badgeKey: null },
  { to: '/account', icon: '👤', label: 'Konto', badgeKey: null },
];

function useBadgeCounts() {
  const { data: equipment } = useQuery({
    queryKey: ['badge', 'equipment'],
    queryFn: () => fetchCount('/api/v1/equipment'),
    staleTime: 60000,
  });
  const { data: projectsActive } = useQuery({
    queryKey: ['badge', 'projects-active'],
    queryFn: () => fetchCount('/api/v1/projects', { status: 'active' }),
    staleTime: 60000,
  });
  const { data: quotesOpen } = useQuery({
    queryKey: ['badge', 'quotes-open'],
    queryFn: () => fetchCount('/api/v1/quotes', { status: 'sent' }),
    staleTime: 60000,
  });
  const { data: invoicesOpen } = useQuery({
    queryKey: ['badge', 'invoices-open'],
    queryFn: () => fetchCount('/api/v1/invoices', { status: 'sent' }),
    staleTime: 60000,
  });
  // Dunning count - try the endpoint, fallback to 0
  const { data: dunning } = useQuery({
    queryKey: ['badge', 'dunning'],
    queryFn: async () => {
      try {
        const token = localStorage.getItem('cd_access_token') || '';
        const res = await fetch('/api/v1/dunning/overdue', {
          headers: { Authorization: `Bearer ${token}` },
        });
        if (!res.ok) return 0;
        const json = await res.json();
        const items = json?.data;
        return Array.isArray(items) ? items.length : 0;
      } catch { return 0; }
    },
    staleTime: 60000,
  });

  return {
    equipment: equipment || 0,
    projects_active: projectsActive || 0,
    quotes_open: quotesOpen || 0,
    invoices_open: invoicesOpen || 0,
    dunning: dunning || 0,
  } as Record<string, number>;
}

export default function AppLayout() {
  const navigate = useNavigate();
  const { isExpanded, isPinned, expand, collapse, pin } = useSidebarStore();
  const badges = useBadgeCounts();
  useWebSocket();

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
          {navItems.map((item) => {
            const count = item.badgeKey ? badges[item.badgeKey] : 0;
            return (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.to === '/'}
                className={({ isActive }) =>
                  `sidebar__link ${isActive ? 'sidebar__link--active' : ''}`
                }
              >
                <span className="sidebar__link-icon">
                  {item.icon}
                  {count > 0 && (
                    <span className="sidebar__badge">{count > 99 ? '99+' : count}</span>
                  )}
                </span>
                {isExpanded && <span className="sidebar__link-label">{item.label}</span>}
                {isExpanded && count > 0 && (
                  <span className="sidebar__badge-expanded">{count}</span>
                )}
              </NavLink>
            );
          })}
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
          <div className="app-header__right">
            <NotificationBell />
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
