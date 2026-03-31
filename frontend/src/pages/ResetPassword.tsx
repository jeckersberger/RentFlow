import { useState } from 'react'
import { useNavigate, useParams, Link } from 'react-router-dom'
import { authApi } from '../services/api'
import { Input } from '../components/Form/Input'
import './Login.scss'

function ResetPasswordPage() {
  const navigate = useNavigate()
  const { token } = useParams<{ token: string }>()

  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<{
    newPassword?: string
    confirmPassword?: string
  }>({})
  const [success, setSuccess] = useState(false)
  const [isLoading, setIsLoading] = useState(false)

  const validateForm = () => {
    const errors: typeof fieldErrors = {}

    if (!newPassword) {
      errors.newPassword = 'Passwort ist erforderlich'
    } else if (newPassword.length < 8) {
      errors.newPassword = 'Passwort muss mindestens 8 Zeichen lang sein'
    }

    if (!confirmPassword) {
      errors.confirmPassword = 'Bitte Passwort bestaetigen'
    } else if (newPassword !== confirmPassword) {
      errors.confirmPassword = 'Passwoerter stimmen nicht ueberein'
    }

    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')

    if (!validateForm()) return

    if (!token) {
      setError('Kein gueltiger Reset-Token vorhanden.')
      return
    }

    setIsLoading(true)

    try {
      await authApi.resetPassword(token, newPassword)
      setSuccess(true)
    } catch (err: unknown) {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const errorResponse = (err as any)?.response
      let errorMessage =
        'Passwort konnte nicht zurueckgesetzt werden. Der Token ist ungueltig oder abgelaufen.'

      if (errorResponse?.data?.message) {
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
            <span className="login-card__logo-icon">🔒</span>
          </div>
          <h1 className="login-card__title">Neues Passwort</h1>
          <p className="login-card__subtitle">
            Geben Sie Ihr neues Passwort ein.
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
              Ihr Passwort wurde erfolgreich zurueckgesetzt. Sie koennen sich
              jetzt mit Ihrem neuen Passwort anmelden.
            </div>
            <button
              type="button"
              className="login-form__submit"
              onClick={() => navigate('/login')}
            >
              Zum Login
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
              id="new-password"
              label="Neues Passwort"
              type="password"
              value={newPassword}
              onChange={(e) => {
                setNewPassword(e.target.value)
                if (fieldErrors.newPassword)
                  setFieldErrors({ ...fieldErrors, newPassword: undefined })
              }}
              placeholder="Mindestens 8 Zeichen"
              disabled={isLoading}
              error={fieldErrors.newPassword}
              autoComplete="new-password"
              autoFocus
            />

            <Input
              id="confirm-password"
              label="Passwort bestaetigen"
              type="password"
              value={confirmPassword}
              onChange={(e) => {
                setConfirmPassword(e.target.value)
                if (fieldErrors.confirmPassword)
                  setFieldErrors({
                    ...fieldErrors,
                    confirmPassword: undefined,
                  })
              }}
              placeholder="Passwort wiederholen"
              disabled={isLoading}
              error={fieldErrors.confirmPassword}
              autoComplete="new-password"
            />

            <button
              type="submit"
              className="login-form__submit"
              disabled={isLoading}
            >
              {isLoading ? 'Wird gespeichert...' : 'Passwort zuruecksetzen'}
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

      <div className="login-version">CrateDesk</div>
    </div>
  )
}

export default ResetPasswordPage
