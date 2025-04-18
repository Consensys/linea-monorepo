import { useEffect, useState } from "react";
import { useAccount } from "wagmi";
import BridgeTwoLogo from "@/components/bridge/bridge-two-logo";
import styles from "./claiming.module.scss";
import SettingIcon from "@/assets/icons/setting.svg";
import AdvancedSettings from "@/components/bridge/modal/advanced-settings";
import Skeleton from "@/components/bridge/claiming/skeleton";
import ReceivedAmount from "./received-amount";
import Fees from "./fees";
import { useFormStore, useChainStore } from "@/stores";
import BridgeMode from "./bridge-mode";

export default function Claiming() {
  const { isConnected } = useAccount();
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();

  const [loading, setLoading] = useState<boolean>(false);
  const [showAdvancedSettingsModal, setShowAdvancedSettingsModal] = useState<boolean>(false);

  const amount = useFormStore((state) => state.amount);
  const balance = useFormStore((state) => state.balance);
  const isTokenCanonicalUSDC = useFormStore((state) => state.isTokenCanonicalUSDC);

  const originChainBalanceTooLow = amount && balance < amount;

  useEffect(() => {
    setLoading(true);
    const timeout = setTimeout(() => {
      setLoading(false);
    }, 1000);

    return () => clearTimeout(timeout);
  }, [amount]);

  if (!amount || amount <= 0n) return null;
  if (isConnected && originChainBalanceTooLow) return null;

  return (
    <div className={styles["wrapper"]}>
      <div className={styles.top}>
        <p className={styles.title}>Receive</p>
        <div className={styles.config}>
          <BridgeMode />
          {
            // There is no auto-claiming for USDC via CCTPV2
            !isTokenCanonicalUSDC() && (
              <button className={styles.setting} type="button" onClick={() => setShowAdvancedSettingsModal(true)}>
                <SettingIcon />
              </button>
            )
          }
        </div>
      </div>

      {loading ? (
        <Skeleton />
      ) : (
        <div className={styles.content}>
          <div className={styles.result}>
            <BridgeTwoLogo
              src1={fromChain?.iconPath ?? ""}
              src2={toChain?.iconPath ?? ""}
              alt1={fromChain?.nativeCurrency.symbol ?? ""}
              alt2={toChain?.nativeCurrency.symbol ?? ""}
            />
            <ReceivedAmount />
          </div>
          <Fees />
        </div>
      )}
      <AdvancedSettings
        isModalOpen={showAdvancedSettingsModal}
        onCloseModal={() => setShowAdvancedSettingsModal(false)}
      />
    </div>
  );
}
