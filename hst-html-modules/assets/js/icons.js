/**
 * MT5 SVG icon helper — use across nav, tables, and detail headers.
 */
(function (global) {
  function detectBase() {
    var scripts = document.getElementsByTagName("script");
    for (var i = 0; i < scripts.length; i++) {
      var src = scripts[i].getAttribute("src") || "";
      if (src.indexOf("icons.js") !== -1) {
        return src.replace(/assets\/js\/icons\.js.*$/, "assets/icons/svg/");
      }
    }
    if (location.pathname.indexOf("/modules/") !== -1) {
      return "../../assets/icons/svg/";
    }
    return "assets/icons/svg/";
  }

  var base = detectBase();
  var version = "4";

  function url(id) {
    return base + id + ".svg?v=" + version;
  }

  function escapeAttr(s) {
    return String(s).replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;");
  }

  function img(id, title) {
    var t = title ? ' title="' + escapeAttr(title) + '"' : "";
    return '<img class="hst-icon" src="' + escapeAttr(url(id)) + '" alt=""' + t + ">";
  }

  function use(id, title) {
    var t = title ? ' title="' + escapeAttr(title) + '"' : "";
    return '<svg class="hst-icon" aria-hidden="true"' + t + '><use href="' +
      escapeAttr(base.replace(/svg\/$/, "sprite.svg#icon-" + id)) + '"></use></svg>';
  }

  global.HSTIcons = {
    base: base,
    url: url,
    img: img,
    use: use,
    setBase: function (b) { base = b; this.base = b; }
  };
})(typeof window !== "undefined" ? window : this);
