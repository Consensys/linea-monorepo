import BridgeTwoLogo from "@/components/v2/bridge/bridge-two-logo";
import styles from "./claiming.module.scss";
import SettingIcon from "@/assets/icons/setting.svg";
import ClockIcon from "@/assets/icons/clock.svg";
import AttentionIcon from "@/assets/icons/attention.svg";
import Image from "next/image";
import { useState } from "react";
import ManualClaim from "@/components/v2/bridge/modal/manual-claim";
import EstimatedTime from "@/components/v2/bridge/modal/estimated-time";
import GasFees from "@/components/v2/bridge/modal/gas-fees";
import AdvancedSettings from "@/components/v2/bridge/modal/advanced-settings";
import BridgeModeDropdown from "@/components/v2/bridge/bridge-mode-dropdown";

export type BridgeModeOption = {
  value: string;
  label: string;
  image: string;
};

const options: BridgeModeOption[] = [
  { value: "native", label: "Native Bridge", image: "/images/logo/linea-rounded.svg" },
  { value: "across", label: "Across", image: "/images/logo/across.svg" },
];

export default function Claiming() {
  const [selectedMode, setSelectedMode] = useState<BridgeModeOption>(options[0]);
  const [showManualClaimModal, setShowManualClaimModal] = useState<boolean>(false);
  const [showEstimatedTimeModal, setShowEstimatedTimeModal] = useState<boolean>(false);
  const [showGasFeesModal, setShowGasFeesModal] = useState<boolean>(false);
  const [showAdvancedSettingsModal, setShowAdvancedSettingsModal] = useState<boolean>(false);

  return (
    <div className={styles["wrapper"]}>
      <div className={styles.top}>
        <p className={styles.title}>Get on Linea</p>
        <div className={styles.config}>
          <BridgeModeDropdown selectedMode={selectedMode} setSelectedMode={setSelectedMode} options={options} />
          <button className={styles.setting} type="button" onClick={() => setShowAdvancedSettingsModal(true)}>
            <SettingIcon />
          </button>
        </div>
      </div>
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
        <button type="button" className={styles["gas-fees"]} onClick={() => setShowGasFeesModal(true)}>
          <Image src="/images/logo/ethereum-rounded.svg" width={12} height={12} alt="eth" />
          <p className={styles["estimate-crypto"]}>0.00019087 ETH</p>
          <p className={styles["estimate-amount"]}>{"($10.54)"}</p>
        </button>
        <button type="button" className={styles.time} onClick={() => setShowEstimatedTimeModal(true)}>
          <ClockIcon />
          <span>~ 20 mins</span>
        </button>
        <button type="button" className={styles.manual} onClick={() => setShowManualClaimModal(true)}>
          <AttentionIcon />
          <span>Manual</span>
        </button>
      </div>
      {selectedMode.value === "native" ? (
        <div className={styles.bottom}>
          Can&apos;t wait?{" "}
          <button type="button" onClick={() => setSelectedMode(options[1])}>
            Speed up with Across
          </button>
        </div>
      ) : null}
      <ManualClaim isModalOpen={showManualClaimModal} onCloseModal={() => setShowManualClaimModal(false)} />
      <EstimatedTime isModalOpen={showEstimatedTimeModal} onCloseModal={() => setShowEstimatedTimeModal(false)} />
      <GasFees isModalOpen={showGasFeesModal} onCloseModal={() => setShowGasFeesModal(false)} />
      <AdvancedSettings
        isModalOpen={showAdvancedSettingsModal}
        onCloseModal={() => setShowAdvancedSettingsModal(false)}
      />
    </div>
  );
}
