import { requireEnv } from "../hardhat/utils";
import { ethers } from "hardhat";

/*
    *******************************************************************************************
    1. Set the ADMIN_ADDRESS to the Safe address
    2. Set the PROXY_ADDRESS for the contract 
    3. Set the CONTRACT_TYPE of the proxy - e.g. ZkEvmV2
    4. Set the CONTRACT_ROLES comma separated, e.g "0x356a809dfdea9198dd76fb76bf6d403ecf13ea675eb89e1eda2db2c4a4676a26,0x1185e52d62bfbbea270e57d3d09733d221b53ab7a18bae82bb3c6c74bab16d82,0x0000000000000000000000000000000000000000000000000000000000000000"
    *******************************************************************************************
    NB: Be sure to have use the roles initially set to the security council EOA before changing
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/operational/grantContractRoles.ts
    *******************************************************************************************
*/

async function main() {
  const adminAddress = requireEnv("ADMIN_ADDRESS");
  const proxyAddress = requireEnv("PROXY_ADDRESS");
  const contractType = requireEnv("CONTRACT_TYPE");
  const contractRoles = requireEnv("CONTRACT_ROLES");

  const contract = await ethers.getContractAt(contractType, proxyAddress);

  const rolesArray = contractRoles?.split(",");
  for (let i = 0; i < rolesArray.length; i++) {
    console.log(`Granting ${rolesArray[i]} to ${adminAddress}`);
    const tx = await contract.grantRole(rolesArray[i], adminAddress);
    console.log("Waiting for transaction to process");
    await tx.wait();
  }

  console.log("Done");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
