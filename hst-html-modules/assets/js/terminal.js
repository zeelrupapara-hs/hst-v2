/**
 * Minimal UI behaviour for static MT5-style terminal HTML.
 * No API calls — navigation and toolbox tabs only.
 */
(function () {
  function initNavTree(root) {
    root.querySelectorAll(".chevron:not(.empty)").forEach(function (chev) {
      chev.addEventListener("click", function (e) {
        e.stopPropagation();
        var item = chev.closest(".nav-branch");
        if (!item) return;
        var child = item.querySelector(":scope > ul");
        if (!child) return;
        var open = child.style.display !== "none";
        child.style.display = open ? "none" : "block";
        chev.textContent = open ? "▶" : "▼";
      });
    });

    root.querySelectorAll(".nav-item[data-module]").forEach(function (el) {
      el.addEventListener("click", function () {
        if (el.classList.contains("disabled")) return;
        root.querySelectorAll(".nav-item.active").forEach(function (a) {
          a.classList.remove("active");
        });
        el.classList.add("active");
        var src = el.getAttribute("data-module");
        var frame = document.getElementById("moduleFrame");
        if (frame && src) frame.src = src;
        var status = document.getElementById("statusModule");
        if (status) status.textContent = el.querySelector(".label").textContent;
      });
    });
  }

  function initToolbox(root) {
    var tabs = root.querySelectorAll(".toolbox-tabs button");
    var panels = root.querySelectorAll(".toolbox-panel");
    tabs.forEach(function (tab) {
      tab.addEventListener("click", function () {
        var id = tab.getAttribute("data-panel");
        tabs.forEach(function (t) { t.classList.remove("active"); });
        panels.forEach(function (p) { p.classList.remove("active"); });
        tab.classList.add("active");
        var panel = root.querySelector("#" + id);
        if (panel) panel.classList.add("active");
      });
    });
  }

  function isFullscreen() {
    return !!(document.fullscreenElement || document.webkitFullscreenElement);
  }

  function requestFullscreen() {
    var el = document.documentElement;
    var fn = el.requestFullscreen || el.webkitRequestFullscreen;
    if (!fn) return Promise.reject(new Error("Fullscreen not supported"));
    return fn.call(el);
  }

  function exitFullscreen() {
    var fn = document.exitFullscreen || document.webkitExitFullscreen;
    if (!fn) return Promise.reject(new Error("Fullscreen not supported"));
    return fn.call(document);
  }

  function toggleFullscreen() {
    if (isFullscreen()) return exitFullscreen();
    return requestFullscreen();
  }

  function updateFullscreenUi() {
    var on = isFullscreen();
    document.body.classList.toggle("is-fullscreen", on);
    var btn = document.getElementById("btnFullscreen");
    if (btn) {
      btn.textContent = on ? "❐" : "□";
      btn.title = on ? "Exit Full Screen (F11)" : "Full Screen (F11)";
      btn.setAttribute("aria-label", on ? "Exit Full Screen" : "Full Screen");
    }
    var check = document.getElementById("viewMenuFullscreenCheck");
    if (check) check.textContent = on ? "✓" : "";
  }

  function initFullscreen() {
    var btn = document.getElementById("btnFullscreen");
    if (btn) {
      btn.addEventListener("click", function () {
        toggleFullscreen().catch(function () {});
      });
    }

    document.addEventListener("fullscreenchange", updateFullscreenUi);
    document.addEventListener("webkitfullscreenchange", updateFullscreenUi);

    document.addEventListener("keydown", function (e) {
      if (e.key !== "F11") return;
      e.preventDefault();
      toggleFullscreen().catch(function () {});
    });

    updateFullscreenUi();
  }

  function initViewMenu() {
    var trigger = document.getElementById("menuView");
    if (!trigger) return;

    var menu = null;

    function closeMenu() {
      if (menu && menu.parentNode) menu.parentNode.removeChild(menu);
      menu = null;
      trigger.classList.remove("active");
      document.removeEventListener("mousedown", onDocDown, true);
      document.removeEventListener("keydown", onKeyDown, true);
    }

    function onDocDown(e) {
      if (menu && (menu.contains(e.target) || trigger.contains(e.target))) return;
      closeMenu();
    }

    function onKeyDown(e) {
      if (e.key === "Escape") closeMenu();
    }

    trigger.addEventListener("click", function (e) {
      e.stopPropagation();
      if (menu) {
        closeMenu();
        return;
      }

      menu = document.createElement("div");
      menu.className = "ctx-menu menu-dropdown";
      menu.setAttribute("role", "menu");

      var row = document.createElement("div");
      row.className = "ctx-item";
      row.id = "viewMenuFullscreen";
      row.setAttribute("role", "menuitem");

      var icon = document.createElement("span");
      icon.className = "ctx-icon";
      row.appendChild(icon);

      var label = document.createElement("span");
      label.className = "ctx-label";
      label.textContent = "Full Screen";
      row.appendChild(label);

      var check = document.createElement("span");
      check.className = "ctx-check";
      check.id = "viewMenuFullscreenCheck";
      check.textContent = isFullscreen() ? "✓" : "";
      row.appendChild(check);

      var shortcut = document.createElement("span");
      shortcut.className = "ctx-shortcut";
      shortcut.textContent = "F11";
      row.appendChild(shortcut);

      row.addEventListener("click", function () {
        toggleFullscreen().catch(function () {});
        closeMenu();
      });

      menu.appendChild(row);
      document.body.appendChild(menu);
      trigger.classList.add("active");

      var rect = trigger.getBoundingClientRect();
      menu.style.left = Math.round(rect.left) + "px";
      menu.style.top = Math.round(rect.bottom) + "px";

      document.addEventListener("mousedown", onDocDown, true);
      document.addEventListener("keydown", onKeyDown, true);
    });
  }

  document.addEventListener("DOMContentLoaded", function () {
    var nav = document.getElementById("navTree");
    if (nav) initNavTree(nav);
    var toolbox = document.getElementById("toolbox");
    if (toolbox) initToolbox(toolbox);
    initFullscreen();
    initViewMenu();
  });
})();
