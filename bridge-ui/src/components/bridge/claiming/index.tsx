import { useEffect, useState } from "react";

import dynamic from "next/dynamic";
import Link from "next/link";
import { useConnection } from "wagmi";

import SettingIcon from "@/assets/icons/setting.svg";
import BridgeTwoLogo from "@/components/bridge/bridge-two-logo";
import Skeleton from "@/components/bridge/claiming/skeleton";
import { ETH_SYMBOL } from "@/constants/tokens";
import { useL1MessageServiceLiquidity } from "@/hooks";
import { useChainStore } from "@/stores/chainStore";
import { useFormStore } from "@/stores/formStoreProvider";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import { BridgeProvider, CCTPMode, ChainLayer } from "@/types";
import { isCctp } from "@/utils/tokens";

import BridgeMode from "./bridge-mode";
import styles from "./claiming.module.scss";
import Fees from "./fees";
import ReceivedAmount from "./received-amount";

const AdvancedSettings = dynamic(() => import("@/components/bridge/modal/advanced-settings"), {
  ssr: false,
});

export default function Claiming() {
  const { isConnected } = useConnection();
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const setHideNoFeesPill = useNativeBridgeNavigationStore.useSetHideNoFeesPill();

  const [showAdvancedSettingsModal, setShowAdvancedSettingsModal] = useState<boolean>(false);

  const amount = useFormStore((state) => state.amount);
  const balance = useFormStore((state) => state.balance);
  const token = useFormStore((state) => state.token);
  const cctpMode = useFormStore((state) => state.cctpMode);
  const originChainBalanceTooLow = amount && balance < amount;
  const isL2ToL1EthFlow =
    fromChain.layer === ChainLayer.L2 && toChain.layer === ChainLayer.L1 && token.symbol === ETH_SYMBOL;

  const { isLowLiquidity: hasLowL1MessageServiceBalance, isLoading: isLiquidityLoading } = useL1MessageServiceLiquidity(
    {
      toChain,
      isL2ToL1Eth: isL2ToL1EthFlow,
      withdrawalAmount: amount ?? 0n,
    },
  );

  const loading = isLiquidityLoading && !!amount && amount > 0n;

  useEffect(() => {
    setHideNoFeesPill(token.bridgeProvider === BridgeProvider.CCTP && cctpMode === CCTPMode.FAST);
    return () => setHideNoFeesPill(false);
  }, [cctpMode, token.bridgeProvider, setHideNoFeesPill]);

  const showSettingIcon = fromChain.layer !== ChainLayer.L2 && !isCctp(token) && !loading;

  if (!amount || amount <= 0n) return null;
  if (isConnected && originChainBalanceTooLow) return null;

  return (
    <div className={styles["wrapper"]}>
      <div className={styles.top}>
        <p className={styles.title}>Receive</p>
        <div className={styles.config}>
          <BridgeMode />
          {showSettingIcon && (
            <button className={styles.setting} type="button" onClick={() => setShowAdvancedSettingsModal(true)}>
              <SettingIcon />
            </button>
          )}
        </div>
      </div>

      {loading ? (
        <Skeleton />
      ) : (
        <div className={styles.content}>
          <div className={styles.result}>
            <BridgeTwoLogo
              src1={token?.image ?? ""}
              src2={toChain?.iconPath ?? ""}
              alt1={token?.symbol ?? ""}
              alt2={toChain?.nativeCurrency.symbol ?? ""}
            />
            <ReceivedAmount />
          </div>
          <Fees />
        </div>
      )}
      {hasLowL1MessageServiceBalance && (
        <p className={styles.warning}>
          The bridge is currently congested.{" "}
          <Link href="https://linea.build/hub/bridge" target="_blank" rel="noopener noreferrer">
            Learn more.
          </Link>
        </p>
      )}
      {showAdvancedSettingsModal && (
        <AdvancedSettings
          isModalOpen={showAdvancedSettingsModal}
          onCloseModal={() => setShowAdvancedSettingsModal(false)}
        />
      )}
    </div>
  );
}
