import { IBlockchainClient } from "@consensys/linea-shared-utils";
import { Address, getContract, GetContractReturnType, parseEventLogs, PublicClient, TransactionReceipt } from "viem";
import { IDashboard } from "../../core/clients/contracts/IDashboard.js";
import { DashboardABI } from "../../core/abis/Dashboard.js";

/**
 * Client for interacting with Dashboard smart contracts.
 * Provides methods for extracting payment information from transaction receipts by decoding contract events.
 */
export class DashboardContractClient implements IDashboard<TransactionReceipt> {
  private readonly contract: GetContractReturnType<typeof DashboardABI, PublicClient, Address>;

  /**
   * Creates a new DashboardContractClient instance.
   *
   * @param {IBlockchainClient<PublicClient, TransactionReceipt>} contractClientLibrary - Blockchain client for reading contract data.
   * @param {Address} contractAddress - The address of the Dashboard contract.
   */
  constructor(
    private readonly contractClientLibrary: IBlockchainClient<PublicClient, TransactionReceipt>,
    private readonly contractAddress: Address,
  ) {
    this.contract = getContract({
      abi: DashboardABI,
      address: contractAddress,
      client: this.contractClientLibrary.getBlockchainClient(),
    });
  }

  /**
   * Gets the address of the Dashboard contract.
   *
   * @returns {Address} The contract address.
   */
  getAddress(): Address {
    return this.contractAddress;
  }

  /**
   * Gets the viem contract instance.
   *
   * @returns {GetContractReturnType} The contract instance.
   */
  getContract(): GetContractReturnType {
    return this.contract;
  }

  /**
   * Extracts the node operator fee amount from a transaction receipt by decoding FeeDisbursed events.
   * Only decodes logs emitted by this contract. Skips unrelated logs (from the same contract or different ABIs).
   * If event not found, returns 0n.
   *
   * @param {TransactionReceipt} txReceipt - The transaction receipt to search for FeeDisbursed events.
   * @returns {bigint} The fee amount from the FeeDisbursed event, or 0n if the event is not found.
   */
  getNodeOperatorFeesPaidFromTxReceipt(txReceipt: TransactionReceipt): bigint {
    const logs = parseEventLogs({
      abi: this.contract.abi,
      eventName: "FeeDisbursed",
      logs: txReceipt.logs,
    });

    const fee =
      logs.find((log) => log.address.toLowerCase() === this.contractAddress.toLowerCase())?.args.fee ?? 0n;
    return fee;
  }
}

