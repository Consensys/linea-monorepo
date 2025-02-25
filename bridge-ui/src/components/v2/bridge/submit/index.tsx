import { useFormContext } from "react-hook-form";
import { parseUnits } from "viem";
import { useAccount, useBalance } from "wagmi";
import { NetworkLayer } from "@/config";
import { useBridge } from "@/hooks";
import { useChainStore } from "@/stores/chainStore";
import Button from "@/components/v2/ui/button";
import WalletIcon from "@/assets/icons/wallet.svg";
import DestinationAddress from "@/components/v2/bridge/modal/destination-address";
import { MouseEventHandler, useState } from "react";
import styles from "./submit.module.scss";

type Props = {
  disabled?: boolean;
  setIsDestinationAddressOpen: MouseEventHandler<HTMLButtonElement>;
};

export function Submit({ disabled = false, setIsDestinationAddressOpen }: Props) {
  const [showChangeAddressModal, setShowChangeAddressModal] = useState<boolean>(false);

  // Form
  const { watch, formState } = useFormContext();
  const { errors } = formState;

  const [watchAmount, watchAllowance, watchClaim, watchBalance] = watch(["amount", "allowance", "claim", "balance"]);

  // Context
  const token = useChainStore.useToken();
  const toChainId = useChainStore.useToChain()?.id;
  const networkLayer = useChainStore.useNetworkLayer();

  // Wagmi
  const { bridgeEnabled } = useBridge();
  const { address } = useAccount();
  const { data: destinationChainBalance } = useBalance({
    address,
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    chainId: toChainId,
    query: {
      enabled: !!address && !!toChainId,
    },
  });

  const originChainBalanceTooLow =
    token !== null &&
    (errors?.amount?.message !== undefined ||
      parseUnits(watchBalance, token.decimals) < parseUnits(watchAmount, token.decimals));

  const destinationBalanceTooLow =
    watchClaim === "manual" && destinationChainBalance && destinationChainBalance.value === 0n;

  const buttonText = originChainBalanceTooLow
    ? "Insufficient funds"
    : destinationBalanceTooLow
      ? "Bridge anyway"
      : "Bridge";

  const handleCloseModal = () => {
    setShowChangeAddressModal(false);
  };

  return (
    <div className={styles.container}>
      <Button disabled={disabled} fullWidth>
        Review Bridge
      </Button>
      <button type="button" className={styles["wallet-icon"]} onClick={setIsDestinationAddressOpen}>
        <WalletIcon />
      </button>
      <DestinationAddress
        isModalOpen={showChangeAddressModal}
        onCloseModal={handleCloseModal}
        defaultAddress="0xE9493bF17dyhxzkD23dE93F17hdyh73"
      />
    </div>
  );
}
