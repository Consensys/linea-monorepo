import { BigNumber, Event } from "ethers";
import { MessageInDb } from "../postman/utils/types";
import { Direction, MessageStatus } from "../postman/utils/enums";
import { ParsedEvent } from "../utils/types";

export class EventParser {
  public static parseEvent<TEvent extends Event>(event: TEvent): ParsedEvent<TEvent> {
    return {
      args: event.args,
      blockNumber: event.blockNumber,
      logIndex: event.logIndex,
      contractAddress: event.address,
      transactionHash: event.transactionHash,
    };
  }

  public static filterAndParseEvents<TEvent extends Event>(events: TEvent[]): ParsedEvent<TEvent>[] {
    return events.filter((e) => e.removed === false).map(this.parseEvent);
  }

  public static parsedEventToMessage<TEvent extends ParsedEvent<Event>>(
    event: TEvent,
    direction: Direction,
    status: MessageStatus,
  ): MessageInDb & { logIndex: number } {
    return {
      messageSender: event.args?._from,
      destination: event.args?._to,
      fee: BigNumber.from(event.args?._fee).toString(),
      value: BigNumber.from(event.args?._value).toString(),
      messageNonce: BigNumber.from(event.args?._nonce).toNumber(),
      calldata: event.args?._calldata,
      messageHash: event.args?._messageHash,
      messageContractAddress: event.contractAddress,
      sentBlockNumber: event.blockNumber,
      logIndex: event.logIndex,
      claimNumberOfRetry: 0,
      direction,
      status,
    };
  }
}
