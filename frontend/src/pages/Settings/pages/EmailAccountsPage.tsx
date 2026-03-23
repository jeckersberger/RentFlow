import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { mailApi } from '../../../services/api'
import { useNotificationStore } from '../../../stores/notificationStore'
import { Plus, Pencil, Trash2, TestTube, X } from 'lucide-react'
import '../Settings.module.scss'

interface MailboxForm {
  type: 'general' | 'invoices' | 'personal'
  email: string
  label: string
  imap_host: string
  imap_port: number
  smtp_host: string
  smtp_port: number
  username: string
  password: string
  tls: boolean
}

const EMPTY_FORM: MailboxForm = {
  type: 'general',
  email: '',
  label: '',
  imap_host: '',
  imap_port: 993,
  smtp_host: '',
  smtp_port: 587,
  username: '',
  password: '',
  tls: true,
}

const TYPE_LABELS: Record<string, string> = {
  general: 'Allgemein',
  invoices: 'Rechnungen',
  personal: 'Persönlich',
}

function EmailAccountsPage() {
  const queryClient = useQueryClient()
  const addNotification = useNotificationStore(s => s.addNotification)
  const [modalOpen, setModalOpen] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<MailboxForm>(EMPTY_FORM)
  const [testingId, setTestingId] = useState<string | null>(null)

  const { data: mailboxes = [], isLoading } = useQuery({
    queryKey: ['mailboxes'],
    queryFn: () => mailApi.listMailboxes(),
  })

  const createMutation = useMutation({
    mutationFn: (data: MailboxForm) => mailApi.createMailbox(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mailboxes'] })
      addNotification('Mailbox erfolgreich erstellt', 'success', { duration: 3000 })
      closeModal()
    },
    onError: () => {
      addNotification('Fehler beim Erstellen der Mailbox', 'error', { duration: 4000 })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: MailboxForm }) => mailApi.updateMailbox(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mailboxes'] })
      addNotification('Mailbox aktualisiert', 'success', { duration: 3000 })
      closeModal()
    },
    onError: () => {
      addNotification('Fehler beim Aktualisieren', 'error', { duration: 4000 })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => mailApi.deleteMailbox(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mailboxes'] })
      addNotification('Mailbox gelöscht', 'success', { duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Löschen', 'error', { duration: 4000 })
    },
  })

  const testMutation = useMutation({
    mutationFn: (id: string) => mailApi.testMailbox(id),
    onSuccess: () => {
      addNotification('Verbindung erfolgreich!', 'success', { title: 'Verbindungstest', duration: 4000 })
      setTestingId(null)
    },
    onError: () => {
      addNotification('Verbindung fehlgeschlagen', 'error', { title: 'Verbindungstest', duration: 4000 })
      setTestingId(null)
    },
  })

  const openCreate = () => {
    setEditingId(null)
    setForm(EMPTY_FORM)
    setModalOpen(true)
  }

  const openEdit = (mailbox: any) => {
    setEditingId(mailbox.id)
    setForm({
      type: mailbox.type || 'general',
      email: mailbox.email || '',
      label: mailbox.label || '',
      imap_host: mailbox.imap_host || '',
      imap_port: mailbox.imap_port || 993,
      smtp_host: mailbox.smtp_host || '',
      smtp_port: mailbox.smtp_port || 587,
      username: mailbox.username || '',
      password: '',
      tls: mailbox.tls ?? true,
    })
    setModalOpen(true)
  }

  const closeModal = () => {
    setModalOpen(false)
    setEditingId(null)
    setForm(EMPTY_FORM)
  }

  const handleSave = () => {
    if (!form.email || !form.imap_host || !form.smtp_host) {
      addNotification('Bitte alle Pflichtfelder ausfüllen', 'warning', { duration: 3000 })
      return
    }
    if (editingId) {
      updateMutation.mutate({ id: editingId, data: form })
    } else {
      createMutation.mutate(form)
    }
  }

  const handleDelete = (id: string, label: string) => {
    if (confirm(`Mailbox "${label}" wirklich löschen?`)) {
      deleteMutation.mutate(id)
    }
  }

  const handleTest = (id: string) => {
    setTestingId(id)
    testMutation.mutate(id)
  }

  const inputStyle: React.CSSProperties = {
    background: 'var(--glass-bg-input)',
    border: '1px solid var(--color-border-strong)',
    borderRadius: '8px',
    padding: '8px 12px',
    fontSize: '14px',
    color: 'var(--color-text-primary)',
    width: '100%',
  }

  const labelStyle: React.CSSProperties = {
    fontSize: '13px',
    fontWeight: 600,
    color: 'var(--color-text-secondary)',
    marginBottom: '4px',
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <div>
          <h2 style={{ fontSize: '20px', fontWeight: 700, color: 'var(--color-text-primary)', margin: 0 }}>
            E-Mail-Konten
          </h2>
          <p style={{ fontSize: '13px', color: 'var(--color-text-secondary)', margin: '4px 0 0' }}>
            Konfigurieren Sie Ihre Mailboxen für den E-Mail-Empfang und -Versand
          </p>
        </div>
        <button
          onClick={openCreate}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: '6px',
            background: 'linear-gradient(135deg, #00d4ff 0%, #8b5cf6 100%)',
            color: '#0a0f1a',
            border: 'none',
            borderRadius: '8px',
            padding: '8px 16px',
            fontSize: '13px',
            fontWeight: 600,
            cursor: 'pointer',
            boxShadow: '0 2px 8px rgba(0, 212, 255, 0.25)',
          }}
        >
          <Plus size={16} />
          Mailbox hinzufügen
        </button>
      </div>

      {/* Mailbox list */}
      {isLoading ? (
        <div style={{ padding: '40px', textAlign: 'center', color: 'var(--color-text-muted)' }}>
          Lade Mailboxen...
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '12px' }}>
          {Array.isArray(mailboxes) && mailboxes.length > 0 ? (
            mailboxes.map((mb: any) => (
              <div
                key={mb.id}
                style={{
                  background: 'var(--glass-bg)',
                  backdropFilter: 'blur(16px)',
                  WebkitBackdropFilter: 'blur(16px)',
                  border: '1px solid var(--color-border)',
                  borderRadius: '12px',
                  padding: '16px 20px',
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  transition: 'border-color 0.2s, box-shadow 0.2s',
                }}
                onMouseEnter={e => {
                  e.currentTarget.style.borderColor = 'rgba(0, 212, 255, 0.2)'
                  e.currentTarget.style.boxShadow = '0 0 20px rgba(0, 212, 255, 0.05)'
                }}
                onMouseLeave={e => {
                  e.currentTarget.style.borderColor = 'var(--color-border)'
                  e.currentTarget.style.boxShadow = 'none'
                }}
              >
                <div>
                  <div style={{ display: 'flex', gap: '8px', alignItems: 'center', marginBottom: '4px' }}>
                    <span style={{ fontWeight: 600, fontSize: '14px', color: 'var(--color-text-primary)' }}>
                      {mb.label || TYPE_LABELS[mb.type] || mb.type}
                    </span>
                    <span
                      style={{
                        padding: '2px 8px',
                        borderRadius: '9999px',
                        fontSize: '10px',
                        fontWeight: 600,
                        background: 'rgba(0, 212, 255, 0.1)',
                        color: 'var(--color-primary)',
                      }}
                    >
                      {TYPE_LABELS[mb.type] || mb.type}
                    </span>
                  </div>
                  <div style={{ fontSize: '13px', color: 'var(--color-text-secondary)' }}>
                    {mb.email}
                  </div>
                  <div style={{ fontSize: '11px', color: 'var(--color-text-muted)', marginTop: '2px' }}>
                    IMAP: {mb.imap_host}:{mb.imap_port} | SMTP: {mb.smtp_host}:{mb.smtp_port}
                    {mb.tls && ' | TLS'}
                  </div>
                </div>
                <div style={{ display: 'flex', gap: '8px' }}>
                  <button
                    onClick={() => handleTest(mb.id)}
                    disabled={testingId === mb.id}
                    style={{
                      background: 'transparent',
                      border: '1px solid var(--color-border)',
                      borderRadius: '6px',
                      padding: '6px 10px',
                      cursor: 'pointer',
                      color: 'var(--color-text-secondary)',
                      display: 'flex',
                      alignItems: 'center',
                      gap: '4px',
                      fontSize: '12px',
                      opacity: testingId === mb.id ? 0.6 : 1,
                    }}
                    title="Verbindung testen"
                  >
                    <TestTube size={14} />
                    {testingId === mb.id ? 'Teste...' : 'Testen'}
                  </button>
                  <button
                    onClick={() => openEdit(mb)}
                    style={{
                      background: 'transparent',
                      border: '1px solid var(--color-border)',
                      borderRadius: '6px',
                      padding: '6px 8px',
                      cursor: 'pointer',
                      color: 'var(--color-text-secondary)',
                    }}
                    title="Bearbeiten"
                  >
                    <Pencil size={14} />
                  </button>
                  <button
                    onClick={() => handleDelete(mb.id, mb.label || mb.email)}
                    style={{
                      background: 'transparent',
                      border: '1px solid rgba(239, 68, 68, 0.3)',
                      borderRadius: '6px',
                      padding: '6px 8px',
                      cursor: 'pointer',
                      color: 'var(--color-danger)',
                    }}
                    title="Löschen"
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            ))
          ) : (
            <div
              style={{
                background: 'var(--glass-bg)',
                border: '1px solid var(--color-border)',
                borderRadius: '12px',
                padding: '48px',
                textAlign: 'center',
                color: 'var(--color-text-muted)',
              }}
            >
              <p style={{ fontSize: '14px', marginBottom: '8px' }}>Keine Mailboxen konfiguriert</p>
              <p style={{ fontSize: '12px' }}>Fügen Sie eine Mailbox hinzu, um E-Mails zu empfangen und zu senden.</p>
            </div>
          )}
        </div>
      )}

      {/* Modal */}
      {modalOpen && (
        <div
          style={{
            position: 'fixed',
            inset: 0,
            zIndex: 1000,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            background: 'rgba(0, 0, 0, 0.6)',
            backdropFilter: 'blur(4px)',
          }}
          onClick={closeModal}
        >
          <div
            style={{
              background: 'var(--glass-bg)',
              backdropFilter: 'blur(24px)',
              WebkitBackdropFilter: 'blur(24px)',
              border: '1px solid var(--color-border)',
              borderRadius: '16px',
              padding: '28px',
              width: '520px',
              maxWidth: '95vw',
              maxHeight: '90vh',
              overflowY: 'auto',
              boxShadow: '0 24px 64px rgba(0, 0, 0, 0.4)',
            }}
            onClick={e => e.stopPropagation()}
          >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '20px' }}>
              <h3 style={{ fontSize: '18px', fontWeight: 700, color: 'var(--color-text-primary)', margin: 0 }}>
                {editingId ? 'Mailbox bearbeiten' : 'Mailbox hinzufügen'}
              </h3>
              <button
                onClick={closeModal}
                style={{ background: 'none', border: 'none', cursor: 'pointer', color: 'var(--color-text-muted)', padding: '4px' }}
              >
                <X size={18} />
              </button>
            </div>

            <div style={{ display: 'flex', flexDirection: 'column', gap: '16px' }}>
              {/* Typ */}
              <div>
                <div style={labelStyle}>Typ</div>
                <select
                  value={form.type}
                  onChange={e => setForm(prev => ({ ...prev, type: e.target.value as any }))}
                  style={{ ...inputStyle, cursor: 'pointer' }}
                >
                  <option value="general">Allgemein</option>
                  <option value="invoices">Rechnungen</option>
                  <option value="personal">Persönlich</option>
                </select>
              </div>

              {/* Label */}
              <div>
                <div style={labelStyle}>Bezeichnung</div>
                <input
                  type="text"
                  value={form.label}
                  onChange={e => setForm(prev => ({ ...prev, label: e.target.value }))}
                  placeholder="z.B. Allgemein"
                  style={inputStyle}
                />
              </div>

              {/* Email */}
              <div>
                <div style={labelStyle}>E-Mail-Adresse</div>
                <input
                  type="email"
                  value={form.email}
                  onChange={e => setForm(prev => ({ ...prev, email: e.target.value }))}
                  placeholder="info@firma.de"
                  style={inputStyle}
                />
              </div>

              {/* IMAP */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 100px', gap: '12px' }}>
                <div>
                  <div style={labelStyle}>IMAP Server</div>
                  <input
                    type="text"
                    value={form.imap_host}
                    onChange={e => setForm(prev => ({ ...prev, imap_host: e.target.value }))}
                    placeholder="imap.provider.de"
                    style={inputStyle}
                  />
                </div>
                <div>
                  <div style={labelStyle}>Port</div>
                  <input
                    type="number"
                    value={form.imap_port}
                    onChange={e => setForm(prev => ({ ...prev, imap_port: parseInt(e.target.value) || 993 }))}
                    style={inputStyle}
                  />
                </div>
              </div>

              {/* SMTP */}
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 100px', gap: '12px' }}>
                <div>
                  <div style={labelStyle}>SMTP Server</div>
                  <input
                    type="text"
                    value={form.smtp_host}
                    onChange={e => setForm(prev => ({ ...prev, smtp_host: e.target.value }))}
                    placeholder="smtp.provider.de"
                    style={inputStyle}
                  />
                </div>
                <div>
                  <div style={labelStyle}>Port</div>
                  <input
                    type="number"
                    value={form.smtp_port}
                    onChange={e => setForm(prev => ({ ...prev, smtp_port: parseInt(e.target.value) || 587 }))}
                    style={inputStyle}
                  />
                </div>
              </div>

              {/* Username */}
              <div>
                <div style={labelStyle}>Benutzername</div>
                <input
                  type="text"
                  value={form.username}
                  onChange={e => setForm(prev => ({ ...prev, username: e.target.value }))}
                  placeholder="benutzername@firma.de"
                  style={inputStyle}
                />
              </div>

              {/* Password */}
              <div>
                <div style={labelStyle}>Passwort</div>
                <input
                  type="password"
                  value={form.password}
                  onChange={e => setForm(prev => ({ ...prev, password: e.target.value }))}
                  placeholder={editingId ? 'Leer lassen um nicht zu ändern' : 'Passwort'}
                  style={inputStyle}
                />
              </div>

              {/* TLS Toggle */}
              <div style={{ display: 'flex', alignItems: 'center', gap: '10px' }}>
                <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer' }}>
                  <input
                    type="checkbox"
                    checked={form.tls}
                    onChange={e => setForm(prev => ({ ...prev, tls: e.target.checked }))}
                    style={{ width: '16px', height: '16px', accentColor: '#00d4ff' }}
                  />
                  <span style={{ fontSize: '13px', color: 'var(--color-text-primary)' }}>
                    TLS/SSL verwenden
                  </span>
                </label>
              </div>
            </div>

            {/* Modal Actions */}
            <div style={{ display: 'flex', gap: '10px', justifyContent: 'flex-end', marginTop: '24px', paddingTop: '16px', borderTop: '1px solid var(--color-border)' }}>
              <button
                onClick={closeModal}
                style={{
                  background: 'transparent',
                  border: '1px solid var(--color-border)',
                  borderRadius: '8px',
                  padding: '8px 16px',
                  fontSize: '13px',
                  color: 'var(--color-text-primary)',
                  cursor: 'pointer',
                }}
              >
                Abbrechen
              </button>
              <button
                onClick={handleSave}
                disabled={createMutation.isPending || updateMutation.isPending}
                style={{
                  background: 'linear-gradient(135deg, #00d4ff 0%, #8b5cf6 100%)',
                  color: '#0a0f1a',
                  border: 'none',
                  borderRadius: '8px',
                  padding: '8px 20px',
                  fontSize: '13px',
                  fontWeight: 600,
                  cursor: 'pointer',
                  boxShadow: '0 2px 8px rgba(0, 212, 255, 0.25)',
                  opacity: (createMutation.isPending || updateMutation.isPending) ? 0.7 : 1,
                }}
              >
                {(createMutation.isPending || updateMutation.isPending) ? 'Speichern...' : 'Speichern'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default EmailAccountsPage
