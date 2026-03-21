import { create } from 'zustand'

interface ScanResult {
  code: string
  timestamp: number
  type?: 'qr' | 'barcode' | 'unknown'
}

interface ScannerStore {
  isScanning: boolean
  lastScannedCode: string | null
  scanHistory: ScanResult[]
  cameraPermission: 'granted' | 'denied' | 'prompt' | null
  startScan: () => void
  stopScan: () => void
  addScanResult: (code: string, type?: 'qr' | 'barcode' | 'unknown') => void
  clearHistory: () => void
  setCameraPermission: (permission: 'granted' | 'denied' | 'prompt') => void
}

export const useScannerStore = create<ScannerStore>((set) => ({
  isScanning: false,
  lastScannedCode: null,
  scanHistory: [],
  cameraPermission: null,

  startScan: () =>
    set({ isScanning: true }),

  stopScan: () =>
    set({ isScanning: false }),

  addScanResult: (code: string, type = 'unknown') =>
    set((state) => ({
      lastScannedCode: code,
      scanHistory: [
        {
          code,
          timestamp: Date.now(),
          type,
        },
        ...state.scanHistory,
      ].slice(0, 100), // Keep last 100 scans
    })),

  clearHistory: () =>
    set({
      scanHistory: [],
      lastScannedCode: null,
    }),

  setCameraPermission: (permission: 'granted' | 'denied' | 'prompt') =>
    set({ cameraPermission: permission }),
}))
