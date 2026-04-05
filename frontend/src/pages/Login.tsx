import { useState, useEffect, useRef } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'
import { authApi } from '../services/api'
import { Input } from '../components/Form/Input'
import './Login.scss'

function LoginPage() {
  const navigate = useNavigate()
  const login = useAuthStore((state) => state.login)
  const loginInputRef = useRef<HTMLInputElement>(null)

  const [loginId, setLoginId] = useState('')
  const [password, setPassword] = useState('')
  const [rememberMe, setRememberMe] = useState(false)
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<{ login?: string; password?: string }>({})
  const [isLoading, setIsLoading] = useState(false)

  useEffect(() => {
    const remembered = localStorage.getItem('rememberedLogin')
    if (remembered) {
      setLoginId(remembered)
      setRememberMe(true)
    }
    loginInputRef.current?.focus()
  }, [])

  const validateForm = () => {
    const errors: typeof fieldErrors = {}

    if (!loginId) {
      errors.login = 'Benutzername oder E-Mail ist erforderlich'
    }

    if (!password) {
      errors.password = 'Passwort ist erforderlich'
    } else if (password.length < 6) {
      errors.password = 'Passwort muss mindestens 6 Zeichen haben'
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
      // Backend accepts username or email in the "email" field
      const response = await authApi.login(loginId, password)
      login(response.token, response.refreshToken || '', response.user)

      // Try to fetch full user profile (non-blocking)
      try {
        const profile = await authApi.getCurrentUser()
        if (profile) {
          useAuthStore.getState().setUser({
            id: profile.id || profile.data?.id,
            email: profile.email || profile.data?.email || '',
            name: profile.name || profile.data?.name || profile.username || loginId,
          })
        }
      } catch {
        // Ignore - basic user info from login is sufficient
      }

      if (rememberMe) {
        localStorage.setItem('rememberedLogin', loginId)
      } else {
        localStorage.removeItem('rememberedLogin')
      }

      navigate('/')
    } catch (err: unknown) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const errorResponse = (err as any)?.response
      let errorMessage = 'Anmeldung fehlgeschlagen. Bitte überprüfen Sie Ihre Anmeldedaten.'

      if (errorResponse?.status === 401) {
        errorMessage = 'Ungültiger Benutzername/E-Mail oder Passwort.'
      } else if (errorResponse?.status === 429) {
        errorMessage = 'Zu viele Anmeldeversuche. Bitte warten Sie ein paar Minuten.'
      } else if (errorResponse?.data?.message) {
        errorMessage = errorResponse.data.message
      } else if ((err as Error)?.message) {
        errorMessage = (err as Error).message
      }

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
          <h1 className="login-card__title">CrateDesk</h1>
          <p className="login-card__tagline">Professionelle Vermietungssoftware</p>
          <p className="login-card__subtitle">
            Verwaltung von Veranstaltungsausstattung
          </p>
        </div>

        <form className="login-form" onSubmit={handleSubmit}>
          {error && (
            <div className="login-form__error" role="alert">
              <strong>Fehler:</strong> {error}
            </div>
          )}

          <Input
            id="login"
            ref={loginInputRef}
            label="Benutzername oder E-Mail"
            type="text"
            value={loginId}
            onChange={(e) => {
              setLoginId(e.target.value)
              if (fieldErrors.login) setFieldErrors({ ...fieldErrors, login: undefined })
            }}
            placeholder="jeck oder admin@example.com"
            disabled={isLoading}
            error={fieldErrors.login}
            autoComplete="username"
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

          <div className="login-form__options">
            <div className="login-form__remember">
              <input
                id="remember-me"
                type="checkbox"
                checked={rememberMe}
                onChange={(e) => setRememberMe(e.target.checked)}
                disabled={isLoading}
              />
              <label htmlFor="remember-me">Anmeldedaten merken</label>
            </div>
            <button
              type="button"
              className="login-form__forgot"
              onClick={() => navigate('/forgot-password')}
              disabled={isLoading}
            >
              Passwort vergessen?
            </button>
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
            <p className="login-card__credentials-title">Melden Sie sich mit Ihrem Benutzernamen oder E-Mail an.</p>
          </div>
        </div>
      </div>

      <div className="login-legal-links">
        <Link to="/impressum">Impressum</Link>
        <span className="login-legal-links__separator">|</span>
        <Link to="/datenschutz">Datenschutz</Link>
      </div>

      <div className="login-version">
        CrateDesk
      </div>
    </div>
  )
}

export default LoginPage
