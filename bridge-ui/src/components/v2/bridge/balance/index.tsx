import { useQueryClient } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { useBlockNumber } from "wagmi";
import { formatBalance } from "@/utils/format";
import { useFormContext } from "react-hook-form";
import { useChainStore } from "@/stores/chainStore";
import { useTokenBalance } from "@/hooks/useTokenBalance";
import WalletIcon from "@/assets/icons/wallet.svg";
import styles from "./balance.module.scss";
import DestinationAddress from "../modal/destination-address";

export function Balance() {
  const [showChangeAddressModal, setShowChangeAddressModal] = useState<boolean>(false);
  // Context
  const { token, networkLayer } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
  }));

  const tokenAddress = token?.[networkLayer];
  // Wagmi
  const queryClient = useQueryClient();
  const { data: blockNumber } = useBlockNumber({ watch: true });
  const { balance, queryKey } = useTokenBalance(tokenAddress, token?.decimals);

  // Form
  const { setValue } = useFormContext();

  useEffect(() => {
    setValue("balance", balance);
  }, [balance, setValue, token?.decimals]);

  useEffect(() => {
    if (blockNumber && blockNumber % 5n === 0n) {
      queryClient.invalidateQueries({ queryKey });
    }
  }, [blockNumber, queryClient, queryKey]);

  const handleCloseModal = () => {
    setShowChangeAddressModal(false);
  };

  return (
    <div className={styles.balance}>
      <span>{balance && `${formatBalance(balance)} ${token?.symbol}`} available</span>
      <WalletIcon className={styles["wallet-icon"]} onClick={() => setShowChangeAddressModal(true)} />
      <DestinationAddress
        isModalOpen={showChangeAddressModal}
        onCloseModal={handleCloseModal}
        defaultAddress="0xE9493bF17dyhxzkD23dE93F17hdyh73"
      />
    </div>
  );
}
