import { useEffect, useRef, useCallback } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'

// =============================================================================
// Types
// =============================================================================

export interface KeyboardShortcut {
  /** Unique id */
  id: string
  /** Key(s) the user must press – e.g. "ctrl+k", "g d" (sequence) */
  keys: string
  /** Human-readable label shown in the help modal */
  label: string
  /** Category for grouping in the help modal */
  category: ShortcutCategory
  /** The action to run */
  action: () => void
  /** When true the shortcut is only active on matching routes */
  when?: () => boolean
}

export type ShortcutCategory = 'navigation' | 'aktionen' | 'allgemein'

export const CATEGORY_LABELS: Record<ShortcutCategory, string> = {
  navigation: 'Navigation',
  aktionen: 'Aktionen',
  allgemein: 'Allgemein',
}

// =============================================================================
// Shortcut registry (singleton so other components can read it)
// =============================================================================

let _shortcuts: KeyboardShortcut[] = []
let _listeners: Array<() => void> = []

export function getShortcuts(): KeyboardShortcut[] {
  return _shortcuts
}

/** Subscribe to shortcut list changes (used by ShortcutsHelp). */
export function onShortcutsChange(fn: () => void) {
  _listeners.push(fn)
  return () => {
    _listeners = _listeners.filter((l) => l !== fn)
  }
}

function setShortcuts(next: KeyboardShortcut[]) {
  _shortcuts = next
  _listeners.forEach((fn) => fn())
}

// =============================================================================
// Event for opening the shortcuts help modal
// =============================================================================

export const SHORTCUTS_HELP_EVENT = 'rentflow:shortcuts-help'

export function openShortcutsHelp() {
  window.dispatchEvent(new CustomEvent(SHORTCUTS_HELP_EVENT))
}

// =============================================================================
// Hook
// =============================================================================

