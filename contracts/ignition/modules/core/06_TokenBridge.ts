import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const TokenBridgeModule = buildModule("TokenBridge", (m) => {
  const defaultAdmin = m.getParameter<string>("defaultAdmin");
  const messageService = m.getParameter<string>("messageService");
  const tokenBeacon = m.getParameter<string>("tokenBeacon");
  const sourceChainId = m.getParameter<number>("sourceChainId");
  const targetChainId = m.getParameter<number>("targetChainId");
  const remoteSender = m.getParameter<string>("remoteSender");
  const reservedTokens = m.getParameter<string[]>("reservedTokens", []);
  const roleAddresses = m.getParameter<unknown[]>("roleAddresses");
  const pauseTypeRoles = m.getParameter<unknown[]>("pauseTypeRoles");
  const unpauseTypeRoles = m.getParameter<unknown[]>("unpauseTypeRoles");

  const implementation = m.contract("TokenBridge", [], {
    id: "TokenBridge_Implementation",
  });

  const proxyAdmin = m.contract("src/_testing/integration/ProxyAdmin.sol:ProxyAdmin", [], {
    id: "TokenBridge_ProxyAdmin",
  });

  const initData = m.encodeFunctionCall(implementation, "initialize", [
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
    "src/_testing/integration/TransparentUpgradeableProxy.sol:TransparentUpgradeableProxy",
    [implementation, proxyAdmin, initData],
    { id: "TokenBridge_Proxy" },
  );

  return { proxy, proxyAdmin, implementation };
});

export default TokenBridgeModule;
