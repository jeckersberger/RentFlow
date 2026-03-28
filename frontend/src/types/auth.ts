export interface User {
  id: string;
  email: string;
  role: string;
  tenant_id: string;
}

export interface LoginResponse {
  access_token: string;
  user: User;
}

export interface AuthState {
  accessToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
}
