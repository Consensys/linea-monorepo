import { ethers } from "hardhat";
import { OPERATOR_ROLE } from "../../test/common/constants";
import { LineaRollupInit__factory, LineaRollup__factory } from "../../typechain-types";

const main = async () => {
  const initialL2BlockNumber = "1987654321";
  const initialStateRootHash = "0x3450000000000000000000000000000000000000000000000000000000000345";

  const proxyAdminContract = "0xd1A02bfB124F5e3970d46111586100E72e7B56bB";
  const proxyContract = "0xA5d372Cc31C02E945949F8240716c420d8ECBe44";
  const NewImplementation = "0x1A4635Bd57705D72df06246947A78DAF959D6902";
  const accountGrantRevokeRole = "0xDdf06fce2C4230A99377E413B0553CDbdf39ef61";
  const pauseType = "0x0000000000000000000000000000000000000000000000000000000000000000";
  const rateLimitAmount = "10000000000";
  const minimumFee = "10000000000";

  const days = 7;
  const delay = days * 24 * 3600;

  console.log("Encoded TX Output:");
  console.log("\n");

  const upgradeCallWithReinitSystemMigrationBlock = LineaRollup__factory.createInterface().encodeFunctionData(
    "initializeSystemMigrationBlock",
    [1000000],
  );

  console.log("upgradeCallWithReinitSystemMigrationBlock", upgradeCallWithReinitSystemMigrationBlock);

  const upgradeCallUsingSecurityCouncil = ethers.concat([
    "0x99a88ec4",
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "address"], [proxyContract, NewImplementation]),
  ]);

  console.log("Encoded Tx Schedule Upgrade from Security Council :", "\n", upgradeCallUsingSecurityCouncil);
  console.log("\n");

  const upgradeScheduleCallwithZodiac = ethers.concat([
    "0x01d5062a",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32", "uint256"],
      [proxyAdminContract, 0, upgradeCallUsingSecurityCouncil, ethers.ZeroHash, ethers.ZeroHash, delay],
    ),
  ]);

  console.log("Delay is set to:", delay);

  console.log(
    "Encoded TX Schedule Upgrade using Zodiac with ",
    days,
    "day delay:",
    "\n",
    upgradeScheduleCallwithZodiac,
  );
  console.log("\n");

  const upgradeCallWithReinitializationUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [
        proxyContract,
        NewImplementation,
        LineaRollupInit__factory.createInterface().encodeFunctionData("initializeV2", [
          initialL2BlockNumber,
          initialStateRootHash,
        ]),
      ],
    ),
  ]);

  console.log(
    "Encoded Tx Upgrade with Reinitialization from Security Council :",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");

  const encodeGrantRole = ethers.concat([
    "0x2f2ff15d",
    ethers.AbiCoder.defaultAbiCoder().encode(["bytes32", "address"], [OPERATOR_ROLE, accountGrantRevokeRole]),
  ]);

  console.log("encodeGrantRole:", "\n", encodeGrantRole);
  console.log("\n");

  const encodeRevokeRole = ethers.concat([
    "0xd547741f",
    ethers.AbiCoder.defaultAbiCoder().encode(["bytes32", "address"], [OPERATOR_ROLE, accountGrantRevokeRole]),
  ]);

  console.log("encodeRevokeRole:", "\n", encodeRevokeRole);
  console.log("\n");

  const encodePauseByType = ethers.concat([
    "0x8264bd82",
    ethers.AbiCoder.defaultAbiCoder().encode(["bytes32"], [pauseType]),
  ]);

  console.log("encodePauseByType:", "\n", encodePauseByType);
  console.log("\n");

  const encodeResetLimitAmount = ethers.concat([
    "0x557eac73",
    ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [rateLimitAmount]),
  ]);

  console.log("encodeResetLimitAmount:", "\n", encodeResetLimitAmount);
  console.log("\n");

  const encodeSetMinimumFee = ethers.concat([
    "0x182a7506",
    ethers.AbiCoder.defaultAbiCoder().encode(["uint256"], [minimumFee]),
  ]);

  console.log("encodeSetMinimumFee:", "\n", encodeSetMinimumFee);
  console.log("\n");
};

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
