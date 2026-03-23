export function AIInsightsWidget() {
  const insights = [
    {
      id: 'i1',
      type: 'optimization',
      title: 'Equipment-Auslastung optimieren',
      description: 'Die Lichtausrüstung hat eine 35% höhere Nachfrage als Verfügbarkeit. Erwägen Sie zusätzliche Moving Heads.',
      priority: 'high',
      icon: '💡',
    },
    {
      id: 'i2',
      type: 'warning',
      title: 'Wartung überfällig',
      description: 'MA Lighting grandMA3 ist 2 Wochen über dem geplanten Wartungsintervall.',
      priority: 'critical',
      icon: '⚠️',
    },
    {
      id: 'i3',
      type: 'trend',
      title: 'Umsatztrend positiv',
      description: 'Der Umsatz ist in den letzten 3 Monaten um 18% gestiegen. Hochsaison beginnt im April.',
      priority: 'info',
      icon: '📈',
    },
    {
      id: 'i4',
      type: 'suggestion',
      title: 'Cross-Selling Möglichkeit',
      description: 'Kunden die Audio mieten, buchen zu 72% auch Lichtequipment. Bundle-Angebote empfohlen.',
      priority: 'medium',
      icon: '🎯',
    },
  ]

  const priorityClass = (p: string) => {
    switch (p) {
      case 'critical': return 'ai-insight--critical'
      case 'high': return 'ai-insight--high'
      case 'medium': return 'ai-insight--medium'
      default: return 'ai-insight--info'
    }
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
