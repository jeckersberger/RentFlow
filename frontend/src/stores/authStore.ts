import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface User {
  id?: string
  email: string
  name: string
}

interface AuthStore {
  token: string | null
  user: User | null
  tenantId: string | null
  isAuthenticated: boolean
  login: (token: string, user: User) => void
  logout: () => void
  setUser: (user: User) => void
  setToken: (token: string) => void
  setTenantId: (tenantId: string) => void
}

function decodeJWT(token: string): Record<string, unknown> {
  try {
    return JSON.parse(atob(token.split('.')[1]))
  } catch {
    return {}
  }
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      token: null,
      user: null,
      tenantId: null,
      isAuthenticated: false,

      login: (token: string, user: User) => {
        const payload = decodeJWT(token)
        const tenantId = (payload.tenant_id as string) || null
        set({
          token,
          user,
          tenantId,
          isAuthenticated: true,
        })
      },

      logout: () =>
        set({
          token: null,
          user: null,
          tenantId: null,
          isAuthenticated: false,
        }),

      setUser: (user: User) =>
        set({ user }),

      setToken: (token: string) =>
        set({ token }),

      setTenantId: (tenantId: string) =>
        set({ tenantId }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({
        token: state.token,
        user: state.user,
        tenantId: state.tenantId,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
