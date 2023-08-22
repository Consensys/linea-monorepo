import * as dotenv from "dotenv";
import { LineaSDK } from "../src/lib";

dotenv.config();

async function main() {
  const sdk = new LineaSDK({
    l1RpcUrl: process.env.L1_RPC_URL ?? "",
    l2RpcUrl: process.env.L2_RPC_URL ?? "",
    l1SignerPrivateKey: process.env.L1_SIGNER_PRIVATE_KEY ?? "",
    l2SignerPrivateKey: process.env.L2_SIGNER_PRIVATE_KEY ?? "",
    network: "linea-goerli",
    mode: "read-write",
  });

  const l1Contract = sdk.getL1Contract();
  const l2Contract = sdk.getL2Contract();

  console.log(await l2Contract.getMessageStatus("0x3e1bdceea0a4d5693af825a29d44fc41b7db7c4947121362a130326b82b84e65"));
  console.log(await l1Contract.getMessageStatus("0x28e9e11b53d624500f7610377c97877bb1ecb3127a88f7eba84dd7a146891946"));
  const message = await l2Contract.getMessageByMessageHash(
    "0xe36bb6d4122a2874692c7fa5bf189cfa6f80c77da0414d26b3a728b97aa18ee5",
  );
  const messageByTx = await l1Contract.getMessagesByTransactionHash(
    "0xeaeaa2f8bab82aa7d2d53770545399fe9783434bd8a53e5aa93abfadaa19df51",
  );

  const receipt = await l1Contract.getTransactionReceiptByMessageHash(
    "0x28e9e11b53d624500f7610377c97877bb1ecb3127a88f7eba84dd7a146891946",
  );

  console.log({ message, messageByTx, receipt });
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error("", error);
    process.exit(1);
  });
