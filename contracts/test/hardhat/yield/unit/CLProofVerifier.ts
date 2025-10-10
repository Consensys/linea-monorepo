import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { SSZMerkleTree, TestCLProofVerifier } from "contracts/typechain-types";
import { deployTestCLProofVerifier, ValidatorWitness } from "../helpers";
import {
  ACTIVE_0X01_VALIDATOR,
  generateBeaconHeader,
  generateValidator,
  prepareLocalMerkleTree,
  randomBytes32,
  setBeaconBlockRoot,
} from "../helpers/proof";
import { ethers } from "hardhat";
import { FAR_FUTURE_EXIT_EPOCH } from "../../common/constants/yield";

describe("BLS", () => {
  let verifier: TestCLProofVerifier;
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
    verifier = await loadFixture(deployTestCLProofVerifier);
    // test mocker
    const mockRoot = randomBytes32();
    const timestamp = await setBeaconBlockRoot(mockRoot);
    expect(await verifier.getParentBlockRoot(timestamp)).to.equal(mockRoot);
  });

  it("should verify precalculated validator object in merkle tree", async () => {
    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(
      ACTIVE_0X01_VALIDATOR.witness.validator,
    );
    const beaconHeaderMerkle = await sszMerkleTree.getBeaconBlockHeaderProof(ACTIVE_0X01_VALIDATOR.beaconBlockHeader);
    const validatorGIndex = await verifier.getValidatorGI(ACTIVE_0X01_VALIDATOR.witness.validatorIndex, 0);

    // Verify (ValidatorContainer) leaf against (StateRoot) Merkle root
    await sszMerkleTree.verifyProof(
      ACTIVE_0X01_VALIDATOR.witness.proof,
      ACTIVE_0X01_VALIDATOR.beaconBlockHeader.stateRoot,
      validatorMerkle.root,
      validatorGIndex,
    );

    // Verify (StateRoot) leaf against (BeaconBlockRoot) Merkle root
    await sszMerkleTree.verifyProof(
      [...beaconHeaderMerkle.proof],
      beaconHeaderMerkle.root,
      ACTIVE_0X01_VALIDATOR.beaconBlockHeader.stateRoot,
      beaconHeaderMerkle.index,
    );

    // concatentate all proofs to match PG style
    const concatenatedProof = [...ACTIVE_0X01_VALIDATOR.witness.proof, ...beaconHeaderMerkle.proof];

    const timestamp = await setBeaconBlockRoot(ACTIVE_0X01_VALIDATOR.blockRoot);

    const validatorWitness: ValidatorWitness = {
      proof: concatenatedProof,
      pubkey: ACTIVE_0X01_VALIDATOR.witness.validator.pubkey,
      validatorIndex: ACTIVE_0X01_VALIDATOR.witness.validatorIndex,
      childBlockTimestamp: BigInt(timestamp),
      slot: BigInt(ACTIVE_0X01_VALIDATOR.beaconBlockHeader.slot),
      proposerIndex: BigInt(ACTIVE_0X01_VALIDATOR.beaconBlockHeader.proposerIndex),
      effectiveBalance: BigInt(ACTIVE_0X01_VALIDATOR.witness.validator.effectiveBalance),
      activationEpoch: BigInt(ACTIVE_0X01_VALIDATOR.witness.validator.activationEpoch),
      activationEligibilityEpoch: BigInt(ACTIVE_0X01_VALIDATOR.witness.validator.activationEligibilityEpoch),
    };

    // PG style proof verification from PK+WC to BeaconBlockRoot
    await verifier.validateValidatorContainerForPermissionlessUnstake(
      validatorWitness,
      ACTIVE_0X01_VALIDATOR.witness.validator.withdrawalCredentials,
    );
  });

  it("can verify against dynamic merkle tree", async () => {
    const validator = generateValidator();
    // Hardcoded values in our fork of CLProofVerifier.sol
    validator.container.exitEpoch = FAR_FUTURE_EXIT_EPOCH;
    validator.container.withdrawableEpoch = FAR_FUTURE_EXIT_EPOCH;
    validator.container.slashed = false;

    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(validator.container);

    // Verify (PK+WC) leaf against (ValidatorRoot) Merkle root
    await sszMerkleTree.verifyProof(
      [...validatorMerkle.proof],
      validatorMerkle.root,
      validatorMerkle.parentNode,
      validatorMerkle.parentIndex,
    );

    // deploy new verifier with new gIFirstValidator
    const factory = await ethers.getContractFactory("TestCLProofVerifier");
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

    await newVerifier.validateValidatorContainerForPermissionlessUnstake(
      {
        validatorIndex,
        proof: [...proof],
        pubkey: validator.container.pubkey,
        childBlockTimestamp: timestamp,
        slot: beaconHeader.slot,
        proposerIndex: beaconHeader.proposerIndex,
        effectiveBalance: validator.container.effectiveBalance,
        activationEpoch: validator.container.activationEpoch,
        activationEligibilityEpoch: validator.container.activationEligibilityEpoch,
      },
      validator.container.withdrawalCredentials,
    );
  });

  it("should change gIndex on pivot slot", async () => {
    const pivotSlot = 1000;
    const giPrev = randomBytes32();
    const giCurr = randomBytes32();

    const factory = await ethers.getContractFactory("TestCLProofVerifier");
    const newVerifier = await factory.deploy(giPrev, giCurr, pivotSlot);
    await newVerifier.waitForDeployment();

    expect(await newVerifier.getValidatorGI(0n, pivotSlot - 1)).to.equal(giPrev);
    expect(await newVerifier.getValidatorGI(0n, pivotSlot)).to.equal(giCurr);
    expect(await newVerifier.getValidatorGI(0n, pivotSlot + 1)).to.equal(giCurr);
  });
});
