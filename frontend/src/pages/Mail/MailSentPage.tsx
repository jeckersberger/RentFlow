import { useState, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import DOMPurify from 'dompurify'
import { mailApi } from '../../services/api'
import styles from './Mail.module.scss'

interface SentEmail {
  id: string
  to: string
  subject: string
  preview: string
  body?: string
  date: string
  status: 'Zugestellt' | 'Geöffnet' | 'Fehler'
  project?: string
}

const STATUS_MAP: Record<string, SentEmail['status']> = {
  delivered: 'Zugestellt',
  opened: 'Geöffnet',
  read: 'Geöffnet',
  error: 'Fehler',
  failed: 'Fehler',
  sent: 'Zugestellt',
}

function MailSentPage() {
  const navigate = useNavigate()
  const [selectedEmail, setSelectedEmail] = useState<SentEmail | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [sentEmails, setSentEmails] = useState<SentEmail[]>([])

  // Fetch sent mails from API
  const { data: apiSentMails, isLoading } = useQuery({
    queryKey: ['mails', 'outbound'],
    queryFn: () => mailApi.list({ direction: 'outbound' }),
    retry: 1,
  })

  // Map API data to local state
  useEffect(() => {
    if (apiSentMails && Array.isArray(apiSentMails)) {
      const mapped: SentEmail[] = apiSentMails.map((m: any) => ({
        id: m.id,
        to: m.to_address || m.to || '',
        subject: m.subject || '',
        preview: m.preview || m.body_text?.substring(0, 120) || '',
        body: m.body_html || m.body_text || '',
        date: m.sent_at || m.created_at || '',
        status: STATUS_MAP[m.delivery_status] || STATUS_MAP[m.status] || 'Zugestellt',
        project: m.project_name || m.project || undefined,
      }))
      setSentEmails(mapped)
    } else {
      setSentEmails([])
    }
  }, [apiSentMails])

  const filteredEmails = sentEmails.filter(e =>
    e.to.toLowerCase().includes(searchQuery.toLowerCase()) ||
    e.subject.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const formatDate = (dateStr: string) => {
    const date = new Date(dateStr)
    const today = new Date()
    if (date.toDateString() === today.toDateString()) {
      return date.toLocaleTimeString('de-DE', { hour: '2-digit', minute: '2-digit' })
    }
    return date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit' })
  }

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'Zugestellt': return styles.statusDelivered
      case 'Geöffnet': return styles.statusOpened
      case 'Fehler': return styles.statusError
      default: return ''
    }
  }

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Gesendet</h1>
          <p className={styles.subtitle}>Gesendete Nachrichten und E-Mails einsehen</p>
        </div>
        <div className={styles.headerActions}>
          <button className={styles.btnPrimary} onClick={() => navigate('/mail/compose')}>
            + Verfassen
          </button>
        </div>
      </div>

      {/* Filter */}
      <div className={styles.filterBar}>
        <input
          type="text"
          className={styles.searchInput}
          placeholder="Gesendete E-Mails durchsuchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content: List + Detail */}
      <div className={styles.mailLayout}>
        {/* Email List */}
        <div className={styles.mailList}>
          {isLoading ? (
            <div className={styles.emptyState}>
              <h3 className={styles.emptyTitle}>Laden...</h3>
              <p className={styles.emptyDescription}>Gesendete E-Mails werden geladen.</p>
            </div>
          ) : filteredEmails.length === 0 ? (
            <div className={styles.emptyState}>
              <div className={styles.emptyIcon}>{'📤'}</div>
              <h3 className={styles.emptyTitle}>Keine gesendeten E-Mails</h3>
              <p className={styles.emptyDescription}>
                {searchQuery ? 'Keine E-Mails gefunden. Versuchen Sie eine andere Suche.' : 'Sie haben noch keine E-Mails gesendet.'}
              </p>
            </div>
          ) : (
            filteredEmails.map(email => (
              <div
                key={email.id}
                className={`${styles.mailItem} ${selectedEmail?.id === email.id ? styles.mailItemSelected : ''}`}
                onClick={() => setSelectedEmail(email)}
              >
                <div className={styles.mailItemHeader}>
                  <span className={styles.mailFrom}>An: {email.to}</span>
                  <span className={styles.mailDate}>{formatDate(email.date)}</span>
                </div>
                <div className={styles.mailSubject}>{email.subject}</div>
                <div className={styles.mailPreview}>{email.preview}</div>
                <div className={styles.sentMeta}>
                  <span className={`${styles.sentStatus} ${getStatusClass(email.status)}`}>
                    {email.status}
                  </span>
                  {email.project && (
                    <span className={styles.projectBadge}>{email.project}</span>
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
                  <span className={styles.detailFrom}>An: {selectedEmail.to}</span>
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
                <div className={styles.detailStatusRow}>
                  <span className={`${styles.sentStatusLarge} ${getStatusClass(selectedEmail.status)}`}>
                    {selectedEmail.status}
                  </span>
                  {selectedEmail.project && (
                    <span className={styles.projectBadgeLarge}>{selectedEmail.project}</span>
                  )}
                </div>
              </div>
              <div className={styles.detailBody}>
                {selectedEmail.body ? (
                  <div dangerouslySetInnerHTML={{ __html: DOMPurify.sanitize(selectedEmail.body) }} />
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
                <button className={styles.btnSecondary}>Erneut senden</button>
                <button className={styles.btnSecondary}>Weiterleiten</button>
              </div>
            </>
          ) : (
            <div className={styles.detailEmpty}>
              <div className={styles.emptyIcon}>📧</div>
              <p>Wählen Sie eine E-Mail aus, um sie zu lesen</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default MailSentPage
