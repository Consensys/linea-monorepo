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
import { getContractsAddressesByChainId, Message } from "@consensys/linea-sdk-core";

export type ClaimOnL2Parameters<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
> = UnionEvaluate<UnionOmit<FormattedTransactionRequest<derivedChain>, "data" | "to" | "from">> &
  Partial<GetChainParameter<chain, chainOverride>> &
  Partial<GetAccountParameter<account>> &
  Omit<Message, "messageHash" | "nonce"> & { messageNonce: bigint; feeRecipient?: Address };

export type ClaimOnL2ReturnType = SendTransactionReturnType;

export async function claimOnL2<
  chain extends Chain | undefined,
  account extends Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
>(
  client: Client<Transport, chain, account>,
  parameters: ClaimOnL2Parameters<chain, account, chainOverride, derivedChain>,
): Promise<ClaimOnL2ReturnType> {
  const {
    account: account_ = client.account,
    from,
    to,
    fee,
    value,
    messageNonce,
    calldata,
    feeRecipient,
    ...tx
  } = parameters;

  const account = account_ ? parseAccount(account_) : client.account;
  if (!account) {
    throw new BaseError("Account is required to send a transaction");
  }

  const chainId = client.chain?.id;

  if (!chainId) {
    throw new BaseError("No chain id found in l2 client");
  }

  const l2MessageServiceAddress = getContractsAddressesByChainId(chainId).messageService;

  return sendTransaction(client, {
    to: l2MessageServiceAddress,
    account,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              internalType: "address",
              name: "_from",
              type: "address",
            },
            {
              internalType: "address",
              name: "_to",
              type: "address",
            },
            {
              internalType: "uint256",
              name: "_fee",
              type: "uint256",
            },
            {
              internalType: "uint256",
              name: "_value",
              type: "uint256",
            },
            {
              internalType: "address payable",
              name: "_feeRecipient",
              type: "address",
            },
            {
              internalType: "bytes",
              name: "_calldata",
              type: "bytes",
            },
            {
              internalType: "uint256",
              name: "_nonce",
              type: "uint256",
            },
          ],
          name: "claimMessage",
          outputs: [],
          stateMutability: "nonpayable",
          type: "function",
        },
      ],
      functionName: "claimMessage",
      args: [from, to, fee, value, feeRecipient ?? zeroAddress, calldata, messageNonce],
    }),
    ...tx,
  } as SendTransactionParameters);
}
