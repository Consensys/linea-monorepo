import { buildModule } from "@nomicfoundation/hardhat-ignition/modules";

const PlonkVerifierModule = buildModule("PlonkVerifier", (m) => {
  const verifierContractName = m.getParameter<string>("verifierContractName");
  const chainId = m.getParameter<string>("chainId");
  const baseFee = m.getParameter<string>("baseFee");
  const coinbase = m.getParameter<string>("coinbase");
  const l2MessageServiceAddress = m.getParameter<string>("l2MessageServiceAddress");

  const mimc = m.contract("Mimc", [], { id: "Mimc" });

  const verifier = m.contract(
    verifierContractName,
    [
      [
        { value: chainId, name: "chainId" },
        { value: baseFee, name: "baseFee" },
        { value: coinbase, name: "coinbase" },
        { value: l2MessageServiceAddress, name: "l2MessageServiceAddress" },
      ],
    ],
    {
      id: "PlonkVerifier",
      libraries: { Mimc: mimc },
    },
  );

  return { mimc, verifier };
});

export default PlonkVerifierModule;
