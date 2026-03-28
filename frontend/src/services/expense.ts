import api from './api';
import type { ExpenseCategory, Expense, ExpenseReceipt } from '@/types/expense';

export async function listCategories(): Promise<ExpenseCategory[]> {
  return api.get('/api/v1/expense-categories') as unknown as ExpenseCategory[];
}

export async function getCategory(id: string): Promise<ExpenseCategory> {
  return api.get(`/api/v1/expense-categories/${id}`) as unknown as ExpenseCategory;
}

export async function createCategory(body: Partial<ExpenseCategory>): Promise<ExpenseCategory> {
  return api.post('/api/v1/expense-categories', body) as unknown as ExpenseCategory;
}

export async function updateCategory(id: string, body: Partial<ExpenseCategory>): Promise<ExpenseCategory> {
  return api.put(`/api/v1/expense-categories/${id}`, body) as unknown as ExpenseCategory;
}

export async function listExpenses(params?: { project_id?: string; status?: string }): Promise<Expense[]> {
  return api.get('/api/v1/expenses', { params }) as unknown as Expense[];
}

export async function getExpense(id: string): Promise<Expense> {
  return api.get(`/api/v1/expenses/${id}`) as unknown as Expense;
}

export async function createExpense(body: Partial<Expense>): Promise<Expense> {
  return api.post('/api/v1/expenses', body) as unknown as Expense;
}

export async function updateExpense(id: string, body: Partial<Expense>): Promise<Expense> {
  return api.put(`/api/v1/expenses/${id}`, body) as unknown as Expense;
}

export async function approveExpense(id: string): Promise<Expense> {
  return api.patch(`/api/v1/expenses/${id}/approve`) as unknown as Expense;
}

export async function listReceipts(expenseId: string): Promise<ExpenseReceipt[]> {
  return api.get(`/api/v1/expenses/${expenseId}/receipts`) as unknown as ExpenseReceipt[];
}
