import {
  getLegacyDetailSrc,
  getModuleDef,
} from "../../features/modules/registry.js";
import { LegacyModuleFrame } from "./LegacyModuleFrame.jsx";
import { ListModule } from "./ListModule.jsx";
import { SettingsModule } from "./SettingsModule.jsx";
import { SplitModule } from "./SplitModule.jsx";
import { DatafeedsModule } from "../datafeeds/DatafeedsModule.jsx";
import { SymbolSettingsModule } from "../symbols/SymbolSettingsModule.jsx";

export function ModuleView({ panel, moduleId, recordId }) {
  const def = getModuleDef(panel, moduleId);

  if (!def) {
    return (
      <div className="module-root">
        <p className="module-note">Unknown module: {moduleId}</p>
      </div>
    );
  }

  if (recordId && def.detailKind === "symbol" && def.type === "split") {
    return <SplitModule config={def} panel={panel} moduleId={moduleId} />;
  }

  if (recordId && def.detailKind === "datafeed" && def.type === "split") {
    return <DatafeedsModule config={def} panel={panel} moduleId={moduleId} />;
  }

  if (recordId && def.detailKind === "symbol") {
    return (
      <div className="module-root module-center-host">
        <SymbolSettingsModule symbolId={recordId} />
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

  if (def.type === "split") {
    if (def.detailKind === "datafeed") {
      return <DatafeedsModule config={def} panel={panel} moduleId={moduleId} />;
    }
    return <SplitModule config={def} panel={panel} moduleId={moduleId} />;
  }

  return (
    <ListModule
      config={def}
      detailPath={
        def.detailKind === "symbol" || def.legacyDetail
          ? (id) => `/${panel}/${moduleId}/${encodeURIComponent(id)}`
          : null
      }
    />
  );
}
