import { OnChainMessageStatus } from "@consensys/linea-sdk";
import { useMemo } from "react";

type StatusTextProps = {
  status: OnChainMessageStatus;
};

const useFormattedStatus = (status: OnChainMessageStatus): JSX.Element => {
  return useMemo(() => {
    switch (status) {
      case OnChainMessageStatus.CLAIMABLE:
        return <span className="text-secondary">Ready to claim</span>;
      case OnChainMessageStatus.UNKNOWN:
        return <span className="text-primary">Pending</span>;
      case OnChainMessageStatus.CLAIMED:
        return <span className="text-success">Completed</span>;
      default:
        throw new Error(`Incorrect transaction status: ${status}`);
    }
  }, [status]);
};

export default function StatusText({ status }: StatusTextProps) {
  return useFormattedStatus(status);
}
