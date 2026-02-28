import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const ImplementationForProxyModule = buildModule("ImplementationForProxy", (m) => {
  const contractName = m.getParameter<string>("contractName");

  const implementation = m.contract(contractName, [], {
    id: "Implementation",
  });

  return { implementation };
});

export default ImplementationForProxyModule;
