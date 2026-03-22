import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { auditApi } from '../../services/api'
import type { AuditEntry, ChainValidationResult } from '../../types/audit'
import styles from './Audit.module.scss'

function AuditPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterOperation, setFilterOperation] = useState<string>('all')
  const [filterEntityType, setFilterEntityType] = useState<string>('all')
  const [filterUser, setFilterUser] = useState<string>('all')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [validationResult, setValidationResult] = useState<ChainValidationResult | null>(null)
  const [showValidation, setShowValidation] = useState(false)

  // Fetch audit logs
  const { data: auditEntries = [], isLoading: isLoadingAudit } = useQuery({
    queryKey: ['audit-log', { startDate, endDate, operation: filterOperation, entityType: filterEntityType, userId: filterUser }],
    queryFn: () => auditApi.logs({ start_date: startDate, end_date: endDate, operation: filterOperation, entity_type: filterEntityType, user_id: filterUser }),
    staleTime: 1000 * 60 * 5,
  })

  // Verify chain mutation
  const verifyMutation = useMutation({
    mutationFn: () => auditApi.verify(),
    onSuccess: (data) => {
      setValidationResult(data)
      setShowValidation(true)
    },
  })

  // Export mutation
  const exportMutation = useMutation({
    mutationFn: () => auditApi.export({ start_date: startDate, end_date: endDate, operation: filterOperation }),
    onSuccess: () => {
      // Handle export success
    },
  })

  const filteredEntries = (auditEntries as AuditEntry[]).filter(entry => {
    const matchesSearch = searchQuery === '' ||
      entry.entity_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.user_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      entry.service.toLowerCase().includes(searchQuery.toLowerCase())

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
        matchesDateRange = matchesDateRange && entryTime <= new Date(endDate).getTime()
      }
    }

    return matchesSearch && matchesOperation && matchesEntityType && matchesUser && matchesDateRange
  })

  const handleValidateChain = () => {
    verifyMutation.mutate()
  }

  const handleExport = () => {
    exportMutation.mutate()
  }

  const uniqueUsers = Array.from(new Set((auditEntries as AuditEntry[]).map(e => e.user_id))).map(id => {
    const entry = (auditEntries as AuditEntry[]).find(e => e.user_id === id)
    return { id, name: entry?.user_name || id }
  })

  return (
    <div className={styles.auditPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>📋 Audit-Log</h1>
          <p className={styles.subtitle}>Volltext-Suche und GoBD-Prüfbericht</p>
        </div>
        <div className={styles.headerActions}>
          <button
            className="btn btn--primary"
            onClick={handleValidateChain}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            ✓ Chain validieren
          </button>
          <button
            className="btn"
            onClick={handleExport}
            style={{ padding: 'var(--spacing-3) var(--spacing-5)' }}
          >
            📥 GoBD Export
          </button>
        </div>
      </div>

      {/* Validation Result */}
      {showValidation && validationResult && (
        <div className={styles.validationResult}>
          <div className={styles.validationHeader}>
            <span style={{ fontSize: 'var(--font-size-xl)' }}>
              {validationResult.valid ? '✓' : '✗'}
            </span>
            <div>
              <h3 className={styles.validationTitle}>
                {validationResult.valid ? 'Chain-Validierung erfolgreich' : 'Chain-Validierung fehlgeschlagen'}
              </h3>
              <p className={styles.validationTime}>
                {new Date(validationResult.timestamp).toLocaleString('de-DE')}
              </p>
            </div>
          </div>
          <div className={styles.validationDetails}>
            <span><strong>Einträge geprüft:</strong> {validationResult.entries_checked}</span>
            <span><strong>Ungültige Einträge:</strong> {validationResult.invalid_entries}</span>
            <span><strong>Verifikations-Hash:</strong> {validationResult.verification_hash.substring(0, 30)}...</span>
          </div>
          <button
            className="btn btn--sm"
            onClick={() => setShowValidation(false)}
          >
            Schließen
          </button>
        </div>
      )}

      {/* Filters */}
      <section className={styles.filtersSection}>
        <div className={styles.searchBox}>
          <input
            type="text"
            className={styles.searchInput}
            placeholder="Volltext-Suche (Entity ID, User, Service)..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
        </div>

        <div className={styles.filtersGrid}>
          <div className={styles.filterGroup}>
            <label htmlFor="operation-filter" className={styles.filterLabel}>Operation:</label>
            <select
              id="operation-filter"
              className={styles.filterSelect}
              value={filterOperation}
              onChange={(e) => setFilterOperation(e.target.value)}
            >
              <option value="all">Alle</option>
              <option value="CREATE">CREATE</option>
              <option value="READ">READ</option>
              <option value="UPDATE">UPDATE</option>
              <option value="DELETE">DELETE</option>
              <option value="EXPORT">EXPORT</option>
              <option value="LOGIN">LOGIN</option>
            </select>
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="entity-filter" className={styles.filterLabel}>Entity Type:</label>
            <select
              id="entity-filter"
              className={styles.filterSelect}
              value={filterEntityType}
              onChange={(e) => setFilterEntityType(e.target.value)}
            >
              <option value="all">Alle</option>
              <option value="equipment">Equipment</option>
              <option value="project">Project</option>
              <option value="crew">Crew</option>
              <option value="invoice">Invoice</option>
              <option value="user">User</option>
              <option value="system">System</option>
            </select>
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="user-filter" className={styles.filterLabel}>User:</label>
            <select
              id="user-filter"
              className={styles.filterSelect}
              value={filterUser}
              onChange={(e) => setFilterUser(e.target.value)}
            >
              <option value="all">Alle</option>
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
              onChange={(e) => setStartDate(e.target.value)}
            />
          </div>

          <div className={styles.filterGroup}>
            <label htmlFor="end-date" className={styles.filterLabel}>Bis:</label>
            <input
              id="end-date"
              type="date"
              className={styles.filterInput}
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
            />
          </div>
        </div>
      </section>

      {/* Audit Entries Table */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Audit-Einträge ({filteredEntries.length})</h2>

        {isLoadingAudit ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>⏳</div>
            <p className={styles.emptyText}>Laden...</p>
          </div>
        ) : filteredEntries.length === 0 ? (
          <div className={styles.emptyState}>
            <div className={styles.emptyIcon}>📋</div>
            <h3 className={styles.emptyTitle}>Keine Einträge gefunden</h3>
            <p className={styles.emptyText}>Versuchen Sie, Ihre Suchkriterien zu ändern.</p>
          </div>
        ) : (
          <div className={styles.tableWrapper}>
            <table className={styles.table}>
              <thead>
                <tr>
                  <th>Zeit</th>
                  <th>Service</th>
                  <th>Operation</th>
                  <th>Entity</th>
                  <th>User</th>
                  <th>Checksum</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {filteredEntries.map((entry: AuditEntry) => (
                  <tr key={entry.id} className={styles.tableRow}>
                    <td className={styles.cellTime}>
                      {new Date(entry.timestamp).toLocaleString('de-DE', {
                        year: '2-digit',
                        month: '2-digit',
                        day: '2-digit',
                        hour: '2-digit',
                        minute: '2-digit',
                        second: '2-digit',
                      })}
                    </td>
                    <td className={styles.cellService}>{entry.service}</td>
                    <td className={styles.cellOperation}>
                      <span className={`${styles.operationBadge} ${styles[`operation--${entry.operation.toLowerCase()}`]}`}>
                        {entry.operation}
                      </span>
                    </td>
                    <td className={styles.cellEntity}>
                      <div className={styles.entityInfo}>
                        <span className={styles.entityType}>{entry.entity_type}</span>
                        <span className={styles.entityId}>{entry.entity_id}</span>
                      </div>
                    </td>
                    <td className={styles.cellUser}>{entry.user_name}</td>
                    <td className={styles.cellChecksum}>
                      <code className={styles.checksum}>{entry.checksum.substring(0, 15)}...</code>
                    </td>
                    <td className={styles.cellStatus}>
                      <span className={`${styles.statusBadge} ${styles[`status--${entry.status}`]}`}>
                        {entry.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}

export default AuditPage
