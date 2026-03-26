import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './i18n/config'
import './styles/global.scss'

// Force clear ALL Service Worker caches and re-register
if ('serviceWorker' in navigator) {
  // Clear ALL caches (not just api-cache)
  caches.keys().then(names => {
    names.forEach(name => caches.delete(name))
  })
  // Force SW update
  navigator.serviceWorker.getRegistrations().then(regs => {
    regs.forEach(reg => reg.update())
  })
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
