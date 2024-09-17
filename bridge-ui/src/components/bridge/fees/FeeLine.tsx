import { BsDashLg } from "react-icons/bs";
import { MdInfo } from "react-icons/md";
import { Tooltip } from "@/components/tooltip";

interface FeeLineProps {
  label: string;
  value: string | undefined;
  tooltip?: string;
  tooltipClassName?: string;
}

export const FeeLine: React.FC<FeeLineProps> = ({ label, value, tooltip, tooltipClassName }) => (
  <div className="flex justify-between text-[#C0C0C0]">
    <div className="flex items-center gap-2">
      <span>{label}:</span>
      {tooltip && (
        <Tooltip text={tooltip} className={tooltipClassName}>
          <MdInfo />
        </Tooltip>
      )}
    </div>
    <span>{value ? value : <BsDashLg />}</span>
  </div>
);
