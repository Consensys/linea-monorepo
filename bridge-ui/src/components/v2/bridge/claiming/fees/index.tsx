import { useState } from "react";
import Image from "next/image";
import { useFormContext } from "react-hook-form";
import { BridgeForm } from "@/models";
import { BridgeType } from "@/config/config";
import styles from "./fees.module.scss";
import { useChainStore } from "@/stores/chainStore";
import GasFees from "../../modal/gas-fees";
import AcrossFees from "../../modal/across-fees";
import ManualClaim from "../manual-claim";
import NoFees from "./no-fees";
import EstimatedTime from "../estimated-time";

export default function Fees() {
  const fromChain = useChainStore.useFromChain();
  const { watch } = useFormContext<BridgeForm>();
  const [showGasFeesModal, setShowGasFeesModal] = useState<boolean>(false);
  const [showAcrossFeesModal, setShowAcrossFeesModal] = useState<boolean>(false);

  const [mode, claim] = watch(["mode", "claim"]);

  const isNoFees = claim === "manual" && mode === BridgeType.NATIVE;

  const handleShowFees = () => {
    if (mode === BridgeType.NATIVE) {
      setShowGasFeesModal(true);
    } else if (mode === BridgeType.ACROSS) {
      setShowAcrossFeesModal(true);
    }
  };

  return (
    <>
      <div className={styles.estimate}>
        {isNoFees ? (
          <NoFees iconPath={fromChain.iconPath} />
        ) : (
          <button type="button" className={styles["gas-fees"]} onClick={handleShowFees}>
            <Image src={fromChain?.iconPath ?? ""} width={12} height={12} alt="eth" />
            <p className={styles["estimate-crypto"]}>0.00019087 ETH</p>
            <p className={styles["estimate-amount"]}>{"($10.54)"}</p>
          </button>
        )}
        <EstimatedTime />
        {claim === "manual" && <ManualClaim />}
      </div>
      <GasFees isModalOpen={showGasFeesModal} onCloseModal={() => setShowGasFeesModal(false)} />
      <AcrossFees isModalOpen={showAcrossFeesModal} onCloseModal={() => setShowAcrossFeesModal(false)} />
    </>
  );
}
