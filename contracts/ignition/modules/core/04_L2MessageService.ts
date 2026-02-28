import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule";

export interface L2MessageServiceInitParams {
  rateLimitPeriod: number;
  rateLimitAmount: string;
  securityCouncil: string;
  roleAddresses: Array<{ addressWithRole: string; role: string }>;
  pauseTypeRoles: Array<{ pauseType: number; role: string }>;
  unpauseTypeRoles: Array<{ pauseType: number; role: string }>;
}

const L2MessageServiceModule = buildModule("L2MessageService", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const rateLimitPeriod = m.getParameter<number>("rateLimitPeriod");
  const rateLimitAmount = m.getParameter<string>("rateLimitAmount");
  const securityCouncil = m.getParameter<string>("securityCouncil");
  const roleAddresses = m.getParameter<Array<{ addressWithRole: string; role: string }>>("roleAddresses");
  const pauseTypeRoles = m.getParameter<Array<{ pauseType: number; role: string }>>("pauseTypeRoles");
  const unpauseTypeRoles = m.getParameter<Array<{ pauseType: number; role: string }>>("unpauseTypeRoles");

  const implementation = m.contract("L2MessageService", [], {
    id: "L2MessageServiceImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initialize", [
    rateLimitPeriod,
    rateLimitAmount,
    securityCouncil,
    roleAddresses,
    pauseTypeRoles,
    unpauseTypeRoles,
  ]);

  const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initializeData], {
    id: "L2MessageServiceProxy",
  });

  const l2MessageService = m.contractAt("L2MessageService", proxy, {
    id: "L2MessageService",
  });

  return { proxyAdmin, implementation, proxy, l2MessageService };
});

export default L2MessageServiceModule;
