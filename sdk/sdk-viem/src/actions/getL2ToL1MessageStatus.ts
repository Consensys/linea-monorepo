import { Account, BaseError, Chain, Client, Hex, Transport } from "viem";
import { getContractEvents, readContract } from "viem/actions";
import { getContractsAddressesByChainId, OnChainMessageStatus } from "@consensys/linea-sdk-core";
import { getMessageSentEvents } from "./getMessageSentEvents";

export type GetL2ToL1MessageStatusParameters<chain extends Chain | undefined, account extends Account | undefined> = {
  l2Client: Client<Transport, chain, account>;
  messageHash: Hex;
};

export type GetL2ToL1MessageStatusReturnType = OnChainMessageStatus;

/**
 * Returns the status of an L2 to L1 message on Linea.
 *
 * @returns The status of the L2 to L1 message.
 * @param client - Client to use
 * @param args - {@link GetL2ToL1MessageStatusParameters}
 *
 * @example
 * import { createPublicClient, http } from 'viem'
 * import { mainnet, linea } from 'viem/chains'
 * import { getL2ToL1MessageStatus } from '@consensys/linea-sdk-viem'
 *
 * const client = createPublicClient({
 *   chain: mainnet,
 *   transport: http(),
 * });
 *
 * const l2Client = createPublicClient({
 *  chain: linea,
 *  transport: http(),
 * });
 *
 * const messageStatus = await getL2ToL1MessageStatus(client, {
 *   l2Client,
 *   messageHash: '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
 * });
 */
export async function getL2ToL1MessageStatus<
  chain extends Chain | undefined,
  account extends Account | undefined,
  chainL2 extends Chain | undefined,
  accountL2 extends Account | undefined,
>(
  client: Client<Transport, chain, account>,
  parameters: GetL2ToL1MessageStatusParameters<chainL2, accountL2>,
): Promise<GetL2ToL1MessageStatusReturnType> {
  const { l2Client, messageHash } = parameters;

  if (!client.chain) {
    throw new BaseError("Client is required to get L2 to L1 message status.");
  }

  const [messageSentEvent] = await getMessageSentEvents(l2Client, { args: { _messageHash: messageHash } });

  if (!messageSentEvent) {
    throw new BaseError(`Message hash does not exist on L2. Message hash: ${messageHash}`);
  }
  const lineaRollupAddress = getContractsAddressesByChainId(client.chain.id).messageService;

  const [[l2MessagingBlockAnchoredEvent], isMessageClaimed] = await Promise.all([
    getContractEvents(client, {
      address: lineaRollupAddress,
      abi: [
        {
          anonymous: false,
          inputs: [{ indexed: true, internalType: "uint256", name: "l2Block", type: "uint256" }],
          name: "L2MessagingBlockAnchored",
          type: "event",
        },
      ] as const,
      eventName: "L2MessagingBlockAnchored",
      args: {
        l2Block: messageSentEvent.blockNumber,
      },
      fromBlock: "earliest",
      toBlock: "latest",
    }),
    readContract(client, {
      address: lineaRollupAddress,
      abi: [
        {
          inputs: [{ internalType: "uint256", name: "_messageNumber", type: "uint256" }],
          name: "isMessageClaimed",
          outputs: [{ internalType: "bool", name: "isClaimed", type: "bool" }],
          stateMutability: "view",
          type: "function",
        },
      ],
      functionName: "isMessageClaimed",
      args: [messageSentEvent.messageNonce],
    }),
  ]);

  if (isMessageClaimed) {
    return OnChainMessageStatus.CLAIMED;
  }

  if (l2MessagingBlockAnchoredEvent) {
    return OnChainMessageStatus.CLAIMABLE;
  }

  return OnChainMessageStatus.UNKNOWN;
}
