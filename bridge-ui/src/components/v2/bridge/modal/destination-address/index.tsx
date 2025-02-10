import Modal from "@/components/v2/modal";
import styles from "./destination-address.module.scss";
import Button from "@/components/v2/ui/button";
import { useEffect, useState } from "react";
import clsx from "clsx";
import XCircleIcon from "@/assets/icons/x-circle.svg";
import { useFormContext } from "react-hook-form";
import { isAddress } from "viem";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
  defaultAddress: string;
};

const type = "error";

export default function DestinationAddress({ isModalOpen, onCloseModal, defaultAddress }: Props) {
  const { register, formState, setValue, setError, clearErrors, watch } = useFormContext();
  const { errors } = formState;

  const watchRecipient = watch("recipient");

  useEffect(() => {
    if (watchRecipient && !isAddress(watchRecipient)) {
      setError("recipient", {
        type: "custom",
        message: "Invalid address",
      });
    } else {
      clearErrors("recipient");
    }
  }, [watchRecipient, setError, clearErrors]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCloseModal();
  };
  const handleResetInput = () => {
    setValue("recipient", defaultAddress);
  };

  return (
    <Modal title="Destination address" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={clsx(styles["modal-inner"], styles[type])}>
        <form onSubmit={handleSubmit}>
          <label htmlFor="address">To address</label>
          <div className={styles["input-container"]}>
            <input
              type="text"
              id="address"
              autoCorrect="off"
              autoComplete="off"
              spellCheck="false"
              maxLength={42}
              {...register("recipient", {
                validate: (value) => !value || isAddress(value) || "Invalid address",
              })}
            />

            <button
              type="button"
              className={clsx(styles.reset, {
                [styles["show"]]: watchRecipient && watchRecipient !== defaultAddress,
              })}
              onClick={handleResetInput}
            >
              <XCircleIcon />
            </button>
          </div>
          {errors.recipient && <p className={styles["message-text"]}>{errors.recipient.message?.toString()}</p>}

          <Button className={styles["btn-save"]} type="submit" fullWidth>
            Save
          </Button>
        </form>
      </div>
    </Modal>
  );
}
