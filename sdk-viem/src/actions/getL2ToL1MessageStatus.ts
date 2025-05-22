import { Account, BaseError, Chain, Client, Hex, Transport } from "viem";
import { OnChainMessageStatus } from "../types/message";
import { getMessageSentEvents } from "./getMessageSentEvents";
import { getContractEvents, readContract } from "viem/actions";
import { getBridgeContractAddresses } from "./getBridgeContractAddresses";

export type GetL2ToL1MessageStatusParameters<chain extends Chain | undefined, account extends Account | undefined> = {
  l2Client: Client<Transport, chain, account>;
  messageHash: Hex;
};

export type GetL2ToL1MessageStatusReturnType = OnChainMessageStatus;

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

  const [messageSentEvent] = await getMessageSentEvents(l2Client, { args: { _messageHash: messageHash } });

  if (!messageSentEvent) {
    throw new BaseError(`Message hash does not exist on L2. Message hash: ${messageHash}`);
  }

  const [[l2MessagingBlockAnchoredEvent], isMessageClaimed] = await Promise.all([
    getContractEvents(client, {
      address: getBridgeContractAddresses(client).lineaRollup,
      abi: [
        {
          anonymous: false,
          inputs: [
            {
              indexed: true,
              internalType: "uint256",
              name: "l2Block",
              type: "uint256",
            },
          ],
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
      address: getBridgeContractAddresses(client).lineaRollup,
      abi: [
        {
          inputs: [
            {
              internalType: "uint256",
              name: "_messageNumber",
              type: "uint256",
            },
          ],
          name: "isMessageClaimed",
          outputs: [
            {
              internalType: "bool",
              name: "isClaimed",
              type: "bool",
            },
          ],
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
