import i18n from 'i18next'
import { initReactI18next } from 'react-i18next'

// Existing namespaces
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

// New namespaces
import enSettings from './locales/en/settings.json'
import enScanner from './locales/en/scanner.json'
import enWarehouse from './locales/en/warehouse.json'
import enContacts from './locales/en/contacts.json'
import enCrew from './locales/en/crew.json'
import enTransport from './locales/en/transport.json'
import enFinance from './locales/en/finance.json'
import enWorkshop from './locales/en/workshop.json'
import enReports from './locales/en/reports.json'
import enAi from './locales/en/ai.json'
import enMail from './locales/en/mail.json'

import deSettings from './locales/de/settings.json'
import deScanner from './locales/de/scanner.json'
import deWarehouse from './locales/de/warehouse.json'
import deContacts from './locales/de/contacts.json'
import deCrew from './locales/de/crew.json'
import deTransport from './locales/de/transport.json'
import deFinance from './locales/de/finance.json'
import deWorkshop from './locales/de/workshop.json'
import deReports from './locales/de/reports.json'
import deAi from './locales/de/ai.json'
import deMail from './locales/de/mail.json'

import frCommon from './locales/fr/common.json'
import nlCommon from './locales/nl/common.json'

const resources = {
  en: {
    common: enCommon,
    auth: enAuth,
    equipment: enEquipment,
    projects: enProjects,
    invoices: enInvoices,
    settings: enSettings,
    scanner: enScanner,
    warehouse: enWarehouse,
    contacts: enContacts,
    crew: enCrew,
    transport: enTransport,
    finance: enFinance,
    workshop: enWorkshop,
    reports: enReports,
    ai: enAi,
    mail: enMail,
  },
  de: {
    common: deCommon,
    auth: deAuth,
    equipment: deEquipment,
    projects: deProjects,
    invoices: deInvoices,
    settings: deSettings,
    scanner: deScanner,
    warehouse: deWarehouse,
    contacts: deContacts,
    crew: deCrew,
    transport: deTransport,
    finance: deFinance,
    workshop: deWorkshop,
    reports: deReports,
    ai: deAi,
    mail: deMail,
  },
  fr: {
    common: frCommon,
  },
  nl: {
    common: nlCommon,
  },
}

i18n
  .use(initReactI18next)
  .init({
    resources,
    lng: localStorage.getItem('rentflow_language') || 'de',
    fallbackLng: 'de',
    interpolation: {
      escapeValue: false,
    },
    ns: [
      'common',
      'auth',
      'equipment',
      'projects',
      'invoices',
      'settings',
      'scanner',
      'warehouse',
      'contacts',
      'crew',
      'transport',
      'finance',
      'workshop',
      'reports',
      'ai',
      'mail',
    ],
    defaultNS: 'common',
  })

export default i18n
