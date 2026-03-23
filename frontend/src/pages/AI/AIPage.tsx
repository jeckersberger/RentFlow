import { useState } from 'react'
import { Link } from 'react-router-dom'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid,
  Tooltip, ResponsiveContainer, Legend, Area, AreaChart
} from 'recharts'
import styles from './AI.module.scss'

// ============================================================================
// DEMO DATA
// ============================================================================

const AI_PROVIDERS = [
  { name: 'Claude 3.5 Sonnet', status: 'active' as const, model: 'claude-3-5-sonnet', requests: 1247 },
  { name: 'GPT-4o', status: 'inactive' as const, model: 'gpt-4o', requests: 0 },
  { name: 'Gemini Pro', status: 'inactive' as const, model: 'gemini-pro', requests: 0 },
  { name: 'Ollama (Lokal)', status: 'inactive' as const, model: 'llama3', requests: 0 },
]

const PRICE_SUGGESTIONS = [
  { id: '1', name: 'Shure SM58', category: 'Audio', currentPrice: 15, suggestedPrice: 22, confidence: 'Hoch' as const, reason: 'Marktdurchschnitt liegt bei 25 EUR/Tag' },
  { id: '2', name: 'Sennheiser EW 100 G4', category: 'Audio', currentPrice: 35, suggestedPrice: 42, confidence: 'Hoch' as const, reason: 'Hohe Nachfrage, wenig Angebot' },
  { id: '3', name: 'Martin MAC Aura XB', category: 'Licht', currentPrice: 85, suggestedPrice: 95, confidence: 'Mittel' as const, reason: 'Saisonale Nachfragesteigerung' },
  { id: '4', name: 'Blackmagic ATEM Mini Pro', category: 'Video', currentPrice: 45, suggestedPrice: 55, confidence: 'Hoch' as const, reason: 'Streaming-Events stark nachgefragt' },
  { id: '5', name: 'QSC K12.2', category: 'Audio', currentPrice: 40, suggestedPrice: 38, confidence: 'Niedrig' as const, reason: 'Marktsaettigung, Preis leicht senken' },
  { id: '6', name: 'Chauvet Rogue R2 Wash', category: 'Licht', currentPrice: 55, suggestedPrice: 65, confidence: 'Mittel' as const, reason: 'Festival-Saison steht bevor' },
]

const DEMAND_FORECAST = [
  { month: 'Apr', audio: 65, licht: 45, video: 55 },
  { month: 'Mai', audio: 78, licht: 72, video: 60 },
  { month: 'Jun', audio: 92, licht: 95, video: 70 },
  { month: 'Jul', audio: 88, licht: 85, video: 65 },
  { month: 'Aug', audio: 75, licht: 70, video: 58 },
  { month: 'Sep', audio: 82, licht: 80, video: 72 },
]

const REVENUE_IMPACT = [
  { name: 'Aktuell', umsatz: 12400 },
  { name: 'Mit KI-Preisen', umsatz: 14850 },
]

const MAINTENANCE_PREDICTIONS = [
  { id: '1', equipment: 'Martin MAC Aura XB #3', lastMaintenance: '2025-12-15', predictedDate: '2026-04-10', confidence: 92, reason: 'Lampe hat 1.850 von 2.000 Betriebsstunden erreicht' },
  { id: '2', equipment: 'QSC K12.2 #7', lastMaintenance: '2025-11-20', predictedDate: '2026-04-05', confidence: 85, reason: 'Ventilator zeigt erhoehte Laufgeraeusche (Sensordaten)' },
  { id: '3', equipment: 'Shure ULXD4 #2', lastMaintenance: '2026-01-10', predictedDate: '2026-05-15', confidence: 78, reason: 'Batterie-Ladezyklen nahe Maximum' },
  { id: '4', equipment: 'Blackmagic ATEM 2 M/E', lastMaintenance: '2025-10-05', predictedDate: '2026-04-20', confidence: 71, reason: 'Firmware-Instabilitaeten bei letzten 3 Einsaetzen' },
]

