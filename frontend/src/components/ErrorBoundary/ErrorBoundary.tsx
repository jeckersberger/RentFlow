import { Component, type ErrorInfo, type ReactNode } from 'react'
import styles from './ErrorBoundary.module.scss'

interface ErrorBoundaryProps {
  children: ReactNode
  fallback?: ReactNode
}

interface ErrorBoundaryState {
  hasError: boolean
  error: Error | null
  errorInfo: ErrorInfo | null
}

class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props)
    this.state = { hasError: false, error: null, errorInfo: null }
  }

  static getDerivedStateFromError(error: Error): Partial<ErrorBoundaryState> {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('[ErrorBoundary] Unhandled error:', error)
    console.error('[ErrorBoundary] Component stack:', errorInfo.componentStack)
    this.setState({ errorInfo })
  }

  handleReload = () => {
    window.location.reload()
  }

  handleDashboard = () => {
    window.location.href = '/'
  }

  handleReset = () => {
    this.setState({ hasError: false, error: null, errorInfo: null })
  }

  render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback
      }

      return (
        <div className={styles.container}>
          <div className={styles.card}>
            <div className={styles.icon}>!</div>
            <h1 className={styles.title}>Etwas ist schiefgelaufen</h1>
            <p className={styles.description}>
              Ein unerwarteter Fehler ist aufgetreten. Bitte versuchen Sie es erneut oder laden Sie die Seite neu.
            </p>

            {this.state.error && (
              <details className={styles.details}>
                <summary>Technische Details</summary>
                <pre className={styles.errorText}>
                  {this.state.error.message}
                  {this.state.error.stack && (
                    <>
                      {'\n\n--- Stack Trace ---\n'}
                      {this.state.error.stack}
                    </>
                  )}
                  {this.state.errorInfo?.componentStack && (
                    <>
                      {'\n\n--- Component Stack ---\n'}
                      {this.state.errorInfo.componentStack}
                    </>
                  )}
                </pre>
              </details>
            )}

            <div className={styles.actions}>
              <button className={styles.buttonPrimary} onClick={this.handleReload}>
                Seite neu laden
              </button>
              <button className={styles.buttonSecondary} onClick={this.handleDashboard}>
                Zum Dashboard
              </button>
            </div>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}

export default ErrorBoundary
