import { expect } from "chai";
import { TestRlpEncoder } from "../../../typechain-types";
import { deployFromFactory } from "../common/deployment";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";

describe("RlpEncoder Library", () => {
  let contract: TestRlpEncoder;

  async function deployTestEip1559RlpEncoderFixture() {
    return deployFromFactory("TestRlpEncoder");
  }
  beforeEach(async () => {
    contract = (await loadFixture(deployTestEip1559RlpEncoderFixture)) as TestRlpEncoder;
  });

  describe("RLP Encoding", () => {
    describe("Bool Encoding", () => {
      it("Encodes false correctly", async () => {
        const encoded = await contract.encodeBool(false);
        expect(encoded).equal("0x80");
      });

      it("Encodes true correctly", async () => {
        const encoded = await contract.encodeBool(true);
        expect(encoded).equal("0x01");
      });
    });

    describe("String Encoding", () => {
      it("Encodes blank string correctly", async () => {
        const encoded = await contract.encodeString("");
        expect(encoded).equal("0x80");
      });

      it("Encodes a short string correctly", async () => {
        const encoded = await contract.encodeString("short");
        const expected = "0x8573686f7274";
        expect(encoded).equal(expected);
      });

      it("Encodes a long string correctly", async () => {
        const encoded = await contract.encodeString(
          "This is a string that is quite long and needs some different encoding",
        );
        const expected =
          "0xb84554686973206973206120737472696e672074686174206973207175697465206c6f6e6720616e64206e6565647320736f6d6520646966666572656e7420656e636f64696e67";
        expect(encoded).equal(expected);
      });
    });
  });
});
