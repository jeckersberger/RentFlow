import { useNavigate } from 'react-router-dom'

interface PlaceholderPageProps {
  title: string
  icon: string
  description: string
}

function PlaceholderPage({ title, icon, description }: PlaceholderPageProps) {
  const navigate = useNavigate()

  return (
    <div style={{ padding: '2rem', maxWidth: '600px', margin: '0 auto' }}>
      <button
        onClick={() => navigate(-1)}
        style={{
          background: 'none',
          border: '1px solid var(--gray-a6)',
          borderRadius: '8px',
          padding: '0.5rem 1rem',
          cursor: 'pointer',
          color: 'var(--gray-11)',
          fontSize: '0.875rem',
          marginBottom: '2rem',
          display: 'flex',
          alignItems: 'center',
          gap: '0.5rem',
        }}
      >
        ← Zurück
      </button>

      <div
        style={{
          textAlign: 'center',
          padding: '3rem 2rem',
          background: 'var(--color-surface)',
          borderRadius: '12px',
          border: '1px solid var(--gray-a4)',
        }}
      >
        <div style={{ fontSize: '3rem', marginBottom: '1rem' }}>{icon}</div>
        <h1 style={{ fontSize: '1.5rem', fontWeight: 600, color: 'var(--gray-12)', marginBottom: '0.5rem' }}>
          {title}
        </h1>
        <p style={{ color: 'var(--gray-10)', marginBottom: '1.5rem', fontSize: '0.95rem' }}>
          {description}
        </p>
        <div
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            padding: '0.5rem 1rem',
            background: 'var(--accent-a3)',
            color: 'var(--accent-11)',
            borderRadius: '8px',
            fontSize: '0.85rem',
            fontWeight: 500,
          }}
        >
          🚧 In Entwicklung
        </div>
      </div>
    </div>
  )
}

// Transport
export function NewTourPage() {
  return (
    <PlaceholderPage
      icon="🚛"
      title="Neue Tour erstellen"
      description="Planen Sie eine neue Transporttour mit Fahrzeugen, Routen und Zeitfenstern."
    />
  )
}

export function NewVehiclePage() {
  return (
    <PlaceholderPage
      icon="🚐"
      title="Neues Fahrzeug anlegen"
      description="Erfassen Sie ein neues Fahrzeug mit Typ, Kennzeichen und technischen Daten."
    />
  )
}

export function VehicleDetailPage() {
  return (
    <PlaceholderPage
      icon="🔧"
      title="Fahrzeug-Details"
      description="Detailansicht mit Fahrzeugdaten, Wartungshistorie und aktuellem Status."
    />
  )
}

// Maintenance
export function NewMaintenanceTaskPage() {
  return (
    <PlaceholderPage
      icon="🔨"
      title="Neue Wartungsaufgabe"
      description="Erstellen Sie eine neue Wartungs- oder Reparaturaufgabe für Ihre Geräte."
    />
  )
}

export function NewECheckPage() {
  return (
    <PlaceholderPage
      icon="⚡"
      title="E-Check durchführen"
      description="Starten Sie eine elektrische Sicherheitsprüfung nach DGUV Vorschrift 3."
    />
  )
}

export function MaintenancePlanDetailPage() {
  return (
    <PlaceholderPage
      icon="📋"
      title="Wartungsplan-Details"
      description="Übersicht über geplante Wartungsintervalle und durchgeführte Prüfungen."
    />
  )
}

// Crew
export function NewCrewMemberPage() {
  return (
    <PlaceholderPage
      icon="👤"
      title="Neuer Mitarbeiter"
      description="Legen Sie einen neuen Mitarbeiter mit Kontaktdaten und Qualifikationen an."
    />
  )
}

// Documents
export function DocumentDetailPage() {
  return (
    <PlaceholderPage
      icon="📄"
      title="Dokument-Details"
      description="Detailansicht mit Dokumentinformationen, Versionen und Freigaben."
    />
  )
}

export function NewDocumentPage() {
  return (
    <PlaceholderPage
      icon="📝"
      title="Neues Dokument"
      description="Laden Sie ein neues Dokument hoch oder erstellen Sie eine Vorlage."
    />
  )
}

// Insurance
export function NewClaimPage() {
  return (
    <PlaceholderPage
      icon="🛡️"
      title="Neuen Schadensfall melden"
      description="Erfassen Sie einen neuen Versicherungsfall mit Schadensbeschreibung und Fotos."
    />
  )
}
