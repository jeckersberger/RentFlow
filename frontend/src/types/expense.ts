export interface ExpenseCategory {
  id: string;
  tenant_id: string;
  name: string;
  code: string;
  description: string;
  is_active: boolean;
  created_at: string;
}

export interface Expense {
  id: string;
  tenant_id: string;
  project_id: string;
  category_id: string;
  description: string;
  amount: number;
  currency: string;
  expense_date: string;
  receipt_number: string;
  vendor: string;
  status: string;
  approved_by: string;
  approved_at: string;
  notes: string;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface ExpenseReceipt {
  id: string;
  expense_id: string;
  tenant_id: string;
  file_name: string;
  file_path: string;
  file_size: number;
  mime_type: string;
  uploaded_by: string;
  created_at: string;
}
