import { useChainStore } from "@/stores/chainStore";
import { useFormContext } from "react-hook-form";
import { Stepper } from "../ui";
import { parseUnits } from "viem";

const STEPS = ["Approve", "Bridge"];

export function ERC20Stepper() {
  const token = useChainStore((state) => state.token);

  const { watch } = useFormContext();

  const watchAmount = watch("amount");
  const watchAllowance = watch("allowance");

  const isETHTransfer = token && token.symbol === "ETH";
  const isApproved =
    !isETHTransfer &&
    watchAmount &&
    watchAmount > 0 &&
    watchAllowance &&
    token?.decimals &&
    watchAllowance >= parseUnits(watchAmount, token?.decimals);

  return <Stepper steps={STEPS} activeStep={isApproved ? 1 : 0} />;
}
