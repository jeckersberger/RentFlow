/**
 * Centralized status label translations for use across all pages.
 * Provides domain-specific helper functions and a general-purpose lookup.
 */

// ---------------------------------------------------------------------------
// Domain-specific translation helpers
// ---------------------------------------------------------------------------

/** Equipment status */
export const equipmentStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'available': 'Verfügbar',
    'reserved': 'Reserviert',
    'rented': 'Vermietet',
    'checked_out': 'Vermietet',
    'maintenance': 'In Wartung',
    'in_maintenance': 'In Wartung',
    'damaged': 'Beschädigt',
    'retired': 'Ausgemustert',
    'lost': 'Verloren',
    'missing': 'Vermisst',
    'in_use': 'In Gebrauch',
  }
  return map[status?.toLowerCase()] || status
}

/** Project status */
export const projectStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'draft': 'Entwurf',
    'quote': 'Angebot',
    'quoted': 'Angebot',
    'planning': 'In Planung',
    'confirmed': 'Bestätigt',
    'active': 'In Bearbeitung',
    'in_progress': 'In Bearbeitung',
    'completed': 'Abgeschlossen',
    'cancelled': 'Storniert',
    'canceled': 'Storniert',
    'invoiced': 'Abgerechnet',
    'archived': 'Archiviert',
  }
  return map[status?.toLowerCase()] || status
}

/** Invoice status */
export const invoiceStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'draft': 'Entwurf',
    'sent': 'Gesendet',
    'paid': 'Bezahlt',
    'overdue': 'Überfällig',
    'cancelled': 'Storniert',
    'partial': 'Teilweise bezahlt',
    'partially_paid': 'Teilweise bezahlt',
    'credited': 'Gutgeschrieben',
  }
  return map[status?.toLowerCase()] || status
}

/** Crew / Assignment status */
export const assignmentStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'planned': 'Geplant',
    'scheduled': 'Geplant',
    'confirmed': 'Bestätigt',
    'active': 'Aktiv',
    'in_progress': 'Laufend',
    'completed': 'Abgeschlossen',
    'cancelled': 'Storniert',
  }
  return map[status?.toLowerCase()] || status
}

/** Crew availability */
export const availabilityLabel = (status: string): string => {
  const map: Record<string, string> = {
    'available': 'Verfügbar',
    'busy': 'Beschäftigt',
    'on_leave': 'Urlaub',
    'sick': 'Krank',
  }
  return map[status?.toLowerCase()] || status
}

/** Equipment condition */
export const conditionLabel = (condition: string): string => {
  const map: Record<string, string> = {
    'new': 'Neu',
    'excellent': 'Sehr gut',
    'very_good': 'Sehr gut',
    'good': 'Gut',
    'fair': 'Befriedigend',
    'poor': 'Mangelhaft',
    'defective': 'Defekt',
  }
  return map[condition?.toLowerCase()] || condition
}

/** Booking status */
export const bookingStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'pending': 'Anfrage offen',
    'accepted': 'Zugesagt',
    'declined': 'Abgesagt',
    'alternative': 'Alternative',
  }
  return map[status?.toLowerCase()] || status
}

/** Transport / Tour status */
export const tourStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'planned': 'Geplant',
    'loading': 'Beladung',
    'in_transit': 'Unterwegs',
    'delivered': 'Geliefert',
    'completed': 'Abgeschlossen',
  }
  return map[status?.toLowerCase()] || status
}

/** Vehicle status */
export const vehicleStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'available': 'Verfügbar',
    'in_use': 'Im Einsatz',
    'maintenance': 'Wartung',
  }
  return map[status?.toLowerCase()] || status
}

/** Maintenance task status */
export const maintenanceStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'pending': 'Ausstehend',
    'planned': 'Geplant',
    'overdue': 'Überfällig',
    'in_progress': 'In Bearbeitung',
    'completed': 'Abgeschlossen',
    'cancelled': 'Storniert',
  }
  return map[status?.toLowerCase()] || status
}

