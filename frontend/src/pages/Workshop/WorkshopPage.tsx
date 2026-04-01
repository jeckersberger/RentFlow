import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { equipmentApi, maintenanceApi } from '../../services/api'
import { SkeletonTable } from '../../components/Skeleton/SkeletonLoader'
import type { ElectricalTest, MaintenancePlan } from '../../types/maintenance'
import './Workshop.scss'

type TabKey = 'repairs' | 'inspections' | 'lost' | 'inventory'

interface Repair {
  id: string
  equipmentName: string
  defect: string
  priority: 'Hoch' | 'Mittel' | 'Niedrig'
  status: 'Offen' | 'In Arbeit' | 'Erledigt'
  assignedTo: string
  createdAt: string
}

interface Inspection {
  id: string
  equipmentName: string
  nextDate: string
  type: string
  status: 'Fällig' | 'Geplant' | 'Bestanden' | 'Überfällig'
}

interface LostItem {
  id: string
  equipmentName: string
  lastSeen: string
  project: string
  reportedBy: string
  reportedAt: string
}

const PRIORITY_MAP: Record<string, 'Hoch' | 'Mittel' | 'Niedrig'> = {
  critical: 'Hoch',
  high: 'Hoch',
  medium: 'Mittel',
  low: 'Niedrig',
}

const STATUS_MAP: Record<string, 'Offen' | 'In Arbeit' | 'Erledigt'> = {
  pending: 'Offen',
  planned: 'Offen',
  overdue: 'Offen',
  in_progress: 'In Arbeit',
  completed: 'Erledigt',
  cancelled: 'Erledigt',
}

const INTERVAL_LABELS: Record<string, string> = {
  interval: 'Intervall',
  after_use: 'Nach Einsatz',
  hours_based: 'Betriebsstunden',
}

const RESULT_LABELS: Record<string, string> = {
  passed: 'Bestanden',
  failed: 'Nicht bestanden',
  conditional: 'Bedingt bestanden',
}

const RESULT_CLASS: Record<string, string> = {
  passed: 'statusDone',
  failed: 'statusDanger',
  conditional: 'statusWarning',
}

