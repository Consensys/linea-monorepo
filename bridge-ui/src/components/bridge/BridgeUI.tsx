"use client";

import React from "react";
import { useAccount } from "wagmi";
import WrongNetwork from "./WrongNetwork";
import TermsModal from "../terms/TermsModal";
import Bridge from "@/components/bridge/forms/Bridge";
import { AnimatePresence, motion } from "framer-motion";
import { NetworkType } from "@/config";
import { useChainStore } from "@/stores/chainStore";

const BridgeUI: React.FC = () => {
  const networkType = useChainStore((state) => state.networkType);

  // Hooks
  const { isConnected } = useAccount();

  return (
    <AnimatePresence>
      <motion.div
        key="bridge-tech-operator"
        initial={{ opacity: 0, y: 200 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.4 }}
        onAnimationStart={() => {
          if (!isConnected) {
            document.body.style.overflow = "hidden";
          }
        }}
        onAnimationComplete={() => {
          document.body.style.overflow = "";
        }}
      >
        <div className="space-y-5">
          <div className="card w-full bg-base-100 shadow-xl md:w-[500px]">
            {networkType !== NetworkType.WRONG_NETWORK || !isConnected ? <Bridge /> : <WrongNetwork />}
          </div>
          {/* <Debug /> */}
        </div>
      </motion.div>
      <TermsModal />
    </AnimatePresence>
  );
};

export default BridgeUI;
