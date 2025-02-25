import ConnectButton from "@/components/v2/connect-button";
import FaqHelp from "@/components/v2/bridge/faq-help";
import TokenList from "@/components/v2/bridge/token-list";
import { useAccount } from "wagmi";
import { Amount } from "@/components/v2/bridge/amount";
import SwapChain from "@/components/v2/bridge/swap-chain";
import FromChain from "@/components/v2/bridge/from-chain";
import ToChain from "@/components/v2/bridge/to-chain";
import Claiming from "@/components/v2/bridge/claiming";
import styles from "./bridge-form.module.scss";
import { Submit } from "@/components/v2/bridge/submit";
import TransactionPaperIcon from "@/assets/icons/transaction-paper.svg";
import Setting from "@/components/v2/setting";
import { useEffect, useState } from "react";
import { DestinationAddress } from "../destination-address";
import Button from "../../ui/button";
import { useNativeBridgeNavigationStore } from "@/stores/nativeBridgeNavigationStore";
import { FormProvider, useForm } from "react-hook-form";
import { BridgeForm as BridgeFormModel } from "@/models";
import { useTokenStore } from "@/stores/tokenStoreProvider";
import { BridgeType, TokenType } from "@/config/config";
import { useChainStore } from "@/stores/chainStore";

export default function BridgeForm() {
  const { isConnected } = useAccount();
  const token = useChainStore.useToken();
  const [isDestinationAddressOpen, setIsDestinationAddressOpen] = useState(false);
  const setIsTransactionHistoryOpen = useNativeBridgeNavigationStore.useSetIsTransactionHistoryOpen();
  const setIsBridgeOpen = useNativeBridgeNavigationStore.useSetIsBridgeOpen();
  const configContextValue = useTokenStore((state) => state.tokensList);

  const methods = useForm<BridgeFormModel>({
    defaultValues: {
      token: configContextValue?.UNKNOWN[0],
      claim: token?.type === TokenType.ETH ? "auto" : "manual",
      amount: "",
      minFees: 0n,
      gasFees: 0n,
      bridgingAllowed: false,
      balance: "0",
      mode: BridgeType.NATIVE,
    },
  });

  return (
    <FormProvider {...methods}>
      <form>
        <div className={styles["form-wrapper"]}>
          <div className={styles.headline}>
            <div className={styles["action"]}>
              <Button
                variant="link"
                onClick={() => {
                  setIsBridgeOpen(false);
                  setIsTransactionHistoryOpen(true);
                }}
              >
                <TransactionPaperIcon className={styles["transaction-icon"]} />
              </Button>
              <Setting />
            </div>
          </div>
          <div className={styles["exchange"]}>
            <FromChain />
            <div className={styles["swap-chain-container"]}>
              <SwapChain />
            </div>
            <ToChain />
          </div>
          <div className={styles["amount-wrapper"]}>
            <Amount />
            <div className={styles["right"]}>
              <TokenList />
            </div>
          </div>
          <Claiming />
          {isDestinationAddressOpen && (
            <div className={styles["destination-address-wrapper"]}>
              <DestinationAddress />
            </div>
          )}
          <div className={styles["connect-btn-wrapper"]}>
            {isConnected ? (
              <Submit setIsDestinationAddressOpen={() => setIsDestinationAddressOpen((prev) => !prev)} />
            ) : (
              <ConnectButton fullWidth text={"Connect wallet"} />
            )}
          </div>
          <FaqHelp isMobile />
        </div>
      </form>
      <FaqHelp />
    </FormProvider>
  );
}
