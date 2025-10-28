import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { SSZMerkleTree, TestValidatorContainerProofVerifier } from "contracts/typechain-types";
import { deployTestValidatorContainerProofVerifier, ValidatorWitness } from "../helpers";
import {
  ACTIVE_0X01_VALIDATOR_PROOF,
  generateBeaconHeader,
  generateEIP4478Witness,
  generateValidator,
  prepareLocalMerkleTree,
  randomBytes32,
  randomInt,
  setBeaconBlockRoot,
} from "../helpers/proof";
import { ethers } from "hardhat";
import {
  GI_FIRST_VALIDATOR_PREV,
  GI_FIRST_VALIDATOR_CURR,
  PIVOT_SLOT,
  SHARD_COMMITTEE_PERIOD,
  SLOTS_PER_EPOCH,
} from "../../common/constants";
import { expectRevertWithCustomError } from "../../common/helpers";

// TODO Constructor params

describe("ValidatorContainerProofVerifier", () => {
  let verifier: TestValidatorContainerProofVerifier;
  let sszMerkleTree: SSZMerkleTree;
  let firstValidatorLeafIndex: bigint;
  let lastValidatorIndex: bigint;
  let localTreeGiFirstValidator: GIndex;

  before(async () => {
    const localTree = await prepareLocalMerkleTree();
    localTreeGiFirstValidator = localTree.gIFirstValidator;
    sszMerkleTree = localTree.sszMerkleTree;
    firstValidatorLeafIndex = localTree.firstValidatorLeafIndex;
    // populate merkle tree with validators
    for (let i = 1; i < 100; i++) {
      await sszMerkleTree.addValidatorLeaf(generateValidator().container);
    }
    // after adding validators, all newly added validator indexes will +n from this
    lastValidatorIndex = (await sszMerkleTree.leafCount()) - 1n - firstValidatorLeafIndex;
  });

  beforeEach(async () => {
    verifier = await loadFixture(deployTestValidatorContainerProofVerifier);
    // test mocker
    const mockRoot = randomBytes32();
    const timestamp = await setBeaconBlockRoot(mockRoot);
    expect(await verifier.getParentBlockRoot(timestamp)).to.equal(mockRoot);
  });

  describe("constructor", () => {
    it("It should have the correct GI_FIRST_VALIDATOR_PREV_PREV", async () => {
      expect(await verifier.GI_FIRST_VALIDATOR_PREV()).eq(GI_FIRST_VALIDATOR_PREV);
    });
    it("It should have the correct GI_FIRST_VALIDATOR_PREV_CURR", async () => {
      expect(await verifier.GI_FIRST_VALIDATOR_CURR()).eq(GI_FIRST_VALIDATOR_CURR);
    });
    it("It should have the correct PIVOT_SLOT", async () => {
      expect(await verifier.PIVOT_SLOT()).eq(PIVOT_SLOT);
    });
  });

  it("should verify precalculated 0x01 validator object in merkle tree", async () => {
    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(
      ACTIVE_0X01_VALIDATOR_PROOF.witness.validator,
    );
    const beaconHeaderMerkle = await sszMerkleTree.getBeaconBlockHeaderProof(
      ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader,
    );
    const validatorGIndex = await verifier.getValidatorGI(ACTIVE_0X01_VALIDATOR_PROOF.witness.validatorIndex, 0);

    // Verify (ValidatorContainer) leaf against (StateRoot) Merkle root
    await sszMerkleTree.verifyProof(
      ACTIVE_0X01_VALIDATOR_PROOF.witness.proof,
      ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.stateRoot,
      validatorMerkle.root,
      validatorGIndex,
    );

    // Verify (StateRoot) leaf against (BeaconBlockRoot) Merkle root
    await sszMerkleTree.verifyProof(
      [...beaconHeaderMerkle.proof],
      beaconHeaderMerkle.root,
      ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.stateRoot,
      beaconHeaderMerkle.index,
    );

    // concatentate all proofs to match PG style
    const concatenatedProof = [...ACTIVE_0X01_VALIDATOR_PROOF.witness.proof, ...beaconHeaderMerkle.proof];

    const timestamp = await setBeaconBlockRoot(ACTIVE_0X01_VALIDATOR_PROOF.blockRoot);

    const validatorWitness: ValidatorWitness = {
      proof: concatenatedProof,
      validatorIndex: ACTIVE_0X01_VALIDATOR_PROOF.witness.validatorIndex,
      childBlockTimestamp: BigInt(timestamp),
      slot: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.slot),
      proposerIndex: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.proposerIndex),
      effectiveBalance: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.effectiveBalance),
      activationEpoch: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.activationEpoch),
      activationEligibilityEpoch: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.activationEligibilityEpoch),
    };

    // PG style proof verification from PK+WC to BeaconBlockRoot
    await verifier.verifyActiveValidatorContainer(
      validatorWitness,
      ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.pubkey,
      ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.withdrawalCredentials,
    );
  });

  it("should revert if no beacon chain root found for the timestamp", async () => {
    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(
      ACTIVE_0X01_VALIDATOR_PROOF.witness.validator,
    );
    const beaconHeaderMerkle = await sszMerkleTree.getBeaconBlockHeaderProof(
      ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader,
    );
    const validatorGIndex = await verifier.getValidatorGI(ACTIVE_0X01_VALIDATOR_PROOF.witness.validatorIndex, 0);

    // Verify (ValidatorContainer) leaf against (StateRoot) Merkle root
    await sszMerkleTree.verifyProof(
      ACTIVE_0X01_VALIDATOR_PROOF.witness.proof,
      ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.stateRoot,
      validatorMerkle.root,
      validatorGIndex,
    );

    // Verify (StateRoot) leaf against (BeaconBlockRoot) Merkle root
    await sszMerkleTree.verifyProof(
      [...beaconHeaderMerkle.proof],
      beaconHeaderMerkle.root,
      ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.stateRoot,
      beaconHeaderMerkle.index,
    );

    // concatentate all proofs to match PG style
    const concatenatedProof = [...ACTIVE_0X01_VALIDATOR_PROOF.witness.proof, ...beaconHeaderMerkle.proof];

    const validatorWitness: ValidatorWitness = {
      proof: concatenatedProof,
      validatorIndex: ACTIVE_0X01_VALIDATOR_PROOF.witness.validatorIndex,
      childBlockTimestamp: 0n,
      slot: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.slot),
      proposerIndex: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.beaconBlockHeader.proposerIndex),
      effectiveBalance: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.effectiveBalance),
      activationEpoch: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.activationEpoch),
      activationEligibilityEpoch: BigInt(ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.activationEligibilityEpoch),
    };

    // PG style proof verification from PK+WC to BeaconBlockRoot
    const call = verifier.verifyActiveValidatorContainer(
      validatorWitness,
      ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.pubkey,
      ACTIVE_0X01_VALIDATOR_PROOF.witness.validator.withdrawalCredentials,
    );
    expectRevertWithCustomError(verifier, call, "RootNotFound");
  });

  it("can verify against dynamic merkle tree", async () => {
    const validator = generateValidator();

    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(validator.container);

    // Verify (PK+WC) leaf against (ValidatorRoot) Merkle root
    await sszMerkleTree.verifyProof(
      [...validatorMerkle.proof],
      validatorMerkle.root,
      validatorMerkle.parentNode,
      validatorMerkle.parentIndex,
    );

    // deploy new verifier with new gIFirstValidator
    const factory = await ethers.getContractFactory("TestValidatorContainerProofVerifier");
    const newVerifier = await factory.deploy(localTreeGiFirstValidator, localTreeGiFirstValidator, 0);
    await newVerifier.waitForDeployment();

    // add validator to CL state merkle tree
    await sszMerkleTree.addValidatorLeaf(validator.container);
    const validatorIndex = lastValidatorIndex + 1n;
    const stateRoot = await sszMerkleTree.getMerkleRoot();

    const validatorLeafIndex = firstValidatorLeafIndex + validatorIndex;
    const stateProof = await sszMerkleTree.getMerkleProof(validatorLeafIndex);
    const validatorGIndex = await sszMerkleTree.getGeneralizedIndex(validatorLeafIndex);

    expect(await newVerifier.getValidatorGI(validatorIndex, 0)).to.equal(validatorGIndex);

    // Verify (ValidatorContainer) leaf against (StateRoot) Merkle root
    await sszMerkleTree.verifyProof([...stateProof], stateRoot, validatorMerkle.root, validatorGIndex);

    // Pass ValidatorNotActiveForLongEnough() error
    const activationEpoch = validator.container.activationEpoch;
    const minimumSlot = 32n * (activationEpoch + 256n) + 1n;

    const beaconHeader = generateBeaconHeader(stateRoot, Number(minimumSlot));
    const beaconMerkle = await sszMerkleTree.getBeaconBlockHeaderProof(beaconHeader);

    // Verify (StateRoot) leaf against (BeaconBlockRoot) Merkle root
    await sszMerkleTree.verifyProof([...beaconMerkle.proof], beaconMerkle.root, stateRoot, beaconMerkle.index);

    const timestamp = await setBeaconBlockRoot(beaconMerkle.root);
    const proof = [...stateProof, ...beaconMerkle.proof];

    await newVerifier.verifyActiveValidatorContainer(
      {
        validatorIndex,
        proof: [...proof],
        childBlockTimestamp: timestamp,
        slot: beaconHeader.slot,
        proposerIndex: beaconHeader.proposerIndex,
        effectiveBalance: validator.container.effectiveBalance,
        activationEpoch: validator.container.activationEpoch,
        activationEligibilityEpoch: validator.container.activationEligibilityEpoch,
      },
      validator.container.pubkey,
      validator.container.withdrawalCredentials,
    );
  });

  it("should change gIndex on pivot slot", async () => {
    const pivotSlot = 1000;
    const giPrev = randomBytes32();
    const giCurr = randomBytes32();

    const factory = await ethers.getContractFactory("TestValidatorContainerProofVerifier");
    const newVerifier = await factory.deploy(giPrev, giCurr, pivotSlot);
    await newVerifier.waitForDeployment();

    expect(await newVerifier.getValidatorGI(0n, pivotSlot - 1)).to.equal(giPrev);
    expect(await newVerifier.getValidatorGI(0n, pivotSlot)).to.equal(giCurr);
    expect(await newVerifier.getValidatorGI(0n, pivotSlot + 1)).to.equal(giCurr);
  });

  it("should validate proof with different gIndex", async () => {
    const provenValidator = generateValidator();
    const pivotSlot = 100000;
    provenValidator.container.activationEpoch = BigInt(Math.floor(pivotSlot / 32) - 257);

    const prepareCLState = async (gIndex: string, slot: number) => {
      const {
        sszMerkleTree: localTree,
        gIFirstValidator,
        firstValidatorLeafIndex: localFirstValidatorLeafIndex,
      } = await prepareLocalMerkleTree(gIndex);
      await localTree.addValidatorLeaf(provenValidator.container);

      const gIndexProven = await localTree.getGeneralizedIndex(localFirstValidatorLeafIndex + 1n);
      const stateProof = await localTree.getMerkleProof(localFirstValidatorLeafIndex + 1n);
      const beaconHeader = generateBeaconHeader(await localTree.getMerkleRoot(), slot);
      const beaconMerkle = await localTree.getBeaconBlockHeaderProof(beaconHeader);
      const proof = [...stateProof, ...beaconMerkle.proof];

      return {
        localTree,
        gIFirstValidator,
        gIndexProven,
        proof: [...proof],
        beaconHeader,
        beaconRoot: beaconMerkle.root,
      };
    };

    const [prev, curr] = await Promise.all([
      prepareCLState("0x0000000000000000000000000000000000000000000000000056000000000028", pivotSlot - 1),
      prepareCLState("0x0000000000000000000000000000000000000000000000000096000000000028", pivotSlot + 1),
    ]);

    // current CL state
    const factory = await ethers.getContractFactory("TestValidatorContainerProofVerifier");
    const newVerifier = await factory.deploy(prev.gIFirstValidator, curr.gIFirstValidator, pivotSlot);
    await newVerifier.waitForDeployment();

    expect(await newVerifier.getValidatorGI(1n, pivotSlot - 1)).to.equal(prev.gIndexProven);
    expect(await newVerifier.getValidatorGI(1n, pivotSlot)).to.equal(curr.gIndexProven);
    expect(await newVerifier.getValidatorGI(1n, pivotSlot + 1)).to.equal(curr.gIndexProven);

    // // prev works
    const timestampPrev = await setBeaconBlockRoot(prev.beaconRoot);
    await newVerifier.verifyActiveValidatorContainer(
      {
        proof: prev.proof,
        validatorIndex: 1n,
        childBlockTimestamp: timestampPrev,
        slot: prev.beaconHeader.slot,
        proposerIndex: prev.beaconHeader.proposerIndex,
        effectiveBalance: provenValidator.container.effectiveBalance,
        activationEpoch: provenValidator.container.activationEpoch,
        activationEligibilityEpoch: provenValidator.container.activationEligibilityEpoch,
      },
      provenValidator.container.pubkey,
      provenValidator.container.withdrawalCredentials,
    );

    await ethers.provider.send("hardhat_mine", [ethers.toBeHex(1), ethers.toBeHex(1)]);

    // curr works
    const timestampCurr = await setBeaconBlockRoot(curr.beaconRoot);
    await newVerifier.verifyActiveValidatorContainer(
      {
        proof: curr.proof,
        validatorIndex: 1n,
        childBlockTimestamp: timestampCurr,
        slot: curr.beaconHeader.slot,
        proposerIndex: curr.beaconHeader.proposerIndex,
        effectiveBalance: provenValidator.container.effectiveBalance,
        activationEpoch: provenValidator.container.activationEpoch,
        activationEligibilityEpoch: provenValidator.container.activationEligibilityEpoch,
      },
      provenValidator.container.pubkey,
      provenValidator.container.withdrawalCredentials,
    );

    // prev fails on curr slot
    await expect(
      newVerifier.verifyActiveValidatorContainer(
        {
          proof: [...prev.proof],
          validatorIndex: 1n,
          childBlockTimestamp: timestampCurr,
          // invalid slot to get wrong GIndex
          slot: curr.beaconHeader.slot,
          proposerIndex: curr.beaconHeader.proposerIndex,
          effectiveBalance: provenValidator.container.effectiveBalance,
          activationEpoch: provenValidator.container.activationEpoch,
          activationEligibilityEpoch: provenValidator.container.activationEligibilityEpoch,
        },
        provenValidator.container.pubkey,
        provenValidator.container.withdrawalCredentials,
      ),
    ).to.be.revertedWithCustomError(newVerifier, "InvalidSlot");
  });

  it("should verify fabricated 0x02 validator object in merkle tree", async () => {
    const randomAddress = ethers.Wallet.createRandom().address;
    const { eip4788Witness, beaconHeaderMerkleSubtreeProof } = await generateEIP4478Witness(
      sszMerkleTree,
      verifier,
      randomAddress,
    );
    const concatenatedProof = [...eip4788Witness.witness.proof, ...beaconHeaderMerkleSubtreeProof];
    const timestamp = await setBeaconBlockRoot(eip4788Witness.blockRoot);

    const validatorWitness: ValidatorWitness = {
      proof: concatenatedProof,
      validatorIndex: eip4788Witness.witness.validatorIndex,
      childBlockTimestamp: BigInt(timestamp),
      slot: BigInt(eip4788Witness.beaconBlockHeader.slot),
      proposerIndex: BigInt(eip4788Witness.beaconBlockHeader.proposerIndex),
      effectiveBalance: BigInt(eip4788Witness.witness.validator.effectiveBalance),
      activationEpoch: BigInt(eip4788Witness.witness.validator.activationEpoch),
      activationEligibilityEpoch: BigInt(eip4788Witness.witness.validator.activationEligibilityEpoch),
    };

    // PG style proof verification from PK+WC to BeaconBlockRoot
    await verifier.verifyActiveValidatorContainer(
      validatorWitness,
      eip4788Witness.witness.validator.pubkey,
      eip4788Witness.witness.validator.withdrawalCredentials,
    );
  });

  it("should revert for Validator that has not been active for long enough", async () => {
    const slot = randomInt(1743359);
    const epoch = BigInt(slot) / SLOTS_PER_EPOCH;
    const activationEpoch = epoch - SHARD_COMMITTEE_PERIOD + 1n;

    const { container } = generateValidator();
    const validatorWitness: ValidatorWitness = {
      proof: [],
      validatorIndex: BigInt(randomInt(1743359)),
      effectiveBalance: container.effectiveBalance,
      childBlockTimestamp: BigInt(randomInt(1743359)),
      slot: BigInt(slot),
      activationEpoch,
      activationEligibilityEpoch: activationEpoch - 10n,
      proposerIndex: BigInt(randomInt(1743359)),
    };

    await expectRevertWithCustomError(
      verifier,
      verifier.validateActivationEpoch(validatorWitness),
      "ValidatorNotActiveForLongEnough",
    );
  });
});
