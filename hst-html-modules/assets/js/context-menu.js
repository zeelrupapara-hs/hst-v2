/**
 * MT5-style right-click context menus for list modules.
 */
(function (global) {
  var registry = {};
  var navCache = null;
  var navLoading = null;
  var openMenu = null;
  var openSubmenu = null;

  function esc(s) {
    return String(s == null ? "" : s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/"/g, "&quot;");
  }

  function register(key, builder) {
    registry[key] = builder;
  }

  async function loadNav() {
    if (navCache) return navCache;
    if (navLoading) return navLoading;
    navLoading = (async function () {
      if (!global.HSTApi || !HSTApi.isLoggedIn()) {
        navCache = { can: {} };
        return navCache;
      }
      var res = await HSTApi.get("/api/v1/navigation");
      navCache = res.ok && res.data ? res.data : { can: {} };
      return navCache;
    })();
    return navLoading;
  }

  function can(ctx, right) {
    if (!right) return true;
    if (ctx.can && ctx.can[right]) return true;
    if (!global.HSTApi || !HSTApi.isLoggedIn()) return true;
    return false;
  }

  function visible(item, ctx) {
    if (item.hidden && item.hidden(ctx)) return false;
    if (item.right && !can(ctx, item.right)) return false;
    if (item.when && !item.when(ctx)) return false;
    return true;
  }

  function enabled(item, ctx) {
    if (item.disabled && item.disabled(ctx)) return false;
    return true;
  }

  function checked(item, ctx) {
    if (typeof item.checked === "function") return !!item.checked(ctx);
    return !!item.checked;
  }

  function resolveItems(spec, ctx) {
    if (typeof spec === "function") return spec(ctx) || [];
    return spec || [];
  }

  function closeSubmenu() {
    if (openSubmenu && openSubmenu.parentNode) openSubmenu.parentNode.removeChild(openSubmenu);
    openSubmenu = null;
  }

  function closeMenu() {
    closeSubmenu();
    if (openMenu && openMenu.parentNode) openMenu.parentNode.removeChild(openMenu);
    openMenu = null;
    document.removeEventListener("mousedown", onDocDown, true);
    document.removeEventListener("keydown", onKeyDown, true);
    document.removeEventListener("scroll", closeMenu, true);
  }

  function onDocDown(e) {
    var inMenu = openMenu && openMenu.contains(e.target);
    var inSub = openSubmenu && openSubmenu.contains(e.target);
    if (!inMenu && !inSub) closeMenu();
  }

  function onKeyDown(e) {
    if (e.key === "Escape") closeMenu();
  }

  function showSubmenu(parentEl, items, ctx) {
    closeSubmenu();
    var menu = buildMenu(items, ctx);
    menu.classList.add("ctx-menu", "ctx-submenu");
    document.body.appendChild(menu);
    openSubmenu = menu;

    var rect = parentEl.getBoundingClientRect();
    menu.style.left = Math.round(rect.right - 2) + "px";
    menu.style.top = Math.round(rect.top) + "px";

    var mb = menu.getBoundingClientRect();
    if (mb.right > window.innerWidth) menu.style.left = Math.round(rect.left - mb.width + 2) + "px";
    if (mb.bottom > window.innerHeight) menu.style.top = Math.round(window.innerHeight - mb.height - 4) + "px";
  }

  function buildMenu(items, ctx) {
    var menu = document.createElement("div");
    menu.className = "ctx-menu";
    resolveItems(items, ctx).forEach(function (item) {
      if (!visible(item, ctx)) return;
      if (item.type === "separator") {
        var sep = document.createElement("div");
        sep.className = "ctx-sep";
        menu.appendChild(sep);
        return;
      }

      var row = document.createElement("div");
      row.className = "ctx-item";
      if (!enabled(item, ctx)) row.classList.add("disabled");

      var icon = document.createElement("span");
      icon.className = "ctx-icon";
      if (item.iconId) {
        var base = location.pathname.indexOf("/modules/") >= 0
          ? "../../assets/icons/svg/"
          : "assets/icons/svg/";
        var img = document.createElement("img");
        img.className = "hst-icon";
        img.src = base + item.iconId + ".svg?v=3";
        img.alt = "";
        icon.appendChild(img);
      } else {
        icon.textContent = item.icon || "";
      }
      row.appendChild(icon);

      var label = document.createElement("span");
      label.className = "ctx-label";
      label.textContent = item.label || "";
      row.appendChild(label);

      if (item.check) {
        var mark = document.createElement("span");
        mark.className = "ctx-check";
        mark.textContent = checked(item, ctx) ? "✓" : "";
        row.appendChild(mark);
      }

      if (item.shortcut) {
        var shortcut = document.createElement("span");
        shortcut.className = "ctx-shortcut";
        shortcut.textContent = item.shortcut;
        row.appendChild(shortcut);
      }

      if (item.children) {
        row.classList.add("has-children");
        var arrow = document.createElement("span");
        arrow.className = "ctx-arrow";
        arrow.textContent = "▶";
        row.appendChild(arrow);
        row.addEventListener("mouseenter", function () {
          if (!enabled(item, ctx)) return;
          showSubmenu(row, item.children, ctx);
        });
      } else if (item.action) {
        row.addEventListener("click", function (e) {
          e.stopPropagation();
          if (!enabled(item, ctx)) return;
          closeMenu();
          item.action(ctx);
        });
      }

      menu.appendChild(row);
    });
    return menu;
  }

  function show(key, ctx, x, y) {
    var builder = registry[key];
    if (!builder) return;
    closeMenu();

    if (navCache) {
      ctx.can = navCache.can || {};
      ctx.rights = navCache.rights || {};
    } else {
      ctx.can = {};
      ctx.rights = {};
      loadNav();
    }

    var menu = buildMenu(builder, ctx);
    if (!menu.childElementCount) return;

    document.body.appendChild(menu);
    openMenu = menu;
    menu.style.left = Math.round(x) + "px";
    menu.style.top = Math.round(y) + "px";

    var rect = menu.getBoundingClientRect();
    if (rect.right > window.innerWidth) menu.style.left = Math.round(x - rect.width) + "px";
    if (rect.bottom > window.innerHeight) menu.style.top = Math.round(y - rect.height) + "px";

    document.addEventListener("mousedown", onDocDown, true);
    document.addEventListener("keydown", onKeyDown, true);
    document.addEventListener("scroll", closeMenu, true);
  }

  function toast(msg) {
    var el = document.querySelector(".ctx-toast");
    if (!el) {
      el = document.createElement("div");
      el.className = "ctx-toast";
      document.body.appendChild(el);
    }
    el.textContent = msg;
    el.classList.add("show");
    clearTimeout(toast._t);
    toast._t = setTimeout(function () { el.classList.remove("show"); }, 2800);
  }

  function copyText(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).catch(function () {
        fallbackCopy(text);
      });
    } else {
      fallbackCopy(text);
    }
  }

  function fallbackCopy(text) {
    var ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.left = "-9999px";
    document.body.appendChild(ta);
    ta.select();
    try { document.execCommand("copy"); } catch (_) { /* ignore */ }
    document.body.removeChild(ta);
  }

  global.HSTContextMenu = {
    register: register,
    show: show,
    close: closeMenu,
    toast: toast,
    copyText: copyText,
    can: can
  };
})(window);
