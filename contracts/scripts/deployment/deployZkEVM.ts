import { deployFromFactory, deployUpgradableFromFactory, requireEnv } from "../hardhat/utils";

/*
    *******************************************************************************************
    1. Set the ZKEVMV2_INITIAL_STATE_ROOT_HASH
    2. Set the ZKEVMV2_INITIAL_L2_BLOCK_NUMBER
    3. Set the ZKEVMV2_SECURITY_COUNCIL 
    4. Set the ZKEVMV2_OPERATORS 
    5. Set the ZKEVMV2_RATE_LIMIT_PERIOD in Seconds
    6. Set the ZKEVMV2_RATE_LIMIT_AMOUNT in Wei
    *******************************************************************************************
    NB: use the verifier.address output as input for scripts/deployment/setVerifierAddress.ts 
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/deployment/deployZkEVM.ts
    *******************************************************************************************
*/

async function main() {
  const verifierContractName = process.env.VERIFIER_CONTRACT_NAME || "PlonkVerifier";

  // PLONK VERIFIER
  const verifier = await deployFromFactory(verifierContractName);
  console.log(`PlonkVerifier deployed at ${verifier.address}`);

  await deployZKEVM(verifier.address);
}

async function deployZKEVM(verifierAddress: string) {
  // ZKEVMV2 DEPLOYED AS UPGRADEABLE PROXY
  const ZkEvmV2_initialStateRootHash = requireEnv("ZKEVMV2_INITIAL_STATE_ROOT_HASH");
  const ZkEvmV2_initialL2BlockNumber = requireEnv("ZKEVMV2_INITIAL_L2_BLOCK_NUMBER");
  const ZKEVMV2_securityCouncil = requireEnv("ZKEVMV2_SECURITY_COUNCIL");
  const ZKEVMV2_operators = requireEnv("ZKEVMV2_OPERATORS");
  const ZKEVMV2_rateLimitPeriodInSeconds = requireEnv("ZKEVMV2_RATE_LIMIT_PERIOD");
  const ZKEVMV2_rateLimitAmountInWei = requireEnv("ZKEVMV2_RATE_LIMIT_AMOUNT");

  console.log(`Setting operators ${ZKEVMV2_operators}`);
  const zkEvmV2 = await deployUpgradableFromFactory(
    "ZkEvmV2",
    [
      ZkEvmV2_initialStateRootHash,
      ZkEvmV2_initialL2BlockNumber,
      verifierAddress,
      ZKEVMV2_securityCouncil,
      ZKEVMV2_operators?.split(","),
      ZKEVMV2_rateLimitPeriodInSeconds,
      ZKEVMV2_rateLimitAmountInWei,
    ],
    {
      initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256)",
      unsafeAllow: ["constructor"],
    },
  );

  console.log(`ZkEvmV2 deployed at ${zkEvmV2.address}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
