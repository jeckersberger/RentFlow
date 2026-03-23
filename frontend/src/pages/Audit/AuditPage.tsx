import { useState, useMemo, Fragment } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { auditApi } from '../../services/api'
import type { AuditEntry, ChainValidationResult } from '../../types/audit'
import styles from './Audit.module.scss'

// No mock data - use real API only

const PAGE_SIZE = 15

function AuditPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterOperation, setFilterOperation] = useState<string>('all')
  const [filterEntityType, setFilterEntityType] = useState<string>('all')
  const [filterUser, setFilterUser] = useState<string>('all')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [validationResult, setValidationResult] = useState<ChainValidationResult | null>(null)
  const [showValidation, setShowValidation] = useState(false)
  const [expandedRow, setExpandedRow] = useState<string | null>(null)
  const [currentPage, setCurrentPage] = useState(1)

  // Fetch audit logs
  const { data: rawAuditEntries, isLoading: isLoadingAudit } = useQuery({
    queryKey: ['audit-log', { startDate, endDate, operation: filterOperation, entityType: filterEntityType, userId: filterUser }],
    queryFn: () => auditApi.logs({ start_date: startDate, end_date: endDate, operation: filterOperation, entity_type: filterEntityType, user_id: filterUser }),
    staleTime: 1000 * 60 * 5,
  })

  // Use API data, handle various response formats
  const auditEntries: AuditEntry[] = useMemo(() => {
    if (!rawAuditEntries) return []
    const raw = rawAuditEntries as any
    if (Array.isArray(raw)) return raw
    if (Array.isArray(raw.data)) return raw.data
    if (Array.isArray(raw.items)) return raw.items
    return []
  }, [rawAuditEntries])

  // Verify chain mutation
  const verifyMutation = useMutation({
    mutationFn: () => auditApi.verify(),
    onSuccess: (data) => {
      const result: ChainValidationResult = {
        valid: data?.valid ?? true,
        entries_checked: data?.entries_checked ?? auditEntries.length,
        invalid_entries: data?.invalid_entries ?? 0,
        timestamp: data?.timestamp || new Date().toISOString(),
        verification_hash: data?.verification_hash || '',
      }
      setValidationResult(result)
      setShowValidation(true)
    },
    onError: (err: any) => {
      setValidationResult({
        valid: false,
        entries_checked: 0,
        invalid_entries: 0,
        timestamp: new Date().toISOString(),
        verification_hash: `Fehler: ${err?.message || 'Validierung fehlgeschlagen'}`,
      })
      setShowValidation(true)
    },
  })

  // Export mutation
  const exportMutation = useMutation({
    mutationFn: () => auditApi.export({ start_date: startDate, end_date: endDate, operation: filterOperation }),
    onSuccess: () => {
      alert('GoBD-Export wurde erfolgreich erstellt.')
    },
    onError: (err: any) => {
      alert(`GoBD-Export fehlgeschlagen: ${err?.message || 'Unbekannter Fehler'}`)
    },
  })

  const filteredEntries = useMemo(() => {
    return auditEntries.filter(entry => {
      const matchesSearch = searchQuery === '' ||
        entry.entity_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
        entry.user_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        entry.service.toLowerCase().includes(searchQuery.toLowerCase()) ||
        JSON.stringify(entry.changes).toLowerCase().includes(searchQuery.toLowerCase())

      const matchesOperation = filterOperation === 'all' || entry.operation === filterOperation
      const matchesEntityType = filterEntityType === 'all' || entry.entity_type === filterEntityType
      const matchesUser = filterUser === 'all' || entry.user_id === filterUser

      let matchesDateRange = true
      if (startDate || endDate) {
        const entryTime = new Date(entry.timestamp).getTime()
        if (startDate) {
          matchesDateRange = matchesDateRange && entryTime >= new Date(startDate).getTime()
        }
        if (endDate) {
          matchesDateRange = matchesDateRange && entryTime <= new Date(endDate + 'T23:59:59').getTime()
        }
      }

      return matchesSearch && matchesOperation && matchesEntityType && matchesUser && matchesDateRange
    })
  }, [auditEntries, searchQuery, filterOperation, filterEntityType, filterUser, startDate, endDate])

  // Pagination
  const totalPages = Math.ceil(filteredEntries.length / PAGE_SIZE)
  const paginatedEntries = filteredEntries.slice((currentPage - 1) * PAGE_SIZE, currentPage * PAGE_SIZE)

  const handleValidateChain = () => {
    verifyMutation.mutate()
  }

  const handleExport = () => {
    exportMutation.mutate()
  }

  const uniqueUsers = useMemo(() => {
    return Array.from(new Set(auditEntries.map(e => e.user_id))).map(id => {
      const entry = auditEntries.find(e => e.user_id === id)
      return { id, name: entry?.user_name || id }
    })
  }, [auditEntries])

  const toggleExpand = (id: string) => {
    setExpandedRow(prev => prev === id ? null : id)
  }

  const renderChanges = (changes: Record<string, unknown>) => {
    const entries = Object.entries(changes)
    if (entries.length === 0) return <span style={{ color: 'var(--color-text-secondary)' }}>Keine Details</span>

    return (
      <div style={{ display: 'flex', flexDirection: 'column', gap: 'var(--spacing-2)' }}>
        {entries.map(([key, val]) => {
          const change = val as any
          const oldVal = change?.old ?? '--'
          const newVal = change?.new ?? '--'
          return (
            <div key={key} style={{
              display: 'grid',
              gridTemplateColumns: '120px 1fr 24px 1fr',
              gap: 'var(--spacing-2)',
              alignItems: 'center',
              padding: 'var(--spacing-1) 0',
              borderBottom: '1px solid rgba(255,255,255,0.03)',
            }}>
              <span style={{ fontWeight: 600, color: 'var(--color-text-primary)', fontSize: 'var(--font-size-xs)' }}>{key}</span>
              <code style={{
                background: 'rgba(239, 68, 68, 0.1)',
                color: '#ef4444',
                padding: '2px 6px',
                borderRadius: '4px',
                fontSize: 'var(--font-size-xs)',
                wordBreak: 'break-all',
              }}>
                {typeof oldVal === 'object' ? JSON.stringify(oldVal) : String(oldVal)}
              </code>
              <span style={{ color: 'var(--color-text-secondary)', textAlign: 'center', fontSize: 'var(--font-size-sm)' }}>{'\u2192'}</span>
              <code style={{
                background: 'rgba(16, 185, 129, 0.1)',
                color: '#10b981',
                padding: '2px 6px',
                borderRadius: '4px',
                fontSize: 'var(--font-size-xs)',
                wordBreak: 'break-all',
              }}>
                {typeof newVal === 'object' ? JSON.stringify(newVal) : String(newVal)}
              </code>
            </div>
          )
        })}
      </div>
    )
  }

  const getOperationLabel = (op: string): string => {
    switch (op) {
      case 'CREATE': return 'Erstellt'
      case 'UPDATE': return 'Aktualisiert'
      case 'DELETE': return 'Geloescht'
      case 'READ': return 'Gelesen'
      case 'EXPORT': return 'Exportiert'
      case 'LOGIN': return 'Anmeldung'
      case 'LOGOUT': return 'Abmeldung'
      default: return op
    }
  }

  const getEntityTypeLabel = (et: string): string => {
    switch (et) {
      case 'equipment': return 'Equipment'
      case 'project': return 'Projekt'
      case 'invoice': return 'Rechnung'
      case 'user': return 'Benutzer'
      case 'crew': return 'Crew'
      case 'system': return 'System'
      default: return et
    }
  }

  return (
    <div className={styles.auditPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>Audit-Log</h1>
          <p className={styles.subtitle}>
            GoBD-konformes Protokoll aller Systemaktionen -- {filteredEntries.length} Eintraege
          </p>
        </div>
        <div className={styles.headerActions}>
          <button
            className="btn btn--primary"
            onClick={handleValidateChain}
            disabled={verifyMutation.isPending}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            {verifyMutation.isPending ? 'Validiere...' : 'Chain validieren'}
          </button>
          <button
            className="btn"
            onClick={handleExport}
            disabled={exportMutation.isPending}
            style={{
              padding: 'var(--spacing-3) var(--spacing-5)',
              background: 'var(--color-bg-secondary)',
              border: '1px solid var(--color-border)',
              color: 'var(--color-text-primary)',
            }}
          >
            {exportMutation.isPending ? 'Exportiere...' : 'GoBD Export'}
          </button>
        </div>
      </div>

      {/* Validation Result */}
      {showValidation && validationResult && (
        <div className={styles.validationResult} style={{
          borderColor: validationResult.valid ? '#10b981' : '#ef4444',
        }}>
          <div className={styles.validationHeader}>
            <span style={{
              fontSize: '2rem',
              color: validationResult.valid ? '#10b981' : '#ef4444',
              lineHeight: 1,
            }}>
              {validationResult.valid ? '\u2713' : '\u2717'}
            </span>
            <div>
              <h3 className={styles.validationTitle}>
                {validationResult.valid ? 'Chain-Validierung erfolgreich' : 'Chain-Validierung fehlgeschlagen'}
              </h3>
              <p className={styles.validationTime}>
                Geprueft am {new Date(validationResult.timestamp).toLocaleString('de-DE')}
              </p>
            </div>
          </div>
          <div className={styles.validationDetails}>
            <span><strong>Eintraege geprueft:</strong> {validationResult.entries_checked}</span>
            <span><strong>Ungueltige Eintraege:</strong> {validationResult.invalid_entries}</span>
            <span><strong>Verifikations-Hash:</strong> <code style={{ fontFamily: 'monospace', fontSize: 'var(--font-size-xs)' }}>{validationResult.verification_hash.substring(0, 40)}...</code></span>
          </div>
          <button
            className="btn btn--sm"
            onClick={() => setShowValidation(false)}
            style={{ background: 'var(--color-bg-tertiary)', border: '1px solid var(--color-border)', color: 'var(--color-text-primary)' }}
          >
            Schliessen
          </button>
        </div>
      )}

      {/* Filters */}
      <section className={styles.filtersSection}>
        <div className={styles.searchBox}>
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Volltext-Suche (Entity ID, User, Service, Aenderungen)..."
            value={searchQuery}
            onChange={(e) => { setSearchQuery(e.target.value); setCurrentPage(1) }}
          />
        </div>

        <div className={styles.filtersGrid}>
          <div className={styles.filterGroup}>
            <label htmlFor="operation-filter" className={styles.filterLabel}>Operation:</label>
            <select
              id="operation-filter"
              className={styles.filterSelect}
              value={filterOperation}
              onChange={(e) => { setFilterOperation(e.target.value); setCurrentPage(1) }}
            >
              <option value="all">Alle Operationen</option>
              <option value="CREATE">CREATE - Erstellen</option>
              <option value="UPDATE">UPDATE - Aktualisieren</option>
              <option value="DELETE">DELETE - Loeschen</option>
              <option value="READ">READ - Lesen</option>
              <option value="EXPORT">EXPORT - Exportieren</option>
              <option value="LOGIN">LOGIN - Anmeldung</option>
            </select>
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="entity-filter" className={styles.filterLabel}>Entitaet:</label>
            <select
              id="entity-filter"
              className={styles.filterSelect}
              value={filterEntityType}
              onChange={(e) => { setFilterEntityType(e.target.value); setCurrentPage(1) }}
            >
              <option value="all">Alle Typen</option>
              <option value="equipment">Equipment</option>
              <option value="project">Projekt</option>
              <option value="invoice">Rechnung</option>
              <option value="user">Benutzer</option>
              <option value="crew">Crew</option>
              <option value="system">System</option>
            </select>
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="user-filter" className={styles.filterLabel}>Benutzer:</label>
            <select
              id="user-filter"
              className={styles.filterSelect}
              value={filterUser}
              onChange={(e) => { setFilterUser(e.target.value); setCurrentPage(1) }}
            >
              <option value="all">Alle Benutzer</option>
              {uniqueUsers.map(user => (
                <option key={user.id} value={user.id}>{user.name}</option>
              ))}
            </select>
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="start-date" className={styles.filterLabel}>Von:</label>
            <input
              id="start-date"
              type="date"
              className={styles.filterInput}
              value={startDate}
              onChange={(e) => { setStartDate(e.target.value); setCurrentPage(1) }}
            />
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="end-date" className={styles.filterLabel}>Bis:</label>
            <input
              id="end-date"
              type="date"
              className={styles.filterInput}
              value={endDate}
              onChange={(e) => { setEndDate(e.target.value); setCurrentPage(1) }}
            />
          </div>
        </div>
      </section>

      {/* Audit Entries Table */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>
          Audit-Eintraege ({filteredEntries.length})
          {filteredEntries.length !== auditEntries.length && (
            <span style={{ fontWeight: 400, fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', marginLeft: 'var(--spacing-2)' }}>
              (gefiltert von {auditEntries.length})
            </span>
          )}
        </h2>

        {isLoadingAudit ? (
          <div className={styles.emptyState}>
            <div style={{ fontSize: 'var(--font-size-xl)', marginBottom: 'var(--spacing-3)' }}>Laden...</div>
            <p className={styles.emptyText}>Audit-Eintraege werden abgerufen...</p>
          </div>
        ) : filteredEntries.length === 0 ? (
          <div className={styles.emptyState}>
            <h3 className={styles.emptyTitle}>Keine Eintraege gefunden</h3>
            <p className={styles.emptyText}>
              {auditEntries.length === 0
                ? 'Es sind noch keine Audit-Eintraege vorhanden. Eintraege werden automatisch bei Systemaktionen erstellt.'
                : 'Versuchen Sie, Ihre Suchkriterien zu aendern.'}
            </p>
          </div>
        ) : (
          <>
            <div className={styles.tableWrapper}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th style={{ width: 30 }}></th>
                    <th>Zeitpunkt</th>
                    <th>Benutzer</th>
                    <th>Aktion</th>
                    <th>Entitaet</th>
                    <th>Details</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {paginatedEntries.map((entry: AuditEntry) => (
                    <Fragment key={entry.id}>
                      <tr
                        className={styles.tableRow}
                        onClick={() => toggleExpand(entry.id)}
                        style={{ cursor: 'pointer' }}
                      >
                        <td style={{ textAlign: 'center', color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)' }}>
                          {expandedRow === entry.id ? '\u25BC' : '\u25B6'}
                        </td>
                        <td className={styles.cellTime}>
                          {new Date(entry.timestamp).toLocaleString('de-DE', {
                            day: '2-digit', month: '2-digit', year: '2-digit',
                            hour: '2-digit', minute: '2-digit', second: '2-digit',
                          })}
                        </td>
                        <td className={styles.cellUser}>{entry.user_name}</td>
                        <td className={styles.cellOperation}>
                          <span className={`${styles.operationBadge} ${styles[`operation--${entry.operation.toLowerCase()}`]}`}>
                            {getOperationLabel(entry.operation)}
                          </span>
                        </td>
                        <td className={styles.cellEntity}>
                          <div className={styles.entityInfo}>
                            <span className={styles.entityType}>{getEntityTypeLabel(entry.entity_type)}</span>
                            <span className={styles.entityId}>{entry.entity_id}</span>
                          </div>
                        </td>
                        <td style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                          {Object.keys(entry.changes).length} Aenderung(en): {Object.keys(entry.changes).join(', ')}
                        </td>
                        <td className={styles.cellStatus}>
                          <span className={`${styles.statusBadge} ${styles[`status--${entry.status}`]}`}>
                            {entry.status === 'success' ? 'OK' : 'Fehler'}
                          </span>
                        </td>
                      </tr>
                      {expandedRow === entry.id && (
                        <tr key={`${entry.id}-details`}>
                          <td colSpan={7} style={{ padding: 0 }}>
                            <div style={{
                              padding: 'var(--spacing-4) var(--spacing-5)',
                              background: 'rgba(0, 212, 255, 0.02)',
                              borderTop: '1px solid rgba(0, 212, 255, 0.1)',
                              borderBottom: '1px solid rgba(0, 212, 255, 0.1)',
                            }}>
                              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 'var(--spacing-4)', marginBottom: 'var(--spacing-4)' }}>
                                <div>
                                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-1)', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>Service</div>
                                  <div style={{ color: 'var(--color-primary)', fontWeight: 500 }}>{entry.service}</div>
                                </div>
                                <div>
                                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-1)', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>IP-Adresse</div>
                                  <div style={{ fontFamily: 'monospace', fontSize: 'var(--font-size-sm)' }}>{entry.ip_address}</div>
                                </div>
                                <div style={{ gridColumn: '1 / -1' }}>
                                  <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-1)', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>Checksum</div>
                                  <code style={{
                                    display: 'block',
                                    padding: 'var(--spacing-2) var(--spacing-3)',
                                    background: 'var(--color-bg-tertiary)',
                                    borderRadius: 'var(--radius-sm)',
                                    fontSize: 'var(--font-size-xs)',
                                    fontFamily: 'monospace',
                                    color: 'var(--color-text-secondary)',
                                    wordBreak: 'break-all',
                                  }}>
                                    {entry.checksum}
                                  </code>
                                </div>
                              </div>
                              <div style={{ fontSize: 'var(--font-size-xs)', color: 'var(--color-text-secondary)', marginBottom: 'var(--spacing-2)', textTransform: 'uppercase', letterSpacing: '0.5px', fontWeight: 600 }}>
                                Aenderungen (Alt {'\u2192'} Neu)
                              </div>
                              {renderChanges(entry.changes)}
                            </div>
                          </td>
                        </tr>
                      )}
                    </Fragment>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Pagination */}
            {totalPages > 1 && (
              <div style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                gap: 'var(--spacing-2)',
                marginTop: 'var(--spacing-4)',
              }}>
                <button
                  className="btn btn--sm"
                  style={{ background: 'var(--color-bg-secondary)', border: '1px solid var(--color-border)', color: 'var(--color-text-primary)' }}
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                >
                  Zurueck
                </button>
                {Array.from({ length: totalPages }, (_, i) => i + 1).map(page => (
                  <button
                    key={page}
                    className="btn btn--sm"
                    style={{
                      background: currentPage === page ? 'var(--color-primary)' : 'var(--color-bg-secondary)',
                      border: '1px solid ' + (currentPage === page ? 'var(--color-primary)' : 'var(--color-border)'),
                      color: currentPage === page ? '#fff' : 'var(--color-text-primary)',
                      minWidth: 36,
                    }}
                    onClick={() => setCurrentPage(page)}
                  >
                    {page}
                  </button>
                ))}
                <button
                  className="btn btn--sm"
                  style={{ background: 'var(--color-bg-secondary)', border: '1px solid var(--color-border)', color: 'var(--color-text-primary)' }}
                  onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                  disabled={currentPage === totalPages}
                >
                  Weiter
                </button>
              </div>
            )}
          </>
        )}
      </section>
    </div>
  )
}

export default AuditPage
