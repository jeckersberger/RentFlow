import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './i18n/config'
import './styles/global.scss'

// Clear stale Service Worker caches on load
if ('serviceWorker' in navigator) {
  caches.keys().then(names => {
    names.forEach(name => {
      if (name.includes('api-cache') || name.includes('api-lists-cache')) {
        caches.delete(name)
      }
    })
  })
}

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
)
