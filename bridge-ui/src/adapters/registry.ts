import { Chain, Token } from "@/types";

import { cctpAdapter } from "./cctp";
import { nativeAdapter } from "./native";

import type { BridgeAdapter } from "./types";

const adapters: readonly BridgeAdapter[] = [nativeAdapter, cctpAdapter];

function isEnabled(adapter: BridgeAdapter): boolean {
  return adapter.isEnabled?.() ?? true;
}

export function getAdapter(token: Token, fromChain: Chain, toChain: Chain): BridgeAdapter | undefined {
  return adapters.find((a) => isEnabled(a) && a.canHandle(token, fromChain, toChain));
}

export function getAdapterById(id: string): BridgeAdapter | undefined {
  return adapters.find((a) => isEnabled(a) && a.id === id);
}

export function getAllAdapters(): readonly BridgeAdapter[] {
  return adapters.filter(isEnabled);
}
