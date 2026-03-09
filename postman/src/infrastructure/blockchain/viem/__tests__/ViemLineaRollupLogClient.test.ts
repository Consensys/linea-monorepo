import { describe, it, expect, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import * as viemActions from "viem/actions";

import { ViemLineaRollupLogClient } from "../ViemLineaRollupLogClient";

import type { PublicClient } from "viem";

const CONTRACT_ADDRESS = "0x1000000000000000000000000000000000000001";
const FROM_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
const TO_ADDRESS = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";
const MSG_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010";
const TX_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020";

jest.mock("viem/actions", () => ({
  getContractEvents: jest.fn(),
}));

describe("ViemLineaRollupLogClient", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;
  let logClient: ViemLineaRollupLogClient;
  const getContractEventsMock = viemActions.getContractEvents as jest.MockedFunction<
    typeof viemActions.getContractEvents
  >;

  beforeEach(() => {
    publicClient = mock<PublicClient>();
    logClient = new ViemLineaRollupLogClient(publicClient, CONTRACT_ADDRESS);
    getContractEventsMock.mockReset();
  });

  describe("getMessageSentEvents", () => {
    it("returns mapped MessageSent events", async () => {
      getContractEventsMock.mockResolvedValue([
        {
          removed: false,
          blockNumber: 100n,
          transactionHash: TX_HASH,
          logIndex: 2,
          args: {
            _messageHash: MSG_HASH,
            _from: FROM_ADDRESS,
            _to: TO_ADDRESS,
            _fee: 0n,
            _value: 1000n,
            _nonce: 5n,
            _calldata: "0x",
          },
        },
      ] as never);

      const events = await logClient.getMessageSentEvents({ fromBlock: 90, toBlock: 110 });

      expect(events).toHaveLength(1);
      expect(events[0]).toMatchObject({
        messageHash: MSG_HASH,
        messageSender: FROM_ADDRESS,
        destination: TO_ADDRESS,
        fee: 0n,
        value: 1000n,
        messageNonce: 5n,
        calldata: "0x",
        blockNumber: 100,
        transactionHash: TX_HASH,
        logIndex: 2,
      });
    });

    it("filters out removed events", async () => {
      getContractEventsMock.mockResolvedValue([
        {
          removed: true,
          blockNumber: 100n,
          logIndex: 0,
          transactionHash: TX_HASH,
          args: {
            _messageHash: MSG_HASH,
            _from: FROM_ADDRESS,
            _to: TO_ADDRESS,
            _fee: 0n,
            _value: 0n,
            _nonce: 1n,
            _calldata: "0x",
          },
        },
      ] as never);

      const events = await logClient.getMessageSentEvents({});
      expect(events).toHaveLength(0);
    });

    it("filters by fromBlockLogIndex within same block", async () => {
      getContractEventsMock.mockResolvedValue([
        {
          removed: false,
          blockNumber: 100n,
          logIndex: 0,
          transactionHash: TX_HASH,
          args: {
            _messageHash: MSG_HASH,
            _from: FROM_ADDRESS,
            _to: TO_ADDRESS,
            _fee: 0n,
            _value: 0n,
            _nonce: 1n,
            _calldata: "0x",
          },
        },
        {
          removed: false,
          blockNumber: 100n,
          logIndex: 3,
          transactionHash: TX_HASH,
          args: {
            _messageHash: MSG_HASH,
            _from: FROM_ADDRESS,
            _to: TO_ADDRESS,
            _fee: 0n,
            _value: 0n,
            _nonce: 2n,
            _calldata: "0x",
          },
        },
      ] as never);

      const events = await logClient.getMessageSentEvents({ fromBlock: 100, fromBlockLogIndex: 2 });
      expect(events).toHaveLength(1);
      expect(events[0].messageNonce).toBe(2n);
    });
  });

  describe("getL2MessagingBlockAnchoredEvents", () => {
    it("returns mapped L2MessagingBlockAnchored events", async () => {
      getContractEventsMock.mockResolvedValue([
        {
          removed: false,
          blockNumber: 200n,
          transactionHash: TX_HASH,
          logIndex: 1,
          args: { l2Block: 150n },
        },
      ] as never);

      const events = await logClient.getL2MessagingBlockAnchoredEvents({
        filters: { l2Block: 150n },
        fromBlock: 190,
      });

      expect(events).toHaveLength(1);
      expect(events[0]).toMatchObject({
        l2Block: 150n,
        blockNumber: 200,
        transactionHash: TX_HASH,
        logIndex: 1,
      });
    });
  });

  describe("getMessageClaimedEvents", () => {
    it("returns mapped MessageClaimed events", async () => {
      getContractEventsMock.mockResolvedValue([
        {
          removed: false,
          blockNumber: 300n,
          transactionHash: TX_HASH,
          logIndex: 4,
          args: { _messageHash: MSG_HASH },
        },
      ] as never);

      const events = await logClient.getMessageClaimedEvents({
        filters: { messageHash: MSG_HASH },
        fromBlock: 250,
      });

      expect(events).toHaveLength(1);
      expect(events[0]).toMatchObject({
        messageHash: MSG_HASH,
        blockNumber: 300,
        transactionHash: TX_HASH,
        logIndex: 4,
      });
    });
  });
});
