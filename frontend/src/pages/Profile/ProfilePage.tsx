import { useState, useMemo } from 'react'
import { useAuthStore } from '../../stores/authStore'
import { useThemeStore } from '../../stores/themeStore'
import { useNotificationStore } from '../../stores/notificationStore'
import { authApi } from '../../services/api'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  User,
  Mail,
  Phone,
  Shield,
  Camera,
  Lock,
  Eye,
  EyeOff,
  Bell,
  Monitor,
  Sun,
  Moon,
  Globe,
  Calendar,
  Laptop,
  LogOut,
  Save,
  AlertTriangle,
} from 'lucide-react'
import './Profile.scss'

// Password strength calculation
function getPasswordStrength(password: string): { score: number; label: string; color: string } {
  let score = 0
  if (password.length >= 8) score++
  if (password.length >= 12) score++
  if (/[a-z]/.test(password) && /[A-Z]/.test(password)) score++
  if (/\d/.test(password)) score++
  if (/[^a-zA-Z0-9]/.test(password)) score++

  if (score <= 1) return { score, label: 'Sehr schwach', color: 'danger' }
  if (score === 2) return { score, label: 'Schwach', color: 'warning' }
  if (score === 3) return { score, label: 'Mittel', color: 'yellow' }
  if (score === 4) return { score, label: 'Stark', color: 'success' }
  return { score, label: 'Sehr stark', color: 'cyan' }
}

