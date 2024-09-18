type CollapseProps = {
  title: string;
  children: React.ReactNode;
};

const Collapse: React.FC<CollapseProps> = ({ title, children }) => {
  return (
    <details className="collapse collapse-arrow rounded-none border-2 border-card bg-cardBg hover:border-primary">
      <summary className="collapse-title px-8 text-xl font-medium text-white">{title}</summary>
      <div className="collapse-content flex flex-col gap-2 px-8 text-justify">{children}</div>
    </details>
  );
};

export default Collapse;
