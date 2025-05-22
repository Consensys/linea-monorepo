import { createPublicClient, http } from "viem";
import { linea, mainnet } from "viem/chains";
import { publicActionsL2 } from "./decorators/publicL2";
import { publicActionsL1 } from "./decorators/publicL1";
import { getTransactionReceiptByMessageHash } from "./actions/getTransactionReceiptByMessageHash";

async function main() {
  const l2Client = createPublicClient({
    chain: linea,
    transport: http("https://linea-mainnet.infura.io/v3/9d60b7d314be4567adf4530f4b9dd801"),
  }).extend(publicActionsL2());

  const l1Client = createPublicClient({
    chain: mainnet,
    transport: http("https://mainnet.infura.io/v3/9d60b7d314be4567adf4530f4b9dd801"),
  }).extend(publicActionsL1());

  const proof = await l1Client.getMessageProof({
    messageHash: "0xfc89f3eb8c72aa9dd0456b3ad276442c65aafead3f2a7b55fd425a8a32ca8b0c",
    l2Client,
  });
  console.log("Message Proof:", proof);

  const status = await l2Client.getMessageStatus({
    messageHash: "0xd674cee8ebd16feed8c5f8b1c9e617460d7752f50b40f660f9bc925e66086860",
  });

  const l1Status = await l1Client.getMessageStatus({
    l2Client,
    messageHash: "0xdaf11ba86ecbed738c0ba2d82066997a361213b65a2c6b33ab769ba549d2bcd1",
  });

  console.log("Message Status:", status);
  console.log("L1 Message Status:", l1Status);

  const l1Message = await l1Client.getTransactionReceiptByMessageHash({
    messageHash: "0xd674cee8ebd16feed8c5f8b1c9e617460d7752f50b40f660f9bc925e66086860",
  });

  const l2Message = await getTransactionReceiptByMessageHash(l2Client, {
    messageHash: "0xdaf11ba86ecbed738c0ba2d82066997a361213b65a2c6b33ab769ba549d2bcd1",
  });

  console.log({
    l1Message,
    l2Message,
  });
}

main()
  .then(() => {
    process.exit(0);
  })
  .catch((error) => {
    console.error("Error:", error);
    process.exit(1);
  });
