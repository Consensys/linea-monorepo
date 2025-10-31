import { IBlockchainClient, ILogger } from "@consensys/linea-shared-utils";
import {
  Address,
  encodeFunctionData,
  getContract,
  GetContractReturnType,
  PublicClient,
  TransactionReceipt,
  WatchContractEventReturnType,
} from "viem";
import { LazyOracleABI } from "../../core/abis/LazyOracle.js";
import {
  ILazyOracle,
  UpdateVaultDataParams,
  LazyOracleReportData,
  WaitForVaultReportDataEventResult,
  VaultReportResult,
} from "../../core/clients/contracts/ILazyOracle.js";
import { OperationTrigger } from "../../core/metrics/LineaNativeYieldAutomationServiceMetrics.js";

export class LazyOracleContractClient implements ILazyOracle<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof LazyOracleABI, PublicClient, Address>;
  constructor(
    private readonly logger: ILogger,
    private readonly contractClientLibrary: IBlockchainClient<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
    private readonly pollIntervalMs: number,
    private readonly eventWatchTimeoutMs: number,
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
    const returnVal = {
      timestamp: resp[0],
      refSlot: resp[1],
      treeRoot: resp[2],
      reportCid: resp[3],
    };
    this.logger.debug("latestReportData", { returnVal });
    return returnVal;
  }

  async updateVaultData(params: UpdateVaultDataParams): Promise<TransactionReceipt> {
    const { vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof } =
      params;
    const calldata = encodeFunctionData({
      abi: this.contract.abi,
      functionName: "updateVaultData",
      args: [vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof],
    });
    this.logger.debug(`updateVaultData started`, { params });
    const txReceipt = await this.contractClientLibrary.sendSignedTransaction(this.contractAddress, calldata);
    this.logger.info(`updateVaultData succeeded, txHash=${txReceipt.transactionHash}`, { params });
    return txReceipt;
  }

  async simulateUpdateVaultData(params: UpdateVaultDataParams): Promise<void> {
    const { vault, totalValue, cumulativeLidoFees, liabilityShares, maxLiabilityShares, slashingReserve, proof } =
      params;
    await this.contract.simulate.updateVaultData([
      vault,
      totalValue,
      cumulativeLidoFees,
      liabilityShares,
      maxLiabilityShares,
      slashingReserve,
      proof,
    ]);
  }

  async waitForVaultsReportDataUpdatedEvent(): Promise<WaitForVaultReportDataEventResult> {
    // Create placeholder Promise resolve fn
    let resolvePromise: (value: WaitForVaultReportDataEventResult) => void;
    // Create Promise that we will return
    // Set placeholder resolve fn, to resolve fn here - decouple Promise creation from resolve fn creation
    const waitForEvent = new Promise<WaitForVaultReportDataEventResult>((resolve) => {
      resolvePromise = resolve;
    });

    // Create placeholders for unwatch fns
    let unwatchEvent: WatchContractEventReturnType | undefined;
    let unwatchTimeout: (() => void) | undefined;

    const cleanup = () => {
      if (unwatchTimeout) {
        unwatchTimeout();
        unwatchTimeout = undefined;
      }
      if (unwatchEvent) {
        unwatchEvent();
        unwatchEvent = undefined;
      }
    };

    // Start timeout
    this.logger.info(`waitForVaultsReportDataUpdatedEvent started with timeout=${this.eventWatchTimeoutMs}ms`);
    const timeoutId = setTimeout(() => {
      cleanup();
      this.logger.info(`waitForVaultsReportDataUpdatedEvent timed out after timeout=${this.eventWatchTimeoutMs}ms`);
      resolvePromise({ result: OperationTrigger.TIMEOUT });
    }, this.eventWatchTimeoutMs);
    unwatchTimeout = () => clearTimeout(timeoutId);

    // Start event watching
    unwatchEvent = this.contractClientLibrary.getBlockchainClient().watchContractEvent({
      address: this.contractAddress,
      abi: this.contract.abi,
      eventName: "VaultsReportDataUpdated",
      pollingInterval: this.pollIntervalMs,
      onLogs: (logs) => {
        // Filter out removed logs - could be due to reorg
        const nonRemovedLogs = logs.filter((log) => !log.removed);
        if (nonRemovedLogs.length === 0) {
          this.logger.warn("waitForVaultsReportDataUpdatedEvent: Dropped VaultsReportDataUpdated event");
          return;
        }

        if (nonRemovedLogs.length !== logs.length) {
          this.logger.debug("waitForVaultsReportDataUpdatedEvent: Ignored removed reorg logs", { logs });
        }

        // Filter out logs that don't fit our interface
        const firstEvent = nonRemovedLogs[0];
        if (
          firstEvent.args?.timestamp === undefined ||
          firstEvent.args?.refSlot === undefined ||
          firstEvent.args?.root === undefined ||
          firstEvent.args?.cid === undefined
        ) {
          return;
        }

        // Success -> cleanup and return
        cleanup();
        const result: VaultReportResult = {
          result: OperationTrigger.VAULTS_REPORT_DATA_UPDATED_EVENT,
          txHash: firstEvent.transactionHash,
          report: {
            timestamp: firstEvent.args?.timestamp,
            refSlot: firstEvent.args?.refSlot,
            treeRoot: firstEvent.args?.root,
            reportCid: firstEvent.args?.cid,
          },
        };
        this.logger.info("waitForVaultsReportDataUpdatedEvent detected", {
          result,
        });
        resolvePromise(result);
      },
      onError: (error) => {
        // Tolerate errors, we don't want to interrupt L1MessageService<->YieldProvider rebalancing
        this.logger.error("waitForVaultsReportDataUpdatedEvent error", { error });
      },
    });

    return await waitForEvent;
  }
}
