import { useState, useEffect, useRef, useCallback } from 'react'
import QRCode from 'qrcode'
import { Modal } from '../Modal/Modal'
import { authApi } from '../../services/api'
import './QRLoginModal.scss'

interface QRLoginModalProps {
  isOpen: boolean
  onClose: () => void
}

export function QRLoginModal({ isOpen, onClose }: QRLoginModalProps) {
  const [, setQrToken] = useState<string | null>(null)
  const [qrDataUrl, setQrDataUrl] = useState<string | null>(null)
  const [, setExpiresAt] = useState<Date | null>(null)
  const [secondsLeft, setSecondsLeft] = useState(0)
  const [status, setStatus] = useState<'loading' | 'ready' | 'redeemed' | 'expired' | 'error'>('loading')
  const [errorMsg, setErrorMsg] = useState('')
  const pollingRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const cleanup = useCallback(() => {
    if (pollingRef.current) {
      clearInterval(pollingRef.current)
      pollingRef.current = null
    }
    if (countdownRef.current) {
      clearInterval(countdownRef.current)
      countdownRef.current = null
    }
  }, [])

  const generateToken = useCallback(async () => {
    setStatus('loading')
    setErrorMsg('')
    setQrToken(null)
    setQrDataUrl(null)
    cleanup()

    try {
      const result = await authApi.generateQRToken()
      const token = result.qr_token
      const expires = new Date(result.expires_at)

      setQrToken(token)
      setExpiresAt(expires)
      setStatus('ready')

      // Generate QR code as data URL
      const dataUrl = await QRCode.toDataURL(token, {
        width: 280,
        margin: 2,
        color: {
          dark: '#000000',
          light: '#ffffff',
        },
        errorCorrectionLevel: 'M',
      })
      setQrDataUrl(dataUrl)

      // Start countdown
      const updateCountdown = () => {
        const now = new Date()
        const diff = Math.max(0, Math.floor((expires.getTime() - now.getTime()) / 1000))
        setSecondsLeft(diff)
        if (diff <= 0) {
          setStatus('expired')
          cleanup()
        }
      }
      updateCountdown()
      countdownRef.current = setInterval(updateCountdown, 1000)

      // Start polling for redemption
      pollingRef.current = setInterval(async () => {
        try {
          const statusResult = await authApi.getQRStatus(token)
          if (statusResult.redeemed) {
            setStatus('redeemed')
            cleanup()
          } else if (statusResult.status === 'expired') {
            setStatus('expired')
            cleanup()
          }
        } catch {
          // Ignore polling errors silently
        }
      }, 3000)
    } catch (err) {
      setStatus('error')
      setErrorMsg('QR-Token konnte nicht generiert werden.')
      console.error('QR token generation failed:', err)
    }
  }, [cleanup])

  useEffect(() => {
    if (isOpen) {
      generateToken()
    } else {
      cleanup()
      setStatus('loading')
      setQrToken(null)
      setQrDataUrl(null)
    }
    return cleanup
  }, [isOpen, generateToken, cleanup])

  const formatTime = (seconds: number): string => {
    const m = Math.floor(seconds / 60)
    const s = seconds % 60
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="QR-Code fuer Scanner" size="sm">
      <div className="qr-login-modal">
        {status === 'loading' && (
          <div className="qr-login-modal__loading">
            <div className="qr-login-modal__spinner" />
            <p>QR-Code wird generiert...</p>
          </div>
        )}

        {status === 'ready' && qrDataUrl && (
          <>
            <p className="qr-login-modal__description">
              Scanne diesen QR-Code mit der RentFlow Scanner-App, um dich einzuloggen.
            </p>
            <div className="qr-login-modal__qr-container">
              <img src={qrDataUrl} alt="QR-Code fuer Scanner Login" className="qr-login-modal__qr-image" />
            </div>
            <div className={`qr-login-modal__countdown ${secondsLeft <= 60 ? 'qr-login-modal__countdown--warning' : ''}`}>
              Gueltig noch {formatTime(secondsLeft)}
            </div>
            <p className="qr-login-modal__hint">
              Der Code ist 5 Minuten gueltig und kann nur einmal verwendet werden.
            </p>
          </>
        )}

        {status === 'redeemed' && (
          <div className="qr-login-modal__status qr-login-modal__status--success">
            <div className="qr-login-modal__status-icon">&#10003;</div>
            <p>Scanner erfolgreich eingeloggt!</p>
            <p className="qr-login-modal__hint">Du kannst dieses Fenster jetzt schliessen.</p>
          </div>
        )}

        {status === 'expired' && (
          <div className="qr-login-modal__status qr-login-modal__status--expired">
            <p>Der QR-Code ist abgelaufen.</p>
            <button className="qr-login-modal__retry-btn" onClick={generateToken}>
              Neuen QR-Code generieren
            </button>
          </div>
        )}

        {status === 'error' && (
          <div className="qr-login-modal__status qr-login-modal__status--error">
            <p>{errorMsg}</p>
            <button className="qr-login-modal__retry-btn" onClick={generateToken}>
              Erneut versuchen
            </button>
          </div>
        )}
      </div>
    </Modal>
  )
}
