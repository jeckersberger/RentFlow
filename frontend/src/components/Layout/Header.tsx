import { useLocation } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useState } from 'react'
import './Header.scss'

const pageTitles: Record<string, string> = {
  '/': 'Dashboard',
  '/dashboard': 'Dashboard',
  '/equipment': 'Equipment',
  '/projects': 'Projekte',
  '/invoices': 'Rechnungen',
  '/scanner': 'Scanner',
  '/warehouse': 'Lager',
  '/transport': 'Transport',
  '/maintenance': 'Wartung',
  '/crew': 'Crew',
  '/documents': 'Dokumente',
  '/insurance': 'Versicherung',
  '/reports': 'Reports',
  '/ai': 'KI-Assistent',
  '/workflows': 'Workflows',
  '/federation': 'Federation',
  '/audit': 'Audit-Log',
  '/admin': 'Admin',
  '/settings': 'Einstellungen',
}

function getPageTitle(pathname: string): string {
  // Exact match first
  if (pageTitles[pathname]) return pageTitles[pathname]
  // Try matching the base path (e.g., /equipment/123 -> /equipment)
  const basePath = '/' + pathname.split('/')[1]
  return pageTitles[basePath] || 'RentFlow'
}

function Header() {
  const location = useLocation()
  const user = useAuthStore((state) => state.user)
  const [searchQuery, setSearchQuery] = useState('')

  const pageTitle = getPageTitle(location.pathname)

  const userInitials = user?.name
    ? user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
    : user?.email?.[0]?.toUpperCase() || '?'

  return (
    <header className="header">
      <div className="header__content">
        <div className="header__left">
          <h1 className="header__title">{pageTitle}</h1>
        </div>

        <div className="header__center">
          <div className="header__search">
            <span className="header__search-icon">🔍</span>
            <input
              type="text"
              className="header__search-input"
              placeholder="Suchen... (Ctrl+K)"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        <div className="header__right">
          <div className="header__user">
            <div className="header__user-avatar">{userInitials}</div>
            <div className="header__user-info">
              <p className="header__user-name">{user?.name || user?.email || 'Benutzer'}</p>
              <p className="header__user-email">{user?.email}</p>
            </div>
          </div>
        </div>
      </div>
    </header>
  )
}

export default Header
