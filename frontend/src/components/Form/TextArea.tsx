import './Form.module.scss'

interface TextAreaProps
  extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string
  error?: string
  helperText?: string
}

export function TextArea({
  label,
  error,
  helperText,
  id,
  ...props
}: TextAreaProps) {
  const textareaId = id || `textarea-${Math.random()}`

  return (
    <div className="form-group">
      {label && (
        <label htmlFor={textareaId} className="form-label">
          {label}
        </label>
      )}
      <textarea
        id={textareaId}
        className={`form-textarea ${error ? 'form-textarea--error' : ''}`}
        {...props}
      />
      {error && <span className="form-error">{error}</span>}
      {helperText && <span className="form-helper">{helperText}</span>}
    </div>
  )
}
