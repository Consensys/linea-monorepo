type CollapseProps = {
  title: string;
  children: React.ReactNode;
};

export const Collapse: React.FC<CollapseProps> = ({ title, children }) => {
  return (
    <details className="collapse collapse-arrow rounded-none border-2 border-card bg-cardBg hover:border-primary">
      <summary className="collapse-title text-xl font-medium text-white">{title}</summary>
      <div className="collapse-content">{children}</div>
    </details>
  );
};
