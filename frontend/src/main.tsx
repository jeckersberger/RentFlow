import React from 'react';
import ReactDOM from 'react-dom/client';
import { useAuthStore } from './stores/authStore';
import App from './App';
import './i18n';
import './styles/global.scss';

// Restore auth state from localStorage before first render
useAuthStore.getState().init();

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
