import Modal from "@/components/v2/modal";
import styles from "./estimated-time.module.scss";
import Button from "@/components/v2/ui/button";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function EstimatedTime({ isModalOpen, onCloseModal }: Props) {
  return (
    <Modal title="Estimated time" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <p className={styles["text"]}>Linea has an approximate 20 minute delay on deposits as a security measure.</p>
        <Button fullWidth onClick={onCloseModal}>
          OK
        </Button>
      </div>
    </Modal>
  );
}
