import { useState } from "react";
import EstimatedTimeModal from "../../modal/estimated-time";
import ClockIcon from "@/assets/icons/clock.svg";
import styles from "./estimated-time.module.scss";
import { useChainStore } from "@/stores";
import { ChainLayer } from "@/types";

export default function EstimatedTime() {
  const isL1Network = useChainStore((state) => state.fromChain.layer === ChainLayer.L1);
  const [showEstimatedTimeModal, setShowEstimatedTimeModal] = useState<boolean>(false);

  const estimatedTime = isL1Network ? "~ 20 mins" : "~ 8-32 hours";
  const estimatedTimeType = isL1Network ? "deposit" : "withdraw";

  return (
    <>
      <button type="button" className={styles.time} onClick={() => setShowEstimatedTimeModal(true)}>
        <ClockIcon />
        <span>{estimatedTime}</span>
      </button>
      <EstimatedTimeModal
        type={estimatedTimeType}
        isModalOpen={showEstimatedTimeModal}
        onCloseModal={() => setShowEstimatedTimeModal(false)}
      />
    </>
  );
}
