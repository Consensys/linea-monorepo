import { ethers } from "hardhat";
import { LineaRollup } from "../../typechain-types";

async function main() {
  // TESTING ON L1

  const factory = await ethers.getContractFactory("LineaRollup");
  const proxyContract = factory.attach("0x41B186Dc7C46f08ADFdCe21Da1b07f605819E9Ab") as LineaRollup;

  const verifier0 = await proxyContract.verifiers(0);
  const verifier1 = await proxyContract.verifiers(1);
  const verifier2 = await proxyContract.verifiers(2);

  console.log({ verifier0, verifier1, verifier2 });

  const limitInWei = await proxyContract.limitInWei();
  const systemMigrationBlock = await proxyContract.systemMigrationBlock();

  console.log({ "systemMigrationBlock:": systemMigrationBlock, "limitInWei:": limitInWei });

  const outboxL1L2Mapping = await proxyContract.outboxL1L2MessageStatus(
    "0x0706629f2d6735d342c62fda7484482f77d9c2f53133e5b391d55b81a820ce27",
  );
  console.log("outboxL1L2Mapping :", outboxL1L2Mapping);

  const dataStartingBlock = await proxyContract.dataStartingBlock(
    "0x0000000000000000000000000000000000000000000000000000000000000001",
  );
  const dataEndingBlock = await proxyContract.dataEndingBlock(
    "0x0000000000000000000000000000000000000000000000000000000000000001",
  );

  console.log("dataStartingBlock :", dataStartingBlock);
  console.log("dataEndingBlock :", dataEndingBlock);
  console.log("L1 Test done");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
