import { useRef, useState } from 'react'
import './Form.module.scss'

interface FileUploadProps {
  label?: string
  accept?: string
  multiple?: boolean
  error?: string
  onChange: (files: File[]) => void
}

export function FileUpload({
  label,
  accept,
  multiple = false,
  error,
  onChange,
}: FileUploadProps) {
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [files, setFiles] = useState<File[]>([])

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFiles = Array.from(e.target.files || [])
    setFiles(selectedFiles)
    onChange(selectedFiles)
  }

  const handleRemoveFile = (index: number) => {
    const newFiles = files.filter((_, i) => i !== index)
    setFiles(newFiles)
    onChange(newFiles)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  return (
    <div className="form-group">
      {label && <label className="form-label">{label}</label>}

      <div
        className={`file-upload ${error ? 'file-upload--error' : ''}`}
        onClick={() => fileInputRef.current?.click()}
      >
        <input
          ref={fileInputRef}
          type="file"
          multiple={multiple}
          accept={accept}
          onChange={handleFileChange}
          className="file-upload__input"
        />
        <div className="file-upload__content">
          <span className="file-upload__icon">📁</span>
          <p className="file-upload__text">
            Click to upload or drag and drop
          </p>
        </div>
      </div>

      {files.length > 0 && (
        <div className="file-upload__list">
          {files.map((file, index) => (
            <div key={`${file.name}-${index}`} className="file-upload__item">
              <span className="file-upload__filename">{file.name}</span>
              <button
                type="button"
                className="file-upload__remove"
                onClick={() => handleRemoveFile(index)}
              >
                ✕
              </button>
            </div>
          ))}
        </div>
      )}

      {error && <span className="form-error">{error}</span>}
    </div>
  )
}
