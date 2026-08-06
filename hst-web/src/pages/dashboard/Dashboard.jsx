import { PanelGroup, Panel, PanelResizeHandle } from "react-resizable-panels";
import useGlobalStore from "../../store/useGlobalStore";
import MarketWatch from "../../components/marketWatch";
import Chart from "../../components/chart";
import Order from "../../components/order";
import Activity from "../../components/activity";

const Dashboard = () => {
  const panels = useGlobalStore((state) => state.panels);
  const symbol = useGlobalStore((state) => state.symbol);

  return (
    <div className="h-full w-full">
      <PanelGroup direction="horizontal">
        {panels.marketWatch && (
          <>
            <Panel id="marketWatch" order={1} defaultSize={25} minSize={25}>
              <MarketWatch />
            </Panel>

            <PanelResizeHandle className="w-[2px] bg-theme-border" />
          </>
        )}

        <Panel
          id="mainGroup"
          order={2}
          defaultSize={panels.marketWatch ? 75 : 100}
          minSize={50}
        >
          <PanelGroup direction="vertical">
            {(panels.chart || panels.order) && (
              <>
                <Panel
                  id="chartGroup"
                  order={1}
                  defaultSize={panels.activity ? 70 : 100}
                  minSize={20}
                >
                  <PanelGroup direction="horizontal">
                    {panels.chart && (
                      <>
                        <Panel
                          id="chart"
                          order={1}
                          defaultSize={panels.order ? 75 : 100}
                          // minSize={75}
                        >
                          <Chart />
                        </Panel>

                        {panels.order && (
                          <PanelResizeHandle className="w-[2px] bg-theme-border" />
                        )}
                      </>
                    )}

                    {panels.order && (
                      <Panel
                        id="order"
                        order={2}
                        defaultSize={panels.chart ? 25 : 100}
                        minSize={25}
                      >
                        <Order symbolId={symbol} />
                      </Panel>
                    )}
                  </PanelGroup>
                </Panel>

                <PanelResizeHandle className="h-[2px] bg-theme-border" />
              </>
            )}

            {panels.activity && (
              <Panel id="activity" order={2} defaultSize={30} minSize={10}>
                <Activity />
              </Panel>
            )}
          </PanelGroup>
        </Panel>
      </PanelGroup>
    </div>
  );
};

export default Dashboard;
