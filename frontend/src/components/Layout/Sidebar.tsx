import { Link, useLocation } from 'react-router-dom'
import './Sidebar.scss'

interface NavItem {
  label: string
  href: string
  icon: string
}

const navItems: NavItem[] = [
  { label: 'Dashboard', href: '/', icon: '📊' },
  { label: 'Equipment', href: '/equipment', icon: '📦' },
  { label: 'Projects', href: '/projects', icon: '📋' },
  { label: 'Invoices', href: '/invoices', icon: '💰' },
  { label: 'Crew', href: '/crew', icon: '👥' },
  { label: 'Reports', href: '/reports', icon: '📈' },
]

function Sidebar() {
  const location = useLocation()

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
        <p className="sidebar__version">v1.0.0</p>
      </div>
    </aside>
  )
}

export default Sidebar
