import * as dotenv from "dotenv";
import { LineaSDK, OnChainMessageStatus } from "../src";

dotenv.config();

async function main() {
  const sdk = new LineaSDK({
    l1RpcUrlOrProvider: process.env.L1_RPC_URL ?? "",
    l2RpcUrlOrProvider: process.env.L2_RPC_URL ?? "",
    l1SignerPrivateKeyOrWallet: process.env.L1_SIGNER_PRIVATE_KEY ?? "",
    l2SignerPrivateKeyOrWallet: process.env.L2_SIGNER_PRIVATE_KEY ?? "",
    network: "linea-mainnet",
    mode: "read-write",
  });

  const localL1ContractAddress = undefined; // need to be provided if network is "custom"
  const localL2ContractAddress = undefined; // need to be provided if network is "custom"

  const l1Contract = sdk.getL1Contract(localL1ContractAddress, localL2ContractAddress);
  const l2Contract = sdk.getL2Contract(localL2ContractAddress);

  // L2 -> L1 message, check its status on L1, and claim it on L1 if it's claimable
  const l2MessageHash = "0x9ab654972fbb4d513e76aea43cdb9bad1cb747bbfcebacd6b6f01bbaf0df28d9"; // from a real message sent on L2 and claimed on L1
  const l2MessageStatusOnL1 = await l1Contract.getMessageStatus(l2MessageHash);
  const l2MessageSentEvent = await l2Contract.getMessageByMessageHash(l2MessageHash);
  const l2MessageSentTxReceipt = await l2Contract.getTransactionReceiptByMessageHash(l2MessageHash);
  const [l2MessageSentEventByTxHash] =
    (await l2Contract.getMessagesByTransactionHash(l2MessageSentTxReceipt!.hash)) ?? [];
  if (l2MessageSentEvent?.messageHash !== l2MessageSentEventByTxHash.messageHash) {
    console.log("Something wrong here, the two message hashes should be equal");
  }
  console.log({ l2MessageStatusOnL1, l2MessageSentEvent, l2MessageSentTxReceipt });

  if (l2MessageStatusOnL1 === OnChainMessageStatus.CLAIMABLE) {
    const l1ClaimingService = sdk.getL1ClaimingService(localL1ContractAddress, localL2ContractAddress);
    const estimateClaimGas = await l1ClaimingService.estimateClaimMessageGas(l2MessageSentEvent!);
    console.log({ estimateClaimGas });

    const l1ClaimTxResponse = await l1ClaimingService.claimMessage(l2MessageSentEvent!, {
      maxPriorityFeePerGas: 100000000000n,
      maxFeePerGas: 100000000000n,
    });

    const l1ClaimTxReceipt = await l1ClaimTxResponse.wait();
    console.log({ l1ClaimTxReceipt });

    const finalL2MessageStatusOnL1 = await l1Contract.getMessageStatus(l2MessageHash);
    console.log({ finalL2MessageStatusOnL1 });
  }
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
