export type DocumentType = 'offer' | 'invoice' | 'delivery_note' | 'contract' | 'other'
export type DocumentStatus = 'draft' | 'generated' | 'sent' | 'signed' | 'archived'
export type SignatureStatus = 'pending' | 'signed' | 'rejected'

export interface Document {
  id: string
  number: string
  type: DocumentType
  title: string
  status: DocumentStatus
  signature_status: SignatureStatus
  created_date: string
  sent_date?: string
  signed_date?: string
  project_id?: string
  project_name?: string
  recipient: string
  file_url?: string
  generated_by: string
}

export interface DocumentSignature {
  id: string
  document_id: string
  signer_name: string
  signer_email: string
  signed_date: string
  signature_url?: string
  status: SignatureStatus
}

export interface GoBDVerification {
  document_id: string
  hash: string
  timestamp: string
  verified: boolean
  chain_valid: boolean
}
