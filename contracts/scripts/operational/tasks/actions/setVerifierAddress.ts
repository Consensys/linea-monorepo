import type { NewTaskActionFunction } from "hardhat/types/tasks";
import { getTaskCliOrEnvValue } from "../../../../common/helpers/environmentHelper.js";

interface TaskArgs {
  proxyAddress?: string;
  contractType?: string;
  newVerifierAddress?: string;
  proofType?: string;
}

const action: NewTaskActionFunction<TaskArgs> = async (taskArgs, hre) => {
  const connection = await hre.network.connect();
  const { ethers } = connection;

  const proxyAddress = getTaskCliOrEnvValue(taskArgs, "proxyAddress", "PROXY_ADDRESS");
  const contractType = getTaskCliOrEnvValue(taskArgs, "contractType", "CONTRACT_TYPE");
  const newVerifierAddress = getTaskCliOrEnvValue(taskArgs, "newVerifierAddress", "NEW_VERIFIER_ADDRESS");
  const proofType = getTaskCliOrEnvValue(taskArgs, "proofType", "PROOF_TYPE");

  if (proxyAddress === undefined) {
    throw new Error("Please specify a proxy address e.g: --proxy-address 0x... or PROXY_ADDRESS=0x...");
  }

  if (contractType === undefined) {
    throw new Error("Please specify a contract type e.g: --contract-type LineaRollup or CONTRACT_TYPE=LineaRollup");
  }

  if (newVerifierAddress === undefined) {
    throw new Error(
      "Please specify a new verifier address e.g: --new-verifier-address 0x... or NEW_VERIFIER_ADDRESS=0x...",
    );
  }

  if (proofType === undefined) {
    throw new Error("Please specify a proof type e.g: --proof-type 1 or PROOF_TYPE=1");
  }

  const contract = await ethers.getContractAt(contractType, proxyAddress);

  console.log(`Setting verifier address to ${newVerifierAddress} for proof type ${proofType}`);
  const tx = await contract.setVerifierAddress(newVerifierAddress, proofType);
  console.log("Waiting for transaction to process");
  await tx.wait();

  console.log("Done");
};

export default action;
