"use client";

import React, { useContext } from "react";
import { UIContext } from "@/contexts/ui.context";
import ToolTip from "../toolTip/ToolTip";
import Shortcut from "../shortcut/Shortcut";

const Hero: React.FC = () => {
  const { toggleShowBridge } = useContext(UIContext);
  return (
    <>
      <div className="relative z-10 flex h-full flex-col items-center justify-center gap-8 pt-24 text-center">
        <h1 className="text-4xl leading-tight text-white md:text-[4rem]">
          How would you like <br /> to bridge your funds?
        </h1>
        <div className="flex flex-wrap justify-center gap-3">
          <a
            href="https://portfolio.metamask.io/bridge?destChain=59144"
            target="_blank"
            className="btn-custom btn btn-primary rounded-full text-sm font-medium uppercase md:text-[0.9375rem]"
            rel="noopener"
          >
            Metamask bridge
          </a>
          <a
            href="https://linea.build/apps?types=bridge"
            target="_blank"
            rel="noreferrer"
            className="btn-custom btn btn-outline btn-primary rounded-full border-primary text-sm font-medium uppercase !text-white hover:!text-black md:text-[0.9375rem]"
          >
            Third-party bridges
          </a>
          <button
            id="native-bridge-btn"
            className="btn-custom btn rounded-full bg-white text-sm font-medium uppercase text-black hover:bg-primary md:text-[0.9375rem]"
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
