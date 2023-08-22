import { requireEnv } from "../hardhat/utils";
import { ethers } from "hardhat";
import { BigNumber } from "ethers";

/*
    *******************************************************************************************
    1. Deploy the verifier and get the address
    2. Run this script matching the correct PROOF_TYPE
    *******************************************************************************************

    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/operational/setVerifierAddress.ts
    *******************************************************************************************
*/

async function main() {
  const proofType = requireEnv("VERIFIER_PROOF"); // todo rename this once checking it doesn't break anything else
  const zkEvmAddress = requireEnv("ZKEVMV2_ADDRESS");
  const verifierAddress = requireEnv("VERIFIER_ADDRESS");

  const zkEvmV2 = await ethers.getContractAt("ZkEvmV2", zkEvmAddress);

  console.log(`Setting verifier address ${verifierAddress} of type ${proofType}`);
  const tx = await zkEvmV2.setVerifierAddress(verifierAddress, BigNumber.from(proofType));

  console.log("Waiting for transaction to process");
  await tx.wait();

  const checkVerifierIsSet = await zkEvmV2.verifiers(BigNumber.from(proofType));
  console.log(`ZkEvmV2 implementation added ${checkVerifierIsSet} as new verifier`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
