import { useFormContext } from "react-hook-form";
import { parseUnits } from "viem";
import { useAccount, useBalance } from "wagmi";
import { NetworkLayer } from "@/config";
import { useBridge } from "@/hooks";
import { useChainStore } from "@/stores/chainStore";
import Button from "@/components/v2/ui/button";

type Props = {
  disabled?: boolean;
};

export function Submit({ disabled = false }: Props) {
  // Form
  const { watch, formState } = useFormContext();
  const { errors } = formState;

  const [watchAmount, watchAllowance, watchClaim, watchBalance] = watch(["amount", "allowance", "claim", "balance"]);

  // Context
  const token = useChainStore((state) => state.token);
  const networkLayer = useChainStore((state) => state.networkLayer);
  const toChainId = useChainStore((state) => state.toChain?.id);

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

  const originChainBalanceTooLow =
    token !== null &&
    (errors?.amount?.message !== undefined ||
      parseUnits(watchBalance, token.decimals) < parseUnits(watchAmount, token.decimals));

  const destinationBalanceTooLow =
    watchClaim === "manual" && destinationChainBalance && destinationChainBalance.value === 0n;

  const buttonText = originChainBalanceTooLow
    ? "Insufficient funds"
    : destinationBalanceTooLow
      ? "Bridge anyway"
      : "Bridge";

  return (
    <Button disabled={disabled} fullWidth>
      Bridge
    </Button>
  );
}
