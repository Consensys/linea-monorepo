import useEthBridgeTxArgs from "./useEthBridgeTxArgs";
import useERC20BridgeTxArgs from "./useERC20BridgeTxArgs";
import useERC20ApproveTxArgs from "./useERC20ApproveTxArgs";
import useAllowance from "../useAllowance";

const useTransactionArgs = () => {
  const ethBridgeTxArgs = useEthBridgeTxArgs();
  const { allowance } = useAllowance();
  const erc20BridgeTxArgs = useERC20BridgeTxArgs({ allowance });
  const erc20ApproveTxArgs = useERC20ApproveTxArgs({ allowance });

  if (erc20ApproveTxArgs) {
    return erc20ApproveTxArgs;
  }

  return ethBridgeTxArgs || erc20BridgeTxArgs;
};

export default useTransactionArgs;
