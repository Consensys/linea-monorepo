import { Message } from "./types";

export class FeeEstimationError extends Error {
  constructor(message: string) {
    super(message);
  }
}

export class GasEstimationError<T extends Message> extends Error {
  public rejectedMessage?: T;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  constructor(e: any, rejectedMessage?: T) {
    super(e);
    Error.captureStackTrace(this, this.constructor);
    this.rejectedMessage = rejectedMessage;
  }
}
