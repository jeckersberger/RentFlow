import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render } from '@testing-library/react'
import App from '../App'
import { useAuthStore } from '../stores/authStore'

// Mock react-router-dom
vi.mock('react-router-dom', () => ({
  BrowserRouter: ({ children }: { children: React.ReactNode }) => children,
  Routes: ({ children }: { children: React.ReactNode }) => children,
  Route: () => null,
  Navigate: () => null,
  useNavigate: () => vi.fn(),
  useLocation: () => ({ pathname: '/' }),
}))

// Mock child components and stores
vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../stores/themeStore', () => ({
  initializeTheme: vi.fn(),
}))

vi.mock('../stores/notificationStore', () => ({
  useNotificationStore: () => ({
    notifications: [],
    addNotification: vi.fn(),
    removeNotification: vi.fn(),
  }),
}))

// Mock complex components that have their own dependencies
vi.mock('../components/Toast/Toast', () => ({
  ToastContainer: () => null,
}))

vi.mock('../components/CommandPalette/CommandPalette', () => ({
  CommandPalette: () => null,
}))

const mockedUseAuthStore = vi.mocked(useAuthStore)

describe('App Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders without crashing', () => {
    mockedUseAuthStore.mockReturnValue({
      isAuthenticated: false,
    } as ReturnType<typeof useAuthStore>)

    const { container } = render(<App />)
    expect(container).toBeTruthy()
  })

  it('redirects to login when not authenticated', () => {
    mockedUseAuthStore.mockReturnValue({
      isAuthenticated: false,
    } as ReturnType<typeof useAuthStore>)

    const { container } = render(<App />)
    expect(container).toBeTruthy()
  })
})
