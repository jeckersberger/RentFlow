import { useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { crewApi, bookingApi } from '../../../services/api'
import styles from '../ProjectDetail.module.scss'

interface CrewTabProps {
  project: Project
}

interface CrewAssignment {
  id: string
  project_id: string
  crew_member_id: string
  crew_member_name?: string
  first_name?: string
  last_name?: string
  role: string
  hours?: number
  planned_hours?: number
  daily_rate?: number
  hourly_rate?: number
  rate?: number
  status?: string
}

interface BookingRequest {
  id: string
  assignment_id: string
  crew_member_id: string
  crew_member_name: string
  token: string
  status: string
  response_message?: string
  responded_at?: string
  created_at: string
}

interface CrewConflict {
  project_id: string
  project_name?: string
  member_id?: string
  member_name?: string
  start_date: string
  end_date: string
  role?: string
}

const bookingStatusConfig: Record<string, { label: string; bg: string; color: string }> = {
  pending: { label: 'Anfrage offen', bg: 'rgba(245, 158, 11, 0.15)', color: '#f59e0b' },
  accepted: { label: 'Zugesagt', bg: 'rgba(16, 185, 129, 0.15)', color: '#10b981' },
  declined: { label: 'Abgesagt', bg: 'rgba(239, 68, 68, 0.15)', color: '#ef4444' },
  alternative: { label: 'Alternative', bg: 'rgba(139, 92, 246, 0.15)', color: '#8b5cf6' },
}

export function CrewTab({ project }: CrewTabProps) {
  const queryClient = useQueryClient()
  const [sendingBooking, setSendingBooking] = useState<string | null>(null)
  const [bookingMessage, setBookingMessage] = useState('')
  const [showMessageFor, setShowMessageFor] = useState<string | null>(null)

  // Crew conflict detection state
  const [checkingConflictFor, setCheckingConflictFor] = useState<string | null>(null)
  const [crewConflicts, setCrewConflicts] = useState<CrewConflict[]>([])
  const [showCrewConflictWarning, setShowCrewConflictWarning] = useState<string | null>(null)

  const { data: assignmentsData, isLoading, error } = useQuery({
    queryKey: ['project-crew-assignments', project.id],
    queryFn: () => crewApi.listAssignments({ project_id: project.id } as any),
  })

  const { data: bookingsRaw } = useQuery({
    queryKey: ['booking-requests'],
    queryFn: () => bookingApi.list(),
  })

  // Handle different response shapes
  const assignments: CrewAssignment[] = Array.isArray(assignmentsData)
    ? assignmentsData
    : assignmentsData?.data ?? assignmentsData?.items ?? []

  const bookings: BookingRequest[] = Array.isArray(bookingsRaw)
    ? bookingsRaw
    : bookingsRaw?.data ?? []

  // Map assignment_id -> booking for quick lookup
  const bookingByAssignment = new Map<string, BookingRequest>()
  for (const b of bookings) {
    bookingByAssignment.set(b.assignment_id, b)
  }

  const totalCost = assignments.reduce((sum, p) => {
    const hours = p.planned_hours || p.hours || 0
    const rate = p.rate || p.hourly_rate || p.daily_rate || 0
    return sum + (hours * rate)
  }, 0)

  const getDisplayName = (a: CrewAssignment) => {
    if (a.crew_member_name) return a.crew_member_name
    if (a.first_name || a.last_name) return `${a.first_name || ''} ${a.last_name || ''}`.trim()
    return null
  }

  const formatDate = (d: string) =>
    new Date(d).toLocaleDateString('de-DE')

  // Check for crew double-booking before sending a booking request
  const handleCheckCrewConflictsAndBook = async (assignmentId: string, memberId: string) => {
    setCheckingConflictFor(assignmentId)
    setCrewConflicts([])
    setShowCrewConflictWarning(null)

    try {
      const conflicts = await crewApi.checkConflicts(
        memberId,
        project.start_date,
        project.end_date,
      )

      if (conflicts && conflicts.length > 0) {
        // Filter out current project from conflicts
        const otherConflicts = conflicts.filter(
          (c: CrewConflict) => c.project_id !== project.id
        )
        if (otherConflicts.length > 0) {
          setCrewConflicts(otherConflicts)
          setShowCrewConflictWarning(assignmentId)
          setCheckingConflictFor(null)
          return
        }
      }

      // No conflicts, proceed with booking
      setShowMessageFor(assignmentId)
    } catch {
      // If conflict check fails, proceed anyway
      setShowMessageFor(assignmentId)
    } finally {
      setCheckingConflictFor(null)
    }
  }

  const handleSendBooking = async (assignmentId: string) => {
    setSendingBooking(assignmentId)
    try {
      await bookingApi.create({ assignment_id: assignmentId, message: bookingMessage })
      setBookingMessage('')
      setShowMessageFor(null)
      setShowCrewConflictWarning(null)
      setCrewConflicts([])
      queryClient.invalidateQueries({ queryKey: ['booking-requests'] })
    } catch {
      // silently handle - user sees no change which indicates failure
    } finally {
      setSendingBooking(null)
    }
  }

  const handleForceBooking = (assignmentId: string) => {
    setShowCrewConflictWarning(null)
    setCrewConflicts([])
    setShowMessageFor(assignmentId)
  }

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>
          Crew & Team
          {!isLoading && ` (${assignments.length})`}
        </h3>
        <button className="btn btn--primary" disabled>
          + Position hinzufuegen
        </button>
      </div>

      {isLoading ? (
        <div className={styles.emptyState}>
          <p className={styles.emptyStateText}>Crew-Daten werden geladen...</p>
        </div>
      ) : error ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u26A0\uFE0F'}</div>
          <h4 className={styles.emptyStateTitle}>Fehler beim Laden</h4>
          <p className={styles.emptyStateText}>
            Crew-Daten konnten nicht geladen werden.
          </p>
        </div>
      ) : assignments.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F465}'}</div>
          <h4 className={styles.emptyStateTitle}>Keine Crew-Positionen</h4>
          <p className={styles.emptyStateText}>
            Fuegen Sie Positionen wie Soundtechniker, Lichttechniker oder Stagehand hinzu
            und weisen Sie Crew-Mitglieder zu.
          </p>
        </div>
      ) : (
        <>
          <div style={{ overflowX: 'auto' }}>
            <table className={styles.dataTable}>
              <thead>
                <tr>
                  <th>Rolle</th>
                  <th>Zugewiesene Person</th>
                  <th style={{ textAlign: 'right' }}>Stunden</th>
                  {assignments.some(a => a.rate || a.hourly_rate || a.daily_rate) && (
                    <>
                      <th style={{ textAlign: 'right' }}>Stundensatz</th>
                      <th style={{ textAlign: 'right' }}>Kosten</th>
                    </>
                  )}
                  <th style={{ textAlign: 'center' }}>Status</th>
                  <th style={{ textAlign: 'center' }}>Buchung</th>
                </tr>
              </thead>
              <tbody>
                {assignments.map((pos) => {
                  const name = getDisplayName(pos)
                  const hours = pos.planned_hours || pos.hours || 0
                  const rate = pos.rate || pos.hourly_rate || pos.daily_rate || 0
                  const hasRates = assignments.some(a => a.rate || a.hourly_rate || a.daily_rate)
                  const booking = bookingByAssignment.get(pos.id)
                  const bConfig = booking ? bookingStatusConfig[booking.status] : null

                  return (
                    <tr key={pos.id}>
                      <td>{pos.role || 'Techniker'}</td>
                      <td>
                        {name || (
                          <span style={{ color: 'var(--color-text-muted)' }}>Nicht zugewiesen</span>
                        )}
                      </td>
                      <td style={{ textAlign: 'right' }}>{hours > 0 ? `${hours}h` : '\u2014'}</td>
                      {hasRates && (
                        <>
                          <td style={{ textAlign: 'right' }}>
                            {rate > 0 ? `\u20AC${rate.toFixed(2)}/h` : '\u2014'}
                          </td>
                          <td style={{ textAlign: 'right' }}>
                            {hours > 0 && rate > 0
                              ? `\u20AC${(hours * rate).toFixed(2)}`
                              : '\u2014'}
                          </td>
                        </>
                      )}
                      <td style={{ textAlign: 'center' }}>
                        <span
                          style={{
                            display: 'inline-block',
                            padding: '2px 8px',
                            borderRadius: '9999px',
                            fontSize: '0.75rem',
                            fontWeight: 500,
                            background: pos.status === 'confirmed'
                              ? 'rgba(16, 185, 129, 0.15)'
                              : pos.status === 'cancelled'
                              ? 'rgba(239, 68, 68, 0.15)'
                              : 'rgba(0, 212, 255, 0.1)',
                            color: pos.status === 'confirmed'
                              ? '#10b981'
                              : pos.status === 'cancelled'
                              ? '#ef4444'
                              : 'var(--color-primary)',
                          }}
                        >
                          {pos.status || 'zugewiesen'}
                        </span>
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        {/* Crew conflict warning */}
                        {showCrewConflictWarning === pos.id && crewConflicts.length > 0 ? (
                          <div style={{
                            background: 'rgba(245, 158, 11, 0.15)',
                            border: '1px solid rgba(245, 158, 11, 0.3)',
                            borderRadius: '6px',
                            padding: '0.5rem',
                            textAlign: 'left',
                            fontSize: '0.75rem',
                          }}>
                            <div style={{ fontWeight: 600, color: '#f59e0b', marginBottom: '0.25rem' }}>
                              {'\u26A0\uFE0F'} Doppelbuchung!
                            </div>
                            <div style={{ color: 'var(--color-text-secondary)' }}>
                              {name} ist bereits eingeteilt:
                            </div>
                            <ul style={{ margin: '0.25rem 0', paddingLeft: '1rem' }}>
                              {crewConflicts.map((c, i) => (
                                <li key={i}>
                                  {c.project_name || c.project_id} ({formatDate(c.start_date)} - {formatDate(c.end_date)})
                                </li>
                              ))}
                            </ul>
                            <div style={{ display: 'flex', gap: '4px', marginTop: '0.25rem' }}>
                              <button
                                style={{
                                  padding: '2px 6px',
                                  fontSize: '0.7rem',
                                  background: 'linear-gradient(135deg, #00d4ff, #8b5cf6)',
                                  color: '#fff',
                                  border: 'none',
                                  borderRadius: '4px',
                                  cursor: 'pointer',
                                }}
                                onClick={() => handleForceBooking(pos.id)}
                              >
                                Trotzdem buchen
                              </button>
                              <button
                                style={{
                                  padding: '2px 6px',
                                  fontSize: '0.7rem',
                                  background: 'transparent',
                                  color: 'var(--color-text-muted)',
                                  border: '1px solid rgba(255,255,255,0.1)',
                                  borderRadius: '4px',
                                  cursor: 'pointer',
                                }}
                                onClick={() => { setShowCrewConflictWarning(null); setCrewConflicts([]) }}
                              >
                                Abbrechen
                              </button>
                            </div>
                          </div>
                        ) : booking ? (
                          <span
                            style={{
                              display: 'inline-block',
                              padding: '2px 8px',
                              borderRadius: '9999px',
                              fontSize: '0.75rem',
                              fontWeight: 500,
                              background: bConfig?.bg || 'rgba(0, 212, 255, 0.1)',
                              color: bConfig?.color || 'var(--color-primary)',
                            }}
                            title={booking.response_message || undefined}
                          >
                            {bConfig?.label || booking.status}
                          </span>
                        ) : pos.crew_member_id ? (
                          showMessageFor === pos.id ? (
                            <div style={{ display: 'flex', gap: '4px', alignItems: 'center', justifyContent: 'center' }}>
                              <input
                                type="text"
                                placeholder="Nachricht (optional)"
                                value={bookingMessage}
                                onChange={(e) => setBookingMessage(e.target.value)}
                                style={{
                                  padding: '2px 6px',
                                  fontSize: '0.75rem',
                                  background: 'rgba(255,255,255,0.05)',
                                  border: '1px solid rgba(255,255,255,0.1)',
                                  borderRadius: '4px',
                                  color: 'inherit',
                                  width: '120px',
                                }}
                              />
                              <button
                                style={{
                                  padding: '2px 8px',
                                  fontSize: '0.7rem',
                                  background: 'linear-gradient(135deg, #00d4ff, #8b5cf6)',
                                  color: '#fff',
                                  border: 'none',
                                  borderRadius: '4px',
                                  cursor: 'pointer',
                                  whiteSpace: 'nowrap',
                                }}
                                onClick={() => handleSendBooking(pos.id)}
                                disabled={sendingBooking === pos.id}
                              >
                                {sendingBooking === pos.id ? '...' : 'Senden'}
                              </button>
                              <button
                                style={{
                                  padding: '2px 6px',
                                  fontSize: '0.7rem',
                                  background: 'transparent',
                                  color: 'var(--color-text-muted)',
                                  border: '1px solid rgba(255,255,255,0.1)',
                                  borderRadius: '4px',
                                  cursor: 'pointer',
                                }}
                                onClick={() => { setShowMessageFor(null); setBookingMessage('') }}
                              >
                                X
                              </button>
                            </div>
                          ) : (
                            <button
                              style={{
                                padding: '2px 8px',
                                fontSize: '0.7rem',
                                background: 'linear-gradient(135deg, #00d4ff, #8b5cf6)',
                                color: '#fff',
                                border: 'none',
                                borderRadius: '4px',
                                cursor: 'pointer',
                                whiteSpace: 'nowrap',
                              }}
                              onClick={() => handleCheckCrewConflictsAndBook(pos.id, pos.crew_member_id)}
                              disabled={sendingBooking !== null || checkingConflictFor === pos.id}
                            >
                              {checkingConflictFor === pos.id ? 'Pruefe...' : 'Buchungsanfrage senden'}
                            </button>
                          )
                        ) : (
                          <span style={{ color: 'var(--color-text-muted)', fontSize: '0.75rem' }}>{'\u2014'}</span>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
              {totalCost > 0 && (
                <tfoot>
                  <tr>
                    <td colSpan={4}>Gesamtkosten Crew</td>
                    <td style={{ textAlign: 'right' }}>{'\u20AC'}{totalCost.toFixed(2)}</td>
                    <td />
                    <td />
                  </tr>
                </tfoot>
              )}
            </table>
          </div>
        </>
      )}
    </div>
  )
}
