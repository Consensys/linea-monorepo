import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { BigNumber, ContractTransaction } from "ethers";
import { ethers } from "hardhat";
import { TestZkEvmV2 } from "../typechain-types";
import { BAD_STARTING_HASH, INITIAL_WITHDRAW_LIMIT, ONE_DAY_IN_SECONDS } from "./utils/constants";
import { deployUpgradableFromFactory } from "./utils/deployment";
import { getProverTestData, getTransactionsToBeDecoded } from "./utils/helpers";

describe("ZK EVM V2 contract with full verifier", () => {
  let zkEvm: TestZkEvmV2;
  let multiRollupZkEvm: TestZkEvmV2;

  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  let admin: SignerWithAddress;
  let verifier: string;
  let securityCouncil: SignerWithAddress;
  let operator: SignerWithAddress;

  const PROOF_MODE = "Full";

  const { proof, blocks, parentStateRootHash, firstBlockNumber } = getProverTestData(PROOF_MODE, "output-file.json");
  const {
    proof: proofRollup1,
    blocks: blocksRollup1,
    parentStateRootHash: parentStateRootHashRollup1,
    firstBlockNumber: firstBlockNumberRollup1,
  } = getProverTestData(PROOF_MODE, "rollup-1.json");
  const {
    proof: proofRollup2,
    blocks: blocksRollup2,
    parentStateRootHash: parentStateRootHashRollup2,
  } = getProverTestData(PROOF_MODE, "rollup-2.json");

  async function deployZkEvmFixture() {
    const PlonkVerifierFactory = await ethers.getContractFactory("PlonkVerifierFull");
    const plonkVerifier = await PlonkVerifierFactory.deploy();
    await plonkVerifier.deployed();

    verifier = plonkVerifier.address;

    const multiRollupZkEvm = (await deployUpgradableFromFactory(
      "TestZkEvmV2",
      [
        parentStateRootHashRollup1,
        firstBlockNumberRollup1 - 1,
        verifier,
        securityCouncil.address,
        [operator.address],
        ONE_DAY_IN_SECONDS,
        INITIAL_WITHDRAW_LIMIT,
      ],
      {
        initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256)",
        unsafeAllow: ["constructor"],
      },
    )) as TestZkEvmV2;

    const zkEvm = (await deployUpgradableFromFactory(
      "TestZkEvmV2",
      [
        parentStateRootHash,
        firstBlockNumber - 1,
        verifier,
        securityCouncil.address,
        [operator.address],
        ONE_DAY_IN_SECONDS,
        INITIAL_WITHDRAW_LIMIT,
      ],
      {
        initializer: "initialize(bytes32,uint256,address,address,address[],uint256,uint256)",
        unsafeAllow: ["constructor"],
      },
    )) as TestZkEvmV2;

    return { zkEvm, multiRollupZkEvm };
  }

  before(async () => {
    [admin, securityCouncil, operator] = await ethers.getSigners();
  });

  beforeEach(async () => {
    const contracts = await loadFixture(deployZkEvmFixture);
    zkEvm = contracts.zkEvm;
    multiRollupZkEvm = contracts.multiRollupZkEvm;
  });

  describe("When not paused", () => {
    describe("Multiple in a row with full proof", () => {
      it("Should fail when starting rootHash does not match last known block starting hash", async () => {
        await expect(
          multiRollupZkEvm
            .connect(operator)
            .finalizeBlocks(blocksRollup1, proof, 0, BAD_STARTING_HASH, { gasLimit: 10_000_000 }),
        ).to.be.revertedWithCustomError(multiRollupZkEvm, "StartingRootHashDoesNotMatch");
      });

      it("Should finalize multiple rollups and change the current timestamp", async () => {
        const { currentTimestampAfterSecondCall } = await finalizeMultipleBlocks();
        expect(currentTimestampAfterSecondCall.toNumber()).to.equal(
          blocksRollup2[blocksRollup2.length - 1].l2BlockTimestamp,
        );
      });

      it("Should finalize multiple rollups and change the current block number", async () => {
        const { initialCurrentBlockNumber, currentBlockNumberAfterSecondCall } = await finalizeMultipleBlocks();
        expect(currentBlockNumberAfterSecondCall).to.equal(
          initialCurrentBlockNumber.add(blocksRollup1.length).add(blocksRollup2.length),
        );
      });

      it("Should finalize multiple rollups and set the root hashes for each block", async () => {
        const { initialCurrentBlockNumber, currentBlockNumberAfterFirstCall } = await finalizeMultipleBlocks();

        const lastStateRootHashForFirstCall = await multiRollupZkEvm.stateRootHashes(
          initialCurrentBlockNumber.add(blocksRollup1.length),
        );
        expect(lastStateRootHashForFirstCall).to.equal(blocksRollup1[blocksRollup1.length - 1].blockRootHash);

        const lastStateRootHashForSecondCall = await multiRollupZkEvm.stateRootHashes(
          currentBlockNumberAfterFirstCall.add(blocksRollup2.length),
        );
        expect(lastStateRootHashForSecondCall).to.equal(blocksRollup2[blocksRollup2.length - 1].blockRootHash);
      });

      it("Should finalize multiple rollups and emit BlockFinalized", async () => {
        const {
          firstRollupTransaction,
          secondRollupTransaction,
          initialCurrentBlockNumber,
          currentBlockNumberAfterFirstCall,
        } = await finalizeMultipleBlocks();

        let blockNumber = initialCurrentBlockNumber.toNumber();
        let expectedEventData = blocksRollup1.map(({ blockRootHash }) => {
          blockNumber++;
          return { blockNumber, stateRootHash: blockRootHash };
        });

        const { events } = await firstRollupTransaction.wait();

        expect(events).to.not.be.undefined;

        if (events) {
          const filteredEvents = events.filter((event) => event.event === "BlockFinalized");
          expect(filteredEvents.length).to.be.equal(expectedEventData.length);

          // check for topic with non empty data
          for (let i = 0; i < filteredEvents.length; i++) {
            expect(filteredEvents[i].args?.blockNumber).to.deep.equal(expectedEventData[i].blockNumber);
            expect(filteredEvents[i].args?.stateRootHash).to.deep.equal(expectedEventData[i].stateRootHash);
          }
        }

        blockNumber = currentBlockNumberAfterFirstCall.toNumber();
        expectedEventData = blocksRollup2.map(({ blockRootHash }) => {
          blockNumber++;
          return { blockNumber, stateRootHash: blockRootHash };
        });

        const { events: eventsSet2 } = await secondRollupTransaction.wait();

        expect(eventsSet2).to.not.be.undefined;

        if (eventsSet2) {
          const filteredEvents = eventsSet2.filter((event) => event.event === "BlockFinalized");
          expect(filteredEvents.length).to.be.equal(expectedEventData.length);

          // check for topic with non empty data
          for (let i = 0; i < filteredEvents.length; i++) {
            expect(filteredEvents[i].args?.blockNumber).to.deep.equal(expectedEventData[i].blockNumber);
            expect(filteredEvents[i].args?.stateRootHash).to.deep.equal(expectedEventData[i].stateRootHash);
          }
        }
      });

      it("Should finalize multiple blocks and emit BlocksVerificationDone", async () => {
        const {
          firstRollupTransaction,
          secondRollupTransaction,
          currentBlockNumberAfterFirstCall,
          currentBlockNumberAfterSecondCall,
        } = await finalizeMultipleBlocks();

        await expect(firstRollupTransaction)
          .to.emit(multiRollupZkEvm, "BlocksVerificationDone")
          // blocks are zero based, based on initial data and this should be 1 less
          .withArgs(
            currentBlockNumberAfterFirstCall,
            parentStateRootHashRollup1,
            blocksRollup1[blocksRollup1.length - 1].blockRootHash,
          );

        await expect(secondRollupTransaction)
          .to.emit(multiRollupZkEvm, "BlocksVerificationDone")
          // blocks are zero based, based on initial data and this should be 1 less
          .withArgs(
            currentBlockNumberAfterSecondCall,
            parentStateRootHashRollup2,
            blocksRollup2[blocksRollup2.length - 1].blockRootHash,
          );
      });
    });

    describe("finalizeBlocks with full proof", () => {
      it("Should finalize blocks", async () => {
        const txHashes: string[][] = [];
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          const txs = await zkEvm.extractMessageHashes(tx);
          txHashes.push(txs);
          await zkEvm.addL1L2MessageHash(tx);
        }

        const tx = await zkEvm
          .connect(operator)
          .finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 });
        const { events } = await tx.wait();

        expect(events).to.not.be.undefined;

        if (events) {
          const filteredEvents = events.filter((event) => event.event === "L1L2MessagesReceivedOnL2");
          expect(filteredEvents.length).to.equal(txHashes.length);

          for (let i = 0; i < filteredEvents.length; i++) {
            expect(filteredEvents[i].args?.messageHashes).to.deep.equal(txHashes[i]);
          }
        }
      });

      it("Should finalize blocks and change currentL2BlockNumber", async () => {
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }

        const currentBlockNumberKnownBeforeExecution = await zkEvm.currentL2BlockNumber();
        expect(currentBlockNumberKnownBeforeExecution).to.equal(firstBlockNumber - 1);

        await zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 });

        const currentBlockNumberKnownAfterExecution = await zkEvm.currentL2BlockNumber();
        expect(currentBlockNumberKnownAfterExecution).to.equal(
          currentBlockNumberKnownBeforeExecution.add(blocks.length),
        );
      });

      it("Should finalize blocks and change currentTimestamp", async () => {
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }

        let currentTimestamp = await zkEvm.currentTimestamp();
        expect(currentTimestamp).to.equal(0);

        await zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 });

        currentTimestamp = await zkEvm.currentTimestamp();
        expect(currentTimestamp).to.equal(blocks[blocks.length - 1].l2BlockTimestamp);
      });

      it("Should finalize blocks and emit BlockFinalized", async () => {
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }

        let blockNumber = firstBlockNumber - 1;
        const expectedEventData = blocks.map(({ blockRootHash }) => {
          blockNumber++;
          return { blockNumber, stateRootHash: blockRootHash };
        });

        const tx = await zkEvm
          .connect(operator)
          .finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 });

        const { events } = await tx.wait();

        expect(events).to.not.be.undefined;

        if (events) {
          const filteredEvents = events.filter((event) => event.event === "BlockFinalized");
          expect(filteredEvents.length).to.be.equal(expectedEventData.length);

          // check for topic with non empty data
          for (let i = 0; i < filteredEvents.length; i++) {
            expect(filteredEvents[i].args?.blockNumber).to.deep.equal(expectedEventData[i].blockNumber);
            expect(filteredEvents[i].args?.stateRootHash).to.deep.equal(expectedEventData[i].stateRootHash);
          }
        }
      });

      it("Should finalize blocks and emit BlocksVerificationDone", async () => {
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }
        const previousKnownStartingNumber = await zkEvm.currentL2BlockNumber();
        await expect(
          zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 }),
        )
          .to.emit(zkEvm, "BlocksVerificationDone")
          // blocks are zero based, based on initial data and this should be 1 less
          .withArgs(
            previousKnownStartingNumber.add(blocks.length),
            parentStateRootHash,
            blocks[blocks.length - 1].blockRootHash,
          );
      });

      it("Should fail to process with future BlockTimeStamp", async () => {
        const { blocks } = getProverTestData(PROOF_MODE, "output-file.json");

        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }

        blocks[blocks.length - 1].l2BlockTimestamp = 3123456789;
        await expect(
          zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 }),
        ).to.be.revertedWithCustomError(zkEvm, "BlockTimestampError");
      });

      it.skip("Should fail to process with duplicate data", async () => {
        const { blocks } = getProverTestData(PROOF_MODE, "rollup-2.json");

        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }

        await expect(
          zkEvm
            .connect(operator)
            .finalizeBlocks([...blocks, ...blocks], proof, 0, parentStateRootHash, { gasLimit: 15_000_000 }),
        ).to.be.revertedWithCustomError(zkEvm, "MessageAlreadyReceived");
      });

      it.skip("Should fail when messages not marked as sent", async () => {
        await expect(
          zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 }),
        ).to.be.revertedWithCustomError(zkEvm, "L1L2MessageNotSent");
      });

      it("Should fail when proof does not match", async () => {
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }

        const data = getProverTestData(PROOF_MODE, "output-file.json");
        // remove a transaction breaking the hashes
        data.blocks[0].transactions.push(data.blocks[0].transactions[0]);

        await expect(
          zkEvm.connect(operator).finalizeBlocks(data.blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 }),
        ).to.be.revertedWithCustomError(zkEvm, "InvalidProof");
      });

      it("Should set the state hash for each block", async () => {
        for (const tx of getTransactionsToBeDecoded(blocks)) {
          await zkEvm.addL1L2MessageHash(tx);
        }
        const previousKnownStartingNumber = await zkEvm.currentL2BlockNumber();

        await zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, parentStateRootHash, { gasLimit: 10_000_000 });

        const blockHash = await zkEvm.stateRootHashes(previousKnownStartingNumber.add(blocks.length));
        expect(blockHash).to.equal(blocks[blocks.length - 1].blockRootHash);
      });

      it("Should fail when starting rootHash does not match last known block starting hash", async () => {
        await expect(
          zkEvm.connect(operator).finalizeBlocks(blocks, proof, 0, BAD_STARTING_HASH, { gasLimit: 10_000_000 }),
        ).to.be.revertedWithCustomError(zkEvm, "StartingRootHashDoesNotMatch");
      });
    });
  });

  async function finalizeMultipleBlocks(): Promise<{
    firstRollupTransaction: ContractTransaction;
    secondRollupTransaction: ContractTransaction;
    initialCurrentBlockNumber: BigNumber;
    currentBlockNumberAfterFirstCall: BigNumber;
    currentBlockNumberAfterSecondCall: BigNumber;
    initialCurrentTimestamp: BigNumber;
    currentTimestampAfterFirstCall: BigNumber;
    currentTimestampAfterSecondCall: BigNumber;
  }> {
    const txHashes: string[][] = [];

    for (const tx of getTransactionsToBeDecoded(blocksRollup1)) {
      const txs = await multiRollupZkEvm.extractMessageHashes(tx);
      txHashes.push(txs);
      await multiRollupZkEvm.addL1L2MessageHash(tx);
    }

    const initialCurrentBlockNumber = await multiRollupZkEvm.currentL2BlockNumber();
    expect(initialCurrentBlockNumber).to.equal(firstBlockNumberRollup1 - 1);

    const initialCurrentTimestamp = await multiRollupZkEvm.currentTimestamp();
    expect(initialCurrentTimestamp).to.equal(0);

    const firstRollupTransaction = await multiRollupZkEvm
      .connect(operator)
      .finalizeBlocks(blocksRollup1, proofRollup1, 0, parentStateRootHashRollup1, { gasLimit: 10_000_000 });

    const currentBlockNumberAfterFirstCall = await multiRollupZkEvm.currentL2BlockNumber();
    expect(currentBlockNumberAfterFirstCall).to.equal(initialCurrentBlockNumber.add(blocksRollup1.length));

    const currentTimestampAfterFirstCall = await multiRollupZkEvm.currentTimestamp();
    expect(currentTimestampAfterFirstCall).to.equal(blocksRollup1[blocksRollup1.length - 1].l2BlockTimestamp);

    for (const tx of getTransactionsToBeDecoded(blocksRollup2)) {
      const txs = await multiRollupZkEvm.extractMessageHashes(tx);
      txHashes.push(txs);
      await multiRollupZkEvm.addL1L2MessageHash(tx);
    }

    const secondRollupTransaction = await multiRollupZkEvm
      .connect(operator)
      .finalizeBlocks(blocksRollup2, proofRollup2, 0, parentStateRootHashRollup2, { gasLimit: 10_000_000 });

    const currentBlockNumberAfterSecondCall = await multiRollupZkEvm.currentL2BlockNumber();
    expect(currentBlockNumberAfterSecondCall).to.equal(currentBlockNumberAfterFirstCall.add(blocksRollup2.length));

    const currentTimestampAfterSecondCall = await multiRollupZkEvm.currentTimestamp();
    expect(currentTimestampAfterSecondCall).to.equal(blocksRollup2[blocksRollup2.length - 1].l2BlockTimestamp);

    return {
      firstRollupTransaction,
      secondRollupTransaction,
      initialCurrentBlockNumber,
      currentBlockNumberAfterFirstCall,
      currentBlockNumberAfterSecondCall,
      initialCurrentTimestamp,
      currentTimestampAfterFirstCall,
      currentTimestampAfterSecondCall,
    };
  }
});