const ACTIVITY_LOG = [
  { id: '1', timestamp: '2026-03-23 14:32', type: 'price', description: 'Preis fuer "Sennheiser EW 100 G4" auf 42 EUR/Tag optimiert' },
  { id: '2', timestamp: '2026-03-23 11:15', type: 'mail', description: 'E-Mail von EventPro GmbH automatisch Projekt "Sommerfest 2026" zugeordnet' },
  { id: '3', timestamp: '2026-03-23 09:45', type: 'maintenance', description: 'Wartungswarnung fuer "Martin MAC Aura XB #3" generiert' },
  { id: '4', timestamp: '2026-03-22 16:20', type: 'forecast', description: 'Nachfrageprognose aktualisiert: Licht-Equipment +35% im Juni erwartet' },
  { id: '5', timestamp: '2026-03-22 14:10', type: 'price', description: 'Preisvorschlag fuer 6 Geraete generiert (+19,7% Umsatzpotenzial)' },
  { id: '6', timestamp: '2026-03-22 10:30', type: 'recognition', description: 'Neues Equipment "Chauvet Rogue R3X Wash" per Foto erkannt und angelegt' },
  { id: '7', timestamp: '2026-03-21 17:00', type: 'mail', description: 'Angebot-Anfrage von TechEvents AG automatisch als Lead erfasst' },
]

// ============================================================================
// HELPER COMPONENTS
// ============================================================================

function ConfidenceBadge({ level }: { level: 'Hoch' | 'Mittel' | 'Niedrig' }) {
  const cls = level === 'Hoch' ? styles.badgeHigh : level === 'Mittel' ? styles.badgeMedium : styles.badgeLow
  return <span className={`${styles.badge} ${cls}`}>{level}</span>
}

function ActivityIcon({ type }: { type: string }) {
  const icons: Record<string, string> = {
    price: '\u20AC',
    mail: '\u2709',
    maintenance: '\u2699',
    forecast: '\u2197',
    recognition: '\u{1F4F7}',
  }
  return <span className={styles.activityIcon}>{icons[type] || '\u2022'}</span>
}

const CustomTooltip = ({ active, payload, label }: { active?: boolean; payload?: Array<{ name: string; value: number; color: string }>; label?: string }) => {
  if (active && payload && payload.length) {
    return (
      <div className={styles.chartTooltip}>
        <p className={styles.chartTooltipLabel}>{label}</p>
        {payload.map((entry, index) => (
          <p key={index} style={{ color: entry.color, margin: '2px 0', fontSize: '0.8rem' }}>
            {entry.name}: {entry.value}%
          </p>
        ))}
      </div>
    )
  }
  return null
}

const RevenueTooltip = ({ active, payload, label }: { active?: boolean; payload?: Array<{ value: number }>; label?: string }) => {
  if (active && payload && payload.length) {
    return (
      <div className={styles.chartTooltip}>
        <p className={styles.chartTooltipLabel}>{label}</p>
        <p style={{ color: '#00d4ff', margin: '2px 0', fontSize: '0.8rem' }}>
          {new Intl.NumberFormat('de-DE', { style: 'currency', currency: 'EUR' }).format(payload[0].value)}
        </p>
      </div>
    )
  }
  return null
}

// ============================================================================
// MAIN COMPONENT
// ============================================================================

