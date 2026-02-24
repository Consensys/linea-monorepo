import { ethers, upgrades } from "hardhat";
import { DeployFunction } from "hardhat-deploy/types";
import { tryVerifyContract, getRequiredEnvVar } from "../common/helpers";
import { LineaRollup__factory } from "contracts/typechain-types";

const func: DeployFunction = async function () {
  const contractName = "LineaRollup";

  const proxyAddress = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS");
  const forcedTransactionFeeInWei = getRequiredEnvVar("LINEA_ROLLUP_FORCED_TRANSACTION_FEE_IN_WEI");
  const addressFilter = getRequiredEnvVar("LINEA_ROLLUP_ADDRESS_FILTER");

  const factory = await ethers.getContractFactory(contractName);

  console.log("Deploying Contract...");
  const newContract = await upgrades.deployImplementation(factory, {
    kind: "transparent",
  });

  const contract = newContract.toString();

  console.log(`Contract deployed at ${contract}`);

  // The encoding should be used through the safe.
  // THIS IS JUST A SAMPLE AND WILL BE ADJUSTED WHEN NEEDED FOR GENERATING THE CALLDATA FOR THE UPGRADE CALL
  // https://www.4byte.directory/signatures/?bytes4_signature=0x9623609d
  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [
        proxyAddress,
        newContract,
        LineaRollup__factory.createInterface().encodeFunctionData("reinitializeLineaRollupV9", [
          forcedTransactionFeeInWei,
          addressFilter,
        ]),
      ],
    ),
  ]);

  console.log(
    "Encoded Tx Upgrade with Reinitialization from Security Council:",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");

  await tryVerifyContract(contract);
};

export default func;
func.tags = ["LineaRollupWithReinitialization"];
