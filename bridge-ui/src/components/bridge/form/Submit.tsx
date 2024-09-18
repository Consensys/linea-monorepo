import { NetworkLayer } from "@/config";
import { useBridge } from "@/hooks";
import { useChainStore } from "@/stores/chainStore";
import { useFormContext } from "react-hook-form";
import ApproveERC20 from "./ApproveERC20";
import { Button } from "../../ui";
import { useAccount, useBalance } from "wagmi";
import { cn } from "@/utils/cn";
import { parseUnits } from "viem";

type SubmitProps = {
  isLoading: boolean;
  isWaitingLoading: boolean;
};

export function Submit({ isLoading = false, isWaitingLoading = false }: SubmitProps) {
  // Form
  const { watch, formState } = useFormContext();
  const { errors } = formState;

  const [watchAmount, watchAllowance, watchClaim] = watch(["amount", "allowance", "claim"]);

  // Context
  const { token, networkLayer, toChainId } = useChainStore((state) => ({
    token: state.token,
    networkLayer: state.networkLayer,
    toChainId: state.toChain?.id,
  }));

  // Wagmi
  const { bridgeEnabled } = useBridge();
  const { address } = useAccount();
  const { data: destinationChainBalance } = useBalance({
    address,
    chainId: toChainId,
    query: {
      enabled: !!address && !!toChainId,
    },
  });

  const isERC20Token = token && networkLayer !== NetworkLayer.UNKNOWN && token[networkLayer];
  const isButtonDisabled = !bridgeEnabled(watchAmount, watchAllowance || BigInt(0), errors);
  const isETHTransfer = token && token.symbol === "ETH";
  const showApproveERC20 =
    !isETHTransfer &&
    (!watchAllowance || (token?.decimals && watchAllowance < parseUnits(watchAmount, token.decimals)));

  // TODO: refactor this
  const destinationBalanceTooLow =
    watchClaim === "manual" && destinationChainBalance && destinationChainBalance.value === 0n;

  const buttonText = errors?.amount?.message
    ? "Insufficient balance"
    : destinationBalanceTooLow
      ? "Bridge anyway"
      : "Bridge";

  return isETHTransfer ? (
    <Button
      type="submit"
      className={cn("w-full text-lg font-normal", {
        "btn-secondary": destinationBalanceTooLow,
      })}
      disabled={isButtonDisabled}
      loading={isLoading || isWaitingLoading}
    >
      {buttonText}
    </Button>
  ) : showApproveERC20 && isERC20Token ? (
    <ApproveERC20 />
  ) : (
    <Button
      id="submit-erc-btn"
      className="w-full text-lg font-normal"
      disabled={isButtonDisabled}
      loading={isLoading || isWaitingLoading}
      type="submit"
    >
      {buttonText}
    </Button>
  );
}