function ProfilePage() {
  const user = useAuthStore((state) => state.user)
  const setUser = useAuthStore((state) => state.setUser)
  const isDarkMode = useThemeStore((state) => state.isDarkMode)
  const toggleDarkMode = useThemeStore((state) => state.toggleDarkMode)
  const addNotification = useNotificationStore((state) => state.addNotification)
  const queryClient = useQueryClient()

  // Profile form state
  const nameParts = (user?.name || '').split(' ')
  const [firstName, setFirstName] = useState(nameParts[0] || '')
  const [lastName, setLastName] = useState(nameParts.slice(1).join(' ') || '')
  const [email, setEmail] = useState(user?.email || '')
  const [phone, setPhone] = useState('')

  // Password form state
  const [currentPassword, setCurrentPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [showCurrentPassword, setShowCurrentPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  // Notification preferences state
  const [emailNotifications, setEmailNotifications] = useState(true)
  const [pushNotifications, setPushNotifications] = useState(false)
  const [dailySummary, setDailySummary] = useState(true)

  // Language state
  const [language, setLanguage] = useState('de')
  const [dateFormat, setDateFormat] = useState('DD.MM.YYYY')

  // Computed values
  const userInitials = useMemo(() => {
    if (user?.name) {
      return user.name.split(' ').map(n => n[0]).join('').toUpperCase().slice(0, 2)
    }
    return user?.email?.[0]?.toUpperCase() || '?'
  }, [user])

  const passwordStrength = useMemo(() => getPasswordStrength(newPassword), [newPassword])
  const passwordsMatch = confirmPassword === '' || newPassword === confirmPassword
  const canChangePassword = currentPassword.length > 0 && newPassword.length >= 8 && newPassword === confirmPassword

  // Sessions query
  const { data: sessions = [] } = useQuery({
    queryKey: ['sessions'],
    queryFn: () => authApi.getSessions(),
  })

  // Update profile mutation
  const updateProfileMutation = useMutation({
    mutationFn: () => authApi.updateProfile({
      first_name: firstName,
      last_name: lastName,
      email,
      phone,
    }),
    onSuccess: () => {
      const fullName = `${firstName} ${lastName}`.trim()
      setUser({ ...user!, name: fullName, email })
      addNotification('Profil erfolgreich aktualisiert.', 'success', { title: 'Gespeichert' })
    },
    onError: () => {
      addNotification('Fehler beim Speichern des Profils.', 'error', { title: 'Fehler' })
    },
  })

  // Change password mutation
  const changePasswordMutation = useMutation({
    mutationFn: () => authApi.changePassword({
      current_password: currentPassword,
      new_password: newPassword,
    }),
    onSuccess: () => {
      setCurrentPassword('')
      setNewPassword('')
      setConfirmPassword('')
      addNotification('Passwort erfolgreich geändert.', 'success', { title: 'Passwort geändert' })
    },
    onError: () => {
      addNotification('Fehler beim Ändern des Passworts. Bitte überprüfe dein aktuelles Passwort.', 'error', { title: 'Fehler' })
    },
  })

  // Revoke sessions mutation
  const revokeSessionsMutation = useMutation({
    mutationFn: () => authApi.revokeOtherSessions(),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['sessions'] })
      addNotification('Alle anderen Sitzungen wurden beendet.', 'success', { title: 'Sitzungen beendet' })
    },
    onError: () => {
      addNotification('Fehler beim Beenden der Sitzungen.', 'error', { title: 'Fehler' })
    },
  })

  const handleSaveProfile = (e: React.FormEvent) => {
    e.preventDefault()
    updateProfileMutation.mutate()
  }

  const handleChangePassword = (e: React.FormEvent) => {
    e.preventDefault()
    if (!canChangePassword) return
    changePasswordMutation.mutate()
  }

  const formatSessionDate = (dateStr: string) => {
    const date = new Date(dateStr)
    return date.toLocaleDateString('de-DE', {
      day: '2-digit',
      month: '2-digit',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  }

  return (
    <div className="profile">
      {/* Profile Header */}
      <div className="profile__header">
        <div className="profile__avatarSection">
          <div className="profile__avatarRing">
            <div className="profile__avatar">{userInitials}</div>
          </div>
          <div className="profile__headerInfo">
            <h1 className="profile__name">{user?.name || 'Benutzer'}</h1>
            <p className="profile__email">{user?.email}</p>
            <div className="profile__badges">
              <span className="profile__roleBadge">
                <Shield size={12} />
                Administrator
              </span>
              <span className="profile__dateBadge">
                <Calendar size={12} />
                Mitglied seit März 2025
              </span>
            </div>
          </div>
        </div>
      </div>

      <div className="profile__grid">
        {/* Persönliche Daten */}
        <div className="profile__card">
          <div className="profile__cardHeader">
            <User size={20} />
            <h2>Persönliche Daten</h2>
          </div>
          <form onSubmit={handleSaveProfile} className="profile__form">
            <div className="profile__formRow">
              <div className="profile__formGroup">
                <label className="profile__label">Vorname</label>
                <input
                  type="text"
                  className="profile__input"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  placeholder="Vorname"
                />
              </div>
              <div className="profile__formGroup">
                <label className="profile__label">Nachname</label>
                <input
                  type="text"
                  className="profile__input"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  placeholder="Nachname"
                />
              </div>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">
                <Mail size={14} />
                E-Mail
              </label>
              <input
                type="email"
                className="profile__input"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="E-Mail Adresse"
              />
              <span className="profile__hint">Bei Änderung wird eine Bestätigungs-E-Mail gesendet.</span>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">
                <Phone size={14} />
                Telefon
              </label>
              <input
                type="tel"
                className="profile__input"
                value={phone}
                onChange={(e) => setPhone(e.target.value)}
                placeholder="+49 123 456789"
              />
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">
                <Shield size={14} />
                Position / Rolle
              </label>
              <input
                type="text"
                className="profile__input profile__input--readonly"
                value="Administrator"
                readOnly
                disabled
              />
              <span className="profile__hint">Wird vom Admin festgelegt.</span>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">
                <Camera size={14} />
                Profilbild
              </label>
              <div className="profile__uploadArea">
                <Camera size={24} />
                <span>Bild hochladen oder hierher ziehen</span>
                <span className="profile__uploadHint">JPG, PNG, max. 5 MB</span>
              </div>
            </div>

            <button
              type="submit"
              className="profile__btn"
              disabled={updateProfileMutation.isPending}
            >
              <Save size={16} />
              {updateProfileMutation.isPending ? 'Speichern...' : 'Speichern'}
            </button>
          </form>
        </div>

        {/* Passwort ändern */}
        <div className="profile__card">
          <div className="profile__cardHeader">
            <Lock size={20} />
            <h2>Passwort ändern</h2>
          </div>
          <form onSubmit={handleChangePassword} className="profile__form">
            <div className="profile__formGroup">
              <label className="profile__label">Aktuelles Passwort</label>
              <div className="profile__inputWrapper">
                <input
                  type={showCurrentPassword ? 'text' : 'password'}
                  className="profile__input"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                  placeholder="Aktuelles Passwort eingeben"
                />
                <button
                  type="button"
                  className="profile__inputToggle"
                  onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                >
                  {showCurrentPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">Neues Passwort</label>
              <div className="profile__inputWrapper">
                <input
                  type={showNewPassword ? 'text' : 'password'}
                  className="profile__input"
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="Neues Passwort eingeben"
                />
                <button
                  type="button"
                  className="profile__inputToggle"
                  onClick={() => setShowNewPassword(!showNewPassword)}
                >
                  {showNewPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
              {newPassword.length > 0 && (
                <div className="profile__strengthBar">
                  <div className="profile__strengthTrack">
                    <div
                      className={`profile__strengthFill ${`profile__strengthFill--${passwordStrength.color}`}`}
                      style={{ width: `${(passwordStrength.score / 5) * 100}%` }}
                    />
                  </div>
                  <span className={`profile__strengthLabel ${`profile__strengthLabel--${passwordStrength.color}`}`}>
                    {passwordStrength.label}
                  </span>
                </div>
              )}
              <span className="profile__hint">Mindestens 8 Zeichen, Gross-/Kleinbuchstaben, Zahlen und Sonderzeichen empfohlen.</span>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">Neues Passwort bestätigen</label>
              <div className="profile__inputWrapper">
                <input
                  type={showConfirmPassword ? 'text' : 'password'}
                  className={`profile__input ${!passwordsMatch ? 'profile__input--error' : ''}`}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="Neues Passwort wiederholen"
                />
                <button
                  type="button"
                  className="profile__inputToggle"
                  onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                >
                  {showConfirmPassword ? <EyeOff size={16} /> : <Eye size={16} />}
                </button>
              </div>
              {!passwordsMatch && (
                <span className="profile__error">
                  <AlertTriangle size={12} />
                  Passwörter stimmen nicht überein.
                </span>
              )}
            </div>

            <button
              type="submit"
              className="profile__btn"
              disabled={!canChangePassword || changePasswordMutation.isPending}
            >
              <Lock size={16} />
              {changePasswordMutation.isPending ? 'Wird geändert...' : 'Passwort ändern'}
            </button>
          </form>
        </div>

        {/* Benachrichtigungs-Präferenzen */}
        <div className="profile__card">
          <div className="profile__cardHeader">
            <Bell size={20} />
            <h2>Benachrichtigungen</h2>
          </div>
          <div className="profile__toggleList">
            <div className="profile__toggleItem">
              <div className="profile__toggleInfo">
                <span className="profile__toggleLabel">E-Mail Benachrichtigungen</span>
                <span className="profile__toggleDesc">Erhalte wichtige Updates per E-Mail</span>
              </div>
              <button
                className={`profile__toggle ${emailNotifications ? 'profile__toggle--active' : ''}`}
                onClick={() => setEmailNotifications(!emailNotifications)}
              >
                <span className="profile__toggleKnob" />
              </button>
            </div>

            <div className="profile__toggleItem">
              <div className="profile__toggleInfo">
                <span className="profile__toggleLabel">Push Benachrichtigungen</span>
                <span className="profile__toggleDesc">Browser-Benachrichtigungen für Echtzeit-Updates</span>
              </div>
              <button
                className={`profile__toggle ${pushNotifications ? 'profile__toggle--active' : ''}`}
                onClick={() => setPushNotifications(!pushNotifications)}
              >
                <span className="profile__toggleKnob" />
              </button>
            </div>

            <div className="profile__toggleItem">
              <div className="profile__toggleInfo">
                <span className="profile__toggleLabel">Tägliche Zusammenfassung</span>
                <span className="profile__toggleDesc">Täglicher Überblick über alle Aktivitäten</span>
              </div>
              <button
                className={`profile__toggle ${dailySummary ? 'profile__toggle--active' : ''}`}
                onClick={() => setDailySummary(!dailySummary)}
              >
                <span className="profile__toggleKnob" />
              </button>
            </div>
          </div>
          <a href="/settings/email" className="profile__link">
            Alle Benachrichtigungs-Einstellungen &rarr;
          </a>
        </div>

        {/* Aktive Sitzungen */}
        <div className="profile__card">
          <div className="profile__cardHeader">
            <Monitor size={20} />
            <h2>Aktive Sitzungen</h2>
          </div>
          <div className="profile__sessionList">
            {(sessions as Array<{ id: string; browser: string; ip: string; last_active: string; current: boolean }>).map((session) => (
              <div
                key={session.id}
                className={`profile__sessionItem ${session.current ? 'profile__sessionItem--current' : ''}`}
              >
                <div className="profile__sessionIcon">
                  <Laptop size={18} />
                </div>
                <div className="profile__sessionInfo">
                  <span className="profile__sessionBrowser">
                    {session.browser}
                    {session.current && <span className="profile__sessionCurrent">Aktuelle Sitzung</span>}
                  </span>
                  <span className="profile__sessionMeta">
                    IP: {session.ip} &middot; Zuletzt aktiv: {formatSessionDate(session.last_active)}
                  </span>
                </div>
              </div>
            ))}
          </div>
          <button
            className="profile__btn profile__btn--danger"
            onClick={() => revokeSessionsMutation.mutate()}
            disabled={revokeSessionsMutation.isPending}
          >
            <LogOut size={16} />
            {revokeSessionsMutation.isPending ? 'Wird beendet...' : 'Alle anderen Sitzungen beenden'}
          </button>
        </div>

        {/* Darstellung */}
        <div className="profile__card">
          <div className="profile__cardHeader">
            <Globe size={20} />
            <h2>Darstellung</h2>
          </div>
          <div className="profile__form">
            <div className="profile__formGroup">
              <label className="profile__label">Design</label>
              <div className="profile__themeSwitch">
                <button
                  className={`profile__themeOption ${isDarkMode ? 'profile__themeOption--active' : ''}`}
                  onClick={() => { if (!isDarkMode) toggleDarkMode() }}
                >
                  <Moon size={16} />
                  Dunkel
                </button>
                <button
                  className={`profile__themeOption ${!isDarkMode ? 'profile__themeOption--active' : ''}`}
                  onClick={() => { if (isDarkMode) toggleDarkMode() }}
                >
                  <Sun size={16} />
                  Hell
                </button>
              </div>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">Sprache</label>
              <select
                className="profile__select"
                value={language}
                onChange={(e) => setLanguage(e.target.value)}
              >
                <option value="de">Deutsch</option>
                <option value="en">English</option>
              </select>
            </div>

            <div className="profile__formGroup">
              <label className="profile__label">Datumsformat</label>
              <select
                className="profile__select"
                value={dateFormat}
                onChange={(e) => setDateFormat(e.target.value)}
              >
                <option value="DD.MM.YYYY">DD.MM.YYYY (23.03.2026)</option>
                <option value="YYYY-MM-DD">YYYY-MM-DD (2026-03-23)</option>
                <option value="MM/DD/YYYY">MM/DD/YYYY (03/23/2026)</option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

export default ProfilePage
