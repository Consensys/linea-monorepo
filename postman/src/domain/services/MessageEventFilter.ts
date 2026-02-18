import { compileExpression, useDotAccessOperator } from "filtrex";

import { isEmptyBytes } from "../utils/isEmptyBytes";

import type { ICalldataDecoder } from "../ports/ICalldataDecoder";
import type { ILogger } from "../ports/ILogger";
import type { MessageSent } from "../types/blockchain";

export type MessageFilterConfig = {
  isEOAEnabled: boolean;
  isCalldataEnabled: boolean;
};

export class MessageEventFilter {
  constructor(
    private readonly calldataDecoder: ICalldataDecoder | null,
    private readonly config: MessageFilterConfig,
    private readonly logger: ILogger,
  ) {}

  public shouldProcess(
    event: MessageSent,
    calldataFilter?: {
      criteriaExpression: string;
      calldataFunctionInterface: string;
    },
  ): boolean {
    const hasEmptyCalldata = isEmptyBytes(event.calldata);
    const basicProcess = hasEmptyCalldata ? this.config.isEOAEnabled : this.config.isCalldataEnabled;

    if (!basicProcess) {
      this.logger.debug(
        "Message has been excluded because target address is not an EOA or calldata is not empty: messageHash=%s",
        event.messageHash,
      );
      return false;
    }

    if (!hasEmptyCalldata && this.config.isCalldataEnabled && !this.matchesCriteria(event, calldataFilter)) {
      return false;
    }

    return true;
  }

  private matchesCriteria(
    event: MessageSent,
    filters?: { criteriaExpression: string; calldataFunctionInterface: string },
  ): boolean {
    if (!filters || !this.calldataDecoder) {
      return true;
    }

    const decoded = this.calldataDecoder.decode(event.calldata);
    if (!decoded) {
      return false;
    }

    const context = {
      calldata: {
        funcSignature: event.calldata.slice(0, 10),
        ...this.convertBigInts(decoded.args),
      },
    };

    const passesFilter = this.evaluateExpression(filters.criteriaExpression, context);

    if (!passesFilter) {
      this.logger.debug(
        "Message has been excluded because it does not match the criteria: criteria=%s messageHash=%s transactionHash=%s",
        filters.criteriaExpression,
        event.messageHash,
        event.transactionHash,
      );
      return false;
    }

    return true;
  }

  private evaluateExpression(expression: string, context: unknown): boolean {
    try {
      const compiledFilter = compileExpression(expression, { customProp: useDotAccessOperator });
      const passesFilter = compiledFilter(context);
      return passesFilter === true;
    } catch {
      return false;
    }
  }

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  private convertBigInts(data: any): any {
    if (typeof data === "bigint") {
      return Number(data);
    }

    if (Array.isArray(data)) {
      return data.map((item) => this.convertBigInts(item));
    }

    if (data !== null && typeof data === "object") {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      const result: Record<string, any> = {};
      for (const key in data) {
        if (Object.prototype.hasOwnProperty.call(data, key)) {
          result[key] = this.convertBigInts(data[key]);
        }
      }
      return result;
    }

    return data;
  }
}
