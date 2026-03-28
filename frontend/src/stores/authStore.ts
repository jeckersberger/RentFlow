import { create } from 'zustand';
import type { AuthState, User } from '../types/auth';
import { login as loginApi } from '../services/auth';

const STORAGE_KEY_TOKEN = 'ef_access_token';
const STORAGE_KEY_USER = 'ef_user';

interface AuthStore extends AuthState {
  login: (email: string, password: string, tenantSlug: string) => Promise<void>;
  logout: () => void;
  init: () => void;
}

export const useAuthStore = create<AuthStore>((set) => ({
  accessToken: null,
  user: null,
  isAuthenticated: false,

  login: async (email: string, password: string, tenantSlug: string) => {
    const response = await loginApi(email, password, tenantSlug);
    const { access_token, user } = response;

    localStorage.setItem(STORAGE_KEY_TOKEN, access_token);
    localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(user));

    set({
      accessToken: access_token,
      user,
      isAuthenticated: true,
    });
  },

  logout: () => {
    localStorage.removeItem(STORAGE_KEY_TOKEN);
    localStorage.removeItem(STORAGE_KEY_USER);

    set({
      accessToken: null,
      user: null,
      isAuthenticated: false,
    });
  },

  init: () => {
    const token = localStorage.getItem(STORAGE_KEY_TOKEN);
    const userJson = localStorage.getItem(STORAGE_KEY_USER);

    if (token && userJson) {
      try {
        const user: User = JSON.parse(userJson);
        set({
          accessToken: token,
          user,
          isAuthenticated: true,
        });
      } catch {
        localStorage.removeItem(STORAGE_KEY_TOKEN);
        localStorage.removeItem(STORAGE_KEY_USER);
      }
    }
  },
}));
