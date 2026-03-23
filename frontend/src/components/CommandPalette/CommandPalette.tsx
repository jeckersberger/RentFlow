import { useEffect, useState, useRef, useCallback, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { equipmentApi, projectApi, invoiceApi, contactApi } from '../../services/api'
import styles from './CommandPalette.module.scss'

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface CommandItem {
  id: string
  title: string
  subtitle?: string
  icon: string
  category: Category
  path: string
  shortcut?: string
  keywords?: string
}

type Category =
  | 'recent'
  | 'navigation'
  | 'equipment'
  | 'projekte'
  | 'kontakte'
  | 'rechnungen'
  | 'aktionen'

const CATEGORY_LABELS: Record<Category, string> = {
  recent: 'Zuletzt besucht',
  navigation: 'Navigation',
  equipment: 'Equipment',
  projekte: 'Projekte',
  kontakte: 'Kontakte',
  rechnungen: 'Rechnungen',
  aktionen: 'Aktionen',
}

const CATEGORY_ORDER: Category[] = [
  'recent',
  'navigation',
  'equipment',
  'projekte',
  'kontakte',
  'rechnungen',
  'aktionen',
]

// ---------------------------------------------------------------------------
// Static data
// ---------------------------------------------------------------------------

const NAVIGATION_ITEMS: CommandItem[] = [
  { id: 'nav-dashboard', title: 'Dashboard', subtitle: 'Übersicht', icon: '📊', category: 'navigation', path: '/', shortcut: 'G D' },
  { id: 'nav-equipment', title: 'Equipment', subtitle: 'Alle Geräte', icon: '🎛️', category: 'navigation', path: '/equipment', shortcut: 'G E' },
  { id: 'nav-projects', title: 'Projekte', subtitle: 'Alle Projekte', icon: '📁', category: 'navigation', path: '/projects', shortcut: 'G P' },
  { id: 'nav-invoices', title: 'Rechnungen', subtitle: 'Alle Rechnungen', icon: '💰', category: 'navigation', path: '/invoices', shortcut: 'G I' },
  { id: 'nav-contacts', title: 'Kontakte', subtitle: 'CRM', icon: '👥', category: 'navigation', path: '/contacts', shortcut: 'G C' },
  { id: 'nav-calendar', title: 'Kalender', subtitle: 'Terminübersicht', icon: '📅', category: 'navigation', path: '/calendar' },
  { id: 'nav-quotes', title: 'Angebote', subtitle: 'Kostenvoranschläge', icon: '📝', category: 'navigation', path: '/quotes' },
  { id: 'nav-scanner', title: 'Scanner', subtitle: 'Barcode / QR', icon: '📷', category: 'navigation', path: '/scanner', shortcut: 'G S' },
  { id: 'nav-warehouse', title: 'Lager', subtitle: 'Lagerübersicht', icon: '🏭', category: 'navigation', path: '/warehouse', shortcut: 'G W' },
  { id: 'nav-transport', title: 'Transport', subtitle: 'Touren & Fahrzeuge', icon: '🚛', category: 'navigation', path: '/transport' },
  { id: 'nav-maintenance', title: 'Wartung', subtitle: 'Wartungspläne', icon: '🔧', category: 'navigation', path: '/maintenance' },
  { id: 'nav-crew', title: 'Crew', subtitle: 'Mitarbeiter', icon: '👷', category: 'navigation', path: '/crew' },
  { id: 'nav-timetracking', title: 'Zeiterfassung', subtitle: 'Stunden erfassen', icon: '⏱️', category: 'navigation', path: '/time-tracking' },
  { id: 'nav-documents', title: 'Dokumente', subtitle: 'Dateiverwaltung', icon: '📄', category: 'navigation', path: '/documents' },
  { id: 'nav-insurance', title: 'Versicherung', subtitle: 'Schadensmeldungen', icon: '🛡️', category: 'navigation', path: '/insurance' },
  { id: 'nav-reports', title: 'Berichte', subtitle: 'Auswertungen', icon: '📈', category: 'navigation', path: '/reports' },
  { id: 'nav-mail', title: 'E-Mail', subtitle: 'Posteingang', icon: '✉️', category: 'navigation', path: '/mail/inbox' },
  { id: 'nav-settings', title: 'Einstellungen', subtitle: 'Konfiguration', icon: '⚙️', category: 'navigation', path: '/settings' },
  { id: 'nav-workshop', title: 'Werkstatt', subtitle: 'Reparaturen', icon: '🛠️', category: 'navigation', path: '/workshop' },
]

const ACTION_ITEMS: CommandItem[] = [
  { id: 'act-new-equipment', title: 'Neues Equipment', subtitle: 'Equipment anlegen', icon: '➕', category: 'aktionen', path: '/equipment/new', shortcut: 'Ctrl+N' },
  { id: 'act-new-project', title: 'Neues Projekt', subtitle: 'Projekt anlegen', icon: '➕', category: 'aktionen', path: '/projects/new', shortcut: 'Ctrl+N' },
  { id: 'act-new-invoice', title: 'Neue Rechnung', subtitle: 'Rechnung erstellen', icon: '➕', category: 'aktionen', path: '/invoices/new', shortcut: 'Ctrl+N' },
  { id: 'act-new-quote', title: 'Neues Angebot', subtitle: 'Angebot erstellen', icon: '➕', category: 'aktionen', path: '/quotes/new' },
  { id: 'act-scanner', title: 'Scanner öffnen', subtitle: 'Barcode scannen', icon: '📷', category: 'aktionen', path: '/scanner' },
  { id: 'act-new-crew', title: 'Neues Crew-Mitglied', subtitle: 'Mitarbeiter anlegen', icon: '➕', category: 'aktionen', path: '/crew/new' },
  { id: 'act-new-document', title: 'Neues Dokument', subtitle: 'Dokument hochladen', icon: '📎', category: 'aktionen', path: '/documents/new' },
]

// ---------------------------------------------------------------------------
// Recent items (localStorage)
// ---------------------------------------------------------------------------

const RECENT_KEY = 'rentflow_recent_pages'
const MAX_RECENT = 5

function getRecentItems(): CommandItem[] {
  try {
    const raw = localStorage.getItem(RECENT_KEY)
    if (!raw) return []
    return JSON.parse(raw) as CommandItem[]
  } catch {
    return []
  }
}

function addRecentItem(item: CommandItem) {
  const recents = getRecentItems().filter((r) => r.path !== item.path)
  recents.unshift({ ...item, id: `recent-${item.id}`, category: 'recent' })
  if (recents.length > MAX_RECENT) recents.length = MAX_RECENT
  localStorage.setItem(RECENT_KEY, JSON.stringify(recents))
}

// ---------------------------------------------------------------------------
// Fuzzy / substring match
// ---------------------------------------------------------------------------

function matchesSearch(query: string, ...fields: (string | undefined)[]): boolean {
  const q = query.toLowerCase()
  return fields.some((f) => f && f.toLowerCase().includes(q))
}

// ---------------------------------------------------------------------------
// API data cache
// ---------------------------------------------------------------------------

interface CachedData {
  equipment: CommandItem[]
  projects: CommandItem[]
  contacts: CommandItem[]
  invoices: CommandItem[]
  loaded: boolean
}

let dataCache: CachedData = {
  equipment: [],
  projects: [],
  contacts: [],
  invoices: [],
  loaded: false,
}

async function loadSearchData(): Promise<CachedData> {
  if (dataCache.loaded) return dataCache

  try {
    const [eqRes, projRes, invRes, contRes] = await Promise.allSettled([
      equipmentApi.list({ limit: 200 }),
      projectApi.list(1, 200),
      invoiceApi.list(1, 200),
      contactApi.list({ limit: 200 }),
    ])

    // Equipment
    if (eqRes.status === 'fulfilled') {
      const items = Array.isArray(eqRes.value) ? eqRes.value : eqRes.value?.data || []
      dataCache.equipment = items.map((e: any) => ({
        id: `eq-${e.id}`,
        title: e.name,
        subtitle: [e.sku, e.barcode, e.status].filter(Boolean).join(' · '),
        icon: '🎛️',
        category: 'equipment' as Category,
        path: `/equipment/${e.id}`,
        keywords: [e.name, e.sku, e.barcode, e.description].filter(Boolean).join(' '),
      }))
    }

    // Projects
    if (projRes.status === 'fulfilled') {
      const items = Array.isArray(projRes.value) ? projRes.value : projRes.value?.items || projRes.value?.data || []
      dataCache.projects = items.map((p: any) => ({
        id: `proj-${p.id}`,
        title: p.name,
        subtitle: [p.client, p.status].filter(Boolean).join(' · '),
        icon: '📁',
        category: 'projekte' as Category,
        path: `/projects/${p.id}`,
        keywords: [p.name, p.client, p.number, p.location].filter(Boolean).join(' '),
      }))
    }

    // Invoices
    if (invRes.status === 'fulfilled') {
      const items = Array.isArray(invRes.value) ? invRes.value : invRes.value?.data || []
      dataCache.invoices = items.map((i: any) => ({
        id: `inv-${i.id}`,
        title: i.number || `Rechnung #${i.id}`,
        subtitle: [i.client_name, i.status, i.total ? `€${i.total.toLocaleString('de-DE')}` : ''].filter(Boolean).join(' · '),
        icon: '💰',
        category: 'rechnungen' as Category,
        path: `/invoices/${i.id}`,
        keywords: [i.number, i.client_name, i.project_name].filter(Boolean).join(' '),
      }))
    }

    // Contacts
    if (contRes.status === 'fulfilled') {
      const items = Array.isArray(contRes.value) ? contRes.value : contRes.value?.data || []
      dataCache.contacts = items.map((c: any) => ({
        id: `con-${c.id}`,
        title: c.name || [c.first_name, c.last_name].filter(Boolean).join(' '),
        subtitle: [c.company, c.email, c.type].filter(Boolean).join(' · '),
        icon: '👥',
        category: 'kontakte' as Category,
        path: `/contacts/${c.id}`,
        keywords: [c.name, c.first_name, c.last_name, c.company, c.email].filter(Boolean).join(' '),
      }))
    }

    dataCache.loaded = true
  } catch {
    // Silently fail — static items will still work
  }

  return dataCache
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function CommandPalette() {
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [debouncedSearch, setDebouncedSearch] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [apiItems, setApiItems] = useState<CachedData>(dataCache)
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)
  const listRef = useRef<HTMLDivElement>(null)

  // Debounce search input
  useEffect(() => {
    const timer = setTimeout(() => setDebouncedSearch(search), 120)
    return () => clearTimeout(timer)
  }, [search])

  // Load API data on first open
  useEffect(() => {
    if (isOpen && !dataCache.loaded) {
      setLoading(true)
      loadSearchData().then((data) => {
        setApiItems({ ...data })
        setLoading(false)
      })
    }
  }, [isOpen])

  // Build flat results list grouped by category
  const { flatResults, groupedResults } = useMemo(() => {
    const q = debouncedSearch.trim()
    const recents = getRecentItems()

    const allItems: CommandItem[] = [
      ...recents,
      ...NAVIGATION_ITEMS,
      ...apiItems.equipment,
      ...apiItems.projects,
      ...apiItems.contacts,
      ...apiItems.invoices,
      ...ACTION_ITEMS,
    ]

    const filtered = q
      ? allItems.filter((item) =>
          matchesSearch(q, item.title, item.subtitle, item.keywords, CATEGORY_LABELS[item.category])
        )
      : allItems

    // Group by category in defined order
    const grouped: { category: Category; label: string; items: CommandItem[] }[] = []
    const flat: CommandItem[] = []

    for (const cat of CATEGORY_ORDER) {
      const items = filtered.filter((i) => i.category === cat)
      if (items.length === 0) continue
      // When searching, skip recent (already mixed in with other categories)
      if (q && cat === 'recent') continue
      // Without search, limit recent to what we have and navigation to 6
      const limited = !q && cat === 'navigation' ? items.slice(0, 8) : items
      grouped.push({ category: cat, label: CATEGORY_LABELS[cat], items: limited })
      flat.push(...limited)
    }

    return { flatResults: flat, groupedResults: grouped }
  }, [debouncedSearch, apiItems])

  // Reset selected index on search change
  useEffect(() => {
    setSelectedIndex(0)
  }, [debouncedSearch])

  // Scroll selected item into view
  useEffect(() => {
    const listEl = listRef.current
    if (!listEl) return
    const selected = listEl.querySelector(`[data-index="${selectedIndex}"]`) as HTMLElement
    if (selected) {
      selected.scrollIntoView({ block: 'nearest' })
    }
  }, [selectedIndex])

  const close = useCallback(() => {
    setIsOpen(false)
    setSearch('')
    setDebouncedSearch('')
    setSelectedIndex(0)
  }, [])

  const handleSelect = useCallback(
    (item: CommandItem) => {
      // Track recent item (use original item, not the recent-prefixed version)
      const original = item.category === 'recent'
        ? { ...item, id: item.id.replace('recent-', ''), category: 'navigation' as Category }
        : item
      addRecentItem(original)
      navigate(item.path)
      close()
    },
    [navigate, close]
  )

  // Global keyboard handler
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Ctrl+K / Cmd+K to toggle
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        if (isOpen) {
          close()
        } else {
          setIsOpen(true)
        }
        return
      }

      if (!isOpen) return

      switch (e.key) {
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex((prev) =>
            prev > 0 ? prev - 1 : flatResults.length - 1
          )
          break
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex((prev) =>
            prev < flatResults.length - 1 ? prev + 1 : 0
          )
          break
        case 'Enter':
          e.preventDefault()
          if (flatResults[selectedIndex]) {
            handleSelect(flatResults[selectedIndex])
          }
          break
        case 'Escape':
          e.preventDefault()
          close()
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [isOpen, selectedIndex, flatResults, handleSelect, close])

  // Focus input on open
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [isOpen])

  if (!isOpen) return null

  let globalIndex = -1

  return (
    <div className={styles.backdrop} onClick={close}>
      <div className={styles.palette} onClick={(e) => e.stopPropagation()}>
        {/* Search bar */}
        <div className={styles.searchBar}>
          <svg className={styles.searchBar__icon} width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            ref={inputRef}
            type="text"
            className={styles.searchBar__input}
            placeholder="Suche nach Seiten, Equipment, Projekten, Rechnungen..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            autoComplete="off"
            spellCheck={false}
          />
          <kbd className={styles.searchBar__kbd}>ESC</kbd>
        </div>

        {/* Results */}
        <div className={styles.results} ref={listRef}>
          {loading && (
            <div className={styles.loading}>
              <span className={styles.loading__spinner} />
              Daten werden geladen...
            </div>
          )}

          {!loading && flatResults.length === 0 && (
            <div className={styles.empty}>
              <span className={styles.empty__icon}>🔍</span>
              <span>Keine Ergebnisse für &ldquo;{debouncedSearch}&rdquo;</span>
            </div>
          )}

          {!loading &&
            groupedResults.map((group) => (
              <div key={group.category} className={styles.group}>
                <div className={styles.group__header}>{group.label}</div>
                {group.items.map((item) => {
                  globalIndex++
                  const idx = globalIndex
                  return (
                    <button
                      key={item.id}
                      data-index={idx}
                      className={`${styles.item} ${idx === selectedIndex ? styles['item--active'] : ''}`}
                      onClick={() => handleSelect(item)}
                      onMouseEnter={() => setSelectedIndex(idx)}
                    >
                      <span className={styles.item__icon}>{item.icon}</span>
                      <div className={styles.item__content}>
                        <span className={styles.item__title}>{item.title}</span>
                        {item.subtitle && (
                          <span className={styles.item__subtitle}>{item.subtitle}</span>
                        )}
                      </div>
                      {item.shortcut && (
                        <kbd className={styles.item__shortcut}>{item.shortcut}</kbd>
                      )}
                      <svg className={styles.item__arrow} width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                        <polyline points="9 18 15 12 9 6" />
                      </svg>
                    </button>
                  )
                })}
              </div>
            ))}
        </div>

        {/* Footer */}
        <div className={styles.footer}>
          <div className={styles.footer__hints}>
            <span className={styles.footer__hint}>
              <kbd className={styles.footer__key}>↑</kbd>
              <kbd className={styles.footer__key}>↓</kbd>
              <span>Navigieren</span>
            </span>
            <span className={styles.footer__hint}>
              <kbd className={styles.footer__key}>↵</kbd>
              <span>Öffnen</span>
            </span>
            <span className={styles.footer__hint}>
              <kbd className={styles.footer__key}>ESC</kbd>
              <span>Schließen</span>
            </span>
          </div>
          <span className={styles.footer__count}>
            {flatResults.length} Ergebnis{flatResults.length !== 1 ? 'se' : ''}
          </span>
        </div>
      </div>
    </div>
  )
}
