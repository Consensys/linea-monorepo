import { Account, Chain, Client, Transport } from "viem";

export type LineaCancelBundleParameters = {
  replacementUUID: string;
};

export type LineaCancelBundleReturnType = boolean;

export async function lineaCancelBundle<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "linea_cancelBundle";
        Parameters: [LineaCancelBundleParameters["replacementUUID"]];
        ReturnType: LineaCancelBundleReturnType;
      },
    ]
  >,
  params: LineaCancelBundleParameters,
): Promise<LineaCancelBundleReturnType> {
  return client.request({
    method: "linea_cancelBundle",
    params: [params.replacementUUID],
  });
}
