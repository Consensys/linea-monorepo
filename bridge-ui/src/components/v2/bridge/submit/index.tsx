import { useFormContext } from "react-hook-form";
import { parseUnits } from "viem";
// import { useBridge } from "@/hooks";
import Button from "@/components/v2/ui/button";
import WalletIcon from "@/assets/icons/wallet.svg";
import DestinationAddress from "@/components/v2/bridge/modal/destination-address";
import { MouseEventHandler, useState } from "react";
import styles from "./submit.module.scss";

type Props = {
  setIsDestinationAddressOpen: MouseEventHandler<HTMLButtonElement>;
};

export function Submit({ setIsDestinationAddressOpen }: Props) {
  const [showChangeAddressModal, setShowChangeAddressModal] = useState<boolean>(false);

  // Form
  const { watch, formState } = useFormContext();
  const { errors } = formState;

  const [watchAmount, balance, token] = watch(["amount", "balance", "token"]);

  // const { bridgeEnabled } = useBridge();

  const originChainBalanceTooLow =
    errors?.amount?.message !== undefined ||
    parseUnits(balance, token.decimals) < parseUnits(watchAmount, token.decimals);

  const disabled = originChainBalanceTooLow || !watchAmount;

  const buttonText = !watchAmount
    ? "Enter an amount"
    : originChainBalanceTooLow
      ? "Insufficient funds"
      : "Review Bridge";

  const handleCloseModal = () => {
    setShowChangeAddressModal(false);
  };

  return (
    <div className={styles.container}>
      <Button disabled={disabled} fullWidth>
        {buttonText}
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
