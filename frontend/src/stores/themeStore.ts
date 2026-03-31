import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface ThemeStore {
  isDarkMode: boolean
  hasExplicitChoice: boolean
  toggleDarkMode: () => void
  setDarkMode: (isDark: boolean) => void
}

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set) => ({
      isDarkMode: true, // Dark is default
      hasExplicitChoice: false,

      toggleDarkMode: () =>
        set((state) => {
          const newMode = !state.isDarkMode
          applyTheme(newMode)
          return { isDarkMode: newMode, hasExplicitChoice: true }
        }),

      setDarkMode: (isDark: boolean) => {
        applyTheme(isDark)
        set({ isDarkMode: isDark, hasExplicitChoice: true })
      },
    }),
    {
      name: 'theme-storage',
      partialize: (state) => ({
        isDarkMode: state.isDarkMode,
        hasExplicitChoice: state.hasExplicitChoice,
      }),
    }
  )
)

function applyTheme(isDark: boolean) {
  const root = document.documentElement
  if (isDark) {
    root.style.colorScheme = 'dark'
    root.classList.remove('light-mode')
    root.classList.add('dark-mode')
  } else {
    root.style.colorScheme = 'light'
    root.classList.remove('dark-mode')
    root.classList.add('light-mode')
  }
}

// Apply theme on load with system preference detection
export function initializeTheme() {
  const state = useThemeStore.getState()

  // If user has never explicitly chosen, detect system preference
  if (!state.hasExplicitChoice) {
    const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
    useThemeStore.setState({ isDarkMode: prefersDark })
    applyTheme(prefersDark)
  } else {
    applyTheme(state.isDarkMode)
  }

  // Listen for system preference changes (only applies if user hasn't made explicit choice)
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
    const currentState = useThemeStore.getState()
    if (!currentState.hasExplicitChoice) {
      useThemeStore.setState({ isDarkMode: e.matches })
      applyTheme(e.matches)
    }
  })
}
