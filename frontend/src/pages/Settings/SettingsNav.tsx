import { NavLink } from 'react-router-dom'
import {
  Building2,
  Users,
  ShieldCheck,
  Puzzle,
  Database,
  Globe,
  Hash,
  FolderTree,
  Tags,
  SlidersHorizontal,
  Bell,
  Mail,
  Mailbox,
  FileText,
  Scale,
  Percent,
  CreditCard,
  Landmark,
  ScrollText,
  Blocks,
  QrCode,
  Smartphone,
  Download,
  type LucideIcon,
} from 'lucide-react'
import './Settings.scss'

interface SettingsNavItem {
  label: string
  href: string
  icon: LucideIcon
}

interface SettingsNavGroup {
  label: string
  items: SettingsNavItem[]
}

const settingsNavGroups: SettingsNavGroup[] = [
  {
    label: 'Firma',
    items: [
      { label: 'Firmendaten', href: '/settings/company', icon: Building2 },
      { label: 'Bankdaten', href: '/settings/bank-details', icon: Landmark },
      { label: 'AGB', href: '/settings/terms-conditions', icon: ScrollText },
    ],
  },
  {
    label: 'Finanzen',
    items: [
      { label: 'Kleinunternehmer', href: '/settings/tax-exemption', icon: Scale },
      { label: 'Zahlungsbedingungen', href: '/settings/payment-terms', icon: CreditCard },
      { label: 'Steuersätze', href: '/settings/vat-schemes', icon: Percent },
      { label: 'Nummernkreise', href: '/settings/number-sequences', icon: Hash },
    ],
  },
  {
    label: 'Benutzer',
    items: [
      { label: 'Benutzer', href: '/settings/users', icon: Users },
      { label: 'Rollen', href: '/settings/roles', icon: ShieldCheck },
    ],
  },
  {
    label: 'System',
    items: [
      { label: 'E-Mail (SMTP)', href: '/settings/email', icon: Mail },
      { label: 'E-Mail-Konten', href: '/settings/email-accounts', icon: Mailbox },
      { label: 'Benachrichtigungen', href: '/settings/notifications', icon: Bell },
      { label: 'Integrationen', href: '/settings/integrations', icon: Puzzle },
      { label: 'Module', href: '/settings/modules', icon: Blocks },
      { label: 'Scanner-Geraete', href: '/settings/scanner-devices', icon: Smartphone },
      { label: 'Sprache & Region', href: '/settings/locale', icon: Globe },
    ],
  },
  {
    label: 'Erweitert',
    items: [
      { label: 'Eigene Felder', href: '/settings/custom-fields', icon: SlidersHorizontal },
      { label: 'Kategorien', href: '/settings/categories', icon: Tags },
      { label: 'Projekttypen', href: '/settings/project-types', icon: FolderTree },
      { label: 'Dokumentvorlagen', href: '/settings/document-templates', icon: FileText },
      { label: 'Labels & QR-Codes', href: '/settings/labels', icon: QrCode },
      { label: 'Backups', href: '/settings/backups', icon: Database },
      { label: 'Software-Update', href: '/settings/update', icon: Download },
    ],
  },
]

function SettingsNav() {
  return (
    <div className="settings-sidebar">
      <h2 className="settings-sidebar__title">Einstellungen</h2>
      {settingsNavGroups.map((group) => (
        <div key={group.label} className="settings-sidebar__group">
          <div className="settings-sidebar__group-label">{group.label}</div>
          {group.items.map((item) => {
            const IconComponent = item.icon
            return (
              <NavLink
                key={item.href}
                to={item.href}
                className={({ isActive }) =>
                  `settings-sidebar__link${isActive ? ' settings-sidebar__link--active' : ''}`
                }
              >
                <IconComponent size={16} />
                {item.label}
              </NavLink>
            )
          })}
        </div>
      ))}
    </div>
  )
}

export default SettingsNav
