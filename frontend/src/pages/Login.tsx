import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { authApi } from '../services/api'
import './Login.scss'

function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setIsLoading(true)

    try {
      const response = await authApi.login(email, password)
      login(response.token, response.user)
      navigate('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Login failed')
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-card__header">
          <h1 className="login-card__title">RentFlow</h1>
          <p className="login-card__subtitle">Event & Equipment Rental Management</p>
        </div>

        <form className="login-form" onSubmit={handleSubmit}>
          {error && <div className="login-form__error">{error}</div>}

          <div className="login-form__group">
            <label htmlFor="email" className="login-form__label">
              Email Address
            </label>
            <input
              id="email"
              type="email"
              className="login-form__input"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="you@example.com"
              required
              disabled={isLoading}
            />
          </div>

          <div className="login-form__group">
            <label htmlFor="password" className="login-form__label">
              Password
            </label>
            <input
              id="password"
              type="password"
              className="login-form__input"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="••••••••"
              required
              disabled={isLoading}
            />
          </div>

          <button
            type="submit"
            className="login-form__submit"
            disabled={isLoading}
          >
            {isLoading ? 'Logging in...' : 'Sign In'}
          </button>
        </form>

        <div className="login-card__footer">
          <p className="login-card__footer-text">
            Demo credentials: admin@example.com / password
          </p>
        </div>
      </div>
    </div>
  )
}

export default LoginPage
