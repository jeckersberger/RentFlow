import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation } from '@tanstack/react-query'
import { mailApi, contactApi, projectApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import './Mail.scss'

interface ComposeForm {
  from_mailbox_id: string
  to: string
  subject: string
  project_id: string
  body: string
}

function MailComposePage() {
  const navigate = useNavigate()
  const addNotification = useNotificationStore(s => s.addNotification)
  const [form, setForm] = useState<ComposeForm>({
    from_mailbox_id: '',
    to: '',
    subject: '',
    project_id: '',
    body: '',
  })
  const [contactSuggestions, setContactSuggestions] = useState<any[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)

  // Fetch mailboxes
  const { data: mailboxes = [] } = useQuery({
    queryKey: ['mailboxes'],
    queryFn: () => mailApi.listMailboxes(),
  })

  // Fetch projects
  const { data: projectsData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectApi.list(1, 100),
  })
  const projects = projectsData?.items || projectsData?.data || []

  // Fetch contacts for autocomplete
  const { data: contactsData } = useQuery({
    queryKey: ['contacts'],
    queryFn: () => contactApi.list({ limit: 200 }),
  })
  const contacts = contactsData?.data || contactsData || []

  // Send mutation
  const sendMutation = useMutation({
    mutationFn: (data: any) => mailApi.send(data),
    onSuccess: () => {
      addNotification('E-Mail wurde erfolgreich gesendet', 'success', { title: 'Gesendet', duration: 4000 })
      navigate('/mail/sent')
    },
    onError: () => {
      addNotification('E-Mail konnte nicht gesendet werden', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const handleToChange = (value: string) => {
    setForm(prev => ({ ...prev, to: value }))
    if (value.length >= 2 && Array.isArray(contacts)) {
      const filtered = contacts.filter((c: any) => {
        const email = c.email || ''
        const name = c.name || c.company_name || ''
        return email.toLowerCase().includes(value.toLowerCase()) ||
               name.toLowerCase().includes(value.toLowerCase())
      }).slice(0, 5)
      setContactSuggestions(filtered)
      setShowSuggestions(filtered.length > 0)
    } else {
      setShowSuggestions(false)
    }
  }

  const selectContact = (contact: any) => {
    setForm(prev => ({ ...prev, to: contact.email || '' }))
    setShowSuggestions(false)
  }

  const handleSend = () => {
    if (!form.to || !form.subject) {
      addNotification('Bitte Empfänger und Betreff ausfüllen', 'warning', { title: 'Fehlende Felder', duration: 3000 })
      return
    }
    const mailbox = Array.isArray(mailboxes) ? mailboxes.find((m: any) => m.id === form.from_mailbox_id) : null
    sendMutation.mutate({
      from_mailbox_id: form.from_mailbox_id || undefined,
      from_email: mailbox?.email || undefined,
      to: form.to,
      subject: form.subject,
      body_text: form.body,
      project_id: form.project_id || undefined,
    })
  }

  const handleSaveDraft = () => {
    addNotification('Entwurf gespeichert (Platzhalter)', 'info', { title: 'Entwurf', duration: 3000 })
  }

  // Auto-select first mailbox
  if (Array.isArray(mailboxes) && mailboxes.length > 0 && !form.from_mailbox_id) {
    setForm(prev => ({ ...prev, from_mailbox_id: mailboxes[0].id }))
  }

  return (
    <div className="page">
      <div className="header">
        <div>
          <h1 className="title">Neue E-Mail</h1>
          <p className="subtitle">Verfassen und senden Sie eine neue Nachricht</p>
        </div>
      </div>

      <div
        style={{
          background: 'var(--glass-bg)',
          backdropFilter: 'blur(16px)',
          WebkitBackdropFilter: 'blur(16px)',
          border: '1px solid var(--color-border)',
          borderRadius: 'var(--radius-card)',
          padding: 'var(--spacing-6)',
          display: 'flex',
          flexDirection: 'column',
          gap: 'var(--spacing-5)',
          animation: 'fadeInUp 0.4s ease-out both',
        }}
      >
        {/* Von / From */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          <label style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600, color: 'var(--color-text-secondary)' }}>
            Von
          </label>
          <select
            value={form.from_mailbox_id}
            onChange={e => setForm(prev => ({ ...prev, from_mailbox_id: e.target.value }))}
            style={{
              background: 'var(--glass-bg-input)',
              border: '1px solid var(--color-border-strong)',
              borderRadius: '8px',
              padding: '8px 12px',
              fontSize: '14px',
              color: 'var(--color-text-primary)',
              cursor: 'pointer',
            }}
          >
            <option value="">Mailbox wählen...</option>
            {Array.isArray(mailboxes) && mailboxes.map((mb: any) => (
              <option key={mb.id} value={mb.id}>
                {mb.label || mb.type} - {mb.email}
              </option>
            ))}
          </select>
        </div>

        {/* An / To */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)', position: 'relative' }}>
          <label style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600, color: 'var(--color-text-secondary)' }}>
            An
          </label>
          <input
            type="email"
            value={form.to}
            onChange={e => handleToChange(e.target.value)}
            onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
            onFocus={() => { if (contactSuggestions.length > 0) setShowSuggestions(true) }}
            placeholder="empfaenger@beispiel.de"
            className="searchInput"
            style={{ width: '100%' }}
          />
          {showSuggestions && (
            <div
              style={{
                position: 'absolute',
                top: '100%',
                left: 0,
                right: 0,
                zIndex: 10,
                background: 'var(--glass-bg)',
                backdropFilter: 'blur(16px)',
                border: '1px solid var(--color-border-strong)',
                borderRadius: '8px',
                marginTop: '4px',
                overflow: 'hidden',
                boxShadow: '0 8px 32px rgba(0, 0, 0, 0.3)',
              }}
            >
              {contactSuggestions.map((c: any, idx: number) => (
                <div
                  key={c.id || idx}
                  onClick={() => selectContact(c)}
                  style={{
                    padding: '8px 12px',
                    cursor: 'pointer',
                    fontSize: '13px',
                    borderBottom: idx < contactSuggestions.length - 1 ? '1px solid var(--color-border-subtle)' : 'none',
                    color: 'var(--color-text-primary)',
                    transition: 'background 0.15s',
                  }}
                  onMouseEnter={e => (e.currentTarget.style.background = 'rgba(0, 212, 255, 0.06)')}
                  onMouseLeave={e => (e.currentTarget.style.background = 'transparent')}
                >
                  <div style={{ fontWeight: 500 }}>{c.name || c.company_name || 'Unbekannt'}</div>
                  <div style={{ fontSize: '11px', color: 'var(--color-text-muted)' }}>{c.email}</div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Betreff / Subject */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          <label style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600, color: 'var(--color-text-secondary)' }}>
            Betreff
          </label>
          <input
            type="text"
            value={form.subject}
            onChange={e => setForm(prev => ({ ...prev, subject: e.target.value }))}
            placeholder="Betreff eingeben..."
            className="searchInput"
            style={{ width: '100%' }}
          />
        </div>

        {/* Projekt / Project */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          <label style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600, color: 'var(--color-text-secondary)' }}>
            Projekt (optional)
          </label>
          <select
            value={form.project_id}
            onChange={e => setForm(prev => ({ ...prev, project_id: e.target.value }))}
            style={{
              background: 'var(--glass-bg-input)',
              border: '1px solid var(--color-border-strong)',
              borderRadius: '8px',
              padding: '8px 12px',
              fontSize: '14px',
              color: 'var(--color-text-primary)',
              cursor: 'pointer',
            }}
          >
            <option value="">Kein Projekt</option>
            {projects.map((p: any) => (
              <option key={p.id} value={p.id}>{p.name}</option>
            ))}
          </select>
        </div>

        {/* Body / Text */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          <label style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600, color: 'var(--color-text-secondary)' }}>
            Nachricht
          </label>
          <textarea
            value={form.body}
            onChange={e => setForm(prev => ({ ...prev, body: e.target.value }))}
            placeholder="Ihre Nachricht hier eingeben..."
            rows={12}
            style={{
              background: 'var(--glass-bg-input)',
              border: '1px solid var(--color-border-strong)',
              borderRadius: '8px',
              padding: '12px',
              fontSize: '14px',
              color: 'var(--color-text-primary)',
              resize: 'vertical',
              lineHeight: '1.6',
              fontFamily: 'inherit',
            }}
          />
        </div>

        {/* Anhänge / Attachments placeholder */}
        <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
          <label style={{ fontSize: 'var(--font-size-sm)', fontWeight: 600, color: 'var(--color-text-secondary)' }}>
            Anhänge
          </label>
          <div
            style={{
              background: 'var(--glass-bg-input)',
              border: '2px dashed var(--color-border-strong)',
              borderRadius: '8px',
              padding: '24px',
              textAlign: 'center',
              color: 'var(--color-text-muted)',
              fontSize: '13px',
              cursor: 'pointer',
              transition: 'border-color 0.2s',
            }}
          >
            Dateien hierher ziehen oder klicken zum Hochladen (demnächst verfügbar)
          </div>
        </div>

        {/* Actions */}
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', justifyContent: 'flex-end', paddingTop: 'var(--spacing-3)', borderTop: '1px solid var(--color-border)' }}>
          <button
            className="btnSecondary"
            onClick={() => navigate(-1)}
          >
            Abbrechen
          </button>
          <button
            className="btnSecondary"
            onClick={handleSaveDraft}
          >
            Als Entwurf speichern
          </button>
          <button
            className="btnPrimary"
            onClick={handleSend}
            disabled={sendMutation.isPending}
            style={sendMutation.isPending ? { opacity: 0.7, cursor: 'not-allowed' } : undefined}
          >
            {sendMutation.isPending ? 'Wird gesendet...' : 'Senden'}
          </button>
        </div>
      </div>
    </div>
  )
}

export default MailComposePage
