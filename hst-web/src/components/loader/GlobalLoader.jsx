import Loader from "./Loader";

const GlobalLoader = ({ isSuspense = false }) => {
  return (
    <div
      className={`flex items-center justify-center w-full bg-theme-bg ${
        isSuspense ? "h-screen" : "h-full"
      }`}
    >
      <Loader size="large" />
    </div>
  );
};

export default GlobalLoader;
