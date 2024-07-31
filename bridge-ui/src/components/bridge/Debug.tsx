"use client";

import { NetworkLayer } from "@/config";
import { useChainStore } from "@/stores/chainStore";

export default function Debug() {
  // Context
  const { networkType, networkLayer, tokenBridgeAddress, fromChain, toChain, token, messageServiceAddress } =
    useChainStore((state) => ({
      networkType: state.networkType,
      networkLayer: state.networkLayer,
      tokenBridgeAddress: state.tokenBridgeAddress,
      fromChain: state.fromChain,
      toChain: state.toChain,
      token: state.token,
      messageServiceAddress: state.messageServiceAddress,
    }));

  return (
    <div className="collapse">
      <input type="checkbox" />
      <button className="btn collapse-title btn-ghost text-xs font-medium opacity-30">Debug</button>
      <div className="collapse-content">
        <div className="card mt-5 w-full bg-base-100 shadow-xl md:w-[470px]">
          <div className="p-5 text-xs">
            <table>
              <tbody>
                <tr className="text-left">
                  <td>networkType:</td>
                  <td>{networkType}</td>
                </tr>
                <tr className="text-left">
                  <td>networkLayer:</td>
                  <td>{networkLayer}</td>
                </tr>
                <tr className="text-left">
                  <td>activeChain:</td>
                  <td>{fromChain?.name}</td>
                </tr>
                <tr className="text-left">
                  <td>alternativeChain:</td>
                  <td>{toChain?.name}</td>
                </tr>
                <tr className="text-left">
                  <td>fromChain:</td>
                  <td>{fromChain?.name}</td>
                </tr>
                <tr className="text-left">
                  <td>toChain:</td>
                  <td>{toChain?.name}</td>
                </tr>
                <tr className="text-left">
                  <td colSpan={2}>
                    <div className="divider"></div>
                  </td>
                </tr>
                <tr className="text-left">
                  <td>tokenBridge address:</td>
                  <td>{tokenBridgeAddress}</td>
                </tr>
                <tr className="text-left">
                  <td>messageService address:</td>
                  <td>{messageServiceAddress}</td>
                </tr>
                <tr className="text-left">
                  <td colSpan={2}>
                    <div className="divider"></div>
                  </td>
                </tr>
                <tr className="text-left">
                  <td>token:</td>
                  <td>{token?.name}</td>
                </tr>
                <tr className="text-left">
                  <td>address:</td>
                  <td>{token && networkLayer !== NetworkLayer.UNKNOWN && token[networkLayer]}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
}
