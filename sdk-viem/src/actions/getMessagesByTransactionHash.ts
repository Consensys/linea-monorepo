import { Account, BaseError, Chain, Client, Hex, parseEventLogs, toEventSignature, Transport } from "viem";
import { getTransactionReceipt } from "viem/actions";
import { ExtendedMessage } from "@consensys/linea-sdk-core";
import { LINEA_MESSAGE_SERVICE_CONTRACTS } from "../constants/address";

export type GetMessagesByTransactionHashParameters = {
  transactionHash: Hex;
};

export type GetMessagesByTransactionHashReturnType = ExtendedMessage[];

export async function getMessagesByTransactionHash<
  chain extends Chain | undefined,
  account extends Account | undefined,
>(
  client: Client<Transport, chain, account>,
  parameters: GetMessagesByTransactionHashParameters,
): Promise<GetMessagesByTransactionHashReturnType> {
  const { transactionHash } = parameters;

  const receipt = await getTransactionReceipt(client, { hash: transactionHash });
  const chainId = client.chain?.id;

  if (!chainId) {
    throw new BaseError("No chain id found in client");
  }

  const logs = receipt.logs.filter(
    (log) =>
      log.address === LINEA_MESSAGE_SERVICE_CONTRACTS[chainId] &&
      log.topics[0] === toEventSignature("MessageSent(address,address,uint256,uint256,uint256,bytes,bytes32)"),
  );

  const parsedLogs = parseEventLogs({
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
    logs: logs,
  });

  return parsedLogs.map((log) => ({
    from: log.args._from!,
    to: log.args._to!,
    fee: log.args._fee!,
    value: log.args._value!,
    nonce: log.args._nonce!,
    calldata: log.args._calldata!,
    messageHash: log.args._messageHash!,
    transactionHash: log.transactionHash,
    blockNumber: log.blockNumber,
  }));
}
