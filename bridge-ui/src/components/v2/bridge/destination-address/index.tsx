import { useEffect } from "react";
import { useAccount } from "wagmi";
import clsx from "clsx";
import { isAddress } from "viem";
import { useFormContext } from "react-hook-form";
import styles from "./destination-address.module.scss";
import XCircleIcon from "@/assets/icons/x-circle.svg";

export function DestinationAddress() {
  const { address } = useAccount();
  const { register, formState, setError, clearErrors, watch, setValue } = useFormContext();
  const { errors } = formState;

  const watchDestinationAddress = watch("destinationAddress");

  useEffect(() => {
    if (address) {
      setValue("destinationAddress", address);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [address]);

  useEffect(() => {
    if (watchDestinationAddress && !isAddress(watchDestinationAddress)) {
      setError("destinationAddress", {
        type: "custom",
        message: "Invalid address",
      });
    } else {
      clearErrors("destinationAddress");
    }
  }, [watchDestinationAddress, setError, clearErrors]);

  const handleResetInput = () => {
    setValue("destinationAddress", address);
  };

  return (
    <div className={styles["destination-address"]}>
      <p className={styles.title}>Send to wallet</p>
      <div className={styles["input-container"]}>
        <input
          type="text"
          id="address"
          autoCorrect="off"
          autoComplete="off"
          spellCheck="false"
          maxLength={42}
          {...register("destinationAddress", {
            validate: (value) => !value || isAddress(value) || "Invalid address",
          })}
        />

        <button
          type="button"
          className={clsx(styles.reset, {
            [styles["show"]]: watchDestinationAddress && watchDestinationAddress !== address,
          })}
          onClick={handleResetInput}
        >
          <XCircleIcon />
        </button>
      </div>
      {errors.destinationAddress && (
        <p className={styles["message-text"]}>{errors.destinationAddress.message?.toString()}</p>
      )}
    </div>
  );
}
