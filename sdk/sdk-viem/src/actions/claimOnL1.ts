import {
  Account,
  Address,
  BaseError,
  Chain,
  Client,
  DeriveChain,
  encodeFunctionData,
  FormattedTransactionRequest,
  GetChainParameter,
  SendTransactionParameters,
  SendTransactionReturnType,
  Transport,
  UnionEvaluate,
  UnionOmit,
  zeroAddress,
} from "viem";
import { GetAccountParameter } from "../types/account";
import { parseAccount } from "viem/utils";
import { sendTransaction } from "viem/actions";
import { getContractsAddressesByChainId, MessageProof } from "@consensys/linea-sdk-core";
import { Message } from "@consensys/linea-sdk-core/src/types/message";

export type ClaimOnL1Parameters<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
> = UnionEvaluate<UnionOmit<FormattedTransactionRequest<derivedChain>, "data" | "to" | "from">> &
  Partial<GetChainParameter<chain, chainOverride>> &
  Partial<GetAccountParameter<account>> &
  Omit<Message, "messageHash"> & { messageProof: MessageProof; feeRecipient?: Address };

export type ClaimOnL1ReturnType = SendTransactionReturnType;

export async function claimOnL1<
  chain extends Chain | undefined,
  account extends Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
>(
  client: Client<Transport, chain, account>,
  parameters: ClaimOnL1Parameters<chain, account, chainOverride, derivedChain>,
): Promise<ClaimOnL1ReturnType> {
  const {
    account: account_ = client.account,
    from,
    to,
    fee,
    value,
    nonce,
    calldata,
    feeRecipient,
    messageProof,
    ...tx
  } = parameters;

  const account = account_ ? parseAccount(account_) : client.account;
  if (!account) {
    throw new BaseError("Account is required to send a transaction");
  }

  const chainId = client.chain?.id;

  if (!chainId) {
    throw new BaseError("No chain id found in l1 client");
  }

  const lineaRollupAddress = getContractsAddressesByChainId(chainId).messageService;

  return sendTransaction(client, {
    to: lineaRollupAddress,
    account,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              components: [
                { internalType: "bytes32[]", name: "proof", type: "bytes32[]" },
                { internalType: "uint256", name: "messageNumber", type: "uint256" },
                { internalType: "uint32", name: "leafIndex", type: "uint32" },
                { internalType: "address", name: "from", type: "address" },
                { internalType: "address", name: "to", type: "address" },
                { internalType: "uint256", name: "fee", type: "uint256" },
                { internalType: "uint256", name: "value", type: "uint256" },
                { internalType: "address payable", name: "feeRecipient", type: "address" },
                { internalType: "bytes32", name: "merkleRoot", type: "bytes32" },
                { internalType: "bytes", name: "data", type: "bytes" },
              ],
              internalType: "struct IL1MessageService.ClaimMessageWithProofParams",
              name: "_params",
              type: "tuple",
            },
          ],
          name: "claimMessageWithProof",
          outputs: [],
          stateMutability: "nonpayable",
          type: "function",
        },
      ],
      functionName: "claimMessageWithProof",
      args: [
        {
          from,
          to,
          fee,
          value,
          feeRecipient: feeRecipient ?? zeroAddress,
          data: calldata,
          messageNumber: nonce,
          merkleRoot: messageProof.root,
          proof: messageProof.proof,
          leafIndex: messageProof.leafIndex,
        },
      ],
    }),
    ...tx,
  } as SendTransactionParameters);
}
