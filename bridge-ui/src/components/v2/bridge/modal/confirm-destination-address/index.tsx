import Modal from "@/components/v2/modal";
import styles from "./confirm-destination-address.module.scss";
import Button from "@/components/v2/ui/button";
import { formatAddress } from "@/utils/format";
import Link from "next/link";
import UnionIcon from "@/assets/icons/union.svg";

type Props = {
  isModalOpen: boolean;
  onCloseModal: () => void;
  onConfirm: () => void;
};

const address = "0xD54e939a03E516CeE353Dc1f6B7517A5A360C984";

export default function ConfirmDestinationAddress({ isModalOpen, onCloseModal, onConfirm }: Props) {
  return (
    <Modal title="Confirm destination address" isOpen={isModalOpen} onClose={onCloseModal}>
      <div className={styles["modal-inner"]}>
        <p className={styles["text"]}>
          Your funds are being bridged to the following address on the destination chain. Please review and confirm
          before proceeding.
        </p>
        <Link href="/" target="_blank" rel="noopenner noreferrer">
          {formatAddress(address)}
          <UnionIcon />
        </Link>
        <Button fullWidth onClick={onConfirm}>
          Confirm and bridge
        </Button>
      </div>
    </Modal>
  );
}
