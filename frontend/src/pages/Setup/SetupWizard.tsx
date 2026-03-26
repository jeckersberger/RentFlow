import { useState, useEffect, useMemo } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMutation } from '@tanstack/react-query'
import { setupApi, configApi, SetupRequest } from '../../services/api'
import {
  Package,
  FolderKanban,
  Receipt,
  Users,
  Inbox,
  Wrench,
  Bot,
  Globe,
  Monitor,
  Sofa,
  Truck,
  type LucideIcon,
} from 'lucide-react'
import { INDUSTRY_GROUPS, INDUSTRY_PROFILES } from '../../config/industryProfiles'
import './SetupWizard.scss'

type StepType = 'token' | 'industry' | 'company' | 'admin' | 'modules' | 'confirm'

const GROUP_ICONS: Record<string, LucideIcon> = {
  Monitor,
  Wrench,
  Sofa,
  Truck,
  Package,
}

interface FormData extends Partial<SetupRequest> {
  confirmPassword?: string
  ust_id?: string
  handelsregister?: string
  geschaeftsfuehrer?: string
  street?: string
  zip?: string
  city?: string
  country?: string
}

interface ModuleDefinition {
  id: string
  name: string
  description: string
  icon: LucideIcon
  default: boolean
}

const MODULES: ModuleDefinition[] = [
  {
    id: 'warehouse',
    name: 'Lagerverwaltung',
    description: 'Equipment, Scanner, Lager, Labels',
    icon: Package,
    default: true,
  },
  {
    id: 'projects',
    name: 'Projektverwaltung',
    description: 'Projekte, Kalender, Engpässe',
    icon: FolderKanban,
    default: true,
  },
  {
    id: 'finance',
    name: 'Finanzen',
    description: 'Rechnungen, Angebote, Kontakte',
    icon: Receipt,
    default: true,
  },
  {
    id: 'team',
    name: 'Personalverwaltung',
    description: 'Crew, Zeiterfassung, Transport',
    icon: Users,
    default: false,
  },
  {
    id: 'communication',
    name: 'Kommunikation',
    description: 'E-Mail Posteingang & Ausgang',
    icon: Inbox,
    default: false,
  },
  {
    id: 'workshop',
    name: 'Werkstatt',
    description: 'Reparaturen, Prüfungen, Bestandszählungen',
    icon: Wrench,
    default: false,
  },
  {
    id: 'ai',
    name: 'KI-Assistent',
    description: 'Preis-Optimierung, Demand Forecasting',
    icon: Bot,
    default: false,
  },
  {
    id: 'federation',
    name: 'Federation',
    description: 'Firmen verbinden, Equipment teilen',
    icon: Globe,
    default: false,
  },
]

const STEPS: { type: StepType; label: string }[] = [
  { type: 'token', label: 'Setup Token' },
  { type: 'industry', label: 'Branche' },
  { type: 'company', label: 'Firmenangaben' },
  { type: 'admin', label: 'Admin-Konto' },
  { type: 'modules', label: 'Module' },
  { type: 'confirm', label: 'Bestätigung' },
]

function getPasswordStrength(password: string): { score: number; label: string; className: string } {
  if (!password) return { score: 0, label: '', className: '' }

  let score = 0
  if (password.length >= 8) score++
  if (password.length >= 12) score++
  if (password.length >= 16) score++
  if (/[A-Z]/.test(password)) score++
  if (/[a-z]/.test(password)) score++
  if (/[0-9]/.test(password)) score++
  if (/[^A-Za-z0-9]/.test(password)) score++

  if (score <= 2) return { score: 1, label: 'Schwach', className: 'weak' }
  if (score <= 4) return { score: 2, label: 'Mittel', className: 'medium' }
  if (score <= 5) return { score: 3, label: 'Stark', className: 'strong' }
  return { score: 4, label: 'Sehr stark', className: 'very-strong' }
}

