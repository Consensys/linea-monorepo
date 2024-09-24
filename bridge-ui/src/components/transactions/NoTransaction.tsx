import Link from "next/link";

export function NoTransactions() {
  return (
    <div className="flex min-h-80 flex-col items-center justify-center gap-8 rounded-lg border-2 border-card bg-cardBg p-4">
      <span className="text-[#C0C0C0]">No bridge transactions found</span>
      <Link href="/" className="btn btn-primary max-w-xs rounded-full uppercase">
        Bridge assets
      </Link>
    </div>
  );
}
