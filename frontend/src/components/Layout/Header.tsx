import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../stores/authStore'
import './Header.scss'

function Header() {
  const navigate = useNavigate()
  const user = useAuthStore((state) => state.user)
  const logout = useAuthStore((state) => state.logout)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <header className="header">
      <div className="header__left">
        <h1 className="header__title">RentFlow</h1>
      </div>

      <div className="header__right">
        <div className="header__user">
          <span className="header__user-name">{user?.email || 'User'}</span>
          <button className="header__logout-btn" onClick={handleLogout}>
            Logout
          </button>
        </div>
      </div>
    </header>
  )
}

export default Header
