/**
 * 错误码常量,与 backend pkg/apierr/codes.go 严格对应。
 * 修改时需同步两侧。
 */
export const Codes = {
  OK: 0,

  Internal: 10000,
  Database: 10001,
  Cache: 10002,
  UpstreamUnavail: 10003,

  NotLoggedIn: 20001,
  SessionExpired: 20002,
  InvalidToken: 20003,
  TokenExpired: 20004,
  IPNotAllowed: 20005,
  ModelNotAllowed: 20006,
  Forbidden: 20010,

  MissingParam: 30001,
  InvalidParam: 30002,

  EmailRegistered: 40001,
  UsernameTaken: 40002,
  WrongPassword: 40003,
  BalanceInsufficient: 40004,
  NoChannel: 40005,
  ModelNotSupported: 40006,
  OrderNotFound: 40007,
  RedeemInvalid: 40008,
  InviteInvalid: 40009,
  DeptBudgetExceeded: 40010,

  RateLimitUser: 50001,
  RateLimitToken: 50002,
  RateLimitIP: 50003,
  RateLimitGlobal: 50004,

  UpstreamError: 60001,
  UpstreamTimeout: 60002,
  UpstreamContentFilter: 60003,
} as const

export type Code = typeof Codes[keyof typeof Codes]
