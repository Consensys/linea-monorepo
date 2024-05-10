import { ethers } from "hardhat";
import { LineaRollupInit__factory, LineaRollup__factory } from "../../typechain-types";

/*******************************USAGE******************************************************************
GOERLI_PRIVATE_KEY=<your_private_key> \
INFURA_API_KEY=<your_infura_key> \
npx hardhat run scripts/gnosis/encodingTX2.ts --network goerli

or

LINEA_GOERLI_PRIVATE_KEY=<your_private_key> \
INFURA_API_KEY=<your_infura_key> \
npx hardhat run scripts/gnosis/encodingTX2.ts --network linea_goerli
*******************************************************************************************************/

//--------------------------------------Config------------------------------------

const main = async () => {
  const initialL2BlockNumber = "1987654321";
  const initialStateRootHash = "0x3450000000000000000000000000000000000000000000000000000000000345";

  const proxyAdminContract = "0xd1A02bfB124F5e3970d46111586100E72e7B56bB";
  const proxyContract = "0x70BaD09280FD342D02fe64119779BC1f0791BAC2";
  const NewImplementation = "0xEAfCfCc565671042cFD5F9Db27914A77478f6066";

  const rollupSwitchoverInSeconds = 60;
  const blockTimeInSeconds = 12;

  const currentBlockNumber = await ethers.provider.getBlockNumber();
  const systemMigrationBlock = currentBlockNumber + Math.floor(+rollupSwitchoverInSeconds / +blockTimeInSeconds);
  // const systemMigrationBlock = 10431018;

  console.log("Encoded TX Output:");
  console.log("\n");
  console.log("currentBlockNumber:", currentBlockNumber);
  console.log("systemMigrationBlock:", systemMigrationBlock);
  console.log("\n");

  //-------------------------UpgradeAndCall Directly with Migration Block--------------------------

  const upgradeCallWithSystemMigrationBlockUsingSecurityCouncil = ethers.concat([
    "0x9623609d",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "address", "bytes"],
      [
        proxyContract,
        NewImplementation,
        LineaRollup__factory.createInterface().encodeFunctionData("initializeSystemMigrationBlock", [
          systemMigrationBlock,
        ]),
      ],
    ),
  ]);

  console.log(
    "Encoded upgradeAndCall with systemMigrationBlock set at block number:",
    systemMigrationBlock,
    "\n",
    upgradeCallWithSystemMigrationBlockUsingSecurityCouncil,
  );
  console.log("\n");

  //---------------------------Upgrade Directly------------------------------

  const upgradeCallUsingSecurityCouncil = ethers.concat([
    "0x99a88ec4",
    ethers.AbiCoder.defaultAbiCoder().encode(["address", "address"], [proxyContract, NewImplementation]),
  ]);

  console.log("Encoded Upgrade call (directly) from Security Council :", "\n", upgradeCallUsingSecurityCouncil);
  console.log("\n");

  //-----------------------Upgrade Directly with Reinitialization----------------------------------

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
    "Encoded upgradeAndCall (directly) with Reinitialization from Security Council :",
    "\n",
    upgradeCallWithReinitializationUsingSecurityCouncil,
  );
  console.log("\n");

  // ----------------------Additional config for Schedule/Execute-------------------------------

  const timelockDaysDelay = 7;
  const timelockDelay = timelockDaysDelay * 24 * 3600;
  console.log("Schedule/Execute with timelock delay of", timelockDaysDelay);
  console.log("\n");

  //-----------------------------Schedule with Migration Block----------------------------------
  const scheduleUpgradeSystemMigrationBlockCallwithZodiac = ethers.concat([
    "0x01d5062a",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32", "uint256"],
      [
        proxyAdminContract,
        0,
        upgradeCallWithSystemMigrationBlockUsingSecurityCouncil,
        ethers.ZeroHash,
        ethers.ZeroHash,
        timelockDelay,
      ],
    ),
  ]);

  console.log(
    "Encoding to be used for Upgrade at block ",
    systemMigrationBlock,
    "calling Timelock Schedule with",
    timelockDaysDelay,
    "days delay using Zodiac :",
    "\n",
    scheduleUpgradeSystemMigrationBlockCallwithZodiac,
  );
  console.log("\n");

  //-------------------------------Execute with Migration Block------------------------------------

  const executeUpgradeSystemMigrationBlockCallwithZodiac = ethers.concat([
    "0x134008d3",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32"],
      [
        proxyAdminContract,
        0,
        upgradeCallWithSystemMigrationBlockUsingSecurityCouncil,
        ethers.ZeroHash,
        ethers.ZeroHash,
      ],
    ),
  ]);

  console.log(
    "Encoding to be used for calling Timelock Execute function for Upgrade at block",
    systemMigrationBlock,
    "using Zodiac :",
    "\n",
    executeUpgradeSystemMigrationBlockCallwithZodiac,
  );
  console.log("\n");

  //----------------------------Schedule--------------------------------------

  const upgradeScheduleCallwithZodiac = ethers.concat([
    "0x01d5062a",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32", "uint256"],
      [proxyAdminContract, 0, upgradeCallUsingSecurityCouncil, ethers.ZeroHash, ethers.ZeroHash, timelockDelay],
    ),
  ]);

  console.log("Delay is set to:", timelockDelay);

  console.log(
    "Encoded schedule Upgrade using Zodiac with ",
    timelockDaysDelay,
    "day delay:",
    "\n",
    upgradeScheduleCallwithZodiac,
  );
  console.log("\n");

  // -------------------------------Execute------------------------------------------

  const upgradeExecuteCallwithZodiac = ethers.concat([
    "0x134008d3",
    ethers.AbiCoder.defaultAbiCoder().encode(
      ["address", "uint256", "bytes", "bytes32", "bytes32"],
      [proxyAdminContract, 0, upgradeCallUsingSecurityCouncil, ethers.ZeroHash, ethers.ZeroHash],
    ),
  ]);

  console.log("Encoded execute Upgrade using Zodiac", "\n", upgradeExecuteCallwithZodiac);
  console.log("\n");
};

main()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error(error);
    process.exit(1);
  });
