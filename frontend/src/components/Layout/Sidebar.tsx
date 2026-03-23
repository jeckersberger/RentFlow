import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useThemeStore } from '../../stores/themeStore'
import { useModuleStore } from '../../stores/moduleStore'
import { useState, useMemo, useCallback, useRef, useEffect } from 'react'
import {
  LayoutDashboard,
  Calendar,
  FolderKanban,
  AlertTriangle,
  Package,
  Warehouse,
  ScanLine,
  Wrench,
  Users,
  Clock,
  Truck,
  FileText,
  Receipt,
  Contact,
  Inbox,
  Send,
  PenSquare,
  BarChart3,
  Bot,
  Settings,
  Globe,
  Shield,
  ChevronLeft,
  ChevronRight,
  Sun,
  Moon,
  LogOut,
  ChevronDown,
  ChevronUp,
  Menu,
  X,
  User,
  type LucideIcon,
} from 'lucide-react'
import { NAV_SHORTCUT_HINTS } from '../../hooks/useKeyboardShortcuts'
import './Sidebar.scss'

interface NavItem {
  label: string
  href: string
  icon: LucideIcon
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    label: 'Planen',
    items: [
      { label: 'Dashboard', href: '/', icon: LayoutDashboard },
      { label: 'Kalender', href: '/calendar', icon: Calendar },
      { label: 'Projekte', href: '/projects', icon: FolderKanban },
      { label: 'Engpässe', href: '/shortages', icon: AlertTriangle },
    ],
  },
  {
    label: 'Lager',
    items: [
      { label: 'Equipment', href: '/equipment', icon: Package },
      { label: 'Lager', href: '/warehouse', icon: Warehouse },
      { label: 'Scanner', href: '/scanner', icon: ScanLine },
      { label: 'Werkstatt', href: '/workshop', icon: Wrench },
    ],
  },
  {
    label: 'Team',
    items: [
      { label: 'Crew', href: '/crew', icon: Users },
      { label: 'Zeiterfassung', href: '/time-tracking', icon: Clock },
      { label: 'Transport', href: '/transport', icon: Truck },
    ],
  },
  {
    label: 'Finanzen',
    items: [
      { label: 'Angebote', href: '/quotes', icon: FileText },
      { label: 'Rechnungen', href: '/invoices', icon: Receipt },
      { label: 'Kontakte', href: '/contacts', icon: Contact },
    ],
  },
  {
    label: 'Kommunikation',
    items: [
      { label: 'Verfassen', href: '/mail/compose', icon: PenSquare },
      { label: 'Posteingang', href: '/mail/inbox', icon: Inbox },
      { label: 'Gesendet', href: '/mail/sent', icon: Send },
    ],
  },
  {
    label: 'Auswerten',
    items: [
      { label: 'Reports', href: '/reports', icon: BarChart3 },
      { label: 'KI-Assistent', href: '/ai', icon: Bot },
    ],
  },
  {
    label: 'Verwalten',
    items: [
      { label: 'Einstellungen', href: '/settings', icon: Settings },
      { label: 'Federation', href: '/federation', icon: Globe },
      { label: 'Audit-Log', href: '/audit', icon: Shield },
    ],
  },
]

// Mapping of module IDs to their nav item paths
const MODULE_NAV_PATHS: Record<string, string[]> = {
  warehouse: ['/', '/equipment', '/scanner', '/warehouse'],
  projects: ['/projects', '/calendar', '/shortages'],
  finance: ['/invoices', '/quotes', '/contacts'],
  team: ['/crew', '/time-tracking', '/transport'],
  communication: ['/mail/compose', '/mail/inbox', '/mail/sent'],
  workshop: ['/workshop'],
  ai: ['/ai'],
  federation: ['/federation'],
}

// Paths that are always visible regardless of module state
const ALWAYS_VISIBLE_PATHS = ['/', '/settings', '/reports', '/audit']

function isPathVisibleForModules(path: string, enabledModules: string[]): boolean {
  if (ALWAYS_VISIBLE_PATHS.includes(path)) return true
  for (const [moduleId, paths] of Object.entries(MODULE_NAV_PATHS)) {
    if (paths.includes(path)) {
      return enabledModules.includes(moduleId)
    }
  }
  // Paths not in any module are always visible (e.g., /admin, /workflows, etc.)
  return true
}

