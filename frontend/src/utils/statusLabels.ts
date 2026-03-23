/**
 * Centralized status label translations for use across all pages.
 * Provides a quick lookup for status strings without requiring
 * the full i18n namespace infrastructure.
 */

export const STATUS_LABELS: Record<string, Record<string, string>> = {
  de: {
    // Equipment statuses
    available: 'Verfügbar',
    reserved: 'Reserviert',
    checked_out: 'Vermietet',
    in_maintenance: 'In Wartung',
    damaged: 'Beschädigt',
    retired: 'Ausgemustert',
    in_use: 'In Gebrauch',

    // Invoice statuses
    draft: 'Entwurf',
    sent: 'Gesendet',
    paid: 'Bezahlt',
    overdue: 'Überfällig',
    cancelled: 'Storniert',
    partial: 'Teilweise bezahlt',

    // Project statuses
    planning: 'Planung',
    confirmed: 'Bestätigt',
    in_progress: 'In Bearbeitung',
    completed: 'Abgeschlossen',
    canceled: 'Storniert',
    active: 'Aktiv',

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

    // Vehicle statuses
    maintenance: 'Wartung',

    // Repair statuses
    open: 'Offen',
    done: 'Erledigt',

    // Inspection statuses
    due: 'Fällig',
    passed: 'Bestanden',

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
    in_maintenance: 'In Maintenance',
    damaged: 'Damaged',
    retired: 'Retired',
    in_use: 'In Use',

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

    // Vehicle statuses
    maintenance: 'Maintenance',

    // Repair statuses
    open: 'Open',
    done: 'Done',

    // Inspection statuses
    due: 'Due',
    passed: 'Passed',

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
