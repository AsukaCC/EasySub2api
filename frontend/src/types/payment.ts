/**
 * Payment System Type Definitions
 */

// ==================== Enums / Union Types ====================

export type OrderStatus =
  | 'PENDING'
  | 'PAID'
  | 'RECHARGING'
  | 'COMPLETED'
  | 'EXPIRED'
  | 'CANCELLED'
  | 'FAILED'
  | 'REFUND_REQUESTED'
  | 'REFUNDING'
  | 'REFUND_PENDING'
  | 'PARTIALLY_REFUNDED'
  | 'REFUNDED'
  | 'REFUND_FAILED'

export type PaymentType = 'alipay' | 'wxpay' | 'alipay_direct' | 'wxpay_direct' | 'stripe' | 'easypay' | 'airwallex'

export type OrderType = 'balance' | 'subscription'

// ==================== Configuration ====================

export interface PaymentConfig {
  payment_enabled: boolean
  min_amount: number
  max_amount: number
  daily_limit: number
  max_pending_orders: number
  order_timeout_minutes: number
  balance_disabled: boolean
  /** Deprecated response field. The effective recharge conversion is always 1:1. */
  balance_recharge_multiplier: number
  /** Deprecated response field retained while older servers are supported. */
  subscription_usd_to_cny_rate: number
  enabled_payment_types: PaymentType[]
  help_image_url: string
  help_text: string
  stripe_publishable_key: string
}

export interface MethodLimit {
  currency?: string
  display_name?: string
  daily_limit: number
  daily_used: number
  daily_remaining: number
  single_min: number
  single_max: number
  fee_rate: number
  available: boolean
}

/** Response from /payment/limits API */
export interface MethodLimitsResponse {
  methods: Record<string, MethodLimit>
  global_min: number  // widest min across all methods; 0 = no minimum
  global_max: number  // widest max across all methods; 0 = no maximum
}

export interface RechargeBonusTier {
  threshold_cny: number
  bonus_points: number
}

export type RefundStatus =
  | 'REQUESTED'
  | 'RESERVED'
  | 'SUBMITTING'
  | 'PENDING'
  | 'SUCCEEDED'
  | 'FAILED'

export interface RefundQuote {
  order_id: string
  currency: 'CNY' | string
  requested_principal_amount: number
  principal_amount: number
  fee_amount: number
  refund_fee_rate: number
  refund_fee_amount: number
  gateway_amount: number
  base_points: number
  bonus_points: number
  points_to_hold: number
  bonus_expired_offset: number
  affiliate_rebate_points: number
  remaining_principal_amount: number
  max_refundable_principal_amount: number
  refund_deadline?: string
  self_service_eligible: boolean
  requires_ticket: boolean
  blocked_reason?: string
}

export interface PaymentRefund {
  id: string
  order_id: string
  ticket_id?: string
  source: 'SELF_SERVICE' | 'TICKET' | 'ADMIN' | string
  status: RefundStatus | string
  currency: 'CNY' | string
  requested_principal_amount: number
  principal_amount: number
  fee_amount: number
  refund_fee_rate: number
  refund_fee_amount: number
  gateway_amount: number
  base_points: number
  bonus_points: number
  bonus_expired_offset: number
  affiliate_rebate_points: number
  provider_refund_id?: string
  reason: string
  error_code?: string
  error_message?: string
  submitted_at?: string
  settled_at?: string
  created_at: string
}

/** Response from /payment/checkout-info API — single call for the payment page */
export interface CheckoutInfoResponse {
  methods: Record<string, MethodLimit>
  global_min: number
  global_max: number
  plans: SubscriptionPlan[]
  balance_disabled: boolean
  /** Deprecated compatibility field. Recharge conversion is fixed at 1:1. */
  balance_recharge_multiplier: number
  /** Fixed point bonuses; the highest qualifying CNY threshold wins. */
  recharge_bonus_tiers?: RechargeBonusTier[]
  /** Deprecated compatibility field. Subscription plans are point-denominated. */
  subscription_usd_to_cny_rate: number
  recharge_fee_rate: number
  help_text: string
  help_image_url: string
  stripe_publishable_key: string
  /** When true, Alipay payments on mobile always show the QR code instead of redirecting */
  alipay_force_qrcode?: boolean
  /** When true, official Alipay mobile orders use precreate plus an Alipay app deep link */
  alipay_mobile_precreate_deep_link?: boolean
  wallet: WalletSummary
}

export interface WalletSummary {
  balance: number
  available_balance: number
  recharge_balance: number
  bonus_balance: number
  overdraft_amount: number
  frozen_balance: number
  frozen_recharge_balance: number
  frozen_bonus_balance: number
  total_balance: number
  next_bonus_expires_at?: string
  next_expiring_bonus_amount: number
}

// ==================== Orders ====================