/** Priority labels */
export const priorityLabel = (priority: string): string => {
  const map: Record<string, string> = {
    'low': 'Niedrig',
    'medium': 'Mittel',
    'high': 'Hoch',
    'critical': 'Kritisch',
  }
  return map[priority?.toLowerCase()] || priority
}

/** Quote status */
export const quoteStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'draft': 'Entwurf',
    'sent': 'Gesendet',
    'accepted': 'Akzeptiert',
    'rejected': 'Abgelehnt',
    'expired': 'Abgelaufen',
    'cancelled': 'Storniert',
  }
  return map[status?.toLowerCase()] || status
}

/** User status */
export const userStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'active': 'Aktiv',
    'invited': 'Einladung ausstehend',
    'pending': 'Ausstehend',
    'inactive': 'Deaktiviert',
    'disabled': 'Deaktiviert',
  }
  return map[status?.toLowerCase()] || status
}

/** E-Check / inspection result status */
export const inspectionStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'pass': 'Bestanden',
    'passed': 'Bestanden',
    'warning': 'Warnung',
    'conditional': 'Bedingt bestanden',
    'fail': 'Durchgefallen',
    'failed': 'Durchgefallen',
    'due': 'Fällig',
  }
  return map[status?.toLowerCase()] || status
}

/** Warehouse / packing status */
export const warehouseStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'packed': 'Gepackt',
    'in_transit': 'Unterwegs',
    'on_site': 'Vor Ort',
    'return_expected': 'Rückgabe erwartet',
  }
  return map[status?.toLowerCase()] || status
}

/** Mail status */
export const mailStatusLabel = (status: string): string => {
  const map: Record<string, string> = {
    'delivered': 'Zugestellt',
    'delivered_mail': 'Zugestellt',
    'opened': 'Geöffnet',
    'error': 'Fehler',
    'sent': 'Gesendet',
  }
  return map[status?.toLowerCase()] || status
}

// ---------------------------------------------------------------------------
// Flat lookup map (union of all domains — German only)
// ---------------------------------------------------------------------------

