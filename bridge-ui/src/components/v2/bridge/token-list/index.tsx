import Image from "next/image";
import { useAccount } from "wagmi";
import TokenModal from "@/components/v2/bridge/modal/token-modal";
import { Button } from "@/components/ui";
import CaretDownIcon from "@/assets/icons/caret-down.svg";
import styles from "./token-list.module.scss";
import { useState } from "react";
import { useFormContext } from "react-hook-form";
import { useSelectedToken } from "@/hooks/useSelectedToken";

export default function TokenList() {
  const [isModalOpen, setIsModalOpen] = useState(false);

  const token = useSelectedToken();

  const { isConnected } = useAccount();
  const { setValue, clearErrors } = useFormContext();

  const openModal = () => setIsModalOpen(true);
  const closeModal = () => setIsModalOpen(false);

  return (
    <div className={styles["wrapper"]}>
      {token && (
        <Button className={styles["token-select-btn"]} disabled={!isConnected} onClick={openModal}>
          <Image src={token.image} alt={token.name} width={24} height={24} />
          {token.symbol}
          <CaretDownIcon className={styles["arrow-down-icon"]} />
        </Button>
      )}
      <TokenModal setValue={setValue} clearErrors={clearErrors} isModalOpen={isModalOpen} onCloseModal={closeModal} />
    </div>
  );
}
