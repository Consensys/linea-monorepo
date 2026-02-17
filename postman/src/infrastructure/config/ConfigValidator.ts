import { isAddress, parseAbiItem } from "viem";

import { validateEventsFiltersConfig } from "../../application/config/getConfig";

import type { ListenerConfig } from "../../application/config/PostmanConfig";

export function isViemAddressValid(address: string): boolean {
  return isAddress(address);
}

export function isViemFunctionInterfaceValid(functionInterface: string): boolean {
  try {
    parseAbiItem(functionInterface);
    return true;
  } catch {
    return false;
  }
}

export function validateEventsFilters(eventFilters: ListenerConfig["eventFilters"]): void {
  validateEventsFiltersConfig(eventFilters, isViemAddressValid, isViemFunctionInterfaceValid);
}
