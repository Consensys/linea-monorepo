export type Mitigation = {
  shouldRetry: boolean;
  retryWithBlocking?: boolean;
  retryPeriodInMs?: number;
  retryNumOfTime?: number;
};

export type ParsedErrorResult = {
  errorCode: string;
  errorMessage?: string;
  data?: string;
  mitigation: Mitigation;
};

export interface IErrorParser {
  parseErrorWithMitigation(error: unknown): ParsedErrorResult | null;
}
