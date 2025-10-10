import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { deployFromFactory } from "contracts/test/hardhat/common/deployment";
import { ZeroHash, sha256, concat, getBytes } from "ethers";
import { TestBLS } from "contracts/typechain-types";

describe("BLS", () => {
  let bls: TestBLS;

  beforeEach(async () => {
    async function deploy() {
      return deployFromFactory("TestBLS");
    }

    bls = (await loadFixture(deploy)) as TestBLS;
  });

  describe("sha256Pair", () => {
    it("zeros + zeros", async () => {
      const left = ZeroHash;
      const right = ZeroHash;

      const expected = sha256(concat([getBytes(left), getBytes(right)]));
      const actual = await bls.sha256Pair(left, right);
      expect(actual).to.equal(expected);
    });

    it("distinct inputs", async () => {
      const left = "0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5";
      const right = "0xf4551dd23f47858f0e66957db62a0bced8cfd5e9cbd63f2fd73672ed0db7c124";

      const expected = sha256(concat([getBytes(left), getBytes(right)]));
      const actual = await bls.sha256Pair(left, right);
      expect(actual).to.equal(expected);
    });
  });
});
