import {
  getLegacyDetailSrc,
  getModuleDef,
} from "../../features/modules/registry.js";
import { LegacyModuleFrame } from "./LegacyModuleFrame.jsx";
import { ListModule } from "./ListModule.jsx";
import { SettingsModule } from "./SettingsModule.jsx";

export function ModuleView({ panel, moduleId, recordId }) {
  const def = getModuleDef(panel, moduleId);

  if (!def) {
    return (
      <div className="module-root">
        <p className="module-note">Unknown module: {moduleId}</p>
      </div>
    );
  }

  if (recordId) {
    return (
      <LegacyModuleFrame
        src={getLegacyDetailSrc(panel, moduleId, recordId)}
        title={def.title}
      />
    );
  }

  if (def.type === "legacy") {
    return <LegacyModuleFrame src={def.legacySrc} title={def.title} />;
  }

  if (def.type === "settings") {
    return <SettingsModule config={def} />;
  }

  return (
    <ListModule
      config={def}
      detailPath={
        def.legacyDetail
          ? (id) => `/${panel}/${moduleId}/${encodeURIComponent(id)}`
          : null
      }
    />
  );
}
