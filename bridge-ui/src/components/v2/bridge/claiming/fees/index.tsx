import { useFormContext } from "react-hook-form";
import { BridgeForm } from "@/models";
import styles from "./fees.module.scss";
import { useChainStore } from "@/stores/chainStore";
import ManualClaim from "../manual-claim";
import EstimatedTime from "../estimated-time";
import WithFees from "./with-fees";

export default function Fees() {
  const fromChain = useChainStore.useFromChain();
  const { watch } = useFormContext<BridgeForm>();

  const claim = watch("claim");

  return (
    <>
      <div className={styles.estimate}>
        {<WithFees iconPath={fromChain.iconPath} />}
        <EstimatedTime />
        {claim === "manual" && <ManualClaim />}
      </div>
    </>
  );
}
