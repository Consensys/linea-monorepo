import Link from "next/link";
import ReloadHistoryButton from "./ReloadHistoryButton";
import { useFetchHistory } from "@/hooks";

export function NoTransactions() {
  const { clearHistory } = useFetchHistory();
  return (
    <div className="rounded-lg border-2 border-card bg-cardBg p-4">
      <ReloadHistoryButton clearHistory={clearHistory} />
      <div className="flex min-h-80 flex-col items-center justify-center gap-8 ">
        <span className="text-[#C0C0C0]">No bridge transactions found</span>
        <Link href="/" className="btn btn-primary max-w-xs rounded-full uppercase">
          Bridge assets
        </Link>
      </div>
    </div>
  );
}
