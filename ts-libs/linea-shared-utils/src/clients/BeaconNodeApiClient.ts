import {
  IBeaconNodeAPIClient,
  PendingPartialWithdrawal,
  PendingPartialWithdrawalResponse,
} from "../core/client/IBeaconNodeApiClient";
import axios from "axios";
import { ILogger } from "../logging/ILogger";

export class BeaconNodeApiClient implements IBeaconNodeAPIClient {
  constructor(
    private readonly logger: ILogger,
    private readonly rpcURL: string,
  ) {}

  async getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[]> {
    const url = `${this.rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    this.logger.debug(`BeaconNodeApiClient: fetching ${url}`);
    const { data } = await axios.get<PendingPartialWithdrawalResponse>(url);
    if (!data?.data) {
      this.logger.warn("BeaconNodeApiClient: no pending_partial_withdrawals returned");
    }
    return data.data ?? [];
  }
}
