import Modal from "@/components/v2/modal";
import styles from "./gas-fees.module.scss";
import Button from "@/components/v2/ui/button";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
  onClickOk: () => void;
};

export default function GasFees({ isModalOpen, onCloseModal, onClickOk }: Props) {
  const allFees = [
    {
      name: "Ethereum fee",
      fee1: "0.00007875 ETH",
      fee2: "$0.303",
    },
    {
      name: "Linea fee",
      fee1: "0.00007875 ETH",
      fee2: "$0.303",
    },
  ];
  return (
    <Modal title="Gas fees" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <p className={styles["text"]}>
          With blockchains you need to pay a fee to submit transactions. Transactions submitted to the network require a
          small amount of gas to ensure they&apos;re confirmed by the network.
        </p>
        <ul className={styles.list}>
          {allFees.map((row, index) => (
            <li key={index}>
              <span>{row.name}</span>
              <div className={styles["two-fee"]}>
                <span className={styles["fee1"]}>{row.fee1}</span>
                <span className={styles["fee2"]}>{row.fee2}</span>
              </div>
            </li>
          ))}
        </ul>
        <Button fullWidth onClick={onClickOk}>
          OK
        </Button>
      </div>
    </Modal>
  );
}
