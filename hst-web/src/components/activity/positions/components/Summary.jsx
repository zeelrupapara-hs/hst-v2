import usePositionStore from "../../../../store/usePositionStore";

const Summary = () => {
  const summary = usePositionStore((state) => state.summary);

  const data = [
    { key: "balance", label: "Balance" },
    { key: "credit", label: "Credit" },
    { key: "equity", label: "Equity" },
    { key: "used_margin", label: "Used Margin" },
    { key: "free_margin", label: "Free Margin" },
    { key: "margin_level", label: "Margin Level" },
  ];

  return (
    <div className="flex gap-5 text-xs">
      {data?.map(({ key, label }) => (
        <div key={key} className="flex flex-wrap items-center justify-between gap-1">
          <span>{label}:</span>
          <span>{summary[key] || 0}</span>
        </div>
      ))}
    </div>
  );
};

export default Summary;
