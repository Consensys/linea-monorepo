import { run } from "hardhat";
import { delay } from "./general";

export async function tryVerifyContract(contractAddress: string) {
  if (process.env.VERIFY_CONTRACT) {
    console.log("Waiting 30 seconds for contract propagation...");
    await delay(30000);
    console.log("Etherscan verification ongoing...");
    // Verify contract
    try {
      await run("verify", {
        address: contractAddress,
      });
    } catch (err) {
      console.log(`Error happened during verification: ${err}`);
    }
    console.log("Etherscan verification done.");
  }
}

export async function tryVerifyContractWithConstructorArgs(
  contractAddress: string,
  contractForVerification: string,
  args: unknown[],
) {
  if (process.env.VERIFY_CONTRACT) {
    console.log("Waiting 30 seconds for contract propagation...");
    await delay(30000);
    console.log("Etherscan verification ongoing...");

    // Verify contract
    try {
      await run("verify:verify", {
        address: contractAddress,
        contract: contractForVerification,
        constructorArguments: args,
      });
    } catch (err) {
      console.log(`Error happened during verification: ${err}`);
    }
    console.log("Etherscan verification done.");
  }
}
