import { Account, Chain, Client, Transport } from "viem";
import {
  getBlockExtraData,
  GetBlockExtraDataParameters,
  GetBlockExtraDataReturnType,
} from "../actions/getBlockExtraData";
import {
  getBridgeContractAddresses,
  GetBridgeContractAddressesReturnType,
} from "../actions/getBridgeContractAddresses";
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

export type PublicActionsL2 = {
  getBlockExtraData(args: GetBlockExtraDataParameters): Promise<GetBlockExtraDataReturnType>;
  getBridgeContractAddresses(): GetBridgeContractAddressesReturnType;
  getMessageStatus(args: GetL1ToL2MessageStatusParameters): Promise<GetL1ToL2MessageStatusReturnType>;
  getMessageByMessageHash(args: GetMessageByMessageHashParameters): Promise<GetMessageByMessageHashReturnType>;
  getMessagesByTransactionHash(
    args: GetMessagesByTransactionHashParameters,
  ): Promise<GetMessagesByTransactionHashReturnType>;
  getTransactionReceiptByMessageHash<chain extends Chain | undefined>(
    args: GetTransactionReceiptByMessageHashParameters,
  ): Promise<GetTransactionReceiptByMessageHashReturnType<chain>>;
};

export function publicActionsL2() {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): PublicActionsL2 => ({
    getBlockExtraData: (args) => getBlockExtraData(client, args),
    getBridgeContractAddresses: () => getBridgeContractAddresses(client),
    getMessageStatus: (args) => getL1ToL2MessageStatus(client, args),
    getMessageByMessageHash: (args) => getMessageByMessageHash(client, args),
    getMessagesByTransactionHash: (args) => getMessagesByTransactionHash(client, args),
    getTransactionReceiptByMessageHash: (args) => getTransactionReceiptByMessageHash(client, args),
  });
}
