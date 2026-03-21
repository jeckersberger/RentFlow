import { useState } from 'react'
import { Input } from '../../components/Form/Input'
import { TextArea } from '../../components/Form/TextArea'
import '../Equipment/Equipment.module.scss'
import './Settings.module.scss'

type SettingsTab = 'company' | 'users' | 'categories' | 'notifications'

const SETTINGS_TABS: Array<{ value: SettingsTab; label: string; icon: string }> = [
  { value: 'company', label: 'Unternehmenseinstellungen', icon: '🏢' },
  { value: 'users', label: 'Benutzerverwaltung', icon: '👥' },
  { value: 'categories', label: 'Kategorien', icon: '📂' },
  { value: 'notifications', label: 'Benachrichtigungen', icon: '🔔' },
]

function SettingsPage() {
  const [activeTab, setActiveTab] = useState<SettingsTab>('company')
  const [companyName, setCompanyName] = useState('Meine Veranstaltungsfirma')
  const [companyAddress, setCompanyAddress] = useState('Musterstraße 123, 12345 Musterstadt')
  const [companyPhone, setCompanyPhone] = useState('+49 (0) 123 456789')
  const [companyEmail, setCompanyEmail] = useState('info@example.com')

  const renderContent = () => {
    switch (activeTab) {
      case 'company':
        return (
          <div className="settings-content">
            <h2 className="form-section__title">Unternehmenseinstellungen</h2>

            <div className="form-section__grid">
              <Input
                label="Unternehmensname"
                value={companyName}
                onChange={(e) => setCompanyName(e.target.value)}
              />
              <Input
                label="E-Mail"
                type="email"
                value={companyEmail}
                onChange={(e) => setCompanyEmail(e.target.value)}
              />
              <Input
                label="Telefon"
                value={companyPhone}
                onChange={(e) => setCompanyPhone(e.target.value)}
              />
            </div>

            <TextArea
              label="Adresse"
              value={companyAddress}
              onChange={(e) => setCompanyAddress(e.target.value)}
              rows={3}
            />

            <div className="form-section__footer">
              <button className="btn btn--primary">
                Speichern
              </button>
            </div>
          </div>
        )

      case 'users':
        return (
          <div className="settings-content">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-4)' }}>
              <h2 className="form-section__title" style={{ margin: 0 }}>
                Benutzer
              </h2>
              <button className="btn btn--primary">
                + Benutzer einladen
              </button>
            </div>

            <div className="users-table">
              <div className="users-table__header">
                <div>Name</div>
                <div>E-Mail</div>
                <div>Rolle</div>
                <div>Status</div>
                <div>Aktion</div>
              </div>

              <div className="users-table__row">
                <div>Admin User</div>
                <div>admin@example.com</div>
                <div>
                  <span className="badge badge--primary">Admin</span>
                </div>
                <div>
                  <span className="badge badge--success">Aktiv</span>
                </div>
                <div>
                  <button className="icon-btn">⋯</button>
                </div>
              </div>

              <div className="users-table__row">
                <div>Team Member 1</div>
                <div>member1@example.com</div>
                <div>
                  <span className="badge badge--secondary">User</span>
                </div>
                <div>
                  <span className="badge badge--success">Aktiv</span>
                </div>
                <div>
                  <button className="icon-btn">⋯</button>
                </div>
              </div>

              <div className="users-table__row">
                <div>Team Member 2</div>
                <div>member2@example.com</div>
                <div>
                  <span className="badge badge--secondary">User</span>
                </div>
                <div>
                  <span className="badge badge--warning">Einladung ausstehend</span>
                </div>
                <div>
                  <button className="icon-btn">⋯</button>
                </div>
              </div>
            </div>
          </div>
        )

      case 'categories':
        return (
          <div className="settings-content">
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 'var(--spacing-4)' }}>
              <h2 className="form-section__title" style={{ margin: 0 }}>
                Ausrüstungskategorien
              </h2>
              <button className="btn btn--primary">
                + Neue Kategorie
              </button>
            </div>

            <div className="categories-list">
              {[
                'Beleuchtung',
                'Ton & Audio',
                'Bühne & Struktur',
                'Projektion',
                'Dekoration',
              ].map((category) => (
                <div key={category} className="category-item">
                  <span className="category-item__name">{category}</span>
                  <div className="category-item__actions">
                    <button className="icon-btn">✏️</button>
                    <button className="icon-btn">🗑️</button>
                  </div>
                </div>
              ))}
            </div>
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
                <span>Benachrichtigung bei Ausrüstung, die bald gewartet werden muss</span>
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
              <button className="btn btn--primary">
                Speichern
              </button>
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

        <div className="settings-main">
          {renderContent()}
        </div>
      </div>
    </div>
  )
}

export default SettingsPage
