import { IBeaconNodeAPIClient, PendingPartialWithdrawal } from "../core/client/IBeaconNodeApiClient";
import axios from "axios";

export class BeaconNodeApiClient implements IBeaconNodeAPIClient {
  constructor(private readonly rpcURL: string) {}

  async getPendingPartialWithdrawals(): Promise<PendingPartialWithdrawal[]> {
    const url = `${this.rpcURL}/eth/v1/beacon/states/head/pending_partial_withdrawals`;
    const { data } = await axios.get<PendingPartialWithdrawal[]>(url);
    return data;
  }
}
