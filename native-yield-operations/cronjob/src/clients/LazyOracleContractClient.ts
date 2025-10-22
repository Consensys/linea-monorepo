import { IContractClientLibrary } from "ts-libs/linea-shared-utils/src/core/client/IContractClientLibrary";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  Hex,
  PublicClient,
  TransactionReceipt,
} from "viem";
import { LazyOracleABI } from "../core/abis/LazyOracle";
import { ILazyOracle, WaitForVaultsReportDataUpdatedEventReturnType } from "../core/services/contracts/ILazyOracle";
import { LazyOracleReportData } from "../core/entities";
import { IBaseContractClient } from "../core/clients/IBaseContractClient";
import { ILogger } from "ts-libs/linea-shared-utils/dist";

export class LazyOracleContractClient implements ILazyOracle<TransactionReceipt>, IBaseContractClient {
  private readonly contract: GetContractReturnType<typeof LazyOracleABI, PublicClient, Address>;
  constructor(
    private readonly contractClientLibrary: IContractClientLibrary<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
    private readonly logger: ILogger,
    private readonly pollIntervalMs: number,
  ) {
    this.contract = getContract({
      abi: LazyOracleABI,
      address: contractAddress,
      client: contractClientLibrary.getBlockchainClient(),
    });
  }

  getAddress(): Address {
    return this.contractAddress;
  }

  getContract(): GetContractReturnType {
    return this.contract;
  }

  async latestReportData(): Promise<LazyOracleReportData> {
    const resp = await this.contract.read.latestReportData();
    return {
      timestamp: resp[0],
      refSlot: resp[1],
      treeRoot: resp[2],
      reportCid: resp[3],
    };
  }

  async updateVaultData(
    vault: Address,
    totalValue: bigint,
    cumulativeLidoFees: bigint,
    liabilityShares: bigint,
    maxLiabilityShares: bigint,
    slashingReserve: bigint,
    proof: Hex[],
  ): Promise<TransactionReceipt | null> {
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "updateVaultData",
      args: [vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof],
    });
    return await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
  }

  waitForVaultsReportDataUpdatedEvent(): WaitForVaultsReportDataUpdatedEventReturnType {
    // Create placeholder variable. Initialize to empty fn so it's callable before being reassigned (TS safety).
    let resolvePromise: () => void = () => {};
    // Create the waitForEvent Promise, create a 2nd reference to 'resolve' fn
    // Decouple Promise creation from resolve fn creation
    const waitForEvent = new Promise<void>((resolve) => {
      resolvePromise = resolve;
    });

    const unwatch = this.contractClientLibrary.getBlockchainClient().watchContractEvent({
      address: this.contractAddress,
      abi: this.contract.abi,
      eventName: "VaultsReportDataUpdated",
      pollingInterval: this.pollIntervalMs,
      onLogs: (logs) => {
        // Filter out reorgs
        const valid = logs.find((l) => !l.removed);
        if (!valid) return;
        this.logger.info("VaultsReportDataUpdated detected");
        // Call resolve through 2nd reference
        resolvePromise();
      },
      onError: (err) => {
        this.logger.error({ err }, "watchContractEvent error");
      },
    });

    return { unwatch, waitForEvent };
  }
}