export function useKeyboardShortcuts() {
  const navigate = useNavigate()
  const location = useLocation()

  // Track pending "g" press for two-key navigation combos
  const pendingG = useRef(false)
  const pendingTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  const pathname = location.pathname

  // -------------------------------------------------------
  // Build the list of shortcuts (re-built when route changes)
  // -------------------------------------------------------
  const buildShortcuts = useCallback((): KeyboardShortcut[] => {
    const list: KeyboardShortcut[] = []

    // -- Allgemein ---------------------------------------------------------

    list.push({
      id: 'cmd-palette',
      keys: 'ctrl+k',
      label: 'Command Palette öffnen',
      category: 'allgemein',
      action: () => {}, // handled by CommandPalette itself
    })

    list.push({
      id: 'shortcuts-help',
      keys: 'ctrl+/',
      label: 'Tastenkürzel anzeigen',
      category: 'allgemein',
      action: () => openShortcutsHelp(),
    })

    list.push({
      id: 'shortcuts-help-qmark',
      keys: '?',
      label: 'Tastenkürzel anzeigen',
      category: 'allgemein',
      action: () => openShortcutsHelp(),
    })

    list.push({
      id: 'escape',
      keys: 'escape',
      label: 'Modal / Sidebar schließen',
      category: 'allgemein',
      action: () => {}, // handled natively by modals
    })

    // -- Navigation (G then X) --------------------------------------------

    const navShortcuts: Array<{ key: string; path: string; label: string }> = [
      { key: 'd', path: '/', label: 'Zum Dashboard' },
      { key: 'e', path: '/equipment', label: 'Zum Equipment' },
      { key: 'p', path: '/projects', label: 'Zu Projekten' },
      { key: 'i', path: '/invoices', label: 'Zu Rechnungen' },
      { key: 's', path: '/scanner', label: 'Zum Scanner' },
      { key: 'c', path: '/contacts', label: 'Zu Kontakten' },
      { key: 'w', path: '/warehouse', label: 'Zum Lager' },
    ]

    for (const ns of navShortcuts) {
      list.push({
        id: `nav-${ns.key}`,
        keys: `g ${ns.key}`,
        label: ns.label,
        category: 'navigation',
        action: () => navigate(ns.path),
      })
    }

    // -- Context-dependent actions ----------------------------------------

    // Ctrl+S — Save (dispatches a custom event that forms can listen to)
    list.push({
      id: 'save',
      keys: 'ctrl+s',
      label: 'Speichern',
      category: 'aktionen',
      action: () => {
        // Dispatch a custom save event so form components can handle it
        window.dispatchEvent(new CustomEvent('rentflow:save'))
      },
    })

    // Ctrl+N — context-dependent "new"
    if (pathname.startsWith('/equipment')) {
      list.push({
        id: 'new-equipment',
        keys: 'ctrl+n',
        label: 'Neues Equipment',
        category: 'aktionen',
        action: () => navigate('/equipment/new'),
        when: () => pathname.startsWith('/equipment'),
      })

      list.push({
        id: 'equipment-labels',
        keys: 'ctrl+l',
        label: 'Equipment-Labels',
        category: 'aktionen',
        action: () => navigate('/equipment/labels'),
        when: () => pathname.startsWith('/equipment'),
      })
    } else if (pathname.startsWith('/projects')) {
      list.push({
        id: 'new-project',
        keys: 'ctrl+n',
        label: 'Neues Projekt',
        category: 'aktionen',
        action: () => navigate('/projects/new'),
        when: () => pathname.startsWith('/projects'),
      })
    } else if (pathname.startsWith('/invoices')) {
      list.push({
        id: 'new-invoice',
        keys: 'ctrl+n',
        label: 'Neue Rechnung',
        category: 'aktionen',
        action: () => navigate('/invoices/new'),
        when: () => pathname.startsWith('/invoices'),
      })
    } else if (pathname.startsWith('/contacts')) {
      list.push({
        id: 'new-contact',
        keys: 'ctrl+n',
        label: 'Neuer Kontakt',
        category: 'aktionen',
        action: () => navigate('/contacts/new'),
        when: () => pathname.startsWith('/contacts'),
      })
    } else {
      // Generic fallback — open command palette for "new" action
      list.push({
        id: 'new-generic',
        keys: 'ctrl+n',
        label: 'Neu erstellen',
        category: 'aktionen',
        action: () => {
          // Trigger Ctrl+K programmatically to let user pick
          window.dispatchEvent(
            new KeyboardEvent('keydown', { key: 'k', ctrlKey: true, bubbles: true })
          )
        },
      })
    }

    return list
  }, [pathname, navigate])

  // -------------------------------------------------------
  // Register / unregister on mount & route change
  // -------------------------------------------------------
  useEffect(() => {
    const shortcuts = buildShortcuts()
    setShortcuts(shortcuts)

    const clearPendingG = () => {
      pendingG.current = false
      if (pendingTimer.current) {
        clearTimeout(pendingTimer.current)
        pendingTimer.current = null
      }
    }

    const handleKeyDown = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement
      const tagName = target.tagName.toLowerCase()
      const isInput = tagName === 'input' || tagName === 'textarea' || tagName === 'select' || target.isContentEditable

      // ---- Modifier combos (work even in inputs for Ctrl+S) ----

      const ctrl = e.ctrlKey || e.metaKey

      if (ctrl && e.key === '/') {
        e.preventDefault()
        openShortcutsHelp()
        return
      }

      // Ctrl+K is handled by CommandPalette — don't interfere
      if (ctrl && e.key === 'k') return

      if (ctrl && e.key === 's') {
        e.preventDefault()
        window.dispatchEvent(new CustomEvent('rentflow:save'))
        return
      }

      if (ctrl && e.key === 'n') {
        e.preventDefault()
        const ctxShortcut = shortcuts.find((s) => s.keys === 'ctrl+n')
        ctxShortcut?.action()
        return
      }

      if (ctrl && e.key === 'l' && pathname.startsWith('/equipment')) {
        e.preventDefault()
        navigate('/equipment/labels')
        return
      }

      // ---- Non-modifier shortcuts (skip when user is typing) ----

      if (isInput) return

      // "?" key for shortcuts help
      if (e.key === '?' && !ctrl && !e.altKey) {
        e.preventDefault()
        openShortcutsHelp()
        return
      }

      // Two-key navigation: G then <key>
      if (pendingG.current) {
        clearPendingG()
        const combo = `g ${e.key.toLowerCase()}`
        const match = shortcuts.find((s) => s.keys === combo)
        if (match) {
          e.preventDefault()
          match.action()
        }
        return
      }

      if (e.key === 'g' && !ctrl && !e.altKey && !e.shiftKey) {
        pendingG.current = true
        pendingTimer.current = setTimeout(clearPendingG, 1000)
        return
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      clearPendingG()
    }
  }, [buildShortcuts, pathname, navigate])
}

// =============================================================================
// Mapping helper: shortcut key hint for a given nav path
// =============================================================================

export const NAV_SHORTCUT_HINTS: Record<string, string> = {
  '/': 'G D',
  '/equipment': 'G E',
  '/projects': 'G P',
  '/invoices': 'G I',
  '/scanner': 'G S',
  '/contacts': 'G C',
  '/warehouse': 'G W',
}
