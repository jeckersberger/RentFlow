import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { Send } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

function EmailSettingsPage() {
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [testResult, setTestResult] = useState<{ ok: boolean; message: string } | null>(null)

  const [form, setForm] = useState({
    host: '',
    port: '587',
    username: '',
    password: '',
    tls: true,
    sender_name: '',
    sender_email: '',
    bcc_email: '',
  })
  const [signature, setSignature] = useState('')

  const { data: smtpConfig, isLoading: smtpLoading } = useQuery({
    queryKey: ['config', 'email.smtp'],
    queryFn: () => configApi.get('email.smtp'),
  })

  const { data: sigConfig, isLoading: sigLoading } = useQuery({
    queryKey: ['config', 'email.signature'],
    queryFn: () => configApi.get('email.signature'),
  })

  useEffect(() => {
    if (smtpConfig) {
      setForm({
        host: smtpConfig.host || '',
        port: smtpConfig.port || '587',
        username: smtpConfig.username || '',
        password: smtpConfig.password || '',
        tls: smtpConfig.tls ?? true,
        sender_name: smtpConfig.sender_name || '',
        sender_email: smtpConfig.sender_email || '',
        bcc_email: smtpConfig.bcc_email || '',
      })
    }
  }, [smtpConfig])

  useEffect(() => {
    if (sigConfig) {
      setSignature(typeof sigConfig === 'string' ? sigConfig : sigConfig.text || '')
    }
  }, [sigConfig])

  const isLoading = smtpLoading || sigLoading

  const saveMutation = useMutation({
    mutationFn: async () => {
      await Promise.all([
        configApi.set('email.smtp', form),
        configApi.set('email.signature', { text: signature }),
      ])
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'email.smtp'] })
      queryClient.invalidateQueries({ queryKey: ['config', 'email.signature'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const testMutation = useMutation({
    mutationFn: () => configApi.testSmtp(),
    onSuccess: () => setTestResult({ ok: true, message: 'Test-E-Mail erfolgreich gesendet!' }),
    onError: () => setTestResult({ ok: false, message: 'SMTP-Test fehlgeschlagen. Bitte Konfiguration prüfen.' }),
  })

  const update = (field: string, value: string | boolean) =>
    setForm((prev) => ({ ...prev, [field]: value }))

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>E-Mail-Einstellungen</h1></div>
      <SkeletonCard count={2} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>E-Mail (SMTP)</h1>
        <p>Konfigurieren Sie den E-Mail-Versand über Ihren SMTP-Server.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">SMTP-Server</h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Host</label>
            <input className="sp-input" value={form.host} onChange={(e) => update('host', e.target.value)} placeholder="smtp.example.com" />
          </div>
          <div className="sp-field">
            <label className="sp-label">Port</label>
            <input className="sp-input" value={form.port} onChange={(e) => update('port', e.target.value)} placeholder="587" />
          </div>
          <div className="sp-field">
            <label className="sp-label">Benutzername</label>
            <input className="sp-input" value={form.username} onChange={(e) => update('username', e.target.value)} />
          </div>
          <div className="sp-field">
            <label className="sp-label">Passwort</label>
            <input className="sp-input" type="password" value={form.password} onChange={(e) => update('password', e.target.value)} />
          </div>
          <div className="sp-field sp-full">
            <label className="sp-toggle">
              <div className="sp-toggle__label">
                <strong>TLS/SSL verwenden</strong>
                <span>Verschlüsselte Verbindung zum SMTP-Server</span>
              </div>
              <input type="checkbox" checked={form.tls} onChange={(e) => update('tls', e.target.checked)} />
            </label>
          </div>
        </div>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Absender</h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Absendername</label>
            <input className="sp-input" value={form.sender_name} onChange={(e) => update('sender_name', e.target.value)} placeholder="CrateDesk GmbH" />
          </div>
          <div className="sp-field">
            <label className="sp-label">Absender-E-Mail</label>
            <input className="sp-input" type="email" value={form.sender_email} onChange={(e) => update('sender_email', e.target.value)} placeholder="info@rentflow.de" />
          </div>
          <div className="sp-field sp-full">
            <label className="sp-label">BCC (Kopie an)</label>
            <input className="sp-input" type="email" value={form.bcc_email} onChange={(e) => update('bcc_email', e.target.value)} placeholder="archiv@rentflow.de" />
          </div>
        </div>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">E-Mail-Signatur</h3>
        <textarea className="sp-textarea" value={signature} onChange={(e) => setSignature(e.target.value)} placeholder="Mit freundlichen Grüßen,&#10;Ihr CrateDesk-Team" />
      </div>

      <div className="sp-footer">
        {testResult && (
          <span className={testResult.ok ? 'sp-msg--success' : 'sp-msg--error'}>{testResult.message}</span>
        )}
        {saveSuccess && <span className="sp-msg--success">Gespeichert!</span>}
        {saveMutation.isError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button className="sp-btn sp-btn--secondary" onClick={() => testMutation.mutate()} disabled={testMutation.isPending}>
          <Send size={14} style={{ marginRight: 6, verticalAlign: 'middle' }} />
          {testMutation.isPending ? 'Teste...' : 'Test senden'}
        </button>
        <button className="sp-btn sp-btn--primary" onClick={() => saveMutation.mutate()} disabled={saveMutation.isPending}>
          {saveMutation.isPending ? 'Speichern...' : 'Speichern'}
        </button>
      </div>
    </div>
  )
}

export default EmailSettingsPage
