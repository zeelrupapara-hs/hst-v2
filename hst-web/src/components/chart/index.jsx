import { PanelGroup, Panel, PanelResizeHandle } from "react-resizable-panels";
import useGlobalStore from "../../store/useGlobalStore";
import ChartSlot from "./components/ChartSlot";
import OneClick from "../oneClick";

const Chart = () => {
  const symbolId = useGlobalStore((state) => state.symbol);
  const chartLayout = useGlobalStore((state) => state.chartLayout);
  const oneClick = useGlobalStore((state) => state.oneClick);

  const renderCharts = () => {
    switch (chartLayout) {
      case 1:
        return (
          <PanelGroup direction="horizontal">
            <Panel id="chart-1" order={1} defaultSize={100}>
              <ChartSlot chartId={1} />
            </Panel>
          </PanelGroup>
        );

      case 2:
        return (
          <PanelGroup direction="horizontal">
            <Panel id="chart-1" order={1} defaultSize={50}>
              <ChartSlot chartId={1} />
            </Panel>
            <PanelResizeHandle className="w-[2px] bg-theme-border" />
            <Panel id="chart-2" order={2} defaultSize={50}>
              <ChartSlot chartId={2} />
            </Panel>
          </PanelGroup>
        );

      case 3:
        return (
          <PanelGroup direction="vertical">
            <Panel id="chart-1" order={1} defaultSize={50}>
              <ChartSlot chartId={1} />
            </Panel>
            <PanelResizeHandle className="h-[2px] bg-theme-border" />
            <Panel id="chart-2" order={2} defaultSize={50}>
              <ChartSlot chartId={2} />
            </Panel>
          </PanelGroup>
        );

      case 4:
        return (
          <PanelGroup direction="vertical">
            <Panel id="row-1" order={1} defaultSize={50}>
              <PanelGroup direction="horizontal">
                <Panel id="chart-1" order={1} defaultSize={50}>
                  <ChartSlot chartId={1} />
                </Panel>
                <PanelResizeHandle className="w-[2px] bg-theme-border" />
                <Panel id="chart-2" order={2} defaultSize={50}>
                  <ChartSlot chartId={2} />
                </Panel>
              </PanelGroup>
            </Panel>
            <PanelResizeHandle className="h-[2px] bg-theme-border" />
            <Panel id="row-2" order={2} defaultSize={50}>
              <PanelGroup direction="horizontal">
                <Panel id="chart-3" order={3} defaultSize={50}>
                  <ChartSlot chartId={3} />
                </Panel>
                <PanelResizeHandle className="w-[2px] bg-theme-border" />
                <Panel id="chart-4" order={4} defaultSize={50}>
                  <ChartSlot chartId={4} />
                </Panel>
              </PanelGroup>
            </Panel>
          </PanelGroup>
        );

      default:
        return null;
    }
  };

  return (
    <div className="relative h-full w-full">
      {renderCharts()}

      {oneClick ? (
        <div className="absolute top-[70px] left-1/2 -translate-x-1/2 p-1.5">
          <OneClick symbolId={symbolId} />
        </div>
      ) : null}
    </div>
  );
};

export default Chart;
