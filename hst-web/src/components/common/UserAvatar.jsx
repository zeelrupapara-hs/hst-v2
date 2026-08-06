import React from "react";
import { Avatar } from "antd";
import { getInitials } from "../../utils/utils";

const UserAvatar = ({ data }) => {
  if (!data) return null;

  return (
    <div className="flex items-center gap-3">
      <div>
        <Avatar
          size="large"
          className={`uppercase ${!data?.profileImage ? "bg-gray-400" : ""}`}
          src={data?.profileImage}
        >
          <p className="text-base">{getInitials(data?.name)}</p>
        </Avatar>
      </div>

      <div>
        <p className="font-semibold">{data?.name}</p>
        <p className="text-light-gray">{data?.email}</p>
      </div>
    </div>
  );
};

export default UserAvatar;
