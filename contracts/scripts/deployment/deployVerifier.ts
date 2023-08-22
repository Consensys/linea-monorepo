import { run } from "hardhat";
import { delay } from "../../utils/storeAddress";
import { deployFromFactory } from "../hardhat/utils";

/*
    *******************************************************************************************
    1. Set the VERIFIER_CONTRACT_NAME - e.g PlonkeVerifyFull
    *******************************************************************************************
    NB: use the verifier.address output as input for scripts/deployment/setVerifierAddress.ts 
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/deployment/deployVerifier.ts
    *******************************************************************************************
*/

async function main() {
  const verifierFull = await deployFromFactory("PlonkVerifierFull");
  console.log(`PlonkVerifierFull deployed at ${verifierFull.address}`);

  const verifierFullLarge = await deployFromFactory("PlonkVerifierFullLarge");
  console.log(`PlonkVerifierFullLarge deployed at ${verifierFullLarge.address}`);

  console.log(`Waiting for 2 minutes before verifying`)
  await delay(120_000);

  await run("verify", {
    address: verifierFull.address,
  });

  await run("verify", {
    address: verifierFullLarge.address,
  });
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
