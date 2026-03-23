import { useEffect, useState, useCallback, useSyncExternalStore } from 'react'
import {
  getShortcuts,
  onShortcutsChange,
  SHORTCUTS_HELP_EVENT,
  CATEGORY_LABELS,
  type KeyboardShortcut,
  type ShortcutCategory,
} from '../../hooks/useKeyboardShortcuts'
import styles from './ShortcutsHelp.module.scss'

// =============================================================================
// Subscribe to the shortcut registry so we always show the latest list
// =============================================================================

function useShortcutList(): KeyboardShortcut[] {
  return useSyncExternalStore(onShortcutsChange, getShortcuts, getShortcuts)
}

// =============================================================================
// Render a key badge (e.g. "Ctrl" + "K")
// =============================================================================

function KeyBadge({ keys }: { keys: string }) {
  // "ctrl+k"  → ["Ctrl", "K"]
  // "g d"     → ["G", "dann", "D"]
  // "?"       → ["?"]
  // "escape"  → ["Esc"]

  const parts: string[] = []

  if (keys.includes('+')) {
    // Modifier combo
    for (const segment of keys.split('+')) {
      const s = segment.trim().toLowerCase()
      if (s === 'ctrl') parts.push('Ctrl')
      else if (s === 'meta') parts.push('Cmd')
      else if (s === 'shift') parts.push('Shift')
      else if (s === 'alt') parts.push('Alt')
      else parts.push(s.toUpperCase())
    }
  } else if (keys.includes(' ')) {
    // Sequence combo (g d)
    const [first, second] = keys.split(' ')
    parts.push(first.toUpperCase(), 'dann', second.toUpperCase())
  } else if (keys === 'escape') {
    parts.push('Esc')
  } else {
    parts.push(keys.toUpperCase())
  }

  return (
    <span className={styles.keys}>
      {parts.map((p, i) =>
        p === 'dann' ? (
          <span key={i} className={styles.keys__separator}>
            dann
          </span>
        ) : (
          <kbd key={i} className={styles.keys__badge}>
            {p}
          </kbd>
        )
      )}
    </span>
  )
}

// =============================================================================
// Component
// =============================================================================

const CATEGORY_ORDER: ShortcutCategory[] = ['navigation', 'aktionen', 'allgemein']

// Deduplicate by keys — keep first occurrence (so context-dependent wins)
function dedupeShortcuts(list: KeyboardShortcut[]): KeyboardShortcut[] {
  const seen = new Set<string>()
  const out: KeyboardShortcut[] = []
  for (const s of list) {
    // Skip the "?" duplicate of Ctrl+/
    if (s.id === 'shortcuts-help-qmark') continue
    const key = `${s.category}:${s.keys}`
    if (seen.has(key)) continue
    seen.add(key)
    out.push(s)
  }
  return out
}

export function ShortcutsHelp() {
  const [isOpen, setIsOpen] = useState(false)
  const shortcuts = useShortcutList()

  const close = useCallback(() => setIsOpen(false), [])

  // Listen for the custom open event
  useEffect(() => {
    const handleOpen = () => setIsOpen(true)
    window.addEventListener(SHORTCUTS_HELP_EVENT, handleOpen)
    return () => window.removeEventListener(SHORTCUTS_HELP_EVENT, handleOpen)
  }, [])

  // Close on Escape
  useEffect(() => {
    if (!isOpen) return
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        e.preventDefault()
        e.stopPropagation()
        close()
      }
    }
    window.addEventListener('keydown', handleKey, true)
    return () => window.removeEventListener('keydown', handleKey, true)
  }, [isOpen, close])

  if (!isOpen) return null

  const deduped = dedupeShortcuts(shortcuts)
  const grouped = CATEGORY_ORDER.map((cat) => ({
    category: cat,
    label: CATEGORY_LABELS[cat],
    items: deduped.filter((s) => s.category === cat),
  })).filter((g) => g.items.length > 0)

  return (
    <div className={styles.backdrop} onClick={close}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className={styles.header}>
          <h2 className={styles.header__title}>Tastenkürzel</h2>
          <button className={styles.header__close} onClick={close} title="Schließen">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>

        {/* Body */}
        <div className={styles.body}>
          {grouped.map((group) => (
            <div key={group.category} className={styles.section}>
              <h3 className={styles.section__title}>{group.label}</h3>
              <ul className={styles.section__list}>
                {group.items.map((s) => (
                  <li key={s.id} className={styles.row}>
                    <span className={styles.row__label}>{s.label}</span>
                    <KeyBadge keys={s.keys} />
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* Footer */}
        <div className={styles.footer}>
          <span className={styles.footer__hint}>
            Drücke <kbd className={styles.footer__kbd}>Esc</kbd> zum Schließen
          </span>
        </div>
      </div>
    </div>
  )
}
