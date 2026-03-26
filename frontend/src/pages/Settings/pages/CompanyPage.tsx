import { useState, useEffect } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { tenantApi, configApi } from '../../../services/api'
import { useAuthStore } from '../../../stores/authStore'
import { useIndustry } from '../../../hooks/useIndustry'
import { INDUSTRY_GROUPS, INDUSTRY_PROFILES } from '../../../config/industryProfiles'
import { SkeletonCard } from '../../../components/Skeleton/SkeletonLoader'
import '../Settings.scss'

interface TenantData {
  id: string
  name: string
  slug: string
  address_street: string
  address_city: string
  address_zip: string
  address_country: string
  currency: string
  tax_rate: number
  invoice_prefix: string
  default_language: string
  vat_id?: string
  trade_register?: string
  managing_director?: string
  phone?: string
  email?: string
}

function CompanyPage() {
  const tenantId = useAuthStore((s) => s.tenantId)
  const queryClient = useQueryClient()
  const [saveSuccess, setSaveSuccess] = useState(false)
  const [saveError, setSaveError] = useState(false)
  const { profileId, setIndustry, isSettingIndustry } = useIndustry()
  const [industrySuccess, setIndustrySuccess] = useState(false)

  const [form, setForm] = useState({
    name: '',
    address_street: '',
    address_zip: '',
    address_city: '',
    address_country: 'Deutschland',
    vat_id: '',
    trade_register: '',
    managing_director: '',
    phone: '',
    email: '',
    tax_number: '',
  })

  const { data: tenantData, isLoading: tenantLoading } = useQuery<TenantData>({
    queryKey: ['tenant', tenantId],
    queryFn: () => tenantApi.getById(tenantId!),
    enabled: !!tenantId,
  })

  const { data: companyConfig, isLoading: configLoading } = useQuery({
    queryKey: ['config', 'company.details'],
    queryFn: () => configApi.get('company.details'),
  })

  const isLoading = tenantLoading || configLoading

  useEffect(() => {
    if (tenantData || companyConfig) {
      setForm((prev) => ({
        ...prev,
        name: tenantData?.name || prev.name || '',
        address_street: tenantData?.address_street || prev.address_street || '',
        address_zip: tenantData?.address_zip || prev.address_zip || '',
        address_city: tenantData?.address_city || prev.address_city || '',
        address_country: tenantData?.address_country || prev.address_country || 'Deutschland',
        phone: tenantData?.phone || prev.phone || '',
        email: tenantData?.email || prev.email || '',
        vat_id: companyConfig?.vat_id || tenantData?.vat_id || prev.vat_id || '',
        trade_register: companyConfig?.trade_register || tenantData?.trade_register || prev.trade_register || '',
        managing_director: companyConfig?.managing_director || tenantData?.managing_director || prev.managing_director || '',
        tax_number: companyConfig?.tax_number || prev.tax_number || '',
      }))
    }
  }, [tenantData, companyConfig])

  const updateTenantMutation = useMutation({
    mutationFn: (data: Partial<TenantData>) => tenantApi.update(tenantId!, data),
  })

  const updateConfigMutation = useMutation({
    mutationFn: (data: Record<string, string>) => configApi.set('company.details', data),
  })

  const handleSave = async () => {
    setSaveError(false)
    try {
      const tenantFields: Partial<TenantData> = {
        name: form.name,
        address_street: form.address_street,
        address_zip: form.address_zip,
        address_city: form.address_city,
        address_country: form.address_country,
        phone: form.phone,
        email: form.email,
      }
      const configFields = {
        trade_register: form.trade_register,
        tax_number: form.tax_number,
        vat_id: form.vat_id,
        managing_director: form.managing_director,
      }

      await Promise.all([
        tenantId ? updateTenantMutation.mutateAsync(tenantFields) : Promise.resolve(),
        updateConfigMutation.mutateAsync(configFields),
      ])

      queryClient.invalidateQueries({ queryKey: ['tenant', tenantId] })
      queryClient.invalidateQueries({ queryKey: ['config', 'company.details'] })
      setSaveSuccess(true)
      setTimeout(() => setSaveSuccess(false), 3000)
    } catch {
      setSaveError(true)
      setTimeout(() => setSaveError(false), 5000)
    }
  }

  const isSaving = updateTenantMutation.isPending || updateConfigMutation.isPending

  const update = (field: string, value: string) =>
    setForm((prev) => ({ ...prev, [field]: value }))

  if (isLoading) return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Firmendaten</h1>
        <p>Verwalten Sie die Stammdaten Ihres Unternehmens.</p>
      </div>
      <SkeletonCard count={3} />
    </div>
  )

  return (
    <div className="sp-page">
      <div className="sp-header">
        <h1>Firmendaten</h1>
        <p>Verwalten Sie die Stammdaten Ihres Unternehmens.</p>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Branche</h3>
        <div className="sp-grid">
          <div className="sp-field sp-full">
            <label className="sp-label">Branchenprofil</label>
            <select
              className="sp-select"
              value={profileId}
              disabled={isSettingIndustry}
              onChange={(e) => {
                setIndustry(e.target.value, {
                  onSuccess: () => {
                    setIndustrySuccess(true)
                    setTimeout(() => setIndustrySuccess(false), 3000)
                  },
                })
              }}
            >
              {INDUSTRY_GROUPS.map((group) => (
                <optgroup key={group.id} label={group.label}>
                  {group.profiles.map((pid) => {
                    const p = INDUSTRY_PROFILES[pid]
                    return p ? (
                      <option key={pid} value={pid}>
                        {p.label}
                      </option>
                    ) : null
                  })}
                </optgroup>
              ))}
            </select>
            <p style={{ marginTop: '0.25rem', fontSize: '0.75rem', color: 'var(--color-text-muted)' }}>
              Das Branchenprofil bestimmt Bezeichnungen, Funktionen und Kategorien.
            </p>
            {industrySuccess && (
              <span className="sp-msg--success" style={{ display: 'inline-block', marginTop: '0.5rem' }}>Branche gespeichert!</span>
            )}
          </div>
        </div>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Unternehmen</h3>
        <div className="sp-grid">
          <div className="sp-field sp-full">
            <label className="sp-label">Unternehmensname</label>
            <input className="sp-input" value={form.name} onChange={(e) => update('name', e.target.value)} />
          </div>
          <div className="sp-field">
            <label className="sp-label">Geschäftsführer</label>
            <input className="sp-input" value={form.managing_director} onChange={(e) => update('managing_director', e.target.value)} />
          </div>
          <div className="sp-field">
            <label className="sp-label">Handelsregister</label>
            <input className="sp-input" value={form.trade_register} onChange={(e) => update('trade_register', e.target.value)} placeholder="z.B. HRB 12345" />
          </div>
          <div className="sp-field">
            <label className="sp-label">Steuernummer</label>
            <input className="sp-input" value={form.tax_number} onChange={(e) => update('tax_number', e.target.value)} placeholder="z.B. 123/456/78901" />
          </div>
        </div>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Adresse</h3>
        <div className="sp-grid">
          <div className="sp-field sp-full">
            <label className="sp-label">Straße</label>
            <input className="sp-input" value={form.address_street} onChange={(e) => update('address_street', e.target.value)} />
          </div>
          <div className="sp-field">
            <label className="sp-label">PLZ</label>
            <input className="sp-input" value={form.address_zip} onChange={(e) => update('address_zip', e.target.value)} />
          </div>
          <div className="sp-field">
            <label className="sp-label">Stadt</label>
            <input className="sp-input" value={form.address_city} onChange={(e) => update('address_city', e.target.value)} />
          </div>
          <div className="sp-field sp-full">
            <label className="sp-label">Land</label>
            <input className="sp-input" value={form.address_country} onChange={(e) => update('address_country', e.target.value)} />
          </div>
        </div>
      </div>

      <div className="sp-card">
        <h3 className="sp-card__title">Kontakt & Steuer</h3>
        <div className="sp-grid">
          <div className="sp-field">
            <label className="sp-label">Telefon</label>
            <input className="sp-input" value={form.phone} onChange={(e) => update('phone', e.target.value)} placeholder="+49 ..." />
          </div>
          <div className="sp-field">
            <label className="sp-label">E-Mail</label>
            <input className="sp-input" type="email" value={form.email} onChange={(e) => update('email', e.target.value)} />
          </div>
          <div className="sp-field">
            <label className="sp-label">USt-IdNr.</label>
            <input className="sp-input" value={form.vat_id} onChange={(e) => update('vat_id', e.target.value)} placeholder="DE123456789" />
          </div>
        </div>
      </div>

      <div className="sp-footer">
        {saveSuccess && <span className="sp-msg--success">Erfolgreich gespeichert!</span>}
        {saveError && <span className="sp-msg--error">Fehler beim Speichern</span>}
        <button className="sp-btn sp-btn--primary" onClick={handleSave} disabled={isSaving}>
          {isSaving ? 'Speichern...' : 'Speichern'}
        </button>
      </div>
    </div>
  )
}

export default CompanyPage
