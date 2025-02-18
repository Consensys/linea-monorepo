import { BridgeForm } from "@/models";
import BridgeTwoLogo from "@/components/v2/bridge/bridge-two-logo";
import styles from "./claiming.module.scss";
import SettingIcon from "@/assets/icons/setting.svg";
import ClockIcon from "@/assets/icons/clock.svg";
import AttentionIcon from "@/assets/icons/attention.svg";
import Image from "next/image";
import { useEffect, useState } from "react";
import { useFormContext } from "react-hook-form";
import ManualClaim from "@/components/v2/bridge/modal/manual-claim";
import EstimatedTime from "@/components/v2/bridge/modal/estimated-time";
import GasFees from "@/components/v2/bridge/modal/gas-fees";
import AcrossFees from "@/components/v2/bridge/modal/across-fees";
import AdvancedSettings from "@/components/v2/bridge/modal/advanced-settings";
import Skeleton from "@/components/v2/bridge/claiming/skeleton";
import { BridgeType } from "@/config/config";

export type BridgeModeOption = {
  value: BridgeType;
  label: string;
  image: string;
};

export default function Claiming() {
  const { watch, setValue } = useFormContext<BridgeForm>();
  const mode = watch("mode");

  const [loading, setLoading] = useState<boolean>(false);
  const [showManualClaimModal, setShowManualClaimModal] = useState<boolean>(false);
  const [showEstimatedTimeModal, setShowEstimatedTimeModal] = useState<boolean>(false);
  const [showGasFeesModal, setShowGasFeesModal] = useState<boolean>(false);
  const [showAcrossFeesModal, setShowAcrossFeesModal] = useState<boolean>(false);
  const [showAdvancedSettingsModal, setShowAdvancedSettingsModal] = useState<boolean>(false);

  const [amount, claim] = watch(["amount", "claim"]);

  const isNoFees = claim === "manual" && mode === BridgeType.NATIVE;

  const handleShowFees = () => {
    if (mode === BridgeType.NATIVE) {
      setShowGasFeesModal(true);
    } else if (mode === BridgeType.ACROSS) {
      setShowAcrossFeesModal(true);
    }
  };

  useEffect(() => {
    setLoading(true);
    const timeout = setTimeout(() => {
      setLoading(false);
    }, 1000);

    return () => clearTimeout(timeout);
  }, [amount]);

  if (!amount) return null;

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
              src1="/images/logo/ethereum-rounded.svg"
              src2="/images/logo/linea-rounded.svg"
              alt1="eth"
              alt2="linea"
            />
            <div className={styles.value}>
              <p className={styles.crypto}>1 ETH</p>
              <p className={styles.amount}>$3,855.45</p>
            </div>
          </div>
          <div className={styles.estimate}>
            {isNoFees ? (
              <button type="button" className={styles["no-fees"]}>
                <Image src="/images/logo/ethereum-rounded.svg" width={12} height={12} alt="eth" />
                <p className={styles["text"]}>No Fees</p>
              </button>
            ) : (
              <button type="button" className={styles["gas-fees"]} onClick={handleShowFees}>
                <Image src="/images/logo/ethereum-rounded.svg" width={12} height={12} alt="eth" />
                <p className={styles["estimate-crypto"]}>0.00019087 ETH</p>
                <p className={styles["estimate-amount"]}>{"($10.54)"}</p>
              </button>
            )}
            <button type="button" className={styles.time} onClick={() => setShowEstimatedTimeModal(true)}>
              <ClockIcon />
              <span>~ 20 mins</span>
            </button>
            {claim === "manual" && (
              <button type="button" className={styles.manual} onClick={() => setShowManualClaimModal(true)}>
                <AttentionIcon />
                <span>Manual</span>
              </button>
            )}
          </div>
          {mode === BridgeType.NATIVE ? (
            <div className={styles.bottom}>
              Can&apos;t wait?{" "}
              <button type="button" onClick={() => setValue("mode", BridgeType.ACROSS)}>
                Speed up
              </button>
            </div>
          ) : null}
        </div>
      )}
      <ManualClaim isModalOpen={showManualClaimModal} onCloseModal={() => setShowManualClaimModal(false)} />
      <EstimatedTime isModalOpen={showEstimatedTimeModal} onCloseModal={() => setShowEstimatedTimeModal(false)} />
      <GasFees isModalOpen={showGasFeesModal} onCloseModal={() => setShowGasFeesModal(false)} />
      <AcrossFees isModalOpen={showAcrossFeesModal} onCloseModal={() => setShowAcrossFeesModal(false)} />
      <AdvancedSettings
        isModalOpen={showAdvancedSettingsModal}
        onCloseModal={() => setShowAdvancedSettingsModal(false)}
      />
    </div>
  );
}
