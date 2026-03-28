import api from './api';
import type { LoginResponse } from '../types/auth';

export async function login(
  email: string,
  password: string,
  tenantSlug: string
): Promise<LoginResponse> {
  const response = await api.post('/api/v1/auth/login', {
    email,
    password,
    tenant_slug: tenantSlug,
  });
  return response as unknown as LoginResponse;
}
