import fs from "fs";
import { config, network as hardhatNetwork } from "hardhat";
import path from "path";

import {
  tryVerifyContract,
  getRequiredEnvVar,
  requireAddressFromRegistryOrEnv,
  LogContractDeployment,
} from "../common/helpers";
import { abi, bytecode } from "./V1/L2MessageServiceV1Deployed.json";
import { deployScript } from "../rocketh/deploy";
import { getUiSigner, withSignerUiSession } from "../scripts/hardhat/signer-ui-bridge";
import { deployUpgradableWithAbiAndByteCode } from "../scripts/hardhat/utils";

const hardhatConnection = await hardhatNetwork.getOrCreate();
const networkName = hardhatConnection.networkName === "default" ? "hardhat" : hardhatConnection.networkName;

const func = withSignerUiSession("04_deploy_L2MessageServiceV1.ts", async function () {
  const mainnetDeployedL2MessageServiceCacheFolder = path.resolve("./deploy/V1/L2MessageServiceV1Cache/");

  const validationFilePath = path.join(config.paths.cache, "validations.json");
  const validationFileBackupPath = path.join(config.paths.cache, "validations_backup.json");

  if (fs.existsSync(validationFilePath)) {
    fs.copyFileSync(validationFilePath, validationFileBackupPath);
  }

  fs.copyFileSync(path.join(mainnetDeployedL2MessageServiceCacheFolder, "validations.json"), validationFilePath);

  const contractName = "L2MessageServiceLineaMainnet";

  const L2MessageService_securityCouncil = requireAddressFromRegistryOrEnv(
    networkName,
    "L2_SECURITY_COUNCIL",
    "L2_SECURITY_COUNCIL",
  );
  const L2MessageService_l1l2MessageSetter = requireAddressFromRegistryOrEnv(
    networkName,
    "L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER",
    "L2_MESSAGE_SERVICE_L1L2_MESSAGE_SETTER",
  );
  const L2MessageService_rateLimitPeriod = getRequiredEnvVar("L2_MESSAGE_SERVICE_RATE_LIMIT_PERIOD");
  const L2MessageService_rateLimitAmount = getRequiredEnvVar("L2_MESSAGE_SERVICE_RATE_LIMIT_AMOUNT");

  const deployer = await getUiSigner();

  const contract = await deployUpgradableWithAbiAndByteCode(
    deployer,
    "L2MessageServiceLineaMainnet",
    JSON.stringify(abi),
    bytecode,
    [
      L2MessageService_securityCouncil,
      L2MessageService_l1l2MessageSetter,
      L2MessageService_rateLimitPeriod,
      L2MessageService_rateLimitAmount,
    ],
    {
      initializer: "initialize(address,address,uint256,uint256)",
      unsafeAllow: ["constructor"],
    },
  );

  await LogContractDeployment(contractName, contract);
  const contractAddress = await contract.getAddress();

  await tryVerifyContract(contractAddress);

  fs.unlinkSync(path.join(config.paths.cache, "validations.json"));
  if (fs.existsSync(validationFileBackupPath)) {
    fs.copyFileSync(validationFileBackupPath, validationFilePath);
    fs.unlinkSync(validationFileBackupPath);
  }
});
export default deployScript(func, {
  tags: ["L2MessageServiceLineaMainnet"],
});
