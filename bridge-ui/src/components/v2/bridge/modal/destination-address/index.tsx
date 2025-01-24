import Modal from "@/components/v2/modal";
import styles from "./destination-address.module.scss";
import Button from "@/components/v2/ui/button";
import { useState } from "react";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
  defaultAddress: string;
};

enum FormStatus {
  SUCCESS = "SUCCESS",
  WARNING = "WARNING",
  ERROR = "ERROR",
}

export default function DestinationAddress({ isModalOpen, onCloseModal, defaultAddress }: Props) {
  const [address, setAddress] = useState<string>(defaultAddress);
  const [error, setError] = useState<string>("");
  const [status, setStatus] = useState<FormStatus>(FormStatus.SUCCESS);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onCloseModal();
  };
  return (
    <Modal title="Destination address" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <form onSubmit={handleSubmit}>
          <label htmlFor="address">To address</label>
          <input type="text" id="address" value={address} onChange={(e) => setAddress(e.target.value)} />
          {error && <p className={styles["error-text"]}>{error}</p>}
          <Button type="submit" fullWidth>
            Save
          </Button>
        </form>
      </div>
    </Modal>
  );
}
