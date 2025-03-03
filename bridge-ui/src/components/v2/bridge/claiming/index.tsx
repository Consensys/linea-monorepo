import { BridgeForm } from "@/models";
import BridgeTwoLogo from "@/components/v2/bridge/bridge-two-logo";
import styles from "./claiming.module.scss";
import SettingIcon from "@/assets/icons/setting.svg";
import { useEffect, useState } from "react";
import { useFormContext } from "react-hook-form";
import AdvancedSettings from "@/components/v2/bridge/modal/advanced-settings";
import Skeleton from "@/components/v2/bridge/claiming/skeleton";
import { BridgeType } from "@/config/config";
import { useChainStore } from "@/stores/chainStore";
import ReceivedAmount from "./received-amount";
import Fees from "./fees";
import { parseUnits } from "viem";

export type BridgeModeOption = {
  value: BridgeType;
  label: string;
  image: string;
};

export default function Claiming() {
  const fromChain = useChainStore.useFromChain();
  const toChain = useChainStore.useToChain();
  const { watch, formState } = useFormContext<BridgeForm>();

  const [loading, setLoading] = useState<boolean>(false);
  const [showAdvancedSettingsModal, setShowAdvancedSettingsModal] = useState<boolean>(false);

  const [amount, balance, token] = watch(["amount", "balance", "token"]);

  const originChainBalanceTooLow =
    formState.errors?.amount?.message !== undefined ||
    parseUnits(balance, token.decimals) < parseUnits(amount, token.decimals);

  useEffect(() => {
    setLoading(true);
    const timeout = setTimeout(() => {
      setLoading(false);
    }, 1000);

    return () => clearTimeout(timeout);
  }, [amount]);

  if (!amount || parseFloat(amount) <= 0) return null;
  if (originChainBalanceTooLow) return null;

  return (
    <div className={styles["wrapper"]}>
      <div className={styles.top}>
        <p className={styles.title}>Receive</p>
        <div className={styles.config}>
          <button className={styles.setting} type="button" onClick={() => setShowAdvancedSettingsModal(true)}>
            <SettingIcon />
          </button>
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
