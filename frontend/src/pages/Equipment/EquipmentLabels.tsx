import { useState, useEffect, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { equipmentApi, categoryApi } from '../../services/api'
import { Equipment, Category } from '../../types/equipment'
import './EquipmentLabels.css'

/**
 * Generates a simple QR code as an SVG data URL.
 * This is a minimal implementation that encodes text into a QR-like visual pattern.
 * For production, consider using a proper QR library.
 * Falls back to the backend API endpoint for actual QR codes.
 */
function generateQRCodeDataUrl(equipmentId: string): string {
  // Use the backend API endpoint to get a real QR code
  // This returns the URL that can be used as img src
  return `/api/v1/equipment/${equipmentId}/qr-code`
}

function EquipmentLabelsPage() {
  const [searchParams] = useSearchParams()
  const [qrUrls, setQrUrls] = useState<Record<string, string>>({})
  const [loadingQR, setLoadingQR] = useState(false)

  // Get IDs from URL params (comma-separated) or show all
  const idsParam = searchParams.get('ids')
  const selectedIds = idsParam ? idsParam.split(',') : null

  // Fetch all equipment (we need the full list to filter)
  const { data: equipmentData, isLoading } = useQuery({
    queryKey: ['equipment-labels-list'],
    queryFn: async () => {
      // Fetch a large batch - labels are typically for all or selected equipment
      return equipmentApi.list({ limit: 500, offset: 0 })
    },
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const allItems: Equipment[] = equipmentData?.data || []
  const items = selectedIds
    ? allItems.filter((item) => selectedIds.includes(item.id))
    : allItems

  const getCategoryName = (categoryId: string) => {
    const cat = categories?.find((c: Category) => c.id === categoryId)
    return cat ? cat.name : ''
  }

  // Fetch QR codes from the backend API as blobs
  const fetchQRCodes = useCallback(async (equipmentItems: Equipment[]) => {
    setLoadingQR(true)
    const urls: Record<string, string> = {}

    // Fetch in parallel, with a concurrency limit
    const batchSize = 10
    for (let i = 0; i < equipmentItems.length; i += batchSize) {
      const batch = equipmentItems.slice(i, i + batchSize)
      const results = await Promise.allSettled(
        batch.map(async (item) => {
          try {
            const blob = await equipmentApi.getQRCode(item.id)
            const url = URL.createObjectURL(blob)
            return { id: item.id, url }
          } catch {
            // Fallback: use the direct API URL (may work if auth is cookie-based)
            return { id: item.id, url: generateQRCodeDataUrl(item.id) }
          }
        })
      )
      results.forEach((result) => {
        if (result.status === 'fulfilled') {
          urls[result.value.id] = result.value.url
        }
      })
    }

    setQrUrls(urls)
    setLoadingQR(false)
  }, [])

  useEffect(() => {
    if (items.length > 0) {
      fetchQRCodes(items)
    }
    // Cleanup blob URLs on unmount
    return () => {
      Object.values(qrUrls).forEach((url) => {
        if (url.startsWith('blob:')) {
          URL.revokeObjectURL(url)
        }
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [items.length, fetchQRCodes])

  const handlePrint = () => {
    window.print()
  }

  if (isLoading) {
    return (
      <div className="labels-page">
        <div className="labels-loading">Ausrüstungsdaten werden geladen...</div>
      </div>
    )
  }

  if (items.length === 0) {
    return (
      <div className="labels-page">
        <div className="labels-toolbar no-print">
          <h1>QR-Labels drucken</h1>
          <p>Keine Ausrüstung gefunden.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="labels-page">
      <div className="labels-toolbar no-print">
        <div>
          <h1>QR-Labels drucken</h1>
          <p>
            {items.length} Label{items.length !== 1 ? 's' : ''} bereit zum Drucken
            {selectedIds && ` (${selectedIds.length} ausgewählt)`}
          </p>
        </div>
        <div style={{ display: 'flex', gap: '12px', alignItems: 'center' }}>
          {loadingQR && <span style={{ color: 'var(--color-text-secondary)' }}>QR-Codes werden geladen...</span>}
          <button
            className="btn btn--primary"
            onClick={handlePrint}
            disabled={loadingQR}
          >
            Drucken
          </button>
        </div>
      </div>

      <div className="labels-grid">
        {items.map((item) => (
          <div key={item.id} className="label-card">
            <div className="label-qr">
              {qrUrls[item.id] ? (
                <img
                  src={qrUrls[item.id]}
                  alt={`QR Code for ${item.name}`}
                  className="label-qr-img"
                />
              ) : (
                <div className="label-qr-fallback">
                  <span className="label-qr-fallback-text">
                    {item.barcode || item.sku}
                  </span>
                </div>
              )}
            </div>
            <div className="label-info">
              <div className="label-name">{item.name}</div>
              <div className="label-sku">{item.barcode || item.sku}</div>
              {getCategoryName(item.category_id) && (
                <div className="label-category">{getCategoryName(item.category_id)}</div>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default EquipmentLabelsPage
