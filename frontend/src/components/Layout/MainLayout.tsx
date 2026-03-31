import { Outlet } from 'react-router-dom'
import Sidebar from './Sidebar'
import Header from './Header'
import UpdateBanner from '../UpdateBanner/UpdateBanner'
import './MainLayout.scss'

function MainLayout() {
  return (
    <div className="main-layout">
      <Sidebar />
      <div className="main-layout__content">
        <Header />
        <UpdateBanner />
        <main className="main-layout__main">
          <Outlet />
        </main>
      </div>
    </div>
  )
}

export default MainLayout
