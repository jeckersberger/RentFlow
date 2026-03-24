import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import { useThemeStore } from '../../stores/themeStore'
import { useModuleStore } from '../../stores/moduleStore'
import React, { useState, useMemo, useCallback, useRef, useEffect } from 'react'
import { useTranslation } from 'react-i18next'
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
  CreditCard,
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
import { useBadgeCounts, useMarkSectionSeen } from '../../hooks/useBadgeCounts'
import './Sidebar.scss'

interface NavItem {
  labelKey: string
  href: string
  icon: LucideIcon
}

interface NavGroup {
  labelKey: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    labelKey: 'nav.planning',
    items: [
      { labelKey: 'nav.dashboard', href: '/', icon: LayoutDashboard },
      { labelKey: 'nav.calendar', href: '/calendar', icon: Calendar },
      { labelKey: 'nav.projects', href: '/projects', icon: FolderKanban },
      { labelKey: 'nav.shortages', href: '/shortages', icon: AlertTriangle },
    ],
  },
  {
    labelKey: 'nav.warehouse',
    items: [
      { labelKey: 'nav.equipment', href: '/equipment', icon: Package },
      { labelKey: 'nav.warehouse_nav', href: '/warehouse', icon: Warehouse },
      { labelKey: 'nav.scanner', href: '/scanner', icon: ScanLine },
      { labelKey: 'nav.workshop', href: '/workshop', icon: Wrench },
    ],
  },
  {
    labelKey: 'nav.team',
    items: [
      { labelKey: 'nav.crew', href: '/crew', icon: Users },
      { labelKey: 'nav.time_tracking', href: '/time-tracking', icon: Clock },
      { labelKey: 'nav.transport', href: '/transport', icon: Truck },
    ],
  },
  {
    labelKey: 'nav.finance',
    items: [
      { labelKey: 'nav.quotes', href: '/quotes', icon: FileText },
      { labelKey: 'nav.invoices', href: '/invoices', icon: Receipt },
      { labelKey: 'nav.expenses', href: '/expenses', icon: CreditCard },
      { labelKey: 'nav.contacts', href: '/contacts', icon: Contact },
    ],
  },
  {
    labelKey: 'nav.communication',
    items: [
      { labelKey: 'nav.compose', href: '/mail/compose', icon: PenSquare },
      { labelKey: 'nav.inbox', href: '/mail/inbox', icon: Inbox },
      { labelKey: 'nav.sent', href: '/mail/sent', icon: Send },
    ],
  },
  {
    labelKey: 'nav.analytics',
    items: [
      { labelKey: 'nav.reports', href: '/reports', icon: BarChart3 },
      { labelKey: 'nav.ai_assistant', href: '/ai', icon: Bot },
    ],
  },
  {
    labelKey: 'nav.admin',
    items: [
      { labelKey: 'nav.settings', href: '/settings', icon: Settings },
      { labelKey: 'nav.federation', href: '/federation', icon: Globe },
      { labelKey: 'nav.audit_log', href: '/audit', icon: Shield },
    ],
  },
]

// Mapping of module IDs to their nav item paths
const MODULE_NAV_PATHS: Record<string, string[]> = {
  warehouse: ['/', '/equipment', '/scanner', '/warehouse'],
  projects: ['/projects', '/calendar', '/shortages'],
  finance: ['/invoices', '/quotes', '/contacts', '/expenses'],
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
  { labelKey: 'nav.dashboard', href: '/', icon: LayoutDashboard },
  { labelKey: 'nav.equipment', href: '/equipment', icon: Package },
  { labelKey: 'nav.scanner', href: '/scanner', icon: ScanLine },
  { labelKey: 'nav.projects', href: '/projects', icon: FolderKanban },
  // "Menu" item is handled separately in the component (opens overlay)
]

function isActiveRoute(pathname: string, href: string): boolean {
  if (href === '/') return pathname === '/'
  return pathname === href || pathname.startsWith(href + '/')
}

