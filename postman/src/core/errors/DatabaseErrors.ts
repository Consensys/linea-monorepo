import { BaseError } from "@consensys/linea-sdk";
import { MessageProps } from "../entities/Message";
import { DatabaseErrorType, DatabaseRepoName } from "../enums/DatabaseEnums";

export class DatabaseAccessError<T extends MessageProps> extends BaseError {
  override name = DatabaseAccessError.name;
  public rejectedMessage?: T;

  constructor(name: DatabaseRepoName, type: DatabaseErrorType, e: Error, rejectedMessage?: T) {
    super(`${name}: ${type} - ${e.message}`);
    this.rejectedMessage = rejectedMessage;
  }
}
