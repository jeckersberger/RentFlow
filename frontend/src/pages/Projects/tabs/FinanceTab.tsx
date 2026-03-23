import { Project } from '../../../types/project'
import styles from '../ProjectDetail.module.scss'

interface FinanceTabProps {
  project: Project
}

export function FinanceTab({ project }: FinanceTabProps) {
  // Derived from project data where available; everything else placeholder
  const budget = project.budget || 0

  const summaryCards = [
    { label: 'Gesamtumsatz', value: budget, cls: styles.financeCardValueRevenue },
    { label: 'Kosten', value: 0, cls: styles.financeCardValueCost },
    { label: 'Gewinn', value: budget, cls: styles.financeCardValueProfit },
    { label: 'Rabatt', value: 0, cls: styles.financeCardValueDiscount },
  ]

  const breakdownRows = [
    { category: 'Equipment', estimated: 0, planned: 0, actual: 0, revenue: 0, discount: 0, profit: 0 },
    { category: 'Crew', estimated: 0, planned: 0, actual: 0, revenue: 0, discount: 0, profit: 0 },
    { category: 'Transport', estimated: 0, planned: 0, actual: 0, revenue: 0, discount: 0, profit: 0 },
    { category: 'Zusatzkosten', estimated: 0, planned: 0, actual: 0, revenue: 0, discount: 0, profit: 0 },
    { category: 'Versicherung', estimated: 0, planned: 0, actual: 0, revenue: 0, discount: 0, profit: 0 },
  ]

  const invoicedPercent = project.status === 'invoiced' ? 100 : 0

  return (
    <div>
      {/* Summary Cards */}
      <div className={styles.financeSummary}>
        {summaryCards.map((card) => (
          <div key={card.label} className={styles.financeCard}>
            <div className={styles.financeCardLabel}>{card.label}</div>
            <div className={`${styles.financeCardValue} ${card.cls}`}>
              {'\u20AC'}{card.value.toFixed(2)}
            </div>
          </div>
        ))}
      </div>

      {/* Invoiced progress */}
      <div className={styles.glassCard} style={{ marginBottom: 'var(--spacing-6)', textAlign: 'center' }}>
        <div className={styles.glassCardTitle} style={{ justifyContent: 'center' }}>
          In Rechnung gestellt
        </div>
        <div style={{
          fontSize: 'var(--font-size-3xl)',
          fontWeight: 'var(--font-weight-bold)',
          color: invoicedPercent > 0 ? 'var(--color-success)' : 'var(--color-text-muted)',
        }}>
          {invoicedPercent}%
        </div>
      </div>

      {/* Quotes Section */}
      <div className={styles.glassCard} style={{ marginBottom: 'var(--spacing-5)' }}>
        <div className={styles.sectionHeader}>
          <div className={styles.glassCardTitle} style={{ margin: 0 }}>Angebote</div>
          <button className="btn btn--secondary" disabled>Angebot erstellen</button>
        </div>
        <div className={styles.emptyState} style={{ padding: 'var(--spacing-6)' }}>
          <p className={styles.emptyStateText}>Noch keine Angebote erstellt.</p>
        </div>
      </div>

      {/* Invoices Section */}
      <div className={styles.glassCard} style={{ marginBottom: 'var(--spacing-5)' }}>
        <div className={styles.sectionHeader}>
          <div className={styles.glassCardTitle} style={{ margin: 0 }}>Rechnungen</div>
          <button className="btn btn--secondary" disabled>Rechnung erstellen</button>
        </div>
        <div className={styles.emptyState} style={{ padding: 'var(--spacing-6)' }}>
          <p className={styles.emptyStateText}>Noch keine Rechnungen erstellt.</p>
        </div>
      </div>

      {/* Financial Breakdown */}
      <div className={styles.glassCard}>
        <div className={styles.glassCardTitle}>Finanzübersicht nach Kategorie</div>
        <div style={{ overflowX: 'auto' }}>
          <table className={styles.dataTable}>
            <thead>
              <tr>
                <th>Kategorie</th>
                <th style={{ textAlign: 'right' }}>Geschätzt</th>
                <th style={{ textAlign: 'right' }}>Geplant</th>
                <th style={{ textAlign: 'right' }}>Tatsächlich</th>
                <th style={{ textAlign: 'right' }}>Umsatz</th>
                <th style={{ textAlign: 'right' }}>Rabatt %</th>
                <th style={{ textAlign: 'right' }}>Gewinn</th>
              </tr>
            </thead>
            <tbody>
              {breakdownRows.map((row) => (
                <tr key={row.category}>
                  <td>{row.category}</td>
                  <td style={{ textAlign: 'right' }}>{'\u20AC'}{row.estimated.toFixed(2)}</td>
                  <td style={{ textAlign: 'right' }}>{'\u20AC'}{row.planned.toFixed(2)}</td>
                  <td style={{ textAlign: 'right' }}>{'\u20AC'}{row.actual.toFixed(2)}</td>
                  <td style={{ textAlign: 'right' }}>{'\u20AC'}{row.revenue.toFixed(2)}</td>
                  <td style={{ textAlign: 'right' }}>{row.discount}%</td>
                  <td style={{ textAlign: 'right' }}>{'\u20AC'}{row.profit.toFixed(2)}</td>
                </tr>
              ))}
            </tbody>
            <tfoot>
              <tr>
                <td>Gesamt</td>
                <td style={{ textAlign: 'right' }}>{'\u20AC'}0.00</td>
                <td style={{ textAlign: 'right' }}>{'\u20AC'}0.00</td>
                <td style={{ textAlign: 'right' }}>{'\u20AC'}0.00</td>
                <td style={{ textAlign: 'right' }}>{'\u20AC'}0.00</td>
                <td style={{ textAlign: 'right' }}>&mdash;</td>
                <td style={{ textAlign: 'right' }}>{'\u20AC'}0.00</td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>
  )
}
