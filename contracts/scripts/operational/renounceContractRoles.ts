import { requireEnv } from "../hardhat/utils";
import { ethers } from "hardhat";

/*
    *******************************************************************************************
    1. Set the OLD_ADMIN_ADDRESS to the EOA address
    2. Set the NEW_ADMIN_ADDRESS that you have just granted the roles to
    3. Set the PROXY_ADDRESS for the contract 
    4. Set the CONTRACT_TYPE of the proxy - e.g. ZkEvmV2
    5. Set the CONTRACT_ROLES comma separated, e.g "0x356a809dfdea9198dd76fb76bf6d403ecf13ea675eb89e1eda2db2c4a4676a26,0x1185e52d62bfbbea270e57d3d09733d221b53ab7a18bae82bb3c6c74bab16d82,0x0000000000000000000000000000000000000000000000000000000000000000"
    *******************************************************************************************
    NB: Be sure to have use the roles initially set to the security council EOA before changing

    DO NOT CALL THIS UNTIL YOU HAVE GIVEN THE NEW ADDRESS ALLL THE ROLES YOU ARE REVOKING

    MAKE SURE THAT THE DEFAULT ADMIN ROLE IS LAST AS IT IS REVOKING/RENOUNCING FROM SELF
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/operational/renounceContractRoles.ts
    *******************************************************************************************
*/

async function main() {
  const oldAdminAddress = requireEnv("OLD_ADMIN_ADDRESS");
  const newAdminAddress = requireEnv("NEW_ADMIN_ADDRESS");
  const proxyAddress = requireEnv("PROXY_ADDRESS");
  const contractType = requireEnv("CONTRACT_TYPE");
  const contractRoles = requireEnv("CONTRACT_ROLES");

  const contract = await ethers.getContractAt(contractType, proxyAddress);

  const rolesArray = contractRoles?.split(",");
  for (let i = 0; i < rolesArray.length; i++) {
    console.log(
      `Checking the new admin ${newAdminAddress} has ${rolesArray[i]} before revoking the old admin ${oldAdminAddress}`,
    );
    const newAdminHasRole = await contract.hasRole(rolesArray[i], newAdminAddress);
    if (newAdminHasRole) {
      console.log(`Revoking ${rolesArray[i]} from ${oldAdminAddress}`);
      const tx = await contract.renounceRole(rolesArray[i], oldAdminAddress);

      console.log("Waiting for transaction to process");
      await tx.wait();
    } else {
      console.log(`New admin does not have ${rolesArray[i]}, skipping`);
    }
  }

  console.log("Done");
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
