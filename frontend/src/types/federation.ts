export type PartnerTrustLevel = 'verified' | 'trusted' | 'provisional' | 'suspended'
export type CertificateStatus = 'valid' | 'expiring' | 'expired' | 'pending'

export interface Partner {
  id: string
  name: string
  company: string
  email: string
  phone: string
  country: string
  trust_level: PartnerTrustLevel
  equipment_availability: number
  certificate_status: CertificateStatus
  certificate_expiry: string
  joined_date: string
  rating: number
}

export interface EquipmentAvailability {
  partner_id: string
  partner_name: string
  equipment_type: string
  quantity_available: number
  location: string
  delivery_days: number
  price_per_day: number
}

export interface SubRentalRequest {
  id: string
  partner_id: string
  partner_name: string
  equipment: string
  quantity: number
  start_date: string
  end_date: string
  status: 'pending' | 'approved' | 'rejected' | 'completed'
  requested_at: string
}

export interface FederationStats {
  total_partners: number
  verified_partners: number
  equipment_offers: number
  active_rentals: number
}
