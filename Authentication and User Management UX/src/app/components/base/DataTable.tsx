import React from 'react';
import { ChevronUp, ChevronDown } from 'lucide-react';

interface Column<T> {
  key: keyof T;
  label: string;
  sortable?: boolean;
  width?: string;
  render?: (value: any, item: T) => React.ReactNode;
}

interface DataTableProps<T extends { id: string | number }> {
  columns: Column<T>[];
  data: T[];
  onRowClick?: (item: T) => void;
  sortBy?: keyof T;
  sortOrder?: 'asc' | 'desc';
  hoverable?: boolean;
}

export function DataTable<T extends { id: string | number }>({
  columns,
  data,
  onRowClick,
  sortBy,
  sortOrder = 'asc',
  hoverable = true,
}: DataTableProps<T>) {
  return (
    <div className="w-full overflow-x-auto border border-gray-200 rounded-lg">
      <table className="w-full">
        {/* Header */}
        <thead>
          <tr className="bg-gray-50 border-b border-gray-200">
            {columns.map((col) => (
              <th
                key={String(col.key)}
                className={`px-6 py-3 text-left text-sm font-semibold text-gray-700 ${
                  col.sortable ? 'cursor-pointer hover:bg-gray-100' : ''
                } ${col.width ? `w-${col.width}` : ''}`}
              >
                <div className="flex items-center gap-2">
                  {col.label}
                  {col.sortable && sortBy === col.key && (
                    <span className="text-blue-600">
                      {sortOrder === 'asc' ? (
                        <ChevronUp size={16} />
                      ) : (
                        <ChevronDown size={16} />
                      )}
                    </span>
                  )}
                </div>
              </th>
            ))}
          </tr>
        </thead>

        {/* Body */}
        <tbody>
          {data.map((item, idx) => (
            <tr
              key={item.id}
              onClick={() => onRowClick?.(item)}
              className={`border-b border-gray-200 ${
                hoverable ? 'hover:bg-gray-50 transition-colors cursor-pointer' : ''
              } ${idx % 2 === 0 ? 'bg-white' : 'bg-gray-50'}`}
            >
              {columns.map((col) => (
                <td key={String(col.key)} className="px-6 py-4 text-sm text-gray-700">
                  {col.render ? col.render(item[col.key], item) : String(item[col.key])}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>

      {/* Empty State */}
      {data.length === 0 && (
        <div className="px-6 py-8 text-center text-gray-500">
          No hay datos disponibles
        </div>
      )}
    </div>
  );
}
