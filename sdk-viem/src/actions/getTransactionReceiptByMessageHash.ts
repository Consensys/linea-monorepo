import { Account, BaseError, Chain, Client, GetTransactionReceiptReturnType, Hex, Transport } from "viem";
import { LINEA_MESSAGE_SERVICE_CONTRACTS } from "../constants/address";
import { getContractEvents, getTransactionReceipt } from "viem/actions";

export type GetTransactionReceiptByMessageHashParameters = {
  messageHash: Hex;
};

export type GetTransactionReceiptByMessageHashReturnType<chain extends Chain | undefined> =
  GetTransactionReceiptReturnType<chain>;

export async function getTransactionReceiptByMessageHash<
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<Transport, chain, account>,
  parameters: GetTransactionReceiptByMessageHashParameters,
): Promise<GetTransactionReceiptByMessageHashReturnType<chain>> {
  const { messageHash } = parameters;

  const chainId = client.chain?.id;

  if (!chainId) {
    throw new BaseError("No chain id found in client");
  }

  const [event] = await getContractEvents(client, {
    address: LINEA_MESSAGE_SERVICE_CONTRACTS[chainId],
    abi: [
      {
        anonymous: false,
        inputs: [
          { indexed: true, internalType: "address", name: "_from", type: "address" },
          { indexed: true, internalType: "address", name: "_to", type: "address" },
          { indexed: false, internalType: "uint256", name: "_fee", type: "uint256" },
          { indexed: false, internalType: "uint256", name: "_value", type: "uint256" },
          { indexed: false, internalType: "uint256", name: "_nonce", type: "uint256" },
          { indexed: false, internalType: "bytes", name: "_calldata", type: "bytes" },
          { indexed: true, internalType: "bytes32", name: "_messageHash", type: "bytes32" },
        ],
        name: "MessageSent",
        type: "event",
      },
    ] as const,
    eventName: "MessageSent",
    args: {
      _messageHash: messageHash,
    },
    fromBlock: "earliest",
    toBlock: "latest",
  });

  if (!event) {
    throw new BaseError(`Message with hash ${messageHash} not found.`);
  }

  const receipt = await getTransactionReceipt(client, { hash: event.transactionHash });

  return receipt;
}
