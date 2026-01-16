import { Chain, Account, Client, Transport } from "viem";

export type GetZkEVMBlockNumberReturnType = number;

export async function getZkEVMBlockNumber<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<
    Transport,
    chain,
    account,
    [
      {
        Method: "rollup_getZkEVMBlockNumber";
        Parameters: [];
        ReturnType: GetZkEVMBlockNumberReturnType;
      },
    ]
  >,
): Promise<GetZkEVMBlockNumberReturnType> {
  return client.request({
    method: "rollup_getZkEVMBlockNumber",
    params: [],
  });
}
