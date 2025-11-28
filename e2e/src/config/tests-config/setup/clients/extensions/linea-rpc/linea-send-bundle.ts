import { Client, Hash, Chain, Account, Transport } from "viem";

export type LineaSendBundleParameters = {
  txs: Hash[];
  replacementUUID: string;
  blockNumber: string;
};

export type LineaSendBundleReturnType = {
  bundleHash: Hash;
};

export async function lineaSendBundle<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_sendBundle";
        Parameters: [LineaSendBundleParameters];
        ReturnType: LineaSendBundleReturnType;
      },
    ]
  >,
  params: LineaSendBundleParameters,
): Promise<LineaSendBundleReturnType> {
  return client.request({
    method: "linea_sendBundle",
    params: [params],
  });
}
