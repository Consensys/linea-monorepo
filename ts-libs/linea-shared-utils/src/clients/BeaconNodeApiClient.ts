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

  async getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[] | undefined> {
    const url = `${this.rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    this.logger.debug(`getPendingPartialWithdrawals making GET request to url=${url}`);
    const { data } = await this.retryService.retry(() => axios.get<PendingPartialWithdrawalResponse>(url));
    if (data === undefined || data?.data === undefined) {
      this.logger.error("Failed GET request to", url);
      return undefined;
    }
    const returnVal = data.data;
    this.logger.debug("getPendingPartialWithdrawals return value", { returnVal });
    return returnVal;
  }
}
