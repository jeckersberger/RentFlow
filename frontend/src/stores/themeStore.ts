import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface ThemeStore {
  isDarkMode: boolean
  toggleDarkMode: () => void
  setDarkMode: (isDark: boolean) => void
}

export const useThemeStore = create<ThemeStore>()(
  persist(
    (set) => ({
      isDarkMode: false,

      toggleDarkMode: () =>
        set((state) => {
          const newMode = !state.isDarkMode
          applyTheme(newMode)
          return { isDarkMode: newMode }
        }),

      setDarkMode: (isDark: boolean) => {
        applyTheme(isDark)
        set({ isDarkMode: isDark })
      },
    }),
    {
      name: 'theme-storage',
      partialize: (state) => ({
        isDarkMode: state.isDarkMode,
      }),
    }
  )
)

function applyTheme(isDark: boolean) {
  const root = document.documentElement
  if (isDark) {
    root.style.colorScheme = 'dark'
    root.classList.add('dark-mode')
  } else {
    root.style.colorScheme = 'light'
    root.classList.remove('dark-mode')
  }
}

// Apply theme on load
export function initializeTheme() {
  const isDark = useThemeStore.getState().isDarkMode
  applyTheme(isDark)
}
