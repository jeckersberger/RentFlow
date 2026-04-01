import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Users } from 'lucide-react'
import { contactApi } from '../../services/api'
import { Input } from '../../components/Form/Input'
import EmptyState from '../../components/EmptyState/EmptyState'
import ErrorState from '../../components/ErrorState/ErrorState'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import { useNotificationStore } from '../../stores/notificationStore'
import ContactForm, { ContactFormData } from './ContactForm'
import ContactImport from './ContactImport'
import { generateCSV, downloadCSV, formatDateForExport } from '../../utils/csvExport'
import './Contacts.scss'

interface Contact {
  id: string
  type: 'company' | 'person'
  company_name: string
  first_name: string
  last_name: string
  email: string
  phone: string
  mobile: string
  website: string
  street: string
  house_number: string
  zip: string
  city: string
  country: string
  vat_id: string
  notes: string
  tags: string[]
  created_at: string
  updated_at: string
}

// Mock contacts for MOCK_MODE / empty backend
const mockContacts: Contact[] = [
  { id: '1', type: 'company', company_name: 'Stadt München', first_name: '', last_name: '', email: 'veranstaltungen@muenchen.de', phone: '+49 89 233-0', mobile: '', website: 'https://muenchen.de', street: 'Marienplatz', house_number: '8', zip: '80331', city: 'München', country: 'Deutschland', vat_id: 'DE129524797', notes: 'Hauptkunde - Stadtfeste und Events', tags: ['kunde', 'behoerde'], created_at: '2025-06-15T10:00:00Z', updated_at: '2026-03-20T14:00:00Z' },
  { id: '2', type: 'company', company_name: 'TechCorp GmbH', first_name: 'Anna', last_name: 'Schmidt', email: 'events@techcorp.de', phone: '+49 69 12345-0', mobile: '+49 170 9876543', website: 'https://techcorp.de', street: 'Kaiserstraße', house_number: '42', zip: '60329', city: 'Frankfurt', country: 'Deutschland', vat_id: 'DE298765432', notes: 'Firmenevents, Gala-Veranstaltungen', tags: ['kunde'], created_at: '2025-08-20T14:00:00Z', updated_at: '2026-03-18T09:00:00Z' },
  { id: '3', type: 'company', company_name: 'Festival GmbH', first_name: 'Markus', last_name: 'Weber', email: 'info@festivalgmbh.de', phone: '+49 7531 123456', mobile: '+49 160 1234567', website: 'https://festivalgmbh.de', street: 'Seestraße', house_number: '15', zip: '78462', city: 'Konstanz', country: 'Deutschland', vat_id: 'DE345678901', notes: 'Open Air Festival Veranstalter', tags: ['kunde', 'festival'], created_at: '2025-09-10T11:00:00Z', updated_at: '2026-02-20T11:00:00Z' },
  { id: '4', type: 'company', company_name: 'AutoBrand AG', first_name: 'Thomas', last_name: 'Müller', email: 'events@autobrand.de', phone: '+49 711 98765-0', mobile: '', website: 'https://autobrand.de', street: 'Mercedesstraße', house_number: '1', zip: '70327', city: 'Stuttgart', country: 'Deutschland', vat_id: 'DE112233445', notes: 'Automobilhersteller - Produktlaunches', tags: ['kunde', 'automotive'], created_at: '2025-10-01T08:00:00Z', updated_at: '2026-02-28T10:00:00Z' },
  { id: '5', type: 'person', company_name: '', first_name: 'Julia', last_name: 'Weber', email: 'julia.weber@gmail.com', phone: '+49 89 555-1234', mobile: '+49 176 12345678', website: '', street: 'Schlossweg', house_number: '7', zip: '87645', city: 'Schwangau', country: 'Deutschland', vat_id: '', notes: 'Hochzeit Familie Weber', tags: ['privatkunde'], created_at: '2026-01-20T16:00:00Z', updated_at: '2026-03-08T10:00:00Z' },
  { id: '6', type: 'company', company_name: 'Sound & Light Verleih', first_name: 'Peter', last_name: 'Braun', email: 'info@soundlight.de', phone: '+49 89 777-8888', mobile: '', website: 'https://soundlight.de', street: 'Industriestraße', house_number: '22', zip: '85748', city: 'Garching', country: 'Deutschland', vat_id: 'DE556677889', notes: 'Kooperationspartner für Sub-Vermietung', tags: ['lieferant', 'partner'], created_at: '2025-07-01T09:00:00Z', updated_at: '2026-03-15T14:00:00Z' },
]

