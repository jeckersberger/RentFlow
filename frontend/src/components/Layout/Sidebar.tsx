import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useThemeStore } from '../../stores/themeStore'
import { useState } from 'react'
import './Sidebar.scss'

interface NavItem {
  label: string
  href: string
  icon: string
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    label: 'Core',
    items: [
      { label: 'Dashboard', href: '/', icon: '📊' },
      { label: 'Equipment', href: '/equipment', icon: '📦' },
      { label: 'Projekte', href: '/projects', icon: '📋' },
      { label: 'Rechnungen', href: '/invoices', icon: '💰' },
    ],
  },
  {
    label: 'Operations',
    items: [
      { label: 'Scanner', href: '/scanner', icon: '📱' },
      { label: 'Lager', href: '/warehouse', icon: '🏢' },
      { label: 'Transport', href: '/transport', icon: '🚛' },
      { label: 'Wartung', href: '/maintenance', icon: '🔧' },
    ],
  },
  {
    label: 'Team',
    items: [
      { label: 'Crew', href: '/crew', icon: '👥' },
      { label: 'Dokumente', href: '/documents', icon: '📄' },
      { label: 'Versicherung', href: '/insurance', icon: '🛡️' },
    ],
  },
  {
    label: 'Reports & Admin',
    items: [
      { label: 'Reports', href: '/reports', icon: '📈' },
      { label: 'KI-Assistent', href: '/ai', icon: '🤖' },
      { label: 'Workflows', href: '/workflows', icon: '⚡' },
      { label: 'Audit-Log', href: '/audit', icon: '📋' },
      { label: 'Admin', href: '/admin', icon: '⚙️' },
    ],
  },
]

// Flat list of key items for mobile bottom nav
const mobileNavItems: NavItem[] = [
  { label: 'Dashboard', href: '/', icon: '📊' },
  { label: 'Equipment', href: '/equipment', icon: '📦' },
  { label: 'Projekte', href: '/projects', icon: '📋' },
  { label: 'Scanner', href: '/scanner', icon: '📱' },
  { label: 'Mehr', href: '/settings', icon: '⚙️' },
]

function isActiveRoute(pathname: string, href: string): boolean {
  if (href === '/') return pathname === '/'
  return pathname === href || pathname.startsWith(href + '/')
}

function Sidebar() {
  const location = useLocation()
  const navigate = useNavigate()
  const isDarkMode = useThemeStore((state) => state.isDarkMode)
  const toggleDarkMode = useThemeStore((state) => state.toggleDarkMode)
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const [collapsed, setCollapsed] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const userInitials = user?.name
    ? user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
    : user?.email?.[0]?.toUpperCase() || '?'

  return (
    <>
      {/* Desktop sidebar */}
      <aside className={`sidebar ${collapsed ? 'sidebar--collapsed' : ''}`}>
        <div className="sidebar__header">
          <div className="sidebar__logo">
            <div className="sidebar__logo-icon">RF</div>
            {!collapsed && <span className="sidebar__logo-text">RentFlow</span>}
          </div>
          <button
            className="sidebar__collapse-btn"
            onClick={() => setCollapsed(!collapsed)}
            title={collapsed ? 'Sidebar einblenden' : 'Sidebar ausblenden'}
          >
            {collapsed ? '»' : '«'}
          </button>
        </div>

        <nav className="sidebar__nav">
          {navGroups.map((group) => (
            <div key={group.label} className="sidebar__nav-group">
              {!collapsed && (
                <div className="sidebar__nav-group-label">{group.label}</div>
              )}
              {group.items.map((item) => (
                <Link
                  key={item.href}
                  to={item.href}
                  className={`sidebar__nav-item ${isActiveRoute(location.pathname, item.href) ? 'sidebar__nav-item--active' : ''}`}
                  title={collapsed ? item.label : undefined}
                >
                  <span className="sidebar__nav-icon">{item.icon}</span>
                  {!collapsed && <span className="sidebar__nav-label">{item.label}</span>}
                </Link>
              ))}
            </div>
          ))}
        </nav>

        <div className="sidebar__footer">
          <button
            className="sidebar__theme-toggle"
            onClick={toggleDarkMode}
            title={isDarkMode ? 'Light Mode' : 'Dark Mode'}
          >
            {isDarkMode ? '☀️' : '🌙'}
          </button>

          <div className="sidebar__user-section">
            <div className="sidebar__user-avatar">{userInitials}</div>
            {!collapsed && (
              <div className="sidebar__user-details">
                <p className="sidebar__user-name">{user?.name || 'Benutzer'}</p>
                <p className="sidebar__user-email">{user?.email}</p>
              </div>
            )}
            <button
              className="sidebar__logout-btn"
              onClick={handleLogout}
              title="Abmelden"
            >
              🚪
            </button>
          </div>

          {!collapsed && <p className="sidebar__version">v1.0.0</p>}
        </div>
      </aside>

      {/* Mobile bottom nav */}
      <nav className="mobile-nav">
        {mobileNavItems.map((item) => (
          <Link
            key={item.href}
            to={item.href}
            className={`mobile-nav__item ${isActiveRoute(location.pathname, item.href) ? 'mobile-nav__item--active' : ''}`}
          >
            <span className="mobile-nav__icon">{item.icon}</span>
            <span className="mobile-nav__label">{item.label}</span>
          </Link>
        ))}
      </nav>
    </>
  )
}

export default Sidebar
