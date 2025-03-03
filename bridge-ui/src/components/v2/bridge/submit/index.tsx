import Button from "@/components/v2/ui/button";
import WalletIcon from "@/assets/icons/wallet.svg";
import { MouseEventHandler } from "react";
import styles from "./submit.module.scss";
import { useFormStore } from "@/stores/formStoreProvider";

type Props = {
  setIsDestinationAddressOpen: MouseEventHandler<HTMLButtonElement>;
};

export function Submit({ setIsDestinationAddressOpen }: Props) {
  const amount = useFormStore((state) => state.amount);
  const balance = useFormStore((state) => state.balance);

  const originChainBalanceTooLow = amount && balance < amount;

  const disabled = originChainBalanceTooLow || !amount || amount <= 0n;

  const buttonText =
    !amount || amount <= 0n ? "Enter an amount" : originChainBalanceTooLow ? "Insufficient funds" : "Review Bridge";

  return (
    <div className={styles.container}>
      <Button disabled={disabled} fullWidth>
        {buttonText}
      </Button>
      <button type="button" className={styles["wallet-icon"]} onClick={setIsDestinationAddressOpen}>
        <WalletIcon />
      </button>
    </div>
  );
}
