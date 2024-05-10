'use client';

import { useContext } from 'react';

import { ChainContext, NetworkLayer } from '@/contexts/chain.context';

export default function Debug() {
  // Context
  const context = useContext(ChainContext);
  const {
    networkType,
    networkLayer,
    tokenBridgeAddress,
    activeChain,
    alternativeChain,
    fromChain,
    toChain,
    token,
    messageServiceAddress,
  } = context;

  return (
    <div className="collapse">
      <input type="checkbox" />
      <button className="collapse-title font-medium btn btn-ghost text-xs opacity-30">Debug</button>
      <div className="collapse-content">
        <div className="card w-full md:w-[470px] bg-base-100 shadow-xl mt-5">
          <div className="text-xs p-5">
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
                  <td>{activeChain?.name}</td>
                </tr>
                <tr className="text-left">
                  <td>alternativeChain:</td>
                  <td>{alternativeChain?.name}</td>
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
