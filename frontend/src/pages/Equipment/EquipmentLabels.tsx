import { useState, useEffect, useMemo, useCallback, useRef } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import QRCode from 'qrcode'
import { jsPDF } from 'jspdf'
import { Printer, Download, Eye } from 'lucide-react'
import { equipmentApi, categoryApi } from '../../services/api'
import { Equipment, Category } from '../../types/equipment'
import styles from './EquipmentLabels.module.scss'

// ============================================================================
// TYPES
// ============================================================================

type LabelFormat = 'small' | 'medium' | 'large' | 'flightcase'

interface LabelFormatConfig {
  key: LabelFormat
  name: string
  size: string
  icon: string
  description: string
}

const LABEL_FORMATS: LabelFormatConfig[] = [
  { key: 'small', name: 'Klein', size: '25x15mm', icon: '[ ]', description: 'Nur QR-Code' },
  { key: 'medium', name: 'Mittel', size: '50x25mm', icon: '[= ]', description: 'QR + Name + SKU' },
  { key: 'large', name: 'Gro\u00df', size: '100x50mm', icon: '[== ]', description: 'QR + Barcode + alle Infos' },
  { key: 'flightcase', name: 'Flightcase', size: 'A5', icon: '[===]', description: 'Inhaltsliste' },
]

// ============================================================================
// QR CODE GENERATION (client-side)
// ============================================================================

async function generateQR(text: string, size: number = 128): Promise<string> {
  try {
    return await QRCode.toDataURL(text, {
      width: size,
      margin: 1,
      color: { dark: '#000000', light: '#ffffff' },
      errorCorrectionLevel: 'M',
    })
  } catch {
    return ''
  }
}

// ============================================================================
// SIMPLE BARCODE RENDERER (Code128-style visual)
// ============================================================================

function BarcodeCanvas({
  value,
  width = 160,
  height = 40,
}: {
  value: string
  width?: number
  height?: number
}) {
  const canvasRef = useRef<HTMLCanvasElement>(null)

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas || !value) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    canvas.width = width
    canvas.height = height
    ctx.fillStyle = '#ffffff'
    ctx.fillRect(0, 0, width, height)

    // Generate a deterministic bar pattern from the value
    const chars = value.split('')
    const totalBars = chars.length * 6 + 10 // approx bars
    const barWidth = Math.max(1, width / totalBars)
    let x = barWidth * 2 // quiet zone

    ctx.fillStyle = '#000000'

    // Start pattern
    ctx.fillRect(x, 0, barWidth, height); x += barWidth * 2
    ctx.fillRect(x, 0, barWidth, height); x += barWidth * 2

    for (const char of chars) {
      const code = char.charCodeAt(0)
      // Generate 6 bars per character based on char code
      for (let bit = 5; bit >= 0; bit--) {
        const isBar = (code >> bit) & 1
        if (isBar) {
          ctx.fillRect(x, 0, barWidth, height)
        }
        x += barWidth
      }
      // separator
      x += barWidth * 0.5
    }

    // End pattern
    ctx.fillRect(x, 0, barWidth, height); x += barWidth * 2
    ctx.fillRect(x, 0, barWidth, height)
  }, [value, width, height])

  return <canvas ref={canvasRef} width={width} height={height} style={{ display: 'block' }} />
}

// ============================================================================
// COMPONENT
// ============================================================================

