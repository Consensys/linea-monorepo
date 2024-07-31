import { Block, JsonRpcProvider, Wallet } from "ethers";
import { describe, afterEach, it, beforeEach } from "@jest/globals";
import { MockProxy, mock, mockClear } from "jest-mock-extended";
import { L2ChainQuerier } from "../L2ChainQuerier";
import { TEST_L1_SIGNER_PRIVATE_KEY } from "../../../../utils/testing/constants";

describe("L2ChainQuerier", () => {
  let providerMock: MockProxy<JsonRpcProvider>;
  let chainQuerier: L2ChainQuerier;
  beforeEach(() => {
    providerMock = mock<JsonRpcProvider>();
    chainQuerier = new L2ChainQuerier(providerMock, new Wallet(TEST_L1_SIGNER_PRIVATE_KEY, providerMock));
  });

  afterEach(() => {
    mockClear(providerMock);
  });

  describe("getBlockExtraData", () => {
    it("should return block extraData", async () => {
      const blockMocked: Block = {
        baseFeePerGas: 7n,
        difficulty: 2n,
        extraData:
          "0x0100989680015eb3c80000ea600000000000000000000000000000000000000024997ceb570c667b9c369d351b384ce97dcfe0dda90696fc3b007b8d7160672548a6716cc33ffe0e4004c555a0c7edd9ddc2545a630f2276a2964dcf856e6ab501",
        gasLimit: 61000000n,
        gasUsed: 138144n,
        hash: "0xc53d0f6b65feabf0422bb897d3f5de2c32d57612f88eb4a366d8076a040a715a",
        miner: "0x0000000000000000000000000000000000000000",
        nonce: "0x0000000000000000",
        number: 1635519,
        parentHash: "0xecc7bd0b6d533b13fc65529e5da174062d93f8f426c32929914a375f00a19cc3",
        receiptsRoot: "0x0cfd91942f25f029d50078225749dbd8601dd922713e783e4210568ed45b6cd3",
        stateRoot: "0xad1346e81f574b511c917cd76c5b70f1d6852b871569bedd9ef02160746e3ffa",
        timestamp: 1718030601,
        transactions: [],
        provider: providerMock,
        parentBeaconBlockRoot: null,
        blobGasUsed: null,
        excessBlobGas: null,
      } as unknown as Block;
      jest.spyOn(providerMock, "getBlock").mockResolvedValue(blockMocked);

      expect(await chainQuerier.getBlockExtraData("latest")).toStrictEqual({
        version: 1,
        fixedCost: 10000000000,
        variableCost: 22983624000,
        ethGasPrice: 60000000,
      });
    });
  });
});
