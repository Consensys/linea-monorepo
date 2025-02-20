import Modal from "@/components/v2/modal";
import CheckShieldIcon from "@/assets/icons/check-shield.svg";

import styles from "./advanced-settings.module.scss";
import ToggleSwitch from "@/components/v2/ui/toggle-switch";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function AdvancedSettings({ isModalOpen, onCloseModal }: Props) {
  return (
    <Modal title="Advanced settings" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <div className={styles["container"]}>
          <div className={styles.content}>
            <CheckShieldIcon className={styles.icon} />
            <div>
              <p className={styles["title"]}>Manual claim on destination</p>
              <p className={styles["text"]}>
                You will need to claim your transaction on the destination chain with an additional transaction that
                requires ETH on the destination chain.
              </p>
            </div>
          </div>
          <div className={styles.toggle}>
            <ToggleSwitch disabled checked />
          </div>
        </div>
      </div>
    </Modal>
  );
}
