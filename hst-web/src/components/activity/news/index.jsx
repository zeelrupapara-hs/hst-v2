import useNewsStore from "../../../store/useNewsStore";
import { openLink } from "../../../utils/utils";

const News = () => {
  const news = useNewsStore((state) => state.news);

  return (
    <div className="h-full flex flex-col overflow-auto">
      {news?.map((item) => (
        <div
          key={item?.id}
          className="p-4 border-b border-theme-border cursor-pointer"
          onClick={() => openLink(item?.link)}
        >
          <p className="mb-1">{item?.title}</p>
          <p className="text-xs">{item?.pub_date}</p>
        </div>
      ))}
    </div>
  );
};

export default News;