function SetupWizard() {
  const navigate = useNavigate()
  const [currentStep, setCurrentStep] = useState<StepType>('token')
  const [completedSteps, setCompletedSteps] = useState<StepType[]>([])
  const [error, setError] = useState('')
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({})
  const [selectedModules, setSelectedModules] = useState<string[]>(
    MODULES.filter((m) => m.default).map((m) => m.id)
  )
  const [selectedGroup, setSelectedGroup] = useState<string | null>(null)
  const [selectedIndustry, setSelectedIndustry] = useState<string>('event_tech')

  const [formData, setFormData] = useState<FormData>({
    setup_token: '',
    company_name: '',
    company_slug: '',
    company_address: '',
    currency: 'EUR',
    tax_rate: 19.0,
    invoice_prefix: 'RE-',
    admin_name: '',
    admin_email: '',
    admin_password: '',
    confirmPassword: '',
    language: 'de',
    ust_id: '',
    handelsregister: '',
    geschaeftsfuehrer: '',
    street: '',
    zip: '',
    city: '',
    country: 'Deutschland',
  })

  const setupMutation = useMutation({
    mutationFn: async (data: SetupRequest) => {
      const result = await setupApi.complete(data)
      // Save selected modules and industry profile to config after setup completes
      try {
        await Promise.all([
          configApi.set('modules.enabled', selectedModules),
          configApi.set('industry_profile', selectedIndustry),
        ])
      } catch {
        // Non-critical: can be configured later in settings
      }
      return result
    },
    onSuccess: () => {
      navigate('/login')
    },
    onError: (err: any) => {
      const errorMessage = err?.response?.data?.message || 'Setup fehlgeschlagen. Bitte versuchen Sie es später erneut.'

      if (err?.response?.status === 403) {
        setError('Setup wurde bereits abgeschlossen.')
      } else {
        setError(errorMessage)
      }
    },
  })

  // Generate slug from company name
  useEffect(() => {
    if (currentStep === 'company' && formData.company_name) {
      const umlautMap: Record<string, string> = { ä: 'ae', ö: 'oe', ü: 'ue' }
      const generatedSlug = formData.company_name
        .toLowerCase()
        .replace(/[äöü]/g, (char) => umlautMap[char] || char)
        .replace(/[^\w\s-]/g, '')
        .replace(/\s+/g, '-')
        .replace(/-+/g, '-')
        .trim()

      setFormData((prev) => ({
        ...prev,
        company_slug: generatedSlug,
      }))
    }
  }, [formData.company_name, currentStep])

  const currentStepIndex = STEPS.findIndex((s) => s.type === currentStep)

  const passwordStrength = useMemo(
    () => getPasswordStrength(formData.admin_password || ''),
    [formData.admin_password]
  )

  const validateToken = (): boolean => {
    const errors: Record<string, string> = {}

    if (!formData.setup_token?.trim()) {
      errors.setup_token = 'Setup-Token ist erforderlich'
    }

    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const validateCompany = (): boolean => {
    const errors: Record<string, string> = {}

    if (!formData.company_name?.trim()) {
      errors.company_name = 'Firmenname ist erforderlich'
    }

    if (!formData.company_slug?.trim()) {
      errors.company_slug = 'Firmenkürzel ist erforderlich'
    } else if (!/^[a-z0-9-]+$/.test(formData.company_slug)) {
      errors.company_slug = 'Firmenkürzel darf nur Kleinbuchstaben, Zahlen und Bindestriche enthalten'
    }

    if (formData.tax_rate! < 0 || formData.tax_rate! > 100) {
      errors.tax_rate = 'Steuersatz muss zwischen 0 und 100 liegen'
    }

    if (!formData.invoice_prefix?.trim()) {
      errors.invoice_prefix = 'Rechnungsprefix ist erforderlich'
    }

    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const validateAdmin = (): boolean => {
    const errors: Record<string, string> = {}

    if (!formData.admin_name?.trim()) {
      errors.admin_name = 'Name des Administrators ist erforderlich'
    }

    if (!formData.admin_email?.trim()) {
      errors.admin_email = 'E-Mail ist erforderlich'
    } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(formData.admin_email)) {
      errors.admin_email = 'Bitte geben Sie eine gültige E-Mail-Adresse ein'
    }

    if (!formData.admin_password?.trim()) {
      errors.admin_password = 'Passwort ist erforderlich'
    } else {
      const pw = formData.admin_password
      if (pw.length < 12 || !/[A-Z]/.test(pw) || !/[a-z]/.test(pw) || !/[0-9]/.test(pw) || !/[^A-Za-z0-9]/.test(pw)) {
        errors.admin_password = 'Passwort erfuellt nicht alle Anforderungen (siehe oben)'
      }
    }

    if (!formData.confirmPassword?.trim()) {
      errors.confirmPassword = 'Passwortbestätigung ist erforderlich'
    } else if (formData.admin_password !== formData.confirmPassword) {
      errors.confirmPassword = 'Passwörter stimmen nicht überein'
    }

    setFieldErrors(errors)
    return Object.keys(errors).length === 0
  }

  const handleNext = () => {
    setError('')

    let isValid = false
    if (currentStep === 'token') isValid = validateToken()
    if (currentStep === 'industry') isValid = !!selectedIndustry // must pick an industry
    if (currentStep === 'company') isValid = validateCompany()
    if (currentStep === 'admin') isValid = validateAdmin()
    if (currentStep === 'modules') isValid = true // modules step always valid

    if (!isValid) return

    if (!completedSteps.includes(currentStep)) {
      setCompletedSteps([...completedSteps, currentStep])
    }

    const nextStep = STEPS[currentStepIndex + 1]
    if (nextStep) {
      setCurrentStep(nextStep.type)
      setFieldErrors({})
    }
  }

  const handleBack = () => {
    const prevStep = STEPS[currentStepIndex - 1]
    if (prevStep) {
      setCurrentStep(prevStep.type)
      setFieldErrors({})
      setError('')
    }
  }

  const handleSubmit = async () => {
    setError('')

    const submitData: SetupRequest = {
      setup_token: formData.setup_token!,
      company_name: formData.company_name!,
      company_slug: formData.company_slug!,
      company_address: formData.company_address || '',
      currency: formData.currency!,
      tax_rate: formData.tax_rate!,
      invoice_prefix: formData.invoice_prefix!,
      admin_name: formData.admin_name!,
      admin_email: formData.admin_email!,
      admin_password: formData.admin_password!,
      language: formData.language!,
    }

    setupMutation.mutate(submitData)
  }

  const toggleModule = (moduleId: string) => {
    setSelectedModules((prev) =>
      prev.includes(moduleId)
        ? prev.filter((id) => id !== moduleId)
        : [...prev, moduleId]
    )
  }

  const renderStepContent = () => {
    switch (currentStep) {
      case 'token':
        return (
          <div className="setup-step">
            <h2 className="setup-step__title">Setup-Token</h2>
            <p className="setup-step__description">
              Geben Sie den Setup-Token ein, der in den Server-Logs angezeigt wurde.
            </p>

            <div className="setup-form__group">
              <label htmlFor="setup_token" className="setup-form__label">
                Setup-Token
              </label>
              <input
                id="setup_token"
                type="password"
                className={`setup-form__input ${fieldErrors.setup_token ? 'setup-form__input--error' : ''}`}
                value={formData.setup_token || ''}
                onChange={(e) => {
                  setFormData({ ...formData, setup_token: e.target.value })
                  if (fieldErrors.setup_token) {
                    setFieldErrors({ ...fieldErrors, setup_token: '' })
                  }
                }}
                placeholder="Z. B. sk_live_abc123..."
                autoComplete="off"
              />
              {fieldErrors.setup_token && (
                <p className="setup-form__error">{fieldErrors.setup_token}</p>
              )}
            </div>
          </div>
        )

      case 'industry':
        return (
          <div className="setup-step">
            <h2 className="setup-step__title">Was vermieten/verleihen Sie?</h2>
            <p className="setup-step__description">
              Waehlen Sie Ihre Branche, um RentFlow optimal fuer Sie einzurichten.
            </p>

            {!selectedGroup ? (
              <div className="module-grid">
                {INDUSTRY_GROUPS.map((group) => {
                  const IconComponent = GROUP_ICONS[group.icon] || Package
                  return (
                    <div
                      key={group.id}
                      className="module-card"
                      onClick={() => setSelectedGroup(group.id)}
                    >
                      <div className="module-card__header">
                        <div className="module-card__icon">
                          <IconComponent size={20} />
                        </div>
                      </div>
                      <div className="module-card__name">{group.label}</div>
                      <div className="module-card__description">
                        {group.profiles.length} Branchen
                      </div>
                    </div>
                  )
                })}
              </div>
            ) : (
              <>
                <button
                  type="button"
                  className="setup-actions__button setup-actions__button--secondary"
                  style={{ marginBottom: '1rem', flex: 'none', width: 'auto', padding: '0.4rem 1rem', fontSize: '0.85rem' }}
                  onClick={() => setSelectedGroup(null)}
                >
                  Zurueck zur Gruppenauswahl
                </button>
                <div className="module-grid">
                  {INDUSTRY_GROUPS.find((g) => g.id === selectedGroup)?.profiles.map((pid) => {
                    const p = INDUSTRY_PROFILES[pid]
                    if (!p) return null
                    const isSelected = selectedIndustry === pid
                    return (
                      <div
                        key={pid}
                        className={`module-card ${isSelected ? 'module-card--active' : ''}`}
                        onClick={() => setSelectedIndustry(pid)}
                      >
                        <div className="module-card__header">
                          <div className="module-card__icon">
                            <Package size={18} />
                          </div>
                          {isSelected && (
                            <div className="module-card__toggle module-card__toggle--on">
                              <div className="module-card__toggle-knob" />
                            </div>
                          )}
                        </div>
                        <div className="module-card__name">{p.label}</div>
                        <div className="module-card__description">
                          {p.categories.slice(0, 4).join(', ')}...
                        </div>
                      </div>
                    )
                  })}
                </div>
              </>
            )}

            {selectedIndustry && INDUSTRY_PROFILES[selectedIndustry] && (
              <p className="setup-step__hint">
                Ausgewaehlt: <strong>{INDUSTRY_PROFILES[selectedIndustry].label}</strong> — Sie koennen dies jederzeit in den Einstellungen aendern.
              </p>
            )}
          </div>
        )

      case 'company':
        return (
          <div className="setup-step">
            <h2 className="setup-step__title">Firmenangaben</h2>
            <p className="setup-step__description">
              Konfigurieren Sie die grundlegenden Informationen Ihres Unternehmens.
            </p>

            <div className="setup-form__group">
              <label htmlFor="company_name" className="setup-form__label">
                Firmenname <span className="setup-form__required">*</span>
              </label>
              <input
                id="company_name"
                type="text"
                className={`setup-form__input ${fieldErrors.company_name ? 'setup-form__input--error' : ''}`}
                value={formData.company_name || ''}
                onChange={(e) => {
                  setFormData({ ...formData, company_name: e.target.value })
                  if (fieldErrors.company_name) {
                    setFieldErrors({ ...fieldErrors, company_name: '' })
                  }
                }}
                placeholder="Z. B. RentFlow GmbH"
              />
              {fieldErrors.company_name && (
                <p className="setup-form__error">{fieldErrors.company_name}</p>
              )}
            </div>

            <div className="setup-form__group">
              <label htmlFor="company_slug" className="setup-form__label">
                Firmenkürzel <span className="setup-form__required">*</span>
              </label>
              <input
                id="company_slug"
                type="text"
                className={`setup-form__input ${fieldErrors.company_slug ? 'setup-form__input--error' : ''}`}
                value={formData.company_slug || ''}
                onChange={(e) => {
                  setFormData({ ...formData, company_slug: e.target.value })
                  if (fieldErrors.company_slug) {
                    setFieldErrors({ ...fieldErrors, company_slug: '' })
                  }
                }}
                placeholder="Z. B. rentflow-gmbh"
              />
              {fieldErrors.company_slug && (
                <p className="setup-form__error">{fieldErrors.company_slug}</p>
              )}
            </div>

            <div className="setup-form__group">
              <label htmlFor="street" className="setup-form__label">
                Strasse + Hausnummer
              </label>
              <input
                id="street"
                type="text"
                className="setup-form__input"
                value={formData.street || ''}
                onChange={(e) => {
                  const street = e.target.value
                  setFormData(prev => ({
                    ...prev,
                    street,
                    company_address: `${street}, ${prev.zip || ''} ${prev.city || ''}, ${prev.country || 'Deutschland'}`.trim(),
                  }))
                }}
                placeholder="z.B. Musterstrasse 12"
              />
            </div>

            <div className="setup-form__row">
              <div className="setup-form__group">
                <label htmlFor="zip" className="setup-form__label">
                  PLZ
                </label>
                <input
                  id="zip"
                  type="text"
                  className="setup-form__input"
                  value={formData.zip || ''}
                  onChange={(e) => {
                    const zip = e.target.value
                    setFormData(prev => ({
                      ...prev,
                      zip,
                      company_address: `${prev.street || ''}, ${zip} ${prev.city || ''}, ${prev.country || 'Deutschland'}`.trim(),
                    }))
                  }}
                  placeholder="z.B. 80331"
                  maxLength={10}
                />
              </div>

              <div className="setup-form__group">
                <label htmlFor="city" className="setup-form__label">
                  Stadt
                </label>
                <input
                  id="city"
                  type="text"
                  className="setup-form__input"
                  value={formData.city || ''}
                  onChange={(e) => {
                    const city = e.target.value
                    setFormData(prev => ({
                      ...prev,
                      city,
                      company_address: `${prev.street || ''}, ${prev.zip || ''} ${city}, ${prev.country || 'Deutschland'}`.trim(),
                    }))
                  }}
                  placeholder="z.B. Muenchen"
                />
              </div>
            </div>

            <div className="setup-form__group">
              <label htmlFor="country" className="setup-form__label">
                Land
              </label>
              <input
                id="country"
                type="text"
                className="setup-form__input"
                value={formData.country || 'Deutschland'}
                onChange={(e) => {
                  const country = e.target.value
                  setFormData(prev => ({
                    ...prev,
                    country,
                    company_address: `${prev.street || ''}, ${prev.zip || ''} ${prev.city || ''}, ${country}`.trim(),
                  }))
                }}
                placeholder="z.B. Deutschland"
              />
            </div>

            <div className="setup-form__row">
              <div className="setup-form__group">
                <label htmlFor="geschaeftsfuehrer" className="setup-form__label">
                  Geschäftsführer
                </label>
                <input
                  id="geschaeftsfuehrer"
                  type="text"
                  className="setup-form__input"
                  value={formData.geschaeftsfuehrer || ''}
                  onChange={(e) => setFormData({ ...formData, geschaeftsfuehrer: e.target.value })}
                  placeholder="Z. B. Max Mustermann"
                />
              </div>

              <div className="setup-form__group">
                <label htmlFor="ust_id" className="setup-form__label">
                  USt-IdNr.
                </label>
                <input
                  id="ust_id"
                  type="text"
                  className="setup-form__input"
                  value={formData.ust_id || ''}
                  onChange={(e) => setFormData({ ...formData, ust_id: e.target.value })}
                  placeholder="Z. B. DE123456789"
                />
              </div>
            </div>

            <div className="setup-form__group">
              <label htmlFor="handelsregister" className="setup-form__label">
                Handelsregister
              </label>
              <input
                id="handelsregister"
                type="text"
                className="setup-form__input"
                value={formData.handelsregister || ''}
                onChange={(e) => setFormData({ ...formData, handelsregister: e.target.value })}
                placeholder="Z. B. HRB 12345, Amtsgericht München"
              />
            </div>

            <div className="setup-form__row">
              <div className="setup-form__group">
                <label htmlFor="currency" className="setup-form__label">
                  Währung
                </label>
                <select
                  id="currency"
                  className="setup-form__select"
                  value={formData.currency || 'EUR'}
                  onChange={(e) => setFormData({ ...formData, currency: e.target.value })}
                >
                  <option value="EUR">EUR - Euro</option>
                  <option value="USD">USD - US-Dollar</option>
                  <option value="CHF">CHF - Schweizer Franken</option>
                </select>
              </div>

              <div className="setup-form__group">
                <label htmlFor="tax_rate" className="setup-form__label">
                  Steuersatz (%)
                </label>
                <input
                  id="tax_rate"
                  type="number"
                  className={`setup-form__input ${fieldErrors.tax_rate ? 'setup-form__input--error' : ''}`}
                  value={formData.tax_rate || 19.0}
                  onChange={(e) => {
                    setFormData({ ...formData, tax_rate: parseFloat(e.target.value) })
                    if (fieldErrors.tax_rate) {
                      setFieldErrors({ ...fieldErrors, tax_rate: '' })
                    }
                  }}
                  step="0.01"
                  min="0"
                  max="100"
                />
                {fieldErrors.tax_rate && (
                  <p className="setup-form__error">{fieldErrors.tax_rate}</p>
                )}
              </div>
            </div>

            <div className="setup-form__group">
              <label htmlFor="invoice_prefix" className="setup-form__label">
                Rechnungsprefix <span className="setup-form__required">*</span>
              </label>
              <input
                id="invoice_prefix"
                type="text"
                className={`setup-form__input ${fieldErrors.invoice_prefix ? 'setup-form__input--error' : ''}`}
                value={formData.invoice_prefix || ''}
                onChange={(e) => {
                  setFormData({ ...formData, invoice_prefix: e.target.value })
                  if (fieldErrors.invoice_prefix) {
                    setFieldErrors({ ...fieldErrors, invoice_prefix: '' })
                  }
                }}
                placeholder="Z. B. RE-"
              />
              {fieldErrors.invoice_prefix && (
                <p className="setup-form__error">{fieldErrors.invoice_prefix}</p>
              )}
            </div>
          </div>
        )

      case 'admin':
        return (
          <div className="setup-step">
            <h2 className="setup-step__title">Admin-Konto</h2>
            <p className="setup-step__description">
              Erstellen Sie das Administratorkonto für die Anmeldung.
            </p>

            <div className="setup-form__group">
              <label htmlFor="admin_name" className="setup-form__label">
                Name <span className="setup-form__required">*</span>
              </label>
              <input
                id="admin_name"
                type="text"
                className={`setup-form__input ${fieldErrors.admin_name ? 'setup-form__input--error' : ''}`}
                value={formData.admin_name || ''}
                onChange={(e) => {
                  setFormData({ ...formData, admin_name: e.target.value })
                  if (fieldErrors.admin_name) {
                    setFieldErrors({ ...fieldErrors, admin_name: '' })
                  }
                }}
                placeholder="Z. B. Max Mustermann"
              />
              {fieldErrors.admin_name && (
                <p className="setup-form__error">{fieldErrors.admin_name}</p>
              )}
            </div>

            <div className="setup-form__group">
              <label htmlFor="admin_email" className="setup-form__label">
                E-Mail-Adresse <span className="setup-form__required">*</span>
              </label>
              <input
                id="admin_email"
                type="email"
                className={`setup-form__input ${fieldErrors.admin_email ? 'setup-form__input--error' : ''}`}
                value={formData.admin_email || ''}
                onChange={(e) => {
                  setFormData({ ...formData, admin_email: e.target.value })
                  if (fieldErrors.admin_email) {
                    setFieldErrors({ ...fieldErrors, admin_email: '' })
                  }
                }}
                placeholder="admin@example.com"
                autoComplete="email"
              />
              {fieldErrors.admin_email && (
                <p className="setup-form__error">{fieldErrors.admin_email}</p>
              )}
            </div>

            <div className="setup-form__group">
              <label htmlFor="admin_password" className="setup-form__label">
                Passwort <span className="setup-form__required">*</span>
              </label>
              <input
                id="admin_password"
                type="password"
                className={`setup-form__input ${fieldErrors.admin_password ? 'setup-form__input--error' : ''}`}
                value={formData.admin_password || ''}
                onChange={(e) => {
                  setFormData({ ...formData, admin_password: e.target.value })
                  if (fieldErrors.admin_password) {
                    setFieldErrors({ ...fieldErrors, admin_password: '' })
                  }
                }}
                placeholder="Mindestens 12 Zeichen"
                autoComplete="new-password"
              />
              {formData.admin_password && (
                <div className="password-strength">
                  <div className="password-strength__bar">
                    <div
                      className={`password-strength__fill password-strength__fill--${passwordStrength.className}`}
                      style={{ width: `${(passwordStrength.score / 4) * 100}%` }}
                    />
                  </div>
                  <span className={`password-strength__label password-strength__label--${passwordStrength.className}`}>
                    {passwordStrength.label}
                  </span>
                </div>
              )}
              {(() => {
                const pw = formData.admin_password || ''
                const checks = [
                  { ok: pw.length >= 12, label: 'Mindestens 12 Zeichen' },
                  { ok: /[A-Z]/.test(pw), label: 'Mindestens ein Grossbuchstabe' },
                  { ok: /[a-z]/.test(pw), label: 'Mindestens ein Kleinbuchstabe' },
                  { ok: /[0-9]/.test(pw), label: 'Mindestens eine Zahl' },
                  { ok: /[^A-Za-z0-9]/.test(pw), label: 'Mindestens ein Sonderzeichen' },
                ]
                return (
                  <div style={{ marginTop: '0.5rem', fontSize: '0.8rem' }}>
                    {checks.map((c, i) => (
                      <div key={i} style={{ display: 'flex', alignItems: 'center', gap: '0.4rem', marginBottom: '0.15rem', color: !pw ? 'var(--color-text-muted)' : c.ok ? '#10b981' : '#ef4444' }}>
                        <span>{!pw ? '\u25CB' : c.ok ? '\u2713' : '\u2717'}</span>
                        <span>{c.label}</span>
                      </div>
                    ))}
                  </div>
                )
              })()}
              {fieldErrors.admin_password && (
                <p className="setup-form__error">{fieldErrors.admin_password}</p>
              )}
            </div>

            <div className="setup-form__group">
              <label htmlFor="confirmPassword" className="setup-form__label">
                Passwort bestätigen <span className="setup-form__required">*</span>
              </label>
              <input
                id="confirmPassword"
                type="password"
                className={`setup-form__input ${fieldErrors.confirmPassword ? 'setup-form__input--error' : ''}`}
                value={formData.confirmPassword || ''}
                onChange={(e) => {
                  setFormData({ ...formData, confirmPassword: e.target.value })
                  if (fieldErrors.confirmPassword) {
                    setFieldErrors({ ...fieldErrors, confirmPassword: '' })
                  }
                }}
                placeholder="Passwort wiederholen"
                autoComplete="new-password"
              />
              {fieldErrors.confirmPassword && (
                <p className="setup-form__error">{fieldErrors.confirmPassword}</p>
              )}
            </div>
          </div>
        )

      case 'modules':
        return (
          <div className="setup-step">
            <h2 className="setup-step__title">Module</h2>
            <p className="setup-step__description">
              Welche Funktionen benötigen Sie? Wählen Sie die Module aus, die Sie nutzen möchten.
            </p>

            <div className="module-grid">
              {MODULES.map((mod) => {
                const isSelected = selectedModules.includes(mod.id)
                const IconComponent = mod.icon
                return (
                  <div
                    key={mod.id}
                    className={`module-card ${isSelected ? 'module-card--active' : ''}`}
                    onClick={() => toggleModule(mod.id)}
                  >
                    <div className="module-card__header">
                      <div className="module-card__icon">
                        <IconComponent size={20} />
                      </div>
                      <div
                        className={`module-card__toggle ${isSelected ? 'module-card__toggle--on' : ''}`}
                        onClick={(e) => {
                          e.stopPropagation()
                          toggleModule(mod.id)
                        }}
                      >
                        <div className="module-card__toggle-knob" />
                      </div>
                    </div>
                    <div className="module-card__name">{mod.name}</div>
                    <div className="module-card__description">{mod.description}</div>
                    {mod.default && (
                      <span className="module-card__badge">Empfohlen</span>
                    )}
                  </div>
                )
              })}
            </div>

            <p className="setup-step__hint">
              Sie können Module jederzeit in den Einstellungen aktivieren oder deaktivieren.
            </p>
          </div>
        )

      case 'confirm':
        return (
          <div className="setup-step">
            <h2 className="setup-step__title">Bestätigung</h2>
            <p className="setup-step__description">
              Überprüfen Sie Ihre Eingaben, bevor Sie das Setup abschließen.
            </p>

            <div className="setup-summary">
              <div className="setup-summary__section">
                <h3 className="setup-summary__heading">Branche</h3>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Branchenprofil:</span>
                  <span className="setup-summary__value">
                    {INDUSTRY_PROFILES[selectedIndustry]?.label || 'Veranstaltungstechnik'}
                  </span>
                </div>
              </div>

              <div className="setup-summary__section">
                <h3 className="setup-summary__heading">Firmenangaben</h3>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Firmenname:</span>
                  <span className="setup-summary__value">{formData.company_name}</span>
                </div>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Firmenkürzel:</span>
                  <span className="setup-summary__value">{formData.company_slug}</span>
                </div>
                {formData.company_address && (
                  <div className="setup-summary__item">
                    <span className="setup-summary__label">Adresse:</span>
                    <span className="setup-summary__value">{formData.company_address}</span>
                  </div>
                )}
                {formData.geschaeftsfuehrer && (
                  <div className="setup-summary__item">
                    <span className="setup-summary__label">Geschäftsführer:</span>
                    <span className="setup-summary__value">{formData.geschaeftsfuehrer}</span>
                  </div>
                )}
                {formData.ust_id && (
                  <div className="setup-summary__item">
                    <span className="setup-summary__label">USt-IdNr.:</span>
                    <span className="setup-summary__value">{formData.ust_id}</span>
                  </div>
                )}
                {formData.handelsregister && (
                  <div className="setup-summary__item">
                    <span className="setup-summary__label">Handelsregister:</span>
                    <span className="setup-summary__value">{formData.handelsregister}</span>
                  </div>
                )}
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Währung:</span>
                  <span className="setup-summary__value">{formData.currency}</span>
                </div>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Steuersatz:</span>
                  <span className="setup-summary__value">{formData.tax_rate}%</span>
                </div>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Rechnungsprefix:</span>
                  <span className="setup-summary__value">{formData.invoice_prefix}</span>
                </div>
              </div>

              <div className="setup-summary__section">
                <h3 className="setup-summary__heading">Admin-Konto</h3>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Name:</span>
                  <span className="setup-summary__value">{formData.admin_name}</span>
                </div>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">E-Mail:</span>
                  <span className="setup-summary__value">{formData.admin_email}</span>
                </div>
                <div className="setup-summary__item">
                  <span className="setup-summary__label">Passwort:</span>
                  <span className="setup-summary__value">••••••••</span>
                </div>
              </div>

              <div className="setup-summary__section">
                <h3 className="setup-summary__heading">Aktive Module</h3>
                <div className="setup-summary__modules">
                  {MODULES.filter((m) => selectedModules.includes(m.id)).map((mod) => {
                    const IconComponent = mod.icon
                    return (
                      <div key={mod.id} className="setup-summary__module-tag">
                        <IconComponent size={14} />
                        <span>{mod.name}</span>
                      </div>
                    )
                  })}
                </div>
                {MODULES.filter((m) => !selectedModules.includes(m.id)).length > 0 && (
                  <p className="setup-summary__module-note">
                    {MODULES.filter((m) => !selectedModules.includes(m.id)).length} weitere Module können später aktiviert werden
                  </p>
                )}
              </div>
            </div>
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className="setup-page">
      <div className="setup-card">
        <div className="setup-card__header">
          <div className="setup-card__logo">
            <span className="setup-card__logo-icon">📦</span>
          </div>
          <h1 className="setup-card__title">RentFlow</h1>
          <p className="setup-card__subtitle">
            Initialisierung - Bitte füllen Sie das Setup-Formular aus
          </p>
        </div>

        <div className="setup-progress">
          <div className="setup-progress__steps">
            {STEPS.map((step, index) => (
              <div
                key={step.type}
                className={`setup-progress__step ${
                  currentStep === step.type
                    ? 'setup-progress__step--active'
                    : completedSteps.includes(step.type)
                    ? 'setup-progress__step--completed'
                    : ''
                }`}
              >
                <div className="setup-progress__dot">
                  {completedSteps.includes(step.type) ? '✓' : index + 1}
                </div>
                {index < STEPS.length - 1 && (
                  <div className="setup-progress__line" />
                )}
              </div>
            ))}
          </div>
        </div>

        <div className="setup-content">
          {error && (
            <div className="setup-error" role="alert">
              <strong>Fehler:</strong> {error}
            </div>
          )}

          {renderStepContent()}

          <div className="setup-actions">
            <button
              type="button"
              className="setup-actions__button setup-actions__button--secondary"
              onClick={handleBack}
              disabled={currentStepIndex === 0 || setupMutation.isPending}
            >
              Zurück
            </button>

            {currentStep !== 'confirm' ? (
              <button
                type="button"
                className="setup-actions__button setup-actions__button--primary"
                onClick={handleNext}
                disabled={setupMutation.isPending}
              >
                Weiter
              </button>
            ) : (
              <button
                type="button"
                className="setup-actions__button setup-actions__button--primary"
                onClick={handleSubmit}
                disabled={setupMutation.isPending}
              >
                {setupMutation.isPending ? 'Setup wird abgeschlossen...' : 'Setup abschließen'}
              </button>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default SetupWizard
