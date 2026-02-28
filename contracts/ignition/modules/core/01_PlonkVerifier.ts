import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

export interface PlonkVerifierParams {
  contractName: string;
  chainId: string;
  baseFee: string;
  coinbase: string;
  l2MessageServiceAddress: string;
}

const PlonkVerifierModule = buildModule("PlonkVerifier", (m) => {
  const contractName = m.getParameter<string>("contractName", "PlonkVerifier");
  const chainId = m.getParameter<string>("chainId");
  const baseFee = m.getParameter<string>("baseFee");
  const coinbase = m.getParameter<string>("coinbase");
  const l2MessageServiceAddress = m.getParameter<string>("l2MessageServiceAddress");

  const mimc = m.contract("Mimc", [], { id: "Mimc" });

  const constructorArgs = [
    [
      { value: chainId, name: "chainId" },
      { value: baseFee, name: "baseFee" },
      { value: coinbase, name: "coinbase" },
      { value: l2MessageServiceAddress, name: "l2MessageServiceAddress" },
    ],
  ];

  const verifier = m.contract(contractName, constructorArgs, {
    id: "PlonkVerifierContract",
    libraries: {
      Mimc: mimc,
    },
  });

  return { mimc, verifier };
});

export default PlonkVerifierModule;
