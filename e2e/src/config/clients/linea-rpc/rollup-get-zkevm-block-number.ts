import type { Client, Transport, Chain, Account } from "viem";

export type GetZkEVMBlockNumberReturnType = number;

export type GetZkEVMBlockNumberRpc = {
  Method: "rollup_getZkEVMBlockNumber";
  Parameters: [];
  ReturnType: GetZkEVMBlockNumberReturnType;
};

export async function getZkEVMBlockNumber(
  client: Client<Transport, Chain | undefined, Account | undefined, [GetZkEVMBlockNumberRpc]>,
): Promise<GetZkEVMBlockNumberReturnType> {
  return client.request({
    method: "rollup_getZkEVMBlockNumber",
    params: [],
  });
}
