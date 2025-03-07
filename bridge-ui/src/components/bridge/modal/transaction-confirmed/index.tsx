import Link from "next/link";
import Modal from "@/components/modal";
import styles from "./transaction-confirmed.module.scss";
import { useNativeBridgeNavigationStore } from "@/stores";
import Button from "@/components/ui/button";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function TransactionConfirmed({ isModalOpen, onCloseModal }: Props) {
  const setIsTransactionHistoryOpen = useNativeBridgeNavigationStore.useSetIsTransactionHistoryOpen();

  return (
    <Modal title="Transaction confirmed!" isOpen={isModalOpen} onClose={onCloseModal} size="lg">
      <div className={styles["modal-inner"]}>
        <p className={styles["text"]}>
          You may now bridge another transaction, check your transaction history, or stay ahead of the curve with the
          latest trending tokens.
        </p>
        <div className={styles["list-button"]}>
          <Link
            className={styles["primary-btn"]}
            href="https://linea.build/ecosystem"
            target="_blank"
            rel="noopenner noreferrer"
          >
            See What&apos;s Trending
          </Link>

          <Button
            variant="link"
            className={styles["secondary-btn"]}
            onClick={() => {
              setIsTransactionHistoryOpen(true);
              onCloseModal();
            }}
          >
            View transactions
          </Button>
        </div>
      </div>
    </Modal>
  );
}
