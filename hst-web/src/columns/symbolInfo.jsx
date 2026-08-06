const defaultRender = (value, record) => (
  <div className="flex flex-col items-center gap-2">
    <span>{value}</span>
    <div className={`w-full h-2 rounded-full ${record?.bgColor}`}></div>
  </div>
);

const columns = [
  {
    title: "Sunday",
    dataIndex: "sunday",
    key: "sunday",
    render: (value, record) => defaultRender(value, record),
  },
  {
    title: "Monday",
    dataIndex: "monday",
    key: "monday",
    render: (value, record) => defaultRender(value, record),
  },
  {
    title: "Tuesday",
    dataIndex: "tuesday",
    key: "tuesday",
    render: (value, record) => defaultRender(value, record),
  },
  {
    title: "Wednesday",
    dataIndex: "wednesday",
    key: "wednesday",
    render: (value, record) => defaultRender(value, record),
  },
  {
    title: "Thursday",
    dataIndex: "thursday",
    key: "thursday",
    render: (value, record) => defaultRender(value, record),
  },
  {
    title: "Friday",
    dataIndex: "friday",
    key: "friday",
    render: (value, record) => defaultRender(value, record),
  },
  {
    title: "Saturday",
    dataIndex: "saturday",
    key: "saturday",
    render: (value, record) => defaultRender(value, record),
  },
];

export default columns;
