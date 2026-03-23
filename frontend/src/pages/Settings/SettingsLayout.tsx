import { Outlet } from 'react-router-dom'
import SettingsNav from './SettingsNav'
import './Settings.scss'

function SettingsLayout() {
  return (
    <div className="settings-layout">
      <SettingsNav />
      <div className="settings-outlet">
        <Outlet />
      </div>
    </div>
  )
}

export default SettingsLayout
