import { create } from 'zustand'

interface ModuleStore {
  enabledModules: string[]
  loaded: boolean
  setEnabledModules: (modules: string[]) => void
  isModuleEnabled: (moduleId: string) => boolean
}

export const useModuleStore = create<ModuleStore>((set, get) => ({
  enabledModules: ['warehouse', 'projects', 'finance'],
  loaded: false,
  setEnabledModules: (modules) => set({ enabledModules: modules, loaded: true }),
  isModuleEnabled: (moduleId) => get().enabledModules.includes(moduleId),
}))