function AIPage() {
  const [acceptedPrices, setAcceptedPrices] = useState<Set<string>>(new Set())

  const handleAcceptPrice = (id: string) => {
    setAcceptedPrices(prev => {
      const next = new Set(prev)
      next.add(id)
      return next
    })
  }

  const handleAcceptAll = () => {
    setAcceptedPrices(new Set(PRICE_SUGGESTIONS.map(s => s.id)))
  }

  const activeProviders = AI_PROVIDERS.filter(p => p.status === 'active').length
  const totalPotential = PRICE_SUGGESTIONS.reduce((sum, s) => sum + (s.suggestedPrice - s.currentPrice), 0)

  return (
    <div className={styles.dashboard}>
      {/* Header */}
      <div className={styles.dashboardHeader}>
        <div>
          <h1 className={styles.dashboardTitle}>KI-Dashboard</h1>
          <p className={styles.dashboardSubtitle}>
            Intelligente Analyse und Optimierung fuer Ihren Verleih
          </p>
        </div>
        <div className={styles.headerStats}>
          <div className={styles.headerStat}>
            <span className={styles.headerStatValue}>{activeProviders}</span>
            <span className={styles.headerStatLabel}>Provider aktiv</span>
          </div>
          <div className={styles.headerStat}>
            <span className={styles.headerStatValue}>{PRICE_SUGGESTIONS.length}</span>
            <span className={styles.headerStatLabel}>Preisvorschlaege</span>
          </div>
          <div className={styles.headerStat}>
            <span className={styles.headerStatValue}>{MAINTENANCE_PREDICTIONS.length}</span>
            <span className={styles.headerStatLabel}>Wartungswarnungen</span>
          </div>
        </div>
      </div>

      {/* Section 1: KI-Status */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>KI-Status</h2>
        <div className={styles.providerGrid}>
          {AI_PROVIDERS.map((provider) => (
            <div
              key={provider.name}
              className={`${styles.providerCard} ${provider.status === 'active' ? styles.providerActive : styles.providerInactive}`}
            >
              <div className={styles.providerHeader}>
                <span className={`${styles.statusDot} ${provider.status === 'active' ? styles.statusDotActive : styles.statusDotInactive}`} />
                <span className={styles.providerName}>{provider.name}</span>
              </div>
              <div className={styles.providerDetails}>
                <span className={styles.providerModel}>{provider.model}</span>
                {provider.status === 'active' ? (
                  <span className={styles.providerRequests}>{provider.requests.toLocaleString('de-DE')} Anfragen</span>
                ) : (
                  <span className={styles.providerInactiveText}>Nicht konfiguriert</span>
                )}
              </div>
              {provider.status === 'inactive' && (
                <Link to="/settings/integrations" className={styles.configureLink}>
                  Konfigurieren
                </Link>
              )}
            </div>
          ))}
        </div>
      </section>

      {/* Section 2: Preis-Optimierung */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Preis-Optimierung</h2>
          <div className={styles.sectionActions}>
            <span className={styles.potentialBadge}>
              +{totalPotential} EUR/Tag Potenzial
            </span>
            <button
              className={styles.btnAcceptAll}
              onClick={handleAcceptAll}
              disabled={acceptedPrices.size === PRICE_SUGGESTIONS.length}
            >
              Alle uebernehmen
            </button>
          </div>
        </div>

        <div className={styles.priceGrid}>
          <div className={styles.priceTableWrapper}>
            <table className={styles.priceTable}>
              <thead>
                <tr>
                  <th>Equipment</th>
                  <th>Kategorie</th>
                  <th style={{ textAlign: 'right' }}>Aktueller Preis</th>
                  <th style={{ textAlign: 'right' }}>KI-Vorschlag</th>
                  <th style={{ textAlign: 'center' }}>Konfidenz</th>
                  <th>Grund</th>
                  <th style={{ textAlign: 'center' }}>Aktion</th>
                </tr>
              </thead>
              <tbody>
                {PRICE_SUGGESTIONS.map((item) => {
                  const isAccepted = acceptedPrices.has(item.id)
                  const diff = item.suggestedPrice - item.currentPrice
                  const isPositive = diff > 0
                  return (
                    <tr key={item.id} className={isAccepted ? styles.rowAccepted : ''}>
                      <td className={styles.equipmentName}>{item.name}</td>
                      <td>
                        <span className={styles.categoryTag}>{item.category}</span>
                      </td>
                      <td style={{ textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
                        {item.currentPrice} EUR/Tag
                      </td>
                      <td style={{ textAlign: 'right', fontVariantNumeric: 'tabular-nums' }}>
                        <span className={isPositive ? styles.priceUp : styles.priceDown}>
                          {isPositive ? '\u2191' : '\u2193'} {item.suggestedPrice} EUR/Tag
                        </span>
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <ConfidenceBadge level={item.confidence} />
                      </td>
                      <td className={styles.reasonCell}>{item.reason}</td>
                      <td style={{ textAlign: 'center' }}>
                        {isAccepted ? (
                          <span className={styles.acceptedCheck}>{'\u2713'} Uebernommen</span>
                        ) : (
                          <button
                            className={styles.btnAccept}
                            onClick={() => handleAcceptPrice(item.id)}
                          >
                            Uebernehmen
                          </button>
                        )}
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          <div className={styles.revenueChart}>
            <h3 className={styles.chartTitle}>Umsatz-Impact</h3>
            <p className={styles.chartDescription}>Monatlicher Umsatz: Aktuell vs. mit KI-Preisen</p>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={REVENUE_IMPACT} barGap={12}>
                <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
                <XAxis dataKey="name" tick={{ fill: '#94a3b8', fontSize: 12 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: '#94a3b8', fontSize: 12 }} axisLine={false} tickLine={false} />
                <Tooltip content={<RevenueTooltip />} />
                <Bar dataKey="umsatz" fill="url(#revenueGradient)" radius={[6, 6, 0, 0]} barSize={60} />
                <defs>
                  <linearGradient id="revenueGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#00d4ff" stopOpacity={0.9} />
                    <stop offset="100%" stopColor="#8b5cf6" stopOpacity={0.6} />
                  </linearGradient>
                </defs>
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </section>

      {/* Section 3: Demand Forecasting */}
      <section className={styles.section}>
        <div className={styles.sectionHeader}>
          <h2 className={styles.sectionTitle}>Nachfrageprognose</h2>
          <span className={styles.forecastPeriod}>Naechste 6 Monate</span>
        </div>
        <div className={styles.forecastCard}>
          <div className={styles.forecastPeaks}>
            <div className={styles.peakBadge}>
              <span className={styles.peakIcon}>{'\u26A1'}</span>
              <div>
                <strong>Festivalsaison</strong>
                <span>Juni - Juli: Licht +95%, Audio +92%</span>
              </div>
            </div>
            <div className={styles.peakBadge}>
              <span className={styles.peakIcon}>{'\u{1F3B5}'}</span>
              <div>
                <strong>Konzertsaison</strong>
                <span>Mai - Juni: Audio-Equipment stark nachgefragt</span>
              </div>
            </div>
            <div className={styles.peakBadge}>
              <span className={styles.peakIcon}>{'\u{1F3AC}'}</span>
              <div>
                <strong>Messe-Herbst</strong>
                <span>September: Video-Equipment +72%</span>
              </div>
            </div>
          </div>
          <ResponsiveContainer width="100%" height={300}>
            <AreaChart data={DEMAND_FORECAST}>
              <defs>
                <linearGradient id="audioGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#00d4ff" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#00d4ff" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="lichtGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0} />
                </linearGradient>
                <linearGradient id="videoGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%" stopColor="#10b981" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#10b981" stopOpacity={0} />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.06)" />
              <XAxis dataKey="month" tick={{ fill: '#94a3b8', fontSize: 12 }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 12 }} axisLine={false} tickLine={false} domain={[0, 100]} unit="%" />
              <Tooltip content={<CustomTooltip />} />
              <Legend
                wrapperStyle={{ fontSize: '12px', color: '#94a3b8' }}
                iconType="circle"
                iconSize={8}
              />
              <Area type="monotone" dataKey="audio" name="Audio" stroke="#00d4ff" fill="url(#audioGrad)" strokeWidth={2} dot={{ r: 4, fill: '#00d4ff' }} />
              <Area type="monotone" dataKey="licht" name="Licht" stroke="#8b5cf6" fill="url(#lichtGrad)" strokeWidth={2} dot={{ r: 4, fill: '#8b5cf6' }} />
              <Area type="monotone" dataKey="video" name="Video" stroke="#10b981" fill="url(#videoGrad)" strokeWidth={2} dot={{ r: 4, fill: '#10b981' }} />
            </AreaChart>
          </ResponsiveContainer>
        </div>
      </section>

      {/* Section 4: Predictive Maintenance */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>Vorausschauende Wartung</h2>
        <div className={styles.maintenanceGrid}>
          {MAINTENANCE_PREDICTIONS.map((item) => (
            <div key={item.id} className={styles.maintenanceCard}>
              <div className={styles.maintenanceHeader}>
                <h4 className={styles.maintenanceEquipment}>{item.equipment}</h4>
                <div className={styles.maintenanceConfidence}>
                  <div className={styles.confidenceBar}>
                    <div
                      className={styles.confidenceFill}
                      style={{ width: `${item.confidence}%` }}
                    />
                  </div>
                  <span className={styles.confidenceValue}>{item.confidence}%</span>
                </div>
              </div>
              <div className={styles.maintenanceDetails}>
                <div className={styles.maintenanceRow}>
                  <span className={styles.maintenanceLabel}>Letzte Wartung</span>
                  <span className={styles.maintenanceValue}>
                    {new Date(item.lastMaintenance).toLocaleDateString('de-DE')}
                  </span>
                </div>
                <div className={styles.maintenanceRow}>
                  <span className={styles.maintenanceLabel}>Vorhergesagt</span>
                  <span className={`${styles.maintenanceValue} ${styles.maintenancePredicted}`}>
                    {new Date(item.predictedDate).toLocaleDateString('de-DE')}
                  </span>
                </div>
                <div className={styles.maintenanceReason}>{item.reason}</div>
              </div>
              <button className={styles.btnMaintenance}>
                Wartung planen
              </button>
            </div>
          ))}
        </div>
      </section>

      {/* Section 5: KI-Aktivitaetslog */}
      <section className={styles.section}>
        <h2 className={styles.sectionTitle}>KI-Aktivitaetslog</h2>
        <div className={styles.activityList}>
          {ACTIVITY_LOG.map((entry) => (
            <div key={entry.id} className={styles.activityItem}>
              <ActivityIcon type={entry.type} />
              <div className={styles.activityContent}>
                <span className={styles.activityDescription}>{entry.description}</span>
                <span className={styles.activityTime}>{entry.timestamp}</span>
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}

export default AIPage
