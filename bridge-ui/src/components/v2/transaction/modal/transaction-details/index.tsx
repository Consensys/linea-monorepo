import Modal from "@/components/v2/modal";
import styles from "./transaction-details.module.scss";
import Button from "@/components/v2/ui/button";
import { Transaction } from "@/components/v2/transaction/list-transaction";
import { formatAddress } from "@/utils/format";
import ArrowRightIcon from "@/assets/icons/arrow-right.svg";
import Link from "next/link";

type Props = {
  transaction: Transaction | undefined;
  isModalOpen: boolean;
  onCloseModal: () => void;
};

export default function TransactionDetails({ transaction, isModalOpen, onCloseModal }: Props) {
  const handleClaim = () => {
    onCloseModal();
  };
  return (
    <Modal title="Transaction details" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <ul className={styles.list}>
          <li>
            <span>Timestamp</span>
            <div className={styles["date-time"]}>
              <span>Dec 18, 2024</span>
              <span>3:16 PM UTC</span>
            </div>
          </li>
          <li>
            <span>Ethereum Tx hash</span>
            <div className={styles.hash}>
              <Link href="/" target="_blank" rel="noopenner noreferrer">
                {formatAddress(transaction?.code)}
              </Link>
              <ArrowRightIcon />
            </div>
          </li>
          <li>
            <span>Linea Tx hash</span>
            <div className={styles.hash}>
              <Link href="/" target="_blank" rel="noopenner noreferrer">
                {formatAddress(transaction?.code)}
              </Link>
              <ArrowRightIcon />
            </div>
          </li>
          <li>
            <span>Gas fee</span>
            <span className={styles.price}>1.1 USD</span>
          </li>
        </ul>
        <Button onClick={handleClaim} fullWidth>
          CLAIM
        </Button>
      </div>
    </Modal>
  );
}
