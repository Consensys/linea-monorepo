'use client';

import React, { useContext } from 'react';
import { UIContext } from '@/contexts/ui.context';
import ToolTip from '../toolTip/ToolTip';
import Shortcut from '../shortcut/Shortcut';

const Hero: React.FC = () => {
  const { toggleShowBridge } = useContext(UIContext);
  return (
    <>
      <div className="flex flex-col items-center justify-center h-full gap-8 text-center relative z-10 pt-24">
        <h1 className="text-4xl md:text-[4rem] leading-tight text-white">
          How would you like <br /> to bridge your funds?
        </h1>
        <div className="flex flex-wrap justify-center gap-3">
          <a
            href="https://portfolio.metamask.io/bridge?destChain=59144"
            target="_blank"
            className="btn btn-custom btn-primary rounded-full text-sm md:text-[0.9375rem] font-medium"
            rel="noopener"
          >
            Metamask bridge
          </a>
          <a
            href="https://linea.build/apps"
            target="_blank"
            rel="noreferrer"
            className="btn btn-custom btn-outline btn-primary border-primary rounded-full !text-white hover:!text-black text-sm md:text-[0.9375rem] font-medium"
          >
            Third-party bridges
          </a>
          <button
            id="native-bridge-btn"
            className="btn btn-custom bg-white text-black rounded-full hover:bg-primary text-sm md:text-[0.9375rem] font-medium"
            onClick={() => toggleShowBridge(true)}
          >
            Linea Native Bridge
            <ToolTip text="Slow: Use this bridge for non time-sensitive bridge transfers" className="min-w-60">
              <svg width={13} height={13} viewBox="0 0 13 13" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path
                  d="M6.50006 11.9879C9.53099 11.9879 11.988 9.53086 11.988 6.49994C11.988 3.46901 9.53099 1.01196 6.50006 1.01196C3.46914 1.01196 1.01208 3.46901 1.01208 6.49994C1.01208 9.53086 3.46914 11.9879 6.50006 11.9879Z"
                  stroke="#121212"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
                <path d="M6.50006 9.20312V6.00793" stroke="#121212" strokeLinecap="round" strokeLinejoin="round" />
                <path d="M6.50006 4.11108H6.50506" stroke="#121212" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </ToolTip>
          </button>
        </div>
      </div>

      <Shortcut />
    </>
  );
};

export default Hero;
