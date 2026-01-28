import axios from "axios";
import { mock } from "jest-mock-extended";

import { PendingDeposit, PendingPartialWithdrawal } from "../../core/client/IBeaconNodeApiClient";
import { IRetryService } from "../../core/services/IRetryService";
import { ILogger } from "../../logging/ILogger";
import { BeaconNodeApiClient } from "../BeaconNodeApiClient";

jest.mock("axios");

const mockedAxios = axios as jest.Mocked<typeof axios>;

describe("BeaconNodeApiClient", () => {
  const rpcURL = "http://localhost:5051";
  let logger: jest.Mocked<ILogger>;
  let retryService: jest.Mocked<IRetryService>;
  let client: BeaconNodeApiClient;

  beforeEach(() => {
    logger = {
      name: "test",
      info: jest.fn(),
      error: jest.fn(),
      warn: jest.fn(),
      debug: jest.fn(),
    };
    retryService = mock<IRetryService>();
    retryService.retry.mockImplementation(async (fn) => fn());
    mockedAxios.get.mockReset();

    client = new BeaconNodeApiClient(logger, retryService, rpcURL);
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  it("fetches and returns pending partial withdrawals", async () => {
    const rawResponseData = [
      { validator_index: "42", amount: "1234", withdrawable_epoch: "10" },
      { validator_index: "43", amount: "5678", withdrawable_epoch: "11" },
    ];
    const expectedParsedData: PendingPartialWithdrawal[] = [
      { validator_index: 42, amount: 1234n, withdrawable_epoch: 10 },
      { validator_index: 43, amount: 5678n, withdrawable_epoch: 11 },
    ];
    mockedAxios.get.mockResolvedValue({
      data: { execution_optimistic: false, finalized: true, data: rawResponseData },
    });

    const result = await client.getPendingPartialWithdrawals();

    const expectedUrl = `${rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    expect(result).toEqual(expectedParsedData);
    expect(retryService.retry).toHaveBeenCalledTimes(1);
    expect(retryService.retry.mock.calls[0][0]).toEqual(expect.any(Function));
    expect(mockedAxios.get).toHaveBeenCalledWith(expectedUrl);
    expect(logger.debug).toHaveBeenNthCalledWith(
      1,
      `getPendingPartialWithdrawals making GET request to url=${expectedUrl}`,
    );
    expect(logger.info).toHaveBeenCalledWith(
      `getPendingPartialWithdrawals succeeded, pendingWithdrawalCount=${expectedParsedData.length}`,
    );
    expect(logger.debug).toHaveBeenNthCalledWith(2, "getPendingPartialWithdrawals return value", {
      returnVal: expectedParsedData,
    });
    expect(logger.error).not.toHaveBeenCalled();
  });

  it("logs an error and returns undefined when response payload is empty", async () => {
    mockedAxios.get.mockResolvedValue({ data: undefined });

    const result = await client.getPendingPartialWithdrawals();

    const expectedUrl = `${rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    expect(result).toBeUndefined();
    expect(logger.error).toHaveBeenCalledWith("Failed GET request to", expectedUrl);
    expect(logger.debug).toHaveBeenCalledTimes(1);
  });

  it("returns an empty array when the API responds with no withdrawals", async () => {
    mockedAxios.get.mockResolvedValue({
      data: { execution_optimistic: false, finalized: true, data: [] },
    });

    const result = await client.getPendingPartialWithdrawals();

    expect(result).toEqual([]);
    expect(logger.info).toHaveBeenCalledWith(`getPendingPartialWithdrawals succeeded, pendingWithdrawalCount=0`);
    expect(logger.error).not.toHaveBeenCalled();
    expect(logger.debug).toHaveBeenCalledTimes(2);
  });

  it("handles null data array in response", async () => {
    mockedAxios.get.mockResolvedValue({
      data: { execution_optimistic: false, finalized: true, data: null },
    });

    const result = await client.getPendingPartialWithdrawals();

    expect(result).toBeUndefined();
    expect(logger.info).toHaveBeenCalledWith(`getPendingPartialWithdrawals succeeded, pendingWithdrawalCount=0`);
    expect(logger.error).not.toHaveBeenCalled();
    expect(logger.debug).toHaveBeenCalledTimes(2);
  });

  describe("getPendingDeposits", () => {
    it("fetches and returns pending deposits", async () => {
      const rawResponseData = [
        {
          pubkey: "0x1234",
          withdrawal_credentials: "0xabcd",
          amount: "32000000000",
          signature: "0x5678",
          slot: "100",
        },
        {
          pubkey: "0x5678",
          withdrawal_credentials: "0xef01",
          amount: "32000000000",
          signature: "0x9abc",
          slot: "101",
        },
      ];
      const expectedParsedData: PendingDeposit[] = [
        {
          pubkey: "0x1234",
          withdrawal_credentials: "0xabcd",
          amount: 32000000000,
          signature: "0x5678",
          slot: 100,
        },
        {
          pubkey: "0x5678",
          withdrawal_credentials: "0xef01",
          amount: 32000000000,
          signature: "0x9abc",
          slot: 101,
        },
      ];
      mockedAxios.get.mockResolvedValue({
        data: { execution_optimistic: false, finalized: true, data: rawResponseData },
      });

      const result = await client.getPendingDeposits();

      const expectedUrl = `${rpcURL}/eth/v1/beacon/states/head/pending_deposits`;
      expect(result).toEqual(expectedParsedData);
      expect(retryService.retry).toHaveBeenCalledTimes(1);
      expect(retryService.retry.mock.calls[0][0]).toEqual(expect.any(Function));
      expect(mockedAxios.get).toHaveBeenCalledWith(expectedUrl);
      expect(logger.debug).toHaveBeenNthCalledWith(1, `getPendingDeposits making GET request to url=${expectedUrl}`);
      expect(logger.info).toHaveBeenCalledWith(
        `getPendingDeposits succeeded, pendingDepositCount=${expectedParsedData.length}`,
      );
      expect(logger.debug).toHaveBeenNthCalledWith(2, "getPendingDeposits return value", {
        returnVal: expectedParsedData,
      });
      expect(logger.error).not.toHaveBeenCalled();
    });

    it("logs an error and returns undefined when response payload is empty", async () => {
      mockedAxios.get.mockResolvedValue({ data: undefined });

      const result = await client.getPendingDeposits();

      const expectedUrl = `${rpcURL}/eth/v1/beacon/states/head/pending_deposits`;
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed GET request to", expectedUrl);
      expect(logger.debug).toHaveBeenCalledTimes(1);
    });

    it("returns an empty array when the API responds with no deposits", async () => {
      mockedAxios.get.mockResolvedValue({
        data: { execution_optimistic: false, finalized: true, data: [] },
      });

      const result = await client.getPendingDeposits();

      expect(result).toEqual([]);
      expect(logger.info).toHaveBeenCalledWith(`getPendingDeposits succeeded, pendingDepositCount=0`);
      expect(logger.error).not.toHaveBeenCalled();
      expect(logger.debug).toHaveBeenCalledTimes(2);
    });

    it("handles null data array in response", async () => {
      mockedAxios.get.mockResolvedValue({
        data: { execution_optimistic: false, finalized: true, data: null },
      });

      const result = await client.getPendingDeposits();

      expect(result).toBeUndefined();
      expect(logger.info).toHaveBeenCalledWith(`getPendingDeposits succeeded, pendingDepositCount=0`);
      expect(logger.error).not.toHaveBeenCalled();
      expect(logger.debug).toHaveBeenCalledTimes(2);
    });
  });

  describe("getCurrentEpoch", () => {
    it("fetches and returns current epoch from beacon head", async () => {
      const slot = 1892897; // This should convert to epoch 59153 (1892897 / 32 = 59153.03125, floored = 59153)
      const expectedEpoch = Math.floor(slot / 32);
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0xe419f40054c11ebd4973a08ee69c79d3ac65b9c30da4465a4d0b4bcb2718e5a6",
            canonical: true,
            header: {
              message: {
                slot: slot.toString(),
                proposer_index: "337206",
                parent_root: "0xe8bbc35c34d8f6e0d90a0428fe630d816439bdab1963a938700abedc9ffcc89d",
                state_root: "0x8311f8ec859a1af3a8e630f957b6828567364385166a98b36fb7e3c4d165cf55",
                body_root: "0x1576d630e634355976c8236a41df261c6fa29fd3bc053e4d15af234faedd9a1a",
              },
              signature:
                "0xa2c5ddf7700b92216160aae2df3f57981cf3be119a25217a72e68d06eae88c6c290127a1c3a41bd9cf2f2f100bdd960703576e2c7b8c1c6d0f0f296a9139c7971cc6b252add08b9aa7ff16a590d4dfd55027d23042afb8c6befeff9e8dcff7a2",
            },
          },
        },
      });

      const result = await client.getCurrentEpoch();

      const expectedUrl = `${rpcURL}/eth/v1/beacon/headers/head`;
      expect(result).toBe(expectedEpoch);
      expect(retryService.retry).toHaveBeenCalledTimes(1);
      expect(mockedAxios.get).toHaveBeenCalledWith(expectedUrl);
      expect(logger.debug).toHaveBeenNthCalledWith(1, `getCurrentEpoch making GET request to url=${expectedUrl}`);
      expect(logger.info).toHaveBeenCalledWith(`getCurrentEpoch succeeded, epoch=${expectedEpoch}, slot=${slot}`);
      expect(logger.debug).toHaveBeenNthCalledWith(2, "getCurrentEpoch return value", {
        epoch: expectedEpoch,
        slot: slot,
      });
      expect(logger.error).not.toHaveBeenCalled();
    });

    it("handles exact epoch boundary slots", async () => {
      const slot = 3200; // Exactly epoch 100 (3200 / 32 = 100)
      const expectedEpoch = 100;
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0x1234",
            canonical: true,
            header: {
              message: {
                slot: slot.toString(),
                proposer_index: "1",
                parent_root: "0x0000",
                state_root: "0x0000",
                body_root: "0x0000",
              },
              signature: "0x0000",
            },
          },
        },
      });

      const result = await client.getCurrentEpoch();

      expect(result).toBe(expectedEpoch);
      expect(logger.info).toHaveBeenCalledWith(`getCurrentEpoch succeeded, epoch=${expectedEpoch}, slot=${slot}`);
    });

    it("floors fractional epoch values correctly", async () => {
      const slot = 63; // Should be epoch 1 (63 / 32 = 1.96875, floored = 1)
      const expectedEpoch = 1;
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0x1234",
            canonical: true,
            header: {
              message: {
                slot: slot.toString(),
                proposer_index: "1",
                parent_root: "0x0000",
                state_root: "0x0000",
                body_root: "0x0000",
              },
              signature: "0x0000",
            },
          },
        },
      });

      const result = await client.getCurrentEpoch();

      expect(result).toBe(expectedEpoch);
    });

    it("logs an error and returns undefined when response payload is empty", async () => {
      mockedAxios.get.mockResolvedValue({ data: undefined });

      const result = await client.getCurrentEpoch();

      const expectedUrl = `${rpcURL}/eth/v1/beacon/headers/head`;
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to get slot from response", { url: expectedUrl });
      expect(logger.debug).toHaveBeenCalledTimes(1);
    });

    it("logs an error and returns undefined when data is missing", async () => {
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
        },
      });

      const result = await client.getCurrentEpoch();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to get slot from response", expect.any(Object));
    });

    it("logs an error and returns undefined when header is missing", async () => {
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0x1234",
            canonical: true,
          },
        },
      });

      const result = await client.getCurrentEpoch();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to get slot from response", expect.any(Object));
    });

    it("logs an error and returns undefined when message is missing", async () => {
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0x1234",
            canonical: true,
            header: {
              signature: "0x0000",
            },
          },
        },
      });

      const result = await client.getCurrentEpoch();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to get slot from response", expect.any(Object));
    });

    it("logs an error and returns undefined when slot is null", async () => {
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0x1234",
            canonical: true,
            header: {
              message: {
                proposer_index: "1",
                parent_root: "0x0000",
                state_root: "0x0000",
                body_root: "0x0000",
              },
              signature: "0x0000",
            },
          },
        },
      });

      const result = await client.getCurrentEpoch();

      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith("Failed to get slot from response", expect.any(Object));
    });

    it("logs an error and returns undefined when slot is invalid (NaN)", async () => {
      mockedAxios.get.mockResolvedValue({
        data: {
          execution_optimistic: false,
          finalized: false,
          data: {
            root: "0x1234",
            canonical: true,
            header: {
              message: {
                slot: "invalid",
                proposer_index: "1",
                parent_root: "0x0000",
                state_root: "0x0000",
                body_root: "0x0000",
              },
              signature: "0x0000",
            },
          },
        },
      });

      const result = await client.getCurrentEpoch();

      const expectedUrl = `${rpcURL}/eth/v1/beacon/headers/head`;
      expect(result).toBeUndefined();
      expect(logger.error).toHaveBeenCalledWith(`Invalid slot value: invalid from`, expectedUrl);
    });
  });
});
