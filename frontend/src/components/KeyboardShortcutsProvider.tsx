import { useKeyboardShortcuts } from '../hooks/useKeyboardShortcuts'

/**
 * Renders nothing — just activates the global keyboard shortcuts hook.
 * Must be placed inside <BrowserRouter>.
 */
export function KeyboardShortcutsProvider() {
  useKeyboardShortcuts()
  return null
}
