import { create } from 'zustand';

interface SidebarState {
  isExpanded: boolean;
  isPinned: boolean;
  toggle: () => void;
  pin: () => void;
  expand: () => void;
  collapse: () => void;
}

export const useSidebarStore = create<SidebarState>((set) => ({
  isExpanded: false,
  isPinned: false,

  toggle: () =>
    set((state) => ({ isExpanded: !state.isExpanded })),

  pin: () =>
    set((state) => ({
      isPinned: !state.isPinned,
      isExpanded: !state.isPinned ? true : state.isExpanded,
    })),

  expand: () =>
    set({ isExpanded: true }),

  collapse: () =>
    set((state) => ({
      isExpanded: state.isPinned ? true : false,
    })),
}));
