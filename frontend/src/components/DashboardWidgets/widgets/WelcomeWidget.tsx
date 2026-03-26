import { useNavigate } from 'react-router-dom'

export function WelcomeWidget() {
  const navigate = useNavigate()

  const steps = [
    {
      number: 1,
      title: 'Firmendaten vervollständigen',
      description: 'Ergänzen Sie Ihre Firmendaten, Steuersätze und Rechnungseinstellungen.',
      path: '/settings/company',
    },
    {
      number: 2,
      title: 'Equipment anlegen',
      description: 'Erfassen Sie Ihr erstes Equipment-Teil mit Barcode und Kategorie.',
      path: '/equipment/new',
    },
    {
      number: 3,
      title: 'Erstes Projekt erstellen',
      description: 'Legen Sie ein Projekt an und weisen Sie Equipment zu.',
      path: '/projects/new',
    },
  ]

  return (
    <div className="welcome-widget">
      <div className="welcome-widget__header">
        <h2 className="welcome-widget__title">Willkommen bei EquipFlow!</h2>
        <p className="welcome-widget__description">
          Hier sind Ihre nächsten Schritte, um loszulegen:
        </p>
      </div>
      <div className="welcome-widget__steps">
        {steps.map((step) => (
          <div
            key={step.number}
            className="welcome-widget__step"
            onClick={() => navigate(step.path)}
          >
            <div className="welcome-widget__step-number">{step.number}</div>
            <div className="welcome-widget__step-content">
              <h3 className="welcome-widget__step-title">{step.title}</h3>
              <p className="welcome-widget__step-desc">{step.description}</p>
            </div>
            <div className="welcome-widget__step-arrow">→</div>
          </div>
        ))}
      </div>
    </div>
  )
}
