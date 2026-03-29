export interface User {
  id: string;
  email: string;
  role: string;
  tenant_id: string;
}

export interface LoginTokens {
  access_token: string;
  refresh_token: string;
  expires_at: number;
}

export interface LoginResponse {
  tokens: LoginTokens;
  user: User;
}

export interface AuthState {
  accessToken: string | null;
  user: User | null;
  isAuthenticated: boolean;
}
