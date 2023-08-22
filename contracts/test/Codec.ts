import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { Contract } from "ethers";
import {
  Add_L1L2_Message_Hashes_Calldata_With_Empty_Array,
  Add_L1L2_Message_Hashes_Calldata_With_Five_Hashes,
  Add_L1L2_Message_Hashes_Calldata_With_One_Hash,
  L1L2_FiveHashes,
  Single_Item_L1L2_HashArray,
} from "./utils/constants";
import { deployFromFactory } from "./utils/deployment";

describe("Codec V2 Library", () => {
  let contract: Contract;

  async function deployTestCodecV2Fixture() {
    return deployFromFactory("TestCodecV2");
  }
  beforeEach(async () => {
    contract = await loadFixture(deployTestCodecV2Fixture);
  });

  describe("Codec Extracts Hashes from addL1L2MessageHashes calldata", () => {
    it("addL1L2MessageHashes extracts five hash array", async () => {
      const hashes = await contract.extractHashesTest(Add_L1L2_Message_Hashes_Calldata_With_Five_Hashes);
      expect(hashes).to.deep.equal(L1L2_FiveHashes);
    });
    it("addL1L2MessageHashes extracts one hash in array", async () => {
      const hashes = await contract.extractHashesTest(Add_L1L2_Message_Hashes_Calldata_With_One_Hash);
      expect(hashes).to.deep.equal(Single_Item_L1L2_HashArray);
    });
    it("addL1L2MessageHashes extracts empty array", async () => {
      const hashes = await contract.extractHashesTest(Add_L1L2_Message_Hashes_Calldata_With_Empty_Array);
      expect(hashes).to.be.empty;
    });
  });
});
