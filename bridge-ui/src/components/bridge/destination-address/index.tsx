import { ChangeEvent, useEffect, useState } from "react";
import { useAccount } from "wagmi";
import { isAddress } from "viem";
import clsx from "clsx";
import styles from "./destination-address.module.scss";
import XCircleIcon from "@/assets/icons/x-circle.svg";
import { useFormStore } from "@/stores";

export function DestinationAddress() {
  const { address } = useAccount();

  const recipient = useFormStore((state) => state.recipient);
  const setRecipient = useFormStore((state) => state.setRecipient);
  const [inputValue, setInputValue] = useState(recipient);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (inputValue && !isAddress(inputValue)) {
      setError("Invalid address");
    } else {
      setError(null);
    }
  }, [inputValue]);

  const handleChange = (e: ChangeEvent<HTMLInputElement>) => {
    setInputValue(() => e.target.value as `0x${string}`);
    if (isAddress(e.target.value)) {
      setRecipient(e.target.value);
    }
  };

  const handleResetInput = () => {
    if (address) {
      setInputValue(address);
      setRecipient(address);
    }
  };

  return (
    <div className={styles["destination-address"]}>
      <p className={styles.title}>Send to wallet</p>
      <div className={styles["input-container"]}>
        <input
          type="text"
          id="address"
          required
          maxLength={42}
          value={inputValue}
          pattern="^0x[a-fA-F0-9]{40}$"
          onChange={handleChange}
        />

        <button
          type="button"
          className={clsx(styles.reset, {
            [styles["show"]]: inputValue !== address,
          })}
          onClick={handleResetInput}
        >
          <XCircleIcon />
        </button>
      </div>
      {error && <p className={styles["message-text"]}>{error.toString()}</p>}
    </div>
  );
}