function EquipmentLabelsPage() {
  const [searchParams] = useSearchParams()

  // State
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [searchFilter, setSearchFilter] = useState('')
  const [labelFormat, setLabelFormat] = useState<LabelFormat>('medium')
  const [columnsPerRow, setColumnsPerRow] = useState(3)
  const [qrCache, setQrCache] = useState<Record<string, string>>({})
  const [generatingPdf, setGeneratingPdf] = useState(false)

  // Data fetching
  const { data: equipmentData, isLoading } = useQuery({
    queryKey: ['equipment-labels-list'],
    queryFn: () => equipmentApi.list({ limit: 500, offset: 0 }),
  })

  const { data: categories } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoryApi.list() as Promise<Category[]>,
    staleTime: 1000 * 60 * 10,
  })

  const allItems: Equipment[] = useMemo(() => equipmentData?.data || [], [equipmentData])

  const getCategoryName = useCallback(
    (categoryId: string) => {
      const cat = categories?.find((c: Category) => c.id === categoryId)
      return cat ? cat.name : ''
    },
    [categories]
  )

  // Pre-select from URL params
  useEffect(() => {
    const idsParam = searchParams.get('ids')
    if (idsParam && allItems.length > 0) {
      setSelectedIds(new Set(idsParam.split(',')))
    }
  }, [searchParams, allItems.length])

  // Generate QR codes client-side for selected items
  useEffect(() => {
    const selected = allItems.filter((item) => selectedIds.has(item.id))
    const missing = selected.filter((item) => !qrCache[item.id])

    if (missing.length === 0) return

    let cancelled = false
    const generate = async () => {
      const newCache: Record<string, string> = {}
      for (const item of missing) {
        if (cancelled) break
        const qrText = item.barcode || item.sku || item.id
        const dataUrl = await generateQR(qrText)
        if (dataUrl) {
          newCache[item.id] = dataUrl
        }
      }
      if (!cancelled) {
        setQrCache((prev) => ({ ...prev, ...newCache }))
      }
    }
    generate()

    return () => { cancelled = true }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedIds, allItems])

  // Filtered list for selector
  const filteredItems = useMemo(() => {
    if (!searchFilter) return allItems
    const q = searchFilter.toLowerCase()
    return allItems.filter(
      (item) =>
        item.name.toLowerCase().includes(q) ||
        item.sku.toLowerCase().includes(q) ||
        (item.barcode && item.barcode.toLowerCase().includes(q))
    )
  }, [allItems, searchFilter])

  // Selected items for printing
  const selectedItems = useMemo(
    () => allItems.filter((item) => selectedIds.has(item.id)),
    [allItems, selectedIds]
  )

  // Preview item (first selected)
  const previewItem = selectedItems[0] || null

  // ---- HANDLERS ----

  const toggleItem = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const toggleAll = () => {
    if (selectedIds.size === filteredItems.length) {
      setSelectedIds(new Set())
    } else {
      setSelectedIds(new Set(filteredItems.map((i) => i.id)))
    }
  }

  const handlePrint = () => {
    window.print()
  }

  const handleDownloadPdf = async () => {
    if (selectedItems.length === 0) return
    setGeneratingPdf(true)

    try {
      const pdf = new jsPDF({ orientation: 'portrait', unit: 'mm', format: 'a4' })
      const pageWidth = 210
      const pageHeight = 297
      const margin = 10
      const availableWidth = pageWidth - margin * 2

      let labelW: number, labelH: number
      switch (labelFormat) {
        case 'small':
          labelW = 25; labelH = 15; break
        case 'medium':
          labelW = 50; labelH = 25; break
        case 'large':
          labelW = 100; labelH = 50; break
        case 'flightcase':
          labelW = 148; labelH = 210; break // A5
        default:
          labelW = 50; labelH = 25
      }

      const cols = labelFormat === 'flightcase' ? 1 : Math.min(columnsPerRow, Math.floor(availableWidth / labelW))
      const gapX = cols > 1 ? Math.min(4, (availableWidth - cols * labelW) / (cols - 1)) : 0
      const gapY = 4

      let x = margin
      let y = margin
      let col = 0

      for (let i = 0; i < selectedItems.length; i++) {
        const item = selectedItems[i]

        // Check page break
        if (y + labelH > pageHeight - margin) {
          pdf.addPage()
          y = margin
          col = 0
          x = margin
        }

        // Draw label border
        pdf.setDrawColor(180, 180, 180)
        pdf.rect(x, y, labelW, labelH)

        // QR Code
        const qrUrl = qrCache[item.id]
        if (qrUrl) {
          const qrSize = labelFormat === 'small' ? 12 : labelFormat === 'medium' ? 16 : labelFormat === 'flightcase' ? 24 : 30
          const qrX = labelFormat === 'small' ? x + (labelW - qrSize) / 2 : x + 2
          const qrY = labelFormat === 'small' ? y + (labelH - qrSize) / 2 : y + 2
          pdf.addImage(qrUrl, 'PNG', qrX, qrY, qrSize, qrSize)
        }

        // Text content (not for small)
        if (labelFormat !== 'small') {
          const textX = labelFormat === 'flightcase' ? x + 28 : labelFormat === 'large' ? x + 36 : x + 20
          const textStartY = y + 5

          pdf.setFontSize(labelFormat === 'large' ? 10 : 8)
          pdf.setFont('helvetica', 'bold')
          const displayName = item.name.length > 30 ? item.name.substring(0, 28) + '...' : item.name
          pdf.text(displayName, textX, textStartY)

          pdf.setFontSize(labelFormat === 'large' ? 8 : 7)
          pdf.setFont('courier', 'normal')
          pdf.text(item.barcode || item.sku, textX, textStartY + (labelFormat === 'large' ? 5 : 4))

          if (labelFormat === 'large') {
            const catName = getCategoryName(item.category_id)
            if (catName) {
              pdf.setFontSize(7)
              pdf.setFont('helvetica', 'normal')
              pdf.text(catName, textX, textStartY + 10)
            }
          }

          if (labelFormat === 'flightcase') {
            // Flightcase: List all selected items as contents
            pdf.setFontSize(12)
            pdf.setFont('helvetica', 'bold')
            pdf.text('Flightcase Inhalt', x + 4, y + 30)

            pdf.setFontSize(7)
            pdf.setFont('helvetica', 'normal')
            let listY = y + 38
            selectedItems.forEach((si, idx) => {
              if (listY < y + labelH - 6) {
                pdf.text(`${idx + 1}. ${si.name} (${si.barcode || si.sku})`, x + 6, listY)
                listY += 5
              }
            })
            // Only one flightcase label
            break
          }
        }

        // Move to next position
        col++
        if (col >= cols) {
          col = 0
          x = margin
          y += labelH + gapY
        } else {
          x += labelW + gapX
        }
      }

      pdf.save(`EquipFlow-Labels-${new Date().toISOString().slice(0, 10)}.pdf`)
    } catch (err) {
      console.error('PDF generation failed:', err)
    } finally {
      setGeneratingPdf(false)
    }
  }

  // ---- RENDER HELPERS ----

  const renderLabelPreview = (item: Equipment) => {
    const qrUrl = qrCache[item.id]
    const catName = getCategoryName(item.category_id)
    const barcodeValue = item.barcode || item.sku

    if (labelFormat === 'small') {
      return (
        <div className={`${styles.labelPreview} ${styles.labelSmall}`}>
          {qrUrl ? (
            <div className={styles.labelQr}>
              <img src={qrUrl} alt="QR" width={44} height={44} />
            </div>
          ) : (
            <div style={{ fontSize: '7px', fontFamily: 'monospace', textAlign: 'center' }}>
              {barcodeValue}
            </div>
          )}
        </div>
      )
    }

    if (labelFormat === 'medium') {
      return (
        <div className={`${styles.labelPreview} ${styles.labelMedium}`}>
          {qrUrl && (
            <div className={styles.labelQr}>
              <img src={qrUrl} alt="QR" width={60} height={60} />
            </div>
          )}
          <div className={styles.labelInfo}>
            <div className={styles.labelName} style={{ fontSize: '11px' }}>{item.name}</div>
            <div className={styles.labelSku} style={{ fontSize: '9px' }}>{barcodeValue}</div>
          </div>
        </div>
      )
    }

    if (labelFormat === 'large') {
      return (
        <div className={`${styles.labelPreview} ${styles.labelLarge}`}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', flexShrink: 0 }}>
            {qrUrl && (
              <div className={styles.labelQr}>
                <img src={qrUrl} alt="QR" width={80} height={80} />
              </div>
            )}
            <div className={styles.labelBarcode}>
              <BarcodeCanvas value={barcodeValue} width={100} height={28} />
              <div className={styles.labelBarcodeText}>{barcodeValue}</div>
            </div>
          </div>
          <div className={styles.labelInfo}>
            <div className={styles.labelName} style={{ fontSize: '14px' }}>{item.name}</div>
            <div className={styles.labelSku} style={{ fontSize: '10px' }}>{barcodeValue}</div>
            {catName && <div className={styles.labelCategory} style={{ fontSize: '9px' }}>{catName}</div>}
            {item.location_id && (
              <div className={styles.labelLocation} style={{ fontSize: '8px' }}>
                Standort: {item.location_id}
              </div>
            )}
            <div className={styles.labelCompanyLogo} style={{ marginTop: '8px', borderTop: '1px solid #ddd', paddingTop: '4px', borderBottom: 'none', marginBottom: 0 }}>
              [Firmenlogo]
            </div>
          </div>
        </div>
      )
    }

    // Flightcase
    return (
      <div className={`${styles.labelPreview} ${styles.labelFlightcase}`}>
        <div className={styles.flightcaseHeader}>
          <div>
            <div className={styles.flightcaseTitle}>Flightcase Inhalt</div>
            <div className={styles.flightcaseSubtitle}>
              {selectedItems.length} Artikel | Erstellt: {new Date().toLocaleDateString('de-DE')}
            </div>
          </div>
          {qrUrl && (
            <div className={styles.labelQr}>
              <img src={qrUrl} alt="QR" width={48} height={48} />
            </div>
          )}
        </div>
        <div className={styles.flightcaseTable}>
          <table>
            <thead>
              <tr>
                <th>#</th>
                <th>Bezeichnung</th>
                <th>SKU / Barcode</th>
                <th>Kategorie</th>
              </tr>
            </thead>
            <tbody>
              {selectedItems.map((si, idx) => (
                <tr key={si.id}>
                  <td>{idx + 1}</td>
                  <td style={{ fontWeight: 600 }}>{si.name}</td>
                  <td style={{ fontFamily: 'monospace', fontSize: '8px' }}>{si.barcode || si.sku}</td>
                  <td>{getCategoryName(si.category_id)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    )
  }

  const renderPrintLabel = (item: Equipment) => {
    const qrUrl = qrCache[item.id]
    const catName = getCategoryName(item.category_id)
    const barcodeValue = item.barcode || item.sku

    if (labelFormat === 'small') {
      return (
        <div className={`${styles.printLabel} ${styles.printLabelSmall}`} key={item.id}>
          {qrUrl ? (
            <div className={styles.printLabelQr}>
              <img src={qrUrl} alt="QR" width={50} height={50} />
            </div>
          ) : (
            <div style={{ fontSize: '8px', fontFamily: 'monospace' }}>{barcodeValue}</div>
          )}
        </div>
      )
    }

    if (labelFormat === 'medium') {
      return (
        <div className={styles.printLabel} key={item.id}>
          {qrUrl && (
            <div className={styles.printLabelQr}>
              <img src={qrUrl} alt="QR" width={64} height={64} />
            </div>
          )}
          <div className={styles.printLabelInfo}>
            <div className={styles.printLabelName}>{item.name}</div>
            <div className={styles.printLabelSku}>{barcodeValue}</div>
          </div>
        </div>
      )
    }

    if (labelFormat === 'large') {
      return (
        <div className={styles.printLabel} key={item.id}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: '4px', flexShrink: 0 }}>
            {qrUrl && (
              <div className={styles.printLabelQr}>
                <img src={qrUrl} alt="QR" width={80} height={80} />
              </div>
            )}
            <div className={styles.printLabelBarcode}>
              <BarcodeCanvas value={barcodeValue} width={120} height={32} />
              <div style={{ fontFamily: 'monospace', fontSize: '7px', fontWeight: 700, textAlign: 'center', marginTop: '1px' }}>
                {barcodeValue}
              </div>
            </div>
          </div>
          <div className={styles.printLabelInfo}>
            <div className={styles.printLabelName}>{item.name}</div>
            <div className={styles.printLabelSku}>{barcodeValue}</div>
            {catName && <div className={styles.printLabelCategory}>{catName}</div>}
            {item.location_id && (
              <div className={styles.printLabelLocation}>Standort: {item.location_id}</div>
            )}
            <div style={{ marginTop: '6px', fontSize: '8px', color: '#999', textAlign: 'right' }}>
              [Firmenlogo]
            </div>
          </div>
        </div>
      )
    }

    // Flightcase
    return (
      <div className={`${styles.printLabel} ${styles.printLabelFlightcase}`} key="flightcase">
        <div className={styles.flightcaseHeader}>
          <div>
            <div className={styles.flightcaseTitle}>Flightcase Inhalt</div>
            <div className={styles.flightcaseSubtitle}>
              {selectedItems.length} Artikel | {new Date().toLocaleDateString('de-DE')}
            </div>
          </div>
          {qrUrl && (
            <div className={styles.printLabelQr}>
              <img src={qrUrl} alt="QR" width={48} height={48} />
            </div>
          )}
        </div>
        <div className={styles.flightcaseTable}>
          <table>
            <thead>
              <tr>
                <th>#</th>
                <th>Bezeichnung</th>
                <th>SKU / Barcode</th>
                <th>Kategorie</th>
              </tr>
            </thead>
            <tbody>
              {selectedItems.map((si, idx) => (
                <tr key={si.id}>
                  <td>{idx + 1}</td>
                  <td style={{ fontWeight: 600 }}>{si.name}</td>
                  <td style={{ fontFamily: 'monospace', fontSize: '8px' }}>{si.barcode || si.sku}</td>
                  <td>{getCategoryName(si.category_id)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    )
  }

  // ---- RENDER ----

  if (isLoading) {
    return (
      <div className={styles.labelsPage}>
        <div className={styles.loading}>Ausr\u00fcstungsdaten werden geladen...</div>
      </div>
    )
  }

  const gridClass = `${styles.labelsGrid} ${
    columnsPerRow === 1
      ? styles.labelsPer1
      : columnsPerRow === 2
        ? styles.labelsPer2
        : columnsPerRow === 4
          ? styles.labelsPer4
          : styles.labelsPer3
  }`

  return (
    <div className={styles.labelsPage}>
      {/* Header */}
      <div className={styles.header}>
        <div>
          <h1 className={styles.headerTitle}>Label-Designer</h1>
          <p className={styles.headerSubtitle}>
            QR-Codes, Barcodes und Etiketten f\u00fcr Ihre Ausr\u00fcstung erstellen und drucken
          </p>
        </div>
        <div className={styles.headerActions}>
          <button
            className={styles.btnSecondary}
            onClick={handleDownloadPdf}
            disabled={selectedItems.length === 0 || generatingPdf}
          >
            <Download size={16} />
            {generatingPdf ? 'PDF wird erstellt...' : 'Als PDF herunterladen'}
          </button>
          <button
            className={styles.btnPrimary}
            onClick={handlePrint}
            disabled={selectedItems.length === 0}
          >
            <Printer size={16} />
            Im Browser drucken
          </button>
        </div>
      </div>

      {/* Main Layout */}
      <div className={styles.mainLayout}>
        {/* Sidebar */}
        <div className={styles.sidebar}>
          {/* Equipment Selector */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Ausr\u00fcstung ausw\u00e4hlen</h3>
            <input
              type="text"
              className={styles.searchInput}
              placeholder="Nach Name oder SKU suchen..."
              value={searchFilter}
              onChange={(e) => setSearchFilter(e.target.value)}
            />
            <div className={styles.selectAllRow} onClick={toggleAll}>
              <input
                type="checkbox"
                checked={filteredItems.length > 0 && selectedIds.size === filteredItems.length}
                readOnly
              />
              <span>Alle ausw\u00e4hlen ({filteredItems.length})</span>
            </div>
            <div className={styles.equipmentList}>
              {filteredItems.map((item) => (
                <label key={item.id} className={styles.equipmentItem}>
                  <input
                    type="checkbox"
                    checked={selectedIds.has(item.id)}
                    onChange={() => toggleItem(item.id)}
                  />
                  <div className={styles.equipmentItemInfo}>
                    <div className={styles.equipmentItemName}>{item.name}</div>
                    <div className={styles.equipmentItemSku}>{item.barcode || item.sku}</div>
                  </div>
                </label>
              ))}
              {filteredItems.length === 0 && (
                <div className={styles.emptyState}>Keine Ausr\u00fcstung gefunden</div>
              )}
            </div>
            {selectedIds.size > 0 && (
              <div className={styles.selectedCount}>
                {selectedIds.size} von {allItems.length} ausgew\u00e4hlt
              </div>
            )}
          </div>

          {/* Format Selector */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Label-Format</h3>
            <div className={styles.formatGrid}>
              {LABEL_FORMATS.map((fmt) => (
                <button
                  key={fmt.key}
                  className={`${styles.formatOption} ${labelFormat === fmt.key ? styles.formatOptionActive : ''}`}
                  onClick={() => setLabelFormat(fmt.key)}
                >
                  <span className={styles.formatIcon}>{fmt.icon}</span>
                  <span className={styles.formatName}>{fmt.name}</span>
                  <span className={styles.formatSize}>{fmt.size}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Print Options */}
          <div className={styles.card}>
            <h3 className={styles.cardTitle}>Druckoptionen</h3>
            <div className={styles.printOptions}>
              {labelFormat !== 'flightcase' && (
                <div className={styles.optionRow}>
                  <label>Anzahl pro Zeile</label>
                  <select
                    value={columnsPerRow}
                    onChange={(e) => setColumnsPerRow(Number(e.target.value))}
                  >
                    <option value={1}>1</option>
                    <option value={2}>2</option>
                    <option value={3}>3</option>
                    <option value={4}>4</option>
                  </select>
                </div>
              )}
              <div className={styles.printButtons}>
                <button
                  className={styles.btnPrimary}
                  onClick={handlePrint}
                  disabled={selectedItems.length === 0}
                >
                  <Printer size={16} />
                  Im Browser drucken
                </button>
                <button
                  className={styles.btnSecondary}
                  onClick={handleDownloadPdf}
                  disabled={selectedItems.length === 0 || generatingPdf}
                >
                  <Download size={16} />
                  {generatingPdf ? 'Wird erstellt...' : 'Als PDF herunterladen'}
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* Content Area */}
        <div className={styles.contentArea}>
          {/* Live Preview */}
          <div className={styles.previewSection}>
            <h3 className={styles.previewTitle}>
              <Eye size={14} style={{ display: 'inline', verticalAlign: 'middle', marginRight: '6px' }} />
              Vorschau (Originalgr\u00f6\u00dfe)
            </h3>
            <div className={styles.previewContainer}>
              {previewItem ? (
                renderLabelPreview(previewItem)
              ) : (
                <div className={styles.previewEmpty}>
                  W\u00e4hlen Sie Ausr\u00fcstung aus, um eine Vorschau zu sehen
                </div>
              )}
            </div>
          </div>

          {/* All Labels Grid (for print) */}
          <div className={styles.labelsSection}>
            <h3 className={styles.cardTitle}>
              Labels ({selectedItems.length})
            </h3>
            {selectedItems.length === 0 ? (
              <div className={styles.emptyState}>
                W\u00e4hlen Sie links Ausr\u00fcstung aus, um Labels zu erzeugen
              </div>
            ) : labelFormat === 'flightcase' ? (
              <div className={gridClass}>
                {renderPrintLabel(selectedItems[0])}
              </div>
            ) : (
              <div className={gridClass}>
                {selectedItems.map((item) => renderPrintLabel(item))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

export default EquipmentLabelsPage
