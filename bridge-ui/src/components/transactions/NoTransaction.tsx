import Link from "next/link";
import RefreshHistoryButton from "./RefreshHistoryButton";
import { useFetchHistory } from "@/hooks";

export function NoTransactions() {
  const { fetchHistory, isLoading } = useFetchHistory();
  return (
    <div className="rounded-lg border-2 border-card bg-cardBg p-4">
      <RefreshHistoryButton fetchHistory={fetchHistory} isLoading={isLoading} />
      <div className="flex min-h-80 flex-col items-center justify-center gap-8 ">
        <span className="text-[#C0C0C0]">No bridge transactions found</span>
        <Link href="/" className="btn btn-primary max-w-xs rounded-full uppercase">
          Bridge assets
        </Link>
      </div>
    </div>
  );
}
