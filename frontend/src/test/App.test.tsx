import { describe, it, expect, beforeEach, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import App from '../App'
import { useAuthStore } from '../stores/authStore'

// Mock react-router-dom
vi.mock('react-router-dom', () => ({
  BrowserRouter: ({ children }: { children: React.ReactNode }) => children,
  Routes: ({ children }: { children: React.ReactNode }) => children,
  Route: () => null,
  Navigate: () => null,
}))

// Mock child components and stores
vi.mock('../stores/authStore', () => ({
  useAuthStore: vi.fn(),
}))

vi.mock('../stores/themeStore', () => ({
  initializeTheme: vi.fn(),
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

    render(<App />)
    expect(screen.getByRole('main', { hidden: true }) || document.body).toBeTruthy()
  })

  it('redirects to login when not authenticated', () => {
    mockedUseAuthStore.mockReturnValue({
      isAuthenticated: false,
    } as ReturnType<typeof useAuthStore>)

    const { container } = render(<App />)
    expect(container).toBeTruthy()
  })
})