export const STATUS_LABELS: Record<string, Record<string, string>> = {
  de: {
    // Equipment statuses
    available: 'Verfügbar',
    reserved: 'Reserviert',
    checked_out: 'Vermietet',
    rented: 'Vermietet',
    in_maintenance: 'In Wartung',
    maintenance: 'In Wartung',
    damaged: 'Beschädigt',
    retired: 'Ausgemustert',
    in_use: 'In Gebrauch',
    lost: 'Verloren',
    missing: 'Vermisst',

    // Invoice statuses
    draft: 'Entwurf',
    sent: 'Gesendet',
    paid: 'Bezahlt',
    overdue: 'Überfällig',
    cancelled: 'Storniert',
    partial: 'Teilweise bezahlt',

    // Project statuses
    planning: 'In Planung',
    confirmed: 'Bestätigt',
    in_progress: 'In Bearbeitung',
    completed: 'Abgeschlossen',
    canceled: 'Storniert',
    active: 'Aktiv',
    quoted: 'Angebot',
    invoiced: 'Abgerechnet',

    // Quote statuses
    accepted: 'Akzeptiert',
    rejected: 'Abgelehnt',
    expired: 'Abgelaufen',

    // Warehouse statuses
    packed: 'Gepackt',
    in_transit: 'Unterwegs',
    on_site: 'Vor Ort',
    return_expected: 'Rückgabe erwartet',

    // Tour statuses
    planned: 'Geplant',
    loading: 'Beladung',
    delivered: 'Geliefert',

    // Crew availability
    busy: 'Beschäftigt',
    on_leave: 'Urlaub',
    sick: 'Krank',

    // Repair statuses
    open: 'Offen',
    done: 'Erledigt',

    // Inspection statuses
    due: 'Fällig',
    passed: 'Bestanden',
    pass: 'Bestanden',
    warning: 'Warnung',
    fail: 'Durchgefallen',

    // Mail statuses
    delivered_mail: 'Zugestellt',
    opened: 'Geöffnet',
    error: 'Fehler',

    // User statuses
    invited: 'Einladung ausstehend',
    pending: 'Ausstehend',
    inactive: 'Deaktiviert',
    disabled: 'Deaktiviert',

    // Condition ratings
    good: 'Gut',
    fair: 'Akzeptabel',
    poor: 'Schlecht',

    // General
    unknown: 'Unbekannt',
    none: 'Keine',
  },
  en: {
    // Equipment statuses
    available: 'Available',
    reserved: 'Reserved',
    checked_out: 'Rented',
    rented: 'Rented',
    in_maintenance: 'In Maintenance',
    maintenance: 'Maintenance',
    damaged: 'Damaged',
    retired: 'Retired',
    in_use: 'In Use',
    lost: 'Lost',
    missing: 'Missing',

    // Invoice statuses
    draft: 'Draft',
    sent: 'Sent',
    paid: 'Paid',
    overdue: 'Overdue',
    cancelled: 'Cancelled',
    partial: 'Partially Paid',

    // Project statuses
    planning: 'Planning',
    confirmed: 'Confirmed',
    in_progress: 'In Progress',
    completed: 'Completed',
    canceled: 'Cancelled',
    active: 'Active',
    quoted: 'Quoted',
    invoiced: 'Invoiced',

    // Quote statuses
    accepted: 'Accepted',
    rejected: 'Rejected',
    expired: 'Expired',

    // Warehouse statuses
    packed: 'Packed',
    in_transit: 'In Transit',
    on_site: 'On Site',
    return_expected: 'Return Expected',

    // Tour statuses
    planned: 'Planned',
    loading: 'Loading',
    delivered: 'Delivered',

    // Crew availability
    busy: 'Busy',
    on_leave: 'On Leave',
    sick: 'Sick',

    // Repair statuses
    open: 'Open',
    done: 'Done',

    // Inspection statuses
    due: 'Due',
    passed: 'Passed',
    pass: 'Pass',
    warning: 'Warning',
    fail: 'Fail',

    // Mail statuses
    delivered_mail: 'Delivered',
    opened: 'Opened',
    error: 'Error',

    // User statuses
    invited: 'Invitation Pending',
    pending: 'Pending',
    inactive: 'Disabled',
    disabled: 'Disabled',

    // Condition ratings
    good: 'Good',
    fair: 'Fair',
    poor: 'Poor',

    // General
    unknown: 'Unknown',
    none: 'None',
  },
}

/**
 * Get a translated status label.
 *
 * @param status - The status key (e.g. 'available', 'checked_out', 'overdue')
 * @param lang - Language code ('de' or 'en'). Defaults to 'de'.
 * @returns The translated label, or the raw status string if no translation exists.
 *
 * @example
 * ```ts
 * getStatusLabel('checked_out')       // 'Vermietet'
 * getStatusLabel('checked_out', 'en') // 'Rented'
 * getStatusLabel('unknown_status')    // 'unknown_status'
 * ```
 */
export function getStatusLabel(status: string, lang: string = 'de'): string {
  return STATUS_LABELS[lang]?.[status] || STATUS_LABELS['de']?.[status] || status
}

/**
 * Get the current language from localStorage (matching i18n config).
 */
export function getCurrentLanguage(): string {
  if (typeof window !== 'undefined' && window.localStorage) {
    return localStorage.getItem('language') || 'de'
  }
  return 'de'
}

/**
 * Get a translated status label using the current app language.
 *
 * @param status - The status key
 * @returns The translated label in the user's current language
 */
export function getLocalizedStatusLabel(status: string): string {
  return getStatusLabel(status, getCurrentLanguage())
}
