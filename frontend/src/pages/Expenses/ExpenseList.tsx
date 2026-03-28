import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { PageWrapper } from '@/components/PageWrapper/PageWrapper';
import { DataTable, Column } from '@/components/DataTable/DataTable';
import { StatusBadge } from '@/components/StatusBadge/StatusBadge';
import * as expenseApi from '@/services/expense';
import type { Expense, ExpenseCategory } from '@/types/expense';
import './ExpenseList.scss';

type Tab = 'expenses' | 'categories';

const expenseColumns: Column<Expense>[] = [
  { key: 'description', label: 'Beschreibung' },
  { key: 'vendor', label: 'Lieferant', render: (row) => row.vendor || '-' },
  {
    key: 'amount',
    label: 'Betrag',
    render: (row) => `${(row.amount / 100).toFixed(2)} €`,
  },
  { key: 'expense_date', label: 'Datum' },
  {
    key: 'status',
    label: 'Status',
    render: (row) => <StatusBadge status={row.status} />,
  },
];

const categoryColumns: Column<ExpenseCategory>[] = [
  { key: 'name', label: 'Name' },
  { key: 'code', label: 'Code', render: (row) => row.code || '-' },
  { key: 'description', label: 'Beschreibung', render: (row) => row.description || '-' },
  {
    key: 'is_active',
    label: 'Status',
    render: (row) => <StatusBadge status={row.is_active ? 'active' : 'cancelled'} />,
  },
];

export default function ExpenseList() {
  const [tab, setTab] = useState<Tab>('expenses');

  const { data: expenses, isLoading: expensesLoading } = useQuery({
    queryKey: ['expenses'],
    queryFn: () => expenseApi.listExpenses(),
  });

  const { data: categories, isLoading: categoriesLoading } = useQuery({
    queryKey: ['expense-categories'],
    queryFn: expenseApi.listCategories,
  });

  const expenseItems: Expense[] = Array.isArray(expenses) ? expenses : [];
  const categoryItems: ExpenseCategory[] = Array.isArray(categories) ? categories : [];

  return (
    <PageWrapper title="Ausgaben">
      <div className="tabs">
        <button className={`tabs__btn ${tab === 'expenses' ? 'tabs__btn--active' : ''}`} onClick={() => setTab('expenses')}>Ausgaben</button>
        <button className={`tabs__btn ${tab === 'categories' ? 'tabs__btn--active' : ''}`} onClick={() => setTab('categories')}>Kategorien</button>
      </div>

      {tab === 'expenses' && (
        <DataTable<Expense>
          columns={expenseColumns}
          data={expenseItems}
          loading={expensesLoading}
          emptyMessage="Keine Ausgaben vorhanden"
        />
      )}

      {tab === 'categories' && (
        <DataTable<ExpenseCategory>
          columns={categoryColumns}
          data={categoryItems}
          loading={categoriesLoading}
          emptyMessage="Keine Kategorien vorhanden"
        />
      )}
    </PageWrapper>
  );
}
