import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tenantApi, userApi, categoryApi } from '../../services/api'
import { useAuthStore } from '../../stores/authStore'
import { Input } from '../../components/Form/Input'
import { Category } from '../../types/equipment'
import '../Equipment/Equipment.scss'
import './Settings.scss'

type SettingsTab = 'company' | 'users' | 'categories' | 'notifications'

const SETTINGS_TABS: Array<{ value: SettingsTab; label: string; icon: string }> = [
  { value: 'company', label: 'Unternehmenseinstellungen', icon: '🏢' },
  { value: 'users', label: 'Benutzerverwaltung', icon: '👥' },
  { value: 'categories', label: 'Kategorien', icon: '📂' },
  { value: 'notifications', label: 'Benachrichtigungen', icon: '🔔' },
]

interface TenantData {
  id: string
  name: string
  slug: string
  address_street: string
  address_city: string
  address_zip: string
  address_country: string
  currency: string
  tax_rate: number
  invoice_prefix: string
  default_language: string
}

interface UserData {
  id: string
  name: string
  email: string
  role: string
  status: string
}

function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('company')
  const tenantId = useAuthStore((s) => s.tenantId)
  const queryClient = useQueryClient()

  // --- Company tab state ---
  const [companyName, setCompanyName] = useState('')
  const [addressStreet, setAddressStreet] = useState('')
  const [addressCity, setAddressCity] = useState('')
  const [addressZip, setAddressZip] = useState('')
  const [addressCountry, setAddressCountry] = useState('')
  const [currency, setCurrency] = useState('')
  const [taxRate, setTaxRate] = useState('')
  const [invoicePrefix, setInvoicePrefix] = useState('')
  const [saveSuccess, setSaveSuccess] = useState(false)

  // --- Category tab state ---
  const [newCategoryName, setNewCategoryName] = useState('')
  const [newCategoryIcon, setNewCategoryIcon] = useState('')
  const [newCategoryColor, setNewCategoryColor] = useState('#6366f1')
  const [showNewCategoryForm, setShowNewCategoryForm] = useState(false)

  // ============================================================================
  // COMPANY TAB - Tenant data
  // ============================================================================
  const {
    data: tenantData,
    isLoading: tenantLoading,
    error: tenantError,
  } = useQuery<TenantData>({
    queryKey: ['tenant', tenantId],
    queryFn: () => tenantApi.getById(tenantId!),
    enabled: !!tenantId,
  })

  useEffect(() => {
    if (tenantData) {
      setCompanyName(tenantData.name || '')
      setAddressStreet(tenantData.address_street || '')
      setAddressCity(tenantData.address_city || '')
      setAddressZip(tenantData.address_zip || '')
      setAddressCountry(tenantData.address_country || '')
      setCurrency(tenantData.currency || 'EUR')
      setTaxRate(String(tenantData.tax_rate ?? ''))
      setInvoicePrefix(tenantData.invoice_prefix || '')
    }
  }, [tenantData])

  const updateTenantMutation = useMutation({
    mutationFn: (data: Partial<TenantData>) => tenantApi.update(tenantId!, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenant', tenantId] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    },
  })

  const handleSaveCompany = () => {
    updateTenantMutation.mutate({
      name: companyName,
      address_street: addressStreet,
      address_city: addressCity,
      address_zip: addressZip,
      address_country: addressCountry,
      currency,
      tax_rate: parseFloat(taxRate) || 0,
      invoice_prefix: invoicePrefix,
    })
  }

  // ============================================================================
  // USERS TAB
  // ============================================================================
  const {
    data: usersData,
    isLoading: usersLoading,
    error: usersError,
  } = useQuery<UserData[]>({
    queryKey: ['users'],
    queryFn: () => userApi.list(),
    enabled: activeTab === 'users',
  })

  // ============================================================================
  // CATEGORIES TAB
  // ============================================================================
  const {
    data: categoriesData,
    isLoading: categoriesLoading,
    error: categoriesError,
  } = useQuery<Category[]>({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list(),
    enabled: activeTab === 'categories',
    staleTime: 1000 * 60 * 10,
  })

  const createCategoryMutation = useMutation({
    mutationFn: (data: { name: string; icon: string; color: string }) =>
      categoryApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] })
      setNewCategoryName('')
      setNewCategoryIcon('')
      setNewCategoryColor('#6366f1')
      setShowNewCategoryForm(false)
    },
  })

  const handleCreateCategory = () => {
    if (!newCategoryName.trim()) return
    createCategoryMutation.mutate({
      name: newCategoryName.trim(),
      icon: newCategoryIcon.trim() || '📦',
      color: newCategoryColor,
    })
  }

  // ============================================================================
  // Helper: role badge
  // ============================================================================
  const roleBadgeClass = (role: string) => {
    switch (role?.toLowerCase()) {
      case 'admin':
      case 'owner':
        return 'badge badge--primary'
      default:
        return 'badge badge--secondary'
    }
  }

  const statusBadgeClass = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'badge badge--success'
      case 'invited':
      case 'pending':
        return 'badge badge--warning'
      default:
        return 'badge badge--secondary'
    }
  }

  const statusLabel = (status: string) => {
    switch (status?.toLowerCase()) {
      case 'active':
        return 'Aktiv'
      case 'invited':
      case 'pending':
        return 'Einladung ausstehend'
      case 'inactive':
      case 'disabled':
        return 'Deaktiviert'
      default:
        return status
    }
  }

  // ============================================================================
  // RENDER
  // ============================================================================
  const renderContent = () => {
    switch (activeTab) {
      case 'company':
        return (
          <div className="settings-content">
            <h2 className="form-section__title">Unternehmenseinstellungen</h2>

            {tenantLoading && <p>Daten werden geladen...</p>}
            {tenantError && (
              <p style={{ color: 'var(--color-error)' }}>
                Fehler beim Laden der Unternehmensdaten:{' '}
                {(tenantError as Error).message}
              </p>
            )}

            {!tenantLoading && !tenantError && (
              <>
                <div className="form-section__grid">
                  <Input
                    label="Unternehmensname"
                    value={companyName}
                    onChange={(e) => setCompanyName(e.target.value)}
                  />
                  <Input
                    label="Rechnungspräfix"
                    value={invoicePrefix}
                    onChange={(e) => setInvoicePrefix(e.target.value)}
                    placeholder="z.B. RF"
                  />
                </div>

                <div className="form-section__grid">
                  <Input
                    label="Straße"
                    value={addressStreet}
                    onChange={(e) => setAddressStreet(e.target.value)}
                  />
                  <Input
                    label="PLZ"
                    value={addressZip}
                    onChange={(e) => setAddressZip(e.target.value)}
                  />
                  <Input
                    label="Stadt"
                    value={addressCity}
                    onChange={(e) => setAddressCity(e.target.value)}
                  />
                  <Input
                    label="Land"
                    value={addressCountry}
                    onChange={(e) => setAddressCountry(e.target.value)}
                  />
                </div>

                <div className="form-section__grid">
                  <Input
                    label="Währung"
                    value={currency}
                    onChange={(e) => setCurrency(e.target.value)}
                    placeholder="EUR"
                  />
                  <Input
                    label="Steuersatz (%)"
                    type="number"
                    value={taxRate}
                    onChange={(e) => setTaxRate(e.target.value)}
                    placeholder="19"
                  />
                </div>

                <div className="form-section__footer">
                  {updateTenantMutation.isError && (
                    <p style={{ color: 'var(--color-error)', marginRight: 'auto' }}>
                      Fehler beim Speichern:{' '}
                      {(updateTenantMutation.error as Error).message}
                    </p>
                  )}
                  {saveSuccess && (
                    <p style={{ color: 'var(--color-success)', marginRight: 'auto' }}>
                      Erfolgreich gespeichert!
                    </p>
                  )}
                  <button
                    className="btn btn--primary"
                    onClick={handleSaveCompany}
                    disabled={updateTenantMutation.isPending}
                  >
                    {updateTenantMutation.isPending ? 'Speichern...' : 'Speichern'}
                  </button>
                </div>
              </>
            )}
          </div>
        )

      case 'users':
        return (
          <div className="settings-content">
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 'var(--spacing-4)',
              }}
            >
              <h2 className="form-section__title" style={{ margin: 0 }}>
                Benutzer
              </h2>
            </div>

            {usersLoading && <p>Benutzer werden geladen...</p>}
            {usersError && (
              <p style={{ color: 'var(--color-error)' }}>
                Fehler beim Laden der Benutzer:{' '}
                {(usersError as Error).message}
              </p>
            )}

            {!usersLoading && !usersError && usersData && (
              <div className="users-table">
                <div className="users-table__header">
                  <div>Name</div>
                  <div>E-Mail</div>
                  <div>Rolle</div>
                  <div>Status</div>
                </div>

                {(Array.isArray(usersData) ? usersData : []).map((user) => (
                  <div key={user.id} className="users-table__row">
                    <div>{user.name}</div>
                    <div>{user.email}</div>
                    <div>
                      <span className={roleBadgeClass(user.role)}>
                        {user.role}
                      </span>
                    </div>
                    <div>
                      <span className={statusBadgeClass(user.status)}>
                        {statusLabel(user.status)}
                      </span>
                    </div>
                  </div>
                ))}

                {(Array.isArray(usersData) ? usersData : []).length === 0 && (
                  <div className="users-table__row">
                    <div style={{ gridColumn: '1 / -1', textAlign: 'center' }}>
                      Keine Benutzer gefunden.
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        )

      case 'categories':
        return (
          <div className="settings-content">
            <div
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: 'var(--spacing-4)',
              }}
            >
              <h2 className="form-section__title" style={{ margin: 0 }}>
                Ausrüstungskategorien
              </h2>
              <button
                className="btn btn--primary"
                onClick={() => setShowNewCategoryForm(!showNewCategoryForm)}
              >
                + Neue Kategorie
              </button>
            </div>

            {showNewCategoryForm && (
              <div
                style={{
                  padding: 'var(--padding-md)',
                  border: '1px solid var(--color-border)',
                  borderRadius: 'var(--radius-md)',
                  marginBottom: 'var(--spacing-4)',
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 'var(--spacing-3)',
                }}
              >
                <div className="form-section__grid">
                  <Input
                    label="Name"
                    value={newCategoryName}
                    onChange={(e) => setNewCategoryName(e.target.value)}
                    placeholder="z.B. Beleuchtung"
                  />
                  <Input
                    label="Icon (Emoji)"
                    value={newCategoryIcon}
                    onChange={(e) => setNewCategoryIcon(e.target.value)}
                    placeholder="z.B. 💡"
                  />
                  <div className="form-group">
                    <label className="form-label">Farbe</label>
                    <input
                      type="color"
                      value={newCategoryColor}
                      onChange={(e) => setNewCategoryColor(e.target.value)}
                      style={{ width: '100%', height: '38px', cursor: 'pointer' }}
                    />
                  </div>
                </div>
                {createCategoryMutation.isError && (
                  <p style={{ color: 'var(--color-error)' }}>
                    Fehler beim Erstellen:{' '}
                    {(createCategoryMutation.error as Error).message}
                  </p>
                )}
                <div style={{ display: 'flex', gap: 'var(--spacing-2)' }}>
                  <button
                    className="btn btn--primary"
                    onClick={handleCreateCategory}
                    disabled={createCategoryMutation.isPending || !newCategoryName.trim()}
                  >
                    {createCategoryMutation.isPending ? 'Erstellen...' : 'Erstellen'}
                  </button>
                  <button
                    className="btn btn--secondary"
                    onClick={() => setShowNewCategoryForm(false)}
                  >
                    Abbrechen
                  </button>
                </div>
              </div>
            )}

            {categoriesLoading && <p>Kategorien werden geladen...</p>}
            {categoriesError && (
              <p style={{ color: 'var(--color-error)' }}>
                Fehler beim Laden der Kategorien:{' '}
                {(categoriesError as Error).message}
              </p>
            )}

            {!categoriesLoading && !categoriesError && (
              <div className="categories-list">
                {(Array.isArray(categoriesData) ? categoriesData : []).map(
                  (category) => (
                    <div key={category.id} className="category-item">
                      <span className="category-item__name">
                        <span style={{ marginRight: 'var(--spacing-2)' }}>
                          {category.icon || '📦'}
                        </span>
                        {category.name}
                        {category.color && (
                          <span
                            style={{
                              display: 'inline-block',
                              width: 12,
                              height: 12,
                              borderRadius: '50%',
                              backgroundColor: category.color,
                              marginLeft: 'var(--spacing-2)',
                              verticalAlign: 'middle',
                            }}
                          />
                        )}
                      </span>
                    </div>
                  )
                )}

                {(Array.isArray(categoriesData) ? categoriesData : []).length ===
                  0 && <p>Keine Kategorien vorhanden.</p>}
              </div>
            )}
          </div>
        )

      case 'notifications':
        return (
          <div className="settings-content">
            <h2 className="form-section__title">Benachrichtigungseinstellungen</h2>

            <div className="notification-settings">
              <label className="notification-toggle">
                <input type="checkbox" defaultChecked />
                <span>E-Mail-Benachrichtigungen für neue Projekte</span>
              </label>

              <label className="notification-toggle">
                <input type="checkbox" defaultChecked />
                <span>
                  Benachrichtigung bei Ausrüstung, die bald gewartet werden muss
                </span>
              </label>

              <label className="notification-toggle">
                <input type="checkbox" defaultChecked />
                <span>Erinnerung an überfällige Rechnungen</span>
              </label>

              <label className="notification-toggle">
                <input type="checkbox" />
                <span>Tägliche Zusammenfassung</span>
              </label>

              <label className="notification-toggle">
                <input type="checkbox" defaultChecked />
                <span>System- und Sicherheitsmitteilungen</span>
              </label>
            </div>

            <div className="form-section__footer">
              <button className="btn btn--primary">Speichern</button>
            </div>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className="settings-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Einstellungen</h1>
          <p className="page-subtitle">
            Verwalten Sie Ihre Kontoeinstellungen und Systemkonfiguration
          </p>
        </div>
      </div>

      <div className="settings-container">
        <div className="settings-nav">
          {SETTINGS_TABS.map((tab) => (
            <button
              key={tab.value}
              className={`settings-nav__item ${
                activeTab === tab.value ? 'settings-nav__item--active' : ''
              }`}
              onClick={() => setActiveTab(tab.value)}
            >
              <span className="settings-nav__icon">{tab.icon}</span>
              <span className="settings-nav__label">{tab.label}</span>
            </button>
          ))}
        </div>

        <div className="settings-main">{renderContent()}</div>
      </div>
    </div>
  )
}

export default SettingsPage
