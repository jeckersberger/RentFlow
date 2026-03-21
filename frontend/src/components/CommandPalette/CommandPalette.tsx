import { useEffect, useState, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import styles from './CommandPalette.module.scss'

interface Command {
  id: string
  name: string
  path: string
  group?: string
}

const commands: Command[] = [
  { id: 'dashboard', name: 'Dashboard', path: '/', group: 'Navigation' },
  { id: 'equipment', name: 'Equipment', path: '/equipment', group: 'Navigation' },
  { id: 'equipment-new', name: 'Neues Equipment', path: '/equipment/new', group: 'Navigation' },
  { id: 'projects', name: 'Projekte', path: '/projects', group: 'Navigation' },
  { id: 'projects-new', name: 'Neues Projekt', path: '/projects/new', group: 'Navigation' },
  { id: 'invoices', name: 'Rechnungen', path: '/invoices', group: 'Navigation' },
  { id: 'scanner', name: 'Scanner', path: '/scanner', group: 'Navigation' },
  { id: 'warehouse', name: 'Lager', path: '/warehouse', group: 'Navigation' },
  { id: 'settings', name: 'Einstellungen', path: '/settings', group: 'Navigation' },
]

// Simple fuzzy match function
function fuzzyMatch(search: string, text: string): boolean {
  const searchLower = search.toLowerCase()
  const textLower = text.toLowerCase()

  let searchIdx = 0
  for (let i = 0; i < textLower.length && searchIdx < searchLower.length; i++) {
    if (textLower[i] === searchLower[searchIdx]) {
      searchIdx++
    }
  }
  return searchIdx === searchLower.length
}

export function CommandPalette() {
  const [isOpen, setIsOpen] = useState(false)
  const [search, setSearch] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(0)
  const navigate = useNavigate()
  const inputRef = useRef<HTMLInputElement>(null)

  const filteredCommands = search
    ? commands.filter((cmd) => fuzzyMatch(search, cmd.name))
    : commands

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      // Cmd+K or Ctrl+K to open/close
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault()
        setIsOpen(!isOpen)
        setSearch('')
        setSelectedIndex(0)
      }

      // Only handle other keys if palette is open
      if (!isOpen) return

      switch (e.key) {
        case 'ArrowUp':
          e.preventDefault()
          setSelectedIndex((prev) =>
            prev > 0 ? prev - 1 : filteredCommands.length - 1
          )
          break
        case 'ArrowDown':
          e.preventDefault()
          setSelectedIndex((prev) =>
            prev < filteredCommands.length - 1 ? prev + 1 : 0
          )
          break
        case 'Enter':
          e.preventDefault()
          if (filteredCommands[selectedIndex]) {
            handleSelectCommand(filteredCommands[selectedIndex])
          }
          break
        case 'Escape':
          e.preventDefault()
          setIsOpen(false)
          break
      }
    }

    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isOpen, selectedIndex, filteredCommands])

  useEffect(() => {
    if (isOpen) {
      inputRef.current?.focus()
    }
  }, [isOpen])

  useEffect(() => {
    setSelectedIndex(0)
  }, [search])

  const handleSelectCommand = (command: Command) => {
    navigate(command.path)
    setIsOpen(false)
    setSearch('')
  }

  if (!isOpen) {
    return null
  }

  return (
    <div className={styles.backdrop} onClick={() => setIsOpen(false)}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <div className={styles.search}>
          <span className={styles.search__icon}>⌘</span>
          <input
            ref={inputRef}
            type="text"
            placeholder="Befehle suchen..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className={styles.search__input}
          />
          <span className={styles.search__hint}>ESC zum Schließen</span>
        </div>

        <div className={styles.list}>
          {filteredCommands.length === 0 ? (
            <div className={styles.empty}>Keine Befehle gefunden</div>
          ) : (
            filteredCommands.map((cmd, idx) => (
              <button
                key={cmd.id}
                className={`${styles.item} ${
                  idx === selectedIndex ? styles['item--selected'] : ''
                }`}
                onClick={() => handleSelectCommand(cmd)}
              >
                <span className={styles.item__name}>{cmd.name}</span>
                <span className={styles.item__hint}>
                  {cmd.group ? `${cmd.group}` : ''}
                </span>
              </button>
            ))
          )}
        </div>

        <div className={styles.footer}>
          <span className={styles.footer__text}>
            ↑↓ zum Navigieren • ↵ zum Wählen
          </span>
        </div>
      </div>
    </div>
  )
}
