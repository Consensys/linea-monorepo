import { Abi, Account, Address, BlockNumber, BlockTag, Chain, Client, ContractEventName, Transport } from "viem";
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
import { L1PublicClient } from "@consensys/linea-sdk-core";
import { StrictFunctionOnly } from "../types/misc";

export type PublicActionsL1<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
> = StrictFunctionOnly<
  L1PublicClient,
  {
    getMessageProof: <
      abi extends Abi | readonly unknown[] = Abi,
      eventName extends ContractEventName<abi> | undefined = ContractEventName<abi> | undefined,
      strict extends boolean | undefined = undefined,
      fromBlock extends BlockNumber | BlockTag | undefined = undefined,
      toBlock extends BlockNumber | BlockTag | undefined = undefined,
    >(
      args: GetMessageProofParameters<chain, account, abi, eventName, strict, fromBlock, toBlock>,
    ) => Promise<GetMessageProofReturnType>;
    getL2ToL1MessageStatus: <
      abi extends Abi | readonly unknown[] = Abi,
      eventName extends ContractEventName<abi> | undefined = ContractEventName<abi> | undefined,
      strict extends boolean | undefined = undefined,
      fromBlock extends BlockNumber | BlockTag | undefined = undefined,
      toBlock extends BlockNumber | BlockTag | undefined = undefined,
    >(
      args: GetL2ToL1MessageStatusParameters<chain, account, abi, eventName, strict, fromBlock, toBlock>,
    ) => Promise<GetL2ToL1MessageStatusReturnType>;
    getMessageByMessageHash: (args: GetMessageByMessageHashParameters) => Promise<GetMessageByMessageHashReturnType>;
    getMessagesByTransactionHash: (
      args: GetMessagesByTransactionHashParameters,
    ) => Promise<GetMessagesByTransactionHashReturnType>;
    getTransactionReceiptByMessageHash: <chain extends Chain | undefined>(
      args: GetTransactionReceiptByMessageHashParameters,
    ) => Promise<GetTransactionReceiptByMessageHashReturnType<chain>>;
  }
>;

export type PublicActionsL1Parameters = {
  lineaRollupAddress: Address;
  l2MessageServiceAddress: Address;
};

export function publicActionsL1(parameters?: PublicActionsL1Parameters) {
  return <
    chain extends Chain | undefined = Chain | undefined,
    account extends Account | undefined = Account | undefined,
  >(
    client: Client<Transport, chain, account>,
  ): PublicActionsL1<chain, account> => ({
    getMessageProof: (args) =>
      getMessageProof(client, {
        ...args,
        ...(parameters
          ? {
              lineaRollupAddress: parameters.lineaRollupAddress,
              l2MessageServiceAddress: parameters.l2MessageServiceAddress,
            }
          : {}),
      }),
    getL2ToL1MessageStatus: (args) =>
      getL2ToL1MessageStatus(client, {
        ...args,
        ...(parameters
          ? {
              lineaRollupAddress: parameters.lineaRollupAddress,
              l2MessageServiceAddress: parameters.l2MessageServiceAddress,
            }
          : {}),
      }),
    getMessageByMessageHash: (args) =>
      getMessageByMessageHash(client, {
        ...args,
        ...(parameters
          ? {
              messageServiceAddress: parameters.lineaRollupAddress,
            }
          : {}),
      }),
    getMessagesByTransactionHash: (args) =>
      getMessagesByTransactionHash(client, {
        ...args,
        ...(parameters
          ? {
              messageServiceAddress: parameters.lineaRollupAddress,
            }
          : {}),
      }),
    getTransactionReceiptByMessageHash: (args) =>
      getTransactionReceiptByMessageHash(client, {
        ...args,
        ...(parameters
          ? {
              messageServiceAddress: parameters.lineaRollupAddress,
            }
          : {}),
      }),
  });
}
