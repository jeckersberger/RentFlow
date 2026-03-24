import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Project } from '../../../types/project'
import { projectApi } from '../../../services/api'
import { useAuthStore } from '../../../stores/authStore'
import styles from '../ProjectDetail.module.scss'

interface CommunicationTabProps {
  project: Project
}

interface ProjectNote {
  id: string
  text: string
  author: string
  created_at: string
}

export function CommunicationTab({ project }: CommunicationTabProps) {
  const queryClient = useQueryClient()
  const user = useAuthStore((s) => s.user)
  const [noteText, setNoteText] = useState('')
  const [showInput, setShowInput] = useState(false)

  // We store notes in the project's notes field as JSON.
  // Parse existing notes from the project.
  const parseNotes = (notesStr?: string): ProjectNote[] => {
    if (!notesStr) return []
    try {
      const parsed = JSON.parse(notesStr)
      if (Array.isArray(parsed)) return parsed
      // If it's a plain string (old format), wrap it as a single note
      return []
    } catch {
      // If it's a plain text notes field, show it as a single legacy note
      if (notesStr.trim()) {
        return [{
          id: 'legacy',
          text: notesStr,
          author: 'System',
          created_at: project.created_at || new Date().toISOString(),
        }]
      }
      return []
    }
  }

  // We'll use a local query that re-fetches the project to get the latest notes
  const { data: projectData } = useQuery({
    queryKey: ['project-detail-notes', project.id],
    queryFn: () => projectApi.getById(project.id),
    initialData: project,
  })

  const currentProject = projectData || project
  const notes = parseNotes(currentProject.notes)

  const saveMutation = useMutation({
    mutationFn: async (newNote: ProjectNote) => {
      const existingNotes = parseNotes(currentProject.notes).filter(n => n.id !== 'legacy')
      // Preserve legacy text as first note if present
      const legacyNote = parseNotes(currentProject.notes).find(n => n.id === 'legacy')
      const allNotes = legacyNote
        ? [{ ...legacyNote, id: `note-${Date.now() - 1}` }, ...existingNotes, newNote]
        : [...existingNotes, newNote]

      return projectApi.update(project.id, {
        notes: JSON.stringify(allNotes),
      })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['project-detail-notes', project.id] })
      queryClient.invalidateQueries({ queryKey: ['project', project.id] })
      setNoteText('')
      setShowInput(false)
    },
  })

  const handleAddNote = () => {
    if (!noteText.trim()) return
    const newNote: ProjectNote = {
      id: `note-${Date.now()}`,
      text: noteText.trim(),
      author: user?.name || user?.email || 'Unbekannt',
      created_at: new Date().toISOString(),
    }
    saveMutation.mutate(newNote)
  }

  const sortedNotes = [...notes].sort(
    (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
  )

  return (
    <div>
      <div className={styles.sectionHeader}>
        <h3 className={styles.sectionTitle}>Kommunikation</h3>
        <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
          <button
            className="btn btn--primary"
            onClick={() => setShowInput(!showInput)}
          >
            {showInput ? 'Abbrechen' : '+ Notiz hinzufuegen'}
          </button>
        </div>
      </div>

      {/* Add Note Input */}
      {showInput && (
        <div style={{
          marginBottom: '1.5rem',
          padding: '1rem',
          background: 'var(--color-surface)',
          border: '1px solid rgba(255,255,255,0.08)',
          borderRadius: '8px',
        }}>
          <textarea
            value={noteText}
            onChange={(e) => setNoteText(e.target.value)}
            placeholder="Notiz eingeben..."
            rows={4}
            style={{
              width: '100%',
              padding: '0.75rem',
              background: 'rgba(255,255,255,0.05)',
              border: '1px solid rgba(255,255,255,0.1)',
              borderRadius: '6px',
              color: 'inherit',
              fontSize: '0.9rem',
              resize: 'vertical',
              fontFamily: 'inherit',
            }}
          />
          <div style={{ display: 'flex', gap: '0.5rem', marginTop: '0.75rem', justifyContent: 'flex-end' }}>
            <button
              className="btn btn--secondary"
              onClick={() => { setShowInput(false); setNoteText('') }}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={handleAddNote}
              disabled={!noteText.trim() || saveMutation.isPending}
            >
              {saveMutation.isPending ? 'Speichern...' : 'Notiz speichern'}
            </button>
          </div>
          {saveMutation.isError && (
            <div style={{
              marginTop: '0.5rem',
              padding: '0.5rem',
              background: 'rgba(239, 68, 68, 0.1)',
              border: '1px solid rgba(239, 68, 68, 0.3)',
              borderRadius: '4px',
              color: '#ef4444',
              fontSize: '0.8rem',
            }}>
              Fehler beim Speichern. Bitte versuchen Sie es erneut.
            </div>
          )}
        </div>
      )}

      {/* Notes Timeline */}
      {sortedNotes.length === 0 ? (
        <div className={styles.emptyState}>
          <div className={styles.emptyStateIcon}>{'\u{1F4AC}'}</div>
          <h4 className={styles.emptyStateTitle}>Keine Eintraege</h4>
          <p className={styles.emptyStateText}>
            Noch keine Kommunikation zu diesem Projekt. Fuegen Sie eine Notiz hinzu.
          </p>
        </div>
      ) : (
        <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          {sortedNotes.map((note) => (
            <div
              key={note.id}
              style={{
                padding: '1rem',
                background: 'var(--color-surface)',
                border: '1px solid rgba(255,255,255,0.06)',
                borderRadius: '8px',
                borderLeft: '3px solid var(--color-primary, #00d4ff)',
              }}
            >
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '0.5rem',
              }}>
                <span style={{
                  fontWeight: 600,
                  fontSize: '0.85rem',
                  color: 'var(--color-text-primary)',
                }}>
                  {note.author}
                </span>
                <span style={{
                  fontSize: '0.75rem',
                  color: 'var(--color-text-muted)',
                }}>
                  {new Date(note.created_at).toLocaleDateString('de-DE', {
                    day: '2-digit',
                    month: '2-digit',
                    year: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </span>
              </div>
              <p style={{
                margin: 0,
                fontSize: '0.9rem',
                color: 'var(--color-text-secondary)',
                lineHeight: 1.5,
                whiteSpace: 'pre-wrap',
              }}>
                {note.text}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
