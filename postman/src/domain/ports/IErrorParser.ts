export enum ErrorCode {
  NETWORK_ERROR = "NETWORK_ERROR",
  NONCE_EXPIRED = "NONCE_EXPIRED",
  INSUFFICIENT_FUNDS = "INSUFFICIENT_FUNDS",
  GAS_FEE_ERROR = "GAS_FEE_ERROR",
  EXECUTION_REVERTED = "EXECUTION_REVERTED",
  ACTION_REJECTED = "ACTION_REJECTED",
  DATABASE_ERROR = "DATABASE_ERROR",
  UNKNOWN_ERROR = "UNKNOWN_ERROR",
}

export type Severity = "warn" | "error";

export type Mitigation = {
  shouldRetry: boolean;
};

export type ParsedErrorResult = {
  errorCode: ErrorCode;
  errorMessage: string;
  data?: string;
  severity: Severity;
  mitigation: Mitigation;
};

export interface IErrorParser {
  parse(error: unknown): ParsedErrorResult;
}
