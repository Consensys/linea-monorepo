import { describe, it, expect } from "@jest/globals";
import ProviderService from "../ProviderService";
import {
  TEST_L1_SIGNER_PRIVATE_KEY,
  TEST_L2_SIGNER_PRIVATE_KEY,
  TEST_RPC_URL,
} from "../../utils/testHelpers/constants";
import { LineaSDK } from "../LineaSDK";
import { L1MessageServiceContract, L2MessageServiceContract } from "../../contracts";
import { NETWORKS } from "../../utils/networks";

describe("LineaSDK", () => {
  describe("getL1Contract", () => {
    it("should return L1MessageServiceContract in read only mode when the 'mode' option is set to 'read-only'", () => {
      const sdk = new LineaSDK({
        mode: "read-only",
        network: "linea-goerli",
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      const l1MessageService = sdk.getL1Contract();
      expect(JSON.stringify(l1MessageService)).toStrictEqual(
        JSON.stringify(
          new L1MessageServiceContract(
            new ProviderService(TEST_RPC_URL).provider,
            NETWORKS["linea-goerli"].l1ContractAddress,
            "read-only",
            undefined,
            undefined,
            undefined,
          ),
        ),
      );
    });

    it("should return L1MessageServiceContract in read and write mode when the 'mode' option is set to 'read-write'", () => {
      const sdk = new LineaSDK({
        mode: "read-write",
        network: "linea-goerli",
        l1SignerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
        l2SignerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      const l1MessageService = sdk.getL1Contract();
      expect(JSON.stringify(l1MessageService)).toStrictEqual(
        JSON.stringify(
          new L1MessageServiceContract(
            new ProviderService(TEST_RPC_URL).provider,
            NETWORKS["linea-goerli"].l1ContractAddress,
            "read-write",
            new ProviderService(TEST_RPC_URL).getSigner(TEST_L1_SIGNER_PRIVATE_KEY),
            undefined,
            undefined,
          ),
        ),
      );
    });

    it("should throw an error when the 'network' option is set to 'localhost' and no local contract address has been provided to the function", () => {
      const sdk = new LineaSDK({
        mode: "read-write",
        network: "localhost",
        l1SignerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
        l2SignerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      expect(() => sdk.getL1Contract()).toThrow("You need to provide a contract address.");
    });

    it("should return L1MessageServiceContract with custom contract address when the network option is set to 'localhost'", () => {
      const sdk = new LineaSDK({
        mode: "read-write",
        network: "localhost",
        l1SignerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
        l2SignerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      const localContractAddress = "0x0000000000000000000000000000000000000001";

      const l1MessageService = sdk.getL1Contract(localContractAddress);
      expect(JSON.stringify(l1MessageService)).toStrictEqual(
        JSON.stringify(
          new L1MessageServiceContract(
            new ProviderService(TEST_RPC_URL).provider,
            localContractAddress,
            "read-write",
            new ProviderService(TEST_RPC_URL).getSigner(TEST_L1_SIGNER_PRIVATE_KEY),
            undefined,
            undefined,
          ),
        ),
      );
    });
  });

  describe("getL2Contract", () => {
    it("should return L2MessageServiceContract in read only mode when the 'mode' option is set to 'read-only'", () => {
      const sdk = new LineaSDK({
        mode: "read-only",
        network: "linea-goerli",
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      const l2MessageService = sdk.getL2Contract();
      expect(JSON.stringify(l2MessageService)).toStrictEqual(
        JSON.stringify(
          new L2MessageServiceContract(
            new ProviderService(TEST_RPC_URL).provider,
            NETWORKS["linea-goerli"].l2ContractAddress,
            "read-only",
            undefined,
            undefined,
            undefined,
          ),
        ),
      );
    });

    it("should return L2MessageServiceContract in read and write mode when the 'mode' option is set to 'read-write'", () => {
      const sdk = new LineaSDK({
        mode: "read-write",
        network: "linea-goerli",
        l1SignerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
        l2SignerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      const l2MessageService = sdk.getL2Contract();
      expect(JSON.stringify(l2MessageService)).toStrictEqual(
        JSON.stringify(
          new L2MessageServiceContract(
            new ProviderService(TEST_RPC_URL).provider,
            NETWORKS["linea-goerli"].l2ContractAddress,
            "read-write",
            new ProviderService(TEST_RPC_URL).getSigner(TEST_L2_SIGNER_PRIVATE_KEY),
            undefined,
            undefined,
          ),
        ),
      );
    });

    it("should throw an error when the 'network' option is set to 'localhost' and no local contract address has been provided to the function", () => {
      const sdk = new LineaSDK({
        mode: "read-write",
        network: "localhost",
        l1SignerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
        l2SignerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      expect(() => sdk.getL2Contract()).toThrow("You need to provide a contract address.");
    });

    it("should return L2MessageServiceContract with custom contract address when the network option is set to 'localhost'", () => {
      const sdk = new LineaSDK({
        mode: "read-write",
        network: "localhost",
        l1SignerPrivateKey: TEST_L1_SIGNER_PRIVATE_KEY,
        l2SignerPrivateKey: TEST_L2_SIGNER_PRIVATE_KEY,
        l1RpcUrl: TEST_RPC_URL,
        l2RpcUrl: TEST_RPC_URL,
      });

      const localContractAddress = "0x0000000000000000000000000000000000000001";

      const l2MessageService = sdk.getL2Contract(localContractAddress);
      expect(JSON.stringify(l2MessageService)).toStrictEqual(
        JSON.stringify(
          new L2MessageServiceContract(
            new ProviderService(TEST_RPC_URL).provider,
            localContractAddress,
            "read-write",
            new ProviderService(TEST_RPC_URL).getSigner(TEST_L2_SIGNER_PRIVATE_KEY),
            undefined,
            undefined,
          ),
        ),
      );
    });
  });
});
