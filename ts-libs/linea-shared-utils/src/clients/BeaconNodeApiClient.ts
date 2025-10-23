import {
  IBeaconNodeAPIClient,
  PendingPartialWithdrawal,
  PendingPartialWithdrawalResponse,
} from "../core/client/IBeaconNodeApiClient";
import axios from "axios";

// TODO - Add logger for error case
export class BeaconNodeApiClient implements IBeaconNodeAPIClient {
  constructor(private readonly rpcURL: string) {}

  async getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[]> {
    const url = `${this.rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    const { data } = await axios.get<PendingPartialWithdrawalResponse>(url);
    return data.data ?? [];
  }
}
