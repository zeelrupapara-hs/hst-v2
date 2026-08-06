const ChartInactiveOverlay = () => (
  <>
    <div
      className="absolute inset-x-0 top-0 z-20 h-[38px] cursor-not-allowed"
      aria-hidden
    />
    <div
      className="absolute bottom-0 left-0 top-[38px] z-20 w-[52px] cursor-not-allowed"
      aria-hidden
    />
    <div className="absolute bottom-0 right-0 top-[38px] left-[52px] z-20 flex items-center justify-center bg-theme-bg">
      <span className="text-sm font-medium text-theme-text/70">No Data Here</span>
    </div>
  </>
);

export default ChartInactiveOverlay;
