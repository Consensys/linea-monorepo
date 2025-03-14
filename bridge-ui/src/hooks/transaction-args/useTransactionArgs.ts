import useEthBridgeTxArgs from "./useEthBridgeTxArgs";
import useERC20BridgeTxArgs from "./useERC20BridgeTxArgs";
import useApproveTxArgs from "./useApproveTxArgs";
import useAllowance from "../useAllowance";
import useDepositForBurnTxArgs from "./cctp/useDepositForBurnTxArgs";

const useTransactionArgs = () => {
  const ethBridgeTxArgs = useEthBridgeTxArgs();
  const { allowance } = useAllowance();
  const erc20BridgeTxArgs = useERC20BridgeTxArgs({ allowance });
  const erc20ApproveTxArgs = useApproveTxArgs({ allowance });
  const cctpDepositForBurnTxArgs = useDepositForBurnTxArgs({ allowance });

  if (erc20ApproveTxArgs) {
    return erc20ApproveTxArgs;
  }

  return ethBridgeTxArgs || erc20BridgeTxArgs || cctpDepositForBurnTxArgs;
};

export default useTransactionArgs;