export interface PaymentOrder {
  id: string
  user_id: string
  amount: number
  pay_amount: number
  /** Base-currency recharge principal (CNY for new recharge orders). */
  principal_amount?: number
  /** Base-currency payment fee. */
  fee_amount?: number
  /** Points credited from the recharge principal at the fixed 1:1 ratio. */
  base_points?: number
  /** Promotional points credited in addition to the base points. */
  bonus_points?: number
  /** Total points credited for the order. */
  credited_points?: number
  /** Affiliate points generated by this order. */
  affiliate_rebate_points?: number
  wallet_amount: number
  wallet_bonus_amount: number
  wallet_recharge_amount: number
  gateway_base_amount: number
  wallet_only: boolean
  currency?: string
  fee_rate: number
  payment_type: string
  out_trade_no: string
  status: OrderStatus
  order_type: OrderType
  created_at: string
  expires_at: string
  paid_at?: string
  completed_at?: string
  refund_amount: number
  refunded_principal_amount?: number
  refunded_fee_amount?: number
  refunded_gateway_amount?: number
  reversed_base_points?: number
  reversed_bonus_points?: number
  reversed_affiliate_points?: number
  refund_deadline?: string
  refund_reason?: string
  refund_requested_at?: string
  refund_requested_by?: string
  refund_request_reason?: string
  plan_id?: string
  provider_instance_id?: string
  bonus_tier_snapshot?: { threshold_cny?: number; bonus_points?: number }
  bonus_expires_at?: string
  bonus_grant_id?: string
}

// ==================== Plans & Channels ====================

export interface SubscriptionPlan {
  id: string
  group_id: string
  group_platform?: string
  group_name?: string
  rate_multiplier?: number
  peak_rate_enabled?: boolean
  peak_start?: string
  peak_end?: string
  peak_rate_multiplier?: number
  daily_limit_usd?: number | null
  weekly_limit_usd?: number | null
  monthly_limit_usd?: number | null
  /** Point-denominated aliases used by the platform-points API contract. */
  daily_limit_points?: number | null
  weekly_limit_points?: number | null
  monthly_limit_points?: number | null
  supported_model_scopes?: string[]
  name: string
  description: string
  price: number
  original_price?: number
  price_points?: number
  original_price_points?: number
  /** Display-only ISO 4217 currency label (e.g. "NZD"); empty means no label */
  currency?: string
  validity_days: number
  validity_unit: string
  /** Stored as JSON string in backend; API layer should parse before use */
  features: string[]
  for_sale: boolean
  sort_order: number
  stock_enabled?: boolean
  stock_quantity?: number | null
  stock_frozen?: number
  stock_available?: number | null
}

export interface PaymentChannel {
  id: string
  group_id?: string
  name: string
  platform: string
  rate_multiplier: number
  description: string
  models: string[]
  features: string[]
  enabled: boolean
}

// ==================== Providers ====================

export interface ProviderInstance {
  id: string
  provider_key: string
  name: string
  config: Record<string, string>
  supported_types: string[]
  enabled: boolean
  payment_mode: string
  refund_enabled: boolean
  allow_user_refund: boolean
  limits: string
  sort_order: number
}

// ==================== Request / Response ====================

export interface CreateOrderRequest {
  amount: number
  payment_type: string
  order_type: string
  plan_id?: string
  return_url?: string
  payment_source?: string
  openid?: string
  wechat_resume_token?: string
  is_mobile?: boolean
  use_balance?: boolean
}

export type CreateOrderResultType = 'order_created' | 'oauth_required' | 'jsapi_ready'

export interface WechatOAuthInfo {
  authorize_url?: string
  appid?: string
  openid?: string
  scope?: string
  state?: string
  redirect_url?: string
}

export interface WechatJSAPIPayload {
  appId?: string
  timeStamp?: string
  nonceStr?: string
  package?: string
  signType?: string
  paySign?: string
}

export interface CreateOrderResult {
  order_id: string
  amount: number
  principal_amount?: number
  fee_amount?: number
  base_points?: number
  bonus_points?: number
  credited_points?: number
  affiliate_rebate_points?: number
  pay_url?: string
  qr_code?: string
  client_secret?: string
  intent_id?: string
  currency?: string
  country_code?: string
  payment_env?: string
  pay_amount: number
  wallet_amount: number
  wallet_bonus_amount: number
  wallet_recharge_amount: number
  gateway_base_amount: number
  wallet_only: boolean
  fee_rate: number
  expires_at: string
  result_type?: CreateOrderResultType
  payment_type?: string
  out_trade_no?: string
  payment_mode?: string
  resume_token?: string
  alipay_mobile_precreate_deep_link?: boolean
  oauth?: WechatOAuthInfo
  jsapi?: WechatJSAPIPayload
  jsapi_payload?: WechatJSAPIPayload
  activation_status?: 'active' | 'pending'
  pending_subscription?: import('./index').PendingSubscription
}

export type CurrencyAmounts = Record<string, number>

export interface DailyPaymentStats {
  date: string
  amount: CurrencyAmounts
  count: number
}

export interface PaymentMethodStats {
  type: string
  amount: CurrencyAmounts
  count: number
}

export interface TopUserPaymentStats {
  user_id: string
  email: string
  amount: number
}

export interface DashboardStats {
  today_amount: CurrencyAmounts
  total_amount: CurrencyAmounts
  today_count: number
  total_count: number
  avg_amount: CurrencyAmounts
  daily_series: DailyPaymentStats[]
  payment_methods: PaymentMethodStats[]
  top_users: Record<string, TopUserPaymentStats[]>
}
