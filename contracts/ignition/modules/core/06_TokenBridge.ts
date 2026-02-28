import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";
import { ProxyAdminModule } from "../lib/ProxyModule.js";

export interface TokenBridgeInitParams {
  defaultAdmin: string;
  messageService: string;
  tokenBeacon: string;
  sourceChainId: bigint;
  targetChainId: bigint;
  remoteSender: string;
  reservedTokens: string[];
  roleAddresses: Array<{ addressWithRole: string; role: string }>;
  pauseTypeRoles: Array<{ pauseType: number; role: string }>;
  unpauseTypeRoles: Array<{ pauseType: number; role: string }>;
}

const TokenBridgeModule = buildModule("TokenBridge", (m) => {
  const { proxyAdmin } = m.useModule(ProxyAdminModule);

  const defaultAdmin = m.getParameter<string>("defaultAdmin");
  const messageService = m.getParameter<string>("messageService");
  const tokenBeacon = m.getParameter<string>("tokenBeacon");
  const sourceChainId = m.getParameter<bigint>("sourceChainId");
  const targetChainId = m.getParameter<bigint>("targetChainId");
  const remoteSender = m.getParameter<string>("remoteSender");
  const reservedTokens = m.getParameter<string[]>("reservedTokens", []);
  const roleAddresses = m.getParameter<Array<{ addressWithRole: string; role: string }>>("roleAddresses");
  const pauseTypeRoles = m.getParameter<Array<{ pauseType: number; role: string }>>("pauseTypeRoles");
  const unpauseTypeRoles = m.getParameter<Array<{ pauseType: number; role: string }>>("unpauseTypeRoles");

  const implementation = m.contract("TokenBridge", [], {
    id: "TokenBridgeImplementation",
  });

  const initializeData = m.encodeFunctionCall(implementation, "initialize", [
    {
      defaultAdmin,
      messageService,
      tokenBeacon,
      sourceChainId,
      targetChainId,
      remoteSender,
      reservedTokens,
      roleAddresses,
      pauseTypeRoles,
      unpauseTypeRoles,
    },
  ]);

  const proxy = m.contract(
    "TransparentUpgradeableProxy",
    [implementation, proxyAdmin, initializeData],
    {
      id: "TokenBridgeProxy",
    },
  );

  const tokenBridge = m.contractAt("TokenBridge", proxy, {
    id: "TokenBridge",
  });

  return { proxyAdmin, implementation, proxy, tokenBridge };
});

export default TokenBridgeModule;
