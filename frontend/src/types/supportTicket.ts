export type SupportTicketCategory = 'ACCOUNT' | 'REFUND'

export type SupportTicketStatus =
  | 'PENDING_ADMIN'
  | 'PENDING_USER'
  | 'IN_PROGRESS'
  | 'RESOLVED'
  | 'CLOSED'
  | 'CANCELLED'

export interface SupportTicket {
  id: string
  user_id: string
  category: SupportTicketCategory
  status: SupportTicketStatus
  origin: 'USER' | 'ADMIN' | 'MIGRATED'
  title: string
  api_key_id?: string
  api_key_name_snapshot?: string
  group_id?: string
  group_name_snapshot?: string
  order_id?: string
  refund_id?: string
  refund_decision: 'NONE' | 'PENDING' | 'APPROVED' | 'REJECTED'
  approved_principal_amount?: number
  reviewer_id?: string
  reviewed_at?: string
  reopen_count: number
  resolved_at?: string
  closed_at?: string
  created_at: string
  updated_at: string
  unread?: boolean
  username?: string
  email?: string
}

export interface SupportTicketMessage {
  id: string
  ticket_id: string
  author_id?: string
  author_role: 'USER' | 'ADMIN' | 'SYSTEM'
  kind: 'COMMENT' | 'SYSTEM'
  body: string
  event_type?: string
  event_data?: string
  created_at: string
}

export interface SupportTicketPermissions {
  can_reply: boolean
  can_cancel: boolean
  can_close: boolean
  can_reopen: boolean
  can_review_refund: boolean
  can_retry_refund: boolean
}

export interface SupportTicketDetail {
  ticket: SupportTicket
  messages: SupportTicketMessage[]
  user?: { id: string; email: string; username: string }
  refund?: Record<string, unknown>
  permissions: SupportTicketPermissions
}

export interface SupportTicketSummary {
  total: number
  unread: number
  can_create_account: boolean
  can_create_refund: boolean
  feature_enabled: boolean
}

export interface SupportTicketPage {
  items: SupportTicket[]
  total: number
  page?: number
  page_size?: number
}
