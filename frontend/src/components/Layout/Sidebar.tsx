import { Link, useLocation } from 'react-router-dom'
import { useThemeStore } from '../../stores/themeStore'
import './Sidebar.scss'

interface NavItem {
  label: string
  href: string
  icon: string
}

const navItems: NavItem[] = [
  { label: 'Dashboard', href: '/', icon: '📊' },
  { label: 'Ausrüstung', href: '/equipment', icon: '📦' },
  { label: 'Projekte', href: '/projects', icon: '📋' },
  { label: 'Rechnungen', href: '/invoices', icon: '💰' },
  { label: 'Scanner', href: '/scanner', icon: '📱' },
  { label: 'Lager', href: '/warehouse', icon: '🏢' },
  { label: 'Transport', href: '/transport', icon: '🚛' },
  { label: 'Wartung', href: '/maintenance', icon: '🔧' },
  { label: 'Crew', href: '/crew', icon: '👥' },
  { label: 'Dokumente', href: '/documents', icon: '📄' },
  { label: 'Versicherung', href: '/insurance', icon: '🛡️' },
  { label: 'Reports', href: '/reports', icon: '📈' },
  { label: 'Einstellungen', href: '/settings', icon: '⚙️' },
]

function Sidebar() {
  const location = useLocation()
  const isDarkMode = useThemeStore((state) => state.isDarkMode)
  const toggleDarkMode = useThemeStore((state) => state.toggleDarkMode)

  return (
    <aside className="sidebar">
      <div className="sidebar__header">
        <div className="sidebar__logo">RF</div>
      </div>

      <nav className="sidebar__nav">
        {navItems.map((item) => (
          <Link
            key={item.href}
            to={item.href}
            className={`sidebar__nav-item ${location.pathname === item.href ? 'sidebar__nav-item--active' : ''}`}
          >
            <span className="sidebar__nav-icon">{item.icon}</span>
            <span className="sidebar__nav-label">{item.label}</span>
          </Link>
        ))}
      </nav>

      <div className="sidebar__footer">
        <button
          className="sidebar__theme-toggle"
          onClick={toggleDarkMode}
          title={isDarkMode ? 'Light Mode' : 'Dark Mode'}
          style={{
            background: 'none',
            border: 'none',
            fontSize: '1.25rem',
            cursor: 'pointer',
            padding: 'var(--spacing-2)',
            borderRadius: 'var(--radius-md)',
            color: 'var(--color-text-secondary)',
            transition: 'all var(--transition-fast)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            width: '100%',
          }}
          onMouseEnter={(e) => (e.currentTarget.style.backgroundColor = 'var(--color-bg-secondary)')}
          onMouseLeave={(e) => (e.currentTarget.style.backgroundColor = 'transparent')}
        >
          {isDarkMode ? '☀️' : '🌙'}
        </button>
        <p className="sidebar__version">v1.0.0</p>
      </div>
    </aside>
  )
}

export default Sidebar
