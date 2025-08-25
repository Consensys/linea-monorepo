import { HardhatRuntimeEnvironment } from "hardhat/types";
import { DeployFunction } from "hardhat-deploy/types";

const func: DeployFunction = async function (hre: HardhatRuntimeEnvironment) {
  const { deployments, getNamedAccounts } = hre;
  const { deploy } = deployments;
  const { deployer } = await getNamedAccounts();

  console.log("Deploying Status Network KarmaTiers contract...");
  console.log("Deployer:", deployer);

  // Deploy KarmaTiers contract
  const karmaTiers = await deploy("KarmaTiers", {
    from: deployer,
    args: [],
    log: true,
    waitConfirmations: 1,
  });

  console.log("KarmaTiers deployed to:", karmaTiers.address);

  // Verify the deployment
  if (karmaTiers.newlyDeployed) {
    console.log("✅ KarmaTiers contract deployed successfully");
  } else {
    console.log("ℹ️  KarmaTiers contract already deployed at:", karmaTiers.address);
  }
};

func.id = "deploy-status-network-karma-tiers";
func.tags = ["StatusNetworkKarmaTiers"];
func.dependencies = [];

export default func;
