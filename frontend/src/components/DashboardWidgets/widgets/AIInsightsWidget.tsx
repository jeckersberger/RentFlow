import { useQuery } from '@tanstack/react-query'
import { api } from '../../../services/api'

interface AIInsight {
  id: string
  type: string
  title: string
  description: string
  priority: string
  icon: string
}

export function AIInsightsWidget() {
  const { data: insightsData } = useQuery<AIInsight[]>({
    queryKey: ['dashboard-ai-insights'],
    queryFn: async () => {
      try {
        const res = await api.get('/api/v1/ai/insights')
        const data = res.data?.data || res.data?.items || res.data || []
        if (Array.isArray(data)) return data
        return []
      } catch {
        return []
      }
    },
    staleTime: 1000 * 60 * 15,
  })

  const insights = insightsData || []

  const priorityClass = (p: string) => {
    switch (p) {
      case 'critical': return 'ai-insight--critical'
      case 'high': return 'ai-insight--high'
      case 'medium': return 'ai-insight--medium'
      default: return 'ai-insight--info'
    }
  }

  if (!insightsData) {
    return (
      <div className="ai-insights-widget">
        <div className="ai-insights-widget__badge">
          <span className="ai-insights-widget__badge-icon">✨</span>
          KI-Empfehlungen
        </div>
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '1rem' }}>Laden...</p>
      </div>
    )
  }

  if (insights.length === 0) {
    return (
      <div className="ai-insights-widget">
        <div className="ai-insights-widget__badge">
          <span className="ai-insights-widget__badge-icon">✨</span>
          KI-Empfehlungen
        </div>
        <p style={{ color: 'var(--color-text-muted)', textAlign: 'center', padding: '1rem' }}>
          Noch keine KI-Empfehlungen verfügbar.
        </p>
      </div>
    )
  }

  return (
    <div className="ai-insights-widget">
      <div className="ai-insights-widget__badge">
        <span className="ai-insights-widget__badge-icon">✨</span>
        KI-Empfehlungen
      </div>
      <div className="ai-insights-widget__list">
        {insights.map((insight) => (
          <div key={insight.id} className={`ai-insight ${priorityClass(insight.priority)}`}>
            <div className="ai-insight__icon">{insight.icon}</div>
            <div className="ai-insight__content">
              <p className="ai-insight__title">{insight.title}</p>
              <p className="ai-insight__description">{insight.description}</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
