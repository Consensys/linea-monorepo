import * as dotenv from "dotenv";
import { LineaSDK } from "../src";
import { ZERO_ADDRESS } from "../src/core/constants";
import { Direction, MessageStatus } from "../src/core/enums/MessageEnums";

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
    feeRecipient: ZERO_ADDRESS,
    calldata: "0x",
    messageNonce: 3105n,
    messageHash: "",
    contractAddress: "",
    sentBlockNumber: 0,
    direction: Direction.L1_TO_L2,
    status: MessageStatus.SENT,
    claimNumberOfRetry: 0,
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
