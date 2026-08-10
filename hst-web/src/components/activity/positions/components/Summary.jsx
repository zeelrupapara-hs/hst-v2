import usePositionStore from "../../../../store/usePositionStore";
import { formatMoney } from "../../../../utils/utils";

const Summary = () => {
  const summary = usePositionStore((state) => state.summary);
  const digits = usePositionStore((state) => state.currencyDigits);

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
          {/* the margin level is a percentage, not money, so the group's digits do not apply */}
          <span>
            {key === "margin_level"
              ? summary[key] || 0
              : formatMoney(summary[key], digits)}
          </span>
        </div>
      ))}
    </div>
  );
};

export default Summary;
