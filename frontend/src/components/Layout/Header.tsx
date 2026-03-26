import { useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useState, useRef, useEffect } from 'react'
import { NotificationCenter } from '../NotificationCenter/NotificationCenter'
import { User, LogOut, ChevronDown, QrCode } from 'lucide-react'
import { QRLoginModal } from '../QRLoginModal/QRLoginModal'
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
  '/profile': 'Profil',
}

function getPageTitle(pathname: string): string {
  // Exact match first
  if (pageTitles[pathname]) return pageTitles[pathname]
  // Try matching the base path (e.g., /equipment/123 -> /equipment)
  const basePath = '/' + pathname.split('/')[1]
  return pageTitles[basePath] || 'EquipFlow'
}

function Header() {
  const location = useLocation()
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const [searchQuery, setSearchQuery] = useState('')
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const [qrModalOpen, setQrModalOpen] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)

  const pageTitle = getPageTitle(location.pathname)

  const userInitials = user?.name
    ? user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
    : user?.email?.[0]?.toUpperCase() || '?'

  // Close menu on outside click
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setUserMenuOpen(false)
      }
    }
    if (userMenuOpen) {
      document.addEventListener('mousedown', handleClickOutside)
      return () => document.removeEventListener('mousedown', handleClickOutside)
    }
  }, [userMenuOpen])

  const handleLogout = () => {
    setUserMenuOpen(false)
    logout()
    navigate('/login')
  }

  return (
    <>
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
          <NotificationCenter />
          <div className="header__user-wrapper" ref={userMenuRef}>
            <button
              className="header__user"
              onClick={() => setUserMenuOpen(!userMenuOpen)}
            >
              <div className="header__user-avatar">{userInitials}</div>
              <div className="header__user-info">
                <p className="header__user-name">{user?.name || user?.email || 'Benutzer'}</p>
                <p className="header__user-email">{user?.email}</p>
              </div>
              <ChevronDown size={14} className={`header__user-chevron ${userMenuOpen ? 'header__user-chevron--open' : ''}`} />
            </button>
            {userMenuOpen && (
              <div className="header__user-menu">
                <button
                  className="header__user-menu-item"
                  onClick={() => { setUserMenuOpen(false); navigate('/profile') }}
                >
                  <User size={16} />
                  <span>Profil</span>
                </button>
                <button
                  className="header__user-menu-item"
                  onClick={() => { setUserMenuOpen(false); setQrModalOpen(true) }}
                >
                  <QrCode size={16} />
                  <span>QR-Code fuer Scanner</span>
                </button>
                <div className="header__user-menu-divider" />
                <button
                  className="header__user-menu-item header__user-menu-item--danger"
                  onClick={handleLogout}
                >
                  <LogOut size={16} />
                  <span>Abmelden</span>
                </button>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
    <QRLoginModal isOpen={qrModalOpen} onClose={() => setQrModalOpen(false)} />
    </>
  )
}

export default Header
