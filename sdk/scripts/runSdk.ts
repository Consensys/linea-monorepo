import * as dotenv from "dotenv";
import { LineaSDK } from "../src";

dotenv.config();

async function main() {
  const sdk = new LineaSDK({
    l1RpcUrlOrProvider: process.env.L1_RPC_URL ?? "",
    l2RpcUrlOrProvider: process.env.L2_RPC_URL ?? "",
    network: "linea-sepolia",
    mode: "read-only",
  });

  const l2Contract = sdk.getL2Contract();
  const data = l2Contract.encodeClaimMessageTransactionData({
    messageSender: "0x5eEeA0e70FFE4F5419477056023c4b0acA016562",
    destination: "0x5eEeA0e70FFE4F5419477056023c4b0acA016562",
    fee: 0n,
    value: 100000000000000000n,
    calldata: "0x",
    messageNonce: 3105n,
    messageHash: "",
  });
  console.log(data);
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
