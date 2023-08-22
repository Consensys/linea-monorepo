import { DatabaseErrorType, DatabaseRepoName } from "./enums";
import { MessageInDb } from "./types";

export class DatabaseAccessError<T extends MessageInDb> extends Error {
  public rejectedMessage?: T;

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  constructor(name: DatabaseRepoName, type: DatabaseErrorType, e: any, rejectedMessage?: T) {
    super(e);
    Error.captureStackTrace(this, this.constructor);
    this.message = `${name}: ${type} for message - ${e.message || e}`;
    this.name = DatabaseAccessError.name;
    this.rejectedMessage = rejectedMessage;
  }
}
