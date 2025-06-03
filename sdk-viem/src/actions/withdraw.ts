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
  Hex,
  SendTransactionParameters,
  SendTransactionReturnType,
  Transport,
  UnionEvaluate,
  UnionOmit,
  zeroAddress,
} from "viem";
import { GetAccountParameter } from "../types/account";
import { parseAccount } from "viem/utils";
import { readContract, sendTransaction } from "viem/actions";
import { getContractsAddressesByChainId } from "@consensys/linea-sdk-core";

export type WithdrawParameters<
  chain extends Chain | undefined = Chain | undefined,
  account extends Account | undefined = Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
> = UnionEvaluate<UnionOmit<FormattedTransactionRequest<derivedChain>, "data" | "to" | "from">> &
  Partial<GetChainParameter<chain, chainOverride>> &
  Partial<GetAccountParameter<account>> & {
    token: Address;
    to: Address;
    amount: bigint;
    data?: Hex;
  };

export type WithdrawReturnType = SendTransactionReturnType;

export async function withdraw<
  chain extends Chain | undefined,
  account extends Account | undefined,
  chainOverride extends Chain | undefined = Chain | undefined,
  derivedChain extends Chain | undefined = DeriveChain<chain, chainOverride>,
>(
  client: Client<Transport, chain, account>,
  parameters: WithdrawParameters<chain, account, chainOverride, derivedChain>,
) {
  const { account: account_ = client.account, token, amount, to, data, ...tx } = parameters;

  const account = account_ ? parseAccount(account_) : client.account;
  if (!account) {
    throw new BaseError("Account is required to send a transaction");
  }

  const chainId = client.chain?.id;

  if (!chainId) {
    throw new BaseError("No chain id found");
  }

  const l2MessageServiceAddress = getContractsAddressesByChainId(chainId).messageService;

  const minimumFeeInWei = await readContract(client, {
    address: l2MessageServiceAddress,
    abi: [
      {
        inputs: [],
        name: "minimumFeeInWei",
        outputs: [{ internalType: "uint256", name: "", type: "uint256" }],
        stateMutability: "view",
        type: "function",
      },
    ],
    functionName: "minimumFeeInWei",
  });

  if (token === zeroAddress) {
    return sendTransaction(client, {
      to: l2MessageServiceAddress,
      value: amount + minimumFeeInWei,
      account,
      data: encodeFunctionData({
        abi: [
          {
            inputs: [
              { internalType: "address", name: "_to", type: "address" },
              { internalType: "uint256", name: "_fee", type: "uint256" },
              { internalType: "bytes", name: "_calldata", type: "bytes" },
            ],
            name: "sendMessage",
            outputs: [],
            stateMutability: "payable",
            type: "function",
          },
        ],
        functionName: "sendMessage",
        args: [to, minimumFeeInWei, data ?? "0x"],
      }),
      ...tx,
    } as SendTransactionParameters);
  }

  const tokenBridgeAddress = getContractsAddressesByChainId(chainId).tokenBridge;

  return sendTransaction(client, {
    to: tokenBridgeAddress,
    value: minimumFeeInWei,
    account,
    data: encodeFunctionData({
      abi: [
        {
          inputs: [
            {
              internalType: "address",
              name: "_token",
              type: "address",
            },
            {
              internalType: "uint256",
              name: "_amount",
              type: "uint256",
            },
            {
              internalType: "address",
              name: "_recipient",
              type: "address",
            },
          ],
          name: "bridgeToken",
          outputs: [],
          stateMutability: "payable",
          type: "function",
        },
      ],
      functionName: "bridgeToken",
      args: [token, amount, to],
    }),
    ...tx,
  } as SendTransactionParameters);
}
