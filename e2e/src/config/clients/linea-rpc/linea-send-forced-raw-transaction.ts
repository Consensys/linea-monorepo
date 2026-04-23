import type { Client, Hash, Hex, Transport, Chain, Account } from "viem";

export type ForcedRawTransaction = {
  transaction: Hex;
  deadline: string;
};

export type LineaSendForcedRawTransactionRpc = {
  Method: "linea_sendForcedRawTransaction";
  Parameters: [ForcedRawTransaction[]];
  ReturnType: Hash[];
};

export async function lineaSendForcedRawTransaction(
  client: Client<Transport, Chain | undefined, Account | undefined, [LineaSendForcedRawTransactionRpc]>,
  params: ForcedRawTransaction[],
): Promise<Hash[]> {
  return client.request({
    method: "linea_sendForcedRawTransaction",
    params: [params],
  });
}
