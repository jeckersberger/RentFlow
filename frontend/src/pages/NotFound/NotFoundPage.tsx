import { useNavigate } from 'react-router-dom'
import { MapPinOff } from 'lucide-react'
import './NotFoundPage.scss'

function NotFoundPage() {
  const navigate = useNavigate()

  return (
    <div className="container">
      <div className="card">
        <div className="iconWrap">
          <MapPinOff size={36} strokeWidth={1.5} />
        </div>
        <h1 className="code">404</h1>
        <h2 className="title">Seite nicht gefunden</h2>
        <p className="description">
          Die angeforderte Seite existiert nicht oder wurde verschoben.
        </p>
        <button className="btn" onClick={() => navigate('/')}>
          Zum Dashboard
        </button>
      </div>
    </div>
  )
}

export default NotFoundPage
