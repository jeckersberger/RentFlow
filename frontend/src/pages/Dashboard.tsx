import { WidgetGrid } from '../components/DashboardWidgets'
import './Dashboard.scss'

function DashboardPage() {
  return (
    <div className="dashboard">
      <div className="dashboard__header">
        <h1 className="dashboard__title">Dashboard</h1>
        <p className="dashboard__subtitle">Willkommen zurück! Hier ist ein Überblick über Ihr Geschäft.</p>
      </div>

      <WidgetGrid />
    </div>
  )
}

export default DashboardPage
