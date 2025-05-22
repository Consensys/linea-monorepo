import { Account, Address, BaseError, Chain, Client, Transport } from "viem";
import { linea, lineaSepolia, mainnet, sepolia } from "viem/chains";
import {
  L2_MESSAGE_SERVICE_MAINNET_ADDRESS,
  L2_MESSAGE_SERVICE_SEPOLIA_ADDRESS,
  LINEA_ROLLUP_MAINNET_ADDRESS,
  LINEA_ROLLUP_SEPOLIA_ADDRESS,
  TOKEN_BRIDGE_LINEA_MAINNET_ADDRESS,
  TOKEN_BRIDGE_LINEA_SEPOLIA_ADDRESS,
  TOKEN_BRIDGE_MAINNET_ADDRESS,
  TOKEN_BRIDGE_SEPOLIA_ADDRESS,
} from "../constants/address";

export type GetBridgeContractAddressesReturnType = {
  lineaRollup: Address;
  l2MessageService: Address;
  l1TokenBridge: Address;
  l2TokenBridge: Address;
};

export function getBridgeContractAddresses<chain extends Chain | undefined, account extends Account | undefined>(
  client: Client<Transport, chain, account>,
): GetBridgeContractAddressesReturnType {
  if (
    client.chain?.id !== lineaSepolia.id &&
    client.chain?.id !== linea.id &&
    client.chain?.id !== sepolia.id &&
    client.chain?.id !== mainnet.id
  ) {
    throw new BaseError("Client chain is not Linea or Linea Sepolia");
  }

  if (client.chain?.id === lineaSepolia.id || client.chain?.id === sepolia.id) {
    return {
      lineaRollup: LINEA_ROLLUP_SEPOLIA_ADDRESS,
      l2MessageService: L2_MESSAGE_SERVICE_SEPOLIA_ADDRESS,
      l1TokenBridge: TOKEN_BRIDGE_SEPOLIA_ADDRESS,
      l2TokenBridge: TOKEN_BRIDGE_LINEA_SEPOLIA_ADDRESS,
    };
  }

  return {
    lineaRollup: LINEA_ROLLUP_MAINNET_ADDRESS,
    l2MessageService: L2_MESSAGE_SERVICE_MAINNET_ADDRESS,
    l1TokenBridge: TOKEN_BRIDGE_MAINNET_ADDRESS,
    l2TokenBridge: TOKEN_BRIDGE_LINEA_MAINNET_ADDRESS,
  };
}
