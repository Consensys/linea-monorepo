import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export const ProxyModule = buildModule("ProxyModule", (m) => {
  const proxyAdmin = m.contract("ProxyAdmin", [], {
    id: "ProxyAdmin",
    from: m.getAccount(0),
  });

  return { proxyAdmin };
});

export function createTransparentProxyModule(
  moduleId: string,
  implementationModuleId: string,
  implementationArtifact: string,
  initFunctionName: string,
  getInitArgs: (m: Parameters<Parameters<typeof buildModule>[1]>[0]) => unknown[],
  getConstructorArgs?: (m: Parameters<Parameters<typeof buildModule>[1]>[0]) => unknown[],
) {
  return buildModule(moduleId, (m) => {
    const constructorArgs = getConstructorArgs ? getConstructorArgs(m) : [];
    const implementation = m.contract(implementationArtifact, constructorArgs, {
      id: `${implementationModuleId}_Implementation`,
    });

    const proxyAdmin = m.contract("ProxyAdmin", [], {
      id: `${implementationModuleId}_ProxyAdmin`,
    });

    const initArgs = getInitArgs(m);
    const initData = m.encodeFunctionCall(implementation, initFunctionName, initArgs);

    const proxy = m.contract("TransparentUpgradeableProxy", [implementation, proxyAdmin, initData], {
      id: `${implementationModuleId}_Proxy`,
    });

    return { proxy, proxyAdmin, implementation };
  });
}
