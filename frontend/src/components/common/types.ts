/**
 * Common component types
 */

export interface Column {
  key: string
  label: string
  sortable?: boolean
  align?: 'start' | 'center' | 'end'
  width?: string
  minWidth?: string
  maxWidth?: string
  /** Allow the user to resize this column on desktop (defaults to true). */
  resizable?: boolean
  headerClass?: string
  cellClass?: string
  formatter?: (value: any, row: any) => string
}
