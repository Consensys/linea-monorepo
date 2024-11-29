import { describe, it, expect } from "@jest/globals";
import { ethers, JsonRpcProvider } from "ethers";
import { TEST_L1_SIGNER_PRIVATE_KEY, TEST_RPC_URL } from "../../utils/testing/constants";
import { Wallet } from "../wallet";

describe("Wallet", () => {
  describe("getWallet", () => {
    it("should throw an error when private key is invalid", () => {
      expect(() => Wallet.getWallet("private-key")).toThrow(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    });

    it("should return ethers Signer", () => {
      expect(
        JSON.stringify(Wallet.getWallet(TEST_L1_SIGNER_PRIVATE_KEY).connect(new JsonRpcProvider(TEST_RPC_URL))),
      ).toStrictEqual(JSON.stringify(new ethers.Wallet(TEST_L1_SIGNER_PRIVATE_KEY, new JsonRpcProvider(TEST_RPC_URL))));
    });
  });
});
