import { useState } from 'react'
import { useNavigate, Link } from 'react-router-dom'
import { authApi } from '../services/api'
import { Input } from '../components/Form/Input'
import './Login.scss'

function ForgotPasswordPage() {
  const navigate = useNavigate()
  const [email, setEmail] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!email) {
      setError('Bitte geben Sie Ihre E-Mail-Adresse ein.')
      return
    }

    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setError('Bitte geben Sie eine gueltige E-Mail-Adresse ein.')
      return
    }

    setIsLoading(true)

    try {
      await authApi.forgotPassword(email)
      setSuccess(true)
    } catch {
      // Always show success to prevent email enumeration
      setSuccess(true)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="login-page">
      <div className="login-card">
        <div className="login-card__header">
          <div className="login-card__logo">
            <span className="login-card__logo-icon">🔑</span>
          </div>
          <h1 className="login-card__title">Passwort vergessen</h1>
          <p className="login-card__subtitle">
            Geben Sie Ihre E-Mail-Adresse ein, um einen Reset-Link zu erhalten.
          </p>
        </div>

        {success ? (
          <div className="login-form" style={{ textAlign: 'center' }}>
            <div
              style={{
                padding: 'var(--spacing-4)',
                marginBottom: 'var(--spacing-4)',
                background: 'rgba(34, 197, 94, 0.08)',
                border: '1px solid rgba(34, 197, 94, 0.25)',
                borderRadius: 'var(--radius-md)',
                color: '#4ade80',
                fontSize: 'var(--font-size-sm)',
              }}
            >
              Falls ein Konto mit dieser E-Mail-Adresse existiert, wurde ein
              Reset-Token generiert. Bitte pruefen Sie die Server-Logs.
            </div>
            <button
              type="button"
              className="login-form__submit"
              onClick={() => navigate('/login')}
            >
              Zurueck zum Login
            </button>
          </div>
        ) : (
          <form className="login-form" onSubmit={handleSubmit}>
            {error && (
              <div className="login-form__error" role="alert">
                <strong>Fehler:</strong> {error}
              </div>
            )}

            <Input
              id="email"
              label="E-Mail-Adresse"
              type="email"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value)
                if (error) setError('')
              }}
              placeholder="ihre@email.com"
              disabled={isLoading}
              autoComplete="email"
              autoFocus
            />

            <button
              type="submit"
              className="login-form__submit"
              disabled={isLoading}
            >
              {isLoading ? 'Wird gesendet...' : 'Reset-Link anfordern'}
            </button>

            <div style={{ textAlign: 'center', marginTop: 'var(--spacing-4)' }}>
              <Link
                to="/login"
                style={{
                  color: 'var(--color-text-muted)',
                  fontSize: 'var(--font-size-sm)',
                  textDecoration: 'none',
                }}
              >
                Zurueck zum Login
              </Link>
            </div>
          </form>
        )}
      </div>

      <div className="login-version">RentFlow v1.0.0</div>
    </div>
  )
}

export default ForgotPasswordPage
