import { useNavigate } from 'react-router-dom'
import { MapPinOff } from 'lucide-react'
import styles from './NotFoundPage.module.scss'

function NotFoundPage() {
  const navigate = useNavigate()

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <div className={styles.iconWrap}>
          <MapPinOff size={36} strokeWidth={1.5} />
        </div>
        <h1 className={styles.code}>404</h1>
        <h2 className={styles.title}>Seite nicht gefunden</h2>
        <p className={styles.description}>
          Die angeforderte Seite existiert nicht oder wurde verschoben.
        </p>
        <button className={styles.btn} onClick={() => navigate('/')}>
          Zum Dashboard
        </button>
      </div>
    </div>
  )
}

export default NotFoundPage
