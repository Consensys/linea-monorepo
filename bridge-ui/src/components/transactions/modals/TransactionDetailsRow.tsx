type TransactionDetailRowProps = {
  label: string;
  value: React.ReactNode | string;
};

const TransactionDetailRow: React.FC<TransactionDetailRowProps> = ({ label, value }) => (
  <div className="flex items-center">
    <label className="w-44 text-[#C0C0C0]">{label}</label>
    <span className="text-[#E5E5E5]">{value}</span>
  </div>
);

export default TransactionDetailRow;
