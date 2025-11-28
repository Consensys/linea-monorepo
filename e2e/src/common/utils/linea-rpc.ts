import { Client, Chain, Account, Transport } from "viem";
import { generateRandomUUIDv4 } from "./random";
import { lineaSendBundle } from "../../config/tests-config/setup/clients/extensions/linea-rpc/linea-send-bundle";

export async function isSendBundleMethodNotFound<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<Transport, chain, account>,
  targetBlockNumber = "0xffff",
) {
  try {
    await lineaSendBundle(client, {
      txs: [],
      replacementUUID: generateRandomUUIDv4(),
      blockNumber: targetBlockNumber,
    });
  } catch (err) {
    if (err instanceof Error) {
      if (err.message === "Method not found") {
        return true;
      }
    }
  }
  return false;
}
