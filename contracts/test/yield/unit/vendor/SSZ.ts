import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { TestGIndex, TestSSZ } from "contracts/typechain-types";
import { deployFromFactory } from "contracts/test/common/deployment";
import { hexlify, randomBytes, ZeroHash, zeroPadBytes } from "ethers";
import { ValidatorContainer } from "contracts/test/yield/helpers/types";
import { expectRevertWithCustomError } from "contracts/test/common/helpers";

describe("SSZ", () => {
  let ssz: TestSSZ;
  let gIndexLib: TestGIndex;

  beforeEach(async () => {
    async function deploy() {
      const [ssz_, gindex_] = await Promise.all([deployFromFactory("TestSSZ"), deployFromFactory("TestGIndex")]);
      return { ssz_: ssz_ as TestSSZ, gindex_: gindex_ as TestGIndex };
    }

    const deployed = await loadFixture(deploy);
    ssz = deployed.ssz_;
    gIndexLib = deployed.gindex_;
  });

  describe("toLittleEndian", () => {
    it("uint: example value", async () => {
      const v = 0x1234567890abcdefn;
      const expected = zeroPadBytes("0xEFCDAB9078563412", 32);
      const actual = await ssz.toLittleEndianUint(v);
      expect(actual).to.equal(expected);
    });

    it("uint: zero", async () => {
      const expected = ZeroHash;
      const actual = await ssz.toLittleEndianUint(0n);
      expect(actual).to.equal(expected);
    });

    it("bool: false", async () => {
      const expected = ZeroHash;
      const actual = await ssz.toLittleEndianBool(false);
      expect(actual).to.equal(expected);
    });

    it("bool: true", async () => {
      const expected = zeroPadBytes("0x01", 32);
      const actual = await ssz.toLittleEndianBool(true);
      expect(actual).to.equal(expected);
    });

    it("fuzz-ish: applying twice returns original", async () => {
      for (let i = 0; i < 10; i++) {
        const v = BigInt(hexlify(randomBytes(32)));
        const once = await ssz.toLittleEndianUint(v);
        const twice = await ssz.toLittleEndianUint(BigInt(once));
        expect(BigInt(twice)).to.equal(v);
      }
    });
  });

  describe("hashTreeRoot(Validator)", () => {
    it("Exited + Slashed", async () => {
      const validator: ValidatorContainer = {
        pubkey: "0x91760f8a17729cfcb68bfc621438e5d9dfa831cd648e7b2b7d33540a7cbfda1257e4405e67cd8d3260351ab3ff71b213",
        withdrawalCredentials: "0x01000000000000000000000006676e8584342cc8b6052cfdf381c3a281f00ac8",
        effectiveBalance: 30000000000n,
        slashed: true,
        activationEligibilityEpoch: 242529n,
        activationEpoch: 242551n,
        exitEpoch: 242556n,
        withdrawableEpoch: 250743n,
      };

      const expected = "0xe4674dc5c27e7d3049fcd298745c00d3e314f03d33c877f64bf071d3b77eb942";
      const actual = await ssz.hashTreeRoot(validator);
      expect(actual).to.equal(expected);
    });

    it("Active", async () => {
      const validator: ValidatorContainer = {
        pubkey: "0x8fb78536e82bcec34e98fff85c907f0a8e6f4b1ccdbf1e8ace26b59eb5a06d16f34e50837f6c490e2ad6a255db8d543b",
        withdrawalCredentials: "0x0023b9d00bf66e7f8071208a85afde59b3148dea046ee3db5d79244880734881",
        effectiveBalance: 32000000000n,
        slashed: false,
        activationEligibilityEpoch: 2593n,
        activationEpoch: 5890n,
        exitEpoch: BigInt("0xffffffffffffffff"),
        withdrawableEpoch: BigInt("0xffffffffffffffff"),
      };

      const actual = await ssz.hashTreeRoot(validator);
      const expected = "0x60fb91184416404ddfc62bef6df9e9a52c910751daddd47ea426aabaf19dfa09";
      expect(actual).to.equal(expected);
    });

    it("Extra bytes in pubkey are ignored", async () => {
      const validator: ValidatorContainer = {
        pubkey:
          "0x8fb78536e82bcec34e98fff85c907f0a8e6f4b1ccdbf1e8ace26b59eb5a06d16f34e50837f6c490e2ad6a255db8d543bDEADBEEFDEADBEEFDEADBEEFDEADBEEF",
        withdrawalCredentials: "0x0023b9d00bf66e7f8071208a85afde59b3148dea046ee3db5d79244880734881",
        effectiveBalance: 32000000000n,
        slashed: false,
        activationEligibilityEpoch: 2593n,
        activationEpoch: 5890n,
        exitEpoch: BigInt("0xffffffffffffffff"),
        withdrawableEpoch: BigInt("0xffffffffffffffff"),
      };

      const actual = await ssz.hashTreeRoot(validator);
      const expected = "0x60fb91184416404ddfc62bef6df9e9a52c910751daddd47ea426aabaf19dfa09";
      expect(actual).to.equal(expected);
    });

    it("All zeroes", async () => {
      const validator: ValidatorContainer = {
        pubkey: "0x" + "0".repeat(96),
        withdrawalCredentials: ZeroHash,
        effectiveBalance: 0n,
        slashed: false,
        activationEligibilityEpoch: 0n,
        activationEpoch: 0n,
        exitEpoch: 0n,
        withdrawableEpoch: 0n,
      };

      const actual = await ssz.hashTreeRoot(validator);
      const expected = "0xfa324a462bcb0f10c24c9e17c326a4e0ebad204feced523eccaf346c686f06ee";
      expect(actual).to.equal(expected);
    });

    it("All ones", async () => {
      const validator: ValidatorContainer = {
        pubkey: "0x" + "f".repeat(96),
        withdrawalCredentials: "0x" + "f".repeat(64),
        effectiveBalance: BigInt("0xffffffffffffffff"),
        slashed: true,
        activationEligibilityEpoch: BigInt("0xffffffffffffffff"),
        activationEpoch: BigInt("0xffffffffffffffff"),
        exitEpoch: BigInt("0xffffffffffffffff"),
        withdrawableEpoch: BigInt("0xffffffffffffffff"),
      };

      const actual = await ssz.hashTreeRoot(validator);
      const expected = "0x29c03a7cc9a8047ff05619a04bb6e60440a791e6ac3fe7d72e6fe9037dd3696f";
      expect(actual).to.equal(expected);
    });
  });

  describe("verifyProof", () => {
    // For the tests below, assume there's the following tree from the bottom up:
    // --
    // 0x0000000000000000000000000000000000000000000000000000000000000000
    // 0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5
    // 0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30
    // 0x21ddb9a356815c3fac1026b6dec5df3124afbadb485c9ba5a3e3398a04b7ba85
    // --
    // 0x0a4b105f69a6f41c3b3efc9bb5ac525b5b557a524039a13c657a916d8eb04451
    // 0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124
    // --
    // 0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c

    it("happy path", async () => {
      const proof = [
        "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
        "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124",
      ];
      await ssz.verifyProof(
        proof,
        "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
        ZeroHash,
        await gIndexLib.pack(4, 0),
      );

      const proof2 = [ZeroHash, "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124"];
      await ssz.verifyProof(
        proof2,
        "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
        "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
        await gIndexLib.pack(5, 0),
      );
    });
    it("one item", async () => {
      const proof = ["0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124"];
      await ssz.verifyProof(
        proof,
        "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
        "0x0a4b105f69a6f41c3b3efc9bb5ac525b5b557a524039a13c657a916d8eb04451",
        await gIndexLib.pack(2, 0),
      );
    });
    it("revert: no proof", async () => {
      const call = ssz.verifyProof(
        [],
        "0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
        ZeroHash,
        await gIndexLib.pack(2, 0),
      );

      await expectRevertWithCustomError(ssz, call, "InvalidProof");
    });
    it("revert: proving root", async () => {
      await expect(
        ssz.verifyProof(
          [],
          "0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
          "0xf5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
          await gIndexLib.pack(1, 0),
        ),
      ).to.be.revertedWithCustomError(ssz, "InvalidProof");
    });
    it("revert: invalid proof", async () => {
      const proof = [
        "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
        "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124",
      ];
      await expect(
        ssz.verifyProof(
          proof,
          "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
          "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
          await gIndexLib.pack(4, 0),
        ),
      ).to.be.revertedWithCustomError(ssz, "InvalidProof");
    });
    it("revert: wrong gindex", async () => {
      const proof = [
        "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
        "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124",
      ];
      await expect(
        ssz.verifyProof(
          proof,
          "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
          ZeroHash,
          await gIndexLib.pack(5, 0),
        ),
      ).to.be.revertedWithCustomError(ssz, "InvalidProof");
    });
    it("revert: BranchHasExtraItem", async () => {
      const proof = [
        "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
        "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124",
      ];
      await expect(
        ssz.verifyProof(
          proof,
          "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
          "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
          await gIndexLib.pack(2, 0),
        ),
      ).to.be.revertedWithCustomError(ssz, "BranchHasExtraItem");
    });
    it("revert: BranchHasMissingItem", async () => {
      const proof = [
        "0xb4c11951957c6f8f642c4af61cd6b24640fec6dc7fc607ee8206a99e92410d30",
        "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124",
      ];
      await expect(
        ssz.verifyProof(
          proof,
          "0xda1c902c54a4386439ce622d7e527dc11decace28ebb902379cba91c4a116b1c",
          "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5",
          await gIndexLib.pack(8, 0),
        ),
      ).to.be.revertedWithCustomError(ssz, "BranchHasMissingItem");
    });
  });
});
