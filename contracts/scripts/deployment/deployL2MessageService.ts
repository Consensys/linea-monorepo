import { deployUpgradableFromFactory, requireEnv } from "../hardhat/utils";

/*
    *******************************************************************************************
    1. Set the L2MSGSERVICE_SECURITY_COUNCIL - e.g EOA or Safe
    2. Set the L2MSGSERVICE_L1L2_MESSAGE_SETTER for message hash anchoring
    3. Set the L2MSGSERVICE_RATE_LIMIT_PERIOD in Seconds
    4. Set the L2MSGSERVICE_RATE_LIMIT_AMOUNT in Wei
    *******************************************************************************************
    NB: use the verifier.address output as input for scripts/deployment/setVerifierAddress.ts 
    *******************************************************************************************
    npx hardhat run --network zkevm_dev scripts/deployment/deployL2MessageService.ts
    *******************************************************************************************
*/

async function main() {
  const L2MessageService_securityCouncil = requireEnv("L2MSGSERVICE_SECURITY_COUNCIL");
  const L2MessageService_l1l2MessageSetter = requireEnv("L2MSGSERVICE_L1L2_MESSAGE_SETTER");
  const L2MessageService_rateLimitPeriod = requireEnv("L2MSGSERVICE_RATE_LIMIT_PERIOD");
  const L2MessageService_rateLimitAmount = requireEnv("L2MSGSERVICE_RATE_LIMIT_AMOUNT");

  const L2implementation = await deployUpgradableFromFactory(
    "L2MessageService",
    [
      L2MessageService_securityCouncil,
      L2MessageService_l1l2MessageSetter,
      L2MessageService_rateLimitPeriod,
      L2MessageService_rateLimitAmount,
    ],
    {
      initializer: "initialize(address,address,uint256,uint256)",
      unsafeAllow: ["constructor"],
    },
  );
  console.log(`L2MessageService deployed at ${L2implementation.address}`);
}

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
