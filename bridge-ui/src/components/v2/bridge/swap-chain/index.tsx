import Tooltip from "@/components/v2/ui/tooltip";
import ArrowRightIcon from "@/assets/icons/arrow-right.svg";
import styles from "./swap-chain.module.scss";
import { useChainStore } from "@/stores/chainStore";
import { useFormContext } from "react-hook-form";
import { BridgeForm } from "@/models";
import { useIsLoggedIn } from "@/lib/dynamic";

export default function SwapChain() {
  const switchChainInStore = useChainStore.useSwitchChain();
  const isLoggedIn = useIsLoggedIn();
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
        disabled={!isLoggedIn}
      >
        <ArrowRightIcon className={styles["arrow-icon"]} />
      </button>
    </Tooltip>
  );
}
