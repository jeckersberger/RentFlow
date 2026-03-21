import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { invoiceApi } from '../../services/api'
import { useNotificationStore } from '../../stores/notificationStore'
import { StatusBadge } from '../../components/StatusBadge/StatusBadge'
import { Modal } from '../../components/Modal/Modal'
import { Select } from '../../components/Form/Select'
import { InvoiceStatus } from '../../types/invoice'
import '../Equipment/Equipment.module.scss'

function InvoiceDetailPage() {
  const { id } = useParams<{ id: string }>()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [showActionModal, setShowActionModal] = useState(false)
  const [selectedAction, setSelectedAction] = useState<string>('')

  const {
    data: invoice,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['invoice', id],
    queryFn: () => invoiceApi.getById(id!),
    enabled: !!id,
  })

  const getActionLabel = (action: string): string => {
    if (action === 'send') return 'Versendet'
    if (action === 'mark-paid') return 'Bezahlt'
    if (action === 'cancel') return 'Storniert'
    return ''
  }

  const { mutate: performAction, isPending } = useMutation({
    mutationFn: async (action: string) => {
      let newStatus: InvoiceStatus = invoice!.status
      if (action === 'send') newStatus = 'sent'
      if (action === 'mark-paid') newStatus = 'paid'
      if (action === 'cancel') newStatus = 'cancelled'

      return invoiceApi.update(id!, { status: newStatus })
    },
    onMutate: async (action: string) => {
      await queryClient.cancelQueries({ queryKey: ['invoice', id] })
      const previous = queryClient.getQueryData(['invoice', id])

      let newStatus: InvoiceStatus = invoice!.status
      if (action === 'send') newStatus = 'sent'
      if (action === 'mark-paid') newStatus = 'paid'
      if (action === 'cancel') newStatus = 'cancelled'

      queryClient.setQueryData(['invoice', id], (old: any) => ({
        ...old,
        status: newStatus,
        paid_date: newStatus === 'paid' ? new Date().toISOString() : old.paid_date,
      }))
      return { previous }
    },
    onSuccess: (_, action) => {
      setShowActionModal(false)
      const actionLabel = getActionLabel(action)
      addNotification(`Rechnung erfolgreich ${actionLabel.toLowerCase()}`, 'success', {
        title: 'Erfolg',
        duration: 3000,
      })
    },
    onError: (_err: any, _vars, context: any) => {
      queryClient.setQueryData(['invoice', id], context.previous)
      addNotification(
        'Fehler beim Aktualisieren der Rechnung. Bitte versuchen Sie es später erneut.',
        'error',
        {
          title: 'Fehler',
          duration: 5000,
        }
      )
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['invoice', id] })
    },
  })

  if (isLoading) return <div className="invoice-detail-page">Wird geladen...</div>
  if (error) return <div className="error-message">Fehler beim Laden der Rechnung</div>
  if (!invoice) return <div className="error-message">Rechnung nicht gefunden</div>

  const isOverdue = new Date(invoice.due_date) < new Date() && invoice.status !== 'paid'

  return (
    <div className="invoice-detail-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Rechnung {invoice.number}</h1>
          <p className="page-subtitle">
            <StatusBadge
              status={invoice.status}
              label={isOverdue ? 'Überfällig' : undefined}
            />
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
          <button className="btn btn--secondary">
            💾 Als PDF speichern
          </button>
          <button
            className="btn btn--primary"
            onClick={() => setShowActionModal(true)}
          >
            Aktion
          </button>
        </div>
      </div>

      <div className="detail-grid">
        <div className="detail-card">
          <h2 className="detail-card__title">Rechnungsdetails</h2>
          <div className="detail-card__content">
            <div className="detail-card__row">
              <span className="detail-card__row-label">Rechnungsnummer</span>
              <span className="detail-card__row-value">{invoice.number}</span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Kunde</span>
              <span className="detail-card__row-value">
                {invoice.client_name || '—'}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Ausstellungsdatum</span>
              <span className="detail-card__row-value">
                {new Date(invoice.issue_date).toLocaleDateString('de-DE')}
              </span>
            </div>

            <div className="detail-card__row">
              <span className="detail-card__row-label">Fälligkeitsdatum</span>
              <span
                className="detail-card__row-value"
                style={{ color: isOverdue ? 'var(--color-danger)' : 'inherit' }}
              >
                {new Date(invoice.due_date).toLocaleDateString('de-DE')}
              </span>
            </div>

            {invoice.paid_date && (
              <div className="detail-card__row">
                <span className="detail-card__row-label">Zahlungsdatum</span>
                <span className="detail-card__row-value">
                  {new Date(invoice.paid_date).toLocaleDateString('de-DE')}
                </span>
              </div>
            )}
          </div>
        </div>

        <div className="detail-card">
          <h2 className="detail-card__title">Zusammenfassung</h2>
          <div className="detail-card__content">
            <div className="status-box">
              <p className="status-box__label">Zwischensumme</p>
              <p className="status-box__value">
                €{invoice.subtotal.toFixed(2)}
              </p>
            </div>

            {invoice.tax_total ? (
              <div className="status-box">
                <p className="status-box__label">Steuern</p>
                <p className="status-box__value">
                  €{invoice.tax_total.toFixed(2)}
                </p>
              </div>
            ) : null}

            <div className="status-box" style={{ backgroundColor: 'var(--color-primary)' }}>
              <p className="status-box__label" style={{ color: 'white' }}>
                Gesamtbetrag
              </p>
              <p className="status-box__value" style={{ color: 'white' }}>
                €{invoice.total.toFixed(2)}
              </p>
            </div>
          </div>
        </div>
      </div>

      {invoice.line_items && invoice.line_items.length > 0 && (
        <div className="detail-card">
          <h2 className="detail-card__title">Positionen</h2>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr style={{ borderBottom: 'var(--card-border-width) solid var(--color-border)' }}>
                <th style={{ textAlign: 'left', padding: 'var(--spacing-3)' }}>
                  Beschreibung
                </th>
                <th style={{ textAlign: 'right', padding: 'var(--spacing-3)' }}>
                  Menge
                </th>
                <th style={{ textAlign: 'right', padding: 'var(--spacing-3)' }}>
                  Einzelpreis
                </th>
                <th style={{ textAlign: 'right', padding: 'var(--spacing-3)' }}>
                  Gesamt
                </th>
              </tr>
            </thead>
            <tbody>
              {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
              {invoice.line_items.map((item: any, idx: number) => (
                <tr
                  key={idx}
                  style={{ borderBottom: 'var(--card-border-width) solid var(--color-border)' }}
                >
                  <td style={{ padding: 'var(--spacing-3)' }}>
                    {item.description}
                  </td>
                  <td style={{ textAlign: 'right', padding: 'var(--spacing-3)' }}>
                    {item.quantity}
                  </td>
                  <td style={{ textAlign: 'right', padding: 'var(--spacing-3)' }}>
                    €{item.unit_price.toFixed(2)}
                  </td>
                  <td style={{ textAlign: 'right', padding: 'var(--spacing-3)' }}>
                    €{item.total.toFixed(2)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <Modal
        isOpen={showActionModal}
        onClose={() => setShowActionModal(false)}
        title="Rechnung bearbeiten"
        size="sm"
        footer={
          <div style={{ display: 'flex', gap: 'var(--spacing-3)' }}>
            <button
              className="btn btn--secondary"
              onClick={() => setShowActionModal(false)}
            >
              Abbrechen
            </button>
            <button
              className="btn btn--primary"
              onClick={() => {
                if (selectedAction) {
                  performAction(selectedAction)
                }
              }}
              disabled={!selectedAction || isPending}
            >
              {isPending ? 'Wird verarbeitet...' : 'Speichern'}
            </button>
          </div>
        }
      >
        <Select
          label="Aktion"
          options={[
            { value: 'send', label: 'Versenden' },
            { value: 'mark-paid', label: 'Als bezahlt markieren' },
            { value: 'cancel', label: 'Stornieren' },
          ]}
          value={selectedAction}
          onChange={(e) => setSelectedAction(e.target.value)}
          placeholder="Aktion auswählen"
        />
      </Modal>
    </div>
  )
}

export default InvoiceDetailPage
