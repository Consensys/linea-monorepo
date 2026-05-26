import { describe, expect, it } from "@jest/globals";

import {
  getPendingTransactions,
  getTransactionsFromPool,
  hasPendingTransactions,
  serializePendingTransaction,
} from "../transactions.js";
import { type Transaction, type Txpool } from "../types.js";

const R = "0x0000000000000000000000000000000000000000000000000000000000000001";
const S = "0x0000000000000000000000000000000000000000000000000000000000000002";

const LEGACY_RAW = "0xdf80018252089400000000000000000000000000000000000000018080250102";
const EIP2930_RAW =
  "0x01f85c0101028261a894000000000000000000000000000000000000000203821234f838f7940000000000000000000000000000000000000003e1a00000000000000000000000000000000000000000000000000000000000000003010102";
const EIP1559_RAW = "0x02e682e7080202648275309400000000000000000000000000000000000000040582abcdc0800102";

const legacyTransaction: Transaction = {
  hash: "0x314b4875720faeb2f2e01f4673ba7439d0fe447171f974e7b12dc4c328e8f7ab",
  nonce: "0x0",
  gas: "0x5208",
  gasPrice: "0x1",
  input: "0x",
  value: "0x0",
  type: "0x0",
  to: "0x0000000000000000000000000000000000000001",
  r: R,
  s: S,
  v: "0x25",
};

const eip2930Transaction: Transaction = {
  hash: "0x3142a140dcbb20ce356da11433239484be63a3127572a8e1e44383993333affa",
  nonce: "0x1",
  gas: "0x61a8",
  gasPrice: "0x2",
  input: "0x1234",
  value: "0x3",
  chainId: "0x1",
  accessList: [
    {
      address: "0x0000000000000000000000000000000000000003",
      storageKeys: ["0x0000000000000000000000000000000000000000000000000000000000000003"],
    },
  ],
  type: "0x1",
  to: "0x0000000000000000000000000000000000000002",
  r: R,
  s: S,
  v: "0x1",
};

const eip1559Transaction: Transaction = {
  hash: "0x9478b3dc6b5759173e28f55f700628118d5c10d492d0e6ae9928c7255c53e9f6",
  nonce: "0x2",
  gas: "0x7530",
  maxFeePerGas: "0x64",
  maxPriorityFeePerGas: "0x2",
  input: "0xabcd",
  value: "0x5",
  chainId: "0xe708",
  accessList: [],
  type: "0x2",
  to: "0x0000000000000000000000000000000000000004",
  r: R,
  s: S,
  yParity: "0x0",
};

describe("synctx transaction utilities", () => {
  describe("serializePendingTransaction", () => {
    it("serializes a signed legacy transaction", () => {
      expect(serializePendingTransaction(legacyTransaction)).toStrictEqual(LEGACY_RAW);
    });

    it("serializes a signed EIP-2930 transaction with an access list", () => {
      expect(serializePendingTransaction(eip2930Transaction)).toStrictEqual(EIP2930_RAW);
    });

    it("serializes a signed EIP-1559 transaction", () => {
      expect(serializePendingTransaction(eip1559Transaction)).toStrictEqual(EIP1559_RAW);
    });

    it("throws when signature fields are missing", () => {
      const transactionWithoutSignature: Partial<Transaction> = { ...legacyTransaction };
      delete transactionWithoutSignature.r;

      expect(() => serializePendingTransaction(transactionWithoutSignature as Transaction)).toThrow(
        `Missing required transaction field r for ${legacyTransaction.hash}`,
      );
    });

    it("throws when the transaction type is unsupported", () => {
      expect(() => serializePendingTransaction({ ...eip1559Transaction, type: "0x3" })).toThrow(
        `Unsupported transaction type 0x3 for ${eip1559Transaction.hash}`,
      );
    });

    it("throws when the serialized transaction hash does not match", () => {
      expect(() =>
        serializePendingTransaction({
          ...legacyTransaction,
          hash: "0x0000000000000000000000000000000000000000000000000000000000000000",
        }),
      ).toThrow("Serialized transaction hash mismatch");
    });
  });

  describe("transaction pool helpers", () => {
    it("normalizes and serializes transactions from a Geth txpool response", () => {
      const pool: Txpool = {
        pending: {
          "0x0000000000000000000000000000000000000001": {
            "0x0": legacyTransaction,
          },
        },
        queued: {},
      };

      const transactions = getTransactionsFromPool("geth", pool);

      expect(hasPendingTransactions("geth", pool)).toBe(true);
      expect(transactions.map((tx) => serializePendingTransaction(tx))).toStrictEqual([LEGACY_RAW]);
    });

    it("normalizes and serializes transactions from a Besu txpool response", () => {
      const transactions = getTransactionsFromPool("besu", [eip1559Transaction]);

      expect(hasPendingTransactions("besu", [eip1559Transaction])).toBe(true);
      expect(transactions.map((tx) => serializePendingTransaction(tx))).toStrictEqual([EIP1559_RAW]);
    });

    it("returns source transactions absent from the target pool", () => {
      expect(getPendingTransactions([legacyTransaction, eip1559Transaction], [legacyTransaction])).toStrictEqual([
        eip1559Transaction,
      ]);
    });
  });
});
