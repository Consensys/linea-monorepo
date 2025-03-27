import { expect } from "chai";
import { TestEip1559RlpEncoder } from "../../../typechain-types";
import { deployFromFactory } from "../common/deployment";
import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import transactionWithoutCalldata from "../_testData/eip1559RlpEncoderTransactions/withoutCalldata.json";
import transactionWithCalldata from "../_testData/eip1559RlpEncoderTransactions/withCalldata.json";
import { Eip1559Transaction } from "../common/types";
import { generateKeccak256BytesDirectly } from "../common/helpers";

describe.only("Eip1559RlpEncoder Library", () => {
  let contract: TestEip1559RlpEncoder;

  async function deployTestEip1559RlpEncoderFixture() {
    return deployFromFactory("TestEip1559RlpEncoder", 1);
  }
  beforeEach(async () => {
    contract = (await loadFixture(deployTestEip1559RlpEncoderFixture)) as TestEip1559RlpEncoder;
  });

  describe("RLP Encoding and hashing", () => {
    it("Succeeds for a transaction without calldata", async () => {
      const { rlpEncodedTransaction, transactionHash } = await contract.encodeEip1559Transaction(
        buildEip1559Transaction(transactionWithoutCalldata.result),
      );

      expect(transactionWithoutCalldata.result.hash).equal(transactionHash);
      expect(transactionWithoutCalldata.rlpEncoded).equal(rlpEncodedTransaction);
      expect(generateKeccak256BytesDirectly(rlpEncodedTransaction)).equal(transactionWithoutCalldata.result.hash);
    });

    it("Succeeds for a transaction with calldata", async () => {
      const { rlpEncodedTransaction, transactionHash } = await contract.encodeEip1559Transaction(
        buildEip1559Transaction(transactionWithCalldata.result),
      );

      expect(transactionWithCalldata.result.hash).equal(transactionHash);
      expect(transactionWithCalldata.rlpEncoded).equal(rlpEncodedTransaction);
      expect(generateKeccak256BytesDirectly(rlpEncodedTransaction)).equal(transactionWithCalldata.result.hash);
    });
  });

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  function buildEip1559Transaction(data: any): Eip1559Transaction {
    return {
      nonce: data.nonce,
      maxPriorityFeePerGas: data.maxPriorityFeePerGas,
      maxFeePerGas: data.maxFeePerGas,
      gasLimit: data.gas,
      to: data.to,
      value: data.value,
      input: data.input,
      v: data.v,
      r: data.r,
      s: data.s,
    };
  }
});
