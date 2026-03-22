import { useState, useRef, useEffect } from 'react'
import { useQuery } from '@tanstack/react-query'
import type { ChatMessage, AIProvider } from '../../types/ai'
import styles from './AI.module.scss'

const mockChatMessages: ChatMessage[] = [
  {
    id: '1',
    role: 'assistant',
    content: 'Hallo! Ich bin dein RentFlow KI-Assistent. Ich kann dir bei der Verwaltung deiner Ausrüstung, Preisoptimierung und Nachfrageprognosen helfen. Wie kann ich dir heute helfen?',
    timestamp: '2026-03-22T09:00:00Z',
  },
  {
    id: '2',
    role: 'user',
    content: 'Können Sie mir helfen, den Preis für meine Bühnenausrüstung zu optimieren?',
    timestamp: '2026-03-22T09:05:00Z',
  },
  {
    id: '3',
    role: 'assistant',
    content: 'Natürlich! Basierend auf den aktuellen Markttrends und Ihrer historischen Nachfragedaten kann ich folgende Optimierungen empfehlen:\n\n1. Bühnenlights: Erhöhung von €150/Tag auf €185/Tag\n2. Mietgestelle: Erhöhung von €45/Tag auf €55/Tag\n3. Soundanlage: Preis stabil bei €200/Tag\n\nDiese Preise basieren auf einer 78% Nachfragequote im März.',
    timestamp: '2026-03-22T09:06:00Z',
  },
]

const aiProviders: AIProvider[] = ['Claude', 'GPT-4o', 'Gemini', 'Mistral', 'Ollama']

function AIPage() {
  const [messages, setMessages] = useState<ChatMessage[]>(mockChatMessages)
  const [inputValue, setInputValue] = useState('')
  const [selectedProvider, setSelectedProvider] = useState<AIProvider>('Claude')
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const { data: _chatMessages = mockChatMessages } = useQuery({
    queryKey: ['chat-messages'],
    queryFn: async () => mockChatMessages,
    staleTime: 1000 * 60 * 5,
  })

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSendMessage = () => {
    if (inputValue.trim() === '') return

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: inputValue,
      timestamp: new Date().toISOString(),
    }

    setMessages(prev => [...prev, userMessage])
    setInputValue('')
    setIsLoading(true)

    // Simulate AI response
    setTimeout(() => {
      const assistantMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: `Ich verstehe deine Anfrage bezüglich "${inputValue}". Dies wird von ${selectedProvider} verarbeitet. Dies ist eine Demo-Antwort - in der produktiven Version würde die echte API-Integration hier stattfinden.`,
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, assistantMessage])
      setIsLoading(false)
    }, 1000)
  }

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  const handleSmartAssetCreator = () => {
    const message: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: 'Öffne Smart Asset Creator für neue Ausrüstung',
      timestamp: new Date().toISOString(),
    }
    setMessages(prev => [...prev, message])

    setTimeout(() => {
      const response: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'Smart Asset Creator wurde geöffnet. Du kannst jetzt neue Ausrüstung mit KI-unterstützten Spezifikationen erstellen.',
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, response])
    }, 500)
  }

  const handlePriceOptimizer = () => {
    const message: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: 'Analysiere meine aktuellen Preise und optimiere sie',
      timestamp: new Date().toISOString(),
    }
    setMessages(prev => [...prev, message])

    setTimeout(() => {
      const response: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'Preisanalyse abgeschlossen! Empfehlungen:\n- Kategorie A: +8% Erhöhung empfohlen\n- Kategorie B: -3% Reduktion empfohlen\n- Kategorie C: Beibehaltung',
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, response])
    }, 800)
  }

  const handleDemandForecast = () => {
    const message: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: 'Erstelle eine Nachfrageprognose für die nächsten 30 Tage',
      timestamp: new Date().toISOString(),
    }
    setMessages(prev => [...prev, message])

    setTimeout(() => {
      const response: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'Nachfrageprognose für die nächsten 30 Tage:\n📈 Woche 1: Hohe Nachfrage (85%)\n📊 Woche 2: Mittlere Nachfrage (62%)\n📉 Woche 3: Hohe Nachfrage (78%)\n📈 Woche 4: Sehr hohe Nachfrage (92%)',
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, response])
    }, 1000)
  }

  return (
    <div className={styles.aiPage}>
      <div className={styles.header}>
        <div>
          <h1 className={styles.title}>🤖 KI-Assistent</h1>
          <p className={styles.subtitle}>Chat-Interface für intelligente Geschäftsentscheidungen</p>
        </div>
      </div>

      <div className={styles.mainContainer}>
        {/* Sidebar with Quick Actions */}
        <aside className={styles.sidebar}>
          <div className={styles.providerSection}>
            <label className={styles.label}>AI Provider:</label>
            <select
              className={styles.providerSelect}
              value={selectedProvider}
              onChange={(e) => setSelectedProvider(e.target.value as AIProvider)}
            >
              {aiProviders.map(provider => (
                <option key={provider} value={provider}>{provider}</option>
              ))}
            </select>
          </div>

          <div className={styles.quickActionsSection}>
            <h3 className={styles.sectionTitle}>Quick Actions</h3>
            <button
              className={styles.quickActionButton}
              onClick={handleSmartAssetCreator}
            >
              <span className={styles.icon}>✨</span>
              <span>Smart Asset Creator</span>
            </button>
            <button
              className={styles.quickActionButton}
              onClick={handlePriceOptimizer}
            >
              <span className={styles.icon}>💰</span>
              <span>Price Optimizer</span>
            </button>
            <button
              className={styles.quickActionButton}
              onClick={handleDemandForecast}
            >
              <span className={styles.icon}>📈</span>
              <span>Demand Forecast</span>
            </button>
          </div>

          <div className={styles.infoSection}>
            <h4 className={styles.infoTitle}>Aktuelle Session</h4>
            <p className={styles.infoText}>
              <strong>Provider:</strong> {selectedProvider}
            </p>
            <p className={styles.infoText}>
              <strong>Nachrichten:</strong> {messages.length}
            </p>
            <p className={styles.infoText}>
              <strong>Status:</strong> <span className={styles.statusOnline}>●</span> Online
            </p>
          </div>
        </aside>

        {/* Chat Area */}
        <div className={styles.chatContainer}>
          <div className={styles.messagesArea}>
            {messages.map((message) => (
              <div
                key={message.id}
                className={`${styles.message} ${styles[`message--${message.role}`]}`}
              >
                <div className={styles.messageContent}>
                  {message.role === 'user' ? '👤 ' : '🤖 '}
                  {message.content.split('\n').map((line, idx) => (
                    <div key={idx}>{line}</div>
                  ))}
                </div>
                <div className={styles.messageTime}>
                  {new Date(message.timestamp).toLocaleTimeString('de-DE', {
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </div>
              </div>
            ))}
            {isLoading && (
              <div className={`${styles.message} ${styles['message--assistant']}`}>
                <div className={styles.messageContent}>
                  🤖 <span className={styles.typing}>Tiping...</span>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>

          <div className={styles.inputArea}>
            <input
              type="text"
              className={styles.input}
              placeholder="Schreibe deine Nachricht hier..."
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              disabled={isLoading}
            />
            <button
              className={styles.sendButton}
              onClick={handleSendMessage}
              disabled={isLoading || inputValue.trim() === ''}
            >
              📤 Senden
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

export default AIPage
