import { Account, Chain, Client, Transport } from "viem";
import { getMessageProof, GetMessageProofParameters, GetMessageProofReturnType } from "../actions/getMessageProof";
import {
  getL2ToL1MessageStatus,
  GetL2ToL1MessageStatusParameters,
  GetL2ToL1MessageStatusReturnType,
} from "../actions/getL2ToL1MessageStatus";
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

export type PublicActionsL1 = {
  getMessageProof<chain extends Chain | undefined, account extends Account | undefined>(
    args: GetMessageProofParameters<chain, account>,
  ): Promise<GetMessageProofReturnType>;
  getMessageStatus<chain extends Chain | undefined, account extends Account | undefined>(
    args: GetL2ToL1MessageStatusParameters<chain, account>,
  ): Promise<GetL2ToL1MessageStatusReturnType>;
  getMessageByMessageHash(args: GetMessageByMessageHashParameters): Promise<GetMessageByMessageHashReturnType>;
  getMessagesByTransactionHash(
    args: GetMessagesByTransactionHashParameters,
  ): Promise<GetMessagesByTransactionHashReturnType>;
  getTransactionReceiptByMessageHash<chain extends Chain | undefined>(
    args: GetTransactionReceiptByMessageHashParameters,
  ): Promise<GetTransactionReceiptByMessageHashReturnType<chain>>;
};

export function publicActionsL1() {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): PublicActionsL1 => ({
    getMessageProof: (args) => getMessageProof(client, args),
    getMessageStatus: (args) => getL2ToL1MessageStatus(client, args),
    getMessageByMessageHash: (args) => getMessageByMessageHash(client, args),
    getMessagesByTransactionHash: (args) => getMessagesByTransactionHash(client, args),
    getTransactionReceiptByMessageHash: (args) => getTransactionReceiptByMessageHash(client, args),
  });
}
