import Modal from "@/components/v2/modal";
import styles from "./destination-address.module.scss";
import Button from "@/components/v2/ui/button";
import { useState } from "react";
import clsx from "clsx";
import XCircleIcon from "@/assets/icons/x-circle.svg";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
  defaultAddress: string;
};

export default function DestinationAddress({ isModalOpen, onCloseModal, defaultAddress }: Props) {
  const [address, setAddress] = useState<string>(defaultAddress);
  const message = "This is your connected address";
  const type = "success";

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCloseModal();
  };
  const handleResetInput = () => {
    setAddress(defaultAddress || "");
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
              value={address}
              onChange={(e) => setAddress(e.target.value)}
            />

            <button
              type="button"
              className={clsx(styles.reset, {
                [styles["show"]]: address && address !== defaultAddress,
              })}
              onClick={handleResetInput}
            >
              <XCircleIcon />
            </button>
          </div>
          {message && <p className={styles["message-text"]}>{message}</p>}
          <Button className={styles["btn-save"]} type="submit" fullWidth>
            Save
          </Button>
        </form>
      </div>
    </Modal>
  );
}
