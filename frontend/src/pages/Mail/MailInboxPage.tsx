import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { mailApi, projectApi } from '../../services/api'
import styles from './Mail.module.scss'

type TabKey = 'general' | 'invoices' | 'personal'

interface Email {
  id: string
  from: string
  subject: string
  preview: string
  body?: string
  date: string
  isRead: boolean
  project?: string
  project_id?: string
  ai_confidence?: number
  tab: TabKey
}

const DEMO_EMAILS: Email[] = [
  { id: '1', from: 'Stadt München - Kulturamt', subject: 'Genehmigung Stadtfest 2026 - Aufbauzeiten', preview: 'Sehr geehrter Herr Berger, hiermit bestätigen wir die Aufbauzeiten für das Stadtfest am Marienplatz. Der Aufbau kann ab...', date: '2026-03-23T09:30:00Z', isRead: false, project: 'Stadtfest München 2026', project_id: '1', ai_confidence: 0.95, tab: 'general' },
  { id: '2', from: 'TechCorp GmbH - Events', subject: 'RE: Angebot Firmen-Gala - Änderungswünsche', preview: 'Vielen Dank für das ausführliche Angebot. Wir hätten noch einige Anpassungswünsche bezüglich der Lichtplanung...', date: '2026-03-23T08:15:00Z', isRead: false, project: 'Firmen-Gala TechCorp', project_id: '2', ai_confidence: 0.88, tab: 'general' },
  { id: '3', from: 'JBL Professional', subject: 'Rechnung #INV-2026-0847 - VTX Lautsprecher', preview: 'Anbei finden Sie die Rechnung für die Lieferung von 4x JBL VTX A12 Line Array Systemen. Zahlungsziel: 30 Tage...', date: '2026-03-22T16:00:00Z', isRead: true, tab: 'invoices' },
  { id: '4', from: 'Martin Professional', subject: 'Rechnung #MP-2026-1234', preview: 'Invoice for 6x MAC Aura XB units delivered on March 15th...', date: '2026-03-22T14:30:00Z', isRead: true, tab: 'invoices' },
  { id: '5', from: 'Festival GmbH', subject: 'Open Air Bodensee - Bühnenplanung', preview: 'Hallo zusammen, anbei die aktualisierten Bühnenpläne für das Festival im Juli. Bitte prüfen Sie die Maße...', date: '2026-03-22T11:00:00Z', isRead: true, project: 'Open Air Festival Bodensee', project_id: '3', ai_confidence: 0.92, tab: 'general' },
  { id: '6', from: 'Sarah Schmidt', subject: 'Urlaubsantrag KW 16', preview: 'Hallo Marco, ich wollte für die KW 16 Urlaub beantragen. Könntest du das bitte genehmigen?', date: '2026-03-21T17:00:00Z', isRead: true, tab: 'personal' },
  { id: '7', from: 'DGUV Prüfservice', subject: 'Erinnerung: E-Check fällig - 3 Geräte', preview: 'Für folgende Geräte steht die nächste elektrische Sicherheitsprüfung an: Chainmaster BGV-D8+, JBL VTX...', date: '2026-03-21T10:00:00Z', isRead: true, tab: 'general' },
  { id: '8', from: 'Thomas Müller', subject: 'Werkstattbericht KW 12', preview: 'Hier der Werkstattbericht für diese Woche. Yamaha CL5 Fader-Reparatur abgeschlossen, Martin MAC Aura noch in Arbeit...', date: '2026-03-20T18:00:00Z', isRead: true, tab: 'personal' },
]

const TAB_MAILBOX_MAP: Record<TabKey, string> = {
  general: 'general',
  invoices: 'invoices',
  personal: 'personal',
}

function MailInboxPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<TabKey>('general')
  const [selectedEmail, setSelectedEmail] = useState<Email | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [emails, setEmails] = useState<Email[]>(DEMO_EMAILS)
  const [assigningProject, setAssigningProject] = useState(false)

  // Fetch mails from API
  const { data: apiMails } = useQuery({
    queryKey: ['mails', 'inbound', activeTab],
    queryFn: () => mailApi.list({ direction: 'inbound', mailbox_type: TAB_MAILBOX_MAP[activeTab] }),
    retry: 1,
  })

  // Fetch projects for assignment dropdown
  const { data: projectsData } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectApi.list(1, 100),
  })

  const projects = projectsData?.items || projectsData?.data || []

  // Merge API data with demo fallback
  useEffect(() => {
    if (apiMails && Array.isArray(apiMails) && apiMails.length > 0) {
      const mapped: Email[] = apiMails.map((m: any) => ({
        id: m.id,
        from: m.from_address || m.from || '',
        subject: m.subject || '',
        preview: m.preview || m.body_text?.substring(0, 120) || '',
        body: m.body_html || m.body_text || '',
        date: m.received_at || m.created_at || '',
        isRead: m.is_read ?? false,
        project: m.project_name || m.project || undefined,
        project_id: m.project_id || undefined,
        ai_confidence: m.ai_confidence || undefined,
        tab: activeTab,
      }))
      setEmails(mapped)
    } else {
      setEmails(DEMO_EMAILS)
    }
  }, [apiMails, activeTab])

  // Mark as read mutation
  const markReadMutation = useMutation({
    mutationFn: (id: string) => mailApi.markAsRead(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mails'] })
    },
  })

  // Assign to project mutation
  const assignProjectMutation = useMutation({
    mutationFn: ({ mailId, projectId }: { mailId: string; projectId: string }) =>
      mailApi.assignToProject(mailId, projectId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mails'] })
      setAssigningProject(false)
    },
  })

  const handleSelectEmail = (email: Email) => {
    setSelectedEmail(email)
    setAssigningProject(false)
    if (!email.isRead) {
      // Mark as read locally
      setEmails(prev =>
        prev.map(e => e.id === email.id ? { ...e, isRead: true } : e)
      )
      // Mark as read on backend
      markReadMutation.mutate(email.id)
    }
  }

  const handleAssignProject = (projectId: string) => {
    if (!selectedEmail) return
    const project = projects.find((p: any) => p.id === projectId)
    assignProjectMutation.mutate(
      { mailId: selectedEmail.id, projectId },
      {
        onSuccess: () => {
          setEmails(prev =>
            prev.map(e =>
              e.id === selectedEmail.id
                ? { ...e, project: project?.name || '', project_id: projectId }
                : e
            )
          )
          setSelectedEmail(prev =>
            prev ? { ...prev, project: project?.name || '', project_id: projectId } : null
          )
        },
      }
    )
  }

  const filteredEmails = emails
    .filter(e => e.tab === activeTab)
    .filter(e =>
      e.from.toLowerCase().includes(searchQuery.toLowerCase()) ||
      e.subject.toLowerCase().includes(searchQuery.toLowerCase())
    )

  const unreadTotal = emails.filter(e => !e.isRead).length

  const tabs: { key: TabKey; label: string; count: number }[] = [
    { key: 'general', label: 'Allgemein', count: emails.filter(e => e.tab === 'general' && !e.isRead).length },
    { key: 'invoices', label: 'Rechnungen', count: emails.filter(e => e.tab === 'invoices' && !e.isRead).length },
    { key: 'personal', label: 'Persönlich', count: emails.filter(e => e.tab === 'personal' && !e.isRead).length },
  ]

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    const today = new Date()
    if (date.toDateString() === today.toDateString()) {
      return date.toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })
    }
    return date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })
  }

  const getConfidenceColor = (confidence: number) => {
    if (confidence >= 0.9) return '#10b981'
    if (confidence >= 0.7) return '#f59e0b'
    return '#ef4444'
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>
            Posteingang
            {unreadTotal > 0 && <span className={styles.unreadCount}>{unreadTotal}</span>}
          </h1>
          <p className={styles.subtitle}>Eingehende Nachrichten und E-Mails verwalten</p>
        </div>
        <div className={styles.headerActions}>
          <button className={styles.btnPrimary} onClick={() => navigate('/mail/compose')}>
            + Verfassen
          </button>
        </div>
      </div>

      {/* Tabs */}
      <div className={styles.tabBar}>
        {tabs.map(tab => (
          <button
            key={tab.key}
            className={`${styles.tab} ${activeTab === tab.key ? styles.tabActive : ''}`}
            onClick={() => { setActiveTab(tab.key); setSelectedEmail(null); }}
          >
            {tab.label}
            {tab.count > 0 && <span className={styles.tabBadge}>{tab.count}</span>}
          </button>
        ))}
      </div>

      {/* Filter */}
      <div className={styles.filterBar}>
        <input
          type="text"
          className={styles.searchInput}
          placeholder="E-Mails durchsuchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content: List + Detail */}
      <div className={styles.mailLayout}>
        {/* Email List */}
        <div className={styles.mailList}>
          {filteredEmails.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>📭</div>
              <h3 className={styles.emptyTitle}>Keine E-Mails</h3>
              <p className={styles.emptyDescription}>Kein Posteingang in dieser Kategorie.</p>
            </div>
          ) : (
            filteredEmails.map(email => (
              <div
                key={email.id}
                className={`${styles.mailItem} ${!email.isRead ? styles.mailItemUnread : ''} ${selectedEmail?.id === email.id ? styles.mailItemSelected : ''}`}
                onClick={() => handleSelectEmail(email)}
              >
                <div className={styles.mailItemHeader}>
                  {!email.isRead && <span className={styles.unreadDot} />}
                  <span className={`${styles.mailFrom} ${!email.isRead ? styles.mailFromBold : ''}`}>
                    {email.from}
                  </span>
                  <span className={styles.mailDate}>{formatDate(email.date)}</span>
                </div>
                <div className={`${styles.mailSubject} ${!email.isRead ? styles.mailSubjectBold : ''}`}>
                  {email.subject}
                </div>
                <div className={styles.mailPreview}>{email.preview}</div>
                <div style={{ display: 'flex', gap: '6px', alignItems: 'center', flexWrap: 'wrap' }}>
                  {email.project && (
                    <span className={styles.projectBadge}>{email.project}</span>
                  )}
                  {email.ai_confidence && (
                    <span
                      style={{
                        display: 'inline-flex',
                        padding: '2px 6px',
                        borderRadius: '9999px',
                        fontSize: '10px',
                        fontWeight: 600,
                        background: `${getConfidenceColor(email.ai_confidence)}20`,
                        color: getConfidenceColor(email.ai_confidence),
                        border: `1px solid ${getConfidenceColor(email.ai_confidence)}40`,
                      }}
                      title={`KI-Zuordnung: ${Math.round(email.ai_confidence * 100)}% Konfidenz`}
                    >
                      KI {Math.round(email.ai_confidence * 100)}%
                    </span>
                  )}
                </div>
              </div>
            ))
          )}
        </div>

        {/* Email Detail */}
        <div className={styles.mailDetail}>
          {selectedEmail ? (
            <>
              <div className={styles.detailHeader}>
                <h2 className={styles.detailSubject}>{selectedEmail.subject}</h2>
                <div className={styles.detailMeta}>
                  <span className={styles.detailFrom}>Von: {selectedEmail.from}</span>
                  <span className={styles.detailDate}>
                    {new Date(selectedEmail.date).toLocaleDateString('de-DE', {
                      day: '2-digit',
                      month: 'long',
                      year: 'numeric',
                      hour: '2-digit',
                      minute: '2-digit',
                    })}
                  </span>
                </div>
                <div style={{ display: 'flex', gap: '8px', alignItems: 'center', marginTop: '8px', flexWrap: 'wrap' }}>
                  {selectedEmail.project && (
                    <span className={styles.projectBadgeLarge}>{selectedEmail.project}</span>
                  )}
                  {selectedEmail.ai_confidence && (
                    <span
                      style={{
                        display: 'inline-flex',
                        padding: '4px 10px',
                        borderRadius: '9999px',
                        fontSize: '12px',
                        fontWeight: 600,
                        background: `${getConfidenceColor(selectedEmail.ai_confidence)}20`,
                        color: getConfidenceColor(selectedEmail.ai_confidence),
                        border: `1px solid ${getConfidenceColor(selectedEmail.ai_confidence)}40`,
                      }}
                    >
                      KI-Zuordnung: {Math.round(selectedEmail.ai_confidence * 100)}%
                    </span>
                  )}
                </div>
              </div>
              <div className={styles.detailBody}>
                {selectedEmail.body ? (
                  <div dangerouslySetInnerHTML={{ __html: selectedEmail.body }} />
                ) : (
                  <>
                    <p>{selectedEmail.preview}</p>
                    <p className={styles.detailPlaceholder}>
                      [Vollständiger E-Mail-Inhalt wird hier angezeigt]
                    </p>
                  </>
                )}
              </div>
              <div className={styles.detailActions}>
                <button className={styles.btnSecondary} onClick={() => navigate('/mail/compose')}>Antworten</button>
                <button className={styles.btnSecondary}>Weiterleiten</button>
                <button className={styles.btnSecondary}>Archivieren</button>
                {/* Project assignment */}
                {assigningProject ? (
                  <select
                    style={{
                      background: 'var(--glass-bg-input)',
                      border: '1px solid var(--color-border-strong)',
                      borderRadius: '8px',
                      padding: '6px 12px',
                      fontSize: '13px',
                      color: 'var(--color-text-primary)',
                      cursor: 'pointer',
                    }}
                    defaultValue=""
                    onChange={e => {
                      if (e.target.value) handleAssignProject(e.target.value)
                    }}
                    autoFocus
                    onBlur={() => setAssigningProject(false)}
                  >
                    <option value="" disabled>Projekt wählen...</option>
                    {projects.map((p: any) => (
                      <option key={p.id} value={p.id}>{p.name}</option>
                    ))}
                  </select>
                ) : (
                  <button
                    className={styles.btnSecondary}
                    onClick={() => setAssigningProject(true)}
                  >
                    Projekt zuordnen
                  </button>
                )}
              </div>
            </>
          ) : (
            <div className={styles.detailEmpty}>
              <div className={styles.emptyIcon}>📬</div>
              <p>Wählen Sie eine E-Mail aus, um sie zu lesen</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default MailInboxPage
