import { useEffect, useState } from "react";
import { Drawer } from "antd";
import { LuX } from "react-icons/lu";
import { getProfile } from "../../api/request/auth";
import usePositionStore from "../../store/usePositionStore";
import UserAvatar from "../common/UserAvatar";
import Icon from "../common/Icon";
import { formatDate } from "../../utils/utils";

const DetailRow = ({ label, value }) => (
  <div className="flex items-start justify-between gap-3 text-xs">
    <span className="text-light-gray shrink-0">{label}</span>
    <span className="text-right break-all">{value ?? "--"}</span>
  </div>
);

const ProfilePanel = ({ open, onClose }) => {
  const summary = usePositionStore((state) => state.summary);
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!open) return;

    let active = true;
    setLoading(true);

    getProfile()
      .then(({ data }) => {
        if (active && data?.data) setProfile(data.data);
      })
      .finally(() => {
        if (active) setLoading(false);
      });

    return () => {
      active = false;
    };
  }, [open]);

  const accountFields = [
    { key: "balance", label: "Balance" },
    { key: "credit", label: "Credit" },
    { key: "equity", label: "Equity" },
    { key: "used_margin", label: "Used Margin" },
    { key: "free_margin", label: "Free Margin" },
    { key: "margin_level", label: "Margin Level" },
    { key: "profit", label: "Profit" },
  ];

  return (
    <Drawer
      open={open}
      onClose={onClose}
      placement="right"
      width={320}
      closable={false}
      className="profile-panel"
      title={
        <div className="flex items-center justify-between">
          <span className="text-sm font-semibold">Profile</span>
          <button onClick={onClose} className="btn-icon-hover">
            <Icon Icon={LuX} size={18} />
          </button>
        </div>
      }
    >
      <div className="flex flex-col gap-5">
        <UserAvatar data={profile} />

        {loading ? (
          <p className="text-xs text-light-gray">Loading profile...</p>
        ) : (
          <>
            <div className="flex flex-col gap-3 border-b border-theme-border pb-4">
              <p className="text-xs font-semibold uppercase tracking-wide text-light-gray">
                Account
              </p>
              <DetailRow label="Login" value={profile?.login} />
              <DetailRow label="Group" value={profile?.group} />
              <DetailRow
                label="Type"
                value={
                  profile?.account_type
                    ? profile.account_type.charAt(0).toUpperCase() +
                      profile.account_type.slice(1)
                    : "--"
                }
              />
              <DetailRow label="Leverage" value={profile?.leverage ? `1:${profile.leverage}` : "--"} />
              <DetailRow
                label="Access"
                value={profile?.read_only ? "Read Only" : "Full Access"}
              />
            </div>

            <div className="flex flex-col gap-3 border-b border-theme-border pb-4">
              <p className="text-xs font-semibold uppercase tracking-wide text-light-gray">
                Contact
              </p>
              <DetailRow label="Email" value={profile?.email} />
              <DetailRow label="Phone" value={profile?.phone} />
              <DetailRow label="Country" value={profile?.country} />
              <DetailRow label="City" value={profile?.city} />
            </div>

            <div className="flex flex-col gap-3 border-b border-theme-border pb-4">
              <p className="text-xs font-semibold uppercase tracking-wide text-light-gray">
                Activity
              </p>
              <DetailRow
                label="Registered"
                value={formatDate(profile?.registration, "DD MMM YYYY")}
              />
              <DetailRow
                label="Last Access"
                value={formatDate(profile?.last_access, "DD MMM YYYY HH:mm")}
              />
            </div>
          </>
        )}

        <div className="flex flex-col gap-3">
          <p className="text-xs font-semibold uppercase tracking-wide text-light-gray">
            Live Summary
          </p>
          {accountFields.map(({ key, label }) => (
            <DetailRow
              key={key}
              label={label}
              value={
                key === "margin_level" && summary?.[key]
                  ? `${summary[key]}%`
                  : summary?.[key] ?? 0
              }
            />
          ))}
        </div>
      </div>
    </Drawer>
  );
};

export default ProfilePanel;
