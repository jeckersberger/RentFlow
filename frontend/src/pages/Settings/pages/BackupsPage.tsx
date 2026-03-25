import { useState, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { configApi } from '../../../services/api'
import { useNotificationStore } from '../../../stores/notificationStore'
import { Download, RefreshCw, HardDrive, Shield, RotateCcw, Upload, AlertTriangle, Trash2, Clock } from 'lucide-react'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

interface BackupEntry {
  filename: string
  size: number
  size_human: string
  created_at: string
  type: string
  status: string
}

function BackupsPage() {
  const queryClient = useQueryClient()
  const addNotification = useNotificationStore((s) => s.addNotification)
  const [restoreConfirmFilename, setRestoreConfirmFilename] = useState<string | null>(null)
  const [restoreConfirmDate, setRestoreConfirmDate] = useState<string>('')
  const [deleteConfirmFilename, setDeleteConfirmFilename] = useState<string | null>(null)
  const [uploadFile, setUploadFile] = useState<File | null>(null)
  const [downloadingFile, setDownloadingFile] = useState<string | null>(null)
  const fileInputRef = useRef<HTMLInputElement>(null)

  const { data: backups, isLoading } = useQuery<BackupEntry[]>({
    queryKey: ['system', 'backups'],
    queryFn: () => configApi.backupList(),
    refetchInterval: 30000,
  })

  const backupNowMutation = useMutation({
    mutationFn: () => configApi.backupNow(),
    onSuccess: () => {
      addNotification('Backup wurde erfolgreich erstellt', 'success')
      queryClient.invalidateQueries({ queryKey: ['system', 'backups'] })
    },
    onError: () => {
      addNotification('Fehler beim Erstellen des Backups', 'error')
    },
  })

  const restoreMutation = useMutation({
    mutationFn: (filename: string) => configApi.backupRestore(filename),
    onSuccess: () => {
      setRestoreConfirmFilename(null)
      addNotification('Backup wurde wiederhergestellt. Die Seite wird neu geladen.', 'success')
      setTimeout(() => window.location.reload(), 3000)
    },
    onError: () => {
      addNotification('Fehler beim Wiederherstellen des Backups', 'error')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (filename: string) => configApi.backupDelete(filename),
    onSuccess: () => {
      setDeleteConfirmFilename(null)
      addNotification('Backup wurde geloescht', 'success')
      queryClient.invalidateQueries({ queryKey: ['system', 'backups'] })
    },
    onError: () => {
      addNotification('Fehler beim Loeschen des Backups', 'error')
    },
  })

  const uploadMutation = useMutation({
    mutationFn: (file: File) => configApi.backupUpload(file),
    onSuccess: () => {
      addNotification('Backup wurde erfolgreich importiert', 'success')
      setUploadFile(null)
      if (fileInputRef.current) fileInputRef.current.value = ''
      queryClient.invalidateQueries({ queryKey: ['system', 'backups'] })
    },
    onError: () => {
      addNotification('Fehler beim Importieren des Backups', 'error')
    },
  })

  const handleDownload = async (filename: string) => {
    setDownloadingFile(filename)
    try {
      await configApi.backupDownload(filename)
    } catch {
      addNotification('Fehler beim Herunterladen des Backups', 'error')
    } finally {
      setDownloadingFile(null)
    }
  }

  const formatDate = (isoDate: string) => {
    try {
      return new Date(isoDate).toLocaleString('de-DE', {
        day: '2-digit',
        month: '2-digit',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      })
    } catch {
      return isoDate
    }
  }

  const backupList: BackupEntry[] = Array.isArray(backups) ? backups : []
  const lastBackup = backupList.length > 0 ? backupList[0] : null

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header"><h1>Backups</h1></div>
      <SkeletonCard count={3} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Backups</h1>
        <p>Sichern und wiederherstellen Sie Ihre Daten.</p>
      </div>

      {/* Backup Status Card */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <HardDrive size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Backup-Status
        </h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Letztes Backup</label>
            <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
              {lastBackup ? formatDate(lastBackup.created_at) : 'Noch kein Backup erstellt'}
            </p>
          </div>
          <div className="sp-field">
            <label className="sp-label">Letzte Groesse</label>
            <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
              {lastBackup ? lastBackup.size_human : '---'}
            </p>
          </div>
          <div className="sp-field">
            <label className="sp-label">Gesamt-Backups</label>
            <p style={{ color: 'var(--color-text-primary)', margin: 0, fontSize: 'var(--font-size-sm)' }}>
              {backupList.length}
            </p>
          </div>
        </div>
      </div>

      {/* Manual Backup */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <Shield size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Manuelles Backup
        </h3>
        <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-4) 0' }}>
          Erstellen Sie ein manuelles Backup aller 18 Service-Datenbanken. Dies kann einige Minuten dauern.
        </p>
        <button
          className="sp-btn sp-btn--primary"
          onClick={() => backupNowMutation.mutate()}
          disabled={backupNowMutation.isPending}
        >
          {backupNowMutation.isPending ? (
            <>
              <RefreshCw size={16} style={{ marginRight: 6, verticalAlign: 'middle', animation: 'spin 1s linear infinite' }} />
              Backup laeuft...
            </>
          ) : (
            <>
              <Download size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
              Jetzt sichern
            </>
          )}
        </button>
      </div>

      {/* Backup History Table */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <Clock size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Backup-Verlauf
        </h3>
        {backupList.length === 0 ? (
          <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: 0 }}>
            Noch keine Backups vorhanden.
          </p>
        ) : (
          <div style={{ overflowX: 'auto' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 'var(--font-size-sm)' }}>
              <thead>
                <tr style={{ borderBottom: '1px solid var(--color-border)' }}>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Datum
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Dateiname
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Groesse
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Typ
                  </th>
                  <th style={{ textAlign: 'left', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Status
                  </th>
                  <th style={{ textAlign: 'right', padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontWeight: 500 }}>
                    Aktion
                  </th>
                </tr>
              </thead>
              <tbody>
                {backupList.map((entry) => (
                  <tr key={entry.filename} style={{ borderBottom: '1px solid var(--color-border)' }}>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)' }}>
                      {formatDate(entry.created_at)}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-secondary)', fontFamily: 'monospace', fontSize: 'var(--font-size-xs, 12px)' }}>
                      {entry.filename}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)' }}>
                      {entry.size_human}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', color: 'var(--color-text-primary)' }}>
                      {entry.type}
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)' }}>
                      <span
                        style={{
                          display: 'inline-block',
                          padding: '2px 8px',
                          borderRadius: 'var(--radius-sm, 4px)',
                          fontSize: 'var(--font-size-xs, 12px)',
                          fontWeight: 500,
                          backgroundColor:
                            entry.status === 'Erfolgreich'
                              ? 'var(--color-success-bg, #dcfce7)'
                              : entry.status.startsWith('Teilweise')
                              ? 'var(--color-warning-bg, #fefce8)'
                              : 'var(--color-error-bg, #fef2f2)',
                          color:
                            entry.status === 'Erfolgreich'
                              ? 'var(--color-success, #16a34a)'
                              : entry.status.startsWith('Teilweise')
                              ? 'var(--color-warning, #ca8a04)'
                              : 'var(--color-error, #dc2626)',
                        }}
                      >
                        {entry.status}
                      </span>
                    </td>
                    <td style={{ padding: 'var(--spacing-2) var(--spacing-3)', textAlign: 'right' }}>
                      <div style={{ display: 'flex', gap: 'var(--spacing-2)', justifyContent: 'flex-end' }}>
                        <button
                          className="sp-btn sp-btn--ghost"
                          style={{ fontSize: 'var(--font-size-xs, 12px)', padding: '2px 8px' }}
                          onClick={() => handleDownload(entry.filename)}
                          disabled={downloadingFile === entry.filename}
                        >
                          {downloadingFile === entry.filename ? (
                            <RefreshCw size={14} style={{ marginRight: 4, verticalAlign: 'middle', animation: 'spin 1s linear infinite' }} />
                          ) : (
                            <Download size={14} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                          )}
                          Download
                        </button>
                        <button
                          className="sp-btn sp-btn--secondary"
                          style={{ fontSize: 'var(--font-size-xs, 12px)', padding: '2px 8px' }}
                          onClick={() => {
                            setRestoreConfirmFilename(entry.filename)
                            setRestoreConfirmDate(formatDate(entry.created_at))
                          }}
                        >
                          <RotateCcw size={14} style={{ marginRight: 4, verticalAlign: 'middle' }} />
                          Wiederherstellen
                        </button>
                        <button
                          className="sp-btn sp-btn--ghost"
                          style={{ fontSize: 'var(--font-size-xs, 12px)', padding: '2px 8px', color: 'var(--color-danger, #dc2626)' }}
                          onClick={() => setDeleteConfirmFilename(entry.filename)}
                        >
                          <Trash2 size={14} style={{ verticalAlign: 'middle' }} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Upload Backup */}
      <div className="sp-card">
        <h3 className="sp-card__title">
          <Upload size={18} style={{ marginRight: 8, verticalAlign: 'text-bottom' }} />
          Backup importieren
        </h3>
        <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-4) 0' }}>
          Laden Sie eine vorhandene Backup-Datei (.tar.gz) hoch, um sie zu importieren.
        </p>
        <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <input
            ref={fileInputRef}
            type="file"
            accept=".tar.gz,.tgz"
            onChange={(e) => setUploadFile(e.target.files?.[0] || null)}
            style={{
              fontSize: 'var(--font-size-sm)',
              color: 'var(--color-text-primary)',
            }}
          />
          <button
            className="sp-btn sp-btn--primary"
            onClick={() => uploadFile && uploadMutation.mutate(uploadFile)}
            disabled={!uploadFile || uploadMutation.isPending}
          >
            {uploadMutation.isPending ? (
              <>
                <RefreshCw size={16} style={{ marginRight: 6, verticalAlign: 'middle', animation: 'spin 1s linear infinite' }} />
                Wird importiert...
              </>
            ) : (
              <>
                <Upload size={16} style={{ marginRight: 6, verticalAlign: 'middle' }} />
                Backup importieren
              </>
            )}
          </button>
        </div>
      </div>

      {/* Restore Confirmation Dialog */}
      {restoreConfirmFilename && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050,
        }} onClick={() => !restoreMutation.isPending && setRestoreConfirmFilename(null)}>
          <div
            className="sp-card"
            style={{ width: 480, maxWidth: '90vw', margin: 0 }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)', marginBottom: 'var(--spacing-4)' }}>
              <AlertTriangle size={24} style={{ color: 'var(--color-danger)', flexShrink: 0 }} />
              <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>
                Backup wiederherstellen
              </h2>
            </div>
            <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-5) 0', lineHeight: 1.6 }}>
              Achtung! Alle aktuellen Daten werden durch das Backup vom <strong>{restoreConfirmDate}</strong> ersetzt.
              Dieser Vorgang kann nicht rueckgaengig gemacht werden. Fortfahren?
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--spacing-3)' }}>
              <button
                className="sp-btn sp-btn--secondary"
                onClick={() => setRestoreConfirmFilename(null)}
                disabled={restoreMutation.isPending}
              >
                Abbrechen
              </button>
              <button
                className="sp-btn sp-btn--danger"
                onClick={() => restoreMutation.mutate(restoreConfirmFilename)}
                disabled={restoreMutation.isPending}
              >
                {restoreMutation.isPending ? (
                  <>
                    <RefreshCw size={16} style={{ marginRight: 6, verticalAlign: 'middle', animation: 'spin 1s linear infinite' }} />
                    Wiederherstellen...
                  </>
                ) : (
                  'Wiederherstellen'
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Delete Confirmation Dialog */}
      {deleteConfirmFilename && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)', display: 'flex',
          alignItems: 'center', justifyContent: 'center', zIndex: 1050,
        }} onClick={() => !deleteMutation.isPending && setDeleteConfirmFilename(null)}>
          <div
            className="sp-card"
            style={{ width: 420, maxWidth: '90vw', margin: 0 }}
            onClick={(e) => e.stopPropagation()}
          >
            <div style={{ display: 'flex', alignItems: 'center', gap: 'var(--spacing-3)', marginBottom: 'var(--spacing-4)' }}>
              <Trash2 size={24} style={{ color: 'var(--color-danger)', flexShrink: 0 }} />
              <h2 style={{ margin: 0, fontSize: 'var(--font-size-lg)', color: 'var(--color-text-primary)' }}>
                Backup loeschen
              </h2>
            </div>
            <p style={{ color: 'var(--color-text-secondary)', fontSize: 'var(--font-size-sm)', margin: '0 0 var(--spacing-5) 0', lineHeight: 1.6 }}>
              Soll das Backup <strong>{deleteConfirmFilename}</strong> endgueltig geloescht werden?
            </p>
            <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 'var(--spacing-3)' }}>
              <button
                className="sp-btn sp-btn--secondary"
                onClick={() => setDeleteConfirmFilename(null)}
                disabled={deleteMutation.isPending}
              >
                Abbrechen
              </button>
              <button
                className="sp-btn sp-btn--danger"
                onClick={() => deleteMutation.mutate(deleteConfirmFilename)}
                disabled={deleteMutation.isPending}
              >
                {deleteMutation.isPending ? 'Loeschen...' : 'Loeschen'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default BackupsPage
