import { upgrades as createUpgrades } from "@openzeppelin/hardhat-upgrades";
import { RollupRevenueVault__factory } from "contracts/typechain-types";
import hre, { network as hardhatNetwork } from "hardhat";

import { tryVerifyContract, getRequiredEnvVar, requireAddressFromRegistryOrEnv } from "../common/helpers";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const { ethers } = hardhatConnection;
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;
const upgrades = await createUpgrades(hre, hardhatConnection);

const func = withSignerUiSession("18_deploy_RollupRevenueVaultWithReinitialization.ts", async function () {
  const signer = await getUiSigner();
  const contractName = "RollupRevenueVault";

  const proxyAddress = requireAddressFromRegistryOrEnv(
    networkName,
    "RollupRevenueVault",
    "ROLLUP_REVENUE_VAULT_ADDRESS",
  );
  const lastInvoiceDate = getRequiredEnvVar("ROLLUP_REVENUE_VAULT_LAST_INVOICE_DATE");
  const securityCouncil = requireAddressFromRegistryOrEnv(networkName, "L2_SECURITY_COUNCIL", "L2_SECURITY_COUNCIL");
  const invoiceSubmitter = requireAddressFromRegistryOrEnv(
    networkName,
    "ROLLUP_REVENUE_VAULT_INVOICE_SUBMITTER",
    "ROLLUP_REVENUE_VAULT_INVOICE_SUBMITTER",
  );
  const burner = requireAddressFromRegistryOrEnv(
    networkName,
    "ROLLUP_REVENUE_VAULT_BURNER",
    "ROLLUP_REVENUE_VAULT_BURNER",
  );
  const invoicePaymentReceiver = requireAddressFromRegistryOrEnv(
    networkName,
    "ROLLUP_REVENUE_VAULT_INVOICE_PAYMENT_RECEIVER",
    "ROLLUP_REVENUE_VAULT_INVOICE_PAYMENT_RECEIVER",
  );
  const tokenBridge = requireAddressFromRegistryOrEnv(networkName, "TokenBridge_L2", "TOKEN_BRIDGE_ADDRESS");
  const messageService = requireAddressFromRegistryOrEnv(networkName, "L2MessageService", "L2_MESSAGE_SERVICE_ADDRESS");
  const l1LineaTokenBurner = requireAddressFromRegistryOrEnv(
    networkName,
    "ROLLUP_REVENUE_VAULT_L1_LINEA_TOKEN_BURNER",
    "ROLLUP_REVENUE_VAULT_L1_LINEA_TOKEN_BURNER",
  );
  const lineaToken = requireAddressFromRegistryOrEnv(
    networkName,
    "ROLLUP_REVENUE_VAULT_LINEA_TOKEN",
    "ROLLUP_REVENUE_VAULT_LINEA_TOKEN",
  );
  const dexSwapAdapter = requireAddressFromRegistryOrEnv(
    networkName,
    "ROLLUP_REVENUE_VAULT_DEX_SWAP_ADAPTER",
    "ROLLUP_REVENUE_VAULT_DEX_SWAP_ADAPTER",
  );

  const factory = await ethers.getContractFactory(contractName, signer);

  console.log("Deploying Contract...");
  const newContract = await upgrades.deployImplementation(factory, {
    kind: "transparent",
  });

  const contractAddress = newContract.toString();

  console.log(`Contract deployed at ${contractAddress}`);

  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [
        proxyAddress,
        newContract,
        RollupRevenueVault__factory.createInterface().encodeFunctionData("initializeRolesAndStorageVariables", [
          lastInvoiceDate,
          securityCouncil,
          invoiceSubmitter,
          burner,
          invoicePaymentReceiver,
          tokenBridge,
          messageService,
          l1LineaTokenBurner,
          lineaToken,
          dexSwapAdapter,
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

  await tryVerifyContract(contractAddress, "src/operational/RollupRevenueVault.sol:RollupRevenueVault");
});

export default deployScript(func, {
  tags: ["RollupRevenueVaultWithReinitialization"],
});
