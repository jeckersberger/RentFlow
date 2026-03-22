import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { AuditEntry, ChainValidationResult } from '../../types/audit'
import styles from './Audit.module.scss'

const mockAuditEntries: AuditEntry[] = [
  {
    id: '1',
    timestamp: '2026-03-22T14:35:00Z',
    service: 'equipment-service',
    operation: 'UPDATE',
    entity_type: 'equipment',
    entity_id: 'eq-001',
    user_id: 'user-123',
    user_name: 'Thomas Müller',
    changes: { price: { old: 150, new: 175 }, status: { old: 'available', new: 'maintenance' } },
    status: 'success',
    checksum: 'a8f3e2c1b9d4e7f2a5c8...',
    ip_address: '192.168.1.100',
  },
  {
    id: '2',
    timestamp: '2026-03-22T13:20:00Z',
    service: 'project-service',
    operation: 'CREATE',
    entity_type: 'project',
    entity_id: 'proj-042',
    user_id: 'user-456',
    user_name: 'Maria Schmidt',
    changes: { created: { name: 'Stadtfest München 2026', status: 'planning' } },
    status: 'success',
    checksum: 'b7e2f1a9c3d5e8g1h4i9...',
    ip_address: '192.168.1.101',
  },
  {
    id: '3',
    timestamp: '2026-03-22T12:15:00Z',
    service: 'crew-service',
    operation: 'DELETE',
    entity_type: 'crew',
    entity_id: 'crew-089',
    user_id: 'user-789',
    user_name: 'Admin User',
    changes: { deleted: { name: 'Archived Member' } },
    status: 'success',
    checksum: 'c6d1e8f2a5b9c3d7e2f1...',
    ip_address: '192.168.1.102',
  },
  {
    id: '4',
    timestamp: '2026-03-22T11:45:00Z',
    service: 'invoice-service',
    operation: 'EXPORT',
    entity_type: 'invoice',
    entity_id: 'inv-567',
    user_id: 'user-123',
    user_name: 'Thomas Müller',
    changes: { exported: { format: 'PDF', filename: 'inv-567.pdf' } },
    status: 'success',
    checksum: 'd5c2b1f8a4e7d9c3f2e8...',
    ip_address: '192.168.1.100',
  },
  {
    id: '5',
    timestamp: '2026-03-22T10:30:00Z',
    service: 'auth-service',
    operation: 'LOGIN',
    entity_type: 'user',
    entity_id: 'user-123',
    user_id: 'user-123',
    user_name: 'Thomas Müller',
    changes: { login: { method: 'password', success: true } },
    status: 'success',
    checksum: 'e4f3c2b1a8d5e7f1c9d2...',
    ip_address: '192.168.1.100',
  },
  {
    id: '6',
    timestamp: '2026-03-22T09:55:00Z',
    service: 'equipment-service',
    operation: 'READ',
    entity_type: 'equipment',
    entity_id: 'eq-002',
    user_id: 'user-456',
    user_name: 'Maria Schmidt',
    changes: { accessed: { details: 'viewed' } },
    status: 'success',
    checksum: 'f3e8d1c9b2a5f7e4d8c1...',
    ip_address: '192.168.1.101',
  },
]

function AuditPage() {
  const [searchQuery, setSearchQuery] = useState('')
  const [filterOperation, setFilterOperation] = useState<string>('all')
  const [filterEntityType, setFilterEntityType] = useState<string>('all')
  const [filterUser, setFilterUser] = useState<string>('all')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [validationResult, setValidationResult] = useState<ChainValidationResult | null>(null)
  const [showValidation, setShowValidation] = useState(false)

  const { data: auditEntries = mockAuditEntries } = useQuery({
    queryKey: ['audit-log'],
    queryFn: async () => mockAuditEntries,
    staleTime: 1000 * 60 * 5,
  })

  const filteredEntries = auditEntries.filter(entry => {
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
    const result: ChainValidationResult = {
      valid: true,
      entries_checked: auditEntries.length,
      invalid_entries: 0,
      timestamp: new Date().toISOString(),
      verification_hash: 'ab12cd34ef56gh78ij90kl12mn34op56qr78st90uv...',
    }
    setValidationResult(result)
    setShowValidation(true)
  }

  const handleExport = () => {
    // Simulate export
    const csvContent = 'Timestamp,Service,Operation,Entity Type,Entity ID,User,Status\n' +
      filteredEntries.map(e =>
        `${e.timestamp},${e.service},${e.operation},${e.entity_type},${e.entity_id},${e.user_name},${e.status}`
      ).join('\n')

    const blob = new Blob([csvContent], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `audit-export-${new Date().toISOString().split('T')[0]}.csv`
    a.click()
  }

  const uniqueUsers = Array.from(new Set(auditEntries.map(e => e.user_id))).map(id => {
    const entry = auditEntries.find(e => e.user_id === id)
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

        {filteredEntries.length === 0 ? (
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
                {filteredEntries.map(entry => (
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
