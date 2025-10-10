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

  it("should verify precalculated 0x01 validator object in merkle tree", async () => {
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
    const factory = await ethers.getContractFactory("TestCLProofVerifier");
    const newVerifier = await factory.deploy(prev.gIFirstValidator, curr.gIFirstValidator, pivotSlot);
    await newVerifier.waitForDeployment();

    expect(await newVerifier.getValidatorGI(1n, pivotSlot - 1)).to.equal(prev.gIndexProven);
    expect(await newVerifier.getValidatorGI(1n, pivotSlot)).to.equal(curr.gIndexProven);
    expect(await newVerifier.getValidatorGI(1n, pivotSlot + 1)).to.equal(curr.gIndexProven);

    // // prev works
    const timestampPrev = await setBeaconBlockRoot(prev.beaconRoot);
    await newVerifier.validateValidatorContainerForPermissionlessUnstake(
      {
        proof: prev.proof,
        validatorIndex: 1n,
        pubkey: provenValidator.container.pubkey,
        childBlockTimestamp: timestampPrev,
        slot: prev.beaconHeader.slot,
        proposerIndex: prev.beaconHeader.proposerIndex,
        effectiveBalance: provenValidator.container.effectiveBalance,
        activationEpoch: provenValidator.container.activationEpoch,
        activationEligibilityEpoch: provenValidator.container.activationEligibilityEpoch,
      },
      provenValidator.container.withdrawalCredentials,
    );

    await ethers.provider.send("hardhat_mine", [ethers.toBeHex(1), ethers.toBeHex(1)]);

    // curr works
    const timestampCurr = await setBeaconBlockRoot(curr.beaconRoot);
    await newVerifier.validateValidatorContainerForPermissionlessUnstake(
      {
        proof: curr.proof,
        validatorIndex: 1n,
        pubkey: provenValidator.container.pubkey,
        childBlockTimestamp: timestampCurr,
        slot: curr.beaconHeader.slot,
        proposerIndex: curr.beaconHeader.proposerIndex,
        effectiveBalance: provenValidator.container.effectiveBalance,
        activationEpoch: provenValidator.container.activationEpoch,
        activationEligibilityEpoch: provenValidator.container.activationEligibilityEpoch,
      },
      provenValidator.container.withdrawalCredentials,
    );

    // prev fails on curr slot
    await expect(
      newVerifier.validateValidatorContainerForPermissionlessUnstake(
        {
          proof: [...prev.proof],
          validatorIndex: 1n,
          pubkey: provenValidator.container.pubkey,
          childBlockTimestamp: timestampCurr,
          // invalid slot to get wrong GIndex
          slot: curr.beaconHeader.slot,
          proposerIndex: curr.beaconHeader.proposerIndex,
          effectiveBalance: provenValidator.container.effectiveBalance,
          activationEpoch: provenValidator.container.activationEpoch,
          activationEligibilityEpoch: provenValidator.container.activationEligibilityEpoch,
        },
        provenValidator.container.withdrawalCredentials,
      ),
    ).to.be.revertedWithCustomError(newVerifier, "InvalidSlot");
  });

  it("should verify precalculated 0x01 validator object converted to 0x02 in merkle tree", async () => {
    // Verify that the 'getRoot' function works as expected
    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(
      ACTIVE_0X01_VALIDATOR.witness.validator,
    );
    const validatorGIndex = await verifier.getValidatorGI(ACTIVE_0X01_VALIDATOR.witness.validatorIndex, 0);

    const expectedStateRoot = await sszMerkleTree.getRoot(
      ACTIVE_0X01_VALIDATOR.witness.proof,
      validatorMerkle.root,
      validatorGIndex,
    );

    expect(expectedStateRoot).to.equal(ACTIVE_0X01_VALIDATOR.beaconBlockHeader.stateRoot);

    // Modify the 0x01 validator to 0x02 validator
    const CONVERTED_ACTIVE_0X01_VALIDATOR = {
      ...ACTIVE_0X01_VALIDATOR,
      witness: { ...ACTIVE_0X01_VALIDATOR.witness, validator: ACTIVE_0X01_VALIDATOR.witness.validator },
    };

    CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.withdrawalCredentials =
      "0x02" + CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.withdrawalCredentials.slice(4);

    // Compute new state root
    const convertedValidatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(
      CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator,
    );
    const convertedValidatorGIndex = await verifier.getValidatorGI(
      CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validatorIndex,
      0,
    );

    const convertedExpectedStateRoot = await sszMerkleTree.getRoot(
      CONVERTED_ACTIVE_0X01_VALIDATOR.witness.proof,
      convertedValidatorMerkle.root,
      convertedValidatorGIndex,
    );

    CONVERTED_ACTIVE_0X01_VALIDATOR.beaconBlockHeader.stateRoot = convertedExpectedStateRoot;

    // Verify (ValidatorContainer) leaf against (StateRoot) Merkle root
    await sszMerkleTree.verifyProof(
      CONVERTED_ACTIVE_0X01_VALIDATOR.witness.proof,
      CONVERTED_ACTIVE_0X01_VALIDATOR.beaconBlockHeader.stateRoot,
      convertedValidatorMerkle.root,
      convertedValidatorGIndex,
    );

    // Compute new BeaconBlockRoot
    const convertedBeaconHeaderMerkle = await sszMerkleTree.getBeaconBlockHeaderProof(
      CONVERTED_ACTIVE_0X01_VALIDATOR.beaconBlockHeader,
    );

    const convertedBeaconRoot = await sszMerkleTree.getRoot(
      [...convertedBeaconHeaderMerkle.proof],
      convertedExpectedStateRoot,
      convertedBeaconHeaderMerkle.index,
    );
    CONVERTED_ACTIVE_0X01_VALIDATOR.blockRoot = convertedBeaconRoot;

    // Verify (StateRoot) leaf against (BeaconBlockRoot) Merkle root
    await sszMerkleTree.verifyProof(
      [...convertedBeaconHeaderMerkle.proof],
      convertedBeaconHeaderMerkle.root,
      CONVERTED_ACTIVE_0X01_VALIDATOR.beaconBlockHeader.stateRoot,
      convertedBeaconHeaderMerkle.index,
    );

    // concatentate all proofs to match PG style
    const concatenatedProof = [...CONVERTED_ACTIVE_0X01_VALIDATOR.witness.proof, ...convertedBeaconHeaderMerkle.proof];

    const timestamp = await setBeaconBlockRoot(CONVERTED_ACTIVE_0X01_VALIDATOR.blockRoot);

    const validatorWitness: ValidatorWitness = {
      proof: concatenatedProof,
      pubkey: CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.pubkey,
      validatorIndex: CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validatorIndex,
      childBlockTimestamp: BigInt(timestamp),
      slot: BigInt(CONVERTED_ACTIVE_0X01_VALIDATOR.beaconBlockHeader.slot),
      proposerIndex: BigInt(CONVERTED_ACTIVE_0X01_VALIDATOR.beaconBlockHeader.proposerIndex),
      effectiveBalance: BigInt(CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.effectiveBalance),
      activationEpoch: BigInt(CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.activationEpoch),
      activationEligibilityEpoch: BigInt(CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.activationEligibilityEpoch),
    };

    // PG style proof verification from PK+WC to BeaconBlockRoot
    await verifier.validateValidatorContainerForPermissionlessUnstake(
      validatorWitness,
      CONVERTED_ACTIVE_0X01_VALIDATOR.witness.validator.withdrawalCredentials,
    );
  });
});
