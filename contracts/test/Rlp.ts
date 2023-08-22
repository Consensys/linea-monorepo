import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { TestRlp } from "../typechain-types";
import { deployFromFactory } from "./utils/deployment";
import { getRLPEncodeTransactions } from "./utils/helpers";

describe("Rlp", () => {
  let contract: TestRlp;
  let account: SignerWithAddress;

  async function deployRlpFixture() {
    return deployFromFactory("TestRlp") as Promise<TestRlp>;
  }

  const { eip1559Transaction, shortEip1559Transaction } = getRLPEncodeTransactions("test-transactions.json");

  before(async () => {
    [account] = await ethers.getSigners();
  });

  beforeEach(async () => {
    contract = await loadFixture(deployRlpFixture);
  });

  describe("Transactions are wrong length", () => {
    it("Should revert when decoding is too short", async () => {
      // note the 10 vs. 9 - the encoding adds a C0 at the end so the 9th element is the empty access list
      await expect(contract.skipTo(ethers.utils.hexDataSlice(shortEip1559Transaction, 1), 9)).to.be.reverted;
    });
  });

  describe("isList", () => {
    it("Should return true if bytes is a list", async () => {
      const list = new Uint8Array([0x12, 0x34, 0x56, 0x78]);
      const encodedList = ethers.utils.RLP.encode([list]);
      expect(await contract.isList(encodedList)).to.be.true;
    });

    it("Should return false if bytes is a not list", async () => {
      const encodedString = ethers.utils.RLP.encode("0x12345678");
      expect(await contract.isList(encodedString)).to.be.false;
    });

    it("Should return false if bytes length === 0", async () => {
      expect(await contract.isList("0x")).to.be.false;
    });
  });

  describe("itemLength", () => {
    it("Should return the byte length for a single byte", async () => {
      const str = "a";
      const encodedList = ethers.utils.RLP.encode(ethers.utils.hexlify(ethers.utils.toUtf8Bytes(str)));
      const itemLength = await contract.itemLength(encodedList);
      expect(itemLength.toNumber()).to.equal(ethers.utils.hexDataLength(encodedList));
    });

    it("Should return the byte length for a 0-55 bytes length", async () => {
      const str = "dog";
      const encodedList = ethers.utils.RLP.encode(ethers.utils.hexlify(ethers.utils.toUtf8Bytes(str)));
      const itemLength = await contract.itemLength(encodedList);
      expect(itemLength.toNumber()).to.equal(ethers.utils.hexDataLength(encodedList));
    });

    it("Should return the byte length for a >55 bytes length", async () => {
      const str = "zoo255zoo255zzzzzzzzzzzzssssssssssssssssssssssssssssssssssssssssssssss";
      const encodedList = ethers.utils.RLP.encode(ethers.utils.hexlify(ethers.utils.toUtf8Bytes(str)));
      const itemLength = await contract.itemLength(encodedList);
      expect(itemLength.toNumber()).to.equal(ethers.utils.hexDataLength(encodedList));
    });

    it("Should return the byte length for a list with length 0-55", async () => {
      const list = [
        ethers.utils.hexlify(ethers.utils.toUtf8Bytes("dog")),
        ethers.utils.hexlify(ethers.utils.toUtf8Bytes("god")),
        ethers.utils.hexlify(ethers.utils.toUtf8Bytes("cat")),
      ];
      const encodedList = ethers.utils.RLP.encode(list);
      const itemLength = await contract.itemLength(encodedList);
      expect(itemLength.toNumber()).to.equal(ethers.utils.hexDataLength(encodedList));
    });

    it("Should return the byte length for a list with length >55", async () => {
      const list = [
        ethers.utils.hexlify(ethers.utils.toUtf8Bytes("dog")),
        ethers.utils.hexlify(ethers.utils.toUtf8Bytes("god")),
        ethers.utils.hexlify(ethers.utils.toUtf8Bytes("cat")),
        ethers.utils.hexlify(
          ethers.utils.toUtf8Bytes("255255zzzzzzzzzzzzaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
        ),
      ];
      const encodedList = ethers.utils.RLP.encode(list);
      const itemLength = await contract.itemLength(encodedList);
      expect(itemLength.toNumber()).to.equal(ethers.utils.hexDataLength(encodedList));
    });
  });

  describe("next", () => {
    it("Should revert if there is no next item", async () => {
      const encodedList = ethers.utils.RLP.encode([]);
      await expect(contract.next(encodedList)).to.be.revertedWithCustomError(contract, "NoNext");
    });

    it("Should return the next RLP item (for eip1559, the nextPtr should be the chain ID)", async () => {
      const [nextItem, itemNextMemPtr] = await contract.next(
        ethers.utils.hexDataSlice(
          "0x02f887800b4d8202918301a48a94000000000000000000000000000000000000000080b864f4b476e100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001497569c1f1e97fb9eca5b5f7c7153f2eb6f9edd312bd4d0dc529f15ae67a7cbec0",
          1,
        ),
      );
      expect(nextItem.len).to.equal(1);
      expect(nextItem.memPtr.toNumber() - itemNextMemPtr.toNumber()).to.equal(0);
    });
  });

  describe("hasNext", () => {
    it("Should return true if the RLP item has a nextPtr", async () => {
      const hasNext = await contract.hasNext(ethers.utils.hexDataSlice(eip1559Transaction, 1));
      expect(hasNext).to.equal(true);
    });
  });

  describe("skipTo", () => {
    it("Should return the destination address in the iteration (for an eip1559 encoded transaction)", async () => {
      const item = await contract.skipTo(ethers.utils.hexDataSlice(eip1559Transaction, 1), 6);
      const destinationAddress = await contract.fromRlpItemToAddress({ memPtr: item.memPtr, len: item.len });
      expect(destinationAddress).to.equal("0x0000000000000000000000000000000000aD0000");
    });

    it("Should revert if it goes beyond the number of elements)", async () => {
      await expect(contract.skipTo(ethers.utils.hexDataSlice(eip1559Transaction, 1), 9)).to.be.revertedWithCustomError(
        contract,
        "MemoryOutOfBounds",
      );
    });

    it("Should not revert if it meets maximum elements)", async () => {
      await expect(
        contract.skipTo(ethers.utils.hexDataSlice(eip1559Transaction, 1), 8),
      ).to.not.be.revertedWithCustomError(contract, "MemoryOutOfBounds");
    });
  });

  describe("iterator", () => {
    it("Should revert if the RLP item is not a list", async () => {
      const encodedAddress = ethers.utils.RLP.encode(account.address);
      await expect(contract.iterator(encodedAddress)).to.be.revertedWithCustomError(contract, "NotList");
    });

    it("Should succeed if the RLP item is a long list", async () => {
      expect(await contract.iterator(ethers.utils.hexDataSlice(eip1559Transaction, 1))).to.not.be.reverted;
    });
  });

  describe("payloadLocation", () => {
    it("Should return the correct payload location", async () => {
      const payload = await contract.payloadLocation(ethers.utils.hexDataSlice(eip1559Transaction, 1));
      expect(payload.itemlen.toNumber()).to.equal(648);
      expect(payload.ptr.toNumber() - payload.rlpItemPtr.toNumber()).to.equal(3);
    });
  });

  describe("toAddress", () => {
    it("Should revert if rlp encoded bytes length !== 21", async () => {
      const encodedAddress = ethers.utils.RLP.encode(`${account.address}01`);
      await expect(contract.toAddress(encodedAddress)).to.be.revertedWithCustomError(contract, "WrongBytesLength");
    });

    it("Should convert rlp encoded bytes to address", async () => {
      const encodedAddress = ethers.utils.RLP.encode(account.address);
      const address = await contract.toAddress(encodedAddress);
      expect(address).to.equal(account.address);
    });
  });

  describe("toBytes", () => {
    it("Should revert if rlp encoded bytes length == 0", async () => {
      await expect(contract.toBytes("0x")).to.be.revertedWithCustomError(contract, "WrongBytesLength");
    });

    it("Should convert rlp encoded bytes to bytes", async () => {
      const encodedBytes = ethers.utils.RLP.encode("0x12345678");
      const bytes = await contract.toBytes(encodedBytes);
      expect(bytes).to.equal("0x12345678");
    });
  });

  describe("toUint", () => {
    it("Should revert if rlp encoded bytes length == 0", async () => {
      await expect(contract.toUint("0x")).to.be.revertedWithCustomError(contract, "WrongBytesLength");
    });

    it("Should revert if rlp encoded bytes length > 33", async () => {
      await expect(
        contract.toUint(
          ethers.utils.hexDataSlice(
            "0x02f887800b4d8202918301a48a94000000000000000000000000000000000000000080b864f4b476e100000000000000000000000000000000000000000000000000000000000000020000000000000000000000000000000000000000000000000000000000000001497569c1f1e97fb9eca5b5f7c7153f2eb6f9edd312bd4d0dc529f15ae67a7cbec0",
            1,
          ),
        ),
      ).to.be.revertedWithCustomError(contract, "WrongBytesLength");
    });

    it("Should convert rlp encoded bytes to uint", async () => {
      const encodedBytes = ethers.utils.RLP.encode("0x28");
      const uint = await contract.toUint(encodedBytes);
      expect(uint).to.equal(40);
    });
  });
});
