import './Form.scss'

interface Option {
  value: string | number
  label: string
}

interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  label?: string
  error?: string
  options: Option[]
  placeholder?: string
}

export function Select({
  label,
  error,
  options,
  placeholder,
  id,
  ...props
}: SelectProps) {
  const selectId = id || `select-${Math.random()}`

  return (
    <div className="form-group">
      {label && (
        <label htmlFor={selectId} className="form-label">
          {label}
        </label>
      )}
      <select
        id={selectId}
        className={`form-select ${error ? 'form-select--error' : ''}`}
        {...props}
      >
        {placeholder && (
          <option value="">{placeholder}</option>
        )}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {error && <span className="form-error">{error}</span>}
    </div>
  )
}
