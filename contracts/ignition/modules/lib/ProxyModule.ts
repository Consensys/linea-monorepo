import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export const ProxyAdminModule = buildModule("ProxyAdmin", (m) => {
  const proxyAdmin = m.contract("ProxyAdmin", [], {
    id: "ProxyAdmin",
  });

  return { proxyAdmin };
});

export default ProxyAdminModule;
