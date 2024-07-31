import { describe, it, beforeEach, expect } from "@jest/globals";
import { Transaction, Wallet, ethers } from "ethers";
import { GoNativeCompressor } from "../GoNativeCompressor";

const TEST_ADDRESS = "0x0000000000000000000000000000000000000001";
const TEST_PRIVATE_KEY = "0x0000000000000000000000000000000000000000000000000000000000000001";

describe("GoNativeCompressor", () => {
  const dataLimit = 800_000;
  let compressor: GoNativeCompressor;

  beforeEach(() => {
    compressor = new GoNativeCompressor(dataLimit);
  });

  describe("getCompressedTxSize", () => {
    it("Should throw an error if an error occured during tx compression", () => {
      const transaction = Transaction.from({
        to: TEST_ADDRESS,
        value: ethers.parseEther("2"),
      });
      const rlpEncodedTransaction = ethers.encodeRlp(transaction.unsignedSerialized);
      const input = ethers.getBytes(rlpEncodedTransaction);

      expect(() => compressor.getCompressedTxSize(input)).toThrow(
        "Error while compressing the transaction: rlp: too few elements for types.DynamicFeeTx",
      );
    });

    it("Should return compressed tx size", async () => {
      const transaction = Transaction.from({
        to: TEST_ADDRESS,
        value: ethers.parseEther("2"),
      });
      const signer = new Wallet(TEST_PRIVATE_KEY);
      const encodedSignedTx = await signer.signTransaction(transaction);

      const rlpEncodedTransaction = ethers.encodeRlp(encodedSignedTx);
      const input = ethers.getBytes(rlpEncodedTransaction);

      const compressedTxSize = compressor.getCompressedTxSize(input);

      expect(compressedTxSize).toStrictEqual(43);
    });
  });
});
