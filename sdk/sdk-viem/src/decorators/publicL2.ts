import { Account, Chain, Client, Transport } from "viem";
import { L2PublicClient } from "@consensys/linea-sdk-core";
import {
  getBlockExtraData,
  GetBlockExtraDataParameters,
  GetBlockExtraDataReturnType,
} from "../actions/getBlockExtraData";
import {
  getL1ToL2MessageStatus,
  GetL1ToL2MessageStatusParameters,
  GetL1ToL2MessageStatusReturnType,
} from "../actions/getL1ToL2MessageStatus";
import {
  getMessageByMessageHash,
  GetMessageByMessageHashParameters,
  GetMessageByMessageHashReturnType,
} from "../actions/getMessageByMessageHash";
import {
  getMessagesByTransactionHash,
  GetMessagesByTransactionHashParameters,
  GetMessagesByTransactionHashReturnType,
} from "../actions/getMessagesByTransactionHash";
import {
  getTransactionReceiptByMessageHash,
  GetTransactionReceiptByMessageHashParameters,
  GetTransactionReceiptByMessageHashReturnType,
} from "../actions/getTransactionReceiptByMessageHash";
import { StrictFunctionOnly } from "../types/misc";

export type PublicActionsL2<chain extends Chain | undefined = Chain | undefined> = StrictFunctionOnly<
  L2PublicClient,
  {
    getBlockExtraData(args: GetBlockExtraDataParameters): Promise<GetBlockExtraDataReturnType>;
    getL1ToL2MessageStatus(args: GetL1ToL2MessageStatusParameters): Promise<GetL1ToL2MessageStatusReturnType>;
    getMessageByMessageHash(args: GetMessageByMessageHashParameters): Promise<GetMessageByMessageHashReturnType>;
    getMessagesByTransactionHash(
      args: GetMessagesByTransactionHashParameters,
    ): Promise<GetMessagesByTransactionHashReturnType>;
    getTransactionReceiptByMessageHash(
      args: GetTransactionReceiptByMessageHashParameters,
    ): Promise<GetTransactionReceiptByMessageHashReturnType<chain>>;
  }
>;

export function publicActionsL2() {
  return <
    transport extends Transport = Transport,
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<transport, chain, account>,
  ): PublicActionsL2<chain> => ({
    getBlockExtraData: (args) => getBlockExtraData(client, args),
    getL1ToL2MessageStatus: (args) => getL1ToL2MessageStatus(client, args),
    getMessageByMessageHash: (args) => getMessageByMessageHash(client, args),
    getMessagesByTransactionHash: (args) => getMessagesByTransactionHash(client, args),
    getTransactionReceiptByMessageHash: (args) => getTransactionReceiptByMessageHash(client, args),
  });
}
