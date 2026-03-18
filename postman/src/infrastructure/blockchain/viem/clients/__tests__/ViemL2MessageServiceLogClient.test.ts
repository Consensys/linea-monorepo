import { describe, it, expect, beforeEach } from "@jest/globals";
import { mock } from "jest-mock-extended";
import * as viemActions from "viem/actions";

import { ViemL2MessageServiceLogClient } from "../ViemL2MessageServiceLogClient";

import type { PublicClient } from "viem";

const CONTRACT_ADDRESS = "0x2000000000000000000000000000000000000002";
const FROM_ADDRESS = "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";
const TO_ADDRESS = "0xbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb";
const MSG_HASH = "0x1010101010101010101010101010101010101010101010101010101010101010";
const TX_HASH = "0x2020202020202020202020202020202020202020202020202020202020202020";

jest.mock("viem/actions", () => ({
  getContractEvents: jest.fn(),
}));

function makeMessageSentEvent(overrides: Partial<{ blockNumber: bigint; logIndex: number; nonce: bigint }> = {}) {
  return {
    removed: false,
    blockNumber: overrides.blockNumber ?? 100n,
    transactionHash: TX_HASH,
    logIndex: overrides.logIndex ?? 0,
    args: {
      _messageHash: MSG_HASH,
      _from: FROM_ADDRESS,
      _to: TO_ADDRESS,
      _fee: 0n,
      _value: 500n,
      _nonce: overrides.nonce ?? 1n,
      _calldata: "0x",
    },
  };
}

describe("ViemL2MessageServiceLogClient", () => {
  let publicClient: ReturnType<typeof mock<PublicClient>>;
  let logClient: ViemL2MessageServiceLogClient;
  const getContractEventsMock = viemActions.getContractEvents as jest.MockedFunction<
    typeof viemActions.getContractEvents
  >;

  beforeEach(() => {
    publicClient = mock<PublicClient>();
    logClient = new ViemL2MessageServiceLogClient(publicClient, CONTRACT_ADDRESS);
    getContractEventsMock.mockReset();
  });

  describe("getMessageSentEvents", () => {
    it("returns mapped MessageSent events", async () => {
      getContractEventsMock.mockResolvedValue([makeMessageSentEvent()] as never);

      const events = await logClient.getMessageSentEvents({ fromBlock: 90n, toBlock: 110n });

      expect(events).toHaveLength(1);
      expect(events[0]).toMatchObject({
        messageHash: MSG_HASH,
        messageSender: FROM_ADDRESS,
        destination: TO_ADDRESS,
        fee: 0n,
        value: 500n,
        messageNonce: 1n,
        blockNumber: 100,
        transactionHash: TX_HASH,
        logIndex: 0,
      });
    });

    it("filters out removed events", async () => {
      getContractEventsMock.mockResolvedValue([{ ...makeMessageSentEvent(), removed: true }] as never);
      const events = await logClient.getMessageSentEvents({});
      expect(events).toHaveLength(0);
    });

    it("filters events by fromBlockLogIndex within the fromBlock", async () => {
      getContractEventsMock.mockResolvedValue([
        makeMessageSentEvent({ blockNumber: 100n, logIndex: 0, nonce: 1n }),
        makeMessageSentEvent({ blockNumber: 100n, logIndex: 5, nonce: 2n }),
        makeMessageSentEvent({ blockNumber: 100n, logIndex: 10, nonce: 3n }),
        makeMessageSentEvent({ blockNumber: 101n, logIndex: 0, nonce: 4n }),
      ] as never);

      const events = await logClient.getMessageSentEvents({
        fromBlock: 100n,
        toBlock: 110n,
        fromBlockLogIndex: 5,
      });

      expect(events).toHaveLength(3);
      expect(events[0].logIndex).toBe(5);
      expect(events[1].logIndex).toBe(10);
      expect(events[2].logIndex).toBe(0);
    });

    it("converts number-typed block params to bigint via toBlockParam", async () => {
      getContractEventsMock.mockResolvedValue([makeMessageSentEvent()] as never);

      await logClient.getMessageSentEvents({
        fromBlock: 90 as unknown as bigint,
        toBlock: 110 as unknown as bigint,
      });

      expect(getContractEventsMock).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({
          fromBlock: 90n,
          toBlock: 110n,
        }),
      );
    });

    it("uses fallback block params when fromBlock and toBlock are undefined", async () => {
      getContractEventsMock.mockResolvedValue([makeMessageSentEvent()] as never);

      await logClient.getMessageSentEvents({});

      expect(getContractEventsMock).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({
          fromBlock: "earliest",
          toBlock: "latest",
        }),
      );
    });

    it("passes string block tags through directly", async () => {
      getContractEventsMock.mockResolvedValue([makeMessageSentEvent()] as never);

      await logClient.getMessageSentEvents({
        toBlock: "finalized" as unknown as bigint,
      });

      expect(getContractEventsMock).toHaveBeenCalledWith(
        publicClient,
        expect.objectContaining({
          toBlock: "finalized",
        }),
      );
    });
  });
});
