import Tooltip from "@/components/v2/ui/tooltip";
import ArrowRightIcon from "@/assets/icons/arrow-right.svg";
import styles from "./swap-chain.module.scss";
import { useChainStore } from "@/stores/chainStore";
import { useFormContext } from "react-hook-form";
import { BridgeForm } from "@/models";

export default function SwapChain() {
  const switchChainInStore = useChainStore((state) => state.switchChain);
  const { reset } = useFormContext<BridgeForm>();

  return (
    <Tooltip text="Switch chains" position="top">
      <button
        className={styles["swap-chain"]}
        onClick={(e) => {
          e.preventDefault();
          e.currentTarget.classList.toggle(styles["rotate-360"]);
          switchChainInStore();
          reset();
        }}
      >
        <ArrowRightIcon className={styles["arrow-icon"]} />
      </button>
    </Tooltip>
  );
}