// Flat list of key items for mobile bottom nav (5 items: last one opens full sidebar overlay)
const mobileNavItems: NavItem[] = [
  { label: 'Dashboard', href: '/', icon: LayoutDashboard },
  { label: 'Equipment', href: '/equipment', icon: Package },
  { label: 'Scanner', href: '/scanner', icon: ScanLine },
  { label: 'Projekte', href: '/projects', icon: FolderKanban },
  // "Menu" item is handled separately in the component (opens overlay)
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
  const enabledModules = useModuleStore((state) => state.enabledModules)
  const modulesLoaded = useModuleStore((state) => state.loaded)
  const [collapsed, setCollapsed] = useState(false)
  const [collapsedGroups, setCollapsedGroups] = useState<Record<string, boolean>>({})
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)

  const closeMobileMenu = useCallback(() => setMobileMenuOpen(false), [])

  // Close user menu on outside click
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

  // Filter nav groups based on enabled modules
  const filteredNavGroups = useMemo(() => {
    // If modules haven't loaded yet, show all items as fallback
    if (!modulesLoaded) return navGroups

    return navGroups
      .map((group) => ({
        ...group,
        items: group.items.filter((item) =>
          isPathVisibleForModules(item.href, enabledModules)
        ),
      }))
      .filter((group) => group.items.length > 0)
  }, [enabledModules, modulesLoaded])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const toggleGroup = (label: string) => {
    setCollapsedGroups((prev) => ({
      ...prev,
      [label]: !prev[label],
    }))
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
            {collapsed ? <ChevronRight size={14} /> : <ChevronLeft size={14} />}
          </button>
        </div>

        <nav className="sidebar__nav">
          {filteredNavGroups.map((group) => {
            const isGroupCollapsed = collapsedGroups[group.label]
            return (
              <div key={group.label} className="sidebar__nav-group">
                {!collapsed && (
                  <button
                    className="sidebar__nav-group-label"
                    onClick={() => toggleGroup(group.label)}
                  >
                    <span>{group.label}</span>
                    {isGroupCollapsed ? <ChevronDown size={12} /> : <ChevronUp size={12} />}
                  </button>
                )}
                {!isGroupCollapsed && group.items.map((item) => {
                  const IconComponent = item.icon
                  const shortcutHint = NAV_SHORTCUT_HINTS[item.href]
                  return (
                    <Link
                      key={item.href}
                      to={item.href}
                      className={`sidebar__nav-item ${isActiveRoute(location.pathname, item.href) ? 'sidebar__nav-item--active' : ''}`}
                      title={collapsed ? item.label : undefined}
                    >
                      <span className="sidebar__nav-icon">
                        <IconComponent size={18} />
                      </span>
                      {!collapsed && <span className="sidebar__nav-label">{item.label}</span>}
                      {!collapsed && shortcutHint && (
                        <span className="sidebar__nav-shortcut">{shortcutHint}</span>
                      )}
                    </Link>
                  )
                })}
              </div>
            )
          })}
        </nav>

        <div className="sidebar__footer">
          <button
            className="sidebar__theme-toggle"
            onClick={toggleDarkMode}
            title={isDarkMode ? 'Hellmodus aktivieren' : 'Dunkelmodus aktivieren'}
          >
            {isDarkMode ? <Sun size={18} /> : <Moon size={18} />}
            {!collapsed && (
              <span className="sidebar__theme-label">
                {isDarkMode ? 'Hellmodus' : 'Dunkelmodus'}
              </span>
            )}
          </button>

          <div className="sidebar__user-section" ref={userMenuRef}>
            <button
              className="sidebar__user-trigger"
              onClick={() => setUserMenuOpen(!userMenuOpen)}
              title={collapsed ? (user?.name || 'Benutzer') : undefined}
            >
              <div className="sidebar__user-avatar">{userInitials}</div>
              {!collapsed && (
                <div className="sidebar__user-details">
                  <p className="sidebar__user-name">{user?.name || 'Benutzer'}</p>
                  <p className="sidebar__user-email">{user?.email}</p>
                </div>
              )}
            </button>
            {userMenuOpen && (
              <div className={`sidebar__user-menu ${collapsed ? 'sidebar__user-menu--collapsed' : ''}`}>
                <button
                  className="sidebar__user-menu-item"
                  onClick={() => { setUserMenuOpen(false); navigate('/profile') }}
                >
                  <User size={16} />
                  <span>Profil</span>
                </button>
                <div className="sidebar__user-menu-divider" />
                <button
                  className="sidebar__user-menu-item sidebar__user-menu-item--danger"
                  onClick={() => { setUserMenuOpen(false); handleLogout() }}
                >
                  <LogOut size={16} />
                  <span>Abmelden</span>
                </button>
              </div>
            )}
          </div>

          {!collapsed && <p className="sidebar__version">v1.0.0</p>}
        </div>
      </aside>

      {/* Mobile sidebar overlay */}
      {mobileMenuOpen && (
        <div className="mobile-overlay" onClick={closeMobileMenu}>
          <aside className="mobile-overlay__sidebar" onClick={(e) => e.stopPropagation()}>
            <div className="mobile-overlay__header">
              <div className="sidebar__logo">
                <div className="sidebar__logo-icon">RF</div>
                <span className="sidebar__logo-text">RentFlow</span>
              </div>
              <button className="mobile-overlay__close" onClick={closeMobileMenu}>
                <X size={20} />
              </button>
            </div>

            <nav className="mobile-overlay__nav">
              {filteredNavGroups.map((group) => (
                <div key={group.label} className="mobile-overlay__group">
                  <span className="mobile-overlay__group-label">{group.label}</span>
                  {group.items.map((item) => {
                    const IconComponent = item.icon
                    return (
                      <Link
                        key={item.href}
                        to={item.href}
                        className={`mobile-overlay__nav-item ${isActiveRoute(location.pathname, item.href) ? 'mobile-overlay__nav-item--active' : ''}`}
                        onClick={closeMobileMenu}
                      >
                        <IconComponent size={18} />
                        <span>{item.label}</span>
                      </Link>
                    )
                  })}
                </div>
              ))}
            </nav>

            <div className="mobile-overlay__footer">
              <button className="mobile-overlay__theme-btn" onClick={toggleDarkMode}>
                {isDarkMode ? <Sun size={18} /> : <Moon size={18} />}
                <span>{isDarkMode ? 'Light Mode' : 'Dark Mode'}</span>
              </button>
              <div className="mobile-overlay__user">
                <div className="sidebar__user-avatar">{userInitials}</div>
                <div className="sidebar__user-details">
                  <p className="sidebar__user-name">{user?.name || 'Benutzer'}</p>
                  <p className="sidebar__user-email">{user?.email}</p>
                </div>
              </div>
              <button
                className="mobile-overlay__logout"
                onClick={() => { closeMobileMenu(); handleLogout(); }}
              >
                <LogOut size={16} />
                <span>Abmelden</span>
              </button>
            </div>
          </aside>
        </div>
      )}

      {/* Mobile bottom nav */}
      <nav className="mobile-nav">
        {mobileNavItems.map((item) => {
          const IconComponent = item.icon
          return (
            <Link
              key={item.href}
              to={item.href}
              className={`mobile-nav__item ${isActiveRoute(location.pathname, item.href) ? 'mobile-nav__item--active' : ''}`}
            >
              <span className="mobile-nav__icon">
                <IconComponent size={20} />
              </span>
              <span className="mobile-nav__label">{item.label}</span>
            </Link>
          )
        })}
        {/* Menu / Hamburger button */}
        <button
          className={`mobile-nav__item ${mobileMenuOpen ? 'mobile-nav__item--active' : ''}`}
          onClick={() => setMobileMenuOpen(true)}
        >
          <span className="mobile-nav__icon">
            <Menu size={20} />
          </span>
          <span className="mobile-nav__label">Menü</span>
        </button>
      </nav>
    </>
  )
}

export default Sidebar