function WorkshopPage() {
  const queryClient = useQueryClient()
  const [activeTab, setActiveTab] = useState<TabKey>('repairs')
  const [searchQuery, setSearchQuery] = useState('')
  const [showRepairModal, setShowRepairModal] = useState(false)
  const [showPlanModal, setShowPlanModal] = useState(false)
  const [showECheckModal, setShowECheckModal] = useState(false)

  // Plan form state
  const [planForm, setPlanForm] = useState({
    equipment_id: '',
    plan_type: 'interval' as 'interval' | 'after_use' | 'hours_based',
    interval_days: '',
    name: '',
    description: '',
    next_due_at: '',
  })

  // E-Check form state
  const [eCheckForm, setECheckForm] = useState({
    equipment_id: '',
    tester_id: '',
    test_date: new Date().toISOString().slice(0, 10),
    test_type: 'vde_0702' as 'vde_0701' | 'vde_0702',
    result: 'passed' as 'passed' | 'failed' | 'conditional',
    insulation_resistance_mohm: '',
    protective_conductor_resistance_ohm: '',
    leakage_current_ma: '',
    visual_inspection_ok: true,
    functional_test_ok: true,
    notes: '',
  })

  // Data queries
  const { data: equipmentData, isLoading: eqLoading } = useQuery({
    queryKey: ['equipment-workshop'],
    queryFn: () => equipmentApi.list({ limit: 100 }),
    staleTime: 1000 * 60 * 5,
  })

  const { data: tasksData, isLoading: tasksLoading } = useQuery({
    queryKey: ['maintenance-tasks-workshop'],
    queryFn: () => maintenanceApi.listTasks(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: plansData, isLoading: plansLoading } = useQuery({
    queryKey: ['maintenance-plans-workshop'],
    queryFn: () => maintenanceApi.listPlans(),
    staleTime: 1000 * 60 * 5,
  })

  const { data: eTestsData, isLoading: eTestsLoading } = useQuery({
    queryKey: ['electrical-tests-workshop'],
    queryFn: () => maintenanceApi.listElectricalTests(),
    staleTime: 1000 * 60 * 5,
  })

  const isLoading = eqLoading || tasksLoading || plansLoading || eTestsLoading

  const equipment = equipmentData?.data || equipmentData?.items || (Array.isArray(equipmentData) ? equipmentData : [])
  const tasks = tasksData?.items || tasksData?.data || (Array.isArray(tasksData) ? tasksData : [])
  const plans: MaintenancePlan[] = plansData?.items || plansData?.data || (Array.isArray(plansData) ? plansData : [])
  const electricalTests: ElectricalTest[] = eTestsData?.items || eTestsData?.data || (Array.isArray(eTestsData) ? eTestsData : [])

  // Mutations
  const { mutate: createPlan, isPending: isCreatingPlan } = useMutation({
    mutationFn: (data: object) => maintenanceApi.createPlan(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['maintenance-plans-workshop'] })
      setShowPlanModal(false)
      setPlanForm({ equipment_id: '', plan_type: 'interval', interval_days: '', name: '', description: '', next_due_at: '' })
    },
  })

  const { mutate: createECheck, isPending: isCreatingECheck } = useMutation({
    mutationFn: (data: object) => maintenanceApi.createElectricalTest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['electrical-tests-workshop'] })
      setShowECheckModal(false)
      setECheckForm({
        equipment_id: '', tester_id: '', test_date: new Date().toISOString().slice(0, 10),
        test_type: 'vde_0702', result: 'passed', insulation_resistance_mohm: '',
        protective_conductor_resistance_ohm: '', leakage_current_ma: '',
        visual_inspection_ok: true, functional_test_ok: true, notes: '',
      })
    },
  })

  // Derived data
  const repairs: Repair[] = useMemo(() => {
    return tasks.map((t: any) => ({
      id: t.id,
      equipmentName: t.equipment_name || 'Unbekannt',
      defect: t.plan_name || t.notes || 'Wartungsaufgabe',
      priority: PRIORITY_MAP[t.priority] || 'Mittel',
      status: STATUS_MAP[t.status] || 'Offen',
      assignedTo: t.assigned_to || '',
      createdAt: t.created_at || t.scheduled_at || '',
    }))
  }, [tasks])

  const inspections: Inspection[] = useMemo(() => {
    const now = new Date()
    return plans.map((p: any) => {
      const nextDate = p.next_due_at || p.next_maintenance || ''
      const isActive = p.is_active ?? (p.status === 'active')
      let status: Inspection['status'] = 'Geplant'
      if (nextDate) {
        const due = new Date(nextDate)
        const daysUntil = Math.ceil((due.getTime() - now.getTime()) / (1000 * 60 * 60 * 24))
        if (daysUntil < 0) status = 'Überfällig'
        else if (daysUntil <= 7) status = 'Fällig'
        else status = 'Geplant'
      }
      if (!isActive && p.status === 'completed') status = 'Bestanden'

      return {
        id: p.id,
        equipmentName: p.equipment_name || p.name || 'Unbekannt',
        nextDate: nextDate,
        type: p.plan_type || p.frequency || 'Wartung',
        status,
      }
    })
  }, [plans])

  const lostItems: LostItem[] = useMemo(() => {
    return equipment
      .filter((e: any) => e.status === 'lost' || e.status === 'missing')
      .map((e: any) => ({
        id: e.id,
        equipmentName: e.name,
        lastSeen: e.location || '',
        project: e.project_name || '',
        reportedBy: '',
        reportedAt: e.updated_at || e.created_at || '',
      }))
  }, [equipment])

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    return isNaN(date.getTime()) ? 'N/A' : date.toLocaleDateString('de-DE', { day: '2-digit', month: '2-digit', year: '2-digit' })
  }

  const tabs: { key: TabKey; label: string; count: number }[] = [
    { key: 'repairs', label: 'Reparaturen', count: repairs.filter(r => r.status !== 'Erledigt').length },
    { key: 'inspections', label: 'Prüfungen', count: inspections.filter(i => i.status === 'Fällig' || i.status === 'Überfällig').length },
    { key: 'lost', label: 'Verlorene Materialien', count: lostItems.length },
    { key: 'inventory', label: 'Bestandszählung', count: 0 },
  ]

  const getPriorityClass = (priority: string) => {
    switch (priority) {
      case 'Hoch': return 'priorityHigh'
      case 'Mittel': return 'priorityMedium'
      case 'Niedrig': return 'priorityLow'
      default: return 'priorityMedium'
    }
  }

  const getStatusClass = (status: string) => {
    switch (status) {
      case 'Offen': return 'statusOpen'
      case 'In Arbeit': return 'statusProgress'
      case 'Erledigt': return 'statusDone'
      case 'Fällig': return 'statusWarning'
      case 'Überfällig': return 'statusDanger'
      case 'Geplant': return 'statusPlanned'
      case 'Bestanden': return 'statusDone'
      default: return ''
    }
  }

  const handleCreatePlan = () => {
    const intervalMap: Record<string, number> = {
      daily: 1, weekly: 7, monthly: 30, yearly: 365,
    }
    const intervalDays = planForm.plan_type === 'interval'
      ? (intervalMap[planForm.interval_days] || parseInt(planForm.interval_days) || 30)
      : undefined

    createPlan({
      equipment_id: planForm.equipment_id,
      plan_type: planForm.plan_type,
      interval_days: intervalDays,
      name: planForm.name,
      description: planForm.description,
    })
  }

  const handleCreateECheck = () => {
    createECheck({
      equipment_id: eCheckForm.equipment_id,
      tester_id: eCheckForm.tester_id,
      test_type: eCheckForm.test_type,
      test_date: new Date(eCheckForm.test_date).toISOString(),
      result: eCheckForm.result,
      insulation_resistance_mohm: eCheckForm.insulation_resistance_mohm ? parseFloat(eCheckForm.insulation_resistance_mohm) : undefined,
      protective_conductor_resistance_ohm: eCheckForm.protective_conductor_resistance_ohm ? parseFloat(eCheckForm.protective_conductor_resistance_ohm) : undefined,
      leakage_current_ma: eCheckForm.leakage_current_ma ? parseFloat(eCheckForm.leakage_current_ma) : undefined,
      visual_inspection_ok: eCheckForm.visual_inspection_ok,
      functional_test_ok: eCheckForm.functional_test_ok,
      notes: eCheckForm.notes,
    })
  }

  const handlePrintLabel = (test: ElectricalTest) => {
    const eqName = test.equipment_name || equipment.find((e: any) => e.id === test.equipment_id)?.name || 'Unbekannt'
    const printWindow = window.open('', '_blank', 'width=400,height=350')
    if (!printWindow) return
    printWindow.document.write(`<!DOCTYPE html>
<html><head><title>Prüfplakette</title>
<style>
  body { font-family: Arial, sans-serif; margin: 0; padding: 20px; }
  .label { width: 70mm; height: 35mm; border: 2px solid #333; border-radius: 6px; padding: 8px 12px; box-sizing: border-box; position: relative; }
  .label-title { font-size: 11px; font-weight: bold; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 4px; }
  .label-result { font-size: 14px; font-weight: bold; color: ${test.result === 'passed' ? '#16a34a' : test.result === 'failed' ? '#dc2626' : '#d97706'}; }
  .label-row { font-size: 9px; color: #333; margin: 2px 0; }
  .label-cert { font-family: monospace; font-size: 8px; color: #666; margin-top: 4px; }
  @media print { body { padding: 0; } }
</style></head><body>
<div class="label">
  <div class="label-title">DGUV V3 / E-Check</div>
  <div class="label-result">${RESULT_LABELS[test.result] || test.result}</div>
  <div class="label-row"><strong>${eqName}</strong></div>
  <div class="label-row">Prüfdatum: ${new Date(test.test_date).toLocaleDateString('de-DE')}</div>
  <div class="label-row">Nächste Prüfung: ${test.next_test_date ? new Date(test.next_test_date).toLocaleDateString('de-DE') : '—'}</div>
  <div class="label-row">Prüfer: ${test.tester_id}</div>
  <div class="label-cert">${test.certificate_number || ''}</div>
</div>
<script>window.onload = function() { window.print(); };</script>
</body></html>`)
    printWindow.document.close()
  }

  const openRepairs = repairs.filter(r => r.status === 'Offen').length
  const inProgressRepairs = repairs.filter(r => r.status === 'In Arbeit').length
  const overdueInspections = inspections.filter(i => i.status === 'Überfällig').length

  return (
    <div className="page">
      <div className="header">
        <div>
          <h1 className="title">Werkstatt</h1>
          <p className="subtitle">Reparaturen, Prüfungen und Bestandsverwaltung</p>
        </div>
        <div className="headerActions">
          <button className="btnPrimary" onClick={() => setShowRepairModal(true)}>
            + Reparatur melden
          </button>
          <button className="btnSecondary" onClick={() => { setActiveTab('inspections'); setShowPlanModal(true) }}>
            + Wartungsplan erstellen
          </button>
          <button className="btnSecondary" onClick={() => { setActiveTab('inspections'); setShowECheckModal(true) }}>
            + E-Check erstellen
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="statsGrid">
        <div className={`statCard ${openRepairs > 0 ? 'statCardDanger' : ''}`}>
          <div className="statLabel">Offene Reparaturen</div>
          <div className="statValue">{openRepairs}</div>
        </div>
        <div className={`statCard ${inProgressRepairs > 0 ? 'statCardWarning' : ''}`}>
          <div className="statLabel">In Arbeit</div>
          <div className="statValue">{inProgressRepairs}</div>
        </div>
        <div className={`statCard ${overdueInspections > 0 ? 'statCardDanger' : ''}`}>
          <div className="statLabel">Überfällige Prüfungen</div>
          <div className="statValue">{overdueInspections}</div>
        </div>
        <div className="statCard">
          <div className="statLabel">E-Checks gesamt</div>
          <div className="statValue">{electricalTests.length}</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="tabBar">
        {tabs.map(tab => (
          <button
            key={tab.key}
            className={`tab ${activeTab === tab.key ? 'tabActive' : ''}`}
            onClick={() => setActiveTab(tab.key)}
          >
            {tab.label}
            {tab.count > 0 && <span className="tabBadge">{tab.count}</span>}
          </button>
        ))}
      </div>

      {/* Filter */}
      <div className="filterBar">
        <input
          type="text"
          className="searchInput"
          placeholder="Suchen..."
          value={searchQuery}
          onChange={e => setSearchQuery(e.target.value)}
        />
      </div>

      {/* Content */}
      <div className="tableCard">
        {isLoading ? (
          <SkeletonTable rows={5} columns={5} />
        ) : activeTab === 'repairs' ? (
          repairs.length === 0 ? (
            <div className="emptyState">
              <div className="emptyIcon">🔧</div>
              <h3 className="emptyTitle">Keine Reparaturen</h3>
              <p className="emptyDescription">Alle Geräte sind in einwandfreiem Zustand.</p>
              <button className="btnPrimary" onClick={() => setShowRepairModal(true)}>Reparatur melden</button>
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Equipment</th>
                  <th>Defektbeschreibung</th>
                  <th>Priorität</th>
                  <th>Status</th>
                  <th>Zugewiesen an</th>
                  <th>Erstellt</th>
                </tr>
              </thead>
              <tbody>
                {repairs.filter(r => r.equipmentName.toLowerCase().includes(searchQuery.toLowerCase())).map(repair => (
                  <tr key={repair.id}>
                    <td className="equipmentName">{repair.equipmentName}</td>
                    <td>{repair.defect}</td>
                    <td><span className={`badge ${getPriorityClass(repair.priority)}`}>{repair.priority}</span></td>
                    <td><span className={`badge ${getStatusClass(repair.status)}`}>{repair.status}</span></td>
                    <td className="assignee">{repair.assignedTo}</td>
                    <td className="dateCell">{formatDate(repair.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        ) : activeTab === 'inspections' ? (
          <div>
            {/* Wartungspläne */}
            <div className="sectionHeader">
              <h3 className="sectionTitle">Wartungspläne</h3>
              <button className="btnSmallPrimary" onClick={() => setShowPlanModal(true)}>
                + Wartungsplan erstellen
              </button>
            </div>
            {plans.length === 0 ? (
              <div className="emptyStateCompact">
                <p className="emptyDescription">Noch keine Wartungspläne angelegt.</p>
              </div>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Equipment</th>
                    <th>Name</th>
                    <th>Intervall</th>
                    <th>Nächste Fälligkeit</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {inspections.filter(i => i.equipmentName.toLowerCase().includes(searchQuery.toLowerCase())).map(insp => (
                    <tr key={insp.id}>
                      <td className="equipmentName">{insp.equipmentName}</td>
                      <td>{plans.find((p: any) => p.id === insp.id)?.name || '—'}</td>
                      <td><span className="typeBadge">{INTERVAL_LABELS[insp.type] || insp.type}</span></td>
                      <td className="dateCell">{formatDate(insp.nextDate)}</td>
                      <td><span className={`badge ${getStatusClass(insp.status)}`}>{insp.status}</span></td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}

            {/* E-Check / DGUV Prüfungen */}
            <div className="sectionHeader" style={{ marginTop: 'var(--spacing-6)' }}>
              <h3 className="sectionTitle">E-Check / DGUV V3 Prüfungen</h3>
              <button className="btnSmallPrimary" onClick={() => setShowECheckModal(true)}>
                + E-Check erstellen
              </button>
            </div>
            {electricalTests.length === 0 ? (
              <div className="emptyStateCompact">
                <p className="emptyDescription">Noch keine E-Checks durchgeführt.</p>
              </div>
            ) : (
              <table className="table">
                <thead>
                  <tr>
                    <th>Equipment</th>
                    <th>Prüfdatum</th>
                    <th>Nächste Prüfung</th>
                    <th>Prüfer</th>
                    <th>Ergebnis</th>
                    <th>Messwerte</th>
                    <th>Aktion</th>
                  </tr>
                </thead>
                <tbody>
                  {electricalTests
                    .filter(t => (t.equipment_name || '').toLowerCase().includes(searchQuery.toLowerCase()))
                    .map((test: ElectricalTest) => (
                    <tr key={test.id}>
                      <td className="equipmentName">{test.equipment_name || 'Unbekannt'}</td>
                      <td className="dateCell">{formatDate(test.test_date)}</td>
                      <td className="dateCell">{test.next_test_date ? formatDate(test.next_test_date) : '—'}</td>
                      <td>{test.tester_id}</td>
                      <td>
                        <span className={`badge ${RESULT_CLASS[test.result] || ''}`}>
                          {RESULT_LABELS[test.result] || test.result}
                        </span>
                      </td>
                      <td className="measurementCell">
                        {test.insulation_resistance_mohm != null && (
                          <span className="measurement">Iso: {test.insulation_resistance_mohm} MOhm</span>
                        )}
                        {test.protective_conductor_resistance_ohm != null && (
                          <span className="measurement">SL: {test.protective_conductor_resistance_ohm} Ohm</span>
                        )}
                        {test.leakage_current_ma != null && (
                          <span className="measurement">Abl: {test.leakage_current_ma} mA</span>
                        )}
                        {test.insulation_resistance_mohm == null && test.protective_conductor_resistance_ohm == null && test.leakage_current_ma == null && '—'}
                      </td>
                      <td>
                        <button
                          className="btnSmallSecondary"
                          onClick={() => handlePrintLabel(test)}
                          title="Prüfplakette drucken"
                        >
                          Plakette
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        ) : activeTab === 'lost' ? (
          lostItems.length === 0 ? (
            <div className="emptyState">
              <div className="emptyIcon">📦</div>
              <h3 className="emptyTitle">Keine verlorenen Materialien</h3>
              <p className="emptyDescription">Alle Materialien sind erfasst.</p>
            </div>
          ) : (
            <table className="table">
              <thead>
                <tr>
                  <th>Equipment</th>
                  <th>Zuletzt gesehen</th>
                  <th>Projekt</th>
                  <th>Gemeldet von</th>
                  <th>Gemeldet am</th>
                </tr>
              </thead>
              <tbody>
                {lostItems.map(item => (
                  <tr key={item.id}>
                    <td className="equipmentName">{item.equipmentName}</td>
                    <td>{item.lastSeen}</td>
                    <td><span className="projectTag">{item.project}</span></td>
                    <td>{item.reportedBy}</td>
                    <td className="dateCell">{formatDate(item.reportedAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          )
        ) : (
          <div className="emptyState">
            <div className="emptyIcon">📊</div>
            <h3 className="emptyTitle">Bestandszählung</h3>
            <p className="emptyDescription">
              Starten Sie eine neue Bestandszählung, um Ihren Lagerbestand abzugleichen.
            </p>
            <button className="btnPrimary">Zählung starten</button>
          </div>
        )}
      </div>

      {/* Repair Modal */}
      {showRepairModal && (
        <div className="modalOverlay" onClick={() => setShowRepairModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modalHeader">
              <h2 className="modalTitle">Reparatur melden</h2>
              <button className="modalClose" onClick={() => setShowRepairModal(false)}>&times;</button>
            </div>
            <div className="modalBody">
              <div className="formGroup">
                <label className="formLabel">Equipment</label>
                <select className="formSelect">
                  <option value="">Equipment auswählen...</option>
                  {equipment.map((e: any) => (
                    <option key={e.id} value={e.id}>{e.name}</option>
                  ))}
                </select>
              </div>
              <div className="formGroup">
                <label className="formLabel">Defektbeschreibung</label>
                <textarea className="formTextarea" rows={3} placeholder="Beschreiben Sie den Defekt..." />
              </div>
              <div className="formGroup">
                <label className="formLabel">Priorität</label>
                <select className="formSelect">
                  <option value="Niedrig">Niedrig</option>
                  <option value="Mittel">Mittel</option>
                  <option value="Hoch">Hoch</option>
                </select>
              </div>
              <div className="formGroup">
                <label className="formLabel">Zuweisen an</label>
                <input type="text" className="formInput" placeholder="Mitarbeitername..." />
              </div>
            </div>
            <div className="modalFooter">
              <button className="btnSecondary" onClick={() => setShowRepairModal(false)}>Abbrechen</button>
              <button className="btnPrimary" onClick={() => setShowRepairModal(false)}>Reparatur erstellen</button>
            </div>
          </div>
        </div>
      )}

      {/* Wartungsplan Modal */}
      {showPlanModal && (
        <div className="modalOverlay" onClick={() => setShowPlanModal(false)}>
          <div className="modal" onClick={e => e.stopPropagation()}>
            <div className="modalHeader">
              <h2 className="modalTitle">Wartungsplan erstellen</h2>
              <button className="modalClose" onClick={() => setShowPlanModal(false)}>&times;</button>
            </div>
            <div className="modalBody">
              <div className="formGroup">
                <label className="formLabel">Equipment *</label>
                <select
                  className="formSelect"
                  value={planForm.equipment_id}
                  onChange={e => setPlanForm(f => ({ ...f, equipment_id: e.target.value }))}
                >
                  <option value="">Equipment auswählen...</option>
                  {equipment.map((e: any) => (
                    <option key={e.id} value={e.id}>{e.name}</option>
                  ))}
                </select>
              </div>
              <div className="formGroup">
                <label className="formLabel">Name *</label>
                <input
                  type="text"
                  className="formInput"
                  placeholder="z.B. Jährliche DGUV V3 Prüfung"
                  value={planForm.name}
                  onChange={e => setPlanForm(f => ({ ...f, name: e.target.value }))}
                />
              </div>
              <div className="formGroup">
                <label className="formLabel">Plantyp</label>
                <select
                  className="formSelect"
                  value={planForm.plan_type}
                  onChange={e => setPlanForm(f => ({ ...f, plan_type: e.target.value as any }))}
                >
                  <option value="interval">Intervall (zeitbasiert)</option>
                  <option value="after_use">Nach Einsatz</option>
                  <option value="hours_based">Betriebsstunden</option>
                </select>
              </div>
              {planForm.plan_type === 'interval' && (
                <div className="formGroup">
                  <label className="formLabel">Intervall</label>
                  <select
                    className="formSelect"
                    value={planForm.interval_days}
                    onChange={e => setPlanForm(f => ({ ...f, interval_days: e.target.value }))}
                  >
                    <option value="">Intervall auswählen...</option>
                    <option value="daily">Täglich</option>
                    <option value="weekly">Wöchentlich</option>
                    <option value="monthly">Monatlich</option>
                    <option value="yearly">Jährlich</option>
                  </select>
                </div>
              )}
              <div className="formGroup">
                <label className="formLabel">Nächste Fälligkeit</label>
                <input
                  type="date"
                  className="formInput"
                  value={planForm.next_due_at}
                  onChange={e => setPlanForm(f => ({ ...f, next_due_at: e.target.value }))}
                />
              </div>
              <div className="formGroup">
                <label className="formLabel">Beschreibung / Checkliste</label>
                <textarea
                  className="formTextarea"
                  rows={4}
                  placeholder="Prüfschritte, Checkliste, Hinweise..."
                  value={planForm.description}
                  onChange={e => setPlanForm(f => ({ ...f, description: e.target.value }))}
                />
              </div>
            </div>
            <div className="modalFooter">
              <button className="btnSecondary" onClick={() => setShowPlanModal(false)}>Abbrechen</button>
              <button
                className="btnPrimary"
                onClick={handleCreatePlan}
                disabled={!planForm.equipment_id || !planForm.name || isCreatingPlan}
              >
                {isCreatingPlan ? 'Wird erstellt...' : 'Wartungsplan erstellen'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* E-Check Modal */}
      {showECheckModal && (
        <div className="modalOverlay" onClick={() => setShowECheckModal(false)}>
          <div className="modalWide" onClick={e => e.stopPropagation()}>
            <div className="modalHeader">
              <h2 className="modalTitle">E-Check / DGUV V3 Prüfung</h2>
              <button className="modalClose" onClick={() => setShowECheckModal(false)}>&times;</button>
            </div>
            <div className="modalBody">
              <div className="formRow">
                <div className="formGroup">
                  <label className="formLabel">Equipment *</label>
                  <select
                    className="formSelect"
                    value={eCheckForm.equipment_id}
                    onChange={e => setECheckForm(f => ({ ...f, equipment_id: e.target.value }))}
                  >
                    <option value="">Equipment auswählen...</option>
                    {equipment.map((e: any) => (
                      <option key={e.id} value={e.id}>{e.name}</option>
                    ))}
                  </select>
                </div>
                <div className="formGroup">
                  <label className="formLabel">Prüfer *</label>
                  <input
                    type="text"
                    className="formInput"
                    placeholder="Name des Prüfers"
                    value={eCheckForm.tester_id}
                    onChange={e => setECheckForm(f => ({ ...f, tester_id: e.target.value }))}
                  />
                </div>
              </div>

              <div className="formRow">
                <div className="formGroup">
                  <label className="formLabel">Prüfdatum *</label>
                  <input
                    type="date"
                    className="formInput"
                    value={eCheckForm.test_date}
                    onChange={e => setECheckForm(f => ({ ...f, test_date: e.target.value }))}
                  />
                </div>
                <div className="formGroup">
                  <label className="formLabel">Prüfnorm</label>
                  <select
                    className="formSelect"
                    value={eCheckForm.test_type}
                    onChange={e => setECheckForm(f => ({ ...f, test_type: e.target.value as any }))}
                  >
                    <option value="vde_0701">VDE 0701 (Instandsetzung)</option>
                    <option value="vde_0702">VDE 0702 (Wiederholungsprüfung)</option>
                  </select>
                </div>
              </div>

              <div className="formGroup">
                <label className="formLabel">Ergebnis *</label>
                <div className="radioGroup">
                  {([['passed', 'Bestanden'], ['conditional', 'Bedingt bestanden'], ['failed', 'Nicht bestanden']] as const).map(([val, label]) => (
                    <label key={val} className={`radioLabel ${eCheckForm.result === val ? 'radioLabelActive' : ''}`}>
                      <input
                        type="radio"
                        name="echeck-result"
                        value={val}
                        checked={eCheckForm.result === val}
                        onChange={() => setECheckForm(f => ({ ...f, result: val }))}
                        className="radioInput"
                      />
                      {label}
                    </label>
                  ))}
                </div>
              </div>

              <div className="formRow">
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={eCheckForm.visual_inspection_ok}
                    onChange={e => setECheckForm(f => ({ ...f, visual_inspection_ok: e.target.checked }))}
                    className="checkboxInput"
                  />
                  Sichtprüfung bestanden
                </label>
                <label className="checkboxLabel">
                  <input
                    type="checkbox"
                    checked={eCheckForm.functional_test_ok}
                    onChange={e => setECheckForm(f => ({ ...f, functional_test_ok: e.target.checked }))}
                    className="checkboxInput"
                  />
                  Funktionsprüfung bestanden
                </label>
              </div>

              <div className="formSectionLabel">Messwerte (optional)</div>
              <div className="formRow">
                <div className="formGroup">
                  <label className="formLabel">Isolationswiderstand (MOhm)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="formInput"
                    placeholder="z.B. 2.5"
                    value={eCheckForm.insulation_resistance_mohm}
                    onChange={e => setECheckForm(f => ({ ...f, insulation_resistance_mohm: e.target.value }))}
                  />
                </div>
                <div className="formGroup">
                  <label className="formLabel">Schutzleiterwiderstand (Ohm)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="formInput"
                    placeholder="z.B. 0.15"
                    value={eCheckForm.protective_conductor_resistance_ohm}
                    onChange={e => setECheckForm(f => ({ ...f, protective_conductor_resistance_ohm: e.target.value }))}
                  />
                </div>
                <div className="formGroup">
                  <label className="formLabel">Ableitstrom (mA)</label>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    className="formInput"
                    placeholder="z.B. 1.2"
                    value={eCheckForm.leakage_current_ma}
                    onChange={e => setECheckForm(f => ({ ...f, leakage_current_ma: e.target.value }))}
                  />
                </div>
              </div>

              <div className="formGroup">
                <label className="formLabel">Bemerkungen</label>
                <textarea
                  className="formTextarea"
                  rows={3}
                  placeholder="Auffälligkeiten, Mängel, Hinweise..."
                  value={eCheckForm.notes}
                  onChange={e => setECheckForm(f => ({ ...f, notes: e.target.value }))}
                />
              </div>
            </div>
            <div className="modalFooter">
              <button className="btnSecondary" onClick={() => setShowECheckModal(false)}>Abbrechen</button>
              <button
                className="btnPrimary"
                onClick={handleCreateECheck}
                disabled={!eCheckForm.equipment_id || !eCheckForm.tester_id || !eCheckForm.test_date || isCreatingECheck}
              >
                {isCreatingECheck ? 'Wird gespeichert...' : 'E-Check speichern'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

export default WorkshopPage
