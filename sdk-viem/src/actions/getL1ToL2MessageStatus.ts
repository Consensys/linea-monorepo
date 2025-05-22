import { Account, Chain, Client, Hex, Transport } from "viem";
import { readContract } from "viem/actions";
import { getBridgeContractAddresses } from "./getBridgeContractAddresses";
import { formatMessageStatus } from "../utils/message";
import { OnChainMessageStatus } from "../types/message";

export type GetL1ToL2MessageStatusReturnType = OnChainMessageStatus;

export type GetL1ToL2MessageStatusParameters = {
  messageHash: Hex;
};

export async function getL1ToL2MessageStatus<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<Transport, chain, account>,
  parameters: GetL1ToL2MessageStatusParameters,
): Promise<GetL1ToL2MessageStatusReturnType> {
  const { messageHash } = parameters;

  const { l2MessageService } = getBridgeContractAddresses(client);

  const status = await readContract(client, {
    address: l2MessageService,
    abi: [
      {
        inputs: [
          {
            internalType: "bytes32",
            name: "messageHash",
            type: "bytes32",
          },
        ],
        name: "inboxL1L2MessageStatus",
        outputs: [
          {
            internalType: "uint256",
            name: "messageStatus",
            type: "uint256",
          },
        ],
        stateMutability: "view",
        type: "function",
      },
    ],
    functionName: "inboxL1L2MessageStatus",
    args: [messageHash],
  });

  return formatMessageStatus(status);
}
