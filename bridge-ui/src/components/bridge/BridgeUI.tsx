'use client';

import React, { useContext } from 'react';
import WrongNetwork from './WrongNetwork';
import { ChainContext, NetworkType } from '@/contexts/chain.context';
import { useIsConnected } from '@/hooks';
import TermsModal from '../terms/TermsModal';
import History from '@/components/history/History';
import Bridge from '@/components/bridge/forms/Bridge';
import { AnimatePresence, motion } from 'framer-motion';

const BridgeUI: React.FC = () => {
  const context = useContext(ChainContext);
  const { networkType } = context;

  // Hooks
  const isConnected = useIsConnected();

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
            document.body.style.overflow = 'hidden';
          }
        }}
        onAnimationComplete={() => {
          document.body.style.overflow = '';
        }}
      >
        <div className="space-y-5">
          <div className="card w-full md:w-[500px] bg-base-100 shadow-xl">
            {networkType !== NetworkType.WRONG_NETWORK || !isConnected ? <Bridge /> : <WrongNetwork />}
          </div>
          {isConnected && <History />}
          {/* <Debug /> */}
        </div>
      </motion.div>
      <TermsModal />
    </AnimatePresence>
  );
};

export default BridgeUI;
