import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'
import enCommon from './locales/en/common.json'
import enAuth from './locales/en/auth.json'
import enEquipment from './locales/en/equipment.json'
import enProjects from './locales/en/projects.json'
import enInvoices from './locales/en/invoices.json'
import deCommon from './locales/de/common.json'
import deAuth from './locales/de/auth.json'
import deEquipment from './locales/de/equipment.json'
import deProjects from './locales/de/projects.json'
import deInvoices from './locales/de/invoices.json'

const resources = {
  en: {
    common: enCommon,
    auth: enAuth,
    equipment: enEquipment,
    projects: enProjects,
    invoices: enInvoices,
  },
  de: {
    common: deCommon,
    auth: deAuth,
    equipment: deEquipment,
    projects: deProjects,
    invoices: deInvoices,
  },
}

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: localStorage.getItem('language') || 'de',
    fallbackLng: 'de',
    interpolation: {
      escapeValue: false,
    },
    ns: ['common', 'auth', 'equipment', 'projects', 'invoices'],
    defaultNS: 'common',
  })

export default i18n
