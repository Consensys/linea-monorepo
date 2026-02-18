export class BaseError extends Error {
  override name = "PostmanCoreError";

  constructor(message?: string) {
    super(message ?? "An error occurred.");
    Error.captureStackTrace(this, this.constructor);
  }
}
