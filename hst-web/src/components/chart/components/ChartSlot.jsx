import useGlobalStore from "../../../store/useGlobalStore";
import useSymbolLive from "../../../hooks/useSymbolLive";
import ChartComponent from "./ChartComponent";

const ChartSlot = ({ chartId }) => {
  const setGlobalStore = useGlobalStore((state) => state.setGlobalStore);
  const chartSymbols = useGlobalStore((state) => state.chartSymbols);
  const setChartSymbol = useGlobalStore((state) => state.setChartSymbol);
  const isDrag = useGlobalStore((state) => state.isDrag);

  const symbolId = chartSymbols[chartId];
  const isLive = useSymbolLive(symbolId);

  const handleDrop = (e) => {
    e.preventDefault();
    const id = e.dataTransfer.getData("id");
    if (id) setChartSymbol(chartId, id);
    setGlobalStore({ isDrag: false });
  };

  if (!symbolId) return null;

  return (
    <div
      onDrop={handleDrop}
      onDragOver={(e) => e.preventDefault()}
      className="relative h-full w-full flex items-center justify-center"
    >
      {isDrag && <div className="absolute h-full w-full top-0 left-0 z-50" />}

      <ChartComponent symbolId={symbolId} isLive={isLive} />
    </div>
  );
};

export default ChartSlot;
