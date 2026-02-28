import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule.js";

export interface RecoverFundsParams {
  securityCouncil: string;
  executorAddress: string;
}

const RecoverFundsModule = buildModule("RecoverFunds", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const securityCouncil = m.getParameter<string>("securityCouncil");
  const executorAddress = m.getParameter<string>("executorAddress");

  const implementation = m.contract("RecoverFunds", [], {
    id: "RecoverFundsImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initialize", [securityCouncil, executorAddress]);

  const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initializeData], {
    id: "RecoverFundsProxy",
  });

  const recoverFunds = m.contractAt("RecoverFunds", proxy, {
    id: "RecoverFunds",
  });

  return { proxyAdmin, implementation, proxy, recoverFunds };
});

export default RecoverFundsModule;
