import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule";

const UpgradeableConsolidationQueuePredeployModule = buildModule("UpgradeableConsolidationQueuePredeploy", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const implementation = m.contract("UpgradeableConsolidationQueuePredeploy", [], {
    id: "UpgradeableConsolidationQueuePredeployImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initialize", []);

  const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initializeData], {
    id: "UpgradeableConsolidationQueuePredeployProxy",
  });

  const upgradeableConsolidationQueuePredeploy = m.contractAt("UpgradeableConsolidationQueuePredeploy", proxy, {
    id: "UpgradeableConsolidationQueuePredeploy",
  });

  return { proxyAdmin, implementation, proxy, upgradeableConsolidationQueuePredeploy };
});

export default UpgradeableConsolidationQueuePredeployModule;
