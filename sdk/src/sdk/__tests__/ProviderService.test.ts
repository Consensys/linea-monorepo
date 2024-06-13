import { describe, it, expect } from "@jest/globals";
import { ethers, JsonRpcProvider } from "ethers";
import ProviderService from "../ProviderService";
import { TEST_L1_SIGNER_PRIVATE_KEY, TEST_RPC_URL } from "../../utils/testing/constants";

describe("ProviderService", () => {
  describe("getSigner", () => {
    it("should throw an error when private key is invalid", () => {
      const providerService = new ProviderService(TEST_RPC_URL);
      expect(() => providerService.getSigner("private-key")).toThrow(
        "Something went wrong when trying to generate Wallet. Please check your private key and the provider url.",
      );
    });

    it("should return ethers Signer", () => {
      const providerService = new ProviderService(TEST_RPC_URL);
      expect(JSON.stringify(providerService.getSigner(TEST_L1_SIGNER_PRIVATE_KEY))).toStrictEqual(
        JSON.stringify(new ethers.Wallet(TEST_L1_SIGNER_PRIVATE_KEY, new JsonRpcProvider(TEST_RPC_URL))),
      );
    });
  });
});
