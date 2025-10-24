import {
  IBeaconNodeAPIClient,
  PendingPartialWithdrawal,
  PendingPartialWithdrawalResponse,
} from "../core/client/IBeaconNodeApiClient";
import axios from "axios";
import { ILogger } from "../logging/ILogger";
import { IRetryService } from "../core/services/IRetryService";

export class BeaconNodeApiClient implements IBeaconNodeAPIClient {
  constructor(
    private readonly logger: ILogger,
    private readonly retryService: IRetryService,
    private readonly rpcURL: string,
  ) {}

  async getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[]> {
    const url = `${this.rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    this.logger.debug(`getPendingPartialWithdrawals making GET request to url=${url}`);
    const { data } = await this.retryService.retry(() => axios.get<PendingPartialWithdrawalResponse>(url));
    if (!data?.data) {
      this.logger.warn("getPendingPartialWithdrawals: no pending_partial_withdrawals returned");
    }
    const returnVal = data.data ?? [];
    this.logger.debug("getPendingPartialWithdrawals return value", returnVal);
    return returnVal;
  }
}
