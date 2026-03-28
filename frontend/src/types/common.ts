export interface PaginatedResponse<T> {
  data: T[];
  meta: { page: number; per_page: number; total: number };
}

export interface ApiError {
  code: string;
  message: string;
}