const projectCounts: Record<string, number> = {
  '1': 3, '2': 1, '3': 1, '4': 1, '5': 1, '6': 0,
}

function getDisplayName(contact: Contact): string {
  if (contact.type === 'company') return contact.company_name
  return `${contact.first_name} ${contact.last_name}`.trim()
}

function getInitials(contact: Contact): string {
  if (contact.type === 'company') {
    return contact.company_name.split(' ').map(w => w[0]).slice(0, 2).join('').toUpperCase()
  }
  return `${contact.first_name?.[0] || ''}${contact.last_name?.[0] || ''}`.toUpperCase()
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: 'numeric' })
}

function ContactsPage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { addNotification } = useNotificationStore()
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState<'' | 'company' | 'person'>('')
  const [tagFilter, setTagFilter] = useState('')
  const [selectedContact, setSelectedContact] = useState<Contact | null>(null)
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [showImport, setShowImport] = useState(false)
  const [editingContact, setEditingContact] = useState<Contact | null>(null)

  const { data: contactsRaw, isLoading: _isLoading, error } = useQuery({
    queryKey: ['contacts', searchQuery],
    queryFn: () => contactApi.list({ search: searchQuery }),
    staleTime: 1000 * 60 * 5,
  })

  // Use real data or fallback to mock
  const contacts: Contact[] = useMemo(() => {
    const raw = contactsRaw?.data || contactsRaw
    if (Array.isArray(raw) && raw.length > 0) return raw
    // Fallback to mock data when backend returns empty or errors
    return mockContacts
  }, [contactsRaw])

  const filteredContacts = useMemo(() => {
    return contacts.filter((c) => {
      const name = getDisplayName(c).toLowerCase()
      const matchesSearch = !searchQuery ||
        name.includes(searchQuery.toLowerCase()) ||
        c.email.toLowerCase().includes(searchQuery.toLowerCase()) ||
        c.city.toLowerCase().includes(searchQuery.toLowerCase())
      const matchesType = !typeFilter || c.type === typeFilter
      const matchesTag = !tagFilter || (c.tags || []).some(t => t.toLowerCase().includes(tagFilter.toLowerCase()))
      return matchesSearch && matchesType && matchesTag
    })
  }, [contacts, searchQuery, typeFilter, tagFilter])

  const { mutate: createContact, isPending: isCreating } = useMutation({
    mutationFn: (data: ContactFormData) => {
      const payload = {
        type: data.type,
        company_name: data.company_name,
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email,
        phone: data.phone,
        mobile: data.mobile,
        website: data.website,
        street: data.street,
        house_number: data.house_number,
        zip: data.zip,
        city: data.city,
        country: data.country,
        vat_id: data.vat_id,
        notes: data.notes,
        tags: data.tags,
      }
      return contactApi.create(payload)
    },
    onSuccess: () => {
      setShowCreateForm(false)
      queryClient.invalidateQueries({ queryKey: ['contacts'] })
      addNotification('Kontakt erfolgreich erstellt', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Erstellen des Kontakts', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const { mutate: updateContact, isPending: isUpdating } = useMutation({
    mutationFn: ({ id, data }: { id: string; data: ContactFormData }) => {
      const payload = {
        type: data.type,
        company_name: data.company_name,
        first_name: data.first_name,
        last_name: data.last_name,
        email: data.email,
        phone: data.phone,
        mobile: data.mobile,
        website: data.website,
        street: data.street,
        house_number: data.house_number,
        zip: data.zip,
        city: data.city,
        country: data.country,
        vat_id: data.vat_id,
        notes: data.notes,
        tags: data.tags,
      }
      return contactApi.update(id, payload)
    },
    onSuccess: () => {
      setEditingContact(null)
      setSelectedContact(null)
      queryClient.invalidateQueries({ queryKey: ['contacts'] })
      addNotification('Kontakt erfolgreich aktualisiert', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Aktualisieren des Kontakts', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const { mutate: deleteContact } = useMutation({
    mutationFn: (id: string) => contactApi.delete(id),
    onSuccess: () => {
      setSelectedContact(null)
      queryClient.invalidateQueries({ queryKey: ['contacts'] })
      addNotification('Kontakt erfolgreich gelöscht', 'success', { title: 'Erfolg', duration: 3000 })
    },
    onError: () => {
      addNotification('Fehler beim Löschen des Kontakts', 'error', { title: 'Fehler', duration: 5000 })
    },
  })

  const exportContactsCSV = () => {
    const CONTACT_HEADERS = [
      { key: 'type', label: 'Typ' },
      { key: 'company_name', label: 'Firmenname' },
      { key: 'first_name', label: 'Vorname' },
      { key: 'last_name', label: 'Nachname' },
      { key: 'email', label: 'E-Mail' },
      { key: 'phone', label: 'Telefon' },
      { key: 'mobile', label: 'Mobil' },
      { key: 'website', label: 'Website' },
      { key: 'street', label: 'Straße' },
      { key: 'house_number', label: 'Hausnummer' },
      { key: 'zip', label: 'PLZ' },
      { key: 'city', label: 'Stadt' },
      { key: 'country', label: 'Land' },
      { key: 'vat_id', label: 'USt-ID' },
      { key: 'notes', label: 'Notizen' },
      { key: 'tags', label: 'Tags' },
    ]
    const rows = filteredContacts.map((c) => ({
      type: c.type === 'company' ? 'Firma' : 'Person',
      company_name: c.company_name || '',
      first_name: c.first_name || '',
      last_name: c.last_name || '',
      email: c.email || '',
      phone: c.phone || '',
      mobile: c.mobile || '',
      website: c.website || '',
      street: c.street || '',
      house_number: c.house_number || '',
      zip: c.zip || '',
      city: c.city || '',
      country: c.country || '',
      vat_id: c.vat_id || '',
      notes: c.notes || '',
      tags: (c.tags || []).join(', '),
    }))
    const csv = generateCSV(CONTACT_HEADERS, rows)
    downloadCSV(csv, `kontakte_export_${formatDateForExport()}.csv`)
    addNotification(`${rows.length} Kontakte exportiert`, 'success', { title: 'Export erfolgreich', duration: 3000 })
  }

  const googleMapsLink = (contact: Contact) => {
    const addr = `${contact.street} ${contact.house_number}, ${contact.zip} ${contact.city}, ${contact.country}`
    return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(addr)}`
  }

  return (
    <div className="contacts-page">
      {/* Header */}
      <div className="page-header">
        <div>
          <h1 className="page-title">Kontakte</h1>
          <p className="page-subtitle">
            {filteredContacts.length} Kontakt{filteredContacts.length !== 1 ? 'e' : ''} gesamt
          </p>
        </div>
        <div style={{ display: 'flex', gap: 'var(--spacing-3)', flexWrap: 'wrap' }}>
          <button
            className="btn btn--secondary"
            onClick={exportContactsCSV}
            disabled={filteredContacts.length === 0}
          >
            Exportieren
          </button>
          <button
            className="btn btn--secondary"
            onClick={() => setShowImport(true)}
          >
            CSV importieren
          </button>
          <button
            className="btn btn--primary"
            onClick={() => setShowCreateForm(true)}
          >
            + Kontakt hinzufügen
          </button>
        </div>
      </div>

      {/* Filters */}
      <div className="filters-bar">
        <Input
          type="text"
          placeholder="Nach Name, E-Mail oder Stadt suchen..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
        />
        <select
          value={typeFilter}
          onChange={(e) => setTypeFilter(e.target.value as '' | 'company' | 'person')}
          style={{
            backgroundColor: 'var(--glass-bg-input-strong)',
            border: '1px solid var(--color-border-strong)',
            color: 'var(--color-text-primary)',
            padding: 'var(--padding-md)',
            borderRadius: 'var(--radius-input)',
            fontSize: 'var(--font-size-sm)',
          }}
        >
          <option value="">Alle Typen</option>
          <option value="company">Firma</option>
          <option value="person">Person</option>
        </select>
        <Input
          type="text"
          placeholder="Tag filtern..."
          value={tagFilter}
          onChange={(e) => setTagFilter(e.target.value)}
        />
      </div>

      {error && !contacts.length && (
        <ErrorState
          variant="generic"
          title="Fehler beim Laden der Kontakte"
          description="Die Kontaktdaten konnten nicht geladen werden. Demo-Daten werden angezeigt."
          onRetry={() => window.location.reload()}
          compact
        />
      )}

      {/* Main Content */}
      <div className="contacts-layout">
        {/* Contacts Table */}
        <div className="contacts-list-container">
          {_isLoading ? (
            <SkeletonTable rows={6} columns={7} />
          ) : filteredContacts.length === 0 ? (
            <EmptyState
              icon={Users}
              title="Keine Kontakte gefunden"
              description={searchQuery || typeFilter
                ? 'Versuchen Sie andere Suchkriterien.'
                : 'Erstellen Sie Ihren ersten Kontakt, um loszulegen.'}
              action={!(searchQuery || typeFilter) ? { label: 'Kontakt hinzufügen', onClick: () => setShowCreateForm(true) } : undefined}
              compact
            />
          ) : (
            <table className="contacts-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Typ</th>
                  <th>E-Mail</th>
                  <th>Telefon</th>
                  <th>Stadt</th>
                  <th>Projekte</th>
                  <th>Letzte Aktivität</th>
                </tr>
              </thead>
              <tbody>
                {filteredContacts.map((contact) => (
                  <tr
                    key={contact.id}
                    className={selectedContact?.id === contact.id ? 'active' : ''}
                    onClick={() => setSelectedContact(contact)}
                  >
                    <td>
                      <div className="contact-name-cell">
                        <div className={`contact-avatar ${`contact-avatar--${contact.type}`}`}>
                          {getInitials(contact)}
                        </div>
                        <span>{getDisplayName(contact)}</span>
                      </div>
                    </td>
                    <td>
                      <span className={`type-badge ${`type-badge--${contact.type}`}`}>
                        {contact.type === 'company' ? 'Firma' : 'Person'}
                      </span>
                    </td>
                    <td>{contact.email || '\u2014'}</td>
                    <td>{contact.phone || '\u2014'}</td>
                    <td>{contact.city || '\u2014'}</td>
                    <td>{projectCounts[contact.id] || 0}</td>
                    <td>{formatDate(contact.updated_at)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        {/* Detail Sidebar */}
        {selectedContact && (
          <div className="contact-detail-sidebar">
            <div className="sidebar-header">
              <div>
                <div className={`sidebar-avatar ${`sidebar-avatar--${selectedContact.type}`}`}>
                  {getInitials(selectedContact)}
                </div>
                <h2 className="sidebar-name">{getDisplayName(selectedContact)}</h2>
                <p className="sidebar-type">
                  {selectedContact.type === 'company' ? 'Firma' : 'Person'}
                  {selectedContact.tags?.length > 0 && (
                    <> &middot; {selectedContact.tags.join(', ')}</>
                  )}
                </p>
              </div>
              <button
                className="sidebar-close"
                onClick={() => setSelectedContact(null)}
                title="Schließen"
              >
                &times;
              </button>
            </div>

            {/* Contact Info */}
            <div className="sidebar-section">
              <h3 className="sidebar-section-title">Kontaktdaten</h3>

              {selectedContact.email && (
                <div className="sidebar-field">
                  <span className="sidebar-field-label">E-Mail</span>
                  <span className="sidebar-field-value">
                    <a href={`mailto:${selectedContact.email}`}>{selectedContact.email}</a>
                  </span>
                </div>
              )}

              {selectedContact.phone && (
                <div className="sidebar-field">
                  <span className="sidebar-field-label">Telefon</span>
                  <span className="sidebar-field-value">
                    <a href={`tel:${selectedContact.phone}`}>{selectedContact.phone}</a>
                  </span>
                </div>
              )}

              {selectedContact.mobile && (
                <div className="sidebar-field">
                  <span className="sidebar-field-label">Mobil</span>
                  <span className="sidebar-field-value">
                    <a href={`tel:${selectedContact.mobile}`}>{selectedContact.mobile}</a>
                  </span>
                </div>
              )}

              {selectedContact.website && (
                <div className="sidebar-field">
                  <span className="sidebar-field-label">Website</span>
                  <span className="sidebar-field-value">
                    <a href={selectedContact.website} target="_blank" rel="noopener noreferrer">
                      {selectedContact.website}
                    </a>
                  </span>
                </div>
              )}
            </div>

            {/* Address */}
            {(selectedContact.street || selectedContact.city) && (
              <div className="sidebar-section">
                <h3 className="sidebar-section-title">Adresse</h3>
                <div className="sidebar-field">
                  <span className="sidebar-field-value">
                    {selectedContact.street && `${selectedContact.street} ${selectedContact.house_number}`}
                    {selectedContact.street && <br />}
                    {selectedContact.zip} {selectedContact.city}
                    {selectedContact.city && <br />}
                    {selectedContact.country}
                  </span>
                </div>
                <a
                  href={googleMapsLink(selectedContact)}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn btn--secondary btn--sm"
                  style={{ marginTop: 'var(--spacing-2)', textDecoration: 'none' }}
                >
                  In Google Maps öffnen
                </a>
              </div>
            )}

            {/* Company-specific */}
            {selectedContact.type === 'company' && selectedContact.vat_id && (
              <div className="sidebar-section">
                <h3 className="sidebar-section-title">Firmendaten</h3>
                <div className="sidebar-field">
                  <span className="sidebar-field-label">USt-ID</span>
                  <span className="sidebar-field-value">{selectedContact.vat_id}</span>
                </div>
              </div>
            )}

            {/* Notes */}
            {selectedContact.notes && (
              <div className="sidebar-section">
                <h3 className="sidebar-section-title">Notizen</h3>
                <p style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-primary)', margin: 0, whiteSpace: 'pre-wrap' }}>
                  {selectedContact.notes}
                </p>
              </div>
            )}

            {/* Associated Projects */}
            <div className="sidebar-section">
              <h3 className="sidebar-section-title">Projekte ({projectCounts[selectedContact.id] || 0})</h3>
              {(projectCounts[selectedContact.id] || 0) > 0 ? (
                <div>
                  <div
                    className="sidebar-project-item"
                    onClick={() => navigate('/projects')}
                  >
                    <span className="sidebar-project-name">Projekte anzeigen</span>
                    <span className="sidebar-project-status">&rarr;</span>
                  </div>
                </div>
              ) : (
                <p style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', margin: 0 }}>
                  Keine Projekte zugeordnet
                </p>
              )}
            </div>

            {/* Communication History */}
            <div className="sidebar-section">
              <h3 className="sidebar-section-title">Kommunikation</h3>
              <p style={{ fontSize: 'var(--font-size-sm)', color: 'var(--color-text-secondary)', margin: 0 }}>
                Erstellt am {formatDate(selectedContact.created_at)}
                <br />
                Zuletzt aktualisiert am {formatDate(selectedContact.updated_at)}
              </p>
            </div>

            {/* Actions */}
            <div className="sidebar-actions">
              <button
                className="btn btn--secondary btn--sm"
                onClick={() => setEditingContact(selectedContact)}
              >
                Bearbeiten
              </button>
              <button
                className="btn btn--danger btn--sm"
                onClick={() => {
                  if (window.confirm(`Kontakt "${getDisplayName(selectedContact)}" wirklich löschen?`)) {
                    deleteContact(selectedContact.id)
                  }
                }}
              >
                Löschen
              </button>
            </div>
          </div>
        )}
      </div>

      {/* Create Contact Modal */}
      <ContactForm
        isOpen={showCreateForm}
        onClose={() => setShowCreateForm(false)}
        onSubmit={(data) => createContact(data)}
        isLoading={isCreating}
        title="Kontakt hinzufügen"
      />

      {/* Edit Contact Modal */}
      {editingContact && (
        <ContactForm
          isOpen={!!editingContact}
          onClose={() => setEditingContact(null)}
          onSubmit={(data) => updateContact({ id: editingContact.id, data })}
          initialData={{
            type: editingContact.type,
            company_name: editingContact.company_name,
            first_name: editingContact.first_name,
            last_name: editingContact.last_name,
            email: editingContact.email,
            phone: editingContact.phone,
            mobile: editingContact.mobile,
            website: editingContact.website,
            street: editingContact.street,
            house_number: editingContact.house_number,
            zip: editingContact.zip,
            city: editingContact.city,
            country: editingContact.country,
            vat_id: editingContact.vat_id,
            notes: editingContact.notes,
            tags: editingContact.tags || [],
          }}
          isLoading={isUpdating}
          title="Kontakt bearbeiten"
        />
      )}

      {/* Import Modal */}
      <ContactImport isOpen={showImport} onClose={() => setShowImport(false)} />
    </div>
  )
}

export default ContactsPage
