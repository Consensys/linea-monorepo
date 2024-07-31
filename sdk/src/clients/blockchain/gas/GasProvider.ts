import { Block, JsonRpcProvider, TransactionReceipt, TransactionRequest, TransactionResponse } from "ethers";
import { DefaultGasProvider } from "./DefaultGasProvider";
import { LineaGasProvider } from "./LineaGasProvider";
import { IChainQuerier } from "../../../core/clients/blockchain/IChainQuerier";
import { GasFees, GasProviderConfig, IGasProvider, LineaGasFees } from "../../../core/clients/blockchain/IGasProvider";
import { Direction } from "../../../core/enums/MessageEnums";
import { BaseError } from "../../../core/errors/Base";

export class GasProvider implements IGasProvider<TransactionRequest> {
  private defaultGasProvider: DefaultGasProvider;
  private lineaGasProvider: LineaGasProvider;

  /**
   * Creates an instance of `GasProvider`.
   *
   * @param {IChainQuerier} chainQuerier - The chain querier for interacting with the blockchain.
   * @param {GasProviderConfig} config - The configuration for the gas provider.
   */
  constructor(
    private chainQuerier: IChainQuerier<
      TransactionReceipt,
      Block,
      TransactionRequest,
      TransactionResponse,
      JsonRpcProvider
    >,
    private readonly config: GasProviderConfig,
  ) {
    this.defaultGasProvider = new DefaultGasProvider(this.chainQuerier, {
      maxFeePerGas: config.maxFeePerGas,
      gasEstimationPercentile: config.gasEstimationPercentile,
      enforceMaxGasFee: config.enforceMaxGasFee,
    });
    this.lineaGasProvider = new LineaGasProvider(this.chainQuerier, {
      maxFeePerGas: config.maxFeePerGas,
      enforceMaxGasFee: config.enforceMaxGasFee,
    });
  }

  /**
   * Fetches gas fee estimates.
   * Determines which gas provider to use based on the presence of transactionRequest.
   *
   * @param {TransactionRequest} [transactionRequest] - Optional transaction request to determine specific gas provider.
   * @returns {Promise<GasFees | LineaGasFees>} A promise that resolves to an object containing gas fee estimates.
   * @throws {BaseError} If transactionRequest is not provided when required.
   */
  public async getGasFees(transactionRequest?: TransactionRequest): Promise<GasFees | LineaGasFees> {
    if (this.config.direction === Direction.L1_TO_L2) {
      if (this.config.enableLineaEstimateGas) {
        if (!transactionRequest) {
          throw new BaseError(
            "You need to provide transaction request as param to call the getGasFees function on Linea.",
          );
        }
        return this.lineaGasProvider.getGasFees(transactionRequest);
      } else {
        const fees = await this.defaultGasProvider.getGasFees();
        const gasLimit = await this.chainQuerier.estimateGas({
          ...transactionRequest,
          maxPriorityFeePerGas: fees.maxPriorityFeePerGas,
          maxFeePerGas: fees.maxFeePerGas,
        });
        return {
          ...fees,
          gasLimit,
        };
      }
    } else {
      return this.defaultGasProvider.getGasFees();
    }
  }

  /**
   * Gets the maximum fee per gas as configured.
   *
   * @returns {bigint} The maximum fee per gas.
   */
  public getMaxFeePerGas(): bigint {
    if (this.config.direction === Direction.L1_TO_L2) {
      return this.lineaGasProvider.getMaxFeePerGas();
    }
    return this.defaultGasProvider.getMaxFeePerGas();
  }
}
