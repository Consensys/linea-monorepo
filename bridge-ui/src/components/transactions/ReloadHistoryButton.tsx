import { Button } from "../ui";

export default function ReloadHistoryButton({ clearHistory }: { clearHistory: () => void }) {
  return (
    <div className="flex justify-end">
      <Button
        id="reload-history-btn"
        variant="link"
        size="sm"
        className="font-light normal-case text-gray-200 no-underline opacity-60 hover:text-primary hover:opacity-100"
        onClick={clearHistory}
      >
        Reload history
      </Button>
    </div>
  );
}
