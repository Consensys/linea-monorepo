import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { SSZMerkleTree, TestCLProofVerifier } from "contracts/typechain-types";
import { deployTestCLProofVerifier, ValidatorWitness } from "../helpers";
import {
  ACTIVE_VALIDATOR,
  generateValidator,
  prepareLocalMerkleTree,
  randomBytes32,
  setBeaconBlockRoot,
} from "../helpers/proof";

describe("BLS", () => {
  let verifier: TestCLProofVerifier;
  let sszMerkleTree: SSZMerkleTree;
  let firstValidatorLeafIndex: bigint;
  let lastValidatorIndex: bigint;

  before(async () => {
    const localTree = await prepareLocalMerkleTree();
    sszMerkleTree = localTree.sszMerkleTree;
    firstValidatorLeafIndex = localTree.firstValidatorLeafIndex;
    // populate merkle tree with validators
    for (let i = 1; i < 100; i++) {
      await sszMerkleTree.addValidatorLeaf(generateValidator().container);
    }
    // after adding validators, all newly added validator indexes will +n from this
    lastValidatorIndex = (await sszMerkleTree.leafCount()) - 1n - firstValidatorLeafIndex;
    console.log(lastValidatorIndex);
  });

  beforeEach(async () => {
    verifier = await loadFixture(deployTestCLProofVerifier);
    // test mocker
    const mockRoot = randomBytes32();
    const timestamp = await setBeaconBlockRoot(mockRoot);
    expect(await verifier.getParentBlockRoot(timestamp)).to.equal(mockRoot);
  });

  it("should verify precalclulated validator object in merkle tree", async () => {
    const validatorMerkle = await sszMerkleTree.getValidatorPubkeyWCParentProof(ACTIVE_VALIDATOR.witness.validator);
    const beaconHeaderMerkle = await sszMerkleTree.getBeaconBlockHeaderProof(ACTIVE_VALIDATOR.beaconBlockHeader);
    const validatorGIndex = await verifier.getValidatorGI(ACTIVE_VALIDATOR.witness.validatorIndex, 0);

    // raw proof verification of (ValidatorContainer) leaf against (StateRoot) Merkle root
    await sszMerkleTree.verifyProof(
      ACTIVE_VALIDATOR.witness.proof,
      ACTIVE_VALIDATOR.beaconBlockHeader.stateRoot,
      validatorMerkle.root,
      validatorGIndex,
    );

    // concatentate all proofs to match PG style
    const concatenatedProof = [...ACTIVE_VALIDATOR.witness.proof, ...beaconHeaderMerkle.proof];

    const timestamp = await setBeaconBlockRoot(ACTIVE_VALIDATOR.blockRoot);

    const validatorWitness: ValidatorWitness = {
      proof: concatenatedProof,
      pubkey: ACTIVE_VALIDATOR.witness.validator.pubkey,
      validatorIndex: ACTIVE_VALIDATOR.witness.validatorIndex,
      childBlockTimestamp: BigInt(timestamp),
      slot: BigInt(ACTIVE_VALIDATOR.beaconBlockHeader.slot),
      proposerIndex: BigInt(ACTIVE_VALIDATOR.beaconBlockHeader.proposerIndex),
      effectiveBalance: BigInt(ACTIVE_VALIDATOR.witness.validator.effectiveBalance),
      activationEpoch: BigInt(ACTIVE_VALIDATOR.witness.validator.activationEpoch),
      activationEligibilityEpoch: BigInt(ACTIVE_VALIDATOR.witness.validator.activationEligibilityEpoch),
    };

    // PG style proof verification from PK+WC to BeaconBlockRoot
    await verifier.validateValidatorContainerForPermissionlessUnstake(
      validatorWitness,
      ACTIVE_VALIDATOR.witness.validator.withdrawalCredentials,
    );
  });
});
