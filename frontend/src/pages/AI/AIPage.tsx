import { useState, useRef, useEffect } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { aiApi } from '../../services/api'
import type { ChatMessage, AIProvider } from '../../types/ai'
import styles from './AI.module.scss'

function AIPage() {
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [inputValue, setInputValue] = useState('')
  const [selectedProvider, setSelectedProvider] = useState<AIProvider>('Claude')
  const [isLoading, setIsLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)

  // Fetch available AI providers
  const { data: aiProviders = ['Claude', 'GPT-4o', 'Gemini', 'Mistral', 'Ollama'] } = useQuery({
    queryKey: ['ai-providers'],
    queryFn: () => aiApi.providers(),
    staleTime: 1000 * 60 * 30,
  })

  // Fetch chat history
  const { data: chatHistory = [] } = useQuery({
    queryKey: ['ai-chat-history'],
    queryFn: () => aiApi.history(),
    staleTime: 1000 * 60 * 5,
  })

  // Initialize messages from history
  useEffect(() => {
    if (chatHistory.length > 0) {
      setMessages(chatHistory)
    }
  }, [chatHistory])

  // Chat mutation
  const chatMutation = useMutation({
    mutationFn: (message: string) => aiApi.chat(message, selectedProvider),
    onSuccess: (response) => {
      const assistantMessage: ChatMessage = {
        id: String(Date.now()),
        role: 'assistant',
        content: response.content || response,
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, assistantMessage])
      setIsLoading(false)
    },
    onError: () => {
      const errorMessage: ChatMessage = {
        id: String(Date.now()),
        role: 'assistant',
        content: 'Es tut mir leid, es gab einen Fehler bei der Verarbeitung Ihrer Anfrage. Bitte versuchen Sie es später erneut.',
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, errorMessage])
      setIsLoading(false)
    },
  })

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  const handleSendMessage = () => {
    if (inputValue.trim() === '' || isLoading) return

    const userMessage: ChatMessage = {
      id: String(Date.now()),
      role: 'user',
      content: inputValue,
      timestamp: new Date().toISOString(),
    }

    setMessages(prev => [...prev, userMessage])
    const messageToSend = inputValue
    setInputValue('')
    setIsLoading(true)

    // Send message to AI API
    chatMutation.mutate(messageToSend)
  }

  const handleKeyPress = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSendMessage()
    }
  }

  const handleSmartAssetCreator = () => {
    handleSendMessage_Internal('Öffne Smart Asset Creator für neue Ausrüstung')
  }

  const handlePriceOptimizer = () => {
    handleSendMessage_Internal('Analysiere meine aktuellen Preise und optimiere sie')
  }

  const handleDemandForecast = () => {
    handleSendMessage_Internal('Erstelle eine Nachfrageprognose für die nächsten 30 Tage')
  }

  const handleSendMessage_Internal = (message: string) => {
    const userMessage: ChatMessage = {
      id: String(Date.now()),
      role: 'user',
      content: message,
      timestamp: new Date().toISOString(),
    }
    setMessages(prev => [...prev, userMessage])
    setIsLoading(true)
    chatMutation.mutate(message)
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
              {aiProviders.map((provider: string) => (
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