function Sidebar() {
  const { t } = useTranslation()
  const location = useLocation()
  const navigate = useNavigate()
  const isDarkMode = useThemeStore((state) => state.isDarkMode)
  const toggleDarkMode = useThemeStore((state) => state.toggleDarkMode)
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)
  const enabledModules = useModuleStore((state) => state.enabledModules)
  const modulesLoaded = useModuleStore((state) => state.loaded)
  const badgeCounts = useBadgeCounts()
  const markSeen = useMarkSectionSeen()
  const [collapsed, setCollapsed] = useState(false)
  const [collapsedGroups, setCollapsedGroups] = useState<Record<string, boolean>>({})
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const [userMenuOpen, setUserMenuOpen] = useState(false)
  const userMenuRef = useRef<HTMLDivElement>(null)

  // Map nav paths to badge counts (new items + urgent items) and their section keys for markSeen
  const sectionKeyMap: Record<string, string> = {
    '/projects': 'projects',
    '/equipment': 'equipment',
    '/invoices': 'invoices',
    '/contacts': 'contacts',
    '/crew': 'crew',
    '/mail/inbox': 'mail',
    '/workshop': 'maintenance',
    '/audit': 'audit',
  }

  // Primary badge: new items since last visit (info/blue)
  // Secondary badge: urgent items that always show (danger/red)
  const badgeMap: Record<string, {
    count: number; variant: 'danger' | 'info' | 'warning';
    urgentCount?: number; urgentVariant?: 'danger' | 'warning';
  }> = {
    '/projects': { count: badgeCounts.projects, variant: 'info' },
    '/equipment': { count: badgeCounts.equipment, variant: 'info' },
    '/invoices': {
      count: badgeCounts.invoices,
      variant: 'info',
      urgentCount: badgeCounts.invoicesOverdue,
      urgentVariant: 'danger',
    },
    '/contacts': { count: badgeCounts.contacts, variant: 'info' },
    '/crew': { count: badgeCounts.crew, variant: badgeCounts.crew > 0 ? 'warning' : 'info' },
    '/mail/inbox': { count: badgeCounts.mailUnread, variant: 'info' },
    '/workshop': {
      count: badgeCounts.maintenance,
      variant: 'info',
      urgentCount: badgeCounts.maintenanceOverdue,
      urgentVariant: 'danger',
    },
    '/audit': { count: badgeCounts.audit, variant: 'info' },
  }

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

  const toggleGroup = (labelKey: string) => {
    setCollapsedGroups((prev) => ({
      ...prev,
      [labelKey]: !prev[labelKey],
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
            title={collapsed ? t('nav.expand_sidebar') : t('nav.collapse_sidebar')}
          >
            {collapsed ? <ChevronRight size={14} /> : <ChevronLeft size={14} />}
          </button>
        </div>

        <nav className="sidebar__nav">
          {filteredNavGroups.map((group) => {
            const isGroupCollapsed = collapsedGroups[group.labelKey]
            return (
              <div key={group.labelKey} className="sidebar__nav-group">
                {!collapsed && (
                  <button
                    className="sidebar__nav-group-label"
                    onClick={() => toggleGroup(group.labelKey)}
                  >
                    <span>{t(group.labelKey)}</span>
                    {isGroupCollapsed ? <ChevronDown size={12} /> : <ChevronUp size={12} />}
                  </button>
                )}
                {!isGroupCollapsed && group.items.map((item) => {
                  const IconComponent = item.icon
                  const shortcutHint = NAV_SHORTCUT_HINTS[item.href]
                  const label = t(item.labelKey)
                  const badge = badgeMap[item.href]
                  const sectionKey = sectionKeyMap[item.href]
                  const hasBadge = badge && (badge.count > 0 || (badge.urgentCount && badge.urgentCount > 0))
                  return (
                    <Link
                      key={item.href}
                      to={item.href}
                      className={`sidebar__nav-item ${isActiveRoute(location.pathname, item.href) ? 'sidebar__nav-item--active' : ''}`}
                      title={collapsed ? label : undefined}
                      onClick={() => sectionKey && markSeen(sectionKey)}
                    >
                      <span className="sidebar__nav-icon">
                        <IconComponent size={18} />
                        {collapsed && hasBadge && (
                          <span className={`sidebar__badge sidebar__badge--${badge.urgentCount && badge.urgentCount > 0 ? (badge.urgentVariant || 'danger') : badge.variant} sidebar__badge--dot`} />
                        )}
                      </span>
                      {!collapsed && <span className="sidebar__nav-label">{label}</span>}
                      {!collapsed && badge?.urgentCount != null && badge.urgentCount > 0 && (
                        <span className={`sidebar__badge sidebar__badge--${badge.urgentVariant || 'danger'}`}>
                          {badge.urgentCount}
                        </span>
                      )}
                      {!collapsed && badge && badge.count > 0 && (
                        <span className={`sidebar__badge sidebar__badge--${badge.variant}`}>
                          {badge.count}
                        </span>
                      )}
                      {!collapsed && shortcutHint && !hasBadge && (
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
            title={isDarkMode ? t('nav.light_mode') : t('nav.dark_mode')}
          >
            {isDarkMode ? <Sun size={18} /> : <Moon size={18} />}
            {!collapsed && (
              <span className="sidebar__theme-label">
                {isDarkMode ? t('nav.light_mode') : t('nav.dark_mode')}
              </span>
            )}
          </button>

          <div className="sidebar__user-section" ref={userMenuRef}>
            <button
              className="sidebar__user-trigger"
              onClick={() => setUserMenuOpen(!userMenuOpen)}
              title={collapsed ? (user?.name || t('nav.profile')) : undefined}
            >
              <div className="sidebar__user-avatar">{userInitials}</div>
              {!collapsed && (
                <div className="sidebar__user-details">
                  <p className="sidebar__user-name">{user?.name || t('nav.profile')}</p>
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
                  <span>{t('nav.profile')}</span>
                </button>
                <div className="sidebar__user-menu-divider" />
                <button
                  className="sidebar__user-menu-item sidebar__user-menu-item--danger"
                  onClick={() => { setUserMenuOpen(false); handleLogout() }}
                >
                  <LogOut size={16} />
                  <span>{t('nav.logout')}</span>
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
                <div key={group.labelKey} className="mobile-overlay__group">
                  <span className="mobile-overlay__group-label">{t(group.labelKey)}</span>
                  {group.items.map((item) => {
                    const IconComponent = item.icon
                    const badge = badgeMap[item.href]
                    const sectionKey = sectionKeyMap[item.href]
                    return (
                      <Link
                        key={item.href}
                        to={item.href}
                        className={`mobile-overlay__nav-item ${isActiveRoute(location.pathname, item.href) ? 'mobile-overlay__nav-item--active' : ''}`}
                        onClick={() => { if (sectionKey) markSeen(sectionKey); closeMobileMenu() }}
                      >
                        <IconComponent size={18} />
                        <span>{t(item.labelKey)}</span>
                        {badge?.urgentCount != null && badge.urgentCount > 0 && (
                          <span className={`sidebar__badge sidebar__badge--${badge.urgentVariant || 'danger'}`}>
                            {badge.urgentCount}
                          </span>
                        )}
                        {badge && badge.count > 0 && (
                          <span className={`sidebar__badge sidebar__badge--${badge.variant}`}>
                            {badge.count}
                          </span>
                        )}
                      </Link>
                    )
                  })}
                </div>
              ))}
            </nav>

            <div className="mobile-overlay__footer">
              <button className="mobile-overlay__theme-btn" onClick={toggleDarkMode}>
                {isDarkMode ? <Sun size={18} /> : <Moon size={18} />}
                <span>{isDarkMode ? t('nav.light_mode') : t('nav.dark_mode')}</span>
              </button>
              <div className="mobile-overlay__user">
                <div className="sidebar__user-avatar">{userInitials}</div>
                <div className="sidebar__user-details">
                  <p className="sidebar__user-name">{user?.name || t('nav.profile')}</p>
                  <p className="sidebar__user-email">{user?.email}</p>
                </div>
              </div>
              <button
                className="mobile-overlay__logout"
                onClick={() => { closeMobileMenu(); handleLogout(); }}
              >
                <LogOut size={16} />
                <span>{t('nav.logout')}</span>
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
              <span className="mobile-nav__label">{t(item.labelKey)}</span>
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
          <span className="mobile-nav__label">{t('nav.menu')}</span>
        </button>
      </nav>
    </>
  )
}

export default React.memo(Sidebar)
