import { upgrades as createUpgrades } from "@openzeppelin/hardhat-upgrades";
import { LineaRollup__factory } from "contracts/typechain-types";
import hre, { network as hardhatNetwork } from "hardhat";

import { tryVerifyContract, getRequiredEnvVar, requireAddressFromRegistryOrEnv } from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;
const upgrades = await createUpgrades(hre, hardhatConnection);

const func = withSignerUiSession("03_deploy_LineaRollupV8WithReinitialization.ts", async function () {
  const signer = await getUiSigner();

  const proxyAddress = requireAddressFromRegistryOrEnv(networkName, "LineaRollup", "LINEA_ROLLUP_ADDRESS");
  const forcedTransactionFeeInWei = getRequiredEnvVar("LINEA_ROLLUP_FORCED_TRANSACTION_FEE_IN_WEI");
  const addressFilter = requireAddressFromRegistryOrEnv(networkName, "AddressFilter", "LINEA_ROLLUP_ADDRESS_FILTER");

  const contractName = "LineaRollup";

  const factory = await ethers.getContractFactory(contractName, signer);

  console.log("Deploying new LineaRollup implementation...");
  const newImplementation = await upgrades.deployImplementation(factory, {
    kind: "transparent",
  });

  const implementationAddress = newImplementation.toString();
  console.log(`Implementation deployed at ${implementationAddress}`);

  // Encoded calldata for upgradeAndCall via ProxyAdmin (selector 0x9623609d).
  // Submit this through the Security Council Safe using upgradeAndCall on the ProxyAdmin.
  // See: https://www.4byte.directory/signatures/?bytes4_signature=0x9623609d
  const upgradeCallWithReinitialization = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [
        proxyAddress,
        newImplementation,
        LineaRollup__factory.createInterface().encodeFunctionData("reinitializeLineaRollupV9", [
          forcedTransactionFeeInWei,
          addressFilter,
        ]),
      ],
    ),
  ]);

  console.log("Encoded upgradeAndCall calldata for reinitializeLineaRollupV9:");
  console.log("\n", upgradeCallWithReinitialization, "\n");

  await tryVerifyContract(implementationAddress);
});

export default deployScript(func, {
  tags: ["LineaRollupV8WithReinitialization"],
});
