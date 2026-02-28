import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export interface ImplementationForProxyParams {
  contractName: string;
}

const ImplementationForProxyModule = buildModule("ImplementationForProxy", (m) => {
  const contractName = m.getParameter<string>("contractName");

  const implementation = m.contract(contractName, [], {
    id: "NewImplementation",
  });

  return { implementation };
});

export default ImplementationForProxyModule;
