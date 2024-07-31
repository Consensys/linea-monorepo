import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import fs from "fs";
import { ethers } from "hardhat";
import path from "path";
import { TestLineaRollup } from "../typechain-types";
import {
  GENESIS_L2_TIMESTAMP,
  HASH_ZERO,
  INITIAL_WITHDRAW_LIMIT,
  ONE_DAY_IN_SECONDS,
  TEST_PUBLIC_VERIFIER_INDEX,
} from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import { generateFinalizationDataFromJSON, generateSubmissionDataFromJSON } from "./utils/helpers";

// This stores initialization data for the smart-contract
import { SignerWithAddress } from "@nomicfoundation/hardhat-ethers/signers";
import INITIALIZATION_DATA from "./testData/integrationWithProver/rolling-hash-history.json";

const submissionsDirectory = `${__dirname}/testData/integrationWithProver/blobSubmissions`;
const finalizationDirectory = `${__dirname}/../../testdata/prover/prover-aggregation/responses`;

function scanAndParseDir(directory: string): {
  start: number;
  stop: number;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  parsed: any;
}[] {
  const entries = fs.readdirSync(directory);
  const parsedList = [];

  for (const entry of entries) {
    const filepath = path.join(directory, entry);
    const dataTxt = fs.readFileSync(filepath, "utf-8");
    const parsed = JSON.parse(dataTxt);
    const blockRange = entry.split("-", 2);
    parsedList.push({
      start: parseInt(blockRange[0]),
      stop: parseInt(blockRange[1]),
      parsed: parsed,
    });
  }

  return parsedList;
}

describe("Linea Rollup  contract", () => {
  let lineaRollup: TestLineaRollup;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let verifier: string;
  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;

  async function deployLineaRollupFixture() {
    const PlonkVerifierFactory = await ethers.getContractFactory("TestPlonkVerifierForDataAggregation");
    const plonkVerifier = await PlonkVerifierFactory.deploy();
    await plonkVerifier.waitForDeployment();

    verifier = await plonkVerifier.getAddress();

    const lineaRollup = (await deployUpgradableFromFactory(
      "TestLineaRollup",
      [
        INITIALIZATION_DATA.initialParentStateRootHash,
        INITIALIZATION_DATA.initialFinalizedBlock,
        verifier,
        securityCouncil.address,
        [operator.address],
        ONE_DAY_IN_SECONDS,
        INITIAL_WITHDRAW_LIMIT,
        GENESIS_L2_TIMESTAMP,
      ],
      {
        initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256,uint256)",
        unsafeAllow: ["constructor"],
      },
    )) as unknown as TestLineaRollup;

    return lineaRollup;
  }

  before(async () => {
    [admin, securityCouncil, operator] = await ethers.getSigners();
  });

  beforeEach(async () => {
    lineaRollup = await loadFixture(deployLineaRollupFixture);
  });

  describe("Finalizing with the data generated from the prover's integration tests", () => {
    beforeEach(async () => {
      await lineaRollup.setupParentDataShnarf(
        INITIALIZATION_DATA.initialParentDataHash,
        INITIALIZATION_DATA.initialShnarf,
      );

      await lineaRollup.setupParentFinalizedStateRoot(
        INITIALIZATION_DATA.initialParentDataHash,
        INITIALIZATION_DATA.initialParentStateRootHash,
      );

      for (const event of INITIALIZATION_DATA.rollingHashHistory) {
        await lineaRollup.setRollingHash(event.messageNumber, event.rollingHash);
      }
      await lineaRollup.setLastTimeStamp(0);
    });

    //skipped as discussed with Alex
    it.skip(
      "Should successfully finalize with previous submission data and data submitted in finalization",
      async () => {
        const submissionsJSON = scanAndParseDir(submissionsDirectory);
        const finalizationsJSON = scanAndParseDir(finalizationDirectory);

        let index = 0;
        for (const submission of submissionsJSON) {
          const submissionContractData = generateSubmissionDataFromJSON(
            submission.start,
            submission.stop,
            submission.parsed,
          );

          if (index == 0) {
            // bypass first submission which is the expected behavior
            submissionContractData.dataParentHash = HASH_ZERO;
            submissionContractData.parentStateRootHash = HASH_ZERO;
          }

          index++;

          // console.log("sending submission", submission.start, submission.stop);
          // NB: we are only interested in the transaction passing. That's why we
          // don't bother checking the emitted events. Also, the events are
          // tested in other separate tests.
          await lineaRollup.connect(operator).submitData(submissionContractData, "0x", { gasLimit: 30_000_000 });
        }

        index = 0;

        for (const finalization of finalizationsJSON) {
          const finalizationData = generateFinalizationDataFromJSON(finalization.parsed);
          if (index == 0) {
            finalizationData.dataParentHash = HASH_ZERO;
          }

          index++;
          // console.log("sending finalization", finalization.start, finalization.stop);
          // NB: we are only interested in the transaction passing. That's why we
          // don't bother checking the emitted events. Also, the events are
          // tested in other separate tests.
          await lineaRollup
            .connect(operator)
            .finalizeCompressedBlocksWithProof(
              finalizationData.aggregatedProof,
              TEST_PUBLIC_VERIFIER_INDEX,
              finalizationData,
              {
                gasLimit: 30_000_000,
              },
            );
        }

        for (const finalization of finalizationsJSON) {
          const finalizationData = generateFinalizationDataFromJSON(finalization.parsed);
          // console.log("sending finalization", finalization.start, finalization.stop);

          // NB: we are only interested in the transaction passing. That's why we
          // don't bother checking the emitted events. Also, the events are
          // tested in other separate tests.
          await lineaRollup
            .connect(operator)
            .finalizeCompressedBlocksWithProof(
              finalizationData.aggregatedProof,
              TEST_PUBLIC_VERIFIER_INDEX,
              finalizationData,
              {
                gasLimit: 30_000_000,
              },
            );
        }
      },
    )
      // The test has an extended timeout period. Without that, it will fail on
      // the CI. This corresponds to 50sec as the timeout is given in ms.
      .timeout(50_000);
  });
});
