import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { authApi } from '../services/api'
import { Input } from '../components/Form/Input'
import './Login.scss'

function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(false)
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<{ email?: string; password?: string }>({})
  const [isLoading, setIsLoading] = useState(false)

  const validateForm = () => {
    const errors: typeof fieldErrors = {}

    if (!email) {
      errors.email = 'Email is required'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      errors.email = 'Please enter a valid email'
    }

    if (!password) {
      errors.password = 'Password is required'
    } else if (password.length < 6) {
      errors.password = 'Password must be at least 6 characters'
    }

    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!validateForm()) {
      return
    }

    setIsLoading(true)

    try {
      const response = await authApi.login(email, password)
      login(response.token, response.user)

      if (rememberMe) {
        localStorage.setItem('rememberedEmail', email)
      } else {
        localStorage.removeItem('rememberedEmail')
      }

      navigate('/')
    } catch (err: any) {
      const errorMessage = err?.response?.data?.message || err?.message || 'Login failed. Please check your credentials.'
      setError(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-card__header">
          <div className="login-card__logo">
            <span className="login-card__logo-icon">📦</span>
          </div>
          <h1 className="login-card__title">RentFlow</h1>
          <p className="login-card__subtitle">
            Professionelle Verwaltung von Veranstaltungsausstattung
          </p>
        </div>

        <form className="login-form" onSubmit={handleSubmit}>
          {error && (
            <div className="login-form__error" role="alert">
              {error}
            </div>
          )}

          <Input
            id="email"
            label="E-Mail-Adresse"
            type="email"
            value={email}
            onChange={(e) => {
              setEmail(e.target.value)
              if (fieldErrors.email) setFieldErrors({ ...fieldErrors, email: undefined })
            }}
            placeholder="admin@example.com"
            disabled={isLoading}
            error={fieldErrors.email}
            autoComplete="email"
            autoFocus
          />

          <Input
            id="password"
            label="Passwort"
            type="password"
            value={password}
            onChange={(e) => {
              setPassword(e.target.value)
              if (fieldErrors.password) setFieldErrors({ ...fieldErrors, password: undefined })
            }}
            placeholder="••••••••"
            disabled={isLoading}
            error={fieldErrors.password}
            autoComplete="current-password"
          />

          <div className="login-form__remember">
            <input
              id="remember-me"
              type="checkbox"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
              disabled={isLoading}
            />
            <label htmlFor="remember-me">E-Mail merken</label>
          </div>

          <button
            type="submit"
            className="login-form__submit"
            disabled={isLoading}
          >
            {isLoading ? 'Wird angemeldet...' : 'Anmelden'}
          </button>
        </form>

        <div className="login-card__footer">
          <div className="login-card__credentials">
            <p className="login-card__credentials-title">Demo-Anmeldedaten:</p>
            <p className="login-card__credentials-text">
              <strong>E-Mail:</strong> admin@example.com
            </p>
            <p className="login-card__credentials-text">
              <strong>Passwort:</strong> password
            </p>
          </div>
        </div>
      </div>
    </div>
  )
}

export default LoginPage
