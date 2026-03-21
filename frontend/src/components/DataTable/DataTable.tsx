import { useState } from 'react'
import './DataTable.module.scss'

export interface Column<T> {
  key: keyof T | string
  label: string
  sortable?: boolean
  render?: (value: any, row: T) => React.ReactNode
  width?: string
}

interface DataTableProps<T> {
  columns: Column<T>[]
  data: T[]
  rowKey: keyof T | ((row: T) => string)
  loading?: boolean
  onRowClick?: (row: T) => void
  pagination?: {
    page: number
    total: number
    limit: number
    onPageChange: (page: number) => void
  }
  onSort?: (key: string, order: 'asc' | 'desc') => void
}

export function DataTable<T extends Record<string, any>>({
  columns,
  data,
  rowKey,
  loading = false,
  onRowClick,
  pagination,
  onSort,
}: DataTableProps<T>) {
  const [sortKey, setSortKey] = useState<string | null>(null)
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>('asc')

  const handleSort = (key: string) => {
    if (!onSort) return

    let newOrder: 'asc' | 'desc' = 'asc'
    if (sortKey === key && sortOrder === 'asc') {
      newOrder = 'desc'
    }

    setSortKey(key)
    setSortOrder(newOrder)
    onSort(key, newOrder)
  }

  const getRowKey = (row: T): string => {
    if (typeof rowKey === 'function') {
      return rowKey(row)
    }
    return String(row[rowKey])
  }

  const getNestedValue = (obj: any, path: string): any => {
    return path.split('.').reduce((current, prop) => current?.[prop], obj)
  }

  const renderCell = (column: Column<T>, row: T): React.ReactNode => {
    const value = getNestedValue(row, String(column.key))
    if (column.render) {
      return column.render(value, row)
    }
    return value ?? '—'
  }

  if (loading) {
    return (
      <div className="data-table">
        <div className="data-table__skeleton">
          {[1, 2, 3, 4, 5].map((i) => (
            <div key={i} className="skeleton-row" />
          ))}
        </div>
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className="data-table">
        <div className="data-table__empty">
          <p>Keine Daten verfügbar</p>
        </div>
      </div>
    )
  }

  return (
    <div className="data-table-wrapper">
      <table className="data-table">
        <thead className="data-table__head">
          <tr>
            {columns.map((column) => (
              <th
                key={String(column.key)}
                className="data-table__th"
                style={{ width: column.width }}
              >
                {column.sortable ? (
                  <button
                    className="data-table__sort-btn"
                    onClick={() => handleSort(String(column.key))}
                  >
                    {column.label}
                    {sortKey === column.key && (
                      <span className="data-table__sort-indicator">
                        {sortOrder === 'asc' ? '↑' : '↓'}
                      </span>
                    )}
                  </button>
                ) : (
                  column.label
                )}
              </th>
            ))}
          </tr>
        </thead>
        <tbody className="data-table__body">
          {data.map((row) => (
            <tr
              key={getRowKey(row)}
              className="data-table__row"
              onClick={() => onRowClick?.(row)}
              style={{ cursor: onRowClick ? 'pointer' : 'default' }}
            >
              {columns.map((column) => (
                <td key={String(column.key)} className="data-table__td">
                  {renderCell(column, row)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      {pagination && (
        <div className="data-table__pagination">
          <div className="pagination__info">
            Seite {pagination.page} von{' '}
            {Math.ceil(pagination.total / pagination.limit)}
          </div>
          <div className="pagination__controls">
            <button
              className="pagination__btn"
              onClick={() => pagination.onPageChange(pagination.page - 1)}
              disabled={pagination.page === 1}
            >
              Vorherige
            </button>
            <span className="pagination__current">
              {pagination.page}
            </span>
            <button
              className="pagination__btn"
              onClick={() => pagination.onPageChange(pagination.page + 1)}
              disabled={
                pagination.page >=
                Math.ceil(pagination.total / pagination.limit)
              }
            >
              Nächste
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
