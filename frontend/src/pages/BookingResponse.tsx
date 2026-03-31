import { useState, useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { bookingApi } from '../services/api'
import './BookingResponse.scss'

interface BookingDetails {
  project_name: string
  project_dates: string
  role: string
  location: string
  message: string
  freelancer_name: string
  status: string
}

function BookingResponse() {
  const { token } = useParams<{ token: string }>()
  const [details, setDetails] = useState<BookingDetails | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [responseStatus, setResponseStatus] = useState<string | null>(null)
  const [alternativeMessage, setAlternativeMessage] = useState('')
  const [showAlternative, setShowAlternative] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  const [submitted, setSubmitted] = useState(false)

  useEffect(() => {
    if (!token) return
    bookingApi.getDetails(token)
      .then((data) => {
        // Handle possible envelope wrapping
        const d = data?.data || data
        setDetails(d)
        if (d.status !== 'pending') {
          setSubmitted(true)
          setResponseStatus(d.status)
        }
      })
      .catch(() => {
        setError('Buchungsanfrage nicht gefunden oder abgelaufen.')
      })
      .finally(() => setLoading(false))
  }, [token])

  const handleRespond = async (status: string, message?: string) => {
    if (!token) return
    setSubmitting(true)
    try {
      await bookingApi.respond(token, { status, message: message || '' })
      setResponseStatus(status)
      setSubmitted(true)
    } catch {
      setError('Fehler beim Senden der Antwort. Bitte versuchen Sie es erneut.')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="booking-page">
        <div className="booking-card">
          <div className="booking-card__loading">Wird geladen...</div>
        </div>
      </div>
    )
  }

  if (error && !details) {
    return (
      <div className="booking-page">
        <div className="booking-card">
          <div className="booking-card__header">
            <h1 className="booking-card__title">RentFlow</h1>
            <p className="booking-card__subtitle">Buchungsanfrage</p>
          </div>
          <div className="booking-card__body">
            <div className="booking-card__error">{error}</div>
          </div>
        </div>
      </div>
    )
  }

  if (submitted) {
    const statusLabels: Record<string, string> = {
      accepted: 'Zugesagt',
      declined: 'Abgesagt',
      alternative: 'Alternative vorgeschlagen',
    }

    return (
      <div className="booking-page">
        <div className="booking-card">
          <div className="booking-card__header">
            <h1 className="booking-card__title">RentFlow</h1>
            <p className="booking-card__subtitle">Buchungsanfrage</p>
          </div>
          <div className="booking-card__body">
            <div className="booking-card__success">
              <div className="booking-card__success-icon">
                {responseStatus === 'accepted' ? '\u2705' : responseStatus === 'declined' ? '\u274C' : '\u{1F4AC}'}
              </div>
              <h2>Antwort erfasst</h2>
              <p>
                Status: <strong>{statusLabels[responseStatus || ''] || responseStatus}</strong>
              </p>
              <p className="booking-card__success-detail">
                Vielen Dank fuer Ihre Rueckmeldung, {details?.freelancer_name}!
              </p>
            </div>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="booking-page">
      <div className="booking-card">
        <div className="booking-card__header">
          <h1 className="booking-card__title">RentFlow</h1>
          <p className="booking-card__subtitle">Buchungsanfrage</p>
        </div>

        <div className="booking-card__body">
          {error && <div className="booking-card__error">{error}</div>}

          <div className="booking-card__greeting">
            Hallo {details?.freelancer_name},
          </div>

          <p className="booking-card__intro">
            Sie haben eine Buchungsanfrage fuer folgendes Projekt erhalten:
          </p>

          <div className="booking-card__details">
            <div className="booking-card__detail-row">
              <span className="booking-card__detail-label">Projekt</span>
              <span className="booking-card__detail-value">{details?.project_name}</span>
            </div>
            <div className="booking-card__detail-row">
              <span className="booking-card__detail-label">Zeitraum</span>
              <span className="booking-card__detail-value">{details?.project_dates}</span>
            </div>
            <div className="booking-card__detail-row">
              <span className="booking-card__detail-label">Rolle</span>
              <span className="booking-card__detail-value">{details?.role}</span>
            </div>
            {details?.location && (
              <div className="booking-card__detail-row">
                <span className="booking-card__detail-label">Ort</span>
                <span className="booking-card__detail-value">{details.location}</span>
              </div>
            )}
          </div>

          {details?.message && (
            <div className="booking-card__message">
              <span className="booking-card__message-label">Nachricht:</span>
              <p>{details.message}</p>
            </div>
          )}

          {showAlternative ? (
            <div className="booking-card__alternative">
              <label className="booking-card__alt-label" htmlFor="alt-message">
                Ihr Alternativvorschlag:
              </label>
              <textarea
                id="alt-message"
                className="booking-card__alt-textarea"
                value={alternativeMessage}
                onChange={(e) => setAlternativeMessage(e.target.value)}
                placeholder="z.B. Ich kann nur vom 15.-18. oder schlage Max Mustermann als Ersatz vor..."
                rows={4}
              />
              <div className="booking-card__alt-actions">
                <button
                  className="booking-btn booking-btn--cancel"
                  onClick={() => setShowAlternative(false)}
                  disabled={submitting}
                >
                  Abbrechen
                </button>
                <button
                  className="booking-btn booking-btn--alternative"
                  onClick={() => handleRespond('alternative', alternativeMessage)}
                  disabled={submitting || !alternativeMessage.trim()}
                >
                  {submitting ? 'Wird gesendet...' : 'Vorschlag senden'}
                </button>
              </div>
            </div>
          ) : (
            <div className="booking-card__actions">
              <button
                className="booking-btn booking-btn--accept"
                onClick={() => handleRespond('accepted')}
                disabled={submitting}
              >
                {submitting ? 'Wird gesendet...' : 'Zusagen'}
              </button>
              <button
                className="booking-btn booking-btn--decline"
                onClick={() => handleRespond('declined')}
                disabled={submitting}
              >
                {submitting ? 'Wird gesendet...' : 'Absagen'}
              </button>
              <button
                className="booking-btn booking-btn--alt-toggle"
                onClick={() => setShowAlternative(true)}
                disabled={submitting}
              >
                Alternative vorschlagen
              </button>
            </div>
          )}
        </div>
      </div>

      <div className="booking-version">
        RentFlow
      </div>
    </div>
  )
}

export default BookingResponse
