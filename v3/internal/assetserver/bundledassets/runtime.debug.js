var __defProp = Object.defineProperty;
var __defProps = Object.defineProperties;
var __getOwnPropDescs = Object.getOwnPropertyDescriptors;
var __getOwnPropSymbols = Object.getOwnPropertySymbols;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __propIsEnum = Object.prototype.propertyIsEnumerable;
var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
var __spreadValues = (a, b) => {
  for (var prop in b || (b = {}))
    if (__hasOwnProp.call(b, prop))
      __defNormalProp(a, prop, b[prop]);
  if (__getOwnPropSymbols)
    for (var prop of __getOwnPropSymbols(b)) {
      if (__propIsEnum.call(b, prop))
        __defNormalProp(a, prop, b[prop]);
    }
  return a;
};
var __spreadProps = (a, b) => __defProps(a, __getOwnPropDescs(b));
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};

// desktop/@wailsio/runtime/src/index.ts
var index_exports = {};
__export(index_exports, {
  Application: () => application_exports,
  Browser: () => browser_exports,
  Call: () => calls_exports,
  CancelError: () => CancelError,
  CancellablePromise: () => CancellablePromise,
  CancelledRejectionError: () => CancelledRejectionError,
  Clipboard: () => clipboard_exports,
  Create: () => create_exports,
  Dialogs: () => dialogs_exports,
  Events: () => events_exports,
  Flags: () => flags_exports,
  IOS: () => ios_exports,
  Screens: () => screens_exports,
  System: () => system_exports,
  WML: () => wml_exports,
  Window: () => window_default,
  clientId: () => clientId,
  getTransport: () => getTransport,
  loadOptionalScript: () => loadOptionalScript,
  objectNames: () => objectNames,
  setTransport: () => setTransport
});

// desktop/@wailsio/runtime/src/wml.ts
var wml_exports = {};
__export(wml_exports, {
  Enable: () => Enable,
  Reload: () => Reload
});

// desktop/@wailsio/runtime/src/browser.ts
var browser_exports = {};
__export(browser_exports, {
  OpenURL: () => OpenURL
});

// desktop/@wailsio/runtime/src/nanoid.ts
var urlAlphabet = "useandom-26T198340PX75pxJACKVERYMINDBUSHWOLF_GQZbfghjklqvwyzrict";
function nanoid(size = 21) {
  let id = "";
  let i = size | 0;
  while (i--) {
    id += urlAlphabet[Math.random() * 64 | 0];
  }
  return id;
}

// desktop/@wailsio/runtime/src/runtime.ts
var runtimeURL = window.location.origin + "/wails/runtime";
var CHUNK_THRESHOLD = 512 * 1024;
var objectNames = Object.freeze({
  Call: 0,
  Clipboard: 1,
  Application: 2,
  Events: 3,
  ContextMenu: 4,
  Dialog: 5,
  Window: 6,
  Screens: 7,
  System: 8,
  Browser: 9,
  CancelCall: 10,
  IOS: 11
});
var clientId = nanoid();
var customTransport = null;
function setTransport(transport) {
  customTransport = transport;
}
function getTransport() {
  return customTransport;
}
function newRuntimeCaller(object, windowName = "") {
  return function(method, args = null) {
    return runtimeCallWithID(object, method, windowName, args);
  };
}
async function runtimeCallWithID(objectID, method, windowName, args) {
  var _a2, _b;
  if (customTransport) {
    return customTransport.call(objectID, method, windowName, args);
  }
  let url = new URL(runtimeURL);
  let body = {
    object: objectID,
    method
  };
  if (args !== null && args !== void 0) {
    body.args = args;
  }
  let headers = {
    ["x-wails-client-id"]: clientId,
    ["Content-Type"]: "application/json"
  };
  if (windowName) {
    headers["x-wails-window-name"] = windowName;
  }
  const bodyStr = JSON.stringify(body);
  let response;
  if (bodyStr.length > CHUNK_THRESHOLD) {
    response = await sendChunked(url, headers, bodyStr);
  } else {
    response = await fetch(url, { method: "POST", headers, body: bodyStr });
  }
  if (!response.ok) {
    throw new Error(await response.text());
  }
  if (((_b = (_a2 = response.headers.get("Content-Type")) == null ? void 0 : _a2.indexOf("application/json")) != null ? _b : -1) !== -1) {
    return response.json();
  } else {
    return response.text();
  }
}
async function sendChunked(url, headers, bodyStr) {
  const chunkId = nanoid();
  const bodyBytes = new TextEncoder().encode(bodyStr);
  const totalChunks = Math.ceil(bodyBytes.length / CHUNK_THRESHOLD);
  for (let i = 0; i < totalChunks - 1; i++) {
    const chunk = bodyBytes.subarray(i * CHUNK_THRESHOLD, (i + 1) * CHUNK_THRESHOLD);
    const resp = await fetch(url, {
      method: "POST",
      headers: __spreadProps(__spreadValues({}, headers), {
        "x-wails-chunk-id": chunkId,
        "x-wails-chunk-index": String(i),
        "x-wails-chunk-total": String(totalChunks)
      }),
      body: chunk
    });
    if (!resp.ok) {
      throw new Error(await resp.text());
    }
    const body = await resp.text();
    if (body.trim()) {
      console.warn("[Wails] Unexpected non-empty body from intermediate chunk:", body);
      throw new Error(body);
    }
  }
  return fetch(url, {
    method: "POST",
    headers: __spreadProps(__spreadValues({}, headers), {
      "x-wails-chunk-id": chunkId,
      "x-wails-chunk-index": String(totalChunks - 1),
      "x-wails-chunk-total": String(totalChunks)
    }),
    body: bodyBytes.subarray((totalChunks - 1) * CHUNK_THRESHOLD)
  });
}

// desktop/@wailsio/runtime/src/browser.ts
var call = newRuntimeCaller(objectNames.Browser);
var BrowserOpenURL = 0;
function OpenURL(url) {
  return call(BrowserOpenURL, { url: url.toString() });
}

// desktop/@wailsio/runtime/src/dialogs.ts
var dialogs_exports = {};
__export(dialogs_exports, {
  Error: () => Error2,
  Info: () => Info,
  OpenFile: () => OpenFile,
  Question: () => Question,
  SaveFile: () => SaveFile,
  Warning: () => Warning
});
window._wails = window._wails || {};
var call2 = newRuntimeCaller(objectNames.Dialog);
var DialogInfo = 0;
var DialogWarning = 1;
var DialogError = 2;
var DialogQuestion = 3;
var DialogOpenFile = 4;
var DialogSaveFile = 5;
function dialog(type, options = {}) {
  return call2(type, options);
}
function Info(options) {
  return dialog(DialogInfo, options);
}
function Warning(options) {
  return dialog(DialogWarning, options);
}
function Error2(options) {
  return dialog(DialogError, options);
}
function Question(options) {
  return dialog(DialogQuestion, options);
}
function OpenFile(options) {
  var _a2;
  return (_a2 = dialog(DialogOpenFile, options)) != null ? _a2 : [];
}
function SaveFile(options) {
  return dialog(DialogSaveFile, options);
}

// desktop/@wailsio/runtime/src/events.ts
var events_exports = {};
__export(events_exports, {
  Emit: () => Emit,
  Off: () => Off,
  OffAll: () => OffAll,
  On: () => On,
  OnMultiple: () => OnMultiple,
  Once: () => Once,
  Types: () => Types,
  WailsEvent: () => WailsEvent
});

// desktop/@wailsio/runtime/src/listener.ts
var eventListeners = /* @__PURE__ */ new Map();
var Listener = class {
  constructor(eventName, callback, maxCallbacks) {
    this.eventName = eventName;
    this.callback = callback;
    this.maxCallbacks = maxCallbacks || -1;
  }
  dispatch(data) {
    try {
      this.callback(data);
    } catch (err) {
      console.error(err);
    }
    if (this.maxCallbacks === -1) return false;
    this.maxCallbacks -= 1;
    return this.maxCallbacks === 0;
  }
};
function listenerOff(listener) {
  let listeners = eventListeners.get(listener.eventName);
  if (!listeners) {
    return;
  }
  listeners = listeners.filter((l) => l !== listener);
  if (listeners.length === 0) {
    eventListeners.delete(listener.eventName);
  } else {
    eventListeners.set(listener.eventName, listeners);
  }
}

// desktop/@wailsio/runtime/src/create.ts
var create_exports = {};
__export(create_exports, {
  Any: () => Any,
  Array: () => Array2,
  ByteSlice: () => ByteSlice,
  Events: () => Events,
  Map: () => Map2,
  Nullable: () => Nullable,
  Struct: () => Struct
});
function Any(source) {
  return source;
}
function ByteSlice(source) {
  return source == null ? "" : source;
}
function Array2(element) {
  if (element === Any) {
    return (source) => source === null ? [] : source;
  }
  return (source) => {
    if (source === null) {
      return [];
    }
    for (let i = 0; i < source.length; i++) {
      source[i] = element(source[i]);
    }
    return source;
  };
}
function Map2(key, value) {
  if (value === Any) {
    return (source) => source === null ? {} : source;
  }
  return (source) => {
    if (source === null) {
      return {};
    }
    for (const key2 in source) {
      source[key2] = value(source[key2]);
    }
    return source;
  };
}
function Nullable(element) {
  if (element === Any) {
    return Any;
  }
  return (source) => source === null ? null : element(source);
}
function Struct(createField) {
  let allAny = true;
  for (const name in createField) {
    if (createField[name] !== Any) {
      allAny = false;
      break;
    }
  }
  if (allAny) {
    return Any;
  }
  return (source) => {
    for (const name in createField) {
      if (name in source) {
        source[name] = createField[name](source[name]);
      }
    }
    return source;
  };
}
var Events = {};

// desktop/@wailsio/runtime/src/event_types.ts
var Types = Object.freeze({
  Windows: Object.freeze({
    APMPowerSettingChange: "windows:APMPowerSettingChange",
    APMPowerStatusChange: "windows:APMPowerStatusChange",
    APMResumeAutomatic: "windows:APMResumeAutomatic",
    APMResumeSuspend: "windows:APMResumeSuspend",
    APMSuspend: "windows:APMSuspend",
    ApplicationStarted: "windows:ApplicationStarted",
    SystemThemeChanged: "windows:SystemThemeChanged",
    WebViewNavigationCompleted: "windows:WebViewNavigationCompleted",
    WindowActive: "windows:WindowActive",
    WindowBackgroundErase: "windows:WindowBackgroundErase",
    WindowClickActive: "windows:WindowClickActive",
    WindowClosing: "windows:WindowClosing",
    WindowDidMove: "windows:WindowDidMove",
    WindowDidResize: "windows:WindowDidResize",
    WindowDPIChanged: "windows:WindowDPIChanged",
    WindowDragDrop: "windows:WindowDragDrop",
    WindowDragEnter: "windows:WindowDragEnter",
    WindowDragLeave: "windows:WindowDragLeave",
    WindowDragOver: "windows:WindowDragOver",
    WindowEndMove: "windows:WindowEndMove",
    WindowEndResize: "windows:WindowEndResize",
    WindowFullscreen: "windows:WindowFullscreen",
    WindowHide: "windows:WindowHide",
    WindowInactive: "windows:WindowInactive",
    WindowKeyDown: "windows:WindowKeyDown",
    WindowKeyUp: "windows:WindowKeyUp",
    WindowKillFocus: "windows:WindowKillFocus",
    WindowNonClientHit: "windows:WindowNonClientHit",
    WindowNonClientMouseDown: "windows:WindowNonClientMouseDown",
    WindowNonClientMouseLeave: "windows:WindowNonClientMouseLeave",
    WindowNonClientMouseMove: "windows:WindowNonClientMouseMove",
    WindowNonClientMouseUp: "windows:WindowNonClientMouseUp",
    WindowPaint: "windows:WindowPaint",
    WindowRestore: "windows:WindowRestore",
    WindowSetFocus: "windows:WindowSetFocus",
    WindowShow: "windows:WindowShow",
    WindowStartMove: "windows:WindowStartMove",
    WindowStartResize: "windows:WindowStartResize",
    WindowUnFullscreen: "windows:WindowUnFullscreen",
    WindowZOrderChanged: "windows:WindowZOrderChanged",
    WindowMinimise: "windows:WindowMinimise",
    WindowUnMinimise: "windows:WindowUnMinimise",
    WindowMaximise: "windows:WindowMaximise",
    WindowUnMaximise: "windows:WindowUnMaximise"
  }),
  Mac: Object.freeze({
    ApplicationDidBecomeActive: "mac:ApplicationDidBecomeActive",
    ApplicationDidChangeBackingProperties: "mac:ApplicationDidChangeBackingProperties",
    ApplicationDidChangeEffectiveAppearance: "mac:ApplicationDidChangeEffectiveAppearance",
    ApplicationDidChangeIcon: "mac:ApplicationDidChangeIcon",
    ApplicationDidChangeOcclusionState: "mac:ApplicationDidChangeOcclusionState",
    ApplicationDidChangeScreenParameters: "mac:ApplicationDidChangeScreenParameters",
    ApplicationDidChangeStatusBarFrame: "mac:ApplicationDidChangeStatusBarFrame",
    ApplicationDidChangeStatusBarOrientation: "mac:ApplicationDidChangeStatusBarOrientation",
    ApplicationDidChangeTheme: "mac:ApplicationDidChangeTheme",
    ApplicationDidFinishLaunching: "mac:ApplicationDidFinishLaunching",
    ApplicationDidHide: "mac:ApplicationDidHide",
    ApplicationDidResignActive: "mac:ApplicationDidResignActive",
    ApplicationDidUnhide: "mac:ApplicationDidUnhide",
    ApplicationDidUpdate: "mac:ApplicationDidUpdate",
    ApplicationShouldHandleReopen: "mac:ApplicationShouldHandleReopen",
    ApplicationWillBecomeActive: "mac:ApplicationWillBecomeActive",
    ApplicationWillFinishLaunching: "mac:ApplicationWillFinishLaunching",
    ApplicationWillHide: "mac:ApplicationWillHide",
    ApplicationWillResignActive: "mac:ApplicationWillResignActive",
    ApplicationWillTerminate: "mac:ApplicationWillTerminate",
    ApplicationWillUnhide: "mac:ApplicationWillUnhide",
    ApplicationWillUpdate: "mac:ApplicationWillUpdate",
    MenuDidAddItem: "mac:MenuDidAddItem",
    MenuDidBeginTracking: "mac:MenuDidBeginTracking",
    MenuDidClose: "mac:MenuDidClose",
    MenuDidDisplayItem: "mac:MenuDidDisplayItem",
    MenuDidEndTracking: "mac:MenuDidEndTracking",
    MenuDidHighlightItem: "mac:MenuDidHighlightItem",
    MenuDidOpen: "mac:MenuDidOpen",
    MenuDidPopUp: "mac:MenuDidPopUp",
    MenuDidRemoveItem: "mac:MenuDidRemoveItem",
    MenuDidSendAction: "mac:MenuDidSendAction",
    MenuDidSendActionToItem: "mac:MenuDidSendActionToItem",
    MenuDidUpdate: "mac:MenuDidUpdate",
    MenuWillAddItem: "mac:MenuWillAddItem",
    MenuWillBeginTracking: "mac:MenuWillBeginTracking",
    MenuWillDisplayItem: "mac:MenuWillDisplayItem",
    MenuWillEndTracking: "mac:MenuWillEndTracking",
    MenuWillHighlightItem: "mac:MenuWillHighlightItem",
    MenuWillOpen: "mac:MenuWillOpen",
    MenuWillPopUp: "mac:MenuWillPopUp",
    MenuWillRemoveItem: "mac:MenuWillRemoveItem",
    MenuWillSendAction: "mac:MenuWillSendAction",
    MenuWillSendActionToItem: "mac:MenuWillSendActionToItem",
    MenuWillUpdate: "mac:MenuWillUpdate",
    WebViewDidCommitNavigation: "mac:WebViewDidCommitNavigation",
    WebViewDidFinishNavigation: "mac:WebViewDidFinishNavigation",
    WebViewDidReceiveServerRedirectForProvisionalNavigation: "mac:WebViewDidReceiveServerRedirectForProvisionalNavigation",
    WebViewDidStartProvisionalNavigation: "mac:WebViewDidStartProvisionalNavigation",
    WindowDidBecomeKey: "mac:WindowDidBecomeKey",
    WindowDidBecomeMain: "mac:WindowDidBecomeMain",
    WindowDidBeginSheet: "mac:WindowDidBeginSheet",
    WindowDidChangeAlpha: "mac:WindowDidChangeAlpha",
    WindowDidChangeBackingLocation: "mac:WindowDidChangeBackingLocation",
    WindowDidChangeBackingProperties: "mac:WindowDidChangeBackingProperties",
    WindowDidChangeCollectionBehavior: "mac:WindowDidChangeCollectionBehavior",
    WindowDidChangeEffectiveAppearance: "mac:WindowDidChangeEffectiveAppearance",
    WindowDidChangeOcclusionState: "mac:WindowDidChangeOcclusionState",
    WindowDidChangeOrderingMode: "mac:WindowDidChangeOrderingMode",
    WindowDidChangeScreen: "mac:WindowDidChangeScreen",
    WindowDidChangeScreenParameters: "mac:WindowDidChangeScreenParameters",
    WindowDidChangeScreenProfile: "mac:WindowDidChangeScreenProfile",
    WindowDidChangeScreenSpace: "mac:WindowDidChangeScreenSpace",
    WindowDidChangeScreenSpaceProperties: "mac:WindowDidChangeScreenSpaceProperties",
    WindowDidChangeSharingType: "mac:WindowDidChangeSharingType",
    WindowDidChangeSpace: "mac:WindowDidChangeSpace",
    WindowDidChangeSpaceOrderingMode: "mac:WindowDidChangeSpaceOrderingMode",
    WindowDidChangeTitle: "mac:WindowDidChangeTitle",
    WindowDidChangeToolbar: "mac:WindowDidChangeToolbar",
    WindowDidDeminiaturize: "mac:WindowDidDeminiaturize",
    WindowDidEndSheet: "mac:WindowDidEndSheet",
    WindowDidEnterFullScreen: "mac:WindowDidEnterFullScreen",
    WindowDidEnterVersionBrowser: "mac:WindowDidEnterVersionBrowser",
    WindowDidExitFullScreen: "mac:WindowDidExitFullScreen",
    WindowDidExitVersionBrowser: "mac:WindowDidExitVersionBrowser",
    WindowDidExpose: "mac:WindowDidExpose",
    WindowDidFocus: "mac:WindowDidFocus",
    WindowDidMiniaturize: "mac:WindowDidMiniaturize",
    WindowDidMove: "mac:WindowDidMove",
    WindowDidOrderOffScreen: "mac:WindowDidOrderOffScreen",
    WindowDidOrderOnScreen: "mac:WindowDidOrderOnScreen",
    WindowDidResignKey: "mac:WindowDidResignKey",
    WindowDidResignMain: "mac:WindowDidResignMain",
    WindowDidResize: "mac:WindowDidResize",
    WindowDidUpdate: "mac:WindowDidUpdate",
    WindowDidUpdateAlpha: "mac:WindowDidUpdateAlpha",
    WindowDidUpdateCollectionBehavior: "mac:WindowDidUpdateCollectionBehavior",
    WindowDidUpdateCollectionProperties: "mac:WindowDidUpdateCollectionProperties",
    WindowDidUpdateShadow: "mac:WindowDidUpdateShadow",
    WindowDidUpdateTitle: "mac:WindowDidUpdateTitle",
    WindowDidUpdateToolbar: "mac:WindowDidUpdateToolbar",
    WindowDidZoom: "mac:WindowDidZoom",
    WindowFileDraggingEntered: "mac:WindowFileDraggingEntered",
    WindowFileDraggingExited: "mac:WindowFileDraggingExited",
    WindowFileDraggingPerformed: "mac:WindowFileDraggingPerformed",
    WindowHide: "mac:WindowHide",
    WindowMaximise: "mac:WindowMaximise",
    WindowUnMaximise: "mac:WindowUnMaximise",
    WindowMinimise: "mac:WindowMinimise",
    WindowUnMinimise: "mac:WindowUnMinimise",
    WindowShouldClose: "mac:WindowShouldClose",
    WindowShow: "mac:WindowShow",
    WindowWillBecomeKey: "mac:WindowWillBecomeKey",
    WindowWillBecomeMain: "mac:WindowWillBecomeMain",
    WindowWillBeginSheet: "mac:WindowWillBeginSheet",
    WindowWillChangeOrderingMode: "mac:WindowWillChangeOrderingMode",
    WindowWillClose: "mac:WindowWillClose",
    WindowWillDeminiaturize: "mac:WindowWillDeminiaturize",
    WindowWillEnterFullScreen: "mac:WindowWillEnterFullScreen",
    WindowWillEnterVersionBrowser: "mac:WindowWillEnterVersionBrowser",
    WindowWillExitFullScreen: "mac:WindowWillExitFullScreen",
    WindowWillExitVersionBrowser: "mac:WindowWillExitVersionBrowser",
    WindowWillFocus: "mac:WindowWillFocus",
    WindowWillMiniaturize: "mac:WindowWillMiniaturize",
    WindowWillMove: "mac:WindowWillMove",
    WindowWillOrderOffScreen: "mac:WindowWillOrderOffScreen",
    WindowWillOrderOnScreen: "mac:WindowWillOrderOnScreen",
    WindowWillResignMain: "mac:WindowWillResignMain",
    WindowWillResize: "mac:WindowWillResize",
    WindowWillUnfocus: "mac:WindowWillUnfocus",
    WindowWillUpdate: "mac:WindowWillUpdate",
    WindowWillUpdateAlpha: "mac:WindowWillUpdateAlpha",
    WindowWillUpdateCollectionBehavior: "mac:WindowWillUpdateCollectionBehavior",
    WindowWillUpdateCollectionProperties: "mac:WindowWillUpdateCollectionProperties",
    WindowWillUpdateShadow: "mac:WindowWillUpdateShadow",
    WindowWillUpdateTitle: "mac:WindowWillUpdateTitle",
    WindowWillUpdateToolbar: "mac:WindowWillUpdateToolbar",
    WindowWillUpdateVisibility: "mac:WindowWillUpdateVisibility",
    WindowWillUseStandardFrame: "mac:WindowWillUseStandardFrame",
    WindowZoomIn: "mac:WindowZoomIn",
    WindowZoomOut: "mac:WindowZoomOut",
    WindowZoomReset: "mac:WindowZoomReset"
  }),
  Linux: Object.freeze({
    ApplicationStartup: "linux:ApplicationStartup",
    SystemThemeChanged: "linux:SystemThemeChanged",
    WindowDeleteEvent: "linux:WindowDeleteEvent",
    WindowDidMove: "linux:WindowDidMove",
    WindowDidResize: "linux:WindowDidResize",
    WindowFocusIn: "linux:WindowFocusIn",
    WindowFocusOut: "linux:WindowFocusOut",
    WindowLoadStarted: "linux:WindowLoadStarted",
    WindowLoadRedirected: "linux:WindowLoadRedirected",
    WindowLoadCommitted: "linux:WindowLoadCommitted",
    WindowLoadFinished: "linux:WindowLoadFinished"
  }),
  iOS: Object.freeze({
    ApplicationDidBecomeActive: "ios:ApplicationDidBecomeActive",
    ApplicationDidEnterBackground: "ios:ApplicationDidEnterBackground",
    ApplicationDidFinishLaunching: "ios:ApplicationDidFinishLaunching",
    ApplicationDidReceiveMemoryWarning: "ios:ApplicationDidReceiveMemoryWarning",
    ApplicationWillEnterForeground: "ios:ApplicationWillEnterForeground",
    ApplicationWillResignActive: "ios:ApplicationWillResignActive",
    ApplicationWillTerminate: "ios:ApplicationWillTerminate",
    WindowDidLoad: "ios:WindowDidLoad",
    WindowWillAppear: "ios:WindowWillAppear",
    WindowDidAppear: "ios:WindowDidAppear",
    WindowWillDisappear: "ios:WindowWillDisappear",
    WindowDidDisappear: "ios:WindowDidDisappear",
    WindowSafeAreaInsetsChanged: "ios:WindowSafeAreaInsetsChanged",
    WindowOrientationChanged: "ios:WindowOrientationChanged",
    WindowTouchBegan: "ios:WindowTouchBegan",
    WindowTouchMoved: "ios:WindowTouchMoved",
    WindowTouchEnded: "ios:WindowTouchEnded",
    WindowTouchCancelled: "ios:WindowTouchCancelled",
    WebViewDidStartNavigation: "ios:WebViewDidStartNavigation",
    WebViewDidFinishNavigation: "ios:WebViewDidFinishNavigation",
    WebViewDidFailNavigation: "ios:WebViewDidFailNavigation",
    WebViewDecidePolicyForNavigationAction: "ios:WebViewDecidePolicyForNavigationAction"
  }),
  Common: Object.freeze({
    ApplicationOpenedWithFile: "common:ApplicationOpenedWithFile",
    ApplicationStarted: "common:ApplicationStarted",
    ApplicationLaunchedWithUrl: "common:ApplicationLaunchedWithUrl",
    ThemeChanged: "common:ThemeChanged",
    WindowClosing: "common:WindowClosing",
    WindowDidMove: "common:WindowDidMove",
    WindowDidResize: "common:WindowDidResize",
    WindowDPIChanged: "common:WindowDPIChanged",
    WindowFilesDropped: "common:WindowFilesDropped",
    WindowFocus: "common:WindowFocus",
    WindowFullscreen: "common:WindowFullscreen",
    WindowHide: "common:WindowHide",
    WindowLostFocus: "common:WindowLostFocus",
    WindowMaximise: "common:WindowMaximise",
    WindowMinimise: "common:WindowMinimise",
    WindowToggleFrameless: "common:WindowToggleFrameless",
    WindowRestore: "common:WindowRestore",
    WindowRuntimeReady: "common:WindowRuntimeReady",
    WindowShow: "common:WindowShow",
    WindowUnFullscreen: "common:WindowUnFullscreen",
    WindowUnMaximise: "common:WindowUnMaximise",
    WindowUnMinimise: "common:WindowUnMinimise",
    WindowZoom: "common:WindowZoom",
    WindowZoomIn: "common:WindowZoomIn",
    WindowZoomOut: "common:WindowZoomOut",
    WindowZoomReset: "common:WindowZoomReset"
  })
});

// desktop/@wailsio/runtime/src/events.ts
window._wails = window._wails || {};
window._wails.dispatchWailsEvent = dispatchWailsEvent;
var call3 = newRuntimeCaller(objectNames.Events);
var EmitMethod = 0;
var WailsEvent = class {
  constructor(name, data) {
    this.name = name;
    this.data = data != null ? data : null;
  }
};
function dispatchWailsEvent(event) {
  let listeners = eventListeners.get(event.name);
  if (!listeners) {
    return;
  }
  let wailsEvent = new WailsEvent(
    event.name,
    event.name in Events ? Events[event.name](event.data) : event.data
  );
  if ("sender" in event) {
    wailsEvent.sender = event.sender;
  }
  listeners = listeners.filter((listener) => !listener.dispatch(wailsEvent));
  if (listeners.length === 0) {
    eventListeners.delete(event.name);
  } else {
    eventListeners.set(event.name, listeners);
  }
}
function OnMultiple(eventName, callback, maxCallbacks) {
  let listeners = eventListeners.get(eventName) || [];
  const thisListener = new Listener(eventName, callback, maxCallbacks);
  listeners.push(thisListener);
  eventListeners.set(eventName, listeners);
  return () => listenerOff(thisListener);
}
function On(eventName, callback) {
  return OnMultiple(eventName, callback, -1);
}
function Once(eventName, callback) {
  return OnMultiple(eventName, callback, 1);
}
function Off(...eventNames) {
  eventNames.forEach((eventName) => eventListeners.delete(eventName));
}
function OffAll() {
  eventListeners.clear();
}
function Emit(name, data) {
  return call3(EmitMethod, new WailsEvent(name, data));
}

// desktop/@wailsio/runtime/src/utils.ts
function debugLog(message) {
  console.log(
    "%c wails3 %c " + message + " ",
    "background: #aa0000; color: #fff; border-radius: 3px 0px 0px 3px; padding: 1px; font-size: 0.7rem",
    "background: #009900; color: #fff; border-radius: 0px 3px 3px 0px; padding: 1px; font-size: 0.7rem"
  );
}
function canTrackButtons() {
  return new MouseEvent("mousedown").buttons === 0;
}
function canAbortListeners() {
  if (!EventTarget || !AbortSignal || !AbortController)
    return false;
  let result = true;
  const target = new EventTarget();
  const controller = new AbortController();
  target.addEventListener("test", () => {
    result = false;
  }, { signal: controller.signal });
  controller.abort();
  target.dispatchEvent(new CustomEvent("test"));
  return result;
}
function eventTarget(event) {
  var _a2;
  if (event.target instanceof HTMLElement) {
    return event.target;
  } else if (!(event.target instanceof HTMLElement) && event.target instanceof Node) {
    return (_a2 = event.target.parentElement) != null ? _a2 : document.body;
  } else {
    return document.body;
  }
}
var isReady = false;
document.addEventListener("DOMContentLoaded", () => {
  isReady = true;
});
function whenReady(callback) {
  if (isReady || document.readyState === "complete") {
    callback();
  } else {
    document.addEventListener("DOMContentLoaded", callback);
  }
}

// desktop/@wailsio/runtime/src/window.ts
var DROP_TARGET_ATTRIBUTE = "data-file-drop-target";
var DROP_TARGET_ACTIVE_CLASS = "file-drop-target-active";
var currentDropTarget = null;
var PositionMethod = 0;
var CenterMethod = 1;
var CloseMethod = 2;
var DisableSizeConstraintsMethod = 3;
var EnableSizeConstraintsMethod = 4;
var FocusMethod = 5;
var ForceReloadMethod = 6;
var FullscreenMethod = 7;
var GetScreenMethod = 8;
var GetZoomMethod = 9;
var HeightMethod = 10;
var HideMethod = 11;
var IsFocusedMethod = 12;
var IsFullscreenMethod = 13;
var IsMaximisedMethod = 14;
var IsMinimisedMethod = 15;
var MaximiseMethod = 16;
var MinimiseMethod = 17;
var NameMethod = 18;
var OpenDevToolsMethod = 19;
var RelativePositionMethod = 20;
var ReloadMethod = 21;
var ResizableMethod = 22;
var RestoreMethod = 23;
var SetPositionMethod = 24;
var SetAlwaysOnTopMethod = 25;
var SetBackgroundColourMethod = 26;
var SetFramelessMethod = 27;
var SetFullscreenButtonEnabledMethod = 28;
var SetMaxSizeMethod = 29;
var SetMinSizeMethod = 30;
var SetRelativePositionMethod = 31;
var SetResizableMethod = 32;
var SetSizeMethod = 33;
var SetTitleMethod = 34;
var SetZoomMethod = 35;
var ShowMethod = 36;
var SizeMethod = 37;
var ToggleFullscreenMethod = 38;
var ToggleMaximiseMethod = 39;
var ToggleFramelessMethod = 40;
var UnFullscreenMethod = 41;
var UnMaximiseMethod = 42;
var UnMinimiseMethod = 43;
var WidthMethod = 44;
var ZoomMethod = 45;
var ZoomInMethod = 46;
var ZoomOutMethod = 47;
var ZoomResetMethod = 48;
var SnapAssistMethod = 49;
var FilesDropped = 50;
var PrintMethod = 51;
var SetScreenMethod = 52;
function getDropTargetElement(element) {
  if (!element) {
    return null;
  }
  return element.closest("[".concat(DROP_TARGET_ATTRIBUTE, "]"));
}
function canResolveFilePaths() {
  var _a2, _b, _c, _d;
  if (((_b = (_a2 = window.chrome) == null ? void 0 : _a2.webview) == null ? void 0 : _b.postMessageWithAdditionalObjects) == null) {
    return false;
  }
  return ((_d = (_c = window._wails) == null ? void 0 : _c.flags) == null ? void 0 : _d.enableFileDrop) === true;
}
function resolveFilePaths(x, y, files) {
  var _a2, _b;
  if ((_b = (_a2 = window.chrome) == null ? void 0 : _a2.webview) == null ? void 0 : _b.postMessageWithAdditionalObjects) {
    window.chrome.webview.postMessageWithAdditionalObjects("file:drop:".concat(x, ":").concat(y), files);
  }
}
var nativeDragActive = false;
function cleanupNativeDrag() {
  nativeDragActive = false;
  if (currentDropTarget) {
    currentDropTarget.classList.remove(DROP_TARGET_ACTIVE_CLASS);
    currentDropTarget = null;
  }
}
function handleDragEnter() {
  var _a2, _b;
  if (((_b = (_a2 = window._wails) == null ? void 0 : _a2.flags) == null ? void 0 : _b.enableFileDrop) === false) {
    return;
  }
  nativeDragActive = true;
}
function handleDragLeave() {
  cleanupNativeDrag();
}
function handleDragOver(x, y) {
  var _a2, _b;
  if (!nativeDragActive) return;
  if (((_b = (_a2 = window._wails) == null ? void 0 : _a2.flags) == null ? void 0 : _b.enableFileDrop) === false) {
    return;
  }
  const targetElement = document.elementFromPoint(x, y);
  const dropTarget = getDropTargetElement(targetElement);
  if (currentDropTarget && currentDropTarget !== dropTarget) {
    currentDropTarget.classList.remove(DROP_TARGET_ACTIVE_CLASS);
  }
  if (dropTarget) {
    dropTarget.classList.add(DROP_TARGET_ACTIVE_CLASS);
    currentDropTarget = dropTarget;
  } else {
    currentDropTarget = null;
  }
}
var callerSym = /* @__PURE__ */ Symbol("caller");
callerSym;
var _Window = class _Window {
  /**
   * Initialises a window object with the specified name.
   *
   * @private
   * @param name - The name of the target window.
   */
  constructor(name = "") {
    this[callerSym] = newRuntimeCaller(objectNames.Window, name);
    for (const method of Object.getOwnPropertyNames(_Window.prototype)) {
      if (method !== "constructor" && typeof this[method] === "function") {
        this[method] = this[method].bind(this);
      }
    }
  }
  /**
   * Gets the specified window.
   *
   * @param name - The name of the window to get.
   * @returns The corresponding window object.
   */
  Get(name) {
    return new _Window(name);
  }
  /**
   * Returns the absolute position of the window.
   *
   * @returns The current absolute position of the window.
   */
  Position() {
    return this[callerSym](PositionMethod);
  }
  /**
   * Centers the window on the screen.
   */
  Center() {
    return this[callerSym](CenterMethod);
  }
  /**
   * Closes the window.
   */
  Close() {
    return this[callerSym](CloseMethod);
  }
  /**
   * Disables min/max size constraints.
   */
  DisableSizeConstraints() {
    return this[callerSym](DisableSizeConstraintsMethod);
  }
  /**
   * Enables min/max size constraints.
   */
  EnableSizeConstraints() {
    return this[callerSym](EnableSizeConstraintsMethod);
  }
  /**
   * Focuses the window.
   */
  Focus() {
    return this[callerSym](FocusMethod);
  }
  /**
   * Forces the window to reload the page assets.
   */
  ForceReload() {
    return this[callerSym](ForceReloadMethod);
  }
  /**
   * Switches the window to fullscreen mode.
   */
  Fullscreen() {
    return this[callerSym](FullscreenMethod);
  }
  /**
   * Returns the screen that the window is on.
   *
   * @returns The screen the window is currently on.
   */
  GetScreen() {
    return this[callerSym](GetScreenMethod);
  }
  /**
   * Returns the current zoom level of the window.
   *
   * @returns The current zoom level.
   */
  GetZoom() {
    return this[callerSym](GetZoomMethod);
  }
  /**
   * Returns the height of the window.
   *
   * @returns The current height of the window.
   */
  Height() {
    return this[callerSym](HeightMethod);
  }
  /**
   * Hides the window.
   */
  Hide() {
    return this[callerSym](HideMethod);
  }
  /**
   * Returns true if the window is focused.
   *
   * @returns Whether the window is currently focused.
   */
  IsFocused() {
    return this[callerSym](IsFocusedMethod);
  }
  /**
   * Returns true if the window is fullscreen.
   *
   * @returns Whether the window is currently fullscreen.
   */
  IsFullscreen() {
    return this[callerSym](IsFullscreenMethod);
  }
  /**
   * Returns true if the window is maximised.
   *
   * @returns Whether the window is currently maximised.
   */
  IsMaximised() {
    return this[callerSym](IsMaximisedMethod);
  }
  /**
   * Returns true if the window is minimised.
   *
   * @returns Whether the window is currently minimised.
   */
  IsMinimised() {
    return this[callerSym](IsMinimisedMethod);
  }
  /**
   * Maximises the window.
   */
  Maximise() {
    return this[callerSym](MaximiseMethod);
  }
  /**
   * Minimises the window.
   */
  Minimise() {
    return this[callerSym](MinimiseMethod);
  }
  /**
   * Returns the name of the window.
   *
   * @returns The name of the window.
   */
  Name() {
    return this[callerSym](NameMethod);
  }
  /**
   * Opens the development tools pane.
   */
  OpenDevTools() {
    return this[callerSym](OpenDevToolsMethod);
  }
  /**
   * Returns the relative position of the window to the screen.
   *
   * @returns The current relative position of the window.
   */
  RelativePosition() {
    return this[callerSym](RelativePositionMethod);
  }
  /**
   * Reloads the page assets.
   */
  Reload() {
    return this[callerSym](ReloadMethod);
  }
  /**
   * Returns true if the window is resizable.
   *
   * @returns Whether the window is currently resizable.
   */
  Resizable() {
    return this[callerSym](ResizableMethod);
  }
  /**
   * Restores the window to its previous state if it was previously minimised, maximised or fullscreen.
   */
  Restore() {
    return this[callerSym](RestoreMethod);
  }
  /**
   * Sets the absolute position of the window.
   *
   * @param x - The desired horizontal absolute position of the window.
   * @param y - The desired vertical absolute position of the window.
   */
  SetPosition(x, y) {
    return this[callerSym](SetPositionMethod, { x, y });
  }
  /**
   * Sets the window to be always on top.
   *
   * @param alwaysOnTop - Whether the window should stay on top.
   */
  SetAlwaysOnTop(alwaysOnTop) {
    return this[callerSym](SetAlwaysOnTopMethod, { alwaysOnTop });
  }
  /**
   * Sets the background colour of the window.
   *
   * @param r - The desired red component of the window background.
   * @param g - The desired green component of the window background.
   * @param b - The desired blue component of the window background.
   * @param a - The desired alpha component of the window background.
   */
  SetBackgroundColour(r, g, b, a) {
    return this[callerSym](SetBackgroundColourMethod, { r, g, b, a });
  }
  /**
   * Removes the window frame and title bar.
   *
   * @param frameless - Whether the window should be frameless.
   */
  SetFrameless(frameless) {
    return this[callerSym](SetFramelessMethod, { frameless });
  }
  /**
   * Disables the system fullscreen button.
   *
   * @param enabled - Whether the fullscreen button should be enabled.
   */
  SetFullscreenButtonEnabled(enabled) {
    return this[callerSym](SetFullscreenButtonEnabledMethod, { enabled });
  }
  /**
   * Sets the maximum size of the window.
   *
   * @param width - The desired maximum width of the window.
   * @param height - The desired maximum height of the window.
   */
  SetMaxSize(width, height) {
    return this[callerSym](SetMaxSizeMethod, { width, height });
  }
  /**
   * Sets the minimum size of the window.
   *
   * @param width - The desired minimum width of the window.
   * @param height - The desired minimum height of the window.
   */
  SetMinSize(width, height) {
    return this[callerSym](SetMinSizeMethod, { width, height });
  }
  /**
   * Sets the relative position of the window to the screen.
   *
   * @param x - The desired horizontal relative position of the window.
   * @param y - The desired vertical relative position of the window.
   */
  SetRelativePosition(x, y) {
    return this[callerSym](SetRelativePositionMethod, { x, y });
  }
  /**
   * Sets whether the window is resizable.
   *
   * @param resizable - Whether the window should be resizable.
   */
  SetResizable(resizable2) {
    return this[callerSym](SetResizableMethod, { resizable: resizable2 });
  }
  /**
   * Sets the size of the window.
   *
   * @param width - The desired width of the window.
   * @param height - The desired height of the window.
   */
  SetSize(width, height) {
    return this[callerSym](SetSizeMethod, { width, height });
  }
  /**
   * Sets the title of the window.
   *
   * @param title - The desired title of the window.
   */
  SetTitle(title) {
    return this[callerSym](SetTitleMethod, { title });
  }
  /**
   * Sets the zoom level of the window.
   *
   * @param zoom - The desired zoom level.
   */
  SetZoom(zoom) {
    return this[callerSym](SetZoomMethod, { zoom });
  }
  /**
   * Shows the window.
   */
  Show() {
    return this[callerSym](ShowMethod);
  }
  /**
   * Returns the size of the window.
   *
   * @returns The current size of the window.
   */
  Size() {
    return this[callerSym](SizeMethod);
  }
  /**
   * Toggles the window between fullscreen and normal.
   */
  ToggleFullscreen() {
    return this[callerSym](ToggleFullscreenMethod);
  }
  /**
   * Toggles the window between maximised and normal.
   */
  ToggleMaximise() {
    return this[callerSym](ToggleMaximiseMethod);
  }
  /**
   * Toggles the window between frameless and normal.
   */
  ToggleFrameless() {
    return this[callerSym](ToggleFramelessMethod);
  }
  /**
   * Un-fullscreens the window.
   */
  UnFullscreen() {
    return this[callerSym](UnFullscreenMethod);
  }
  /**
   * Un-maximises the window.
   */
  UnMaximise() {
    return this[callerSym](UnMaximiseMethod);
  }
  /**
   * Un-minimises the window.
   */
  UnMinimise() {
    return this[callerSym](UnMinimiseMethod);
  }
  /**
   * Returns the width of the window.
   *
   * @returns The current width of the window.
   */
  Width() {
    return this[callerSym](WidthMethod);
  }
  /**
   * Zooms the window.
   */
  Zoom() {
    return this[callerSym](ZoomMethod);
  }
  /**
   * Increases the zoom level of the webview content.
   */
  ZoomIn() {
    return this[callerSym](ZoomInMethod);
  }
  /**
   * Decreases the zoom level of the webview content.
   */
  ZoomOut() {
    return this[callerSym](ZoomOutMethod);
  }
  /**
   * Resets the zoom level of the webview content.
   */
  ZoomReset() {
    return this[callerSym](ZoomResetMethod);
  }
  /**
   * Handles file drops originating from platform-specific code (e.g., macOS/Linux native drag-and-drop).
   * Gathers information about the drop target element and sends it back to the Go backend.
   *
   * @param filenames - An array of file paths (strings) that were dropped.
   * @param x - The x-coordinate of the drop event (CSS pixels).
   * @param y - The y-coordinate of the drop event (CSS pixels).
   */
  HandlePlatformFileDrop(filenames, x, y) {
    var _a2, _b;
    if (((_b = (_a2 = window._wails) == null ? void 0 : _a2.flags) == null ? void 0 : _b.enableFileDrop) === false) {
      return;
    }
    const element = document.elementFromPoint(x, y);
    const dropTarget = getDropTargetElement(element);
    if (!dropTarget) {
      return;
    }
    const elementDetails = {
      id: dropTarget.id,
      classList: Array.from(dropTarget.classList),
      attributes: {}
    };
    for (let i = 0; i < dropTarget.attributes.length; i++) {
      const attr = dropTarget.attributes[i];
      elementDetails.attributes[attr.name] = attr.value;
    }
    const payload = {
      filenames,
      x,
      y,
      elementDetails
    };
    this[callerSym](FilesDropped, payload);
    cleanupNativeDrag();
  }
  /**
   * Moves the window to the center of the specified screen's work area.
   *
   * @param screenID - The ID of the target screen.
   */
  SetScreen(screenID) {
    return this[callerSym](SetScreenMethod, { screenID });
  }
  /* Triggers Windows 11 Snap Assist feature (Windows only).
   * This is equivalent to pressing Win+Z and shows snap layout options.
   */
  SnapAssist() {
    return this[callerSym](SnapAssistMethod);
  }
  /**
   * Opens the print dialog for the window.
   */
  Print() {
    return this[callerSym](PrintMethod);
  }
};
var Window = _Window;
var thisWindow = new Window("");
function setupDropTargetListeners() {
  const docElement = document.documentElement;
  let dragEnterCounter = 0;
  docElement.addEventListener("dragenter", (event) => {
    var _a2, _b, _c;
    if (!((_a2 = event.dataTransfer) == null ? void 0 : _a2.types.includes("Files"))) {
      return;
    }
    event.preventDefault();
    if (((_c = (_b = window._wails) == null ? void 0 : _b.flags) == null ? void 0 : _c.enableFileDrop) === false) {
      event.dataTransfer.dropEffect = "none";
      return;
    }
    dragEnterCounter++;
    const targetElement = document.elementFromPoint(event.clientX, event.clientY);
    const dropTarget = getDropTargetElement(targetElement);
    if (currentDropTarget && currentDropTarget !== dropTarget) {
      currentDropTarget.classList.remove(DROP_TARGET_ACTIVE_CLASS);
    }
    if (dropTarget) {
      dropTarget.classList.add(DROP_TARGET_ACTIVE_CLASS);
      event.dataTransfer.dropEffect = "copy";
      currentDropTarget = dropTarget;
    } else {
      event.dataTransfer.dropEffect = "none";
      currentDropTarget = null;
    }
  }, false);
  docElement.addEventListener("dragover", (event) => {
    var _a2, _b, _c;
    if (!((_a2 = event.dataTransfer) == null ? void 0 : _a2.types.includes("Files"))) {
      return;
    }
    event.preventDefault();
    if (((_c = (_b = window._wails) == null ? void 0 : _b.flags) == null ? void 0 : _c.enableFileDrop) === false) {
      event.dataTransfer.dropEffect = "none";
      return;
    }
    const targetElement = document.elementFromPoint(event.clientX, event.clientY);
    const dropTarget = getDropTargetElement(targetElement);
    if (currentDropTarget && currentDropTarget !== dropTarget) {
      currentDropTarget.classList.remove(DROP_TARGET_ACTIVE_CLASS);
    }
    if (dropTarget) {
      if (!dropTarget.classList.contains(DROP_TARGET_ACTIVE_CLASS)) {
        dropTarget.classList.add(DROP_TARGET_ACTIVE_CLASS);
      }
      event.dataTransfer.dropEffect = "copy";
      currentDropTarget = dropTarget;
    } else {
      event.dataTransfer.dropEffect = "none";
      currentDropTarget = null;
    }
  }, false);
  docElement.addEventListener("dragleave", (event) => {
    var _a2, _b, _c;
    if (!((_a2 = event.dataTransfer) == null ? void 0 : _a2.types.includes("Files"))) {
      return;
    }
    event.preventDefault();
    if (((_c = (_b = window._wails) == null ? void 0 : _b.flags) == null ? void 0 : _c.enableFileDrop) === false) {
      return;
    }
    if (event.relatedTarget === null) {
      return;
    }
    dragEnterCounter--;
    if (dragEnterCounter === 0 || currentDropTarget && !currentDropTarget.contains(event.relatedTarget)) {
      if (currentDropTarget) {
        currentDropTarget.classList.remove(DROP_TARGET_ACTIVE_CLASS);
        currentDropTarget = null;
      }
      dragEnterCounter = 0;
    }
  }, false);
  docElement.addEventListener("drop", (event) => {
    var _a2, _b, _c;
    if (!((_a2 = event.dataTransfer) == null ? void 0 : _a2.types.includes("Files"))) {
      return;
    }
    event.preventDefault();
    if (((_c = (_b = window._wails) == null ? void 0 : _b.flags) == null ? void 0 : _c.enableFileDrop) === false) {
      return;
    }
    dragEnterCounter = 0;
    if (currentDropTarget) {
      currentDropTarget.classList.remove(DROP_TARGET_ACTIVE_CLASS);
      currentDropTarget = null;
    }
    if (canResolveFilePaths()) {
      const files = [];
      if (event.dataTransfer.items) {
        for (const item of event.dataTransfer.items) {
          if (item.kind === "file") {
            const file = item.getAsFile();
            if (file) files.push(file);
          }
        }
      } else if (event.dataTransfer.files) {
        for (const file of event.dataTransfer.files) {
          files.push(file);
        }
      }
      if (files.length > 0) {
        resolveFilePaths(event.clientX, event.clientY, files);
      }
    }
  }, false);
}
if (typeof window !== "undefined" && typeof document !== "undefined") {
  setupDropTargetListeners();
}
var window_default = thisWindow;

// desktop/@wailsio/runtime/src/wml.ts
function sendEvent(eventName, data = null) {
  Emit(eventName, data);
}
function callWindowMethod(windowName, methodName) {
  const targetWindow = window_default.Get(windowName);
  const method = targetWindow[methodName];
  if (typeof method !== "function") {
    console.error("Window method '".concat(methodName, "' not found"));
    return;
  }
  try {
    method.call(targetWindow);
  } catch (e) {
    console.error("Error calling window method '".concat(methodName, "': "), e);
  }
}
function onWMLTriggered(ev) {
  const element = ev.currentTarget;
  function runEffect(choice = "Yes") {
    if (choice !== "Yes")
      return;
    const eventType = element.getAttribute("wml-event") || element.getAttribute("data-wml-event");
    const targetWindow = element.getAttribute("wml-target-window") || element.getAttribute("data-wml-target-window") || "";
    const windowMethod = element.getAttribute("wml-window") || element.getAttribute("data-wml-window");
    const url = element.getAttribute("wml-openurl") || element.getAttribute("data-wml-openurl");
    if (eventType !== null)
      sendEvent(eventType);
    if (windowMethod !== null)
      callWindowMethod(targetWindow, windowMethod);
    if (url !== null)
      void OpenURL(url);
  }
  const confirm = element.getAttribute("wml-confirm") || element.getAttribute("data-wml-confirm");
  if (confirm) {
    Question({
      Title: "Confirm",
      Message: confirm,
      Detached: false,
      Buttons: [
        { Label: "Yes" },
        { Label: "No", IsDefault: true }
      ]
    }).then(runEffect);
  } else {
    runEffect();
  }
}
var controllerSym = /* @__PURE__ */ Symbol("controller");
var triggerMapSym = /* @__PURE__ */ Symbol("triggerMap");
var elementCountSym = /* @__PURE__ */ Symbol("elementCount");
controllerSym;
var AbortControllerRegistry = class {
  constructor() {
    this[controllerSym] = new AbortController();
  }
  /**
   * Returns an options object for addEventListener that ties the listener
   * to the AbortSignal from the current AbortController.
   *
   * @param element - An HTML element
   * @param triggers - The list of active WML trigger events for the specified elements
   */
  set(element, triggers) {
    return { signal: this[controllerSym].signal };
  }
  /**
   * Removes all registered event listeners and resets the registry.
   */
  reset() {
    this[controllerSym].abort();
    this[controllerSym] = new AbortController();
  }
};
triggerMapSym, elementCountSym;
var WeakMapRegistry = class {
  constructor() {
    this[triggerMapSym] = /* @__PURE__ */ new WeakMap();
    this[elementCountSym] = 0;
  }
  /**
   * Sets active triggers for the specified element.
   *
   * @param element - An HTML element
   * @param triggers - The list of active WML trigger events for the specified element
   */
  set(element, triggers) {
    if (!this[triggerMapSym].has(element)) {
      this[elementCountSym]++;
    }
    this[triggerMapSym].set(element, triggers);
    return {};
  }
  /**
   * Removes all registered event listeners.
   */
  reset() {
    if (this[elementCountSym] <= 0)
      return;
    for (const element of document.body.querySelectorAll("*")) {
      if (this[elementCountSym] <= 0)
        break;
      const triggers = this[triggerMapSym].get(element);
      if (triggers != null) {
        this[elementCountSym]--;
      }
      for (const trigger of triggers || [])
        element.removeEventListener(trigger, onWMLTriggered);
    }
    this[triggerMapSym] = /* @__PURE__ */ new WeakMap();
    this[elementCountSym] = 0;
  }
};
var triggerRegistry = canAbortListeners() ? new AbortControllerRegistry() : new WeakMapRegistry();
function addWMLListeners(element) {
  const triggerRegExp = /\S+/g;
  const triggerAttr = element.getAttribute("wml-trigger") || element.getAttribute("data-wml-trigger") || "click";
  const triggers = [];
  let match;
  while ((match = triggerRegExp.exec(triggerAttr)) !== null)
    triggers.push(match[0]);
  const options = triggerRegistry.set(element, triggers);
  for (const trigger of triggers)
    element.addEventListener(trigger, onWMLTriggered, options);
}
function Enable() {
  whenReady(Reload);
}
function Reload() {
  triggerRegistry.reset();
  document.body.querySelectorAll("[wml-event], [wml-window], [wml-openurl], [data-wml-event], [data-wml-window], [data-wml-openurl]").forEach(addWMLListeners);
}

// desktop/compiled/main.js
window.wails = index_exports;
Enable();
if (true) {
  debugLog("Wails Runtime Loaded");
}

// desktop/@wailsio/runtime/src/system.ts
var system_exports = {};
__export(system_exports, {
  Capabilities: () => Capabilities,
  Environment: () => Environment,
  IsAMD64: () => IsAMD64,
  IsARM: () => IsARM,
  IsARM64: () => IsARM64,
  IsDarkMode: () => IsDarkMode,
  IsDebug: () => IsDebug,
  IsLinux: () => IsLinux,
  IsMac: () => IsMac,
  IsWindows: () => IsWindows,
  invoke: () => invoke
});
var call4 = newRuntimeCaller(objectNames.System);
var SystemIsDarkMode = 0;
var SystemEnvironment = 1;
var SystemCapabilities = 2;
var _invoke = (function() {
  var _a2, _b, _c, _d, _e, _f;
  try {
    if ((_b = (_a2 = window.chrome) == null ? void 0 : _a2.webview) == null ? void 0 : _b.postMessage) {
      return window.chrome.webview.postMessage.bind(window.chrome.webview);
    } else if ((_e = (_d = (_c = window.webkit) == null ? void 0 : _c.messageHandlers) == null ? void 0 : _d["external"]) == null ? void 0 : _e.postMessage) {
      return window.webkit.messageHandlers["external"].postMessage.bind(window.webkit.messageHandlers["external"]);
    } else if ((_f = window.wails) == null ? void 0 : _f.invoke) {
      return (msg) => window.wails.invoke(typeof msg === "string" ? msg : JSON.stringify(msg));
    }
  } catch (e) {
  }
  console.warn(
    "\n%c\u26A0\uFE0F Browser Environment Detected %c\n\n%cOnly UI previews are available in the browser. For full functionality, please run the application in desktop mode.\nMore information at: https://v3.wails.io/learn/build/#using-a-browser-for-development\n",
    "background: #ffffff; color: #000000; font-weight: bold; padding: 4px 8px; border-radius: 4px; border: 2px solid #000000;",
    "background: transparent;",
    "color: #ffffff; font-style: italic; font-weight: bold;"
  );
  return null;
})();
function invoke(msg) {
  _invoke == null ? void 0 : _invoke(msg);
}
function IsDarkMode() {
  return call4(SystemIsDarkMode);
}
async function Capabilities() {
  return call4(SystemCapabilities);
}
function Environment() {
  return call4(SystemEnvironment);
}
function IsWindows() {
  var _a2, _b;
  return ((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.OS) === "windows";
}
function IsLinux() {
  var _a2, _b;
  return ((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.OS) === "linux";
}
function IsMac() {
  var _a2, _b;
  return ((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.OS) === "darwin";
}
function IsAMD64() {
  var _a2, _b;
  return ((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.Arch) === "amd64";
}
function IsARM() {
  var _a2, _b;
  return ((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.Arch) === "arm";
}
function IsARM64() {
  var _a2, _b;
  return ((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.Arch) === "arm64";
}
function IsDebug() {
  var _a2, _b;
  return Boolean((_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.Debug);
}

// desktop/@wailsio/runtime/src/contextmenu.ts
window.addEventListener("contextmenu", contextMenuHandler);
var call5 = newRuntimeCaller(objectNames.ContextMenu);
var ContextMenuOpen = 0;
function openContextMenu(id, x, y, data) {
  void call5(ContextMenuOpen, { id, x, y, data });
}
function contextMenuHandler(event) {
  const target = eventTarget(event);
  const customContextMenu = window.getComputedStyle(target).getPropertyValue("--custom-contextmenu").trim();
  if (customContextMenu) {
    event.preventDefault();
    const data = window.getComputedStyle(target).getPropertyValue("--custom-contextmenu-data");
    openContextMenu(customContextMenu, event.clientX, event.clientY, data);
  } else {
    processDefaultContextMenu(event, target);
  }
}
function processDefaultContextMenu(event, target) {
  if (IsDebug()) {
    return;
  }
  switch (window.getComputedStyle(target).getPropertyValue("--default-contextmenu").trim()) {
    case "show":
      return;
    case "hide":
      event.preventDefault();
      return;
  }
  if (target.isContentEditable) {
    return;
  }
  const selection = window.getSelection();
  const hasSelection = selection && selection.toString().length > 0;
  if (hasSelection) {
    for (let i = 0; i < selection.rangeCount; i++) {
      const range = selection.getRangeAt(i);
      const rects = range.getClientRects();
      for (let j = 0; j < rects.length; j++) {
        const rect = rects[j];
        if (document.elementFromPoint(rect.left, rect.top) === target) {
          return;
        }
      }
    }
  }
  if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement) {
    if (hasSelection || !target.readOnly && !target.disabled) {
      return;
    }
  }
  event.preventDefault();
}

// desktop/@wailsio/runtime/src/flags.ts
var flags_exports = {};
__export(flags_exports, {
  GetFlag: () => GetFlag
});
function GetFlag(key) {
  try {
    return window._wails.flags[key];
  } catch (e) {
    throw new Error("Unable to retrieve flag '" + key + "': " + e, { cause: e });
  }
}

// desktop/@wailsio/runtime/src/drag.ts
var canDrag = false;
var dragging = false;
var resizable = false;
var canResize = false;
var resizing = false;
var resizeEdge = "";
var defaultCursor = "auto";
var buttons = 0;
var buttonsTracked = canTrackButtons();
window._wails = window._wails || {};
window._wails.setResizable = (value) => {
  resizable = value;
  if (!resizable) {
    canResize = resizing = false;
    setResize();
  }
};
var dragInitDone = false;
function isMobile() {
  var _a2, _b;
  const os = (_b = (_a2 = window._wails) == null ? void 0 : _a2.environment) == null ? void 0 : _b.OS;
  if (os === "ios" || os === "android") return true;
  const ua = navigator.userAgent || navigator.vendor || window.opera || "";
  return /android|iphone|ipad|ipod|iemobile|wpdesktop/i.test(ua);
}
function tryInitDragHandlers() {
  if (dragInitDone) return;
  if (isMobile()) return;
  window.addEventListener("mousedown", update, { capture: true });
  window.addEventListener("mousemove", update, { capture: true });
  window.addEventListener("mouseup", update, { capture: true });
  for (const ev of ["click", "contextmenu", "dblclick"]) {
    window.addEventListener(ev, suppressEvent, { capture: true });
  }
  dragInitDone = true;
}
tryInitDragHandlers();
document.addEventListener("DOMContentLoaded", tryInitDragHandlers, { once: true });
var dragEnvPolls = 0;
var dragEnvPoll = window.setInterval(() => {
  if (dragInitDone) {
    window.clearInterval(dragEnvPoll);
    return;
  }
  tryInitDragHandlers();
  if (++dragEnvPolls > 100) {
    window.clearInterval(dragEnvPoll);
  }
}, 50);
function suppressEvent(event) {
  if (dragging || resizing) {
    event.stopImmediatePropagation();
    event.stopPropagation();
    event.preventDefault();
  }
}
var MouseDown = 0;
var MouseUp = 1;
var MouseMove = 2;
function update(event) {
  let eventType, eventButtons = event.buttons;
  switch (event.type) {
    case "mousedown":
      eventType = MouseDown;
      if (!buttonsTracked) {
        eventButtons = buttons | 1 << event.button;
      }
      break;
    case "mouseup":
      eventType = MouseUp;
      if (!buttonsTracked) {
        eventButtons = buttons & ~(1 << event.button);
      }
      break;
    default:
      eventType = MouseMove;
      if (!buttonsTracked) {
        eventButtons = buttons;
      }
      break;
  }
  let released = buttons & ~eventButtons;
  let pressed = eventButtons & ~buttons;
  buttons = eventButtons;
  if (eventType === MouseDown && !(pressed & event.button)) {
    released |= 1 << event.button;
    pressed |= 1 << event.button;
  }
  if (eventType !== MouseMove && resizing || dragging && (eventType === MouseDown || event.button !== 0)) {
    event.stopImmediatePropagation();
    event.stopPropagation();
    event.preventDefault();
  }
  if (released & 1) {
    primaryUp(event);
  }
  if (pressed & 1) {
    primaryDown(event);
  }
  if (eventType === MouseMove) {
    onMouseMove(event);
  }
  ;
}
function primaryDown(event) {
  canDrag = false;
  canResize = false;
  if (!IsWindows()) {
    if (event.type === "mousedown" && event.button === 0 && event.detail !== 1) {
      return;
    }
  }
  if (resizeEdge) {
    canResize = true;
    return;
  }
  const target = eventTarget(event);
  const style = window.getComputedStyle(target);
  canDrag = style.getPropertyValue("--wails-draggable").trim() === "drag" && (event.offsetX - parseFloat(style.paddingLeft) < target.clientWidth && event.offsetY - parseFloat(style.paddingTop) < target.clientHeight);
}
function primaryUp(event) {
  canDrag = false;
  dragging = false;
  canResize = false;
  resizing = false;
}
var cursorForEdge = Object.freeze({
  "se-resize": "nwse-resize",
  "sw-resize": "nesw-resize",
  "nw-resize": "nwse-resize",
  "ne-resize": "nesw-resize",
  "w-resize": "ew-resize",
  "n-resize": "ns-resize",
  "s-resize": "ns-resize",
  "e-resize": "ew-resize"
});
function setResize(edge) {
  if (edge) {
    if (!resizeEdge) {
      defaultCursor = document.body.style.cursor;
    }
    document.body.style.cursor = cursorForEdge[edge];
  } else if (!edge && resizeEdge) {
    document.body.style.cursor = defaultCursor;
  }
  resizeEdge = edge || "";
}
function onMouseMove(event) {
  if (canResize && resizeEdge) {
    resizing = true;
    invoke("wails:resize:" + resizeEdge);
  } else if (canDrag) {
    dragging = true;
    invoke("wails:drag");
  }
  if (dragging || resizing) {
    canDrag = canResize = false;
    return;
  }
  if (!resizable || !IsWindows()) {
    if (resizeEdge) {
      setResize();
    }
    return;
  }
  const resizeHandleHeight = GetFlag("system.resizeHandleHeight") || 5;
  const resizeHandleWidth = GetFlag("system.resizeHandleWidth") || 5;
  const cornerExtra = GetFlag("resizeCornerExtra") || 10;
  const rightBorder = window.outerWidth - event.clientX < resizeHandleWidth;
  const leftBorder = event.clientX < resizeHandleWidth;
  const topBorder = event.clientY < resizeHandleHeight;
  const bottomBorder = window.outerHeight - event.clientY < resizeHandleHeight;
  const rightCorner = window.outerWidth - event.clientX < resizeHandleWidth + cornerExtra;
  const leftCorner = event.clientX < resizeHandleWidth + cornerExtra;
  const topCorner = event.clientY < resizeHandleHeight + cornerExtra;
  const bottomCorner = window.outerHeight - event.clientY < resizeHandleHeight + cornerExtra;
  if (!leftCorner && !topCorner && !bottomCorner && !rightCorner) {
    setResize();
  } else if (rightCorner && bottomCorner) setResize("se-resize");
  else if (leftCorner && bottomCorner) setResize("sw-resize");
  else if (leftCorner && topCorner) setResize("nw-resize");
  else if (topCorner && rightCorner) setResize("ne-resize");
  else if (leftBorder) setResize("w-resize");
  else if (topBorder) setResize("n-resize");
  else if (bottomBorder) setResize("s-resize");
  else if (rightBorder) setResize("e-resize");
  else setResize();
}

// desktop/@wailsio/runtime/src/application.ts
var application_exports = {};
__export(application_exports, {
  Hide: () => Hide,
  Quit: () => Quit,
  Show: () => Show
});
var call6 = newRuntimeCaller(objectNames.Application);
var HideMethod2 = 0;
var ShowMethod2 = 1;
var QuitMethod = 2;
function Hide() {
  return call6(HideMethod2);
}
function Show() {
  return call6(ShowMethod2);
}
function Quit() {
  return call6(QuitMethod);
}

// desktop/@wailsio/runtime/src/calls.ts
var calls_exports = {};
__export(calls_exports, {
  ByID: () => ByID,
  ByName: () => ByName,
  Call: () => Call,
  RuntimeError: () => RuntimeError
});

// desktop/@wailsio/runtime/src/callable.ts
var fnToStr = Function.prototype.toString;
var reflectApply = typeof Reflect === "object" && Reflect !== null && Reflect.apply;
var badArrayLike;
var isCallableMarker;
if (typeof reflectApply === "function" && typeof Object.defineProperty === "function") {
  try {
    badArrayLike = Object.defineProperty({}, "length", {
      get: function() {
        throw isCallableMarker;
      }
    });
    isCallableMarker = {};
    reflectApply(function() {
      throw 42;
    }, null, badArrayLike);
  } catch (_) {
    if (_ !== isCallableMarker) {
      reflectApply = null;
    }
  }
} else {
  reflectApply = null;
}
var constructorRegex = /^\s*class\b/;
var isES6ClassFn = function isES6ClassFunction(value) {
  try {
    var fnStr = fnToStr.call(value);
    return constructorRegex.test(fnStr);
  } catch (e) {
    return false;
  }
};
var tryFunctionObject = function tryFunctionToStr(value) {
  try {
    if (isES6ClassFn(value)) {
      return false;
    }
    fnToStr.call(value);
    return true;
  } catch (e) {
    return false;
  }
};
var toStr = Object.prototype.toString;
var objectClass = "[object Object]";
var fnClass = "[object Function]";
var genClass = "[object GeneratorFunction]";
var ddaClass = "[object HTMLAllCollection]";
var ddaClass2 = "[object HTML document.all class]";
var ddaClass3 = "[object HTMLCollection]";
var hasToStringTag = typeof Symbol === "function" && !!Symbol.toStringTag;
var isIE68 = !(0 in [,]);
var isDDA = function isDocumentDotAll() {
  return false;
};
if (typeof document === "object") {
  all = document.all;
  if (toStr.call(all) === toStr.call(document.all)) {
    isDDA = function isDocumentDotAll2(value) {
      if ((isIE68 || !value) && (typeof value === "undefined" || typeof value === "object")) {
        try {
          var str = toStr.call(value);
          return (str === ddaClass || str === ddaClass2 || str === ddaClass3 || str === objectClass) && value("") == null;
        } catch (e) {
        }
      }
      return false;
    };
  }
}
var all;
function isCallableRefApply(value) {
  if (isDDA(value)) {
    return true;
  }
  if (!value) {
    return false;
  }
  if (typeof value !== "function" && typeof value !== "object") {
    return false;
  }
  try {
    reflectApply(value, null, badArrayLike);
  } catch (e) {
    if (e !== isCallableMarker) {
      return false;
    }
  }
  return !isES6ClassFn(value) && tryFunctionObject(value);
}
function isCallableNoRefApply(value) {
  if (isDDA(value)) {
    return true;
  }
  if (!value) {
    return false;
  }
  if (typeof value !== "function" && typeof value !== "object") {
    return false;
  }
  if (hasToStringTag) {
    return tryFunctionObject(value);
  }
  if (isES6ClassFn(value)) {
    return false;
  }
  var strClass = toStr.call(value);
  if (strClass !== fnClass && strClass !== genClass && !/^\[object HTML/.test(strClass)) {
    return false;
  }
  return tryFunctionObject(value);
}
var callable_default = reflectApply ? isCallableRefApply : isCallableNoRefApply;

// desktop/@wailsio/runtime/src/cancellable.ts
var CancelError = class extends Error {
  /**
   * Constructs a new `CancelError` instance.
   * @param message - The error message.
   * @param options - Options to be forwarded to the Error constructor.
   */
  constructor(message, options) {
    super(message, options);
    this.name = "CancelError";
  }
};
var CancelledRejectionError = class extends Error {
  /**
   * Constructs a new `CancelledRejectionError` instance.
   * @param promise - The promise that caused the error originally.
   * @param reason - The rejection reason.
   * @param info - An optional informative message specifying the circumstances in which the error was thrown.
   *               Defaults to the string `"Unhandled rejection in cancelled promise."`.
   */
  constructor(promise, reason, info) {
    super((info != null ? info : "Unhandled rejection in cancelled promise.") + " Reason: " + errorMessage(reason), { cause: reason });
    this.promise = promise;
    this.name = "CancelledRejectionError";
  }
};
var barrierSym = /* @__PURE__ */ Symbol("barrier");
var cancelImplSym = /* @__PURE__ */ Symbol("cancelImpl");
var _a;
var species = (_a = Symbol.species) != null ? _a : /* @__PURE__ */ Symbol("speciesPolyfill");
var CancellablePromise = class _CancellablePromise extends Promise {
  /**
   * Creates a new `CancellablePromise`.
   *
   * @param executor - A callback used to initialize the promise. This callback is passed two arguments:
   *                   a `resolve` callback used to resolve the promise with a value
   *                   or the result of another promise (possibly cancellable),
   *                   and a `reject` callback used to reject the promise with a provided reason or error.
   *                   If the value provided to the `resolve` callback is a thenable _and_ cancellable object
   *                   (it has a `then` _and_ a `cancel` method),
   *                   cancellation requests will be forwarded to that object and the oncancelled will not be invoked anymore.
   *                   If any one of the two callbacks is called _after_ the promise has been cancelled,
   *                   the provided values will be cancelled and resolved as usual,
   *                   but their results will be discarded.
   *                   However, if the resolution process ultimately ends up in a rejection
   *                   that is not due to cancellation, the rejection reason
   *                   will be wrapped in a {@link CancelledRejectionError}
   *                   and bubbled up as an unhandled rejection.
   * @param oncancelled - It is the caller's responsibility to ensure that any operation
   *                      started by the executor is properly halted upon cancellation.
   *                      This optional callback can be used to that purpose.
   *                      It will be called _synchronously_ with a cancellation cause
   *                      when cancellation is requested, _after_ the promise has already rejected
   *                      with a {@link CancelError}, but _before_
   *                      any {@link then}/{@link catch}/{@link finally} callback runs.
   *                      If the callback returns a thenable, the promise returned from {@link cancel}
   *                      will only fulfill after the former has settled.
   *                      Unhandled exceptions or rejections from the callback will be wrapped
   *                      in a {@link CancelledRejectionError} and bubbled up as unhandled rejections.
   *                      If the `resolve` callback is called before cancellation with a cancellable promise,
   *                      cancellation requests on this promise will be diverted to that promise,
   *                      and the original `oncancelled` callback will be discarded.
   */
  constructor(executor, oncancelled) {
    let resolve;
    let reject;
    super((res, rej) => {
      resolve = res;
      reject = rej;
    });
    if (this.constructor[species] !== Promise) {
      throw new TypeError("CancellablePromise does not support transparent subclassing. Please refrain from overriding the [Symbol.species] static property.");
    }
    let promise = {
      promise: this,
      resolve,
      reject,
      get oncancelled() {
        return oncancelled != null ? oncancelled : null;
      },
      set oncancelled(cb) {
        oncancelled = cb != null ? cb : void 0;
      }
    };
    const state = {
      get root() {
        return state;
      },
      resolving: false,
      settled: false
    };
    void Object.defineProperties(this, {
      [barrierSym]: {
        configurable: false,
        enumerable: false,
        writable: true,
        value: null
      },
      [cancelImplSym]: {
        configurable: false,
        enumerable: false,
        writable: false,
        value: cancellerFor(promise, state)
      }
    });
    const rejector = rejectorFor(promise, state);
    try {
      executor(resolverFor(promise, state), rejector);
    } catch (err) {
      if (state.resolving) {
        console.log("Unhandled exception in CancellablePromise executor.", err);
      } else {
        rejector(err);
      }
    }
  }
  /**
   * Cancels immediately the execution of the operation associated with this promise.
   * The promise rejects with a {@link CancelError} instance as reason,
   * with the {@link CancelError#cause} property set to the given argument, if any.
   *
   * Has no effect if called after the promise has already settled;
   * repeated calls in particular are safe, but only the first one
   * will set the cancellation cause.
   *
   * The `CancelError` exception _need not_ be handled explicitly _on the promises that are being cancelled:_
   * cancelling a promise with no attached rejection handler does not trigger an unhandled rejection event.
   * Therefore, the following idioms are all equally correct:
   * ```ts
   * new CancellablePromise((resolve, reject) => { ... }).cancel();
   * new CancellablePromise((resolve, reject) => { ... }).then(...).cancel();
   * new CancellablePromise((resolve, reject) => { ... }).then(...).catch(...).cancel();
   * ```
   * Whenever some cancelled promise in a chain rejects with a `CancelError`
   * with the same cancellation cause as itself, the error will be discarded silently.
   * However, the `CancelError` _will still be delivered_ to all attached rejection handlers
   * added by {@link then} and related methods:
   * ```ts
   * let cancellable = new CancellablePromise((resolve, reject) => { ... });
   * cancellable.then(() => { ... }).catch(console.log);
   * cancellable.cancel(); // A CancelError is printed to the console.
   * ```
   * If the `CancelError` is not handled downstream by the time it reaches
   * a _non-cancelled_ promise, it _will_ trigger an unhandled rejection event,
   * just like normal rejections would:
   * ```ts
   * let cancellable = new CancellablePromise((resolve, reject) => { ... });
   * let chained = cancellable.then(() => { ... }).then(() => { ... }); // No catch...
   * cancellable.cancel(); // Unhandled rejection event on chained!
   * ```
   * Therefore, it is important to either cancel whole promise chains from their tail,
   * as shown in the correct idioms above, or take care of handling errors everywhere.
   *
   * @returns A cancellable promise that _fulfills_ after the cancel callback (if any)
   * and all handlers attached up to the call to cancel have run.
   * If the cancel callback returns a thenable, the promise returned by `cancel`
   * will also wait for that thenable to settle.
   * This enables callers to wait for the cancelled operation to terminate
   * without being forced to handle potential errors at the call site.
   * ```ts
   * cancellable.cancel().then(() => {
   *     // Cleanup finished, it's safe to do something else.
   * }, (err) => {
   *     // Unreachable: the promise returned from cancel will never reject.
   * });
   * ```
   * Note that the returned promise will _not_ handle implicitly any rejection
   * that might have occurred already in the cancelled chain.
   * It will just track whether registered handlers have been executed or not.
   * Therefore, unhandled rejections will never be silently handled by calling cancel.
   */
  cancel(cause) {
    return new _CancellablePromise((resolve) => {
      Promise.all([
        this[cancelImplSym](new CancelError("Promise cancelled.", { cause })),
        currentBarrier(this)
      ]).then(() => resolve(), () => resolve());
    });
  }
  /**
   * Binds promise cancellation to the abort event of the given {@link AbortSignal}.
   * If the signal has already aborted, the promise will be cancelled immediately.
   * When either condition is verified, the cancellation cause will be set
   * to the signal's abort reason (see {@link AbortSignal#reason}).
   *
   * Has no effect if called (or if the signal aborts) _after_ the promise has already settled.
   * Only the first signal to abort will set the cancellation cause.
   *
   * For more details about the cancellation process,
   * see {@link cancel} and the `CancellablePromise` constructor.
   *
   * This method enables `await`ing cancellable promises without having
   * to store them for future cancellation, e.g.:
   * ```ts
   * await longRunningOperation().cancelOn(signal);
   * ```
   * instead of:
   * ```ts
   * let promiseToBeCancelled = longRunningOperation();
   * await promiseToBeCancelled;
   * ```
   *
   * @returns This promise, for method chaining.
   */
  cancelOn(signal) {
    if (signal.aborted) {
      void this.cancel(signal.reason);
    } else {
      signal.addEventListener("abort", () => void this.cancel(signal.reason), { capture: true });
    }
    return this;
  }
  /**
   * Attaches callbacks for the resolution and/or rejection of the `CancellablePromise`.
   *
   * The optional `oncancelled` argument will be invoked when the returned promise is cancelled,
   * with the same semantics as the `oncancelled` argument of the constructor.
   * When the parent promise rejects or is cancelled, the `onrejected` callback will run,
   * _even after the returned promise has been cancelled:_
   * in that case, should it reject or throw, the reason will be wrapped
   * in a {@link CancelledRejectionError} and bubbled up as an unhandled rejection.
   *
   * @param onfulfilled The callback to execute when the Promise is resolved.
   * @param onrejected The callback to execute when the Promise is rejected.
   * @returns A `CancellablePromise` for the completion of whichever callback is executed.
   * The returned promise is hooked up to propagate cancellation requests up the chain, but not down:
   *
   *   - if the parent promise is cancelled, the `onrejected` handler will be invoked with a `CancelError`
   *     and the returned promise _will resolve regularly_ with its result;
   *   - conversely, if the returned promise is cancelled, _the parent promise is cancelled too;_
   *     the `onrejected` handler will still be invoked with the parent's `CancelError`,
   *     but its result will be discarded
   *     and the returned promise will reject with a `CancelError` as well.
   *
   * The promise returned from {@link cancel} will fulfill only after all attached handlers
   * up the entire promise chain have been run.
   *
   * If either callback returns a cancellable promise,
   * cancellation requests will be diverted to it,
   * and the specified `oncancelled` callback will be discarded.
   */
  then(onfulfilled, onrejected, oncancelled) {
    if (!(this instanceof _CancellablePromise)) {
      throw new TypeError("CancellablePromise.prototype.then called on an invalid object.");
    }
    if (!callable_default(onfulfilled)) {
      onfulfilled = identity;
    }
    if (!callable_default(onrejected)) {
      onrejected = thrower;
    }
    if (onfulfilled === identity && onrejected == thrower) {
      return new _CancellablePromise((resolve) => resolve(this));
    }
    const barrier = {};
    this[barrierSym] = barrier;
    return new _CancellablePromise((resolve, reject) => {
      void super.then(
        (value) => {
          var _a2;
          if (this[barrierSym] === barrier) {
            this[barrierSym] = null;
          }
          (_a2 = barrier.resolve) == null ? void 0 : _a2.call(barrier);
          try {
            resolve(onfulfilled(value));
          } catch (err) {
            reject(err);
          }
        },
        (reason) => {
          var _a2;
          if (this[barrierSym] === barrier) {
            this[barrierSym] = null;
          }
          (_a2 = barrier.resolve) == null ? void 0 : _a2.call(barrier);
          try {
            resolve(onrejected(reason));
          } catch (err) {
            reject(err);
          }
        }
      );
    }, async (cause) => {
      try {
        return oncancelled == null ? void 0 : oncancelled(cause);
      } finally {
        await this.cancel(cause);
      }
    });
  }
  /**
   * Attaches a callback for only the rejection of the Promise.
   *
   * The optional `oncancelled` argument will be invoked when the returned promise is cancelled,
   * with the same semantics as the `oncancelled` argument of the constructor.
   * When the parent promise rejects or is cancelled, the `onrejected` callback will run,
   * _even after the returned promise has been cancelled:_
   * in that case, should it reject or throw, the reason will be wrapped
   * in a {@link CancelledRejectionError} and bubbled up as an unhandled rejection.
   *
   * It is equivalent to
   * ```ts
   * cancellablePromise.then(undefined, onrejected, oncancelled);
   * ```
   * and the same caveats apply.
   *
   * @returns A Promise for the completion of the callback.
   * Cancellation requests on the returned promise
   * will propagate up the chain to the parent promise,
   * but not in the other direction.
   *
   * The promise returned from {@link cancel} will fulfill only after all attached handlers
   * up the entire promise chain have been run.
   *
   * If `onrejected` returns a cancellable promise,
   * cancellation requests will be diverted to it,
   * and the specified `oncancelled` callback will be discarded.
   * See {@link then} for more details.
   */
  catch(onrejected, oncancelled) {
    return this.then(void 0, onrejected, oncancelled);
  }
  /**
   * Attaches a callback that is invoked when the CancellablePromise is settled (fulfilled or rejected). The
   * resolved value cannot be accessed or modified from the callback.
   * The returned promise will settle in the same state as the original one
   * after the provided callback has completed execution,
   * unless the callback throws or returns a rejecting promise,
   * in which case the returned promise will reject as well.
   *
   * The optional `oncancelled` argument will be invoked when the returned promise is cancelled,
   * with the same semantics as the `oncancelled` argument of the constructor.
   * Once the parent promise settles, the `onfinally` callback will run,
   * _even after the returned promise has been cancelled:_
   * in that case, should it reject or throw, the reason will be wrapped
   * in a {@link CancelledRejectionError} and bubbled up as an unhandled rejection.
   *
   * This method is implemented in terms of {@link then} and the same caveats apply.
   * It is polyfilled, hence available in every OS/webview version.
   *
   * @returns A Promise for the completion of the callback.
   * Cancellation requests on the returned promise
   * will propagate up the chain to the parent promise,
   * but not in the other direction.
   *
   * The promise returned from {@link cancel} will fulfill only after all attached handlers
   * up the entire promise chain have been run.
   *
   * If `onfinally` returns a cancellable promise,
   * cancellation requests will be diverted to it,
   * and the specified `oncancelled` callback will be discarded.
   * See {@link then} for more details.
   */
  finally(onfinally, oncancelled) {
    if (!(this instanceof _CancellablePromise)) {
      throw new TypeError("CancellablePromise.prototype.finally called on an invalid object.");
    }
    if (!callable_default(onfinally)) {
      return this.then(onfinally, onfinally, oncancelled);
    }
    return this.then(
      (value) => _CancellablePromise.resolve(onfinally()).then(() => value),
      (reason) => _CancellablePromise.resolve(onfinally()).then(() => {
        throw reason;
      }),
      oncancelled
    );
  }
  /**
   * We use the `[Symbol.species]` static property, if available,
   * to disable the built-in automatic subclassing features from {@link Promise}.
   * It is critical for performance reasons that extenders do not override this.
   * Once the proposal at https://github.com/tc39/proposal-rm-builtin-subclassing
   * is either accepted or retired, this implementation will have to be revised accordingly.
   *
   * @ignore
   * @internal
   */
  static get [(barrierSym, cancelImplSym, species)]() {
    return Promise;
  }
  static all(values) {
    let collected = Array.from(values);
    const promise = collected.length === 0 ? _CancellablePromise.resolve(collected) : new _CancellablePromise((resolve, reject) => {
      void Promise.all(collected).then(resolve, reject);
    }, (cause) => cancelAll(promise, collected, cause));
    return promise;
  }
  static allSettled(values) {
    let collected = Array.from(values);
    const promise = collected.length === 0 ? _CancellablePromise.resolve(collected) : new _CancellablePromise((resolve, reject) => {
      void Promise.allSettled(collected).then(resolve, reject);
    }, (cause) => cancelAll(promise, collected, cause));
    return promise;
  }
  static any(values) {
    let collected = Array.from(values);
    const promise = collected.length === 0 ? _CancellablePromise.resolve(collected) : new _CancellablePromise((resolve, reject) => {
      void Promise.any(collected).then(resolve, reject);
    }, (cause) => cancelAll(promise, collected, cause));
    return promise;
  }
  static race(values) {
    let collected = Array.from(values);
    const promise = new _CancellablePromise((resolve, reject) => {
      void Promise.race(collected).then(resolve, reject);
    }, (cause) => cancelAll(promise, collected, cause));
    return promise;
  }
  /**
   * Creates a new cancelled CancellablePromise for the provided cause.
   *
   * @group Static Methods
   */
  static cancel(cause) {
    const p = new _CancellablePromise(() => {
    });
    p.cancel(cause);
    return p;
  }
  /**
   * Creates a new CancellablePromise that cancels
   * after the specified timeout, with the provided cause.
   *
   * If the {@link AbortSignal.timeout} factory method is available,
   * it is used to base the timeout on _active_ time rather than _elapsed_ time.
   * Otherwise, `timeout` falls back to {@link setTimeout}.
   *
   * @group Static Methods
   */
  static timeout(milliseconds, cause) {
    const promise = new _CancellablePromise(() => {
    });
    if (AbortSignal && typeof AbortSignal === "function" && AbortSignal.timeout && typeof AbortSignal.timeout === "function") {
      AbortSignal.timeout(milliseconds).addEventListener("abort", () => void promise.cancel(cause));
    } else {
      setTimeout(() => void promise.cancel(cause), milliseconds);
    }
    return promise;
  }
  static sleep(milliseconds, value) {
    return new _CancellablePromise((resolve) => {
      setTimeout(() => resolve(value), milliseconds);
    });
  }
  /**
   * Creates a new rejected CancellablePromise for the provided reason.
   *
   * @group Static Methods
   */
  static reject(reason) {
    return new _CancellablePromise((_, reject) => reject(reason));
  }
  static resolve(value) {
    if (value instanceof _CancellablePromise) {
      return value;
    }
    return new _CancellablePromise((resolve) => resolve(value));
  }
  /**
   * Creates a new CancellablePromise and returns it in an object, along with its resolve and reject functions
   * and a getter/setter for the cancellation callback.
   *
   * This method is polyfilled, hence available in every OS/webview version.
   *
   * @group Static Methods
   */
  static withResolvers() {
    let result = { oncancelled: null };
    result.promise = new _CancellablePromise((resolve, reject) => {
      result.resolve = resolve;
      result.reject = reject;
    }, (cause) => {
      var _a2;
      (_a2 = result.oncancelled) == null ? void 0 : _a2.call(result, cause);
    });
    return result;
  }
};
function cancellerFor(promise, state) {
  let cancellationPromise = void 0;
  return (reason) => {
    if (!state.settled) {
      state.settled = true;
      state.reason = reason;
      promise.reject(reason);
      void Promise.prototype.then.call(promise.promise, void 0, (err) => {
        if (err !== reason) {
          throw err;
        }
      });
    }
    if (!state.reason || !promise.oncancelled) {
      return;
    }
    cancellationPromise = new Promise((resolve) => {
      try {
        resolve(promise.oncancelled(state.reason.cause));
      } catch (err) {
        Promise.reject(new CancelledRejectionError(promise.promise, err, "Unhandled exception in oncancelled callback."));
      }
    }).catch((reason2) => {
      Promise.reject(new CancelledRejectionError(promise.promise, reason2, "Unhandled rejection in oncancelled callback."));
    });
    promise.oncancelled = null;
    return cancellationPromise;
  };
}
function resolverFor(promise, state) {
  return (value) => {
    if (state.resolving) {
      return;
    }
    state.resolving = true;
    if (value === promise.promise) {
      if (state.settled) {
        return;
      }
      state.settled = true;
      promise.reject(new TypeError("A promise cannot be resolved with itself."));
      return;
    }
    if (value != null && (typeof value === "object" || typeof value === "function")) {
      let then;
      try {
        then = value.then;
      } catch (err) {
        state.settled = true;
        promise.reject(err);
        return;
      }
      if (callable_default(then)) {
        try {
          let cancel = value.cancel;
          if (callable_default(cancel)) {
            const oncancelled = (cause) => {
              Reflect.apply(cancel, value, [cause]);
            };
            if (state.reason) {
              void cancellerFor(__spreadProps(__spreadValues({}, promise), { oncancelled }), state)(state.reason);
            } else {
              promise.oncancelled = oncancelled;
            }
          }
        } catch (e) {
        }
        const newState = {
          root: state.root,
          resolving: false,
          get settled() {
            return this.root.settled;
          },
          set settled(value2) {
            this.root.settled = value2;
          },
          get reason() {
            return this.root.reason;
          }
        };
        const rejector = rejectorFor(promise, newState);
        try {
          Reflect.apply(then, value, [resolverFor(promise, newState), rejector]);
        } catch (err) {
          rejector(err);
        }
        return;
      }
    }
    if (state.settled) {
      return;
    }
    state.settled = true;
    promise.resolve(value);
  };
}
function rejectorFor(promise, state) {
  return (reason) => {
    if (state.resolving) {
      return;
    }
    state.resolving = true;
    if (state.settled) {
      try {
        if (reason instanceof CancelError && state.reason instanceof CancelError && Object.is(reason.cause, state.reason.cause)) {
          return;
        }
      } catch (e) {
      }
      void Promise.reject(new CancelledRejectionError(promise.promise, reason));
    } else {
      state.settled = true;
      promise.reject(reason);
    }
  };
}
function cancelAll(parent, values, cause) {
  const results = [];
  for (const value of values) {
    let cancel;
    try {
      if (!callable_default(value.then)) {
        continue;
      }
      cancel = value.cancel;
      if (!callable_default(cancel)) {
        continue;
      }
    } catch (e) {
      continue;
    }
    let result;
    try {
      result = Reflect.apply(cancel, value, [cause]);
    } catch (err) {
      Promise.reject(new CancelledRejectionError(parent, err, "Unhandled exception in cancel method."));
      continue;
    }
    if (!result) {
      continue;
    }
    results.push(
      (result instanceof Promise ? result : Promise.resolve(result)).catch((reason) => {
        Promise.reject(new CancelledRejectionError(parent, reason, "Unhandled rejection in cancel method."));
      })
    );
  }
  return Promise.all(results);
}
function identity(x) {
  return x;
}
function thrower(reason) {
  throw reason;
}
function errorMessage(err) {
  try {
    if (err instanceof Error || typeof err !== "object" || err.toString !== Object.prototype.toString) {
      return "" + err;
    }
  } catch (e) {
  }
  try {
    return JSON.stringify(err);
  } catch (e) {
  }
  try {
    return Object.prototype.toString.call(err);
  } catch (e) {
  }
  return "<could not convert error to string>";
}
function currentBarrier(promise) {
  var _a2;
  let pwr = (_a2 = promise[barrierSym]) != null ? _a2 : {};
  if (!("promise" in pwr)) {
    Object.assign(pwr, promiseWithResolvers());
  }
  if (promise[barrierSym] == null) {
    pwr.resolve();
    promise[barrierSym] = pwr;
  }
  return pwr.promise;
}
var promiseWithResolvers = Promise.withResolvers;
if (promiseWithResolvers && typeof promiseWithResolvers === "function") {
  promiseWithResolvers = promiseWithResolvers.bind(Promise);
} else {
  promiseWithResolvers = function() {
    let resolve;
    let reject;
    const promise = new Promise((res, rej) => {
      resolve = res;
      reject = rej;
    });
    return { promise, resolve, reject };
  };
}

// desktop/@wailsio/runtime/src/calls.ts
window._wails = window._wails || {};
var call7 = newRuntimeCaller(objectNames.Call);
var cancelCall = newRuntimeCaller(objectNames.CancelCall);
var callResponses = /* @__PURE__ */ new Map();
var CallBinding = 0;
var CancelMethod = 0;
var RuntimeError = class extends Error {
  /**
   * Constructs a new RuntimeError instance.
   * @param message - The error message.
   * @param options - Options to be forwarded to the Error constructor.
   */
  constructor(message, options) {
    super(message, options);
    this.name = "RuntimeError";
  }
};
function generateID() {
  let result;
  do {
    result = nanoid();
  } while (callResponses.has(result));
  return result;
}
function Call(options) {
  const id = generateID();
  const result = CancellablePromise.withResolvers();
  callResponses.set(id, { resolve: result.resolve, reject: result.reject });
  const request = call7(CallBinding, Object.assign({ "call-id": id }, options));
  let running = true;
  request.then((res) => {
    running = false;
    callResponses.delete(id);
    result.resolve(res);
  }, (err) => {
    running = false;
    callResponses.delete(id);
    result.reject(err);
  });
  const cancel = () => {
    callResponses.delete(id);
    return cancelCall(CancelMethod, { "call-id": id }).catch((err) => {
      console.error("Error while requesting binding call cancellation:", err);
    });
  };
  result.oncancelled = () => {
    if (running) {
      return cancel();
    } else {
      return request.then(cancel);
    }
  };
  return result.promise;
}
function ByName(methodName, ...args) {
  return Call({ methodName, args });
}
function ByID(methodID, ...args) {
  return Call({ methodID, args });
}

// desktop/@wailsio/runtime/src/clipboard.ts
var clipboard_exports = {};
__export(clipboard_exports, {
  SetText: () => SetText,
  Text: () => Text
});
var call8 = newRuntimeCaller(objectNames.Clipboard);
var ClipboardSetText = 0;
var ClipboardText = 1;
function SetText(text) {
  return call8(ClipboardSetText, { text });
}
function Text() {
  return call8(ClipboardText);
}

// desktop/@wailsio/runtime/src/screens.ts
var screens_exports = {};
__export(screens_exports, {
  GetAll: () => GetAll,
  GetByID: () => GetByID,
  GetByIndex: () => GetByIndex,
  GetCurrent: () => GetCurrent,
  GetPrimary: () => GetPrimary
});
var call9 = newRuntimeCaller(objectNames.Screens);
var getAll = 0;
var getPrimary = 1;
var getCurrent = 2;
var getByID = 3;
var getByIndex = 4;
function GetAll() {
  return call9(getAll);
}
function GetPrimary() {
  return call9(getPrimary);
}
function GetCurrent() {
  return call9(getCurrent);
}
function GetByID(id) {
  return call9(getByID, { id });
}
function GetByIndex(index) {
  return call9(getByIndex, { index });
}

// desktop/@wailsio/runtime/src/ios.ts
var ios_exports = {};
__export(ios_exports, {
  Device: () => Device,
  Haptics: () => Haptics
});
var call10 = newRuntimeCaller(objectNames.IOS);
var HapticsImpact = 0;
var DeviceInfo = 1;
var Haptics;
((Haptics2) => {
  function Impact(style = "medium") {
    return call10(HapticsImpact, { style });
  }
  Haptics2.Impact = Impact;
})(Haptics || (Haptics = {}));
var Device;
((Device2) => {
  function Info2() {
    return call10(DeviceInfo);
  }
  Device2.Info = Info2;
})(Device || (Device = {}));

// desktop/@wailsio/runtime/src/index.ts
window._wails = window._wails || {};
window._wails.invoke = invoke;
window._wails.clientId = clientId;
window._wails.handlePlatformFileDrop = window_default.HandlePlatformFileDrop.bind(window_default);
window._wails.handleDragEnter = handleDragEnter;
window._wails.handleDragLeave = handleDragLeave;
window._wails.handleDragOver = handleDragOver;
invoke("wails:runtime:ready");
function loadOptionalScript(url) {
  return fetch(url, { method: "HEAD" }).then((response) => {
    if (response.ok) {
      const contentType = (response.headers.get("content-type") || "").toLowerCase();
      if (contentType.includes("javascript")) {
        const script = document.createElement("script");
        script.src = url;
        document.head.appendChild(script);
      }
    }
  }).catch(() => {
  });
}
loadOptionalScript("/wails/custom.js");
export {
  application_exports as Application,
  browser_exports as Browser,
  calls_exports as Call,
  CancelError,
  CancellablePromise,
  CancelledRejectionError,
  clipboard_exports as Clipboard,
  create_exports as Create,
  dialogs_exports as Dialogs,
  events_exports as Events,
  flags_exports as Flags,
  ios_exports as IOS,
  screens_exports as Screens,
  system_exports as System,
  wml_exports as WML,
  window_default as Window,
  clientId,
  getTransport,
  loadOptionalScript,
  objectNames,
  setTransport
};
//# sourceMappingURL=data:application/json;base64,ewogICJ2ZXJzaW9uIjogMywKICAic291cmNlcyI6IFsiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL2luZGV4LnRzIiwgIi4uLy4uL3J1bnRpbWUvZGVza3RvcC9Ad2FpbHNpby9ydW50aW1lL3NyYy93bWwudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL2Jyb3dzZXIudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL25hbm9pZC50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvcnVudGltZS50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvZGlhbG9ncy50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvZXZlbnRzLnRzIiwgIi4uLy4uL3J1bnRpbWUvZGVza3RvcC9Ad2FpbHNpby9ydW50aW1lL3NyYy9saXN0ZW5lci50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvY3JlYXRlLnRzIiwgIi4uLy4uL3J1bnRpbWUvZGVza3RvcC9Ad2FpbHNpby9ydW50aW1lL3NyYy9ldmVudF90eXBlcy50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvdXRpbHMudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL3dpbmRvdy50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvY29tcGlsZWQvbWFpbi5qcyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvc3lzdGVtLnRzIiwgIi4uLy4uL3J1bnRpbWUvZGVza3RvcC9Ad2FpbHNpby9ydW50aW1lL3NyYy9jb250ZXh0bWVudS50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvZmxhZ3MudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL2RyYWcudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL2FwcGxpY2F0aW9uLnRzIiwgIi4uLy4uL3J1bnRpbWUvZGVza3RvcC9Ad2FpbHNpby9ydW50aW1lL3NyYy9jYWxscy50cyIsICIuLi8uLi9ydW50aW1lL2Rlc2t0b3AvQHdhaWxzaW8vcnVudGltZS9zcmMvY2FsbGFibGUudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL2NhbmNlbGxhYmxlLnRzIiwgIi4uLy4uL3J1bnRpbWUvZGVza3RvcC9Ad2FpbHNpby9ydW50aW1lL3NyYy9jbGlwYm9hcmQudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL3NjcmVlbnMudHMiLCAiLi4vLi4vcnVudGltZS9kZXNrdG9wL0B3YWlsc2lvL3J1bnRpbWUvc3JjL2lvcy50cyJdLAogICJzb3VyY2VzQ29udGVudCI6IFsiLypcclxuIF9cdCAgIF9fXHQgIF8gX19cclxufCB8XHQgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuLy8gU2V0dXBcclxud2luZG93Ll93YWlscyA9IHdpbmRvdy5fd2FpbHMgfHwge307XHJcblxyXG5pbXBvcnQgXCIuL2NvbnRleHRtZW51LmpzXCI7XHJcbmltcG9ydCBcIi4vZHJhZy5qc1wiO1xyXG5cclxuLy8gUmUtZXhwb3J0IHB1YmxpYyBBUElcclxuaW1wb3J0ICogYXMgQXBwbGljYXRpb24gZnJvbSBcIi4vYXBwbGljYXRpb24uanNcIjtcclxuaW1wb3J0ICogYXMgQnJvd3NlciBmcm9tIFwiLi9icm93c2VyLmpzXCI7XHJcbmltcG9ydCAqIGFzIENhbGwgZnJvbSBcIi4vY2FsbHMuanNcIjtcclxuaW1wb3J0ICogYXMgQ2xpcGJvYXJkIGZyb20gXCIuL2NsaXBib2FyZC5qc1wiO1xyXG5pbXBvcnQgKiBhcyBDcmVhdGUgZnJvbSBcIi4vY3JlYXRlLmpzXCI7XHJcbmltcG9ydCAqIGFzIERpYWxvZ3MgZnJvbSBcIi4vZGlhbG9ncy5qc1wiO1xyXG5pbXBvcnQgKiBhcyBFdmVudHMgZnJvbSBcIi4vZXZlbnRzLmpzXCI7XHJcbmltcG9ydCAqIGFzIEZsYWdzIGZyb20gXCIuL2ZsYWdzLmpzXCI7XHJcbmltcG9ydCAqIGFzIFNjcmVlbnMgZnJvbSBcIi4vc2NyZWVucy5qc1wiO1xyXG5pbXBvcnQgKiBhcyBTeXN0ZW0gZnJvbSBcIi4vc3lzdGVtLmpzXCI7XHJcbmltcG9ydCAqIGFzIElPUyBmcm9tIFwiLi9pb3MuanNcIjtcclxuaW1wb3J0IFdpbmRvdywgeyBoYW5kbGVEcmFnRW50ZXIsIGhhbmRsZURyYWdMZWF2ZSwgaGFuZGxlRHJhZ092ZXIgfSBmcm9tIFwiLi93aW5kb3cuanNcIjtcclxuaW1wb3J0ICogYXMgV01MIGZyb20gXCIuL3dtbC5qc1wiO1xyXG5cclxuZXhwb3J0IHtcclxuICAgIEFwcGxpY2F0aW9uLFxyXG4gICAgQnJvd3NlcixcclxuICAgIENhbGwsXHJcbiAgICBDbGlwYm9hcmQsXHJcbiAgICBEaWFsb2dzLFxyXG4gICAgRXZlbnRzLFxyXG4gICAgRmxhZ3MsXHJcbiAgICBTY3JlZW5zLFxyXG4gICAgU3lzdGVtLFxyXG4gICAgSU9TLFxyXG4gICAgV2luZG93LFxyXG4gICAgV01MXHJcbn07XHJcblxyXG4vKipcclxuICogQW4gaW50ZXJuYWwgdXRpbGl0eSBjb25zdW1lZCBieSB0aGUgYmluZGluZyBnZW5lcmF0b3IuXHJcbiAqXHJcbiAqIEBpZ25vcmVcclxuICovXHJcbmV4cG9ydCB7IENyZWF0ZSB9O1xyXG5cclxuZXhwb3J0ICogZnJvbSBcIi4vY2FuY2VsbGFibGUuanNcIjtcclxuXHJcbi8vIEV4cG9ydCB0cmFuc3BvcnQgaW50ZXJmYWNlcyBhbmQgdXRpbGl0aWVzXHJcbmV4cG9ydCB7XHJcbiAgICBzZXRUcmFuc3BvcnQsXHJcbiAgICBnZXRUcmFuc3BvcnQsXHJcbiAgICB0eXBlIFJ1bnRpbWVUcmFuc3BvcnQsXHJcbiAgICBvYmplY3ROYW1lcyxcclxuICAgIGNsaWVudElkLFxyXG59IGZyb20gXCIuL3J1bnRpbWUuanNcIjtcclxuXHJcbmltcG9ydCB7IGNsaWVudElkIH0gZnJvbSBcIi4vcnVudGltZS5qc1wiO1xyXG5cclxuLy8gTm90aWZ5IGJhY2tlbmRcclxud2luZG93Ll93YWlscy5pbnZva2UgPSBTeXN0ZW0uaW52b2tlO1xyXG53aW5kb3cuX3dhaWxzLmNsaWVudElkID0gY2xpZW50SWQ7XHJcblxyXG4vLyBSZWdpc3RlciBwbGF0Zm9ybSBoYW5kbGVycyAoaW50ZXJuYWwgQVBJKVxyXG4vLyBOb3RlOiBXaW5kb3cgaXMgdGhlIHRoaXNXaW5kb3cgaW5zdGFuY2UgKGRlZmF1bHQgZXhwb3J0IGZyb20gd2luZG93LnRzKVxyXG4vLyBCaW5kaW5nIGVuc3VyZXMgJ3RoaXMnIGNvcnJlY3RseSByZWZlcnMgdG8gdGhlIGN1cnJlbnQgd2luZG93IGluc3RhbmNlXHJcbndpbmRvdy5fd2FpbHMuaGFuZGxlUGxhdGZvcm1GaWxlRHJvcCA9IFdpbmRvdy5IYW5kbGVQbGF0Zm9ybUZpbGVEcm9wLmJpbmQoV2luZG93KTtcclxuXHJcbi8vIExpbnV4LXNwZWNpZmljIGRyYWcgaGFuZGxlcnMgKEdUSyBpbnRlcmNlcHRzIERPTSBkcmFnIGV2ZW50cylcclxud2luZG93Ll93YWlscy5oYW5kbGVEcmFnRW50ZXIgPSBoYW5kbGVEcmFnRW50ZXI7XHJcbndpbmRvdy5fd2FpbHMuaGFuZGxlRHJhZ0xlYXZlID0gaGFuZGxlRHJhZ0xlYXZlO1xyXG53aW5kb3cuX3dhaWxzLmhhbmRsZURyYWdPdmVyID0gaGFuZGxlRHJhZ092ZXI7XHJcblxyXG5TeXN0ZW0uaW52b2tlKFwid2FpbHM6cnVudGltZTpyZWFkeVwiKTtcclxuXHJcbi8qKlxyXG4gKiBMb2FkcyBhIHNjcmlwdCBmcm9tIHRoZSBnaXZlbiBVUkwgaWYgaXQgZXhpc3RzLlxyXG4gKiBVc2VzIEhFQUQgcmVxdWVzdCB0byBjaGVjayBleGlzdGVuY2UsIHRoZW4gaW5qZWN0cyBhIHNjcmlwdCB0YWcuXHJcbiAqIFNpbGVudGx5IGlnbm9yZXMgaWYgdGhlIHNjcmlwdCBkb2Vzbid0IGV4aXN0LlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIGxvYWRPcHRpb25hbFNjcmlwdCh1cmw6IHN0cmluZyk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgcmV0dXJuIGZldGNoKHVybCwgeyBtZXRob2Q6ICdIRUFEJyB9KVxyXG4gICAgICAgIC50aGVuKHJlc3BvbnNlID0+IHtcclxuICAgICAgICAgICAgaWYgKHJlc3BvbnNlLm9rKSB7XHJcbiAgICAgICAgICAgICAgICAvLyBWZXJpZnkgdGhlIHJlc3BvbnNlIGlzIGFjdHVhbGx5IEphdmFTY3JpcHQgYW5kIG5vdCBhbiBIVE1MIGZhbGxiYWNrXHJcbiAgICAgICAgICAgICAgICAvLyAoZS5nLiBWaXRlIGRldiBzZXJ2ZXIgcmV0dXJucyBpbmRleC5odG1sIGZvciB1bmtub3duIHJvdXRlcylcclxuICAgICAgICAgICAgICAgIGNvbnN0IGNvbnRlbnRUeXBlID0gKHJlc3BvbnNlLmhlYWRlcnMuZ2V0KCdjb250ZW50LXR5cGUnKSB8fCAnJykudG9Mb3dlckNhc2UoKTtcclxuICAgICAgICAgICAgICAgIGlmIChjb250ZW50VHlwZS5pbmNsdWRlcygnamF2YXNjcmlwdCcpKSB7XHJcbiAgICAgICAgICAgICAgICAgICAgY29uc3Qgc2NyaXB0ID0gZG9jdW1lbnQuY3JlYXRlRWxlbWVudCgnc2NyaXB0Jyk7XHJcbiAgICAgICAgICAgICAgICAgICAgc2NyaXB0LnNyYyA9IHVybDtcclxuICAgICAgICAgICAgICAgICAgICBkb2N1bWVudC5oZWFkLmFwcGVuZENoaWxkKHNjcmlwdCk7XHJcbiAgICAgICAgICAgICAgICB9XHJcbiAgICAgICAgICAgIH1cclxuICAgICAgICB9KVxyXG4gICAgICAgIC5jYXRjaCgoKSA9PiB7fSk7IC8vIFNpbGVudGx5IGlnbm9yZSAtIHNjcmlwdCBpcyBvcHRpb25hbFxyXG59XHJcblxyXG4vLyBMb2FkIGN1c3RvbS5qcyBpZiBhdmFpbGFibGUgKHVzZWQgYnkgc2VydmVyIG1vZGUgZm9yIFdlYlNvY2tldCBldmVudHMsIGV0Yy4pXHJcbmxvYWRPcHRpb25hbFNjcmlwdCgnL3dhaWxzL2N1c3RvbS5qcycpO1xyXG4iLCAiLypcclxuIF8gICAgIF9fICAgICBfIF9fXHJcbnwgfCAgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuaW1wb3J0IHsgT3BlblVSTCB9IGZyb20gXCIuL2Jyb3dzZXIuanNcIjtcclxuaW1wb3J0IHsgUXVlc3Rpb24gfSBmcm9tIFwiLi9kaWFsb2dzLmpzXCI7XHJcbmltcG9ydCB7IEVtaXQgfSBmcm9tIFwiLi9ldmVudHMuanNcIjtcclxuaW1wb3J0IHsgY2FuQWJvcnRMaXN0ZW5lcnMsIHdoZW5SZWFkeSB9IGZyb20gXCIuL3V0aWxzLmpzXCI7XHJcbmltcG9ydCBXaW5kb3cgZnJvbSBcIi4vd2luZG93LmpzXCI7XHJcblxyXG4vKipcclxuICogU2VuZHMgYW4gZXZlbnQgd2l0aCB0aGUgZ2l2ZW4gbmFtZSBhbmQgb3B0aW9uYWwgZGF0YS5cclxuICpcclxuICogQHBhcmFtIGV2ZW50TmFtZSAtIC0gVGhlIG5hbWUgb2YgdGhlIGV2ZW50IHRvIHNlbmQuXHJcbiAqIEBwYXJhbSBbZGF0YT1udWxsXSAtIC0gT3B0aW9uYWwgZGF0YSB0byBzZW5kIGFsb25nIHdpdGggdGhlIGV2ZW50LlxyXG4gKi9cclxuZnVuY3Rpb24gc2VuZEV2ZW50KGV2ZW50TmFtZTogc3RyaW5nLCBkYXRhOiBhbnkgPSBudWxsKTogdm9pZCB7XHJcbiAgICBFbWl0KGV2ZW50TmFtZSwgZGF0YSk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDYWxscyBhIG1ldGhvZCBvbiBhIHNwZWNpZmllZCB3aW5kb3cuXHJcbiAqXHJcbiAqIEBwYXJhbSB3aW5kb3dOYW1lIC0gVGhlIG5hbWUgb2YgdGhlIHdpbmRvdyB0byBjYWxsIHRoZSBtZXRob2Qgb24uXHJcbiAqIEBwYXJhbSBtZXRob2ROYW1lIC0gVGhlIG5hbWUgb2YgdGhlIG1ldGhvZCB0byBjYWxsLlxyXG4gKi9cclxuZnVuY3Rpb24gY2FsbFdpbmRvd01ldGhvZCh3aW5kb3dOYW1lOiBzdHJpbmcsIG1ldGhvZE5hbWU6IHN0cmluZykge1xyXG4gICAgY29uc3QgdGFyZ2V0V2luZG93ID0gV2luZG93LkdldCh3aW5kb3dOYW1lKTtcclxuICAgIGNvbnN0IG1ldGhvZCA9ICh0YXJnZXRXaW5kb3cgYXMgYW55KVttZXRob2ROYW1lXTtcclxuXHJcbiAgICBpZiAodHlwZW9mIG1ldGhvZCAhPT0gXCJmdW5jdGlvblwiKSB7XHJcbiAgICAgICAgY29uc29sZS5lcnJvcihgV2luZG93IG1ldGhvZCAnJHttZXRob2ROYW1lfScgbm90IGZvdW5kYCk7XHJcbiAgICAgICAgcmV0dXJuO1xyXG4gICAgfVxyXG5cclxuICAgIHRyeSB7XHJcbiAgICAgICAgbWV0aG9kLmNhbGwodGFyZ2V0V2luZG93KTtcclxuICAgIH0gY2F0Y2ggKGUpIHtcclxuICAgICAgICBjb25zb2xlLmVycm9yKGBFcnJvciBjYWxsaW5nIHdpbmRvdyBtZXRob2QgJyR7bWV0aG9kTmFtZX0nOiBgLCBlKTtcclxuICAgIH1cclxufVxyXG5cclxuLyoqXHJcbiAqIFJlc3BvbmRzIHRvIGEgdHJpZ2dlcmluZyBldmVudCBieSBydW5uaW5nIGFwcHJvcHJpYXRlIFdNTCBhY3Rpb25zIGZvciB0aGUgY3VycmVudCB0YXJnZXQuXHJcbiAqL1xyXG5mdW5jdGlvbiBvbldNTFRyaWdnZXJlZChldjogRXZlbnQpOiB2b2lkIHtcclxuICAgIGNvbnN0IGVsZW1lbnQgPSBldi5jdXJyZW50VGFyZ2V0IGFzIEVsZW1lbnQ7XHJcblxyXG4gICAgZnVuY3Rpb24gcnVuRWZmZWN0KGNob2ljZSA9IFwiWWVzXCIpIHtcclxuICAgICAgICBpZiAoY2hvaWNlICE9PSBcIlllc1wiKVxyXG4gICAgICAgICAgICByZXR1cm47XHJcblxyXG4gICAgICAgIGNvbnN0IGV2ZW50VHlwZSA9IGVsZW1lbnQuZ2V0QXR0cmlidXRlKCd3bWwtZXZlbnQnKSB8fCBlbGVtZW50LmdldEF0dHJpYnV0ZSgnZGF0YS13bWwtZXZlbnQnKTtcclxuICAgICAgICBjb25zdCB0YXJnZXRXaW5kb3cgPSBlbGVtZW50LmdldEF0dHJpYnV0ZSgnd21sLXRhcmdldC13aW5kb3cnKSB8fCBlbGVtZW50LmdldEF0dHJpYnV0ZSgnZGF0YS13bWwtdGFyZ2V0LXdpbmRvdycpIHx8IFwiXCI7XHJcbiAgICAgICAgY29uc3Qgd2luZG93TWV0aG9kID0gZWxlbWVudC5nZXRBdHRyaWJ1dGUoJ3dtbC13aW5kb3cnKSB8fCBlbGVtZW50LmdldEF0dHJpYnV0ZSgnZGF0YS13bWwtd2luZG93Jyk7XHJcbiAgICAgICAgY29uc3QgdXJsID0gZWxlbWVudC5nZXRBdHRyaWJ1dGUoJ3dtbC1vcGVudXJsJykgfHwgZWxlbWVudC5nZXRBdHRyaWJ1dGUoJ2RhdGEtd21sLW9wZW51cmwnKTtcclxuXHJcbiAgICAgICAgaWYgKGV2ZW50VHlwZSAhPT0gbnVsbClcclxuICAgICAgICAgICAgc2VuZEV2ZW50KGV2ZW50VHlwZSk7XHJcbiAgICAgICAgaWYgKHdpbmRvd01ldGhvZCAhPT0gbnVsbClcclxuICAgICAgICAgICAgY2FsbFdpbmRvd01ldGhvZCh0YXJnZXRXaW5kb3csIHdpbmRvd01ldGhvZCk7XHJcbiAgICAgICAgaWYgKHVybCAhPT0gbnVsbClcclxuICAgICAgICAgICAgdm9pZCBPcGVuVVJMKHVybCk7XHJcbiAgICB9XHJcblxyXG4gICAgY29uc3QgY29uZmlybSA9IGVsZW1lbnQuZ2V0QXR0cmlidXRlKCd3bWwtY29uZmlybScpIHx8IGVsZW1lbnQuZ2V0QXR0cmlidXRlKCdkYXRhLXdtbC1jb25maXJtJyk7XHJcblxyXG4gICAgaWYgKGNvbmZpcm0pIHtcclxuICAgICAgICBRdWVzdGlvbih7XHJcbiAgICAgICAgICAgIFRpdGxlOiBcIkNvbmZpcm1cIixcclxuICAgICAgICAgICAgTWVzc2FnZTogY29uZmlybSxcclxuICAgICAgICAgICAgRGV0YWNoZWQ6IGZhbHNlLFxyXG4gICAgICAgICAgICBCdXR0b25zOiBbXHJcbiAgICAgICAgICAgICAgICB7IExhYmVsOiBcIlllc1wiIH0sXHJcbiAgICAgICAgICAgICAgICB7IExhYmVsOiBcIk5vXCIsIElzRGVmYXVsdDogdHJ1ZSB9XHJcbiAgICAgICAgICAgIF1cclxuICAgICAgICB9KS50aGVuKHJ1bkVmZmVjdCk7XHJcbiAgICB9IGVsc2Uge1xyXG4gICAgICAgIHJ1bkVmZmVjdCgpO1xyXG4gICAgfVxyXG59XHJcblxyXG4vLyBQcml2YXRlIGZpZWxkIG5hbWVzLlxyXG5jb25zdCBjb250cm9sbGVyU3ltID0gU3ltYm9sKFwiY29udHJvbGxlclwiKTtcclxuY29uc3QgdHJpZ2dlck1hcFN5bSA9IFN5bWJvbChcInRyaWdnZXJNYXBcIik7XHJcbmNvbnN0IGVsZW1lbnRDb3VudFN5bSA9IFN5bWJvbChcImVsZW1lbnRDb3VudFwiKTtcclxuXHJcbi8qKlxyXG4gKiBBYm9ydENvbnRyb2xsZXJSZWdpc3RyeSBkb2VzIG5vdCBhY3R1YWxseSByZW1lbWJlciBhY3RpdmUgZXZlbnQgbGlzdGVuZXJzOiBpbnN0ZWFkXHJcbiAqIGl0IHRpZXMgdGhlbSB0byBhbiBBYm9ydFNpZ25hbCBhbmQgdXNlcyBhbiBBYm9ydENvbnRyb2xsZXIgdG8gcmVtb3ZlIHRoZW0gYWxsIGF0IG9uY2UuXHJcbiAqL1xyXG5jbGFzcyBBYm9ydENvbnRyb2xsZXJSZWdpc3RyeSB7XHJcbiAgICAvLyBQcml2YXRlIGZpZWxkcy5cclxuICAgIFtjb250cm9sbGVyU3ltXTogQWJvcnRDb250cm9sbGVyO1xyXG5cclxuICAgIGNvbnN0cnVjdG9yKCkge1xyXG4gICAgICAgIHRoaXNbY29udHJvbGxlclN5bV0gPSBuZXcgQWJvcnRDb250cm9sbGVyKCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBSZXR1cm5zIGFuIG9wdGlvbnMgb2JqZWN0IGZvciBhZGRFdmVudExpc3RlbmVyIHRoYXQgdGllcyB0aGUgbGlzdGVuZXJcclxuICAgICAqIHRvIHRoZSBBYm9ydFNpZ25hbCBmcm9tIHRoZSBjdXJyZW50IEFib3J0Q29udHJvbGxlci5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0gZWxlbWVudCAtIEFuIEhUTUwgZWxlbWVudFxyXG4gICAgICogQHBhcmFtIHRyaWdnZXJzIC0gVGhlIGxpc3Qgb2YgYWN0aXZlIFdNTCB0cmlnZ2VyIGV2ZW50cyBmb3IgdGhlIHNwZWNpZmllZCBlbGVtZW50c1xyXG4gICAgICovXHJcbiAgICBzZXQoZWxlbWVudDogRWxlbWVudCwgdHJpZ2dlcnM6IHN0cmluZ1tdKTogQWRkRXZlbnRMaXN0ZW5lck9wdGlvbnMge1xyXG4gICAgICAgIHJldHVybiB7IHNpZ25hbDogdGhpc1tjb250cm9sbGVyU3ltXS5zaWduYWwgfTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJlbW92ZXMgYWxsIHJlZ2lzdGVyZWQgZXZlbnQgbGlzdGVuZXJzIGFuZCByZXNldHMgdGhlIHJlZ2lzdHJ5LlxyXG4gICAgICovXHJcbiAgICByZXNldCgpOiB2b2lkIHtcclxuICAgICAgICB0aGlzW2NvbnRyb2xsZXJTeW1dLmFib3J0KCk7XHJcbiAgICAgICAgdGhpc1tjb250cm9sbGVyU3ltXSA9IG5ldyBBYm9ydENvbnRyb2xsZXIoKTtcclxuICAgIH1cclxufVxyXG5cclxuLyoqXHJcbiAqIFdlYWtNYXBSZWdpc3RyeSBtYXBzIGFjdGl2ZSB0cmlnZ2VyIGV2ZW50cyB0byBlYWNoIERPTSBlbGVtZW50IHRocm91Z2ggYSBXZWFrTWFwLlxyXG4gKiBUaGlzIGVuc3VyZXMgdGhhdCB0aGUgbWFwcGluZyByZW1haW5zIHByaXZhdGUgdG8gdGhpcyBtb2R1bGUsIHdoaWxlIHN0aWxsIGFsbG93aW5nIGdhcmJhZ2VcclxuICogY29sbGVjdGlvbiBvZiB0aGUgaW52b2x2ZWQgZWxlbWVudHMuXHJcbiAqL1xyXG5jbGFzcyBXZWFrTWFwUmVnaXN0cnkge1xyXG4gICAgLyoqIFN0b3JlcyB0aGUgY3VycmVudCBlbGVtZW50LXRvLXRyaWdnZXIgbWFwcGluZy4gKi9cclxuICAgIFt0cmlnZ2VyTWFwU3ltXTogV2Vha01hcDxFbGVtZW50LCBzdHJpbmdbXT47XHJcbiAgICAvKiogQ291bnRzIHRoZSBudW1iZXIgb2YgZWxlbWVudHMgd2l0aCBhY3RpdmUgV01MIHRyaWdnZXJzLiAqL1xyXG4gICAgW2VsZW1lbnRDb3VudFN5bV06IG51bWJlcjtcclxuXHJcbiAgICBjb25zdHJ1Y3RvcigpIHtcclxuICAgICAgICB0aGlzW3RyaWdnZXJNYXBTeW1dID0gbmV3IFdlYWtNYXAoKTtcclxuICAgICAgICB0aGlzW2VsZW1lbnRDb3VudFN5bV0gPSAwO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogU2V0cyBhY3RpdmUgdHJpZ2dlcnMgZm9yIHRoZSBzcGVjaWZpZWQgZWxlbWVudC5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0gZWxlbWVudCAtIEFuIEhUTUwgZWxlbWVudFxyXG4gICAgICogQHBhcmFtIHRyaWdnZXJzIC0gVGhlIGxpc3Qgb2YgYWN0aXZlIFdNTCB0cmlnZ2VyIGV2ZW50cyBmb3IgdGhlIHNwZWNpZmllZCBlbGVtZW50XHJcbiAgICAgKi9cclxuICAgIHNldChlbGVtZW50OiBFbGVtZW50LCB0cmlnZ2Vyczogc3RyaW5nW10pOiBBZGRFdmVudExpc3RlbmVyT3B0aW9ucyB7XHJcbiAgICAgICAgaWYgKCF0aGlzW3RyaWdnZXJNYXBTeW1dLmhhcyhlbGVtZW50KSkgeyB0aGlzW2VsZW1lbnRDb3VudFN5bV0rKzsgfVxyXG4gICAgICAgIHRoaXNbdHJpZ2dlck1hcFN5bV0uc2V0KGVsZW1lbnQsIHRyaWdnZXJzKTtcclxuICAgICAgICByZXR1cm4ge307XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBSZW1vdmVzIGFsbCByZWdpc3RlcmVkIGV2ZW50IGxpc3RlbmVycy5cclxuICAgICAqL1xyXG4gICAgcmVzZXQoKTogdm9pZCB7XHJcbiAgICAgICAgaWYgKHRoaXNbZWxlbWVudENvdW50U3ltXSA8PSAwKVxyXG4gICAgICAgICAgICByZXR1cm47XHJcblxyXG4gICAgICAgIGZvciAoY29uc3QgZWxlbWVudCBvZiBkb2N1bWVudC5ib2R5LnF1ZXJ5U2VsZWN0b3JBbGwoJyonKSkge1xyXG4gICAgICAgICAgICBpZiAodGhpc1tlbGVtZW50Q291bnRTeW1dIDw9IDApXHJcbiAgICAgICAgICAgICAgICBicmVhaztcclxuXHJcbiAgICAgICAgICAgIGNvbnN0IHRyaWdnZXJzID0gdGhpc1t0cmlnZ2VyTWFwU3ltXS5nZXQoZWxlbWVudCk7XHJcbiAgICAgICAgICAgIGlmICh0cmlnZ2VycyAhPSBudWxsKSB7IHRoaXNbZWxlbWVudENvdW50U3ltXS0tOyB9XHJcblxyXG4gICAgICAgICAgICBmb3IgKGNvbnN0IHRyaWdnZXIgb2YgdHJpZ2dlcnMgfHwgW10pXHJcbiAgICAgICAgICAgICAgICBlbGVtZW50LnJlbW92ZUV2ZW50TGlzdGVuZXIodHJpZ2dlciwgb25XTUxUcmlnZ2VyZWQpO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgdGhpc1t0cmlnZ2VyTWFwU3ltXSA9IG5ldyBXZWFrTWFwKCk7XHJcbiAgICAgICAgdGhpc1tlbGVtZW50Q291bnRTeW1dID0gMDtcclxuICAgIH1cclxufVxyXG5cclxuY29uc3QgdHJpZ2dlclJlZ2lzdHJ5ID0gY2FuQWJvcnRMaXN0ZW5lcnMoKSA/IG5ldyBBYm9ydENvbnRyb2xsZXJSZWdpc3RyeSgpIDogbmV3IFdlYWtNYXBSZWdpc3RyeSgpO1xyXG5cclxuLyoqXHJcbiAqIEFkZHMgZXZlbnQgbGlzdGVuZXJzIHRvIHRoZSBzcGVjaWZpZWQgZWxlbWVudC5cclxuICovXHJcbmZ1bmN0aW9uIGFkZFdNTExpc3RlbmVycyhlbGVtZW50OiBFbGVtZW50KTogdm9pZCB7XHJcbiAgICBjb25zdCB0cmlnZ2VyUmVnRXhwID0gL1xcUysvZztcclxuICAgIGNvbnN0IHRyaWdnZXJBdHRyID0gKGVsZW1lbnQuZ2V0QXR0cmlidXRlKCd3bWwtdHJpZ2dlcicpIHx8IGVsZW1lbnQuZ2V0QXR0cmlidXRlKCdkYXRhLXdtbC10cmlnZ2VyJykgfHwgXCJjbGlja1wiKTtcclxuICAgIGNvbnN0IHRyaWdnZXJzOiBzdHJpbmdbXSA9IFtdO1xyXG5cclxuICAgIGxldCBtYXRjaDtcclxuICAgIHdoaWxlICgobWF0Y2ggPSB0cmlnZ2VyUmVnRXhwLmV4ZWModHJpZ2dlckF0dHIpKSAhPT0gbnVsbClcclxuICAgICAgICB0cmlnZ2Vycy5wdXNoKG1hdGNoWzBdKTtcclxuXHJcbiAgICBjb25zdCBvcHRpb25zID0gdHJpZ2dlclJlZ2lzdHJ5LnNldChlbGVtZW50LCB0cmlnZ2Vycyk7XHJcbiAgICBmb3IgKGNvbnN0IHRyaWdnZXIgb2YgdHJpZ2dlcnMpXHJcbiAgICAgICAgZWxlbWVudC5hZGRFdmVudExpc3RlbmVyKHRyaWdnZXIsIG9uV01MVHJpZ2dlcmVkLCBvcHRpb25zKTtcclxufVxyXG5cclxuLyoqXHJcbiAqIFNjaGVkdWxlcyBhbiBhdXRvbWF0aWMgcmVsb2FkIG9mIFdNTCB0byBiZSBwZXJmb3JtZWQgYXMgc29vbiBhcyB0aGUgZG9jdW1lbnQgaXMgZnVsbHkgbG9hZGVkLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIEVuYWJsZSgpOiB2b2lkIHtcclxuICAgIHdoZW5SZWFkeShSZWxvYWQpO1xyXG59XHJcblxyXG4vKipcclxuICogUmVsb2FkcyB0aGUgV01MIHBhZ2UgYnkgYWRkaW5nIG5lY2Vzc2FyeSBldmVudCBsaXN0ZW5lcnMgYW5kIGJyb3dzZXIgbGlzdGVuZXJzLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIFJlbG9hZCgpOiB2b2lkIHtcclxuICAgIHRyaWdnZXJSZWdpc3RyeS5yZXNldCgpO1xyXG4gICAgZG9jdW1lbnQuYm9keS5xdWVyeVNlbGVjdG9yQWxsKCdbd21sLWV2ZW50XSwgW3dtbC13aW5kb3ddLCBbd21sLW9wZW51cmxdLCBbZGF0YS13bWwtZXZlbnRdLCBbZGF0YS13bWwtd2luZG93XSwgW2RhdGEtd21sLW9wZW51cmxdJykuZm9yRWFjaChhZGRXTUxMaXN0ZW5lcnMpO1xyXG59XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQgeyBuZXdSdW50aW1lQ2FsbGVyLCBvYmplY3ROYW1lcyB9IGZyb20gXCIuL3J1bnRpbWUuanNcIjtcclxuXHJcbmNvbnN0IGNhbGwgPSBuZXdSdW50aW1lQ2FsbGVyKG9iamVjdE5hbWVzLkJyb3dzZXIpO1xyXG5cclxuY29uc3QgQnJvd3Nlck9wZW5VUkwgPSAwO1xyXG5cclxuLyoqXHJcbiAqIE9wZW4gYSBicm93c2VyIHdpbmRvdyB0byB0aGUgZ2l2ZW4gVVJMLlxyXG4gKlxyXG4gKiBAcGFyYW0gdXJsIC0gVGhlIFVSTCB0byBvcGVuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gT3BlblVSTCh1cmw6IHN0cmluZyB8IFVSTCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgcmV0dXJuIGNhbGwoQnJvd3Nlck9wZW5VUkwsIHt1cmw6IHVybC50b1N0cmluZygpfSk7XHJcbn1cclxuIiwgIi8vIFNvdXJjZTogaHR0cHM6Ly9naXRodWIuY29tL2FpL25hbm9pZFxyXG5cclxuLy8gVGhlIE1JVCBMaWNlbnNlIChNSVQpXHJcbi8vXHJcbi8vIENvcHlyaWdodCAyMDE3IEFuZHJleSBTaXRuaWsgPGFuZHJleUBzaXRuaWsucnU+XHJcbi8vXHJcbi8vIFBlcm1pc3Npb24gaXMgaGVyZWJ5IGdyYW50ZWQsIGZyZWUgb2YgY2hhcmdlLCB0byBhbnkgcGVyc29uIG9idGFpbmluZyBhIGNvcHkgb2ZcclxuLy8gdGhpcyBzb2Z0d2FyZSBhbmQgYXNzb2NpYXRlZCBkb2N1bWVudGF0aW9uIGZpbGVzICh0aGUgXCJTb2Z0d2FyZVwiKSwgdG8gZGVhbCBpblxyXG4vLyB0aGUgU29mdHdhcmUgd2l0aG91dCByZXN0cmljdGlvbiwgaW5jbHVkaW5nIHdpdGhvdXQgbGltaXRhdGlvbiB0aGUgcmlnaHRzIHRvXHJcbi8vIHVzZSwgY29weSwgbW9kaWZ5LCBtZXJnZSwgcHVibGlzaCwgZGlzdHJpYnV0ZSwgc3VibGljZW5zZSwgYW5kL29yIHNlbGwgY29waWVzIG9mXHJcbi8vIHRoZSBTb2Z0d2FyZSwgYW5kIHRvIHBlcm1pdCBwZXJzb25zIHRvIHdob20gdGhlIFNvZnR3YXJlIGlzIGZ1cm5pc2hlZCB0byBkbyBzbyxcclxuLy8gICAgIHN1YmplY3QgdG8gdGhlIGZvbGxvd2luZyBjb25kaXRpb25zOlxyXG4vL1xyXG4vLyAgICAgVGhlIGFib3ZlIGNvcHlyaWdodCBub3RpY2UgYW5kIHRoaXMgcGVybWlzc2lvbiBub3RpY2Ugc2hhbGwgYmUgaW5jbHVkZWQgaW4gYWxsXHJcbi8vIGNvcGllcyBvciBzdWJzdGFudGlhbCBwb3J0aW9ucyBvZiB0aGUgU29mdHdhcmUuXHJcbi8vXHJcbi8vICAgICBUSEUgU09GVFdBUkUgSVMgUFJPVklERUQgXCJBUyBJU1wiLCBXSVRIT1VUIFdBUlJBTlRZIE9GIEFOWSBLSU5ELCBFWFBSRVNTIE9SXHJcbi8vIElNUExJRUQsIElOQ0xVRElORyBCVVQgTk9UIExJTUlURUQgVE8gVEhFIFdBUlJBTlRJRVMgT0YgTUVSQ0hBTlRBQklMSVRZLCBGSVRORVNTXHJcbi8vIEZPUiBBIFBBUlRJQ1VMQVIgUFVSUE9TRSBBTkQgTk9OSU5GUklOR0VNRU5ULiBJTiBOTyBFVkVOVCBTSEFMTCBUSEUgQVVUSE9SUyBPUlxyXG4vLyBDT1BZUklHSFQgSE9MREVSUyBCRSBMSUFCTEUgRk9SIEFOWSBDTEFJTSwgREFNQUdFUyBPUiBPVEhFUiBMSUFCSUxJVFksIFdIRVRIRVJcclxuLy8gSU4gQU4gQUNUSU9OIE9GIENPTlRSQUNULCBUT1JUIE9SIE9USEVSV0lTRSwgQVJJU0lORyBGUk9NLCBPVVQgT0YgT1IgSU5cclxuLy8gQ09OTkVDVElPTiBXSVRIIFRIRSBTT0ZUV0FSRSBPUiBUSEUgVVNFIE9SIE9USEVSIERFQUxJTkdTIElOIFRIRSBTT0ZUV0FSRS5cclxuXHJcbi8vIFRoaXMgYWxwaGFiZXQgdXNlcyBgQS1aYS16MC05Xy1gIHN5bWJvbHMuXHJcbi8vIFRoZSBvcmRlciBvZiBjaGFyYWN0ZXJzIGlzIG9wdGltaXplZCBmb3IgYmV0dGVyIGd6aXAgYW5kIGJyb3RsaSBjb21wcmVzc2lvbi5cclxuLy8gUmVmZXJlbmNlcyB0byB0aGUgc2FtZSBmaWxlICh3b3JrcyBib3RoIGZvciBnemlwIGFuZCBicm90bGkpOlxyXG4vLyBgJ3VzZWAsIGBhbmRvbWAsIGFuZCBgcmljdCdgXHJcbi8vIFJlZmVyZW5jZXMgdG8gdGhlIGJyb3RsaSBkZWZhdWx0IGRpY3Rpb25hcnk6XHJcbi8vIGAtMjZUYCwgYDE5ODNgLCBgNDBweGAsIGA3NXB4YCwgYGJ1c2hgLCBgamFja2AsIGBtaW5kYCwgYHZlcnlgLCBhbmQgYHdvbGZgXHJcbmNvbnN0IHVybEFscGhhYmV0ID1cclxuICAgICd1c2VhbmRvbS0yNlQxOTgzNDBQWDc1cHhKQUNLVkVSWU1JTkRCVVNIV09MRl9HUVpiZmdoamtscXZ3eXpyaWN0J1xyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIG5hbm9pZChzaXplOiBudW1iZXIgPSAyMSk6IHN0cmluZyB7XHJcbiAgICBsZXQgaWQgPSAnJ1xyXG4gICAgLy8gQSBjb21wYWN0IGFsdGVybmF0aXZlIGZvciBgZm9yICh2YXIgaSA9IDA7IGkgPCBzdGVwOyBpKyspYC5cclxuICAgIGxldCBpID0gc2l6ZSB8IDBcclxuICAgIHdoaWxlIChpLS0pIHtcclxuICAgICAgICAvLyBgfCAwYCBpcyBtb3JlIGNvbXBhY3QgYW5kIGZhc3RlciB0aGFuIGBNYXRoLmZsb29yKClgLlxyXG4gICAgICAgIGlkICs9IHVybEFscGhhYmV0WyhNYXRoLnJhbmRvbSgpICogNjQpIHwgMF1cclxuICAgIH1cclxuICAgIHJldHVybiBpZFxyXG59XHJcbiIsICIvKlxyXG4gXyAgICAgX18gICAgIF8gX19cclxufCB8ICAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQgeyBuYW5vaWQgfSBmcm9tIFwiLi9uYW5vaWQuanNcIjtcclxuXHJcbmNvbnN0IHJ1bnRpbWVVUkwgPSB3aW5kb3cubG9jYXRpb24ub3JpZ2luICsgXCIvd2FpbHMvcnVudGltZVwiO1xyXG5cclxuLy8gU3RheSB1bmRlciBXZWJWaWV3MidzIH4yTUIgcmVxdWVzdCBib2R5IGJ1ZmZlcmluZyBsaW1pdCBpbiBXZWJSZXNvdXJjZVJlcXVlc3RlZC5cclxuY29uc3QgQ0hVTktfVEhSRVNIT0xEID0gNTEyICogMTAyNDtcclxuXHJcbi8vIFJlLWV4cG9ydCBuYW5vaWQgZm9yIGN1c3RvbSB0cmFuc3BvcnQgaW1wbGVtZW50YXRpb25zXHJcbmV4cG9ydCB7IG5hbm9pZCB9O1xyXG5cclxuLy8gT2JqZWN0IE5hbWVzXHJcbmV4cG9ydCBjb25zdCBvYmplY3ROYW1lcyA9IE9iamVjdC5mcmVlemUoe1xyXG4gICAgQ2FsbDogMCxcclxuICAgIENsaXBib2FyZDogMSxcclxuICAgIEFwcGxpY2F0aW9uOiAyLFxyXG4gICAgRXZlbnRzOiAzLFxyXG4gICAgQ29udGV4dE1lbnU6IDQsXHJcbiAgICBEaWFsb2c6IDUsXHJcbiAgICBXaW5kb3c6IDYsXHJcbiAgICBTY3JlZW5zOiA3LFxyXG4gICAgU3lzdGVtOiA4LFxyXG4gICAgQnJvd3NlcjogOSxcclxuICAgIENhbmNlbENhbGw6IDEwLFxyXG4gICAgSU9TOiAxMSxcclxufSk7XHJcbmV4cG9ydCBsZXQgY2xpZW50SWQgPSBuYW5vaWQoKTtcclxuXHJcbi8qKlxyXG4gKiBSdW50aW1lVHJhbnNwb3J0IGRlZmluZXMgdGhlIGludGVyZmFjZSBmb3IgY3VzdG9tIElQQyB0cmFuc3BvcnQgaW1wbGVtZW50YXRpb25zLlxyXG4gKiBJbXBsZW1lbnQgdGhpcyBpbnRlcmZhY2UgdG8gdXNlIFdlYlNvY2tldHMsIGN1c3RvbSBwcm90b2NvbHMsIG9yIGFueSBvdGhlclxyXG4gKiB0cmFuc3BvcnQgbWVjaGFuaXNtIGluc3RlYWQgb2YgdGhlIGRlZmF1bHQgSFRUUCBmZXRjaC5cclxuICovXHJcbmV4cG9ydCBpbnRlcmZhY2UgUnVudGltZVRyYW5zcG9ydCB7XHJcbiAgICAvKipcclxuICAgICAqIFNlbmQgYSBydW50aW1lIGNhbGwgYW5kIHJldHVybiB0aGUgcmVzcG9uc2UuXHJcbiAgICAgKlxyXG4gICAgICogQHBhcmFtIG9iamVjdElEIC0gVGhlIFdhaWxzIG9iamVjdCBJRCAoMD1DYWxsLCAxPUNsaXBib2FyZCwgZXRjLilcclxuICAgICAqIEBwYXJhbSBtZXRob2QgLSBUaGUgbWV0aG9kIElEIHRvIGNhbGxcclxuICAgICAqIEBwYXJhbSB3aW5kb3dOYW1lIC0gT3B0aW9uYWwgd2luZG93IG5hbWVcclxuICAgICAqIEBwYXJhbSBhcmdzIC0gQXJndW1lbnRzIHRvIHBhc3MgKHdpbGwgYmUgSlNPTiBzdHJpbmdpZmllZCBpZiBwcmVzZW50KVxyXG4gICAgICogQHJldHVybnMgUHJvbWlzZSB0aGF0IHJlc29sdmVzIHdpdGggdGhlIHJlc3BvbnNlIGRhdGFcclxuICAgICAqL1xyXG4gICAgY2FsbChvYmplY3RJRDogbnVtYmVyLCBtZXRob2Q6IG51bWJlciwgd2luZG93TmFtZTogc3RyaW5nLCBhcmdzOiBhbnkpOiBQcm9taXNlPGFueT47XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDdXN0b20gdHJhbnNwb3J0IGltcGxlbWVudGF0aW9uIChjYW4gYmUgc2V0IGJ5IHVzZXIpXHJcbiAqL1xyXG5sZXQgY3VzdG9tVHJhbnNwb3J0OiBSdW50aW1lVHJhbnNwb3J0IHwgbnVsbCA9IG51bGw7XHJcblxyXG4vKipcclxuICogU2V0IGEgY3VzdG9tIHRyYW5zcG9ydCBmb3IgYWxsIFdhaWxzIHJ1bnRpbWUgY2FsbHMuXHJcbiAqIFRoaXMgYWxsb3dzIHlvdSB0byByZXBsYWNlIHRoZSBkZWZhdWx0IEhUVFAgZmV0Y2ggdHJhbnNwb3J0IHdpdGhcclxuICogV2ViU29ja2V0cywgY3VzdG9tIHByb3RvY29scywgb3IgYW55IG90aGVyIG1lY2hhbmlzbS5cclxuICpcclxuICogQHBhcmFtIHRyYW5zcG9ydCAtIFlvdXIgY3VzdG9tIHRyYW5zcG9ydCBpbXBsZW1lbnRhdGlvblxyXG4gKlxyXG4gKiBAZXhhbXBsZVxyXG4gKiBgYGB0eXBlc2NyaXB0XHJcbiAqIGltcG9ydCB7IHNldFRyYW5zcG9ydCB9IGZyb20gJy93YWlscy9ydW50aW1lLmpzJztcclxuICpcclxuICogY29uc3Qgd3NUcmFuc3BvcnQgPSB7XHJcbiAqICAgY2FsbDogYXN5bmMgKG9iamVjdElELCBtZXRob2QsIHdpbmRvd05hbWUsIGFyZ3MpID0+IHtcclxuICogICAgIC8vIFlvdXIgV2ViU29ja2V0IGltcGxlbWVudGF0aW9uXHJcbiAqICAgfVxyXG4gKiB9O1xyXG4gKlxyXG4gKiBzZXRUcmFuc3BvcnQod3NUcmFuc3BvcnQpO1xyXG4gKiBgYGBcclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBzZXRUcmFuc3BvcnQodHJhbnNwb3J0OiBSdW50aW1lVHJhbnNwb3J0IHwgbnVsbCk6IHZvaWQge1xyXG4gICAgY3VzdG9tVHJhbnNwb3J0ID0gdHJhbnNwb3J0O1xyXG59XHJcblxyXG4vKipcclxuICogR2V0IHRoZSBjdXJyZW50IHRyYW5zcG9ydCAodXNlZnVsIGZvciBleHRlbmRpbmcvd3JhcHBpbmcpXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gZ2V0VHJhbnNwb3J0KCk6IFJ1bnRpbWVUcmFuc3BvcnQgfCBudWxsIHtcclxuICAgIHJldHVybiBjdXN0b21UcmFuc3BvcnQ7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDcmVhdGVzIGEgbmV3IHJ1bnRpbWUgY2FsbGVyIHdpdGggc3BlY2lmaWVkIElELlxyXG4gKlxyXG4gKiBAcGFyYW0gb2JqZWN0IC0gVGhlIG9iamVjdCB0byBpbnZva2UgdGhlIG1ldGhvZCBvbi5cclxuICogQHBhcmFtIHdpbmRvd05hbWUgLSBUaGUgbmFtZSBvZiB0aGUgd2luZG93LlxyXG4gKiBAcmV0dXJuIFRoZSBuZXcgcnVudGltZSBjYWxsZXIgZnVuY3Rpb24uXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gbmV3UnVudGltZUNhbGxlcihvYmplY3Q6IG51bWJlciwgd2luZG93TmFtZTogc3RyaW5nID0gJycpIHtcclxuICAgIHJldHVybiBmdW5jdGlvbiAobWV0aG9kOiBudW1iZXIsIGFyZ3M6IGFueSA9IG51bGwpIHtcclxuICAgICAgICByZXR1cm4gcnVudGltZUNhbGxXaXRoSUQob2JqZWN0LCBtZXRob2QsIHdpbmRvd05hbWUsIGFyZ3MpO1xyXG4gICAgfTtcclxufVxyXG5cclxuYXN5bmMgZnVuY3Rpb24gcnVudGltZUNhbGxXaXRoSUQob2JqZWN0SUQ6IG51bWJlciwgbWV0aG9kOiBudW1iZXIsIHdpbmRvd05hbWU6IHN0cmluZywgYXJnczogYW55KTogUHJvbWlzZTxhbnk+IHtcclxuICAgIC8vIFVzZSBjdXN0b20gdHJhbnNwb3J0IGlmIGF2YWlsYWJsZVxyXG4gICAgaWYgKGN1c3RvbVRyYW5zcG9ydCkge1xyXG4gICAgICAgIHJldHVybiBjdXN0b21UcmFuc3BvcnQuY2FsbChvYmplY3RJRCwgbWV0aG9kLCB3aW5kb3dOYW1lLCBhcmdzKTtcclxuICAgIH1cclxuXHJcbiAgICAvLyBEZWZhdWx0IEhUVFAgZmV0Y2ggdHJhbnNwb3J0XHJcbiAgICBsZXQgdXJsID0gbmV3IFVSTChydW50aW1lVVJMKTtcclxuXHJcbiAgICBsZXQgYm9keTogeyBvYmplY3Q6IG51bWJlcjsgbWV0aG9kOiBudW1iZXIsIGFyZ3M/OiBhbnkgfSA9IHtcclxuICAgICAgb2JqZWN0OiBvYmplY3RJRCxcclxuICAgICAgbWV0aG9kXHJcbiAgICB9XHJcbiAgICBpZiAoYXJncyAhPT0gbnVsbCAmJiBhcmdzICE9PSB1bmRlZmluZWQpIHtcclxuICAgICAgYm9keS5hcmdzID0gYXJncztcclxuICAgIH1cclxuXHJcbiAgICBsZXQgaGVhZGVyczogUmVjb3JkPHN0cmluZywgc3RyaW5nPiA9IHtcclxuICAgICAgICBbXCJ4LXdhaWxzLWNsaWVudC1pZFwiXTogY2xpZW50SWQsXHJcbiAgICAgICAgW1wiQ29udGVudC1UeXBlXCJdOiBcImFwcGxpY2F0aW9uL2pzb25cIlxyXG4gICAgfVxyXG4gICAgaWYgKHdpbmRvd05hbWUpIHtcclxuICAgICAgICBoZWFkZXJzW1wieC13YWlscy13aW5kb3ctbmFtZVwiXSA9IHdpbmRvd05hbWU7XHJcbiAgICB9XHJcblxyXG4gICAgY29uc3QgYm9keVN0ciA9IEpTT04uc3RyaW5naWZ5KGJvZHkpO1xyXG4gICAgbGV0IHJlc3BvbnNlOiBSZXNwb25zZTtcclxuICAgIGlmIChib2R5U3RyLmxlbmd0aCA+IENIVU5LX1RIUkVTSE9MRCkge1xyXG4gICAgICAgIHJlc3BvbnNlID0gYXdhaXQgc2VuZENodW5rZWQodXJsLCBoZWFkZXJzLCBib2R5U3RyKTtcclxuICAgIH0gZWxzZSB7XHJcbiAgICAgICAgcmVzcG9uc2UgPSBhd2FpdCBmZXRjaCh1cmwsIHsgbWV0aG9kOiAnUE9TVCcsIGhlYWRlcnMsIGJvZHk6IGJvZHlTdHIgfSk7XHJcbiAgICB9XHJcbiAgICBpZiAoIXJlc3BvbnNlLm9rKSB7XHJcbiAgICAgICAgdGhyb3cgbmV3IEVycm9yKGF3YWl0IHJlc3BvbnNlLnRleHQoKSk7XHJcbiAgICB9XHJcblxyXG4gICAgaWYgKChyZXNwb25zZS5oZWFkZXJzLmdldChcIkNvbnRlbnQtVHlwZVwiKT8uaW5kZXhPZihcImFwcGxpY2F0aW9uL2pzb25cIikgPz8gLTEpICE9PSAtMSkge1xyXG4gICAgICAgIHJldHVybiByZXNwb25zZS5qc29uKCk7XHJcbiAgICB9IGVsc2Uge1xyXG4gICAgICAgIHJldHVybiByZXNwb25zZS50ZXh0KCk7XHJcbiAgICB9XHJcbn1cclxuXHJcbi8vIHNlbmRDaHVua2VkIHNwbGl0cyBhIGxhcmdlIHNlcmlhbGlzZWQgcmVxdWVzdCBib2R5IGludG8gQ0hVTktfVEhSRVNIT0xELXNpemVkXHJcbi8vIGJ5dGUgY2h1bmtzIGFuZCBzZW5kcyB0aGVtIHNlcmlhbGx5LiAgRW5jb2RpbmcgdG8gVVRGLTggYnl0ZXMgYmVmb3JlIHNsaWNpbmdcclxuLy8gcHJldmVudHMgY29ycnVwdGlvbiBvZiBub24tQk1QIGNoYXJhY3RlcnMgKHN1cnJvZ2F0ZSBwYWlycykgdGhhdCB3b3VsZCBvY2N1clxyXG4vLyB3aGVuIHNwbGl0dGluZyBhdCBKYXZhU2NyaXB0IHN0cmluZyBpbmRpY2VzLiAgVGhlIEdvIHRyYW5zcG9ydCBhc3NlbWJsZXMgdGhlXHJcbi8vIHJhdyBieXRlcyBiZWZvcmUgcHJvY2Vzc2luZy4gIE9ubHkgdGhlIGZpbmFsIGNodW5rJ3MgcmVzcG9uc2UgY2FycmllcyB0aGUgUlBDIHJlc3VsdC5cclxuYXN5bmMgZnVuY3Rpb24gc2VuZENodW5rZWQodXJsOiBVUkwsIGhlYWRlcnM6IFJlY29yZDxzdHJpbmcsIHN0cmluZz4sIGJvZHlTdHI6IHN0cmluZyk6IFByb21pc2U8UmVzcG9uc2U+IHtcclxuICAgIGNvbnN0IGNodW5rSWQgPSBuYW5vaWQoKTtcclxuICAgIGNvbnN0IGJvZHlCeXRlcyA9IG5ldyBUZXh0RW5jb2RlcigpLmVuY29kZShib2R5U3RyKTtcclxuICAgIGNvbnN0IHRvdGFsQ2h1bmtzID0gTWF0aC5jZWlsKGJvZHlCeXRlcy5sZW5ndGggLyBDSFVOS19USFJFU0hPTEQpO1xyXG5cclxuICAgIGZvciAobGV0IGkgPSAwOyBpIDwgdG90YWxDaHVua3MgLSAxOyBpKyspIHtcclxuICAgICAgICBjb25zdCBjaHVuayA9IGJvZHlCeXRlcy5zdWJhcnJheShpICogQ0hVTktfVEhSRVNIT0xELCAoaSArIDEpICogQ0hVTktfVEhSRVNIT0xEKTtcclxuICAgICAgICBjb25zdCByZXNwID0gYXdhaXQgZmV0Y2godXJsLCB7XHJcbiAgICAgICAgICAgIG1ldGhvZDogJ1BPU1QnLFxyXG4gICAgICAgICAgICBoZWFkZXJzOiB7XHJcbiAgICAgICAgICAgICAgICAuLi5oZWFkZXJzLFxyXG4gICAgICAgICAgICAgICAgJ3gtd2FpbHMtY2h1bmstaWQnOiBjaHVua0lkLFxyXG4gICAgICAgICAgICAgICAgJ3gtd2FpbHMtY2h1bmstaW5kZXgnOiBTdHJpbmcoaSksXHJcbiAgICAgICAgICAgICAgICAneC13YWlscy1jaHVuay10b3RhbCc6IFN0cmluZyh0b3RhbENodW5rcyksXHJcbiAgICAgICAgICAgIH0sXHJcbiAgICAgICAgICAgIGJvZHk6IGNodW5rLFxyXG4gICAgICAgIH0pO1xyXG4gICAgICAgIGlmICghcmVzcC5vaykge1xyXG4gICAgICAgICAgICB0aHJvdyBuZXcgRXJyb3IoYXdhaXQgcmVzcC50ZXh0KCkpO1xyXG4gICAgICAgIH1cclxuICAgICAgICBjb25zdCBib2R5ID0gYXdhaXQgcmVzcC50ZXh0KCk7XHJcbiAgICAgICAgaWYgKGJvZHkudHJpbSgpKSB7XHJcbiAgICAgICAgICAgIGNvbnNvbGUud2FybignW1dhaWxzXSBVbmV4cGVjdGVkIG5vbi1lbXB0eSBib2R5IGZyb20gaW50ZXJtZWRpYXRlIGNodW5rOicsIGJvZHkpO1xyXG4gICAgICAgICAgICB0aHJvdyBuZXcgRXJyb3IoYm9keSk7XHJcbiAgICAgICAgfVxyXG4gICAgfVxyXG5cclxuICAgIHJldHVybiBmZXRjaCh1cmwsIHtcclxuICAgICAgICBtZXRob2Q6ICdQT1NUJyxcclxuICAgICAgICBoZWFkZXJzOiB7XHJcbiAgICAgICAgICAgIC4uLmhlYWRlcnMsXHJcbiAgICAgICAgICAgICd4LXdhaWxzLWNodW5rLWlkJzogY2h1bmtJZCxcclxuICAgICAgICAgICAgJ3gtd2FpbHMtY2h1bmstaW5kZXgnOiBTdHJpbmcodG90YWxDaHVua3MgLSAxKSxcclxuICAgICAgICAgICAgJ3gtd2FpbHMtY2h1bmstdG90YWwnOiBTdHJpbmcodG90YWxDaHVua3MpLFxyXG4gICAgICAgIH0sXHJcbiAgICAgICAgYm9keTogYm9keUJ5dGVzLnN1YmFycmF5KCh0b3RhbENodW5rcyAtIDEpICogQ0hVTktfVEhSRVNIT0xEKSxcclxuICAgIH0pO1xyXG59XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQge25ld1J1bnRpbWVDYWxsZXIsIG9iamVjdE5hbWVzfSBmcm9tIFwiLi9ydW50aW1lLmpzXCI7XHJcblxyXG4vLyBzZXR1cFxyXG53aW5kb3cuX3dhaWxzID0gd2luZG93Ll93YWlscyB8fCB7fTtcclxuXHJcbmNvbnN0IGNhbGwgPSBuZXdSdW50aW1lQ2FsbGVyKG9iamVjdE5hbWVzLkRpYWxvZyk7XHJcblxyXG4vLyBEZWZpbmUgY29uc3RhbnRzIGZyb20gdGhlIGBtZXRob2RzYCBvYmplY3QgaW4gVGl0bGUgQ2FzZVxyXG5jb25zdCBEaWFsb2dJbmZvID0gMDtcclxuY29uc3QgRGlhbG9nV2FybmluZyA9IDE7XHJcbmNvbnN0IERpYWxvZ0Vycm9yID0gMjtcclxuY29uc3QgRGlhbG9nUXVlc3Rpb24gPSAzO1xyXG5jb25zdCBEaWFsb2dPcGVuRmlsZSA9IDQ7XHJcbmNvbnN0IERpYWxvZ1NhdmVGaWxlID0gNTtcclxuXHJcbmV4cG9ydCBpbnRlcmZhY2UgT3BlbkZpbGVEaWFsb2dPcHRpb25zIHtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgZGlyZWN0b3JpZXMgY2FuIGJlIGNob3Nlbi4gKi9cclxuICAgIENhbkNob29zZURpcmVjdG9yaWVzPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgZmlsZXMgY2FuIGJlIGNob3Nlbi4gKi9cclxuICAgIENhbkNob29zZUZpbGVzPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgZGlyZWN0b3JpZXMgY2FuIGJlIGNyZWF0ZWQuICovXHJcbiAgICBDYW5DcmVhdGVEaXJlY3Rvcmllcz86IGJvb2xlYW47XHJcbiAgICAvKiogSW5kaWNhdGVzIGlmIGhpZGRlbiBmaWxlcyBzaG91bGQgYmUgc2hvd24uICovXHJcbiAgICBTaG93SGlkZGVuRmlsZXM/OiBib29sZWFuO1xyXG4gICAgLyoqIEluZGljYXRlcyBpZiBhbGlhc2VzIHNob3VsZCBiZSByZXNvbHZlZC4gKi9cclxuICAgIFJlc29sdmVzQWxpYXNlcz86IGJvb2xlYW47XHJcbiAgICAvKiogSW5kaWNhdGVzIGlmIG11bHRpcGxlIHNlbGVjdGlvbiBpcyBhbGxvd2VkLiAqL1xyXG4gICAgQWxsb3dzTXVsdGlwbGVTZWxlY3Rpb24/OiBib29sZWFuO1xyXG4gICAgLyoqIEluZGljYXRlcyBpZiB0aGUgZXh0ZW5zaW9uIHNob3VsZCBiZSBoaWRkZW4uICovXHJcbiAgICBIaWRlRXh0ZW5zaW9uPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgaGlkZGVuIGV4dGVuc2lvbnMgY2FuIGJlIHNlbGVjdGVkLiAqL1xyXG4gICAgQ2FuU2VsZWN0SGlkZGVuRXh0ZW5zaW9uPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgZmlsZSBwYWNrYWdlcyBzaG91bGQgYmUgdHJlYXRlZCBhcyBkaXJlY3Rvcmllcy4gKi9cclxuICAgIFRyZWF0c0ZpbGVQYWNrYWdlc0FzRGlyZWN0b3JpZXM/OiBib29sZWFuO1xyXG4gICAgLyoqIEluZGljYXRlcyBpZiBvdGhlciBmaWxlIHR5cGVzIGFyZSBhbGxvd2VkLiAqL1xyXG4gICAgQWxsb3dzT3RoZXJGaWxldHlwZXM/OiBib29sZWFuO1xyXG4gICAgLyoqIEFycmF5IG9mIGZpbGUgZmlsdGVycy4gKi9cclxuICAgIEZpbHRlcnM/OiBGaWxlRmlsdGVyW107XHJcbiAgICAvKiogVGl0bGUgb2YgdGhlIGRpYWxvZy4gKi9cclxuICAgIFRpdGxlPzogc3RyaW5nO1xyXG4gICAgLyoqIE1lc3NhZ2UgdG8gc2hvdyBpbiB0aGUgZGlhbG9nLiAqL1xyXG4gICAgTWVzc2FnZT86IHN0cmluZztcclxuICAgIC8qKiBUZXh0IHRvIGRpc3BsYXkgb24gdGhlIGJ1dHRvbi4gKi9cclxuICAgIEJ1dHRvblRleHQ/OiBzdHJpbmc7XHJcbiAgICAvKiogRGlyZWN0b3J5IHRvIG9wZW4gaW4gdGhlIGRpYWxvZy4gKi9cclxuICAgIERpcmVjdG9yeT86IHN0cmluZztcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgdGhlIGRpYWxvZyBzaG91bGQgYXBwZWFyIGRldGFjaGVkIGZyb20gdGhlIG1haW4gd2luZG93LiAqL1xyXG4gICAgRGV0YWNoZWQ/OiBib29sZWFuO1xyXG59XHJcblxyXG5leHBvcnQgaW50ZXJmYWNlIFNhdmVGaWxlRGlhbG9nT3B0aW9ucyB7XHJcbiAgICAvKiogRGVmYXVsdCBmaWxlbmFtZSB0byB1c2UgaW4gdGhlIGRpYWxvZy4gKi9cclxuICAgIEZpbGVuYW1lPzogc3RyaW5nO1xyXG4gICAgLyoqIEluZGljYXRlcyBpZiBkaXJlY3RvcmllcyBjYW4gYmUgY2hvc2VuLiAqL1xyXG4gICAgQ2FuQ2hvb3NlRGlyZWN0b3JpZXM/OiBib29sZWFuO1xyXG4gICAgLyoqIEluZGljYXRlcyBpZiBmaWxlcyBjYW4gYmUgY2hvc2VuLiAqL1xyXG4gICAgQ2FuQ2hvb3NlRmlsZXM/OiBib29sZWFuO1xyXG4gICAgLyoqIEluZGljYXRlcyBpZiBkaXJlY3RvcmllcyBjYW4gYmUgY3JlYXRlZC4gKi9cclxuICAgIENhbkNyZWF0ZURpcmVjdG9yaWVzPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgaGlkZGVuIGZpbGVzIHNob3VsZCBiZSBzaG93bi4gKi9cclxuICAgIFNob3dIaWRkZW5GaWxlcz86IGJvb2xlYW47XHJcbiAgICAvKiogSW5kaWNhdGVzIGlmIGFsaWFzZXMgc2hvdWxkIGJlIHJlc29sdmVkLiAqL1xyXG4gICAgUmVzb2x2ZXNBbGlhc2VzPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgdGhlIGV4dGVuc2lvbiBzaG91bGQgYmUgaGlkZGVuLiAqL1xyXG4gICAgSGlkZUV4dGVuc2lvbj86IGJvb2xlYW47XHJcbiAgICAvKiogSW5kaWNhdGVzIGlmIGhpZGRlbiBleHRlbnNpb25zIGNhbiBiZSBzZWxlY3RlZC4gKi9cclxuICAgIENhblNlbGVjdEhpZGRlbkV4dGVuc2lvbj86IGJvb2xlYW47XHJcbiAgICAvKiogSW5kaWNhdGVzIGlmIGZpbGUgcGFja2FnZXMgc2hvdWxkIGJlIHRyZWF0ZWQgYXMgZGlyZWN0b3JpZXMuICovXHJcbiAgICBUcmVhdHNGaWxlUGFja2FnZXNBc0RpcmVjdG9yaWVzPzogYm9vbGVhbjtcclxuICAgIC8qKiBJbmRpY2F0ZXMgaWYgb3RoZXIgZmlsZSB0eXBlcyBhcmUgYWxsb3dlZC4gKi9cclxuICAgIEFsbG93c090aGVyRmlsZXR5cGVzPzogYm9vbGVhbjtcclxuICAgIC8qKiBBcnJheSBvZiBmaWxlIGZpbHRlcnMuICovXHJcbiAgICBGaWx0ZXJzPzogRmlsZUZpbHRlcltdO1xyXG4gICAgLyoqIFRpdGxlIG9mIHRoZSBkaWFsb2cuICovXHJcbiAgICBUaXRsZT86IHN0cmluZztcclxuICAgIC8qKiBNZXNzYWdlIHRvIHNob3cgaW4gdGhlIGRpYWxvZy4gKi9cclxuICAgIE1lc3NhZ2U/OiBzdHJpbmc7XHJcbiAgICAvKiogVGV4dCB0byBkaXNwbGF5IG9uIHRoZSBidXR0b24uICovXHJcbiAgICBCdXR0b25UZXh0Pzogc3RyaW5nO1xyXG4gICAgLyoqIERpcmVjdG9yeSB0byBvcGVuIGluIHRoZSBkaWFsb2cuICovXHJcbiAgICBEaXJlY3Rvcnk/OiBzdHJpbmc7XHJcbiAgICAvKiogSW5kaWNhdGVzIGlmIHRoZSBkaWFsb2cgc2hvdWxkIGFwcGVhciBkZXRhY2hlZCBmcm9tIHRoZSBtYWluIHdpbmRvdy4gKi9cclxuICAgIERldGFjaGVkPzogYm9vbGVhbjtcclxufVxyXG5cclxuZXhwb3J0IGludGVyZmFjZSBNZXNzYWdlRGlhbG9nT3B0aW9ucyB7XHJcbiAgICAvKiogVGhlIHRpdGxlIG9mIHRoZSBkaWFsb2cgd2luZG93LiAqL1xyXG4gICAgVGl0bGU/OiBzdHJpbmc7XHJcbiAgICAvKiogVGhlIG1haW4gbWVzc2FnZSB0byBzaG93IGluIHRoZSBkaWFsb2cuICovXHJcbiAgICBNZXNzYWdlPzogc3RyaW5nO1xyXG4gICAgLyoqIEFycmF5IG9mIGJ1dHRvbiBvcHRpb25zIHRvIHNob3cgaW4gdGhlIGRpYWxvZy4gKi9cclxuICAgIEJ1dHRvbnM/OiBCdXR0b25bXTtcclxuICAgIC8qKiBUcnVlIGlmIHRoZSBkaWFsb2cgc2hvdWxkIGFwcGVhciBkZXRhY2hlZCBmcm9tIHRoZSBtYWluIHdpbmRvdyAoaWYgYXBwbGljYWJsZSkuICovXHJcbiAgICBEZXRhY2hlZD86IGJvb2xlYW47XHJcbn1cclxuXHJcbmV4cG9ydCBpbnRlcmZhY2UgQnV0dG9uIHtcclxuICAgIC8qKiBUZXh0IHRoYXQgYXBwZWFycyB3aXRoaW4gdGhlIGJ1dHRvbi4gKi9cclxuICAgIExhYmVsPzogc3RyaW5nO1xyXG4gICAgLyoqIFRydWUgaWYgdGhlIGJ1dHRvbiBzaG91bGQgY2FuY2VsIGFuIG9wZXJhdGlvbiB3aGVuIGNsaWNrZWQuICovXHJcbiAgICBJc0NhbmNlbD86IGJvb2xlYW47XHJcbiAgICAvKiogVHJ1ZSBpZiB0aGUgYnV0dG9uIHNob3VsZCBiZSB0aGUgZGVmYXVsdCBhY3Rpb24gd2hlbiB0aGUgdXNlciBwcmVzc2VzIGVudGVyLiAqL1xyXG4gICAgSXNEZWZhdWx0PzogYm9vbGVhbjtcclxufVxyXG5cclxuZXhwb3J0IGludGVyZmFjZSBGaWxlRmlsdGVyIHtcclxuICAgIC8qKiBEaXNwbGF5IG5hbWUgZm9yIHRoZSBmaWx0ZXIsIGl0IGNvdWxkIGJlIFwiVGV4dCBGaWxlc1wiLCBcIkltYWdlc1wiIGV0Yy4gKi9cclxuICAgIERpc3BsYXlOYW1lPzogc3RyaW5nO1xyXG4gICAgLyoqIFBhdHRlcm4gdG8gbWF0Y2ggZm9yIHRoZSBmaWx0ZXIsIGUuZy4gXCIqLnR4dDsqLm1kXCIgZm9yIHRleHQgbWFya2Rvd24gZmlsZXMuICovXHJcbiAgICBQYXR0ZXJuPzogc3RyaW5nO1xyXG59XHJcblxyXG4vKipcclxuICogUHJlc2VudHMgYSBkaWFsb2cgb2Ygc3BlY2lmaWVkIHR5cGUgd2l0aCB0aGUgZ2l2ZW4gb3B0aW9ucy5cclxuICpcclxuICogQHBhcmFtIHR5cGUgLSBEaWFsb2cgdHlwZS5cclxuICogQHBhcmFtIG9wdGlvbnMgLSBPcHRpb25zIGZvciB0aGUgZGlhbG9nLlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCByZXNvbHZlcyB3aXRoIHJlc3VsdCBvZiBkaWFsb2cuXHJcbiAqL1xyXG5mdW5jdGlvbiBkaWFsb2codHlwZTogbnVtYmVyLCBvcHRpb25zOiBNZXNzYWdlRGlhbG9nT3B0aW9ucyB8IE9wZW5GaWxlRGlhbG9nT3B0aW9ucyB8IFNhdmVGaWxlRGlhbG9nT3B0aW9ucyA9IHt9KTogUHJvbWlzZTxhbnk+IHtcclxuICAgIHJldHVybiBjYWxsKHR5cGUsIG9wdGlvbnMpO1xyXG59XHJcblxyXG4vKipcclxuICogUHJlc2VudHMgYW4gaW5mbyBkaWFsb2cuXHJcbiAqXHJcbiAqIEBwYXJhbSBvcHRpb25zIC0gRGlhbG9nIG9wdGlvbnNcclxuICogQHJldHVybnMgQSBwcm9taXNlIHRoYXQgcmVzb2x2ZXMgd2l0aCB0aGUgbGFiZWwgb2YgdGhlIGNob3NlbiBidXR0b24uXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSW5mbyhvcHRpb25zOiBNZXNzYWdlRGlhbG9nT3B0aW9ucyk6IFByb21pc2U8c3RyaW5nPiB7IHJldHVybiBkaWFsb2coRGlhbG9nSW5mbywgb3B0aW9ucyk7IH1cclxuXHJcbi8qKlxyXG4gKiBQcmVzZW50cyBhIHdhcm5pbmcgZGlhbG9nLlxyXG4gKlxyXG4gKiBAcGFyYW0gb3B0aW9ucyAtIERpYWxvZyBvcHRpb25zLlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCByZXNvbHZlcyB3aXRoIHRoZSBsYWJlbCBvZiB0aGUgY2hvc2VuIGJ1dHRvbi5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBXYXJuaW5nKG9wdGlvbnM6IE1lc3NhZ2VEaWFsb2dPcHRpb25zKTogUHJvbWlzZTxzdHJpbmc+IHsgcmV0dXJuIGRpYWxvZyhEaWFsb2dXYXJuaW5nLCBvcHRpb25zKTsgfVxyXG5cclxuLyoqXHJcbiAqIFByZXNlbnRzIGFuIGVycm9yIGRpYWxvZy5cclxuICpcclxuICogQHBhcmFtIG9wdGlvbnMgLSBEaWFsb2cgb3B0aW9ucy5cclxuICogQHJldHVybnMgQSBwcm9taXNlIHRoYXQgcmVzb2x2ZXMgd2l0aCB0aGUgbGFiZWwgb2YgdGhlIGNob3NlbiBidXR0b24uXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gRXJyb3Iob3B0aW9uczogTWVzc2FnZURpYWxvZ09wdGlvbnMpOiBQcm9taXNlPHN0cmluZz4geyByZXR1cm4gZGlhbG9nKERpYWxvZ0Vycm9yLCBvcHRpb25zKTsgfVxyXG5cclxuLyoqXHJcbiAqIFByZXNlbnRzIGEgcXVlc3Rpb24gZGlhbG9nLlxyXG4gKlxyXG4gKiBAcGFyYW0gb3B0aW9ucyAtIERpYWxvZyBvcHRpb25zLlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCByZXNvbHZlcyB3aXRoIHRoZSBsYWJlbCBvZiB0aGUgY2hvc2VuIGJ1dHRvbi5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBRdWVzdGlvbihvcHRpb25zOiBNZXNzYWdlRGlhbG9nT3B0aW9ucyk6IFByb21pc2U8c3RyaW5nPiB7IHJldHVybiBkaWFsb2coRGlhbG9nUXVlc3Rpb24sIG9wdGlvbnMpOyB9XHJcblxyXG4vKipcclxuICogUHJlc2VudHMgYSBmaWxlIHNlbGVjdGlvbiBkaWFsb2cgdG8gcGljayBvbmUgb3IgbW9yZSBmaWxlcyB0byBvcGVuLlxyXG4gKlxyXG4gKiBAcGFyYW0gb3B0aW9ucyAtIERpYWxvZyBvcHRpb25zLlxyXG4gKiBAcmV0dXJucyBTZWxlY3RlZCBmaWxlIG9yIGxpc3Qgb2YgZmlsZXMsIG9yIGEgYmxhbmsgc3RyaW5nL2VtcHR5IGxpc3QgaWYgbm8gZmlsZSBoYXMgYmVlbiBzZWxlY3RlZC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBPcGVuRmlsZShvcHRpb25zOiBPcGVuRmlsZURpYWxvZ09wdGlvbnMgJiB7IEFsbG93c011bHRpcGxlU2VsZWN0aW9uOiB0cnVlIH0pOiBQcm9taXNlPHN0cmluZ1tdPjtcclxuZXhwb3J0IGZ1bmN0aW9uIE9wZW5GaWxlKG9wdGlvbnM6IE9wZW5GaWxlRGlhbG9nT3B0aW9ucyAmIHsgQWxsb3dzTXVsdGlwbGVTZWxlY3Rpb24/OiBmYWxzZSB8IHVuZGVmaW5lZCB9KTogUHJvbWlzZTxzdHJpbmc+O1xyXG5leHBvcnQgZnVuY3Rpb24gT3BlbkZpbGUob3B0aW9uczogT3BlbkZpbGVEaWFsb2dPcHRpb25zKTogUHJvbWlzZTxzdHJpbmcgfCBzdHJpbmdbXT47XHJcbmV4cG9ydCBmdW5jdGlvbiBPcGVuRmlsZShvcHRpb25zOiBPcGVuRmlsZURpYWxvZ09wdGlvbnMpOiBQcm9taXNlPHN0cmluZyB8IHN0cmluZ1tdPiB7IHJldHVybiBkaWFsb2coRGlhbG9nT3BlbkZpbGUsIG9wdGlvbnMpID8/IFtdOyB9XHJcblxyXG4vKipcclxuICogUHJlc2VudHMgYSBmaWxlIHNlbGVjdGlvbiBkaWFsb2cgdG8gcGljayBhIGZpbGUgdG8gc2F2ZS5cclxuICpcclxuICogQHBhcmFtIG9wdGlvbnMgLSBEaWFsb2cgb3B0aW9ucy5cclxuICogQHJldHVybnMgU2VsZWN0ZWQgZmlsZSwgb3IgYSBibGFuayBzdHJpbmcgaWYgbm8gZmlsZSBoYXMgYmVlbiBzZWxlY3RlZC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBTYXZlRmlsZShvcHRpb25zOiBTYXZlRmlsZURpYWxvZ09wdGlvbnMpOiBQcm9taXNlPHN0cmluZz4geyByZXR1cm4gZGlhbG9nKERpYWxvZ1NhdmVGaWxlLCBvcHRpb25zKTsgfVxyXG4iLCAiLypcclxuIF9cdCAgIF9fXHQgIF8gX19cclxufCB8XHQgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuaW1wb3J0IHsgbmV3UnVudGltZUNhbGxlciwgb2JqZWN0TmFtZXMgfSBmcm9tIFwiLi9ydW50aW1lLmpzXCI7XHJcbmltcG9ydCB7IGV2ZW50TGlzdGVuZXJzLCBMaXN0ZW5lciwgbGlzdGVuZXJPZmYgfSBmcm9tIFwiLi9saXN0ZW5lci5qc1wiO1xyXG5pbXBvcnQgeyBFdmVudHMgYXMgQ3JlYXRlIH0gZnJvbSBcIi4vY3JlYXRlLmpzXCI7XHJcbmltcG9ydCB7IFR5cGVzIH0gZnJvbSBcIi4vZXZlbnRfdHlwZXMuanNcIjtcclxuXHJcbi8vIFNldHVwXHJcbndpbmRvdy5fd2FpbHMgPSB3aW5kb3cuX3dhaWxzIHx8IHt9O1xyXG53aW5kb3cuX3dhaWxzLmRpc3BhdGNoV2FpbHNFdmVudCA9IGRpc3BhdGNoV2FpbHNFdmVudDtcclxuXHJcbmNvbnN0IGNhbGwgPSBuZXdSdW50aW1lQ2FsbGVyKG9iamVjdE5hbWVzLkV2ZW50cyk7XHJcbmNvbnN0IEVtaXRNZXRob2QgPSAwO1xyXG5cclxuZXhwb3J0ICogZnJvbSBcIi4vZXZlbnRfdHlwZXMuanNcIjtcclxuXHJcbi8qKlxyXG4gKiBBIHRhYmxlIG9mIGRhdGEgdHlwZXMgZm9yIGFsbCBrbm93biBldmVudHMuXHJcbiAqIFdpbGwgYmUgbW9ua2V5LXBhdGNoZWQgYnkgdGhlIGJpbmRpbmcgZ2VuZXJhdG9yLlxyXG4gKi9cclxuZXhwb3J0IGludGVyZmFjZSBDdXN0b21FdmVudHMge31cclxuXHJcbi8qKlxyXG4gKiBFaXRoZXIgYSBrbm93biBldmVudCBuYW1lIG9yIGFuIGFyYml0cmFyeSBzdHJpbmcuXHJcbiAqL1xyXG5leHBvcnQgdHlwZSBXYWlsc0V2ZW50TmFtZTxFIGV4dGVuZHMga2V5b2YgQ3VzdG9tRXZlbnRzID0ga2V5b2YgQ3VzdG9tRXZlbnRzPiA9IEUgfCAoc3RyaW5nICYge30pO1xyXG5cclxuLyoqXHJcbiAqIFVuaW9uIG9mIGFsbCBrbm93biBzeXN0ZW0gZXZlbnQgbmFtZXMuXHJcbiAqL1xyXG50eXBlIFN5c3RlbUV2ZW50TmFtZSA9IHtcclxuICAgIFtLIGluIGtleW9mICh0eXBlb2YgVHlwZXMpXTogKHR5cGVvZiBUeXBlcylbS11ba2V5b2YgKCh0eXBlb2YgVHlwZXMpW0tdKV1cclxufSBleHRlbmRzIChpbmZlciBNKSA/IE1ba2V5b2YgTV0gOiBuZXZlcjtcclxuXHJcbi8qKlxyXG4gKiBUaGUgZGF0YSB0eXBlIGFzc29jaWF0ZWQgdG8gYSBnaXZlbiBldmVudC5cclxuICovXHJcbmV4cG9ydCB0eXBlIFdhaWxzRXZlbnREYXRhPEUgZXh0ZW5kcyBXYWlsc0V2ZW50TmFtZSA9IFdhaWxzRXZlbnROYW1lPiA9XHJcbiAgICBFIGV4dGVuZHMga2V5b2YgQ3VzdG9tRXZlbnRzID8gQ3VzdG9tRXZlbnRzW0VdIDogKEUgZXh0ZW5kcyBTeXN0ZW1FdmVudE5hbWUgPyB2b2lkIDogYW55KTtcclxuXHJcbi8qKlxyXG4gKiBUaGUgdHlwZSBvZiBoYW5kbGVycyBmb3IgYSBnaXZlbiBldmVudC5cclxuICovXHJcbmV4cG9ydCB0eXBlIFdhaWxzRXZlbnRDYWxsYmFjazxFIGV4dGVuZHMgV2FpbHNFdmVudE5hbWUgPSBXYWlsc0V2ZW50TmFtZT4gPSAoZXY6IFdhaWxzRXZlbnQ8RT4pID0+IHZvaWQ7XHJcblxyXG4vKipcclxuICogUmVwcmVzZW50cyBhIHN5c3RlbSBldmVudCBvciBhIGN1c3RvbSBldmVudCBlbWl0dGVkIHRocm91Z2ggd2FpbHMtcHJvdmlkZWQgZmFjaWxpdGllcy5cclxuICovXHJcbmV4cG9ydCBjbGFzcyBXYWlsc0V2ZW50PEUgZXh0ZW5kcyBXYWlsc0V2ZW50TmFtZSA9IFdhaWxzRXZlbnROYW1lPiB7XHJcbiAgICAvKipcclxuICAgICAqIFRoZSBuYW1lIG9mIHRoZSBldmVudC5cclxuICAgICAqL1xyXG4gICAgbmFtZTogRTtcclxuXHJcbiAgICAvKipcclxuICAgICAqIE9wdGlvbmFsIGRhdGEgYXNzb2NpYXRlZCB3aXRoIHRoZSBlbWl0dGVkIGV2ZW50LlxyXG4gICAgICovXHJcbiAgICBkYXRhOiBXYWlsc0V2ZW50RGF0YTxFPjtcclxuXHJcbiAgICAvKipcclxuICAgICAqIE5hbWUgb2YgdGhlIG9yaWdpbmF0aW5nIHdpbmRvdy4gT21pdHRlZCBmb3IgYXBwbGljYXRpb24gZXZlbnRzLlxyXG4gICAgICogV2lsbCBiZSBvdmVycmlkZGVuIGlmIHNldCBtYW51YWxseS5cclxuICAgICAqL1xyXG4gICAgc2VuZGVyPzogc3RyaW5nO1xyXG5cclxuICAgIGNvbnN0cnVjdG9yKG5hbWU6IEUsIGRhdGE6IFdhaWxzRXZlbnREYXRhPEU+KTtcclxuICAgIGNvbnN0cnVjdG9yKG5hbWU6IFdhaWxzRXZlbnREYXRhPEU+IGV4dGVuZHMgbnVsbCB8IHZvaWQgPyBFIDogbmV2ZXIpXHJcbiAgICBjb25zdHJ1Y3RvcihuYW1lOiBFLCBkYXRhPzogYW55KSB7XHJcbiAgICAgICAgdGhpcy5uYW1lID0gbmFtZTtcclxuICAgICAgICB0aGlzLmRhdGEgPSBkYXRhID8/IG51bGw7XHJcbiAgICB9XHJcbn1cclxuXHJcbmZ1bmN0aW9uIGRpc3BhdGNoV2FpbHNFdmVudChldmVudDogYW55KSB7XHJcbiAgICBsZXQgbGlzdGVuZXJzID0gZXZlbnRMaXN0ZW5lcnMuZ2V0KGV2ZW50Lm5hbWUpO1xyXG4gICAgaWYgKCFsaXN0ZW5lcnMpIHtcclxuICAgICAgICByZXR1cm47XHJcbiAgICB9XHJcblxyXG4gICAgbGV0IHdhaWxzRXZlbnQgPSBuZXcgV2FpbHNFdmVudChcclxuICAgICAgICBldmVudC5uYW1lLFxyXG4gICAgICAgIChldmVudC5uYW1lIGluIENyZWF0ZSkgPyBDcmVhdGVbZXZlbnQubmFtZV0oZXZlbnQuZGF0YSkgOiBldmVudC5kYXRhXHJcbiAgICApO1xyXG4gICAgaWYgKCdzZW5kZXInIGluIGV2ZW50KSB7XHJcbiAgICAgICAgd2FpbHNFdmVudC5zZW5kZXIgPSBldmVudC5zZW5kZXI7XHJcbiAgICB9XHJcblxyXG4gICAgbGlzdGVuZXJzID0gbGlzdGVuZXJzLmZpbHRlcihsaXN0ZW5lciA9PiAhbGlzdGVuZXIuZGlzcGF0Y2god2FpbHNFdmVudCkpO1xyXG4gICAgaWYgKGxpc3RlbmVycy5sZW5ndGggPT09IDApIHtcclxuICAgICAgICBldmVudExpc3RlbmVycy5kZWxldGUoZXZlbnQubmFtZSk7XHJcbiAgICB9IGVsc2Uge1xyXG4gICAgICAgIGV2ZW50TGlzdGVuZXJzLnNldChldmVudC5uYW1lLCBsaXN0ZW5lcnMpO1xyXG4gICAgfVxyXG59XHJcblxyXG4vKipcclxuICogUmVnaXN0ZXIgYSBjYWxsYmFjayBmdW5jdGlvbiB0byBiZSBjYWxsZWQgbXVsdGlwbGUgdGltZXMgZm9yIGEgc3BlY2lmaWMgZXZlbnQuXHJcbiAqXHJcbiAqIEBwYXJhbSBldmVudE5hbWUgLSBUaGUgbmFtZSBvZiB0aGUgZXZlbnQgdG8gcmVnaXN0ZXIgdGhlIGNhbGxiYWNrIGZvci5cclxuICogQHBhcmFtIGNhbGxiYWNrIC0gVGhlIGNhbGxiYWNrIGZ1bmN0aW9uIHRvIGJlIGNhbGxlZCB3aGVuIHRoZSBldmVudCBpcyB0cmlnZ2VyZWQuXHJcbiAqIEBwYXJhbSBtYXhDYWxsYmFja3MgLSBUaGUgbWF4aW11bSBudW1iZXIgb2YgdGltZXMgdGhlIGNhbGxiYWNrIGNhbiBiZSBjYWxsZWQgZm9yIHRoZSBldmVudC4gT25jZSB0aGUgbWF4aW11bSBudW1iZXIgaXMgcmVhY2hlZCwgdGhlIGNhbGxiYWNrIHdpbGwgbm8gbG9uZ2VyIGJlIGNhbGxlZC5cclxuICogQHJldHVybnMgQSBmdW5jdGlvbiB0aGF0LCB3aGVuIGNhbGxlZCwgd2lsbCB1bnJlZ2lzdGVyIHRoZSBjYWxsYmFjayBmcm9tIHRoZSBldmVudC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBPbk11bHRpcGxlPEUgZXh0ZW5kcyBXYWlsc0V2ZW50TmFtZSA9IFdhaWxzRXZlbnROYW1lPihldmVudE5hbWU6IEUsIGNhbGxiYWNrOiBXYWlsc0V2ZW50Q2FsbGJhY2s8RT4sIG1heENhbGxiYWNrczogbnVtYmVyKSB7XHJcbiAgICBsZXQgbGlzdGVuZXJzID0gZXZlbnRMaXN0ZW5lcnMuZ2V0KGV2ZW50TmFtZSkgfHwgW107XHJcbiAgICBjb25zdCB0aGlzTGlzdGVuZXIgPSBuZXcgTGlzdGVuZXIoZXZlbnROYW1lLCBjYWxsYmFjaywgbWF4Q2FsbGJhY2tzKTtcclxuICAgIGxpc3RlbmVycy5wdXNoKHRoaXNMaXN0ZW5lcik7XHJcbiAgICBldmVudExpc3RlbmVycy5zZXQoZXZlbnROYW1lLCBsaXN0ZW5lcnMpO1xyXG4gICAgcmV0dXJuICgpID0+IGxpc3RlbmVyT2ZmKHRoaXNMaXN0ZW5lcik7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBSZWdpc3RlcnMgYSBjYWxsYmFjayBmdW5jdGlvbiB0byBiZSBleGVjdXRlZCB3aGVuIHRoZSBzcGVjaWZpZWQgZXZlbnQgb2NjdXJzLlxyXG4gKlxyXG4gKiBAcGFyYW0gZXZlbnROYW1lIC0gVGhlIG5hbWUgb2YgdGhlIGV2ZW50IHRvIHJlZ2lzdGVyIHRoZSBjYWxsYmFjayBmb3IuXHJcbiAqIEBwYXJhbSBjYWxsYmFjayAtIFRoZSBjYWxsYmFjayBmdW5jdGlvbiB0byBiZSBjYWxsZWQgd2hlbiB0aGUgZXZlbnQgaXMgdHJpZ2dlcmVkLlxyXG4gKiBAcmV0dXJucyBBIGZ1bmN0aW9uIHRoYXQsIHdoZW4gY2FsbGVkLCB3aWxsIHVucmVnaXN0ZXIgdGhlIGNhbGxiYWNrIGZyb20gdGhlIGV2ZW50LlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIE9uPEUgZXh0ZW5kcyBXYWlsc0V2ZW50TmFtZSA9IFdhaWxzRXZlbnROYW1lPihldmVudE5hbWU6IEUsIGNhbGxiYWNrOiBXYWlsc0V2ZW50Q2FsbGJhY2s8RT4pOiAoKSA9PiB2b2lkIHtcclxuICAgIHJldHVybiBPbk11bHRpcGxlKGV2ZW50TmFtZSwgY2FsbGJhY2ssIC0xKTtcclxufVxyXG5cclxuLyoqXHJcbiAqIFJlZ2lzdGVycyBhIGNhbGxiYWNrIGZ1bmN0aW9uIHRvIGJlIGV4ZWN1dGVkIG9ubHkgb25jZSBmb3IgdGhlIHNwZWNpZmllZCBldmVudC5cclxuICpcclxuICogQHBhcmFtIGV2ZW50TmFtZSAtIFRoZSBuYW1lIG9mIHRoZSBldmVudCB0byByZWdpc3RlciB0aGUgY2FsbGJhY2sgZm9yLlxyXG4gKiBAcGFyYW0gY2FsbGJhY2sgLSBUaGUgY2FsbGJhY2sgZnVuY3Rpb24gdG8gYmUgY2FsbGVkIHdoZW4gdGhlIGV2ZW50IGlzIHRyaWdnZXJlZC5cclxuICogQHJldHVybnMgQSBmdW5jdGlvbiB0aGF0LCB3aGVuIGNhbGxlZCwgd2lsbCB1bnJlZ2lzdGVyIHRoZSBjYWxsYmFjayBmcm9tIHRoZSBldmVudC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBPbmNlPEUgZXh0ZW5kcyBXYWlsc0V2ZW50TmFtZSA9IFdhaWxzRXZlbnROYW1lPihldmVudE5hbWU6IEUsIGNhbGxiYWNrOiBXYWlsc0V2ZW50Q2FsbGJhY2s8RT4pOiAoKSA9PiB2b2lkIHtcclxuICAgIHJldHVybiBPbk11bHRpcGxlKGV2ZW50TmFtZSwgY2FsbGJhY2ssIDEpO1xyXG59XHJcblxyXG4vKipcclxuICogUmVtb3ZlcyBldmVudCBsaXN0ZW5lcnMgZm9yIHRoZSBzcGVjaWZpZWQgZXZlbnQgbmFtZXMuXHJcbiAqXHJcbiAqIEBwYXJhbSBldmVudE5hbWVzIC0gVGhlIG5hbWUgb2YgdGhlIGV2ZW50cyB0byByZW1vdmUgbGlzdGVuZXJzIGZvci5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBPZmYoLi4uZXZlbnROYW1lczogW1dhaWxzRXZlbnROYW1lLCAuLi5XYWlsc0V2ZW50TmFtZVtdXSk6IHZvaWQge1xyXG4gICAgZXZlbnROYW1lcy5mb3JFYWNoKGV2ZW50TmFtZSA9PiBldmVudExpc3RlbmVycy5kZWxldGUoZXZlbnROYW1lKSk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBSZW1vdmVzIGFsbCBldmVudCBsaXN0ZW5lcnMuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gT2ZmQWxsKCk6IHZvaWQge1xyXG4gICAgZXZlbnRMaXN0ZW5lcnMuY2xlYXIoKTtcclxufVxyXG5cclxuLyoqXHJcbiAqIEVtaXRzIGFuIGV2ZW50LlxyXG4gKlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCB3aWxsIGJlIGZ1bGZpbGxlZCBvbmNlIHRoZSBldmVudCBoYXMgYmVlbiBlbWl0dGVkLiAgUmVzb2x2ZXMgdG8gdHJ1ZSBpZiB0aGUgZXZlbnQgd2FzIGNhbmNlbGxlZC5cclxuICogQHBhcmFtIG5hbWUgLSBUaGUgbmFtZSBvZiB0aGUgZXZlbnQgdG8gZW1pdFxyXG4gKiBAcGFyYW0gZGF0YSAtIFRoZSBkYXRhIHRoYXQgd2lsbCBiZSBzZW50IHdpdGggdGhlIGV2ZW50XHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gRW1pdDxFIGV4dGVuZHMgV2FpbHNFdmVudE5hbWUgPSBXYWlsc0V2ZW50TmFtZT4obmFtZTogRSwgZGF0YTogV2FpbHNFdmVudERhdGE8RT4pOiBQcm9taXNlPGJvb2xlYW4+XHJcbmV4cG9ydCBmdW5jdGlvbiBFbWl0PEUgZXh0ZW5kcyBXYWlsc0V2ZW50TmFtZSA9IFdhaWxzRXZlbnROYW1lPihuYW1lOiBXYWlsc0V2ZW50RGF0YTxFPiBleHRlbmRzIG51bGwgfCB2b2lkID8gRSA6IG5ldmVyKTogUHJvbWlzZTxib29sZWFuPlxyXG5leHBvcnQgZnVuY3Rpb24gRW1pdDxFIGV4dGVuZHMgV2FpbHNFdmVudE5hbWUgPSBXYWlsc0V2ZW50TmFtZT4obmFtZTogV2FpbHNFdmVudERhdGE8RT4sIGRhdGE/OiBhbnkpOiBQcm9taXNlPGJvb2xlYW4+IHtcclxuICAgIHJldHVybiBjYWxsKEVtaXRNZXRob2QsICBuZXcgV2FpbHNFdmVudChuYW1lLCBkYXRhKSlcclxufVxyXG5cclxuIiwgIi8qXHJcbiBfXHQgICBfX1x0ICBfIF9fXHJcbnwgfFx0IC8gL19fXyBfKF8pIC9fX19fXHJcbnwgfCAvfCAvIC8gX18gYC8gLyAvIF9fXy9cclxufCB8LyB8LyAvIC9fLyAvIC8gKF9fICApXHJcbnxfXy98X18vXFxfXyxfL18vXy9fX19fL1xyXG5UaGUgZWxlY3Ryb24gYWx0ZXJuYXRpdmUgZm9yIEdvXHJcbihjKSBMZWEgQW50aG9ueSAyMDE5LXByZXNlbnRcclxuKi9cclxuXHJcbi8vIFRoZSBmb2xsb3dpbmcgdXRpbGl0aWVzIGhhdmUgYmVlbiBmYWN0b3JlZCBvdXQgb2YgLi9ldmVudHMudHNcclxuLy8gZm9yIHRlc3RpbmcgcHVycG9zZXMuXHJcblxyXG5leHBvcnQgY29uc3QgZXZlbnRMaXN0ZW5lcnMgPSBuZXcgTWFwPHN0cmluZywgTGlzdGVuZXJbXT4oKTtcclxuXHJcbmV4cG9ydCBjbGFzcyBMaXN0ZW5lciB7XHJcbiAgICBldmVudE5hbWU6IHN0cmluZztcclxuICAgIGNhbGxiYWNrOiAoZGF0YTogYW55KSA9PiB2b2lkO1xyXG4gICAgbWF4Q2FsbGJhY2tzOiBudW1iZXI7XHJcblxyXG4gICAgY29uc3RydWN0b3IoZXZlbnROYW1lOiBzdHJpbmcsIGNhbGxiYWNrOiAoZGF0YTogYW55KSA9PiB2b2lkLCBtYXhDYWxsYmFja3M6IG51bWJlcikge1xyXG4gICAgICAgIHRoaXMuZXZlbnROYW1lID0gZXZlbnROYW1lO1xyXG4gICAgICAgIHRoaXMuY2FsbGJhY2sgPSBjYWxsYmFjaztcclxuICAgICAgICB0aGlzLm1heENhbGxiYWNrcyA9IG1heENhbGxiYWNrcyB8fCAtMTtcclxuICAgIH1cclxuXHJcbiAgICBkaXNwYXRjaChkYXRhOiBhbnkpOiBib29sZWFuIHtcclxuICAgICAgICB0cnkge1xyXG4gICAgICAgICAgICB0aGlzLmNhbGxiYWNrKGRhdGEpO1xyXG4gICAgICAgIH0gY2F0Y2ggKGVycikge1xyXG4gICAgICAgICAgICBjb25zb2xlLmVycm9yKGVycik7XHJcbiAgICAgICAgfVxyXG5cclxuICAgICAgICBpZiAodGhpcy5tYXhDYWxsYmFja3MgPT09IC0xKSByZXR1cm4gZmFsc2U7XHJcbiAgICAgICAgdGhpcy5tYXhDYWxsYmFja3MgLT0gMTtcclxuICAgICAgICByZXR1cm4gdGhpcy5tYXhDYWxsYmFja3MgPT09IDA7XHJcbiAgICB9XHJcbn1cclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBsaXN0ZW5lck9mZihsaXN0ZW5lcjogTGlzdGVuZXIpOiB2b2lkIHtcclxuICAgIGxldCBsaXN0ZW5lcnMgPSBldmVudExpc3RlbmVycy5nZXQobGlzdGVuZXIuZXZlbnROYW1lKTtcclxuICAgIGlmICghbGlzdGVuZXJzKSB7XHJcbiAgICAgICAgcmV0dXJuO1xyXG4gICAgfVxyXG5cclxuICAgIGxpc3RlbmVycyA9IGxpc3RlbmVycy5maWx0ZXIobCA9PiBsICE9PSBsaXN0ZW5lcik7XHJcbiAgICBpZiAobGlzdGVuZXJzLmxlbmd0aCA9PT0gMCkge1xyXG4gICAgICAgIGV2ZW50TGlzdGVuZXJzLmRlbGV0ZShsaXN0ZW5lci5ldmVudE5hbWUpO1xyXG4gICAgfSBlbHNlIHtcclxuICAgICAgICBldmVudExpc3RlbmVycy5zZXQobGlzdGVuZXIuZXZlbnROYW1lLCBsaXN0ZW5lcnMpO1xyXG4gICAgfVxyXG59XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG4vKipcclxuICogQW55IGlzIGEgZHVtbXkgY3JlYXRpb24gZnVuY3Rpb24gZm9yIHNpbXBsZSBvciB1bmtub3duIHR5cGVzLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIEFueTxUID0gYW55Pihzb3VyY2U6IGFueSk6IFQge1xyXG4gICAgcmV0dXJuIHNvdXJjZTtcclxufVxyXG5cclxuLyoqXHJcbiAqIEJ5dGVTbGljZSBpcyBhIGNyZWF0aW9uIGZ1bmN0aW9uIHRoYXQgcmVwbGFjZXNcclxuICogbnVsbCBzdHJpbmdzIHdpdGggZW1wdHkgc3RyaW5ncy5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBCeXRlU2xpY2Uoc291cmNlOiBhbnkpOiBzdHJpbmcge1xyXG4gICAgcmV0dXJuICgoc291cmNlID09IG51bGwpID8gXCJcIiA6IHNvdXJjZSk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBBcnJheSB0YWtlcyBhIGNyZWF0aW9uIGZ1bmN0aW9uIGZvciBhbiBhcmJpdHJhcnkgdHlwZVxyXG4gKiBhbmQgcmV0dXJucyBhbiBpbi1wbGFjZSBjcmVhdGlvbiBmdW5jdGlvbiBmb3IgYW4gYXJyYXlcclxuICogd2hvc2UgZWxlbWVudHMgYXJlIG9mIHRoYXQgdHlwZS5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBBcnJheTxUID0gYW55PihlbGVtZW50OiAoc291cmNlOiBhbnkpID0+IFQpOiAoc291cmNlOiBhbnkpID0+IFRbXSB7XHJcbiAgICBpZiAoZWxlbWVudCA9PT0gQW55KSB7XHJcbiAgICAgICAgcmV0dXJuIChzb3VyY2UpID0+IChzb3VyY2UgPT09IG51bGwgPyBbXSA6IHNvdXJjZSk7XHJcbiAgICB9XHJcblxyXG4gICAgcmV0dXJuIChzb3VyY2UpID0+IHtcclxuICAgICAgICBpZiAoc291cmNlID09PSBudWxsKSB7XHJcbiAgICAgICAgICAgIHJldHVybiBbXTtcclxuICAgICAgICB9XHJcbiAgICAgICAgZm9yIChsZXQgaSA9IDA7IGkgPCBzb3VyY2UubGVuZ3RoOyBpKyspIHtcclxuICAgICAgICAgICAgc291cmNlW2ldID0gZWxlbWVudChzb3VyY2VbaV0pO1xyXG4gICAgICAgIH1cclxuICAgICAgICByZXR1cm4gc291cmNlO1xyXG4gICAgfTtcclxufVxyXG5cclxuLyoqXHJcbiAqIE1hcCB0YWtlcyBjcmVhdGlvbiBmdW5jdGlvbnMgZm9yIHR3byBhcmJpdHJhcnkgdHlwZXNcclxuICogYW5kIHJldHVybnMgYW4gaW4tcGxhY2UgY3JlYXRpb24gZnVuY3Rpb24gZm9yIGFuIG9iamVjdFxyXG4gKiB3aG9zZSBrZXlzIGFuZCB2YWx1ZXMgYXJlIG9mIHRob3NlIHR5cGVzLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIE1hcDxLIGV4dGVuZHMgUHJvcGVydHlLZXkgPSBhbnksIFYgPSBhbnk+KGtleTogKHNvdXJjZTogYW55KSA9PiBLLCB2YWx1ZTogKHNvdXJjZTogYW55KSA9PiBWKTogKHNvdXJjZTogYW55KSA9PiBSZWNvcmQ8SywgVj4ge1xyXG4gICAgaWYgKHZhbHVlID09PSBBbnkpIHtcclxuICAgICAgICByZXR1cm4gKHNvdXJjZSkgPT4gKHNvdXJjZSA9PT0gbnVsbCA/IHt9IDogc291cmNlKTtcclxuICAgIH1cclxuXHJcbiAgICByZXR1cm4gKHNvdXJjZSkgPT4ge1xyXG4gICAgICAgIGlmIChzb3VyY2UgPT09IG51bGwpIHtcclxuICAgICAgICAgICAgcmV0dXJuIHt9O1xyXG4gICAgICAgIH1cclxuICAgICAgICBmb3IgKGNvbnN0IGtleSBpbiBzb3VyY2UpIHtcclxuICAgICAgICAgICAgc291cmNlW2tleV0gPSB2YWx1ZShzb3VyY2Vba2V5XSk7XHJcbiAgICAgICAgfVxyXG4gICAgICAgIHJldHVybiBzb3VyY2U7XHJcbiAgICB9O1xyXG59XHJcblxyXG4vKipcclxuICogTnVsbGFibGUgdGFrZXMgYSBjcmVhdGlvbiBmdW5jdGlvbiBmb3IgYW4gYXJiaXRyYXJ5IHR5cGVcclxuICogYW5kIHJldHVybnMgYSBjcmVhdGlvbiBmdW5jdGlvbiBmb3IgYSBudWxsYWJsZSB2YWx1ZSBvZiB0aGF0IHR5cGUuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gTnVsbGFibGU8VCA9IGFueT4oZWxlbWVudDogKHNvdXJjZTogYW55KSA9PiBUKTogKHNvdXJjZTogYW55KSA9PiAoVCB8IG51bGwpIHtcclxuICAgIGlmIChlbGVtZW50ID09PSBBbnkpIHtcclxuICAgICAgICByZXR1cm4gQW55O1xyXG4gICAgfVxyXG5cclxuICAgIHJldHVybiAoc291cmNlKSA9PiAoc291cmNlID09PSBudWxsID8gbnVsbCA6IGVsZW1lbnQoc291cmNlKSk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBTdHJ1Y3QgdGFrZXMgYW4gb2JqZWN0IG1hcHBpbmcgZmllbGQgbmFtZXMgdG8gY3JlYXRpb24gZnVuY3Rpb25zXHJcbiAqIGFuZCByZXR1cm5zIGFuIGluLXBsYWNlIGNyZWF0aW9uIGZ1bmN0aW9uIGZvciBhIHN0cnVjdC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBTdHJ1Y3QoY3JlYXRlRmllbGQ6IFJlY29yZDxzdHJpbmcsIChzb3VyY2U6IGFueSkgPT4gYW55Pik6XHJcbiAgICA8VSBleHRlbmRzIFJlY29yZDxzdHJpbmcsIGFueT4gPSBhbnk+KHNvdXJjZTogYW55KSA9PiBVXHJcbntcclxuICAgIGxldCBhbGxBbnkgPSB0cnVlO1xyXG4gICAgZm9yIChjb25zdCBuYW1lIGluIGNyZWF0ZUZpZWxkKSB7XHJcbiAgICAgICAgaWYgKGNyZWF0ZUZpZWxkW25hbWVdICE9PSBBbnkpIHtcclxuICAgICAgICAgICAgYWxsQW55ID0gZmFsc2U7XHJcbiAgICAgICAgICAgIGJyZWFrO1xyXG4gICAgICAgIH1cclxuICAgIH1cclxuICAgIGlmIChhbGxBbnkpIHtcclxuICAgICAgICByZXR1cm4gQW55O1xyXG4gICAgfVxyXG5cclxuICAgIHJldHVybiAoc291cmNlKSA9PiB7XHJcbiAgICAgICAgZm9yIChjb25zdCBuYW1lIGluIGNyZWF0ZUZpZWxkKSB7XHJcbiAgICAgICAgICAgIGlmIChuYW1lIGluIHNvdXJjZSkge1xyXG4gICAgICAgICAgICAgICAgc291cmNlW25hbWVdID0gY3JlYXRlRmllbGRbbmFtZV0oc291cmNlW25hbWVdKTtcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgIH1cclxuICAgICAgICByZXR1cm4gc291cmNlO1xyXG4gICAgfTtcclxufVxyXG5cclxuLyoqXHJcbiAqIE1hcHMga25vd24gZXZlbnQgbmFtZXMgdG8gY3JlYXRpb24gZnVuY3Rpb25zIGZvciB0aGVpciBkYXRhIHR5cGVzLlxyXG4gKiBXaWxsIGJlIG1vbmtleS1wYXRjaGVkIGJ5IHRoZSBiaW5kaW5nIGdlbmVyYXRvci5cclxuICovXHJcbmV4cG9ydCBjb25zdCBFdmVudHM6IFJlY29yZDxzdHJpbmcsIChzb3VyY2U6IGFueSkgPT4gYW55PiA9IHt9O1xyXG4iLCAiLypcclxuIF9cdCAgIF9fXHQgIF8gX19cclxufCB8XHQgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuLy8gQ3luaHlyY2h3eWQgeSBmZmVpbCBob24geW4gYXd0b21hdGlnLiBQRUlESVdDSCBcdTAwQzIgTU9ESVdMXHJcbi8vIFRoaXMgZmlsZSBpcyBhdXRvbWF0aWNhbGx5IGdlbmVyYXRlZC4gRE8gTk9UIEVESVRcclxuXHJcbmV4cG9ydCBjb25zdCBUeXBlcyA9IE9iamVjdC5mcmVlemUoe1xyXG5cdFdpbmRvd3M6IE9iamVjdC5mcmVlemUoe1xyXG5cdFx0QVBNUG93ZXJTZXR0aW5nQ2hhbmdlOiBcIndpbmRvd3M6QVBNUG93ZXJTZXR0aW5nQ2hhbmdlXCIsXHJcblx0XHRBUE1Qb3dlclN0YXR1c0NoYW5nZTogXCJ3aW5kb3dzOkFQTVBvd2VyU3RhdHVzQ2hhbmdlXCIsXHJcblx0XHRBUE1SZXN1bWVBdXRvbWF0aWM6IFwid2luZG93czpBUE1SZXN1bWVBdXRvbWF0aWNcIixcclxuXHRcdEFQTVJlc3VtZVN1c3BlbmQ6IFwid2luZG93czpBUE1SZXN1bWVTdXNwZW5kXCIsXHJcblx0XHRBUE1TdXNwZW5kOiBcIndpbmRvd3M6QVBNU3VzcGVuZFwiLFxyXG5cdFx0QXBwbGljYXRpb25TdGFydGVkOiBcIndpbmRvd3M6QXBwbGljYXRpb25TdGFydGVkXCIsXHJcblx0XHRTeXN0ZW1UaGVtZUNoYW5nZWQ6IFwid2luZG93czpTeXN0ZW1UaGVtZUNoYW5nZWRcIixcclxuXHRcdFdlYlZpZXdOYXZpZ2F0aW9uQ29tcGxldGVkOiBcIndpbmRvd3M6V2ViVmlld05hdmlnYXRpb25Db21wbGV0ZWRcIixcclxuXHRcdFdpbmRvd0FjdGl2ZTogXCJ3aW5kb3dzOldpbmRvd0FjdGl2ZVwiLFxyXG5cdFx0V2luZG93QmFja2dyb3VuZEVyYXNlOiBcIndpbmRvd3M6V2luZG93QmFja2dyb3VuZEVyYXNlXCIsXHJcblx0XHRXaW5kb3dDbGlja0FjdGl2ZTogXCJ3aW5kb3dzOldpbmRvd0NsaWNrQWN0aXZlXCIsXHJcblx0XHRXaW5kb3dDbG9zaW5nOiBcIndpbmRvd3M6V2luZG93Q2xvc2luZ1wiLFxyXG5cdFx0V2luZG93RGlkTW92ZTogXCJ3aW5kb3dzOldpbmRvd0RpZE1vdmVcIixcclxuXHRcdFdpbmRvd0RpZFJlc2l6ZTogXCJ3aW5kb3dzOldpbmRvd0RpZFJlc2l6ZVwiLFxyXG5cdFx0V2luZG93RFBJQ2hhbmdlZDogXCJ3aW5kb3dzOldpbmRvd0RQSUNoYW5nZWRcIixcclxuXHRcdFdpbmRvd0RyYWdEcm9wOiBcIndpbmRvd3M6V2luZG93RHJhZ0Ryb3BcIixcclxuXHRcdFdpbmRvd0RyYWdFbnRlcjogXCJ3aW5kb3dzOldpbmRvd0RyYWdFbnRlclwiLFxyXG5cdFx0V2luZG93RHJhZ0xlYXZlOiBcIndpbmRvd3M6V2luZG93RHJhZ0xlYXZlXCIsXHJcblx0XHRXaW5kb3dEcmFnT3ZlcjogXCJ3aW5kb3dzOldpbmRvd0RyYWdPdmVyXCIsXHJcblx0XHRXaW5kb3dFbmRNb3ZlOiBcIndpbmRvd3M6V2luZG93RW5kTW92ZVwiLFxyXG5cdFx0V2luZG93RW5kUmVzaXplOiBcIndpbmRvd3M6V2luZG93RW5kUmVzaXplXCIsXHJcblx0XHRXaW5kb3dGdWxsc2NyZWVuOiBcIndpbmRvd3M6V2luZG93RnVsbHNjcmVlblwiLFxyXG5cdFx0V2luZG93SGlkZTogXCJ3aW5kb3dzOldpbmRvd0hpZGVcIixcclxuXHRcdFdpbmRvd0luYWN0aXZlOiBcIndpbmRvd3M6V2luZG93SW5hY3RpdmVcIixcclxuXHRcdFdpbmRvd0tleURvd246IFwid2luZG93czpXaW5kb3dLZXlEb3duXCIsXHJcblx0XHRXaW5kb3dLZXlVcDogXCJ3aW5kb3dzOldpbmRvd0tleVVwXCIsXHJcblx0XHRXaW5kb3dLaWxsRm9jdXM6IFwid2luZG93czpXaW5kb3dLaWxsRm9jdXNcIixcclxuXHRcdFdpbmRvd05vbkNsaWVudEhpdDogXCJ3aW5kb3dzOldpbmRvd05vbkNsaWVudEhpdFwiLFxyXG5cdFx0V2luZG93Tm9uQ2xpZW50TW91c2VEb3duOiBcIndpbmRvd3M6V2luZG93Tm9uQ2xpZW50TW91c2VEb3duXCIsXHJcblx0XHRXaW5kb3dOb25DbGllbnRNb3VzZUxlYXZlOiBcIndpbmRvd3M6V2luZG93Tm9uQ2xpZW50TW91c2VMZWF2ZVwiLFxyXG5cdFx0V2luZG93Tm9uQ2xpZW50TW91c2VNb3ZlOiBcIndpbmRvd3M6V2luZG93Tm9uQ2xpZW50TW91c2VNb3ZlXCIsXHJcblx0XHRXaW5kb3dOb25DbGllbnRNb3VzZVVwOiBcIndpbmRvd3M6V2luZG93Tm9uQ2xpZW50TW91c2VVcFwiLFxyXG5cdFx0V2luZG93UGFpbnQ6IFwid2luZG93czpXaW5kb3dQYWludFwiLFxyXG5cdFx0V2luZG93UmVzdG9yZTogXCJ3aW5kb3dzOldpbmRvd1Jlc3RvcmVcIixcclxuXHRcdFdpbmRvd1NldEZvY3VzOiBcIndpbmRvd3M6V2luZG93U2V0Rm9jdXNcIixcclxuXHRcdFdpbmRvd1Nob3c6IFwid2luZG93czpXaW5kb3dTaG93XCIsXHJcblx0XHRXaW5kb3dTdGFydE1vdmU6IFwid2luZG93czpXaW5kb3dTdGFydE1vdmVcIixcclxuXHRcdFdpbmRvd1N0YXJ0UmVzaXplOiBcIndpbmRvd3M6V2luZG93U3RhcnRSZXNpemVcIixcclxuXHRcdFdpbmRvd1VuRnVsbHNjcmVlbjogXCJ3aW5kb3dzOldpbmRvd1VuRnVsbHNjcmVlblwiLFxyXG5cdFx0V2luZG93Wk9yZGVyQ2hhbmdlZDogXCJ3aW5kb3dzOldpbmRvd1pPcmRlckNoYW5nZWRcIixcclxuXHRcdFdpbmRvd01pbmltaXNlOiBcIndpbmRvd3M6V2luZG93TWluaW1pc2VcIixcclxuXHRcdFdpbmRvd1VuTWluaW1pc2U6IFwid2luZG93czpXaW5kb3dVbk1pbmltaXNlXCIsXHJcblx0XHRXaW5kb3dNYXhpbWlzZTogXCJ3aW5kb3dzOldpbmRvd01heGltaXNlXCIsXHJcblx0XHRXaW5kb3dVbk1heGltaXNlOiBcIndpbmRvd3M6V2luZG93VW5NYXhpbWlzZVwiLFxyXG5cdH0pLFxyXG5cdE1hYzogT2JqZWN0LmZyZWV6ZSh7XHJcblx0XHRBcHBsaWNhdGlvbkRpZEJlY29tZUFjdGl2ZTogXCJtYWM6QXBwbGljYXRpb25EaWRCZWNvbWVBY3RpdmVcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkQ2hhbmdlQmFja2luZ1Byb3BlcnRpZXM6IFwibWFjOkFwcGxpY2F0aW9uRGlkQ2hhbmdlQmFja2luZ1Byb3BlcnRpZXNcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkQ2hhbmdlRWZmZWN0aXZlQXBwZWFyYW5jZTogXCJtYWM6QXBwbGljYXRpb25EaWRDaGFuZ2VFZmZlY3RpdmVBcHBlYXJhbmNlXCIsXHJcblx0XHRBcHBsaWNhdGlvbkRpZENoYW5nZUljb246IFwibWFjOkFwcGxpY2F0aW9uRGlkQ2hhbmdlSWNvblwiLFxyXG5cdFx0QXBwbGljYXRpb25EaWRDaGFuZ2VPY2NsdXNpb25TdGF0ZTogXCJtYWM6QXBwbGljYXRpb25EaWRDaGFuZ2VPY2NsdXNpb25TdGF0ZVwiLFxyXG5cdFx0QXBwbGljYXRpb25EaWRDaGFuZ2VTY3JlZW5QYXJhbWV0ZXJzOiBcIm1hYzpBcHBsaWNhdGlvbkRpZENoYW5nZVNjcmVlblBhcmFtZXRlcnNcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkQ2hhbmdlU3RhdHVzQmFyRnJhbWU6IFwibWFjOkFwcGxpY2F0aW9uRGlkQ2hhbmdlU3RhdHVzQmFyRnJhbWVcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkQ2hhbmdlU3RhdHVzQmFyT3JpZW50YXRpb246IFwibWFjOkFwcGxpY2F0aW9uRGlkQ2hhbmdlU3RhdHVzQmFyT3JpZW50YXRpb25cIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkQ2hhbmdlVGhlbWU6IFwibWFjOkFwcGxpY2F0aW9uRGlkQ2hhbmdlVGhlbWVcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkRmluaXNoTGF1bmNoaW5nOiBcIm1hYzpBcHBsaWNhdGlvbkRpZEZpbmlzaExhdW5jaGluZ1wiLFxyXG5cdFx0QXBwbGljYXRpb25EaWRIaWRlOiBcIm1hYzpBcHBsaWNhdGlvbkRpZEhpZGVcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkUmVzaWduQWN0aXZlOiBcIm1hYzpBcHBsaWNhdGlvbkRpZFJlc2lnbkFjdGl2ZVwiLFxyXG5cdFx0QXBwbGljYXRpb25EaWRVbmhpZGU6IFwibWFjOkFwcGxpY2F0aW9uRGlkVW5oaWRlXCIsXHJcblx0XHRBcHBsaWNhdGlvbkRpZFVwZGF0ZTogXCJtYWM6QXBwbGljYXRpb25EaWRVcGRhdGVcIixcclxuXHRcdEFwcGxpY2F0aW9uU2hvdWxkSGFuZGxlUmVvcGVuOiBcIm1hYzpBcHBsaWNhdGlvblNob3VsZEhhbmRsZVJlb3BlblwiLFxyXG5cdFx0QXBwbGljYXRpb25XaWxsQmVjb21lQWN0aXZlOiBcIm1hYzpBcHBsaWNhdGlvbldpbGxCZWNvbWVBY3RpdmVcIixcclxuXHRcdEFwcGxpY2F0aW9uV2lsbEZpbmlzaExhdW5jaGluZzogXCJtYWM6QXBwbGljYXRpb25XaWxsRmluaXNoTGF1bmNoaW5nXCIsXHJcblx0XHRBcHBsaWNhdGlvbldpbGxIaWRlOiBcIm1hYzpBcHBsaWNhdGlvbldpbGxIaWRlXCIsXHJcblx0XHRBcHBsaWNhdGlvbldpbGxSZXNpZ25BY3RpdmU6IFwibWFjOkFwcGxpY2F0aW9uV2lsbFJlc2lnbkFjdGl2ZVwiLFxyXG5cdFx0QXBwbGljYXRpb25XaWxsVGVybWluYXRlOiBcIm1hYzpBcHBsaWNhdGlvbldpbGxUZXJtaW5hdGVcIixcclxuXHRcdEFwcGxpY2F0aW9uV2lsbFVuaGlkZTogXCJtYWM6QXBwbGljYXRpb25XaWxsVW5oaWRlXCIsXHJcblx0XHRBcHBsaWNhdGlvbldpbGxVcGRhdGU6IFwibWFjOkFwcGxpY2F0aW9uV2lsbFVwZGF0ZVwiLFxyXG5cdFx0TWVudURpZEFkZEl0ZW06IFwibWFjOk1lbnVEaWRBZGRJdGVtXCIsXHJcblx0XHRNZW51RGlkQmVnaW5UcmFja2luZzogXCJtYWM6TWVudURpZEJlZ2luVHJhY2tpbmdcIixcclxuXHRcdE1lbnVEaWRDbG9zZTogXCJtYWM6TWVudURpZENsb3NlXCIsXHJcblx0XHRNZW51RGlkRGlzcGxheUl0ZW06IFwibWFjOk1lbnVEaWREaXNwbGF5SXRlbVwiLFxyXG5cdFx0TWVudURpZEVuZFRyYWNraW5nOiBcIm1hYzpNZW51RGlkRW5kVHJhY2tpbmdcIixcclxuXHRcdE1lbnVEaWRIaWdobGlnaHRJdGVtOiBcIm1hYzpNZW51RGlkSGlnaGxpZ2h0SXRlbVwiLFxyXG5cdFx0TWVudURpZE9wZW46IFwibWFjOk1lbnVEaWRPcGVuXCIsXHJcblx0XHRNZW51RGlkUG9wVXA6IFwibWFjOk1lbnVEaWRQb3BVcFwiLFxyXG5cdFx0TWVudURpZFJlbW92ZUl0ZW06IFwibWFjOk1lbnVEaWRSZW1vdmVJdGVtXCIsXHJcblx0XHRNZW51RGlkU2VuZEFjdGlvbjogXCJtYWM6TWVudURpZFNlbmRBY3Rpb25cIixcclxuXHRcdE1lbnVEaWRTZW5kQWN0aW9uVG9JdGVtOiBcIm1hYzpNZW51RGlkU2VuZEFjdGlvblRvSXRlbVwiLFxyXG5cdFx0TWVudURpZFVwZGF0ZTogXCJtYWM6TWVudURpZFVwZGF0ZVwiLFxyXG5cdFx0TWVudVdpbGxBZGRJdGVtOiBcIm1hYzpNZW51V2lsbEFkZEl0ZW1cIixcclxuXHRcdE1lbnVXaWxsQmVnaW5UcmFja2luZzogXCJtYWM6TWVudVdpbGxCZWdpblRyYWNraW5nXCIsXHJcblx0XHRNZW51V2lsbERpc3BsYXlJdGVtOiBcIm1hYzpNZW51V2lsbERpc3BsYXlJdGVtXCIsXHJcblx0XHRNZW51V2lsbEVuZFRyYWNraW5nOiBcIm1hYzpNZW51V2lsbEVuZFRyYWNraW5nXCIsXHJcblx0XHRNZW51V2lsbEhpZ2hsaWdodEl0ZW06IFwibWFjOk1lbnVXaWxsSGlnaGxpZ2h0SXRlbVwiLFxyXG5cdFx0TWVudVdpbGxPcGVuOiBcIm1hYzpNZW51V2lsbE9wZW5cIixcclxuXHRcdE1lbnVXaWxsUG9wVXA6IFwibWFjOk1lbnVXaWxsUG9wVXBcIixcclxuXHRcdE1lbnVXaWxsUmVtb3ZlSXRlbTogXCJtYWM6TWVudVdpbGxSZW1vdmVJdGVtXCIsXHJcblx0XHRNZW51V2lsbFNlbmRBY3Rpb246IFwibWFjOk1lbnVXaWxsU2VuZEFjdGlvblwiLFxyXG5cdFx0TWVudVdpbGxTZW5kQWN0aW9uVG9JdGVtOiBcIm1hYzpNZW51V2lsbFNlbmRBY3Rpb25Ub0l0ZW1cIixcclxuXHRcdE1lbnVXaWxsVXBkYXRlOiBcIm1hYzpNZW51V2lsbFVwZGF0ZVwiLFxyXG5cdFx0V2ViVmlld0RpZENvbW1pdE5hdmlnYXRpb246IFwibWFjOldlYlZpZXdEaWRDb21taXROYXZpZ2F0aW9uXCIsXHJcblx0XHRXZWJWaWV3RGlkRmluaXNoTmF2aWdhdGlvbjogXCJtYWM6V2ViVmlld0RpZEZpbmlzaE5hdmlnYXRpb25cIixcclxuXHRcdFdlYlZpZXdEaWRSZWNlaXZlU2VydmVyUmVkaXJlY3RGb3JQcm92aXNpb25hbE5hdmlnYXRpb246IFwibWFjOldlYlZpZXdEaWRSZWNlaXZlU2VydmVyUmVkaXJlY3RGb3JQcm92aXNpb25hbE5hdmlnYXRpb25cIixcclxuXHRcdFdlYlZpZXdEaWRTdGFydFByb3Zpc2lvbmFsTmF2aWdhdGlvbjogXCJtYWM6V2ViVmlld0RpZFN0YXJ0UHJvdmlzaW9uYWxOYXZpZ2F0aW9uXCIsXHJcblx0XHRXaW5kb3dEaWRCZWNvbWVLZXk6IFwibWFjOldpbmRvd0RpZEJlY29tZUtleVwiLFxyXG5cdFx0V2luZG93RGlkQmVjb21lTWFpbjogXCJtYWM6V2luZG93RGlkQmVjb21lTWFpblwiLFxyXG5cdFx0V2luZG93RGlkQmVnaW5TaGVldDogXCJtYWM6V2luZG93RGlkQmVnaW5TaGVldFwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlQWxwaGE6IFwibWFjOldpbmRvd0RpZENoYW5nZUFscGhhXCIsXHJcblx0XHRXaW5kb3dEaWRDaGFuZ2VCYWNraW5nTG9jYXRpb246IFwibWFjOldpbmRvd0RpZENoYW5nZUJhY2tpbmdMb2NhdGlvblwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlQmFja2luZ1Byb3BlcnRpZXM6IFwibWFjOldpbmRvd0RpZENoYW5nZUJhY2tpbmdQcm9wZXJ0aWVzXCIsXHJcblx0XHRXaW5kb3dEaWRDaGFuZ2VDb2xsZWN0aW9uQmVoYXZpb3I6IFwibWFjOldpbmRvd0RpZENoYW5nZUNvbGxlY3Rpb25CZWhhdmlvclwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlRWZmZWN0aXZlQXBwZWFyYW5jZTogXCJtYWM6V2luZG93RGlkQ2hhbmdlRWZmZWN0aXZlQXBwZWFyYW5jZVwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlT2NjbHVzaW9uU3RhdGU6IFwibWFjOldpbmRvd0RpZENoYW5nZU9jY2x1c2lvblN0YXRlXCIsXHJcblx0XHRXaW5kb3dEaWRDaGFuZ2VPcmRlcmluZ01vZGU6IFwibWFjOldpbmRvd0RpZENoYW5nZU9yZGVyaW5nTW9kZVwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlU2NyZWVuOiBcIm1hYzpXaW5kb3dEaWRDaGFuZ2VTY3JlZW5cIixcclxuXHRcdFdpbmRvd0RpZENoYW5nZVNjcmVlblBhcmFtZXRlcnM6IFwibWFjOldpbmRvd0RpZENoYW5nZVNjcmVlblBhcmFtZXRlcnNcIixcclxuXHRcdFdpbmRvd0RpZENoYW5nZVNjcmVlblByb2ZpbGU6IFwibWFjOldpbmRvd0RpZENoYW5nZVNjcmVlblByb2ZpbGVcIixcclxuXHRcdFdpbmRvd0RpZENoYW5nZVNjcmVlblNwYWNlOiBcIm1hYzpXaW5kb3dEaWRDaGFuZ2VTY3JlZW5TcGFjZVwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlU2NyZWVuU3BhY2VQcm9wZXJ0aWVzOiBcIm1hYzpXaW5kb3dEaWRDaGFuZ2VTY3JlZW5TcGFjZVByb3BlcnRpZXNcIixcclxuXHRcdFdpbmRvd0RpZENoYW5nZVNoYXJpbmdUeXBlOiBcIm1hYzpXaW5kb3dEaWRDaGFuZ2VTaGFyaW5nVHlwZVwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlU3BhY2U6IFwibWFjOldpbmRvd0RpZENoYW5nZVNwYWNlXCIsXHJcblx0XHRXaW5kb3dEaWRDaGFuZ2VTcGFjZU9yZGVyaW5nTW9kZTogXCJtYWM6V2luZG93RGlkQ2hhbmdlU3BhY2VPcmRlcmluZ01vZGVcIixcclxuXHRcdFdpbmRvd0RpZENoYW5nZVRpdGxlOiBcIm1hYzpXaW5kb3dEaWRDaGFuZ2VUaXRsZVwiLFxyXG5cdFx0V2luZG93RGlkQ2hhbmdlVG9vbGJhcjogXCJtYWM6V2luZG93RGlkQ2hhbmdlVG9vbGJhclwiLFxyXG5cdFx0V2luZG93RGlkRGVtaW5pYXR1cml6ZTogXCJtYWM6V2luZG93RGlkRGVtaW5pYXR1cml6ZVwiLFxyXG5cdFx0V2luZG93RGlkRW5kU2hlZXQ6IFwibWFjOldpbmRvd0RpZEVuZFNoZWV0XCIsXHJcblx0XHRXaW5kb3dEaWRFbnRlckZ1bGxTY3JlZW46IFwibWFjOldpbmRvd0RpZEVudGVyRnVsbFNjcmVlblwiLFxyXG5cdFx0V2luZG93RGlkRW50ZXJWZXJzaW9uQnJvd3NlcjogXCJtYWM6V2luZG93RGlkRW50ZXJWZXJzaW9uQnJvd3NlclwiLFxyXG5cdFx0V2luZG93RGlkRXhpdEZ1bGxTY3JlZW46IFwibWFjOldpbmRvd0RpZEV4aXRGdWxsU2NyZWVuXCIsXHJcblx0XHRXaW5kb3dEaWRFeGl0VmVyc2lvbkJyb3dzZXI6IFwibWFjOldpbmRvd0RpZEV4aXRWZXJzaW9uQnJvd3NlclwiLFxyXG5cdFx0V2luZG93RGlkRXhwb3NlOiBcIm1hYzpXaW5kb3dEaWRFeHBvc2VcIixcclxuXHRcdFdpbmRvd0RpZEZvY3VzOiBcIm1hYzpXaW5kb3dEaWRGb2N1c1wiLFxyXG5cdFx0V2luZG93RGlkTWluaWF0dXJpemU6IFwibWFjOldpbmRvd0RpZE1pbmlhdHVyaXplXCIsXHJcblx0XHRXaW5kb3dEaWRNb3ZlOiBcIm1hYzpXaW5kb3dEaWRNb3ZlXCIsXHJcblx0XHRXaW5kb3dEaWRPcmRlck9mZlNjcmVlbjogXCJtYWM6V2luZG93RGlkT3JkZXJPZmZTY3JlZW5cIixcclxuXHRcdFdpbmRvd0RpZE9yZGVyT25TY3JlZW46IFwibWFjOldpbmRvd0RpZE9yZGVyT25TY3JlZW5cIixcclxuXHRcdFdpbmRvd0RpZFJlc2lnbktleTogXCJtYWM6V2luZG93RGlkUmVzaWduS2V5XCIsXHJcblx0XHRXaW5kb3dEaWRSZXNpZ25NYWluOiBcIm1hYzpXaW5kb3dEaWRSZXNpZ25NYWluXCIsXHJcblx0XHRXaW5kb3dEaWRSZXNpemU6IFwibWFjOldpbmRvd0RpZFJlc2l6ZVwiLFxyXG5cdFx0V2luZG93RGlkVXBkYXRlOiBcIm1hYzpXaW5kb3dEaWRVcGRhdGVcIixcclxuXHRcdFdpbmRvd0RpZFVwZGF0ZUFscGhhOiBcIm1hYzpXaW5kb3dEaWRVcGRhdGVBbHBoYVwiLFxyXG5cdFx0V2luZG93RGlkVXBkYXRlQ29sbGVjdGlvbkJlaGF2aW9yOiBcIm1hYzpXaW5kb3dEaWRVcGRhdGVDb2xsZWN0aW9uQmVoYXZpb3JcIixcclxuXHRcdFdpbmRvd0RpZFVwZGF0ZUNvbGxlY3Rpb25Qcm9wZXJ0aWVzOiBcIm1hYzpXaW5kb3dEaWRVcGRhdGVDb2xsZWN0aW9uUHJvcGVydGllc1wiLFxyXG5cdFx0V2luZG93RGlkVXBkYXRlU2hhZG93OiBcIm1hYzpXaW5kb3dEaWRVcGRhdGVTaGFkb3dcIixcclxuXHRcdFdpbmRvd0RpZFVwZGF0ZVRpdGxlOiBcIm1hYzpXaW5kb3dEaWRVcGRhdGVUaXRsZVwiLFxyXG5cdFx0V2luZG93RGlkVXBkYXRlVG9vbGJhcjogXCJtYWM6V2luZG93RGlkVXBkYXRlVG9vbGJhclwiLFxyXG5cdFx0V2luZG93RGlkWm9vbTogXCJtYWM6V2luZG93RGlkWm9vbVwiLFxyXG5cdFx0V2luZG93RmlsZURyYWdnaW5nRW50ZXJlZDogXCJtYWM6V2luZG93RmlsZURyYWdnaW5nRW50ZXJlZFwiLFxyXG5cdFx0V2luZG93RmlsZURyYWdnaW5nRXhpdGVkOiBcIm1hYzpXaW5kb3dGaWxlRHJhZ2dpbmdFeGl0ZWRcIixcclxuXHRcdFdpbmRvd0ZpbGVEcmFnZ2luZ1BlcmZvcm1lZDogXCJtYWM6V2luZG93RmlsZURyYWdnaW5nUGVyZm9ybWVkXCIsXHJcblx0XHRXaW5kb3dIaWRlOiBcIm1hYzpXaW5kb3dIaWRlXCIsXHJcblx0XHRXaW5kb3dNYXhpbWlzZTogXCJtYWM6V2luZG93TWF4aW1pc2VcIixcclxuXHRcdFdpbmRvd1VuTWF4aW1pc2U6IFwibWFjOldpbmRvd1VuTWF4aW1pc2VcIixcclxuXHRcdFdpbmRvd01pbmltaXNlOiBcIm1hYzpXaW5kb3dNaW5pbWlzZVwiLFxyXG5cdFx0V2luZG93VW5NaW5pbWlzZTogXCJtYWM6V2luZG93VW5NaW5pbWlzZVwiLFxyXG5cdFx0V2luZG93U2hvdWxkQ2xvc2U6IFwibWFjOldpbmRvd1Nob3VsZENsb3NlXCIsXHJcblx0XHRXaW5kb3dTaG93OiBcIm1hYzpXaW5kb3dTaG93XCIsXHJcblx0XHRXaW5kb3dXaWxsQmVjb21lS2V5OiBcIm1hYzpXaW5kb3dXaWxsQmVjb21lS2V5XCIsXHJcblx0XHRXaW5kb3dXaWxsQmVjb21lTWFpbjogXCJtYWM6V2luZG93V2lsbEJlY29tZU1haW5cIixcclxuXHRcdFdpbmRvd1dpbGxCZWdpblNoZWV0OiBcIm1hYzpXaW5kb3dXaWxsQmVnaW5TaGVldFwiLFxyXG5cdFx0V2luZG93V2lsbENoYW5nZU9yZGVyaW5nTW9kZTogXCJtYWM6V2luZG93V2lsbENoYW5nZU9yZGVyaW5nTW9kZVwiLFxyXG5cdFx0V2luZG93V2lsbENsb3NlOiBcIm1hYzpXaW5kb3dXaWxsQ2xvc2VcIixcclxuXHRcdFdpbmRvd1dpbGxEZW1pbmlhdHVyaXplOiBcIm1hYzpXaW5kb3dXaWxsRGVtaW5pYXR1cml6ZVwiLFxyXG5cdFx0V2luZG93V2lsbEVudGVyRnVsbFNjcmVlbjogXCJtYWM6V2luZG93V2lsbEVudGVyRnVsbFNjcmVlblwiLFxyXG5cdFx0V2luZG93V2lsbEVudGVyVmVyc2lvbkJyb3dzZXI6IFwibWFjOldpbmRvd1dpbGxFbnRlclZlcnNpb25Ccm93c2VyXCIsXHJcblx0XHRXaW5kb3dXaWxsRXhpdEZ1bGxTY3JlZW46IFwibWFjOldpbmRvd1dpbGxFeGl0RnVsbFNjcmVlblwiLFxyXG5cdFx0V2luZG93V2lsbEV4aXRWZXJzaW9uQnJvd3NlcjogXCJtYWM6V2luZG93V2lsbEV4aXRWZXJzaW9uQnJvd3NlclwiLFxyXG5cdFx0V2luZG93V2lsbEZvY3VzOiBcIm1hYzpXaW5kb3dXaWxsRm9jdXNcIixcclxuXHRcdFdpbmRvd1dpbGxNaW5pYXR1cml6ZTogXCJtYWM6V2luZG93V2lsbE1pbmlhdHVyaXplXCIsXHJcblx0XHRXaW5kb3dXaWxsTW92ZTogXCJtYWM6V2luZG93V2lsbE1vdmVcIixcclxuXHRcdFdpbmRvd1dpbGxPcmRlck9mZlNjcmVlbjogXCJtYWM6V2luZG93V2lsbE9yZGVyT2ZmU2NyZWVuXCIsXHJcblx0XHRXaW5kb3dXaWxsT3JkZXJPblNjcmVlbjogXCJtYWM6V2luZG93V2lsbE9yZGVyT25TY3JlZW5cIixcclxuXHRcdFdpbmRvd1dpbGxSZXNpZ25NYWluOiBcIm1hYzpXaW5kb3dXaWxsUmVzaWduTWFpblwiLFxyXG5cdFx0V2luZG93V2lsbFJlc2l6ZTogXCJtYWM6V2luZG93V2lsbFJlc2l6ZVwiLFxyXG5cdFx0V2luZG93V2lsbFVuZm9jdXM6IFwibWFjOldpbmRvd1dpbGxVbmZvY3VzXCIsXHJcblx0XHRXaW5kb3dXaWxsVXBkYXRlOiBcIm1hYzpXaW5kb3dXaWxsVXBkYXRlXCIsXHJcblx0XHRXaW5kb3dXaWxsVXBkYXRlQWxwaGE6IFwibWFjOldpbmRvd1dpbGxVcGRhdGVBbHBoYVwiLFxyXG5cdFx0V2luZG93V2lsbFVwZGF0ZUNvbGxlY3Rpb25CZWhhdmlvcjogXCJtYWM6V2luZG93V2lsbFVwZGF0ZUNvbGxlY3Rpb25CZWhhdmlvclwiLFxyXG5cdFx0V2luZG93V2lsbFVwZGF0ZUNvbGxlY3Rpb25Qcm9wZXJ0aWVzOiBcIm1hYzpXaW5kb3dXaWxsVXBkYXRlQ29sbGVjdGlvblByb3BlcnRpZXNcIixcclxuXHRcdFdpbmRvd1dpbGxVcGRhdGVTaGFkb3c6IFwibWFjOldpbmRvd1dpbGxVcGRhdGVTaGFkb3dcIixcclxuXHRcdFdpbmRvd1dpbGxVcGRhdGVUaXRsZTogXCJtYWM6V2luZG93V2lsbFVwZGF0ZVRpdGxlXCIsXHJcblx0XHRXaW5kb3dXaWxsVXBkYXRlVG9vbGJhcjogXCJtYWM6V2luZG93V2lsbFVwZGF0ZVRvb2xiYXJcIixcclxuXHRcdFdpbmRvd1dpbGxVcGRhdGVWaXNpYmlsaXR5OiBcIm1hYzpXaW5kb3dXaWxsVXBkYXRlVmlzaWJpbGl0eVwiLFxyXG5cdFx0V2luZG93V2lsbFVzZVN0YW5kYXJkRnJhbWU6IFwibWFjOldpbmRvd1dpbGxVc2VTdGFuZGFyZEZyYW1lXCIsXHJcblx0XHRXaW5kb3dab29tSW46IFwibWFjOldpbmRvd1pvb21JblwiLFxyXG5cdFx0V2luZG93Wm9vbU91dDogXCJtYWM6V2luZG93Wm9vbU91dFwiLFxyXG5cdFx0V2luZG93Wm9vbVJlc2V0OiBcIm1hYzpXaW5kb3dab29tUmVzZXRcIixcclxuXHR9KSxcclxuXHRMaW51eDogT2JqZWN0LmZyZWV6ZSh7XHJcblx0XHRBcHBsaWNhdGlvblN0YXJ0dXA6IFwibGludXg6QXBwbGljYXRpb25TdGFydHVwXCIsXHJcblx0XHRTeXN0ZW1UaGVtZUNoYW5nZWQ6IFwibGludXg6U3lzdGVtVGhlbWVDaGFuZ2VkXCIsXHJcblx0XHRXaW5kb3dEZWxldGVFdmVudDogXCJsaW51eDpXaW5kb3dEZWxldGVFdmVudFwiLFxyXG5cdFx0V2luZG93RGlkTW92ZTogXCJsaW51eDpXaW5kb3dEaWRNb3ZlXCIsXHJcblx0XHRXaW5kb3dEaWRSZXNpemU6IFwibGludXg6V2luZG93RGlkUmVzaXplXCIsXHJcblx0XHRXaW5kb3dGb2N1c0luOiBcImxpbnV4OldpbmRvd0ZvY3VzSW5cIixcclxuXHRcdFdpbmRvd0ZvY3VzT3V0OiBcImxpbnV4OldpbmRvd0ZvY3VzT3V0XCIsXHJcblx0XHRXaW5kb3dMb2FkU3RhcnRlZDogXCJsaW51eDpXaW5kb3dMb2FkU3RhcnRlZFwiLFxyXG5cdFx0V2luZG93TG9hZFJlZGlyZWN0ZWQ6IFwibGludXg6V2luZG93TG9hZFJlZGlyZWN0ZWRcIixcclxuXHRcdFdpbmRvd0xvYWRDb21taXR0ZWQ6IFwibGludXg6V2luZG93TG9hZENvbW1pdHRlZFwiLFxyXG5cdFx0V2luZG93TG9hZEZpbmlzaGVkOiBcImxpbnV4OldpbmRvd0xvYWRGaW5pc2hlZFwiLFxyXG5cdH0pLFxyXG5cdGlPUzogT2JqZWN0LmZyZWV6ZSh7XHJcblx0XHRBcHBsaWNhdGlvbkRpZEJlY29tZUFjdGl2ZTogXCJpb3M6QXBwbGljYXRpb25EaWRCZWNvbWVBY3RpdmVcIixcclxuXHRcdEFwcGxpY2F0aW9uRGlkRW50ZXJCYWNrZ3JvdW5kOiBcImlvczpBcHBsaWNhdGlvbkRpZEVudGVyQmFja2dyb3VuZFwiLFxyXG5cdFx0QXBwbGljYXRpb25EaWRGaW5pc2hMYXVuY2hpbmc6IFwiaW9zOkFwcGxpY2F0aW9uRGlkRmluaXNoTGF1bmNoaW5nXCIsXHJcblx0XHRBcHBsaWNhdGlvbkRpZFJlY2VpdmVNZW1vcnlXYXJuaW5nOiBcImlvczpBcHBsaWNhdGlvbkRpZFJlY2VpdmVNZW1vcnlXYXJuaW5nXCIsXHJcblx0XHRBcHBsaWNhdGlvbldpbGxFbnRlckZvcmVncm91bmQ6IFwiaW9zOkFwcGxpY2F0aW9uV2lsbEVudGVyRm9yZWdyb3VuZFwiLFxyXG5cdFx0QXBwbGljYXRpb25XaWxsUmVzaWduQWN0aXZlOiBcImlvczpBcHBsaWNhdGlvbldpbGxSZXNpZ25BY3RpdmVcIixcclxuXHRcdEFwcGxpY2F0aW9uV2lsbFRlcm1pbmF0ZTogXCJpb3M6QXBwbGljYXRpb25XaWxsVGVybWluYXRlXCIsXHJcblx0XHRXaW5kb3dEaWRMb2FkOiBcImlvczpXaW5kb3dEaWRMb2FkXCIsXHJcblx0XHRXaW5kb3dXaWxsQXBwZWFyOiBcImlvczpXaW5kb3dXaWxsQXBwZWFyXCIsXHJcblx0XHRXaW5kb3dEaWRBcHBlYXI6IFwiaW9zOldpbmRvd0RpZEFwcGVhclwiLFxyXG5cdFx0V2luZG93V2lsbERpc2FwcGVhcjogXCJpb3M6V2luZG93V2lsbERpc2FwcGVhclwiLFxyXG5cdFx0V2luZG93RGlkRGlzYXBwZWFyOiBcImlvczpXaW5kb3dEaWREaXNhcHBlYXJcIixcclxuXHRcdFdpbmRvd1NhZmVBcmVhSW5zZXRzQ2hhbmdlZDogXCJpb3M6V2luZG93U2FmZUFyZWFJbnNldHNDaGFuZ2VkXCIsXHJcblx0XHRXaW5kb3dPcmllbnRhdGlvbkNoYW5nZWQ6IFwiaW9zOldpbmRvd09yaWVudGF0aW9uQ2hhbmdlZFwiLFxyXG5cdFx0V2luZG93VG91Y2hCZWdhbjogXCJpb3M6V2luZG93VG91Y2hCZWdhblwiLFxyXG5cdFx0V2luZG93VG91Y2hNb3ZlZDogXCJpb3M6V2luZG93VG91Y2hNb3ZlZFwiLFxyXG5cdFx0V2luZG93VG91Y2hFbmRlZDogXCJpb3M6V2luZG93VG91Y2hFbmRlZFwiLFxyXG5cdFx0V2luZG93VG91Y2hDYW5jZWxsZWQ6IFwiaW9zOldpbmRvd1RvdWNoQ2FuY2VsbGVkXCIsXHJcblx0XHRXZWJWaWV3RGlkU3RhcnROYXZpZ2F0aW9uOiBcImlvczpXZWJWaWV3RGlkU3RhcnROYXZpZ2F0aW9uXCIsXHJcblx0XHRXZWJWaWV3RGlkRmluaXNoTmF2aWdhdGlvbjogXCJpb3M6V2ViVmlld0RpZEZpbmlzaE5hdmlnYXRpb25cIixcclxuXHRcdFdlYlZpZXdEaWRGYWlsTmF2aWdhdGlvbjogXCJpb3M6V2ViVmlld0RpZEZhaWxOYXZpZ2F0aW9uXCIsXHJcblx0XHRXZWJWaWV3RGVjaWRlUG9saWN5Rm9yTmF2aWdhdGlvbkFjdGlvbjogXCJpb3M6V2ViVmlld0RlY2lkZVBvbGljeUZvck5hdmlnYXRpb25BY3Rpb25cIixcclxuXHR9KSxcclxuXHRDb21tb246IE9iamVjdC5mcmVlemUoe1xyXG5cdFx0QXBwbGljYXRpb25PcGVuZWRXaXRoRmlsZTogXCJjb21tb246QXBwbGljYXRpb25PcGVuZWRXaXRoRmlsZVwiLFxyXG5cdFx0QXBwbGljYXRpb25TdGFydGVkOiBcImNvbW1vbjpBcHBsaWNhdGlvblN0YXJ0ZWRcIixcclxuXHRcdEFwcGxpY2F0aW9uTGF1bmNoZWRXaXRoVXJsOiBcImNvbW1vbjpBcHBsaWNhdGlvbkxhdW5jaGVkV2l0aFVybFwiLFxyXG5cdFx0VGhlbWVDaGFuZ2VkOiBcImNvbW1vbjpUaGVtZUNoYW5nZWRcIixcclxuXHRcdFdpbmRvd0Nsb3Npbmc6IFwiY29tbW9uOldpbmRvd0Nsb3NpbmdcIixcclxuXHRcdFdpbmRvd0RpZE1vdmU6IFwiY29tbW9uOldpbmRvd0RpZE1vdmVcIixcclxuXHRcdFdpbmRvd0RpZFJlc2l6ZTogXCJjb21tb246V2luZG93RGlkUmVzaXplXCIsXHJcblx0XHRXaW5kb3dEUElDaGFuZ2VkOiBcImNvbW1vbjpXaW5kb3dEUElDaGFuZ2VkXCIsXHJcblx0XHRXaW5kb3dGaWxlc0Ryb3BwZWQ6IFwiY29tbW9uOldpbmRvd0ZpbGVzRHJvcHBlZFwiLFxyXG5cdFx0V2luZG93Rm9jdXM6IFwiY29tbW9uOldpbmRvd0ZvY3VzXCIsXHJcblx0XHRXaW5kb3dGdWxsc2NyZWVuOiBcImNvbW1vbjpXaW5kb3dGdWxsc2NyZWVuXCIsXHJcblx0XHRXaW5kb3dIaWRlOiBcImNvbW1vbjpXaW5kb3dIaWRlXCIsXHJcblx0XHRXaW5kb3dMb3N0Rm9jdXM6IFwiY29tbW9uOldpbmRvd0xvc3RGb2N1c1wiLFxyXG5cdFx0V2luZG93TWF4aW1pc2U6IFwiY29tbW9uOldpbmRvd01heGltaXNlXCIsXHJcblx0XHRXaW5kb3dNaW5pbWlzZTogXCJjb21tb246V2luZG93TWluaW1pc2VcIixcclxuXHRcdFdpbmRvd1RvZ2dsZUZyYW1lbGVzczogXCJjb21tb246V2luZG93VG9nZ2xlRnJhbWVsZXNzXCIsXHJcblx0XHRXaW5kb3dSZXN0b3JlOiBcImNvbW1vbjpXaW5kb3dSZXN0b3JlXCIsXHJcblx0XHRXaW5kb3dSdW50aW1lUmVhZHk6IFwiY29tbW9uOldpbmRvd1J1bnRpbWVSZWFkeVwiLFxyXG5cdFx0V2luZG93U2hvdzogXCJjb21tb246V2luZG93U2hvd1wiLFxyXG5cdFx0V2luZG93VW5GdWxsc2NyZWVuOiBcImNvbW1vbjpXaW5kb3dVbkZ1bGxzY3JlZW5cIixcclxuXHRcdFdpbmRvd1VuTWF4aW1pc2U6IFwiY29tbW9uOldpbmRvd1VuTWF4aW1pc2VcIixcclxuXHRcdFdpbmRvd1VuTWluaW1pc2U6IFwiY29tbW9uOldpbmRvd1VuTWluaW1pc2VcIixcclxuXHRcdFdpbmRvd1pvb206IFwiY29tbW9uOldpbmRvd1pvb21cIixcclxuXHRcdFdpbmRvd1pvb21JbjogXCJjb21tb246V2luZG93Wm9vbUluXCIsXHJcblx0XHRXaW5kb3dab29tT3V0OiBcImNvbW1vbjpXaW5kb3dab29tT3V0XCIsXHJcblx0XHRXaW5kb3dab29tUmVzZXQ6IFwiY29tbW9uOldpbmRvd1pvb21SZXNldFwiLFxyXG5cdH0pLFxyXG59KTtcclxuIiwgIi8qXHJcbiBfICAgICBfXyAgICAgXyBfX1xyXG58IHwgIC8gL19fXyBfKF8pIC9fX19fXHJcbnwgfCAvfCAvIC8gX18gYC8gLyAvIF9fXy9cclxufCB8LyB8LyAvIC9fLyAvIC8gKF9fICApXHJcbnxfXy98X18vXFxfXyxfL18vXy9fX19fL1xyXG5UaGUgZWxlY3Ryb24gYWx0ZXJuYXRpdmUgZm9yIEdvXHJcbihjKSBMZWEgQW50aG9ueSAyMDE5LXByZXNlbnRcclxuKi9cclxuXHJcbi8qKlxyXG4gKiBMb2dzIGEgbWVzc2FnZSB0byB0aGUgY29uc29sZSB3aXRoIGN1c3RvbSBmb3JtYXR0aW5nLlxyXG4gKlxyXG4gKiBAcGFyYW0gbWVzc2FnZSAtIFRoZSBtZXNzYWdlIHRvIGJlIGxvZ2dlZC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBkZWJ1Z0xvZyhtZXNzYWdlOiBhbnkpIHtcclxuICAgIC8vIGVzbGludC1kaXNhYmxlLW5leHQtbGluZVxyXG4gICAgY29uc29sZS5sb2coXHJcbiAgICAgICAgJyVjIHdhaWxzMyAlYyAnICsgbWVzc2FnZSArICcgJyxcclxuICAgICAgICAnYmFja2dyb3VuZDogI2FhMDAwMDsgY29sb3I6ICNmZmY7IGJvcmRlci1yYWRpdXM6IDNweCAwcHggMHB4IDNweDsgcGFkZGluZzogMXB4OyBmb250LXNpemU6IDAuN3JlbScsXHJcbiAgICAgICAgJ2JhY2tncm91bmQ6ICMwMDk5MDA7IGNvbG9yOiAjZmZmOyBib3JkZXItcmFkaXVzOiAwcHggM3B4IDNweCAwcHg7IHBhZGRpbmc6IDFweDsgZm9udC1zaXplOiAwLjdyZW0nXHJcbiAgICApO1xyXG59XHJcblxyXG4vKipcclxuICogQ2hlY2tzIHdoZXRoZXIgdGhlIHdlYnZpZXcgc3VwcG9ydHMgdGhlIHtAbGluayBNb3VzZUV2ZW50I2J1dHRvbnN9IHByb3BlcnR5LlxyXG4gKiBMb29raW5nIGF0IHlvdSBtYWNPUyBIaWdoIFNpZXJyYSFcclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBjYW5UcmFja0J1dHRvbnMoKTogYm9vbGVhbiB7XHJcbiAgICByZXR1cm4gKG5ldyBNb3VzZUV2ZW50KCdtb3VzZWRvd24nKSkuYnV0dG9ucyA9PT0gMDtcclxufVxyXG5cclxuLyoqXHJcbiAqIENoZWNrcyB3aGV0aGVyIHRoZSBicm93c2VyIHN1cHBvcnRzIHJlbW92aW5nIGxpc3RlbmVycyBieSB0cmlnZ2VyaW5nIGFuIEFib3J0U2lnbmFsXHJcbiAqIChzZWUgaHR0cHM6Ly9kZXZlbG9wZXIubW96aWxsYS5vcmcvZW4tVVMvZG9jcy9XZWIvQVBJL0V2ZW50VGFyZ2V0L2FkZEV2ZW50TGlzdGVuZXIjc2lnbmFsKS5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBjYW5BYm9ydExpc3RlbmVycygpIHtcclxuICAgIGlmICghRXZlbnRUYXJnZXQgfHwgIUFib3J0U2lnbmFsIHx8ICFBYm9ydENvbnRyb2xsZXIpXHJcbiAgICAgICAgcmV0dXJuIGZhbHNlO1xyXG5cclxuICAgIGxldCByZXN1bHQgPSB0cnVlO1xyXG5cclxuICAgIGNvbnN0IHRhcmdldCA9IG5ldyBFdmVudFRhcmdldCgpO1xyXG4gICAgY29uc3QgY29udHJvbGxlciA9IG5ldyBBYm9ydENvbnRyb2xsZXIoKTtcclxuICAgIHRhcmdldC5hZGRFdmVudExpc3RlbmVyKCd0ZXN0JywgKCkgPT4geyByZXN1bHQgPSBmYWxzZTsgfSwgeyBzaWduYWw6IGNvbnRyb2xsZXIuc2lnbmFsIH0pO1xyXG4gICAgY29udHJvbGxlci5hYm9ydCgpO1xyXG4gICAgdGFyZ2V0LmRpc3BhdGNoRXZlbnQobmV3IEN1c3RvbUV2ZW50KCd0ZXN0JykpO1xyXG5cclxuICAgIHJldHVybiByZXN1bHQ7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBSZXNvbHZlcyB0aGUgY2xvc2VzdCBIVE1MRWxlbWVudCBhbmNlc3RvciBvZiBhbiBldmVudCdzIHRhcmdldC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBldmVudFRhcmdldChldmVudDogRXZlbnQpOiBIVE1MRWxlbWVudCB7XHJcbiAgICBpZiAoZXZlbnQudGFyZ2V0IGluc3RhbmNlb2YgSFRNTEVsZW1lbnQpIHtcclxuICAgICAgICByZXR1cm4gZXZlbnQudGFyZ2V0O1xyXG4gICAgfSBlbHNlIGlmICghKGV2ZW50LnRhcmdldCBpbnN0YW5jZW9mIEhUTUxFbGVtZW50KSAmJiBldmVudC50YXJnZXQgaW5zdGFuY2VvZiBOb2RlKSB7XHJcbiAgICAgICAgcmV0dXJuIGV2ZW50LnRhcmdldC5wYXJlbnRFbGVtZW50ID8/IGRvY3VtZW50LmJvZHk7XHJcbiAgICB9IGVsc2Uge1xyXG4gICAgICAgIHJldHVybiBkb2N1bWVudC5ib2R5O1xyXG4gICAgfVxyXG59XHJcblxyXG4vKioqXHJcbiBUaGlzIHRlY2huaXF1ZSBmb3IgcHJvcGVyIGxvYWQgZGV0ZWN0aW9uIGlzIHRha2VuIGZyb20gSFRNWDpcclxuXHJcbiBCU0QgMi1DbGF1c2UgTGljZW5zZVxyXG5cclxuIENvcHlyaWdodCAoYykgMjAyMCwgQmlnIFNreSBTb2Z0d2FyZVxyXG4gQWxsIHJpZ2h0cyByZXNlcnZlZC5cclxuXHJcbiBSZWRpc3RyaWJ1dGlvbiBhbmQgdXNlIGluIHNvdXJjZSBhbmQgYmluYXJ5IGZvcm1zLCB3aXRoIG9yIHdpdGhvdXRcclxuIG1vZGlmaWNhdGlvbiwgYXJlIHBlcm1pdHRlZCBwcm92aWRlZCB0aGF0IHRoZSBmb2xsb3dpbmcgY29uZGl0aW9ucyBhcmUgbWV0OlxyXG5cclxuIDEuIFJlZGlzdHJpYnV0aW9ucyBvZiBzb3VyY2UgY29kZSBtdXN0IHJldGFpbiB0aGUgYWJvdmUgY29weXJpZ2h0IG5vdGljZSwgdGhpc1xyXG4gbGlzdCBvZiBjb25kaXRpb25zIGFuZCB0aGUgZm9sbG93aW5nIGRpc2NsYWltZXIuXHJcblxyXG4gMi4gUmVkaXN0cmlidXRpb25zIGluIGJpbmFyeSBmb3JtIG11c3QgcmVwcm9kdWNlIHRoZSBhYm92ZSBjb3B5cmlnaHQgbm90aWNlLFxyXG4gdGhpcyBsaXN0IG9mIGNvbmRpdGlvbnMgYW5kIHRoZSBmb2xsb3dpbmcgZGlzY2xhaW1lciBpbiB0aGUgZG9jdW1lbnRhdGlvblxyXG4gYW5kL29yIG90aGVyIG1hdGVyaWFscyBwcm92aWRlZCB3aXRoIHRoZSBkaXN0cmlidXRpb24uXHJcblxyXG4gVEhJUyBTT0ZUV0FSRSBJUyBQUk9WSURFRCBCWSBUSEUgQ09QWVJJR0hUIEhPTERFUlMgQU5EIENPTlRSSUJVVE9SUyBcIkFTIElTXCJcclxuIEFORCBBTlkgRVhQUkVTUyBPUiBJTVBMSUVEIFdBUlJBTlRJRVMsIElOQ0xVRElORywgQlVUIE5PVCBMSU1JVEVEIFRPLCBUSEVcclxuIElNUExJRUQgV0FSUkFOVElFUyBPRiBNRVJDSEFOVEFCSUxJVFkgQU5EIEZJVE5FU1MgRk9SIEEgUEFSVElDVUxBUiBQVVJQT1NFIEFSRVxyXG4gRElTQ0xBSU1FRC4gSU4gTk8gRVZFTlQgU0hBTEwgVEhFIENPUFlSSUdIVCBIT0xERVIgT1IgQ09OVFJJQlVUT1JTIEJFIExJQUJMRVxyXG4gRk9SIEFOWSBESVJFQ1QsIElORElSRUNULCBJTkNJREVOVEFMLCBTUEVDSUFMLCBFWEVNUExBUlksIE9SIENPTlNFUVVFTlRJQUxcclxuIERBTUFHRVMgKElOQ0xVRElORywgQlVUIE5PVCBMSU1JVEVEIFRPLCBQUk9DVVJFTUVOVCBPRiBTVUJTVElUVVRFIEdPT0RTIE9SXHJcbiBTRVJWSUNFUzsgTE9TUyBPRiBVU0UsIERBVEEsIE9SIFBST0ZJVFM7IE9SIEJVU0lORVNTIElOVEVSUlVQVElPTikgSE9XRVZFUlxyXG4gQ0FVU0VEIEFORCBPTiBBTlkgVEhFT1JZIE9GIExJQUJJTElUWSwgV0hFVEhFUiBJTiBDT05UUkFDVCwgU1RSSUNUIExJQUJJTElUWSxcclxuIE9SIFRPUlQgKElOQ0xVRElORyBORUdMSUdFTkNFIE9SIE9USEVSV0lTRSkgQVJJU0lORyBJTiBBTlkgV0FZIE9VVCBPRiBUSEUgVVNFXHJcbiBPRiBUSElTIFNPRlRXQVJFLCBFVkVOIElGIEFEVklTRUQgT0YgVEhFIFBPU1NJQklMSVRZIE9GIFNVQ0ggREFNQUdFLlxyXG5cclxuICoqKi9cclxuXHJcbmxldCBpc1JlYWR5ID0gZmFsc2U7XHJcbmRvY3VtZW50LmFkZEV2ZW50TGlzdGVuZXIoJ0RPTUNvbnRlbnRMb2FkZWQnLCAoKSA9PiB7IGlzUmVhZHkgPSB0cnVlIH0pO1xyXG5cclxuZXhwb3J0IGZ1bmN0aW9uIHdoZW5SZWFkeShjYWxsYmFjazogKCkgPT4gdm9pZCkge1xyXG4gICAgaWYgKGlzUmVhZHkgfHwgZG9jdW1lbnQucmVhZHlTdGF0ZSA9PT0gJ2NvbXBsZXRlJykge1xyXG4gICAgICAgIGNhbGxiYWNrKCk7XHJcbiAgICB9IGVsc2Uge1xyXG4gICAgICAgIGRvY3VtZW50LmFkZEV2ZW50TGlzdGVuZXIoJ0RPTUNvbnRlbnRMb2FkZWQnLCBjYWxsYmFjayk7XHJcbiAgICB9XHJcbn1cclxuIiwgIi8qXHJcbiBfXHQgICBfX1x0ICBfIF9fXHJcbnwgfFx0IC8gL19fXyBfKF8pIC9fX19fXHJcbnwgfCAvfCAvIC8gX18gYC8gLyAvIF9fXy9cclxufCB8LyB8LyAvIC9fLyAvIC8gKF9fICApXHJcbnxfXy98X18vXFxfXyxfL18vXy9fX19fL1xyXG5UaGUgZWxlY3Ryb24gYWx0ZXJuYXRpdmUgZm9yIEdvXHJcbihjKSBMZWEgQW50aG9ueSAyMDE5LXByZXNlbnRcclxuKi9cclxuXHJcbmltcG9ydCB7bmV3UnVudGltZUNhbGxlciwgb2JqZWN0TmFtZXN9IGZyb20gXCIuL3J1bnRpbWUuanNcIjtcclxuaW1wb3J0IHR5cGUgeyBTY3JlZW4gfSBmcm9tIFwiLi9zY3JlZW5zLmpzXCI7XHJcblxyXG4vLyBEcm9wIHRhcmdldCBjb25zdGFudHNcclxuY29uc3QgRFJPUF9UQVJHRVRfQVRUUklCVVRFID0gJ2RhdGEtZmlsZS1kcm9wLXRhcmdldCc7XHJcbmNvbnN0IERST1BfVEFSR0VUX0FDVElWRV9DTEFTUyA9ICdmaWxlLWRyb3AtdGFyZ2V0LWFjdGl2ZSc7XHJcbmxldCBjdXJyZW50RHJvcFRhcmdldDogRWxlbWVudCB8IG51bGwgPSBudWxsO1xyXG5cclxuY29uc3QgUG9zaXRpb25NZXRob2QgICAgICAgICAgICAgICAgICAgID0gMDtcclxuY29uc3QgQ2VudGVyTWV0aG9kICAgICAgICAgICAgICAgICAgICAgID0gMTtcclxuY29uc3QgQ2xvc2VNZXRob2QgICAgICAgICAgICAgICAgICAgICAgID0gMjtcclxuY29uc3QgRGlzYWJsZVNpemVDb25zdHJhaW50c01ldGhvZCAgICAgID0gMztcclxuY29uc3QgRW5hYmxlU2l6ZUNvbnN0cmFpbnRzTWV0aG9kICAgICAgID0gNDtcclxuY29uc3QgRm9jdXNNZXRob2QgICAgICAgICAgICAgICAgICAgICAgID0gNTtcclxuY29uc3QgRm9yY2VSZWxvYWRNZXRob2QgICAgICAgICAgICAgICAgID0gNjtcclxuY29uc3QgRnVsbHNjcmVlbk1ldGhvZCAgICAgICAgICAgICAgICAgID0gNztcclxuY29uc3QgR2V0U2NyZWVuTWV0aG9kICAgICAgICAgICAgICAgICAgID0gODtcclxuY29uc3QgR2V0Wm9vbU1ldGhvZCAgICAgICAgICAgICAgICAgICAgID0gOTtcclxuY29uc3QgSGVpZ2h0TWV0aG9kICAgICAgICAgICAgICAgICAgICAgID0gMTA7XHJcbmNvbnN0IEhpZGVNZXRob2QgICAgICAgICAgICAgICAgICAgICAgICA9IDExO1xyXG5jb25zdCBJc0ZvY3VzZWRNZXRob2QgICAgICAgICAgICAgICAgICAgPSAxMjtcclxuY29uc3QgSXNGdWxsc2NyZWVuTWV0aG9kICAgICAgICAgICAgICAgID0gMTM7XHJcbmNvbnN0IElzTWF4aW1pc2VkTWV0aG9kICAgICAgICAgICAgICAgICA9IDE0O1xyXG5jb25zdCBJc01pbmltaXNlZE1ldGhvZCAgICAgICAgICAgICAgICAgPSAxNTtcclxuY29uc3QgTWF4aW1pc2VNZXRob2QgICAgICAgICAgICAgICAgICAgID0gMTY7XHJcbmNvbnN0IE1pbmltaXNlTWV0aG9kICAgICAgICAgICAgICAgICAgICA9IDE3O1xyXG5jb25zdCBOYW1lTWV0aG9kICAgICAgICAgICAgICAgICAgICAgICAgPSAxODtcclxuY29uc3QgT3BlbkRldlRvb2xzTWV0aG9kICAgICAgICAgICAgICAgID0gMTk7XHJcbmNvbnN0IFJlbGF0aXZlUG9zaXRpb25NZXRob2QgICAgICAgICAgICA9IDIwO1xyXG5jb25zdCBSZWxvYWRNZXRob2QgICAgICAgICAgICAgICAgICAgICAgPSAyMTtcclxuY29uc3QgUmVzaXphYmxlTWV0aG9kICAgICAgICAgICAgICAgICAgID0gMjI7XHJcbmNvbnN0IFJlc3RvcmVNZXRob2QgICAgICAgICAgICAgICAgICAgICA9IDIzO1xyXG5jb25zdCBTZXRQb3NpdGlvbk1ldGhvZCAgICAgICAgICAgICAgICAgPSAyNDtcclxuY29uc3QgU2V0QWx3YXlzT25Ub3BNZXRob2QgICAgICAgICAgICAgID0gMjU7XHJcbmNvbnN0IFNldEJhY2tncm91bmRDb2xvdXJNZXRob2QgICAgICAgICA9IDI2O1xyXG5jb25zdCBTZXRGcmFtZWxlc3NNZXRob2QgICAgICAgICAgICAgICAgPSAyNztcclxuY29uc3QgU2V0RnVsbHNjcmVlbkJ1dHRvbkVuYWJsZWRNZXRob2QgID0gMjg7XHJcbmNvbnN0IFNldE1heFNpemVNZXRob2QgICAgICAgICAgICAgICAgICA9IDI5O1xyXG5jb25zdCBTZXRNaW5TaXplTWV0aG9kICAgICAgICAgICAgICAgICAgPSAzMDtcclxuY29uc3QgU2V0UmVsYXRpdmVQb3NpdGlvbk1ldGhvZCAgICAgICAgID0gMzE7XHJcbmNvbnN0IFNldFJlc2l6YWJsZU1ldGhvZCAgICAgICAgICAgICAgICA9IDMyO1xyXG5jb25zdCBTZXRTaXplTWV0aG9kICAgICAgICAgICAgICAgICAgICAgPSAzMztcclxuY29uc3QgU2V0VGl0bGVNZXRob2QgICAgICAgICAgICAgICAgICAgID0gMzQ7XHJcbmNvbnN0IFNldFpvb21NZXRob2QgICAgICAgICAgICAgICAgICAgICA9IDM1O1xyXG5jb25zdCBTaG93TWV0aG9kICAgICAgICAgICAgICAgICAgICAgICAgPSAzNjtcclxuY29uc3QgU2l6ZU1ldGhvZCAgICAgICAgICAgICAgICAgICAgICAgID0gMzc7XHJcbmNvbnN0IFRvZ2dsZUZ1bGxzY3JlZW5NZXRob2QgICAgICAgICAgICA9IDM4O1xyXG5jb25zdCBUb2dnbGVNYXhpbWlzZU1ldGhvZCAgICAgICAgICAgICAgPSAzOTtcclxuY29uc3QgVG9nZ2xlRnJhbWVsZXNzTWV0aG9kICAgICAgICAgICAgID0gNDA7IFxyXG5jb25zdCBVbkZ1bGxzY3JlZW5NZXRob2QgICAgICAgICAgICAgICAgPSA0MTtcclxuY29uc3QgVW5NYXhpbWlzZU1ldGhvZCAgICAgICAgICAgICAgICAgID0gNDI7XHJcbmNvbnN0IFVuTWluaW1pc2VNZXRob2QgICAgICAgICAgICAgICAgICA9IDQzO1xyXG5jb25zdCBXaWR0aE1ldGhvZCAgICAgICAgICAgICAgICAgICAgICAgPSA0NDtcclxuY29uc3QgWm9vbU1ldGhvZCAgICAgICAgICAgICAgICAgICAgICAgID0gNDU7XHJcbmNvbnN0IFpvb21Jbk1ldGhvZCAgICAgICAgICAgICAgICAgICAgICA9IDQ2O1xyXG5jb25zdCBab29tT3V0TWV0aG9kICAgICAgICAgICAgICAgICAgICAgPSA0NztcclxuY29uc3QgWm9vbVJlc2V0TWV0aG9kICAgICAgICAgICAgICAgICAgID0gNDg7XHJcbmNvbnN0IFNuYXBBc3Npc3RNZXRob2QgICAgICAgICAgICAgICAgICA9IDQ5O1xyXG5jb25zdCBGaWxlc0Ryb3BwZWQgICAgICAgICAgICAgICAgICAgICAgPSA1MDtcclxuY29uc3QgUHJpbnRNZXRob2QgICAgICAgICAgICAgICAgICAgICAgID0gNTE7XHJcbmNvbnN0IFNldFNjcmVlbk1ldGhvZCAgICAgICAgICAgICAgICAgICA9IDUyO1xyXG5cclxuLyoqXHJcbiAqIEZpbmRzIHRoZSBuZWFyZXN0IGRyb3AgdGFyZ2V0IGVsZW1lbnQgYnkgd2Fsa2luZyB1cCB0aGUgRE9NIHRyZWUuXHJcbiAqL1xyXG5mdW5jdGlvbiBnZXREcm9wVGFyZ2V0RWxlbWVudChlbGVtZW50OiBFbGVtZW50IHwgbnVsbCk6IEVsZW1lbnQgfCBudWxsIHtcclxuICAgIGlmICghZWxlbWVudCkge1xyXG4gICAgICAgIHJldHVybiBudWxsO1xyXG4gICAgfVxyXG4gICAgcmV0dXJuIGVsZW1lbnQuY2xvc2VzdChgWyR7RFJPUF9UQVJHRVRfQVRUUklCVVRFfV1gKTtcclxufVxyXG5cclxuLyoqXHJcbiAqIENoZWNrIGlmIHdlIGNhbiB1c2UgV2ViVmlldzIncyBwb3N0TWVzc2FnZVdpdGhBZGRpdGlvbmFsT2JqZWN0cyAoV2luZG93cylcclxuICogQWxzbyBjaGVja3MgdGhhdCBFbmFibGVGaWxlRHJvcCBpcyB0cnVlIGZvciB0aGlzIHdpbmRvdy5cclxuICovXHJcbmZ1bmN0aW9uIGNhblJlc29sdmVGaWxlUGF0aHMoKTogYm9vbGVhbiB7XHJcbiAgICAvLyBNdXN0IGhhdmUgV2ViVmlldzIncyBwb3N0TWVzc2FnZVdpdGhBZGRpdGlvbmFsT2JqZWN0cyBBUEkgKFdpbmRvd3Mgb25seSlcclxuICAgIGlmICgod2luZG93IGFzIGFueSkuY2hyb21lPy53ZWJ2aWV3Py5wb3N0TWVzc2FnZVdpdGhBZGRpdGlvbmFsT2JqZWN0cyA9PSBudWxsKSB7XHJcbiAgICAgICAgcmV0dXJuIGZhbHNlO1xyXG4gICAgfVxyXG4gICAgLy8gTXVzdCBoYXZlIEVuYWJsZUZpbGVEcm9wIHNldCB0byB0cnVlIGZvciB0aGlzIHdpbmRvd1xyXG4gICAgLy8gVGhpcyBmbGFnIGlzIHNldCBieSB0aGUgR28gYmFja2VuZCBkdXJpbmcgcnVudGltZSBpbml0aWFsaXphdGlvblxyXG4gICAgcmV0dXJuICh3aW5kb3cgYXMgYW55KS5fd2FpbHM/LmZsYWdzPy5lbmFibGVGaWxlRHJvcCA9PT0gdHJ1ZTtcclxufVxyXG5cclxuLyoqXHJcbiAqIFNlbmQgZmlsZSBkcm9wIHRvIGJhY2tlbmQgdmlhIFdlYlZpZXcyIChXaW5kb3dzIG9ubHkpXHJcbiAqL1xyXG5mdW5jdGlvbiByZXNvbHZlRmlsZVBhdGhzKHg6IG51bWJlciwgeTogbnVtYmVyLCBmaWxlczogRmlsZVtdKTogdm9pZCB7XHJcbiAgICBpZiAoKHdpbmRvdyBhcyBhbnkpLmNocm9tZT8ud2Vidmlldz8ucG9zdE1lc3NhZ2VXaXRoQWRkaXRpb25hbE9iamVjdHMpIHtcclxuICAgICAgICAod2luZG93IGFzIGFueSkuY2hyb21lLndlYnZpZXcucG9zdE1lc3NhZ2VXaXRoQWRkaXRpb25hbE9iamVjdHMoYGZpbGU6ZHJvcDoke3h9OiR7eX1gLCBmaWxlcyk7XHJcbiAgICB9XHJcbn1cclxuXHJcbi8vIE5hdGl2ZSBkcmFnIHN0YXRlIChMaW51eC9tYWNPUyBpbnRlcmNlcHQgRE9NIGRyYWcgZXZlbnRzKVxyXG5sZXQgbmF0aXZlRHJhZ0FjdGl2ZSA9IGZhbHNlO1xyXG5cclxuLyoqXHJcbiAqIENsZWFucyB1cCBuYXRpdmUgZHJhZyBzdGF0ZSBhbmQgaG92ZXIgZWZmZWN0cy5cclxuICogQ2FsbGVkIG9uIGRyb3Agb3Igd2hlbiBkcmFnIGxlYXZlcyB0aGUgd2luZG93LlxyXG4gKi9cclxuZnVuY3Rpb24gY2xlYW51cE5hdGl2ZURyYWcoKTogdm9pZCB7XHJcbiAgICBuYXRpdmVEcmFnQWN0aXZlID0gZmFsc2U7XHJcbiAgICBpZiAoY3VycmVudERyb3BUYXJnZXQpIHtcclxuICAgICAgICBjdXJyZW50RHJvcFRhcmdldC5jbGFzc0xpc3QucmVtb3ZlKERST1BfVEFSR0VUX0FDVElWRV9DTEFTUyk7XHJcbiAgICAgICAgY3VycmVudERyb3BUYXJnZXQgPSBudWxsO1xyXG4gICAgfVxyXG59XHJcblxyXG4vKipcclxuICogQ2FsbGVkIGZyb20gR28gd2hlbiBhIGZpbGUgZHJhZyBlbnRlcnMgdGhlIHdpbmRvdyBvbiBMaW51eC9tYWNPUy5cclxuICovXHJcbmZ1bmN0aW9uIGhhbmRsZURyYWdFbnRlcigpOiB2b2lkIHtcclxuICAgIC8vIENoZWNrIGlmIGZpbGUgZHJvcHMgYXJlIGVuYWJsZWQgZm9yIHRoaXMgd2luZG93XHJcbiAgICBpZiAoKHdpbmRvdyBhcyBhbnkpLl93YWlscz8uZmxhZ3M/LmVuYWJsZUZpbGVEcm9wID09PSBmYWxzZSkge1xyXG4gICAgICAgIHJldHVybjsgLy8gRmlsZSBkcm9wcyBkaXNhYmxlZCwgZG9uJ3QgYWN0aXZhdGUgZHJhZyBzdGF0ZVxyXG4gICAgfVxyXG4gICAgbmF0aXZlRHJhZ0FjdGl2ZSA9IHRydWU7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDYWxsZWQgZnJvbSBHbyB3aGVuIGEgZmlsZSBkcmFnIGxlYXZlcyB0aGUgd2luZG93IG9uIExpbnV4L21hY09TLlxyXG4gKi9cclxuZnVuY3Rpb24gaGFuZGxlRHJhZ0xlYXZlKCk6IHZvaWQge1xyXG4gICAgY2xlYW51cE5hdGl2ZURyYWcoKTtcclxufVxyXG5cclxuLyoqXHJcbiAqIENhbGxlZCBmcm9tIEdvIGR1cmluZyBmaWxlIGRyYWcgdG8gdXBkYXRlIGhvdmVyIHN0YXRlIG9uIExpbnV4L21hY09TLlxyXG4gKiBAcGFyYW0geCAtIFggY29vcmRpbmF0ZSBpbiBDU1MgcGl4ZWxzXHJcbiAqIEBwYXJhbSB5IC0gWSBjb29yZGluYXRlIGluIENTUyBwaXhlbHNcclxuICovXHJcbmZ1bmN0aW9uIGhhbmRsZURyYWdPdmVyKHg6IG51bWJlciwgeTogbnVtYmVyKTogdm9pZCB7XHJcbiAgICBpZiAoIW5hdGl2ZURyYWdBY3RpdmUpIHJldHVybjtcclxuICAgIFxyXG4gICAgLy8gQ2hlY2sgaWYgZmlsZSBkcm9wcyBhcmUgZW5hYmxlZCBmb3IgdGhpcyB3aW5kb3dcclxuICAgIGlmICgod2luZG93IGFzIGFueSkuX3dhaWxzPy5mbGFncz8uZW5hYmxlRmlsZURyb3AgPT09IGZhbHNlKSB7XHJcbiAgICAgICAgcmV0dXJuOyAvLyBGaWxlIGRyb3BzIGRpc2FibGVkLCBkb24ndCBzaG93IGhvdmVyIGVmZmVjdHNcclxuICAgIH1cclxuICAgIFxyXG4gICAgY29uc3QgdGFyZ2V0RWxlbWVudCA9IGRvY3VtZW50LmVsZW1lbnRGcm9tUG9pbnQoeCwgeSk7XHJcbiAgICBjb25zdCBkcm9wVGFyZ2V0ID0gZ2V0RHJvcFRhcmdldEVsZW1lbnQodGFyZ2V0RWxlbWVudCk7XHJcbiAgICBcclxuICAgIGlmIChjdXJyZW50RHJvcFRhcmdldCAmJiBjdXJyZW50RHJvcFRhcmdldCAhPT0gZHJvcFRhcmdldCkge1xyXG4gICAgICAgIGN1cnJlbnREcm9wVGFyZ2V0LmNsYXNzTGlzdC5yZW1vdmUoRFJPUF9UQVJHRVRfQUNUSVZFX0NMQVNTKTtcclxuICAgIH1cclxuICAgIFxyXG4gICAgaWYgKGRyb3BUYXJnZXQpIHtcclxuICAgICAgICBkcm9wVGFyZ2V0LmNsYXNzTGlzdC5hZGQoRFJPUF9UQVJHRVRfQUNUSVZFX0NMQVNTKTtcclxuICAgICAgICBjdXJyZW50RHJvcFRhcmdldCA9IGRyb3BUYXJnZXQ7XHJcbiAgICB9IGVsc2Uge1xyXG4gICAgICAgIGN1cnJlbnREcm9wVGFyZ2V0ID0gbnVsbDtcclxuICAgIH1cclxufVxyXG5cclxuXHJcblxyXG4vLyBFeHBvcnQgdGhlIGhhbmRsZXJzIGZvciB1c2UgYnkgR28gdmlhIGluZGV4LnRzXHJcbmV4cG9ydCB7IGhhbmRsZURyYWdFbnRlciwgaGFuZGxlRHJhZ0xlYXZlLCBoYW5kbGVEcmFnT3ZlciB9O1xyXG5cclxuLyoqXHJcbiAqIEEgcmVjb3JkIGRlc2NyaWJpbmcgdGhlIHBvc2l0aW9uIG9mIGEgd2luZG93LlxyXG4gKi9cclxuaW50ZXJmYWNlIFBvc2l0aW9uIHtcclxuICAgIC8qKiBUaGUgaG9yaXpvbnRhbCBwb3NpdGlvbiBvZiB0aGUgd2luZG93LiAqL1xyXG4gICAgeDogbnVtYmVyO1xyXG4gICAgLyoqIFRoZSB2ZXJ0aWNhbCBwb3NpdGlvbiBvZiB0aGUgd2luZG93LiAqL1xyXG4gICAgeTogbnVtYmVyO1xyXG59XHJcblxyXG4vKipcclxuICogQSByZWNvcmQgZGVzY3JpYmluZyB0aGUgc2l6ZSBvZiBhIHdpbmRvdy5cclxuICovXHJcbmludGVyZmFjZSBTaXplIHtcclxuICAgIC8qKiBUaGUgd2lkdGggb2YgdGhlIHdpbmRvdy4gKi9cclxuICAgIHdpZHRoOiBudW1iZXI7XHJcbiAgICAvKiogVGhlIGhlaWdodCBvZiB0aGUgd2luZG93LiAqL1xyXG4gICAgaGVpZ2h0OiBudW1iZXI7XHJcbn1cclxuXHJcbi8vIFByaXZhdGUgZmllbGQgbmFtZXMuXHJcbmNvbnN0IGNhbGxlclN5bSA9IFN5bWJvbChcImNhbGxlclwiKTtcclxuXHJcbmNsYXNzIFdpbmRvdyB7XHJcbiAgICAvLyBQcml2YXRlIGZpZWxkcy5cclxuICAgIHByaXZhdGUgW2NhbGxlclN5bV06IChtZXNzYWdlOiBudW1iZXIsIGFyZ3M/OiBhbnkpID0+IFByb21pc2U8YW55PjtcclxuXHJcbiAgICAvKipcclxuICAgICAqIEluaXRpYWxpc2VzIGEgd2luZG93IG9iamVjdCB3aXRoIHRoZSBzcGVjaWZpZWQgbmFtZS5cclxuICAgICAqXHJcbiAgICAgKiBAcHJpdmF0ZVxyXG4gICAgICogQHBhcmFtIG5hbWUgLSBUaGUgbmFtZSBvZiB0aGUgdGFyZ2V0IHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgY29uc3RydWN0b3IobmFtZTogc3RyaW5nID0gJycpIHtcclxuICAgICAgICB0aGlzW2NhbGxlclN5bV0gPSBuZXdSdW50aW1lQ2FsbGVyKG9iamVjdE5hbWVzLldpbmRvdywgbmFtZSlcclxuXHJcbiAgICAgICAgLy8gYmluZCBpbnN0YW5jZSBtZXRob2QgdG8gbWFrZSB0aGVtIGVhc2lseSB1c2FibGUgaW4gZXZlbnQgaGFuZGxlcnNcclxuICAgICAgICBmb3IgKGNvbnN0IG1ldGhvZCBvZiBPYmplY3QuZ2V0T3duUHJvcGVydHlOYW1lcyhXaW5kb3cucHJvdG90eXBlKSkge1xyXG4gICAgICAgICAgICBpZiAoXHJcbiAgICAgICAgICAgICAgICBtZXRob2QgIT09IFwiY29uc3RydWN0b3JcIlxyXG4gICAgICAgICAgICAgICAgJiYgdHlwZW9mICh0aGlzIGFzIGFueSlbbWV0aG9kXSA9PT0gXCJmdW5jdGlvblwiXHJcbiAgICAgICAgICAgICkge1xyXG4gICAgICAgICAgICAgICAgKHRoaXMgYXMgYW55KVttZXRob2RdID0gKHRoaXMgYXMgYW55KVttZXRob2RdLmJpbmQodGhpcyk7XHJcbiAgICAgICAgICAgIH1cclxuICAgICAgICB9XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBHZXRzIHRoZSBzcGVjaWZpZWQgd2luZG93LlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSBuYW1lIC0gVGhlIG5hbWUgb2YgdGhlIHdpbmRvdyB0byBnZXQuXHJcbiAgICAgKiBAcmV0dXJucyBUaGUgY29ycmVzcG9uZGluZyB3aW5kb3cgb2JqZWN0LlxyXG4gICAgICovXHJcbiAgICBHZXQobmFtZTogc3RyaW5nKTogV2luZG93IHtcclxuICAgICAgICByZXR1cm4gbmV3IFdpbmRvdyhuYW1lKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJldHVybnMgdGhlIGFic29sdXRlIHBvc2l0aW9uIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgVGhlIGN1cnJlbnQgYWJzb2x1dGUgcG9zaXRpb24gb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgUG9zaXRpb24oKTogUHJvbWlzZTxQb3NpdGlvbj4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oUG9zaXRpb25NZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ2VudGVycyB0aGUgd2luZG93IG9uIHRoZSBzY3JlZW4uXHJcbiAgICAgKi9cclxuICAgIENlbnRlcigpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKENlbnRlck1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBDbG9zZXMgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgQ2xvc2UoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShDbG9zZU1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBEaXNhYmxlcyBtaW4vbWF4IHNpemUgY29uc3RyYWludHMuXHJcbiAgICAgKi9cclxuICAgIERpc2FibGVTaXplQ29uc3RyYWludHMoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShEaXNhYmxlU2l6ZUNvbnN0cmFpbnRzTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIEVuYWJsZXMgbWluL21heCBzaXplIGNvbnN0cmFpbnRzLlxyXG4gICAgICovXHJcbiAgICBFbmFibGVTaXplQ29uc3RyYWludHMoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShFbmFibGVTaXplQ29uc3RyYWludHNNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogRm9jdXNlcyB0aGUgd2luZG93LlxyXG4gICAgICovXHJcbiAgICBGb2N1cygpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKEZvY3VzTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIEZvcmNlcyB0aGUgd2luZG93IHRvIHJlbG9hZCB0aGUgcGFnZSBhc3NldHMuXHJcbiAgICAgKi9cclxuICAgIEZvcmNlUmVsb2FkKCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oRm9yY2VSZWxvYWRNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogU3dpdGNoZXMgdGhlIHdpbmRvdyB0byBmdWxsc2NyZWVuIG1vZGUuXHJcbiAgICAgKi9cclxuICAgIEZ1bGxzY3JlZW4oKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShGdWxsc2NyZWVuTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJldHVybnMgdGhlIHNjcmVlbiB0aGF0IHRoZSB3aW5kb3cgaXMgb24uXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgVGhlIHNjcmVlbiB0aGUgd2luZG93IGlzIGN1cnJlbnRseSBvbi5cclxuICAgICAqL1xyXG4gICAgR2V0U2NyZWVuKCk6IFByb21pc2U8U2NyZWVuPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShHZXRTY3JlZW5NZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogUmV0dXJucyB0aGUgY3VycmVudCB6b29tIGxldmVsIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgVGhlIGN1cnJlbnQgem9vbSBsZXZlbC5cclxuICAgICAqL1xyXG4gICAgR2V0Wm9vbSgpOiBQcm9taXNlPG51bWJlcj4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oR2V0Wm9vbU1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBSZXR1cm5zIHRoZSBoZWlnaHQgb2YgdGhlIHdpbmRvdy5cclxuICAgICAqXHJcbiAgICAgKiBAcmV0dXJucyBUaGUgY3VycmVudCBoZWlnaHQgb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgSGVpZ2h0KCk6IFByb21pc2U8bnVtYmVyPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShIZWlnaHRNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogSGlkZXMgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgSGlkZSgpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKEhpZGVNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogUmV0dXJucyB0cnVlIGlmIHRoZSB3aW5kb3cgaXMgZm9jdXNlZC5cclxuICAgICAqXHJcbiAgICAgKiBAcmV0dXJucyBXaGV0aGVyIHRoZSB3aW5kb3cgaXMgY3VycmVudGx5IGZvY3VzZWQuXHJcbiAgICAgKi9cclxuICAgIElzRm9jdXNlZCgpOiBQcm9taXNlPGJvb2xlYW4+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKElzRm9jdXNlZE1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBSZXR1cm5zIHRydWUgaWYgdGhlIHdpbmRvdyBpcyBmdWxsc2NyZWVuLlxyXG4gICAgICpcclxuICAgICAqIEByZXR1cm5zIFdoZXRoZXIgdGhlIHdpbmRvdyBpcyBjdXJyZW50bHkgZnVsbHNjcmVlbi5cclxuICAgICAqL1xyXG4gICAgSXNGdWxsc2NyZWVuKCk6IFByb21pc2U8Ym9vbGVhbj4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oSXNGdWxsc2NyZWVuTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJldHVybnMgdHJ1ZSBpZiB0aGUgd2luZG93IGlzIG1heGltaXNlZC5cclxuICAgICAqXHJcbiAgICAgKiBAcmV0dXJucyBXaGV0aGVyIHRoZSB3aW5kb3cgaXMgY3VycmVudGx5IG1heGltaXNlZC5cclxuICAgICAqL1xyXG4gICAgSXNNYXhpbWlzZWQoKTogUHJvbWlzZTxib29sZWFuPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShJc01heGltaXNlZE1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBSZXR1cm5zIHRydWUgaWYgdGhlIHdpbmRvdyBpcyBtaW5pbWlzZWQuXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgV2hldGhlciB0aGUgd2luZG93IGlzIGN1cnJlbnRseSBtaW5pbWlzZWQuXHJcbiAgICAgKi9cclxuICAgIElzTWluaW1pc2VkKCk6IFByb21pc2U8Ym9vbGVhbj4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oSXNNaW5pbWlzZWRNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogTWF4aW1pc2VzIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIE1heGltaXNlKCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oTWF4aW1pc2VNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogTWluaW1pc2VzIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIE1pbmltaXNlKCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oTWluaW1pc2VNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogUmV0dXJucyB0aGUgbmFtZSBvZiB0aGUgd2luZG93LlxyXG4gICAgICpcclxuICAgICAqIEByZXR1cm5zIFRoZSBuYW1lIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIE5hbWUoKTogUHJvbWlzZTxzdHJpbmc+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKE5hbWVNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogT3BlbnMgdGhlIGRldmVsb3BtZW50IHRvb2xzIHBhbmUuXHJcbiAgICAgKi9cclxuICAgIE9wZW5EZXZUb29scygpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKE9wZW5EZXZUb29sc01ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBSZXR1cm5zIHRoZSByZWxhdGl2ZSBwb3NpdGlvbiBvZiB0aGUgd2luZG93IHRvIHRoZSBzY3JlZW4uXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgVGhlIGN1cnJlbnQgcmVsYXRpdmUgcG9zaXRpb24gb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgUmVsYXRpdmVQb3NpdGlvbigpOiBQcm9taXNlPFBvc2l0aW9uPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShSZWxhdGl2ZVBvc2l0aW9uTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJlbG9hZHMgdGhlIHBhZ2UgYXNzZXRzLlxyXG4gICAgICovXHJcbiAgICBSZWxvYWQoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShSZWxvYWRNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogUmV0dXJucyB0cnVlIGlmIHRoZSB3aW5kb3cgaXMgcmVzaXphYmxlLlxyXG4gICAgICpcclxuICAgICAqIEByZXR1cm5zIFdoZXRoZXIgdGhlIHdpbmRvdyBpcyBjdXJyZW50bHkgcmVzaXphYmxlLlxyXG4gICAgICovXHJcbiAgICBSZXNpemFibGUoKTogUHJvbWlzZTxib29sZWFuPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShSZXNpemFibGVNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogUmVzdG9yZXMgdGhlIHdpbmRvdyB0byBpdHMgcHJldmlvdXMgc3RhdGUgaWYgaXQgd2FzIHByZXZpb3VzbHkgbWluaW1pc2VkLCBtYXhpbWlzZWQgb3IgZnVsbHNjcmVlbi5cclxuICAgICAqL1xyXG4gICAgUmVzdG9yZSgpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFJlc3RvcmVNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogU2V0cyB0aGUgYWJzb2x1dGUgcG9zaXRpb24gb2YgdGhlIHdpbmRvdy5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0geCAtIFRoZSBkZXNpcmVkIGhvcml6b250YWwgYWJzb2x1dGUgcG9zaXRpb24gb2YgdGhlIHdpbmRvdy5cclxuICAgICAqIEBwYXJhbSB5IC0gVGhlIGRlc2lyZWQgdmVydGljYWwgYWJzb2x1dGUgcG9zaXRpb24gb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgU2V0UG9zaXRpb24oeDogbnVtYmVyLCB5OiBudW1iZXIpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFNldFBvc2l0aW9uTWV0aG9kLCB7IHgsIHkgfSk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBTZXRzIHRoZSB3aW5kb3cgdG8gYmUgYWx3YXlzIG9uIHRvcC5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0gYWx3YXlzT25Ub3AgLSBXaGV0aGVyIHRoZSB3aW5kb3cgc2hvdWxkIHN0YXkgb24gdG9wLlxyXG4gICAgICovXHJcbiAgICBTZXRBbHdheXNPblRvcChhbHdheXNPblRvcDogYm9vbGVhbik6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oU2V0QWx3YXlzT25Ub3BNZXRob2QsIHsgYWx3YXlzT25Ub3AgfSk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBTZXRzIHRoZSBiYWNrZ3JvdW5kIGNvbG91ciBvZiB0aGUgd2luZG93LlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSByIC0gVGhlIGRlc2lyZWQgcmVkIGNvbXBvbmVudCBvZiB0aGUgd2luZG93IGJhY2tncm91bmQuXHJcbiAgICAgKiBAcGFyYW0gZyAtIFRoZSBkZXNpcmVkIGdyZWVuIGNvbXBvbmVudCBvZiB0aGUgd2luZG93IGJhY2tncm91bmQuXHJcbiAgICAgKiBAcGFyYW0gYiAtIFRoZSBkZXNpcmVkIGJsdWUgY29tcG9uZW50IG9mIHRoZSB3aW5kb3cgYmFja2dyb3VuZC5cclxuICAgICAqIEBwYXJhbSBhIC0gVGhlIGRlc2lyZWQgYWxwaGEgY29tcG9uZW50IG9mIHRoZSB3aW5kb3cgYmFja2dyb3VuZC5cclxuICAgICAqL1xyXG4gICAgU2V0QmFja2dyb3VuZENvbG91cihyOiBudW1iZXIsIGc6IG51bWJlciwgYjogbnVtYmVyLCBhOiBudW1iZXIpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFNldEJhY2tncm91bmRDb2xvdXJNZXRob2QsIHsgciwgZywgYiwgYSB9KTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJlbW92ZXMgdGhlIHdpbmRvdyBmcmFtZSBhbmQgdGl0bGUgYmFyLlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSBmcmFtZWxlc3MgLSBXaGV0aGVyIHRoZSB3aW5kb3cgc2hvdWxkIGJlIGZyYW1lbGVzcy5cclxuICAgICAqL1xyXG4gICAgU2V0RnJhbWVsZXNzKGZyYW1lbGVzczogYm9vbGVhbik6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oU2V0RnJhbWVsZXNzTWV0aG9kLCB7IGZyYW1lbGVzcyB9KTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIERpc2FibGVzIHRoZSBzeXN0ZW0gZnVsbHNjcmVlbiBidXR0b24uXHJcbiAgICAgKlxyXG4gICAgICogQHBhcmFtIGVuYWJsZWQgLSBXaGV0aGVyIHRoZSBmdWxsc2NyZWVuIGJ1dHRvbiBzaG91bGQgYmUgZW5hYmxlZC5cclxuICAgICAqL1xyXG4gICAgU2V0RnVsbHNjcmVlbkJ1dHRvbkVuYWJsZWQoZW5hYmxlZDogYm9vbGVhbik6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oU2V0RnVsbHNjcmVlbkJ1dHRvbkVuYWJsZWRNZXRob2QsIHsgZW5hYmxlZCB9KTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFNldHMgdGhlIG1heGltdW0gc2l6ZSBvZiB0aGUgd2luZG93LlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSB3aWR0aCAtIFRoZSBkZXNpcmVkIG1heGltdW0gd2lkdGggb2YgdGhlIHdpbmRvdy5cclxuICAgICAqIEBwYXJhbSBoZWlnaHQgLSBUaGUgZGVzaXJlZCBtYXhpbXVtIGhlaWdodCBvZiB0aGUgd2luZG93LlxyXG4gICAgICovXHJcbiAgICBTZXRNYXhTaXplKHdpZHRoOiBudW1iZXIsIGhlaWdodDogbnVtYmVyKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShTZXRNYXhTaXplTWV0aG9kLCB7IHdpZHRoLCBoZWlnaHQgfSk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBTZXRzIHRoZSBtaW5pbXVtIHNpemUgb2YgdGhlIHdpbmRvdy5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0gd2lkdGggLSBUaGUgZGVzaXJlZCBtaW5pbXVtIHdpZHRoIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKiBAcGFyYW0gaGVpZ2h0IC0gVGhlIGRlc2lyZWQgbWluaW11bSBoZWlnaHQgb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgU2V0TWluU2l6ZSh3aWR0aDogbnVtYmVyLCBoZWlnaHQ6IG51bWJlcik6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oU2V0TWluU2l6ZU1ldGhvZCwgeyB3aWR0aCwgaGVpZ2h0IH0pO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogU2V0cyB0aGUgcmVsYXRpdmUgcG9zaXRpb24gb2YgdGhlIHdpbmRvdyB0byB0aGUgc2NyZWVuLlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSB4IC0gVGhlIGRlc2lyZWQgaG9yaXpvbnRhbCByZWxhdGl2ZSBwb3NpdGlvbiBvZiB0aGUgd2luZG93LlxyXG4gICAgICogQHBhcmFtIHkgLSBUaGUgZGVzaXJlZCB2ZXJ0aWNhbCByZWxhdGl2ZSBwb3NpdGlvbiBvZiB0aGUgd2luZG93LlxyXG4gICAgICovXHJcbiAgICBTZXRSZWxhdGl2ZVBvc2l0aW9uKHg6IG51bWJlciwgeTogbnVtYmVyKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShTZXRSZWxhdGl2ZVBvc2l0aW9uTWV0aG9kLCB7IHgsIHkgfSk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBTZXRzIHdoZXRoZXIgdGhlIHdpbmRvdyBpcyByZXNpemFibGUuXHJcbiAgICAgKlxyXG4gICAgICogQHBhcmFtIHJlc2l6YWJsZSAtIFdoZXRoZXIgdGhlIHdpbmRvdyBzaG91bGQgYmUgcmVzaXphYmxlLlxyXG4gICAgICovXHJcbiAgICBTZXRSZXNpemFibGUocmVzaXphYmxlOiBib29sZWFuKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShTZXRSZXNpemFibGVNZXRob2QsIHsgcmVzaXphYmxlIH0pO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogU2V0cyB0aGUgc2l6ZSBvZiB0aGUgd2luZG93LlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSB3aWR0aCAtIFRoZSBkZXNpcmVkIHdpZHRoIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKiBAcGFyYW0gaGVpZ2h0IC0gVGhlIGRlc2lyZWQgaGVpZ2h0IG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIFNldFNpemUod2lkdGg6IG51bWJlciwgaGVpZ2h0OiBudW1iZXIpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFNldFNpemVNZXRob2QsIHsgd2lkdGgsIGhlaWdodCB9KTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFNldHMgdGhlIHRpdGxlIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKlxyXG4gICAgICogQHBhcmFtIHRpdGxlIC0gVGhlIGRlc2lyZWQgdGl0bGUgb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgU2V0VGl0bGUodGl0bGU6IHN0cmluZyk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oU2V0VGl0bGVNZXRob2QsIHsgdGl0bGUgfSk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBTZXRzIHRoZSB6b29tIGxldmVsIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKlxyXG4gICAgICogQHBhcmFtIHpvb20gLSBUaGUgZGVzaXJlZCB6b29tIGxldmVsLlxyXG4gICAgICovXHJcbiAgICBTZXRab29tKHpvb206IG51bWJlcik6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oU2V0Wm9vbU1ldGhvZCwgeyB6b29tIH0pO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogU2hvd3MgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgU2hvdygpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFNob3dNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogUmV0dXJucyB0aGUgc2l6ZSBvZiB0aGUgd2luZG93LlxyXG4gICAgICpcclxuICAgICAqIEByZXR1cm5zIFRoZSBjdXJyZW50IHNpemUgb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgU2l6ZSgpOiBQcm9taXNlPFNpemU+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFNpemVNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogVG9nZ2xlcyB0aGUgd2luZG93IGJldHdlZW4gZnVsbHNjcmVlbiBhbmQgbm9ybWFsLlxyXG4gICAgICovXHJcbiAgICBUb2dnbGVGdWxsc2NyZWVuKCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oVG9nZ2xlRnVsbHNjcmVlbk1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBUb2dnbGVzIHRoZSB3aW5kb3cgYmV0d2VlbiBtYXhpbWlzZWQgYW5kIG5vcm1hbC5cclxuICAgICAqL1xyXG4gICAgVG9nZ2xlTWF4aW1pc2UoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShUb2dnbGVNYXhpbWlzZU1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBUb2dnbGVzIHRoZSB3aW5kb3cgYmV0d2VlbiBmcmFtZWxlc3MgYW5kIG5vcm1hbC5cclxuICAgICAqL1xyXG4gICAgVG9nZ2xlRnJhbWVsZXNzKCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oVG9nZ2xlRnJhbWVsZXNzTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFVuLWZ1bGxzY3JlZW5zIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIFVuRnVsbHNjcmVlbigpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFVuRnVsbHNjcmVlbk1ldGhvZCk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBVbi1tYXhpbWlzZXMgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgVW5NYXhpbWlzZSgpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFVuTWF4aW1pc2VNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogVW4tbWluaW1pc2VzIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIFVuTWluaW1pc2UoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShVbk1pbmltaXNlTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJldHVybnMgdGhlIHdpZHRoIG9mIHRoZSB3aW5kb3cuXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgVGhlIGN1cnJlbnQgd2lkdGggb2YgdGhlIHdpbmRvdy5cclxuICAgICAqL1xyXG4gICAgV2lkdGgoKTogUHJvbWlzZTxudW1iZXI+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFdpZHRoTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFpvb21zIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIFpvb20oKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShab29tTWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIEluY3JlYXNlcyB0aGUgem9vbSBsZXZlbCBvZiB0aGUgd2VidmlldyBjb250ZW50LlxyXG4gICAgICovXHJcbiAgICBab29tSW4oKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShab29tSW5NZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogRGVjcmVhc2VzIHRoZSB6b29tIGxldmVsIG9mIHRoZSB3ZWJ2aWV3IGNvbnRlbnQuXHJcbiAgICAgKi9cclxuICAgIFpvb21PdXQoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShab29tT3V0TWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIFJlc2V0cyB0aGUgem9vbSBsZXZlbCBvZiB0aGUgd2VidmlldyBjb250ZW50LlxyXG4gICAgICovXHJcbiAgICBab29tUmVzZXQoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShab29tUmVzZXRNZXRob2QpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogSGFuZGxlcyBmaWxlIGRyb3BzIG9yaWdpbmF0aW5nIGZyb20gcGxhdGZvcm0tc3BlY2lmaWMgY29kZSAoZS5nLiwgbWFjT1MvTGludXggbmF0aXZlIGRyYWctYW5kLWRyb3ApLlxyXG4gICAgICogR2F0aGVycyBpbmZvcm1hdGlvbiBhYm91dCB0aGUgZHJvcCB0YXJnZXQgZWxlbWVudCBhbmQgc2VuZHMgaXQgYmFjayB0byB0aGUgR28gYmFja2VuZC5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0gZmlsZW5hbWVzIC0gQW4gYXJyYXkgb2YgZmlsZSBwYXRocyAoc3RyaW5ncykgdGhhdCB3ZXJlIGRyb3BwZWQuXHJcbiAgICAgKiBAcGFyYW0geCAtIFRoZSB4LWNvb3JkaW5hdGUgb2YgdGhlIGRyb3AgZXZlbnQgKENTUyBwaXhlbHMpLlxyXG4gICAgICogQHBhcmFtIHkgLSBUaGUgeS1jb29yZGluYXRlIG9mIHRoZSBkcm9wIGV2ZW50IChDU1MgcGl4ZWxzKS5cclxuICAgICAqL1xyXG4gICAgSGFuZGxlUGxhdGZvcm1GaWxlRHJvcChmaWxlbmFtZXM6IHN0cmluZ1tdLCB4OiBudW1iZXIsIHk6IG51bWJlcik6IHZvaWQge1xyXG4gICAgICAgIC8vIENoZWNrIGlmIGZpbGUgZHJvcHMgYXJlIGVuYWJsZWQgZm9yIHRoaXMgd2luZG93XHJcbiAgICAgICAgaWYgKCh3aW5kb3cgYXMgYW55KS5fd2FpbHM/LmZsYWdzPy5lbmFibGVGaWxlRHJvcCA9PT0gZmFsc2UpIHtcclxuICAgICAgICAgICAgcmV0dXJuOyAvLyBGaWxlIGRyb3BzIGRpc2FibGVkLCBpZ25vcmUgdGhlIGRyb3BcclxuICAgICAgICB9XHJcbiAgICAgICAgXHJcbiAgICAgICAgY29uc3QgZWxlbWVudCA9IGRvY3VtZW50LmVsZW1lbnRGcm9tUG9pbnQoeCwgeSk7XHJcbiAgICAgICAgY29uc3QgZHJvcFRhcmdldCA9IGdldERyb3BUYXJnZXRFbGVtZW50KGVsZW1lbnQpO1xyXG5cclxuICAgICAgICBpZiAoIWRyb3BUYXJnZXQpIHtcclxuICAgICAgICAgICAgLy8gRHJvcCB3YXMgbm90IG9uIGEgZGVzaWduYXRlZCBkcm9wIHRhcmdldCAtIGlnbm9yZVxyXG4gICAgICAgICAgICByZXR1cm47XHJcbiAgICAgICAgfVxyXG5cclxuICAgICAgICBjb25zdCBlbGVtZW50RGV0YWlscyA9IHtcclxuICAgICAgICAgICAgaWQ6IGRyb3BUYXJnZXQuaWQsXHJcbiAgICAgICAgICAgIGNsYXNzTGlzdDogQXJyYXkuZnJvbShkcm9wVGFyZ2V0LmNsYXNzTGlzdCksXHJcbiAgICAgICAgICAgIGF0dHJpYnV0ZXM6IHt9IGFzIHsgW2tleTogc3RyaW5nXTogc3RyaW5nIH0sXHJcbiAgICAgICAgfTtcclxuICAgICAgICBmb3IgKGxldCBpID0gMDsgaSA8IGRyb3BUYXJnZXQuYXR0cmlidXRlcy5sZW5ndGg7IGkrKykge1xyXG4gICAgICAgICAgICBjb25zdCBhdHRyID0gZHJvcFRhcmdldC5hdHRyaWJ1dGVzW2ldO1xyXG4gICAgICAgICAgICBlbGVtZW50RGV0YWlscy5hdHRyaWJ1dGVzW2F0dHIubmFtZV0gPSBhdHRyLnZhbHVlO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgY29uc3QgcGF5bG9hZCA9IHtcclxuICAgICAgICAgICAgZmlsZW5hbWVzLFxyXG4gICAgICAgICAgICB4LFxyXG4gICAgICAgICAgICB5LFxyXG4gICAgICAgICAgICBlbGVtZW50RGV0YWlscyxcclxuICAgICAgICB9O1xyXG5cclxuICAgICAgICB0aGlzW2NhbGxlclN5bV0oRmlsZXNEcm9wcGVkLCBwYXlsb2FkKTtcclxuICAgICAgICBcclxuICAgICAgICAvLyBDbGVhbiB1cCBuYXRpdmUgZHJhZyBzdGF0ZSBhZnRlciBkcm9wXHJcbiAgICAgICAgY2xlYW51cE5hdGl2ZURyYWcoKTtcclxuICAgIH1cclxuICBcclxuICAgIC8qKlxyXG4gICAgICogTW92ZXMgdGhlIHdpbmRvdyB0byB0aGUgY2VudGVyIG9mIHRoZSBzcGVjaWZpZWQgc2NyZWVuJ3Mgd29yayBhcmVhLlxyXG4gICAgICpcclxuICAgICAqIEBwYXJhbSBzY3JlZW5JRCAtIFRoZSBJRCBvZiB0aGUgdGFyZ2V0IHNjcmVlbi5cclxuICAgICAqL1xyXG4gICAgU2V0U2NyZWVuKHNjcmVlbklEOiBzdHJpbmcpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpc1tjYWxsZXJTeW1dKFNldFNjcmVlbk1ldGhvZCwgeyBzY3JlZW5JRCB9KTtcclxuICAgIH1cclxuXHJcbiAgICAvKiBUcmlnZ2VycyBXaW5kb3dzIDExIFNuYXAgQXNzaXN0IGZlYXR1cmUgKFdpbmRvd3Mgb25seSkuXHJcbiAgICAgKiBUaGlzIGlzIGVxdWl2YWxlbnQgdG8gcHJlc3NpbmcgV2luK1ogYW5kIHNob3dzIHNuYXAgbGF5b3V0IG9wdGlvbnMuXHJcbiAgICAgKi9cclxuICAgIFNuYXBBc3Npc3QoKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIHRoaXNbY2FsbGVyU3ltXShTbmFwQXNzaXN0TWV0aG9kKTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIE9wZW5zIHRoZSBwcmludCBkaWFsb2cgZm9yIHRoZSB3aW5kb3cuXHJcbiAgICAgKi9cclxuICAgIFByaW50KCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgICAgIHJldHVybiB0aGlzW2NhbGxlclN5bV0oUHJpbnRNZXRob2QpO1xyXG4gICAgfVxyXG59XHJcblxyXG4vKipcclxuICogVGhlIHdpbmRvdyB3aXRoaW4gd2hpY2ggdGhlIHNjcmlwdCBpcyBydW5uaW5nLlxyXG4gKi9cclxuY29uc3QgdGhpc1dpbmRvdyA9IG5ldyBXaW5kb3coJycpO1xyXG5cclxuLyoqXHJcbiAqIFNldHMgdXAgZ2xvYmFsIGRyYWcgYW5kIGRyb3AgZXZlbnQgbGlzdGVuZXJzIGZvciBmaWxlIGRyb3BzLlxyXG4gKiBIYW5kbGVzIHZpc3VhbCBmZWVkYmFjayAoaG92ZXIgc3RhdGUpIGFuZCBmaWxlIGRyb3AgcHJvY2Vzc2luZy5cclxuICovXHJcbmZ1bmN0aW9uIHNldHVwRHJvcFRhcmdldExpc3RlbmVycygpIHtcclxuICAgIGNvbnN0IGRvY0VsZW1lbnQgPSBkb2N1bWVudC5kb2N1bWVudEVsZW1lbnQ7XHJcbiAgICBsZXQgZHJhZ0VudGVyQ291bnRlciA9IDA7XHJcblxyXG4gICAgZG9jRWxlbWVudC5hZGRFdmVudExpc3RlbmVyKCdkcmFnZW50ZXInLCAoZXZlbnQpID0+IHtcclxuICAgICAgICBpZiAoIWV2ZW50LmRhdGFUcmFuc2Zlcj8udHlwZXMuaW5jbHVkZXMoJ0ZpbGVzJykpIHtcclxuICAgICAgICAgICAgcmV0dXJuOyAvLyBPbmx5IGhhbmRsZSBmaWxlIGRyYWdzLCBsZXQgb3RoZXIgZHJhZ3MgcGFzcyB0aHJvdWdoXHJcbiAgICAgICAgfVxyXG4gICAgICAgIGV2ZW50LnByZXZlbnREZWZhdWx0KCk7IC8vIEFsd2F5cyBwcmV2ZW50IGRlZmF1bHQgdG8gc3RvcCBicm93c2VyIG5hdmlnYXRpb25cclxuICAgICAgICAvLyBPbiBXaW5kb3dzLCBjaGVjayBpZiBmaWxlIGRyb3BzIGFyZSBlbmFibGVkIGZvciB0aGlzIHdpbmRvd1xyXG4gICAgICAgIGlmICgod2luZG93IGFzIGFueSkuX3dhaWxzPy5mbGFncz8uZW5hYmxlRmlsZURyb3AgPT09IGZhbHNlKSB7XHJcbiAgICAgICAgICAgIGV2ZW50LmRhdGFUcmFuc2Zlci5kcm9wRWZmZWN0ID0gJ25vbmUnOyAvLyBTaG93IFwibm8gZHJvcFwiIGN1cnNvclxyXG4gICAgICAgICAgICByZXR1cm47IC8vIEZpbGUgZHJvcHMgZGlzYWJsZWQsIGRvbid0IHNob3cgaG92ZXIgZWZmZWN0c1xyXG4gICAgICAgIH1cclxuICAgICAgICBkcmFnRW50ZXJDb3VudGVyKys7XHJcbiAgICAgICAgXHJcbiAgICAgICAgY29uc3QgdGFyZ2V0RWxlbWVudCA9IGRvY3VtZW50LmVsZW1lbnRGcm9tUG9pbnQoZXZlbnQuY2xpZW50WCwgZXZlbnQuY2xpZW50WSk7XHJcbiAgICAgICAgY29uc3QgZHJvcFRhcmdldCA9IGdldERyb3BUYXJnZXRFbGVtZW50KHRhcmdldEVsZW1lbnQpO1xyXG5cclxuICAgICAgICAvLyBVcGRhdGUgaG92ZXIgc3RhdGVcclxuICAgICAgICBpZiAoY3VycmVudERyb3BUYXJnZXQgJiYgY3VycmVudERyb3BUYXJnZXQgIT09IGRyb3BUYXJnZXQpIHtcclxuICAgICAgICAgICAgY3VycmVudERyb3BUYXJnZXQuY2xhc3NMaXN0LnJlbW92ZShEUk9QX1RBUkdFVF9BQ1RJVkVfQ0xBU1MpO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgaWYgKGRyb3BUYXJnZXQpIHtcclxuICAgICAgICAgICAgZHJvcFRhcmdldC5jbGFzc0xpc3QuYWRkKERST1BfVEFSR0VUX0FDVElWRV9DTEFTUyk7XHJcbiAgICAgICAgICAgIGV2ZW50LmRhdGFUcmFuc2Zlci5kcm9wRWZmZWN0ID0gJ2NvcHknO1xyXG4gICAgICAgICAgICBjdXJyZW50RHJvcFRhcmdldCA9IGRyb3BUYXJnZXQ7XHJcbiAgICAgICAgfSBlbHNlIHtcclxuICAgICAgICAgICAgZXZlbnQuZGF0YVRyYW5zZmVyLmRyb3BFZmZlY3QgPSAnbm9uZSc7XHJcbiAgICAgICAgICAgIGN1cnJlbnREcm9wVGFyZ2V0ID0gbnVsbDtcclxuICAgICAgICB9XHJcbiAgICB9LCBmYWxzZSk7XHJcblxyXG4gICAgZG9jRWxlbWVudC5hZGRFdmVudExpc3RlbmVyKCdkcmFnb3ZlcicsIChldmVudCkgPT4ge1xyXG4gICAgICAgIGlmICghZXZlbnQuZGF0YVRyYW5zZmVyPy50eXBlcy5pbmNsdWRlcygnRmlsZXMnKSkge1xyXG4gICAgICAgICAgICByZXR1cm47IC8vIE9ubHkgaGFuZGxlIGZpbGUgZHJhZ3NcclxuICAgICAgICB9XHJcbiAgICAgICAgZXZlbnQucHJldmVudERlZmF1bHQoKTsgLy8gQWx3YXlzIHByZXZlbnQgZGVmYXVsdCB0byBzdG9wIGJyb3dzZXIgbmF2aWdhdGlvblxyXG4gICAgICAgIC8vIE9uIFdpbmRvd3MsIGNoZWNrIGlmIGZpbGUgZHJvcHMgYXJlIGVuYWJsZWQgZm9yIHRoaXMgd2luZG93XHJcbiAgICAgICAgaWYgKCh3aW5kb3cgYXMgYW55KS5fd2FpbHM/LmZsYWdzPy5lbmFibGVGaWxlRHJvcCA9PT0gZmFsc2UpIHtcclxuICAgICAgICAgICAgZXZlbnQuZGF0YVRyYW5zZmVyLmRyb3BFZmZlY3QgPSAnbm9uZSc7IC8vIFNob3cgXCJubyBkcm9wXCIgY3Vyc29yXHJcbiAgICAgICAgICAgIHJldHVybjsgLy8gRmlsZSBkcm9wcyBkaXNhYmxlZCwgZG9uJ3Qgc2hvdyBob3ZlciBlZmZlY3RzXHJcbiAgICAgICAgfVxyXG4gICAgICAgIFxyXG4gICAgICAgIC8vIFVwZGF0ZSBkcm9wIHRhcmdldCBhcyBjdXJzb3IgbW92ZXNcclxuICAgICAgICBjb25zdCB0YXJnZXRFbGVtZW50ID0gZG9jdW1lbnQuZWxlbWVudEZyb21Qb2ludChldmVudC5jbGllbnRYLCBldmVudC5jbGllbnRZKTtcclxuICAgICAgICBjb25zdCBkcm9wVGFyZ2V0ID0gZ2V0RHJvcFRhcmdldEVsZW1lbnQodGFyZ2V0RWxlbWVudCk7XHJcbiAgICAgICAgXHJcbiAgICAgICAgaWYgKGN1cnJlbnREcm9wVGFyZ2V0ICYmIGN1cnJlbnREcm9wVGFyZ2V0ICE9PSBkcm9wVGFyZ2V0KSB7XHJcbiAgICAgICAgICAgIGN1cnJlbnREcm9wVGFyZ2V0LmNsYXNzTGlzdC5yZW1vdmUoRFJPUF9UQVJHRVRfQUNUSVZFX0NMQVNTKTtcclxuICAgICAgICB9XHJcbiAgICAgICAgXHJcbiAgICAgICAgaWYgKGRyb3BUYXJnZXQpIHtcclxuICAgICAgICAgICAgaWYgKCFkcm9wVGFyZ2V0LmNsYXNzTGlzdC5jb250YWlucyhEUk9QX1RBUkdFVF9BQ1RJVkVfQ0xBU1MpKSB7XHJcbiAgICAgICAgICAgICAgICBkcm9wVGFyZ2V0LmNsYXNzTGlzdC5hZGQoRFJPUF9UQVJHRVRfQUNUSVZFX0NMQVNTKTtcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgICAgICBldmVudC5kYXRhVHJhbnNmZXIuZHJvcEVmZmVjdCA9ICdjb3B5JztcclxuICAgICAgICAgICAgY3VycmVudERyb3BUYXJnZXQgPSBkcm9wVGFyZ2V0O1xyXG4gICAgICAgIH0gZWxzZSB7XHJcbiAgICAgICAgICAgIGV2ZW50LmRhdGFUcmFuc2Zlci5kcm9wRWZmZWN0ID0gJ25vbmUnO1xyXG4gICAgICAgICAgICBjdXJyZW50RHJvcFRhcmdldCA9IG51bGw7XHJcbiAgICAgICAgfVxyXG4gICAgfSwgZmFsc2UpO1xyXG5cclxuICAgIGRvY0VsZW1lbnQuYWRkRXZlbnRMaXN0ZW5lcignZHJhZ2xlYXZlJywgKGV2ZW50KSA9PiB7XHJcbiAgICAgICAgaWYgKCFldmVudC5kYXRhVHJhbnNmZXI/LnR5cGVzLmluY2x1ZGVzKCdGaWxlcycpKSB7XHJcbiAgICAgICAgICAgIHJldHVybjtcclxuICAgICAgICB9XHJcbiAgICAgICAgZXZlbnQucHJldmVudERlZmF1bHQoKTsgLy8gQWx3YXlzIHByZXZlbnQgZGVmYXVsdCB0byBzdG9wIGJyb3dzZXIgbmF2aWdhdGlvblxyXG4gICAgICAgIC8vIE9uIFdpbmRvd3MsIGNoZWNrIGlmIGZpbGUgZHJvcHMgYXJlIGVuYWJsZWQgZm9yIHRoaXMgd2luZG93XHJcbiAgICAgICAgaWYgKCh3aW5kb3cgYXMgYW55KS5fd2FpbHM/LmZsYWdzPy5lbmFibGVGaWxlRHJvcCA9PT0gZmFsc2UpIHtcclxuICAgICAgICAgICAgcmV0dXJuO1xyXG4gICAgICAgIH1cclxuICAgICAgICBcclxuICAgICAgICAvLyBPbiBMaW51eC9XZWJLaXRHVEsgYW5kIG1hY09TLCBkcmFnbGVhdmUgZmlyZXMgaW1tZWRpYXRlbHkgd2l0aCByZWxhdGVkVGFyZ2V0PW51bGwgd2hlbiBuYXRpdmVcclxuICAgICAgICAvLyBkcmFnIGhhbmRsaW5nIGlzIGludm9sdmVkLiBJZ25vcmUgdGhlc2Ugc3B1cmlvdXMgZXZlbnRzIC0gd2UnbGwgY2xlYW4gdXAgb24gZHJvcCBpbnN0ZWFkLlxyXG4gICAgICAgIGlmIChldmVudC5yZWxhdGVkVGFyZ2V0ID09PSBudWxsKSB7XHJcbiAgICAgICAgICAgIHJldHVybjtcclxuICAgICAgICB9XHJcbiAgICAgICAgXHJcbiAgICAgICAgZHJhZ0VudGVyQ291bnRlci0tO1xyXG4gICAgICAgIFxyXG4gICAgICAgIGlmIChkcmFnRW50ZXJDb3VudGVyID09PSAwIHx8IFxyXG4gICAgICAgICAgICAoY3VycmVudERyb3BUYXJnZXQgJiYgIWN1cnJlbnREcm9wVGFyZ2V0LmNvbnRhaW5zKGV2ZW50LnJlbGF0ZWRUYXJnZXQgYXMgTm9kZSkpKSB7XHJcbiAgICAgICAgICAgIGlmIChjdXJyZW50RHJvcFRhcmdldCkge1xyXG4gICAgICAgICAgICAgICAgY3VycmVudERyb3BUYXJnZXQuY2xhc3NMaXN0LnJlbW92ZShEUk9QX1RBUkdFVF9BQ1RJVkVfQ0xBU1MpO1xyXG4gICAgICAgICAgICAgICAgY3VycmVudERyb3BUYXJnZXQgPSBudWxsO1xyXG4gICAgICAgICAgICB9XHJcbiAgICAgICAgICAgIGRyYWdFbnRlckNvdW50ZXIgPSAwO1xyXG4gICAgICAgIH1cclxuICAgIH0sIGZhbHNlKTtcclxuXHJcbiAgICBkb2NFbGVtZW50LmFkZEV2ZW50TGlzdGVuZXIoJ2Ryb3AnLCAoZXZlbnQpID0+IHtcclxuICAgICAgICBpZiAoIWV2ZW50LmRhdGFUcmFuc2Zlcj8udHlwZXMuaW5jbHVkZXMoJ0ZpbGVzJykpIHtcclxuICAgICAgICAgICAgcmV0dXJuOyAvLyBPbmx5IGhhbmRsZSBmaWxlIGRyb3BzXHJcbiAgICAgICAgfVxyXG4gICAgICAgIGV2ZW50LnByZXZlbnREZWZhdWx0KCk7IC8vIEFsd2F5cyBwcmV2ZW50IGRlZmF1bHQgdG8gc3RvcCBicm93c2VyIG5hdmlnYXRpb25cclxuICAgICAgICAvLyBPbiBXaW5kb3dzLCBjaGVjayBpZiBmaWxlIGRyb3BzIGFyZSBlbmFibGVkIGZvciB0aGlzIHdpbmRvd1xyXG4gICAgICAgIGlmICgod2luZG93IGFzIGFueSkuX3dhaWxzPy5mbGFncz8uZW5hYmxlRmlsZURyb3AgPT09IGZhbHNlKSB7XHJcbiAgICAgICAgICAgIHJldHVybjtcclxuICAgICAgICB9XHJcbiAgICAgICAgZHJhZ0VudGVyQ291bnRlciA9IDA7XHJcbiAgICAgICAgXHJcbiAgICAgICAgaWYgKGN1cnJlbnREcm9wVGFyZ2V0KSB7XHJcbiAgICAgICAgICAgIGN1cnJlbnREcm9wVGFyZ2V0LmNsYXNzTGlzdC5yZW1vdmUoRFJPUF9UQVJHRVRfQUNUSVZFX0NMQVNTKTtcclxuICAgICAgICAgICAgY3VycmVudERyb3BUYXJnZXQgPSBudWxsO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgLy8gT24gV2luZG93cywgaGFuZGxlIGZpbGUgZHJvcHMgdmlhIEphdmFTY3JpcHRcclxuICAgICAgICAvLyBPbiBtYWNPUy9MaW51eCwgbmF0aXZlIGNvZGUgd2lsbCBjYWxsIEhhbmRsZVBsYXRmb3JtRmlsZURyb3BcclxuICAgICAgICBpZiAoY2FuUmVzb2x2ZUZpbGVQYXRocygpKSB7XHJcbiAgICAgICAgICAgIGNvbnN0IGZpbGVzOiBGaWxlW10gPSBbXTtcclxuICAgICAgICAgICAgaWYgKGV2ZW50LmRhdGFUcmFuc2Zlci5pdGVtcykge1xyXG4gICAgICAgICAgICAgICAgZm9yIChjb25zdCBpdGVtIG9mIGV2ZW50LmRhdGFUcmFuc2Zlci5pdGVtcykge1xyXG4gICAgICAgICAgICAgICAgICAgIGlmIChpdGVtLmtpbmQgPT09ICdmaWxlJykge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICBjb25zdCBmaWxlID0gaXRlbS5nZXRBc0ZpbGUoKTtcclxuICAgICAgICAgICAgICAgICAgICAgICAgaWYgKGZpbGUpIGZpbGVzLnB1c2goZmlsZSk7XHJcbiAgICAgICAgICAgICAgICAgICAgfVxyXG4gICAgICAgICAgICAgICAgfVxyXG4gICAgICAgICAgICB9IGVsc2UgaWYgKGV2ZW50LmRhdGFUcmFuc2Zlci5maWxlcykge1xyXG4gICAgICAgICAgICAgICAgZm9yIChjb25zdCBmaWxlIG9mIGV2ZW50LmRhdGFUcmFuc2Zlci5maWxlcykge1xyXG4gICAgICAgICAgICAgICAgICAgIGZpbGVzLnB1c2goZmlsZSk7XHJcbiAgICAgICAgICAgICAgICB9XHJcbiAgICAgICAgICAgIH1cclxuICAgICAgICAgICAgXHJcbiAgICAgICAgICAgIGlmIChmaWxlcy5sZW5ndGggPiAwKSB7XHJcbiAgICAgICAgICAgICAgICByZXNvbHZlRmlsZVBhdGhzKGV2ZW50LmNsaWVudFgsIGV2ZW50LmNsaWVudFksIGZpbGVzKTtcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgIH1cclxuICAgIH0sIGZhbHNlKTtcclxufVxyXG5cclxuLy8gSW5pdGlhbGl6ZSBsaXN0ZW5lcnMgd2hlbiB0aGUgc2NyaXB0IGxvYWRzXHJcbmlmICh0eXBlb2Ygd2luZG93ICE9PSBcInVuZGVmaW5lZFwiICYmIHR5cGVvZiBkb2N1bWVudCAhPT0gXCJ1bmRlZmluZWRcIikge1xyXG4gICAgc2V0dXBEcm9wVGFyZ2V0TGlzdGVuZXJzKCk7XHJcbn1cclxuXHJcbmV4cG9ydCBkZWZhdWx0IHRoaXNXaW5kb3c7XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQgKiBhcyBSdW50aW1lIGZyb20gXCIuLi9Ad2FpbHNpby9ydW50aW1lL3NyY1wiO1xyXG5cclxuLy8gTk9URTogdGhlIGZvbGxvd2luZyBtZXRob2RzIE1VU1QgYmUgaW1wb3J0ZWQgZXhwbGljaXRseSBiZWNhdXNlIG9mIGhvdyBlc2J1aWxkIGluamVjdGlvbiB3b3Jrc1xyXG5pbXBvcnQgeyBFbmFibGUgYXMgRW5hYmxlV01MIH0gZnJvbSBcIi4uL0B3YWlsc2lvL3J1bnRpbWUvc3JjL3dtbFwiO1xyXG5pbXBvcnQgeyBkZWJ1Z0xvZyB9IGZyb20gXCIuLi9Ad2FpbHNpby9ydW50aW1lL3NyYy91dGlsc1wiO1xyXG5cclxud2luZG93LndhaWxzID0gUnVudGltZTtcclxuRW5hYmxlV01MKCk7XHJcblxyXG5pZiAoREVCVUcpIHtcclxuICAgIGRlYnVnTG9nKFwiV2FpbHMgUnVudGltZSBMb2FkZWRcIilcclxufVxyXG4iLCAiLypcclxuIF9cdCAgIF9fXHQgIF8gX19cclxufCB8XHQgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuaW1wb3J0IHsgbmV3UnVudGltZUNhbGxlciwgb2JqZWN0TmFtZXMgfSBmcm9tIFwiLi9ydW50aW1lLmpzXCI7XHJcblxyXG5jb25zdCBjYWxsID0gbmV3UnVudGltZUNhbGxlcihvYmplY3ROYW1lcy5TeXN0ZW0pO1xyXG5cclxuY29uc3QgU3lzdGVtSXNEYXJrTW9kZSA9IDA7XHJcbmNvbnN0IFN5c3RlbUVudmlyb25tZW50ID0gMTtcclxuY29uc3QgU3lzdGVtQ2FwYWJpbGl0aWVzID0gMjtcclxuXHJcbmNvbnN0IF9pbnZva2UgPSAoZnVuY3Rpb24gKCkge1xyXG4gICAgdHJ5IHtcclxuICAgICAgICAvLyBXaW5kb3dzIFdlYlZpZXcyXHJcbiAgICAgICAgaWYgKCh3aW5kb3cgYXMgYW55KS5jaHJvbWU/LndlYnZpZXc/LnBvc3RNZXNzYWdlKSB7XHJcbiAgICAgICAgICAgIHJldHVybiAod2luZG93IGFzIGFueSkuY2hyb21lLndlYnZpZXcucG9zdE1lc3NhZ2UuYmluZCgod2luZG93IGFzIGFueSkuY2hyb21lLndlYnZpZXcpO1xyXG4gICAgICAgIH1cclxuICAgICAgICAvLyBtYWNPUy9pT1MgV0tXZWJWaWV3XHJcbiAgICAgICAgZWxzZSBpZiAoKHdpbmRvdyBhcyBhbnkpLndlYmtpdD8ubWVzc2FnZUhhbmRsZXJzPy5bJ2V4dGVybmFsJ10/LnBvc3RNZXNzYWdlKSB7XHJcbiAgICAgICAgICAgIHJldHVybiAod2luZG93IGFzIGFueSkud2Via2l0Lm1lc3NhZ2VIYW5kbGVyc1snZXh0ZXJuYWwnXS5wb3N0TWVzc2FnZS5iaW5kKCh3aW5kb3cgYXMgYW55KS53ZWJraXQubWVzc2FnZUhhbmRsZXJzWydleHRlcm5hbCddKTtcclxuICAgICAgICB9XHJcbiAgICAgICAgLy8gQW5kcm9pZCBXZWJWaWV3IC0gdXNlcyBhZGRKYXZhc2NyaXB0SW50ZXJmYWNlIHdoaWNoIGV4cG9zZXMgd2luZG93LndhaWxzLmludm9rZVxyXG4gICAgICAgIGVsc2UgaWYgKCh3aW5kb3cgYXMgYW55KS53YWlscz8uaW52b2tlKSB7XHJcbiAgICAgICAgICAgIHJldHVybiAobXNnOiBhbnkpID0+ICh3aW5kb3cgYXMgYW55KS53YWlscy5pbnZva2UodHlwZW9mIG1zZyA9PT0gJ3N0cmluZycgPyBtc2cgOiBKU09OLnN0cmluZ2lmeShtc2cpKTtcclxuICAgICAgICB9XHJcbiAgICB9IGNhdGNoKGUpIHt9XHJcblxyXG4gICAgY29uc29sZS53YXJuKCdcXG4lY1x1MjZBMFx1RkUwRiBCcm93c2VyIEVudmlyb25tZW50IERldGVjdGVkICVjXFxuXFxuJWNPbmx5IFVJIHByZXZpZXdzIGFyZSBhdmFpbGFibGUgaW4gdGhlIGJyb3dzZXIuIEZvciBmdWxsIGZ1bmN0aW9uYWxpdHksIHBsZWFzZSBydW4gdGhlIGFwcGxpY2F0aW9uIGluIGRlc2t0b3AgbW9kZS5cXG5Nb3JlIGluZm9ybWF0aW9uIGF0OiBodHRwczovL3YzLndhaWxzLmlvL2xlYXJuL2J1aWxkLyN1c2luZy1hLWJyb3dzZXItZm9yLWRldmVsb3BtZW50XFxuJyxcclxuICAgICAgICAnYmFja2dyb3VuZDogI2ZmZmZmZjsgY29sb3I6ICMwMDAwMDA7IGZvbnQtd2VpZ2h0OiBib2xkOyBwYWRkaW5nOiA0cHggOHB4OyBib3JkZXItcmFkaXVzOiA0cHg7IGJvcmRlcjogMnB4IHNvbGlkICMwMDAwMDA7JyxcclxuICAgICAgICAnYmFja2dyb3VuZDogdHJhbnNwYXJlbnQ7JyxcclxuICAgICAgICAnY29sb3I6ICNmZmZmZmY7IGZvbnQtc3R5bGU6IGl0YWxpYzsgZm9udC13ZWlnaHQ6IGJvbGQ7Jyk7XHJcbiAgICByZXR1cm4gbnVsbDtcclxufSkoKTtcclxuXHJcbmV4cG9ydCBmdW5jdGlvbiBpbnZva2UobXNnOiBhbnkpOiB2b2lkIHtcclxuICAgIF9pbnZva2U/Lihtc2cpO1xyXG59XHJcblxyXG4vKipcclxuICogUmV0cmlldmVzIHRoZSBzeXN0ZW0gZGFyayBtb2RlIHN0YXR1cy5cclxuICpcclxuICogQHJldHVybnMgQSBwcm9taXNlIHRoYXQgcmVzb2x2ZXMgdG8gYSBib29sZWFuIHZhbHVlIGluZGljYXRpbmcgaWYgdGhlIHN5c3RlbSBpcyBpbiBkYXJrIG1vZGUuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSXNEYXJrTW9kZSgpOiBQcm9taXNlPGJvb2xlYW4+IHtcclxuICAgIHJldHVybiBjYWxsKFN5c3RlbUlzRGFya01vZGUpO1xyXG59XHJcblxyXG4vKipcclxuICogRmV0Y2hlcyB0aGUgY2FwYWJpbGl0aWVzIG9mIHRoZSBhcHBsaWNhdGlvbiBmcm9tIHRoZSBzZXJ2ZXIuXHJcbiAqXHJcbiAqIEByZXR1cm5zIEEgcHJvbWlzZSB0aGF0IHJlc29sdmVzIHRvIGFuIG9iamVjdCBjb250YWluaW5nIHRoZSBjYXBhYmlsaXRpZXMuXHJcbiAqL1xyXG5leHBvcnQgYXN5bmMgZnVuY3Rpb24gQ2FwYWJpbGl0aWVzKCk6IFByb21pc2U8UmVjb3JkPHN0cmluZywgYW55Pj4ge1xyXG4gICAgcmV0dXJuIGNhbGwoU3lzdGVtQ2FwYWJpbGl0aWVzKTtcclxufVxyXG5cclxuZXhwb3J0IGludGVyZmFjZSBPU0luZm8ge1xyXG4gICAgLyoqIFRoZSBicmFuZGluZyBvZiB0aGUgT1MuICovXHJcbiAgICBCcmFuZGluZzogc3RyaW5nO1xyXG4gICAgLyoqIFRoZSBJRCBvZiB0aGUgT1MuICovXHJcbiAgICBJRDogc3RyaW5nO1xyXG4gICAgLyoqIFRoZSBuYW1lIG9mIHRoZSBPUy4gKi9cclxuICAgIE5hbWU6IHN0cmluZztcclxuICAgIC8qKiBUaGUgdmVyc2lvbiBvZiB0aGUgT1MuICovXHJcbiAgICBWZXJzaW9uOiBzdHJpbmc7XHJcbn1cclxuXHJcbmV4cG9ydCBpbnRlcmZhY2UgRW52aXJvbm1lbnRJbmZvIHtcclxuICAgIC8qKiBUaGUgYXJjaGl0ZWN0dXJlIG9mIHRoZSBzeXN0ZW0uICovXHJcbiAgICBBcmNoOiBzdHJpbmc7XHJcbiAgICAvKiogVHJ1ZSBpZiB0aGUgYXBwbGljYXRpb24gaXMgcnVubmluZyBpbiBkZWJ1ZyBtb2RlLCBvdGhlcndpc2UgZmFsc2UuICovXHJcbiAgICBEZWJ1ZzogYm9vbGVhbjtcclxuICAgIC8qKiBUaGUgb3BlcmF0aW5nIHN5c3RlbSBpbiB1c2UuICovXHJcbiAgICBPUzogc3RyaW5nO1xyXG4gICAgLyoqIERldGFpbHMgb2YgdGhlIG9wZXJhdGluZyBzeXN0ZW0uICovXHJcbiAgICBPU0luZm86IE9TSW5mbztcclxuICAgIC8qKiBBZGRpdGlvbmFsIHBsYXRmb3JtIGluZm9ybWF0aW9uLiAqL1xyXG4gICAgUGxhdGZvcm1JbmZvOiBSZWNvcmQ8c3RyaW5nLCBhbnk+O1xyXG59XHJcblxyXG4vKipcclxuICogUmV0cmlldmVzIGVudmlyb25tZW50IGRldGFpbHMuXHJcbiAqXHJcbiAqIEByZXR1cm5zIEEgcHJvbWlzZSB0aGF0IHJlc29sdmVzIHRvIGFuIG9iamVjdCBjb250YWluaW5nIE9TIGFuZCBzeXN0ZW0gYXJjaGl0ZWN0dXJlLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIEVudmlyb25tZW50KCk6IFByb21pc2U8RW52aXJvbm1lbnRJbmZvPiB7XHJcbiAgICByZXR1cm4gY2FsbChTeXN0ZW1FbnZpcm9ubWVudCk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDaGVja3MgaWYgdGhlIGN1cnJlbnQgb3BlcmF0aW5nIHN5c3RlbSBpcyBXaW5kb3dzLlxyXG4gKlxyXG4gKiBAcmV0dXJuIFRydWUgaWYgdGhlIG9wZXJhdGluZyBzeXN0ZW0gaXMgV2luZG93cywgb3RoZXJ3aXNlIGZhbHNlLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIElzV2luZG93cygpOiBib29sZWFuIHtcclxuICAgIHJldHVybiAod2luZG93IGFzIGFueSkuX3dhaWxzPy5lbnZpcm9ubWVudD8uT1MgPT09IFwid2luZG93c1wiO1xyXG59XHJcblxyXG4vKipcclxuICogQ2hlY2tzIGlmIHRoZSBjdXJyZW50IG9wZXJhdGluZyBzeXN0ZW0gaXMgTGludXguXHJcbiAqXHJcbiAqIEByZXR1cm5zIFJldHVybnMgdHJ1ZSBpZiB0aGUgY3VycmVudCBvcGVyYXRpbmcgc3lzdGVtIGlzIExpbnV4LCBmYWxzZSBvdGhlcndpc2UuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSXNMaW51eCgpOiBib29sZWFuIHtcclxuICAgIHJldHVybiAod2luZG93IGFzIGFueSkuX3dhaWxzPy5lbnZpcm9ubWVudD8uT1MgPT09IFwibGludXhcIjtcclxufVxyXG5cclxuLyoqXHJcbiAqIENoZWNrcyBpZiB0aGUgY3VycmVudCBlbnZpcm9ubWVudCBpcyBhIG1hY09TIG9wZXJhdGluZyBzeXN0ZW0uXHJcbiAqXHJcbiAqIEByZXR1cm5zIFRydWUgaWYgdGhlIGVudmlyb25tZW50IGlzIG1hY09TLCBmYWxzZSBvdGhlcndpc2UuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSXNNYWMoKTogYm9vbGVhbiB7XHJcbiAgICByZXR1cm4gKHdpbmRvdyBhcyBhbnkpLl93YWlscz8uZW52aXJvbm1lbnQ/Lk9TID09PSBcImRhcndpblwiO1xyXG59XHJcblxyXG4vKipcclxuICogQ2hlY2tzIGlmIHRoZSBjdXJyZW50IGVudmlyb25tZW50IGFyY2hpdGVjdHVyZSBpcyBBTUQ2NC5cclxuICpcclxuICogQHJldHVybnMgVHJ1ZSBpZiB0aGUgY3VycmVudCBlbnZpcm9ubWVudCBhcmNoaXRlY3R1cmUgaXMgQU1ENjQsIGZhbHNlIG90aGVyd2lzZS5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBJc0FNRDY0KCk6IGJvb2xlYW4ge1xyXG4gICAgcmV0dXJuICh3aW5kb3cgYXMgYW55KS5fd2FpbHM/LmVudmlyb25tZW50Py5BcmNoID09PSBcImFtZDY0XCI7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDaGVja3MgaWYgdGhlIGN1cnJlbnQgYXJjaGl0ZWN0dXJlIGlzIEFSTS5cclxuICpcclxuICogQHJldHVybnMgVHJ1ZSBpZiB0aGUgY3VycmVudCBhcmNoaXRlY3R1cmUgaXMgQVJNLCBmYWxzZSBvdGhlcndpc2UuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSXNBUk0oKTogYm9vbGVhbiB7XHJcbiAgICByZXR1cm4gKHdpbmRvdyBhcyBhbnkpLl93YWlscz8uZW52aXJvbm1lbnQ/LkFyY2ggPT09IFwiYXJtXCI7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDaGVja3MgaWYgdGhlIGN1cnJlbnQgZW52aXJvbm1lbnQgaXMgQVJNNjQgYXJjaGl0ZWN0dXJlLlxyXG4gKlxyXG4gKiBAcmV0dXJucyBSZXR1cm5zIHRydWUgaWYgdGhlIGVudmlyb25tZW50IGlzIEFSTTY0IGFyY2hpdGVjdHVyZSwgb3RoZXJ3aXNlIHJldHVybnMgZmFsc2UuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSXNBUk02NCgpOiBib29sZWFuIHtcclxuICAgIHJldHVybiAod2luZG93IGFzIGFueSkuX3dhaWxzPy5lbnZpcm9ubWVudD8uQXJjaCA9PT0gXCJhcm02NFwiO1xyXG59XHJcblxyXG4vKipcclxuICogUmVwb3J0cyB3aGV0aGVyIHRoZSBhcHAgaXMgYmVpbmcgcnVuIGluIGRlYnVnIG1vZGUuXHJcbiAqXHJcbiAqIEByZXR1cm5zIFRydWUgaWYgdGhlIGFwcCBpcyBiZWluZyBydW4gaW4gZGVidWcgbW9kZS5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBJc0RlYnVnKCk6IGJvb2xlYW4ge1xyXG4gICAgcmV0dXJuIEJvb2xlYW4oKHdpbmRvdyBhcyBhbnkpLl93YWlscz8uZW52aXJvbm1lbnQ/LkRlYnVnKTtcclxufVxyXG5cclxuIiwgIi8qXHJcbiBfXHQgICBfX1x0ICBfIF9fXHJcbnwgfFx0IC8gL19fXyBfKF8pIC9fX19fXHJcbnwgfCAvfCAvIC8gX18gYC8gLyAvIF9fXy9cclxufCB8LyB8LyAvIC9fLyAvIC8gKF9fICApXHJcbnxfXy98X18vXFxfXyxfL18vXy9fX19fL1xyXG5UaGUgZWxlY3Ryb24gYWx0ZXJuYXRpdmUgZm9yIEdvXHJcbihjKSBMZWEgQW50aG9ueSAyMDE5LXByZXNlbnRcclxuKi9cclxuXHJcbmltcG9ydCB7IG5ld1J1bnRpbWVDYWxsZXIsIG9iamVjdE5hbWVzIH0gZnJvbSBcIi4vcnVudGltZS5qc1wiO1xyXG5pbXBvcnQgeyBJc0RlYnVnIH0gZnJvbSBcIi4vc3lzdGVtLmpzXCI7XHJcbmltcG9ydCB7IGV2ZW50VGFyZ2V0IH0gZnJvbSBcIi4vdXRpbHMuanNcIjtcclxuXHJcbi8vIHNldHVwXHJcbndpbmRvdy5hZGRFdmVudExpc3RlbmVyKCdjb250ZXh0bWVudScsIGNvbnRleHRNZW51SGFuZGxlcik7XHJcblxyXG5jb25zdCBjYWxsID0gbmV3UnVudGltZUNhbGxlcihvYmplY3ROYW1lcy5Db250ZXh0TWVudSk7XHJcblxyXG5jb25zdCBDb250ZXh0TWVudU9wZW4gPSAwO1xyXG5cclxuZnVuY3Rpb24gb3BlbkNvbnRleHRNZW51KGlkOiBzdHJpbmcsIHg6IG51bWJlciwgeTogbnVtYmVyLCBkYXRhOiBhbnkpOiB2b2lkIHtcclxuICAgIHZvaWQgY2FsbChDb250ZXh0TWVudU9wZW4sIHtpZCwgeCwgeSwgZGF0YX0pO1xyXG59XHJcblxyXG5mdW5jdGlvbiBjb250ZXh0TWVudUhhbmRsZXIoZXZlbnQ6IE1vdXNlRXZlbnQpIHtcclxuICAgIGNvbnN0IHRhcmdldCA9IGV2ZW50VGFyZ2V0KGV2ZW50KTtcclxuXHJcbiAgICAvLyBDaGVjayBmb3IgY3VzdG9tIGNvbnRleHQgbWVudVxyXG4gICAgY29uc3QgY3VzdG9tQ29udGV4dE1lbnUgPSB3aW5kb3cuZ2V0Q29tcHV0ZWRTdHlsZSh0YXJnZXQpLmdldFByb3BlcnR5VmFsdWUoXCItLWN1c3RvbS1jb250ZXh0bWVudVwiKS50cmltKCk7XHJcblxyXG4gICAgaWYgKGN1c3RvbUNvbnRleHRNZW51KSB7XHJcbiAgICAgICAgZXZlbnQucHJldmVudERlZmF1bHQoKTtcclxuICAgICAgICBjb25zdCBkYXRhID0gd2luZG93LmdldENvbXB1dGVkU3R5bGUodGFyZ2V0KS5nZXRQcm9wZXJ0eVZhbHVlKFwiLS1jdXN0b20tY29udGV4dG1lbnUtZGF0YVwiKTtcclxuICAgICAgICBvcGVuQ29udGV4dE1lbnUoY3VzdG9tQ29udGV4dE1lbnUsIGV2ZW50LmNsaWVudFgsIGV2ZW50LmNsaWVudFksIGRhdGEpO1xyXG4gICAgfSBlbHNlIHtcclxuICAgICAgICBwcm9jZXNzRGVmYXVsdENvbnRleHRNZW51KGV2ZW50LCB0YXJnZXQpO1xyXG4gICAgfVxyXG59XHJcblxyXG5cclxuLypcclxuLS1kZWZhdWx0LWNvbnRleHRtZW51OiBhdXRvOyAoZGVmYXVsdCkgd2lsbCBzaG93IHRoZSBkZWZhdWx0IGNvbnRleHQgbWVudSBpZiBjb250ZW50RWRpdGFibGUgaXMgdHJ1ZSBPUiB0ZXh0IGhhcyBiZWVuIHNlbGVjdGVkIE9SIGVsZW1lbnQgaXMgaW5wdXQgb3IgdGV4dGFyZWFcclxuLS1kZWZhdWx0LWNvbnRleHRtZW51OiBzaG93OyB3aWxsIGFsd2F5cyBzaG93IHRoZSBkZWZhdWx0IGNvbnRleHQgbWVudVxyXG4tLWRlZmF1bHQtY29udGV4dG1lbnU6IGhpZGU7IHdpbGwgYWx3YXlzIGhpZGUgdGhlIGRlZmF1bHQgY29udGV4dCBtZW51XHJcblxyXG5UaGlzIHJ1bGUgaXMgaW5oZXJpdGVkIGxpa2Ugbm9ybWFsIENTUyBydWxlcywgc28gbmVzdGluZyB3b3JrcyBhcyBleHBlY3RlZFxyXG4qL1xyXG5mdW5jdGlvbiBwcm9jZXNzRGVmYXVsdENvbnRleHRNZW51KGV2ZW50OiBNb3VzZUV2ZW50LCB0YXJnZXQ6IEhUTUxFbGVtZW50KSB7XHJcbiAgICAvLyBEZWJ1ZyBidWlsZHMgYWx3YXlzIHNob3cgdGhlIG1lbnVcclxuICAgIGlmIChJc0RlYnVnKCkpIHtcclxuICAgICAgICByZXR1cm47XHJcbiAgICB9XHJcblxyXG4gICAgLy8gUHJvY2VzcyBkZWZhdWx0IGNvbnRleHQgbWVudVxyXG4gICAgc3dpdGNoICh3aW5kb3cuZ2V0Q29tcHV0ZWRTdHlsZSh0YXJnZXQpLmdldFByb3BlcnR5VmFsdWUoXCItLWRlZmF1bHQtY29udGV4dG1lbnVcIikudHJpbSgpKSB7XHJcbiAgICAgICAgY2FzZSAnc2hvdyc6XHJcbiAgICAgICAgICAgIHJldHVybjtcclxuICAgICAgICBjYXNlICdoaWRlJzpcclxuICAgICAgICAgICAgZXZlbnQucHJldmVudERlZmF1bHQoKTtcclxuICAgICAgICAgICAgcmV0dXJuO1xyXG4gICAgfVxyXG5cclxuICAgIC8vIENoZWNrIGlmIGNvbnRlbnRFZGl0YWJsZSBpcyB0cnVlXHJcbiAgICBpZiAodGFyZ2V0LmlzQ29udGVudEVkaXRhYmxlKSB7XHJcbiAgICAgICAgcmV0dXJuO1xyXG4gICAgfVxyXG5cclxuICAgIC8vIENoZWNrIGlmIHRleHQgaGFzIGJlZW4gc2VsZWN0ZWRcclxuICAgIGNvbnN0IHNlbGVjdGlvbiA9IHdpbmRvdy5nZXRTZWxlY3Rpb24oKTtcclxuICAgIGNvbnN0IGhhc1NlbGVjdGlvbiA9IHNlbGVjdGlvbiAmJiBzZWxlY3Rpb24udG9TdHJpbmcoKS5sZW5ndGggPiAwO1xyXG4gICAgaWYgKGhhc1NlbGVjdGlvbikge1xyXG4gICAgICAgIGZvciAobGV0IGkgPSAwOyBpIDwgc2VsZWN0aW9uLnJhbmdlQ291bnQ7IGkrKykge1xyXG4gICAgICAgICAgICBjb25zdCByYW5nZSA9IHNlbGVjdGlvbi5nZXRSYW5nZUF0KGkpO1xyXG4gICAgICAgICAgICBjb25zdCByZWN0cyA9IHJhbmdlLmdldENsaWVudFJlY3RzKCk7XHJcbiAgICAgICAgICAgIGZvciAobGV0IGogPSAwOyBqIDwgcmVjdHMubGVuZ3RoOyBqKyspIHtcclxuICAgICAgICAgICAgICAgIGNvbnN0IHJlY3QgPSByZWN0c1tqXTtcclxuICAgICAgICAgICAgICAgIGlmIChkb2N1bWVudC5lbGVtZW50RnJvbVBvaW50KHJlY3QubGVmdCwgcmVjdC50b3ApID09PSB0YXJnZXQpIHtcclxuICAgICAgICAgICAgICAgICAgICByZXR1cm47XHJcbiAgICAgICAgICAgICAgICB9XHJcbiAgICAgICAgICAgIH1cclxuICAgICAgICB9XHJcbiAgICB9XHJcblxyXG4gICAgLy8gQ2hlY2sgaWYgdGFnIGlzIGlucHV0IG9yIHRleHRhcmVhLlxyXG4gICAgaWYgKHRhcmdldCBpbnN0YW5jZW9mIEhUTUxJbnB1dEVsZW1lbnQgfHwgdGFyZ2V0IGluc3RhbmNlb2YgSFRNTFRleHRBcmVhRWxlbWVudCkge1xyXG4gICAgICAgIGlmIChoYXNTZWxlY3Rpb24gfHwgKCF0YXJnZXQucmVhZE9ubHkgJiYgIXRhcmdldC5kaXNhYmxlZCkpIHtcclxuICAgICAgICAgICAgcmV0dXJuO1xyXG4gICAgICAgIH1cclxuICAgIH1cclxuXHJcbiAgICAvLyBoaWRlIGRlZmF1bHQgY29udGV4dCBtZW51XHJcbiAgICBldmVudC5wcmV2ZW50RGVmYXVsdCgpO1xyXG59XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG4vKipcclxuICogUmV0cmlldmVzIHRoZSB2YWx1ZSBhc3NvY2lhdGVkIHdpdGggdGhlIHNwZWNpZmllZCBrZXkgZnJvbSB0aGUgZmxhZyBtYXAuXHJcbiAqXHJcbiAqIEBwYXJhbSBrZXkgLSBUaGUga2V5IHRvIHJldHJpZXZlIHRoZSB2YWx1ZSBmb3IuXHJcbiAqIEByZXR1cm4gVGhlIHZhbHVlIGFzc29jaWF0ZWQgd2l0aCB0aGUgc3BlY2lmaWVkIGtleS5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBHZXRGbGFnKGtleTogc3RyaW5nKTogYW55IHtcclxuICAgIHRyeSB7XHJcbiAgICAgICAgcmV0dXJuIHdpbmRvdy5fd2FpbHMuZmxhZ3Nba2V5XTtcclxuICAgIH0gY2F0Y2ggKGUpIHtcclxuICAgICAgICB0aHJvdyBuZXcgRXJyb3IoXCJVbmFibGUgdG8gcmV0cmlldmUgZmxhZyAnXCIgKyBrZXkgKyBcIic6IFwiICsgZSwgeyBjYXVzZTogZSB9KTtcclxuICAgIH1cclxufVxyXG4iLCAiLypcclxuIF9cdCAgIF9fXHQgIF8gX19cclxufCB8XHQgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuaW1wb3J0IHsgaW52b2tlLCBJc1dpbmRvd3MgfSBmcm9tIFwiLi9zeXN0ZW0uanNcIjtcclxuaW1wb3J0IHsgR2V0RmxhZyB9IGZyb20gXCIuL2ZsYWdzLmpzXCI7XHJcbmltcG9ydCB7IGNhblRyYWNrQnV0dG9ucywgZXZlbnRUYXJnZXQgfSBmcm9tIFwiLi91dGlscy5qc1wiO1xyXG5cclxuLy8gU2V0dXBcclxubGV0IGNhbkRyYWcgPSBmYWxzZTtcclxubGV0IGRyYWdnaW5nID0gZmFsc2U7XHJcblxyXG5sZXQgcmVzaXphYmxlID0gZmFsc2U7XHJcbmxldCBjYW5SZXNpemUgPSBmYWxzZTtcclxubGV0IHJlc2l6aW5nID0gZmFsc2U7XHJcbmxldCByZXNpemVFZGdlOiBzdHJpbmcgPSBcIlwiO1xyXG5sZXQgZGVmYXVsdEN1cnNvciA9IFwiYXV0b1wiO1xyXG5cclxubGV0IGJ1dHRvbnMgPSAwO1xyXG5jb25zdCBidXR0b25zVHJhY2tlZCA9IGNhblRyYWNrQnV0dG9ucygpO1xyXG5cclxud2luZG93Ll93YWlscyA9IHdpbmRvdy5fd2FpbHMgfHwge307XHJcbndpbmRvdy5fd2FpbHMuc2V0UmVzaXphYmxlID0gKHZhbHVlOiBib29sZWFuKTogdm9pZCA9PiB7XHJcbiAgICByZXNpemFibGUgPSB2YWx1ZTtcclxuICAgIGlmICghcmVzaXphYmxlKSB7XHJcbiAgICAgICAgLy8gU3RvcCByZXNpemluZyBpZiBpbiBwcm9ncmVzcy5cclxuICAgICAgICBjYW5SZXNpemUgPSByZXNpemluZyA9IGZhbHNlO1xyXG4gICAgICAgIHNldFJlc2l6ZSgpO1xyXG4gICAgfVxyXG59O1xyXG5cclxuLy8gRGVmZXIgYXR0YWNoaW5nIG1vdXNlIGxpc3RlbmVycyB1bnRpbCB3ZSBrbm93IHdlJ3JlIG5vdCBvbiBtb2JpbGUuXHJcbmxldCBkcmFnSW5pdERvbmUgPSBmYWxzZTtcclxuZnVuY3Rpb24gaXNNb2JpbGUoKTogYm9vbGVhbiB7XHJcbiAgICBjb25zdCBvcyA9ICh3aW5kb3cgYXMgYW55KS5fd2FpbHM/LmVudmlyb25tZW50Py5PUztcclxuICAgIGlmIChvcyA9PT0gXCJpb3NcIiB8fCBvcyA9PT0gXCJhbmRyb2lkXCIpIHJldHVybiB0cnVlO1xyXG4gICAgLy8gRmFsbGJhY2sgaGV1cmlzdGljIGlmIGVudmlyb25tZW50IG5vdCB5ZXQgc2V0XHJcbiAgICBjb25zdCB1YSA9IG5hdmlnYXRvci51c2VyQWdlbnQgfHwgbmF2aWdhdG9yLnZlbmRvciB8fCAod2luZG93IGFzIGFueSkub3BlcmEgfHwgXCJcIjtcclxuICAgIHJldHVybiAvYW5kcm9pZHxpcGhvbmV8aXBhZHxpcG9kfGllbW9iaWxlfHdwZGVza3RvcC9pLnRlc3QodWEpO1xyXG59XHJcbmZ1bmN0aW9uIHRyeUluaXREcmFnSGFuZGxlcnMoKTogdm9pZCB7XHJcbiAgICBpZiAoZHJhZ0luaXREb25lKSByZXR1cm47XHJcbiAgICBpZiAoaXNNb2JpbGUoKSkgcmV0dXJuO1xyXG4gICAgd2luZG93LmFkZEV2ZW50TGlzdGVuZXIoJ21vdXNlZG93bicsIHVwZGF0ZSwgeyBjYXB0dXJlOiB0cnVlIH0pO1xyXG4gICAgd2luZG93LmFkZEV2ZW50TGlzdGVuZXIoJ21vdXNlbW92ZScsIHVwZGF0ZSwgeyBjYXB0dXJlOiB0cnVlIH0pO1xyXG4gICAgd2luZG93LmFkZEV2ZW50TGlzdGVuZXIoJ21vdXNldXAnLCB1cGRhdGUsIHsgY2FwdHVyZTogdHJ1ZSB9KTtcclxuICAgIGZvciAoY29uc3QgZXYgb2YgWydjbGljaycsICdjb250ZXh0bWVudScsICdkYmxjbGljayddKSB7XHJcbiAgICAgICAgd2luZG93LmFkZEV2ZW50TGlzdGVuZXIoZXYsIHN1cHByZXNzRXZlbnQsIHsgY2FwdHVyZTogdHJ1ZSB9KTtcclxuICAgIH1cclxuICAgIGRyYWdJbml0RG9uZSA9IHRydWU7XHJcbn1cclxuLy8gQXR0ZW1wdCBpbW1lZGlhdGUgaW5pdCAoaW4gY2FzZSBlbnZpcm9ubWVudCBhbHJlYWR5IHByZXNlbnQpXHJcbnRyeUluaXREcmFnSGFuZGxlcnMoKTtcclxuLy8gQWxzbyBhdHRlbXB0IG9uIERPTSByZWFkeVxyXG5kb2N1bWVudC5hZGRFdmVudExpc3RlbmVyKCdET01Db250ZW50TG9hZGVkJywgdHJ5SW5pdERyYWdIYW5kbGVycywgeyBvbmNlOiB0cnVlIH0pO1xyXG4vLyBBcyBhIGxhc3QgcmVzb3J0LCBwb2xsIGZvciBlbnZpcm9ubWVudCBmb3IgYSBzaG9ydCBwZXJpb2RcclxubGV0IGRyYWdFbnZQb2xscyA9IDA7XHJcbmNvbnN0IGRyYWdFbnZQb2xsID0gd2luZG93LnNldEludGVydmFsKCgpID0+IHtcclxuICAgIGlmIChkcmFnSW5pdERvbmUpIHsgd2luZG93LmNsZWFySW50ZXJ2YWwoZHJhZ0VudlBvbGwpOyByZXR1cm47IH1cclxuICAgIHRyeUluaXREcmFnSGFuZGxlcnMoKTtcclxuICAgIGlmICgrK2RyYWdFbnZQb2xscyA+IDEwMCkgeyB3aW5kb3cuY2xlYXJJbnRlcnZhbChkcmFnRW52UG9sbCk7IH1cclxufSwgNTApO1xyXG5cclxuZnVuY3Rpb24gc3VwcHJlc3NFdmVudChldmVudDogRXZlbnQpIHtcclxuICAgIC8vIFN1cHByZXNzIGNsaWNrIGV2ZW50cyB3aGlsZSByZXNpemluZyBvciBkcmFnZ2luZy5cclxuICAgIGlmIChkcmFnZ2luZyB8fCByZXNpemluZykge1xyXG4gICAgICAgIGV2ZW50LnN0b3BJbW1lZGlhdGVQcm9wYWdhdGlvbigpO1xyXG4gICAgICAgIGV2ZW50LnN0b3BQcm9wYWdhdGlvbigpO1xyXG4gICAgICAgIGV2ZW50LnByZXZlbnREZWZhdWx0KCk7XHJcbiAgICB9XHJcbn1cclxuXHJcbi8vIFVzZSBjb25zdGFudHMgdG8gYXZvaWQgY29tcGFyaW5nIHN0cmluZ3MgbXVsdGlwbGUgdGltZXMuXHJcbmNvbnN0IE1vdXNlRG93biA9IDA7XHJcbmNvbnN0IE1vdXNlVXAgICA9IDE7XHJcbmNvbnN0IE1vdXNlTW92ZSA9IDI7XHJcblxyXG5mdW5jdGlvbiB1cGRhdGUoZXZlbnQ6IE1vdXNlRXZlbnQpIHtcclxuICAgIC8vIFdpbmRvd3Mgc3VwcHJlc3NlcyBtb3VzZSBldmVudHMgYXQgdGhlIGVuZCBvZiBkcmFnZ2luZyBvciByZXNpemluZyxcclxuICAgIC8vIHNvIHdlIG5lZWQgdG8gYmUgc21hcnQgYW5kIHN5bnRoZXNpemUgYnV0dG9uIGV2ZW50cy5cclxuXHJcbiAgICBsZXQgZXZlbnRUeXBlOiBudW1iZXIsIGV2ZW50QnV0dG9ucyA9IGV2ZW50LmJ1dHRvbnM7XHJcbiAgICBzd2l0Y2ggKGV2ZW50LnR5cGUpIHtcclxuICAgICAgICBjYXNlICdtb3VzZWRvd24nOlxyXG4gICAgICAgICAgICBldmVudFR5cGUgPSBNb3VzZURvd247XHJcbiAgICAgICAgICAgIGlmICghYnV0dG9uc1RyYWNrZWQpIHsgZXZlbnRCdXR0b25zID0gYnV0dG9ucyB8ICgxIDw8IGV2ZW50LmJ1dHRvbik7IH1cclxuICAgICAgICAgICAgYnJlYWs7XHJcbiAgICAgICAgY2FzZSAnbW91c2V1cCc6XHJcbiAgICAgICAgICAgIGV2ZW50VHlwZSA9IE1vdXNlVXA7XHJcbiAgICAgICAgICAgIGlmICghYnV0dG9uc1RyYWNrZWQpIHsgZXZlbnRCdXR0b25zID0gYnV0dG9ucyAmIH4oMSA8PCBldmVudC5idXR0b24pOyB9XHJcbiAgICAgICAgICAgIGJyZWFrO1xyXG4gICAgICAgIGRlZmF1bHQ6XHJcbiAgICAgICAgICAgIGV2ZW50VHlwZSA9IE1vdXNlTW92ZTtcclxuICAgICAgICAgICAgaWYgKCFidXR0b25zVHJhY2tlZCkgeyBldmVudEJ1dHRvbnMgPSBidXR0b25zOyB9XHJcbiAgICAgICAgICAgIGJyZWFrO1xyXG4gICAgfVxyXG5cclxuICAgIGxldCByZWxlYXNlZCA9IGJ1dHRvbnMgJiB+ZXZlbnRCdXR0b25zO1xyXG4gICAgbGV0IHByZXNzZWQgPSBldmVudEJ1dHRvbnMgJiB+YnV0dG9ucztcclxuXHJcbiAgICBidXR0b25zID0gZXZlbnRCdXR0b25zO1xyXG5cclxuICAgIC8vIFN5bnRoZXNpemUgYSByZWxlYXNlLXByZXNzIHNlcXVlbmNlIGlmIHdlIGRldGVjdCBhIHByZXNzIG9mIGFuIGFscmVhZHkgcHJlc3NlZCBidXR0b24uXHJcbiAgICBpZiAoZXZlbnRUeXBlID09PSBNb3VzZURvd24gJiYgIShwcmVzc2VkICYgZXZlbnQuYnV0dG9uKSkge1xyXG4gICAgICAgIHJlbGVhc2VkIHw9ICgxIDw8IGV2ZW50LmJ1dHRvbik7XHJcbiAgICAgICAgcHJlc3NlZCB8PSAoMSA8PCBldmVudC5idXR0b24pO1xyXG4gICAgfVxyXG5cclxuICAgIC8vIFN1cHByZXNzIGFsbCBidXR0b24gZXZlbnRzIGR1cmluZyBkcmFnZ2luZyBhbmQgcmVzaXppbmcsXHJcbiAgICAvLyB1bmxlc3MgdGhpcyBpcyBhIG1vdXNldXAgZXZlbnQgdGhhdCBpcyBlbmRpbmcgYSBkcmFnIGFjdGlvbi5cclxuICAgIGlmIChcclxuICAgICAgICBldmVudFR5cGUgIT09IE1vdXNlTW92ZSAvLyBGYXN0IHBhdGggZm9yIG1vdXNlbW92ZVxyXG4gICAgICAgICYmIHJlc2l6aW5nXHJcbiAgICAgICAgfHwgKFxyXG4gICAgICAgICAgICBkcmFnZ2luZ1xyXG4gICAgICAgICAgICAmJiAoXHJcbiAgICAgICAgICAgICAgICBldmVudFR5cGUgPT09IE1vdXNlRG93blxyXG4gICAgICAgICAgICAgICAgfHwgZXZlbnQuYnV0dG9uICE9PSAwXHJcbiAgICAgICAgICAgIClcclxuICAgICAgICApXHJcbiAgICApIHtcclxuICAgICAgICBldmVudC5zdG9wSW1tZWRpYXRlUHJvcGFnYXRpb24oKTtcclxuICAgICAgICBldmVudC5zdG9wUHJvcGFnYXRpb24oKTtcclxuICAgICAgICBldmVudC5wcmV2ZW50RGVmYXVsdCgpO1xyXG4gICAgfVxyXG5cclxuICAgIC8vIEhhbmRsZSByZWxlYXNlc1xyXG4gICAgaWYgKHJlbGVhc2VkICYgMSkgeyBwcmltYXJ5VXAoZXZlbnQpOyB9XHJcbiAgICAvLyBIYW5kbGUgcHJlc3Nlc1xyXG4gICAgaWYgKHByZXNzZWQgJiAxKSB7IHByaW1hcnlEb3duKGV2ZW50KTsgfVxyXG5cclxuICAgIC8vIEhhbmRsZSBtb3VzZW1vdmVcclxuICAgIGlmIChldmVudFR5cGUgPT09IE1vdXNlTW92ZSkgeyBvbk1vdXNlTW92ZShldmVudCk7IH07XHJcbn1cclxuXHJcbmZ1bmN0aW9uIHByaW1hcnlEb3duKGV2ZW50OiBNb3VzZUV2ZW50KTogdm9pZCB7XHJcbiAgICAvLyBSZXNldCByZWFkaW5lc3Mgc3RhdGUuXHJcbiAgICBjYW5EcmFnID0gZmFsc2U7XHJcbiAgICBjYW5SZXNpemUgPSBmYWxzZTtcclxuXHJcbiAgICAvLyBJZ25vcmUgcmVwZWF0ZWQgY2xpY2tzIG9uIG1hY09TIGFuZCBMaW51eC5cclxuICAgIGlmICghSXNXaW5kb3dzKCkpIHtcclxuICAgICAgICBpZiAoZXZlbnQudHlwZSA9PT0gJ21vdXNlZG93bicgJiYgZXZlbnQuYnV0dG9uID09PSAwICYmIGV2ZW50LmRldGFpbCAhPT0gMSkge1xyXG4gICAgICAgICAgICByZXR1cm47XHJcbiAgICAgICAgfVxyXG4gICAgfVxyXG5cclxuICAgIGlmIChyZXNpemVFZGdlKSB7XHJcbiAgICAgICAgLy8gUmVhZHkgdG8gcmVzaXplIGlmIHRoZSBwcmltYXJ5IGJ1dHRvbiB3YXMgcHJlc3NlZCBmb3IgdGhlIGZpcnN0IHRpbWUuXHJcbiAgICAgICAgY2FuUmVzaXplID0gdHJ1ZTtcclxuICAgICAgICAvLyBEbyBub3Qgc3RhcnQgZHJhZyBvcGVyYXRpb25zIHdoZW4gb24gcmVzaXplIGVkZ2VzLlxyXG4gICAgICAgIHJldHVybjtcclxuICAgIH1cclxuXHJcbiAgICAvLyBSZXRyaWV2ZSB0YXJnZXQgZWxlbWVudFxyXG4gICAgY29uc3QgdGFyZ2V0ID0gZXZlbnRUYXJnZXQoZXZlbnQpO1xyXG5cclxuICAgIC8vIFJlYWR5IHRvIGRyYWcgaWYgdGhlIHByaW1hcnkgYnV0dG9uIHdhcyBwcmVzc2VkIGZvciB0aGUgZmlyc3QgdGltZSBvbiBhIGRyYWdnYWJsZSBlbGVtZW50LlxyXG4gICAgLy8gSWdub3JlIGNsaWNrcyBvbiB0aGUgc2Nyb2xsYmFyLlxyXG4gICAgY29uc3Qgc3R5bGUgPSB3aW5kb3cuZ2V0Q29tcHV0ZWRTdHlsZSh0YXJnZXQpO1xyXG4gICAgY2FuRHJhZyA9IChcclxuICAgICAgICBzdHlsZS5nZXRQcm9wZXJ0eVZhbHVlKFwiLS13YWlscy1kcmFnZ2FibGVcIikudHJpbSgpID09PSBcImRyYWdcIlxyXG4gICAgICAgICYmIChcclxuICAgICAgICAgICAgZXZlbnQub2Zmc2V0WCAtIHBhcnNlRmxvYXQoc3R5bGUucGFkZGluZ0xlZnQpIDwgdGFyZ2V0LmNsaWVudFdpZHRoXHJcbiAgICAgICAgICAgICYmIGV2ZW50Lm9mZnNldFkgLSBwYXJzZUZsb2F0KHN0eWxlLnBhZGRpbmdUb3ApIDwgdGFyZ2V0LmNsaWVudEhlaWdodFxyXG4gICAgICAgIClcclxuICAgICk7XHJcbn1cclxuXHJcbmZ1bmN0aW9uIHByaW1hcnlVcChldmVudDogTW91c2VFdmVudCkge1xyXG4gICAgLy8gU3RvcCBkcmFnZ2luZyBhbmQgcmVzaXppbmcuXHJcbiAgICBjYW5EcmFnID0gZmFsc2U7XHJcbiAgICBkcmFnZ2luZyA9IGZhbHNlO1xyXG4gICAgY2FuUmVzaXplID0gZmFsc2U7XHJcbiAgICByZXNpemluZyA9IGZhbHNlO1xyXG59XHJcblxyXG5jb25zdCBjdXJzb3JGb3JFZGdlID0gT2JqZWN0LmZyZWV6ZSh7XHJcbiAgICBcInNlLXJlc2l6ZVwiOiBcIm53c2UtcmVzaXplXCIsXHJcbiAgICBcInN3LXJlc2l6ZVwiOiBcIm5lc3ctcmVzaXplXCIsXHJcbiAgICBcIm53LXJlc2l6ZVwiOiBcIm53c2UtcmVzaXplXCIsXHJcbiAgICBcIm5lLXJlc2l6ZVwiOiBcIm5lc3ctcmVzaXplXCIsXHJcbiAgICBcInctcmVzaXplXCI6IFwiZXctcmVzaXplXCIsXHJcbiAgICBcIm4tcmVzaXplXCI6IFwibnMtcmVzaXplXCIsXHJcbiAgICBcInMtcmVzaXplXCI6IFwibnMtcmVzaXplXCIsXHJcbiAgICBcImUtcmVzaXplXCI6IFwiZXctcmVzaXplXCIsXHJcbn0pXHJcblxyXG5mdW5jdGlvbiBzZXRSZXNpemUoZWRnZT86IGtleW9mIHR5cGVvZiBjdXJzb3JGb3JFZGdlKTogdm9pZCB7XHJcbiAgICBpZiAoZWRnZSkge1xyXG4gICAgICAgIGlmICghcmVzaXplRWRnZSkgeyBkZWZhdWx0Q3Vyc29yID0gZG9jdW1lbnQuYm9keS5zdHlsZS5jdXJzb3I7IH1cclxuICAgICAgICBkb2N1bWVudC5ib2R5LnN0eWxlLmN1cnNvciA9IGN1cnNvckZvckVkZ2VbZWRnZV07XHJcbiAgICB9IGVsc2UgaWYgKCFlZGdlICYmIHJlc2l6ZUVkZ2UpIHtcclxuICAgICAgICBkb2N1bWVudC5ib2R5LnN0eWxlLmN1cnNvciA9IGRlZmF1bHRDdXJzb3I7XHJcbiAgICB9XHJcblxyXG4gICAgcmVzaXplRWRnZSA9IGVkZ2UgfHwgXCJcIjtcclxufVxyXG5cclxuZnVuY3Rpb24gb25Nb3VzZU1vdmUoZXZlbnQ6IE1vdXNlRXZlbnQpOiB2b2lkIHtcclxuICAgIGlmIChjYW5SZXNpemUgJiYgcmVzaXplRWRnZSkge1xyXG4gICAgICAgIC8vIFN0YXJ0IHJlc2l6aW5nLlxyXG4gICAgICAgIHJlc2l6aW5nID0gdHJ1ZTtcclxuICAgICAgICBpbnZva2UoXCJ3YWlsczpyZXNpemU6XCIgKyByZXNpemVFZGdlKTtcclxuICAgIH0gZWxzZSBpZiAoY2FuRHJhZykge1xyXG4gICAgICAgIC8vIFN0YXJ0IGRyYWdnaW5nLlxyXG4gICAgICAgIGRyYWdnaW5nID0gdHJ1ZTtcclxuICAgICAgICBpbnZva2UoXCJ3YWlsczpkcmFnXCIpO1xyXG4gICAgfVxyXG5cclxuICAgIGlmIChkcmFnZ2luZyB8fCByZXNpemluZykge1xyXG4gICAgICAgIC8vIEVpdGhlciBkcmFnIG9yIHJlc2l6ZSBpcyBvbmdvaW5nLFxyXG4gICAgICAgIC8vIHJlc2V0IHJlYWRpbmVzcyBhbmQgc3RvcCBwcm9jZXNzaW5nLlxyXG4gICAgICAgIGNhbkRyYWcgPSBjYW5SZXNpemUgPSBmYWxzZTtcclxuICAgICAgICByZXR1cm47XHJcbiAgICB9XHJcblxyXG4gICAgaWYgKCFyZXNpemFibGUgfHwgIUlzV2luZG93cygpKSB7XHJcbiAgICAgICAgaWYgKHJlc2l6ZUVkZ2UpIHsgc2V0UmVzaXplKCk7IH1cclxuICAgICAgICByZXR1cm47XHJcbiAgICB9XHJcblxyXG4gICAgY29uc3QgcmVzaXplSGFuZGxlSGVpZ2h0ID0gR2V0RmxhZyhcInN5c3RlbS5yZXNpemVIYW5kbGVIZWlnaHRcIikgfHwgNTtcclxuICAgIGNvbnN0IHJlc2l6ZUhhbmRsZVdpZHRoID0gR2V0RmxhZyhcInN5c3RlbS5yZXNpemVIYW5kbGVXaWR0aFwiKSB8fCA1O1xyXG5cclxuICAgIC8vIEV4dHJhIHBpeGVscyBmb3IgdGhlIGNvcm5lciBhcmVhcy5cclxuICAgIGNvbnN0IGNvcm5lckV4dHJhID0gR2V0RmxhZyhcInJlc2l6ZUNvcm5lckV4dHJhXCIpIHx8IDEwO1xyXG5cclxuICAgIGNvbnN0IHJpZ2h0Qm9yZGVyID0gKHdpbmRvdy5vdXRlcldpZHRoIC0gZXZlbnQuY2xpZW50WCkgPCByZXNpemVIYW5kbGVXaWR0aDtcclxuICAgIGNvbnN0IGxlZnRCb3JkZXIgPSBldmVudC5jbGllbnRYIDwgcmVzaXplSGFuZGxlV2lkdGg7XHJcbiAgICBjb25zdCB0b3BCb3JkZXIgPSBldmVudC5jbGllbnRZIDwgcmVzaXplSGFuZGxlSGVpZ2h0O1xyXG4gICAgY29uc3QgYm90dG9tQm9yZGVyID0gKHdpbmRvdy5vdXRlckhlaWdodCAtIGV2ZW50LmNsaWVudFkpIDwgcmVzaXplSGFuZGxlSGVpZ2h0O1xyXG5cclxuICAgIC8vIEFkanVzdCBmb3IgY29ybmVyIGFyZWFzLlxyXG4gICAgY29uc3QgcmlnaHRDb3JuZXIgPSAod2luZG93Lm91dGVyV2lkdGggLSBldmVudC5jbGllbnRYKSA8IChyZXNpemVIYW5kbGVXaWR0aCArIGNvcm5lckV4dHJhKTtcclxuICAgIGNvbnN0IGxlZnRDb3JuZXIgPSBldmVudC5jbGllbnRYIDwgKHJlc2l6ZUhhbmRsZVdpZHRoICsgY29ybmVyRXh0cmEpO1xyXG4gICAgY29uc3QgdG9wQ29ybmVyID0gZXZlbnQuY2xpZW50WSA8IChyZXNpemVIYW5kbGVIZWlnaHQgKyBjb3JuZXJFeHRyYSk7XHJcbiAgICBjb25zdCBib3R0b21Db3JuZXIgPSAod2luZG93Lm91dGVySGVpZ2h0IC0gZXZlbnQuY2xpZW50WSkgPCAocmVzaXplSGFuZGxlSGVpZ2h0ICsgY29ybmVyRXh0cmEpO1xyXG5cclxuICAgIGlmICghbGVmdENvcm5lciAmJiAhdG9wQ29ybmVyICYmICFib3R0b21Db3JuZXIgJiYgIXJpZ2h0Q29ybmVyKSB7XHJcbiAgICAgICAgLy8gT3B0aW1pc2F0aW9uOiBvdXQgb2YgYWxsIGNvcm5lciBhcmVhcyBpbXBsaWVzIG91dCBvZiBib3JkZXJzLlxyXG4gICAgICAgIHNldFJlc2l6ZSgpO1xyXG4gICAgfVxyXG4gICAgLy8gRGV0ZWN0IGNvcm5lcnMuXHJcbiAgICBlbHNlIGlmIChyaWdodENvcm5lciAmJiBib3R0b21Db3JuZXIpIHNldFJlc2l6ZShcInNlLXJlc2l6ZVwiKTtcclxuICAgIGVsc2UgaWYgKGxlZnRDb3JuZXIgJiYgYm90dG9tQ29ybmVyKSBzZXRSZXNpemUoXCJzdy1yZXNpemVcIik7XHJcbiAgICBlbHNlIGlmIChsZWZ0Q29ybmVyICYmIHRvcENvcm5lcikgc2V0UmVzaXplKFwibnctcmVzaXplXCIpO1xyXG4gICAgZWxzZSBpZiAodG9wQ29ybmVyICYmIHJpZ2h0Q29ybmVyKSBzZXRSZXNpemUoXCJuZS1yZXNpemVcIik7XHJcbiAgICAvLyBEZXRlY3QgYm9yZGVycy5cclxuICAgIGVsc2UgaWYgKGxlZnRCb3JkZXIpIHNldFJlc2l6ZShcInctcmVzaXplXCIpO1xyXG4gICAgZWxzZSBpZiAodG9wQm9yZGVyKSBzZXRSZXNpemUoXCJuLXJlc2l6ZVwiKTtcclxuICAgIGVsc2UgaWYgKGJvdHRvbUJvcmRlcikgc2V0UmVzaXplKFwicy1yZXNpemVcIik7XHJcbiAgICBlbHNlIGlmIChyaWdodEJvcmRlcikgc2V0UmVzaXplKFwiZS1yZXNpemVcIik7XHJcbiAgICAvLyBPdXQgb2YgYm9yZGVyIGFyZWEuXHJcbiAgICBlbHNlIHNldFJlc2l6ZSgpO1xyXG59XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQgeyBuZXdSdW50aW1lQ2FsbGVyLCBvYmplY3ROYW1lcyB9IGZyb20gXCIuL3J1bnRpbWUuanNcIjtcclxuY29uc3QgY2FsbCA9IG5ld1J1bnRpbWVDYWxsZXIob2JqZWN0TmFtZXMuQXBwbGljYXRpb24pO1xyXG5cclxuY29uc3QgSGlkZU1ldGhvZCA9IDA7XHJcbmNvbnN0IFNob3dNZXRob2QgPSAxO1xyXG5jb25zdCBRdWl0TWV0aG9kID0gMjtcclxuXHJcbi8qKlxyXG4gKiBIaWRlcyBhIGNlcnRhaW4gbWV0aG9kIGJ5IGNhbGxpbmcgdGhlIEhpZGVNZXRob2QgZnVuY3Rpb24uXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gSGlkZSgpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgIHJldHVybiBjYWxsKEhpZGVNZXRob2QpO1xyXG59XHJcblxyXG4vKipcclxuICogQ2FsbHMgdGhlIFNob3dNZXRob2QgYW5kIHJldHVybnMgdGhlIHJlc3VsdC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBTaG93KCk6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgcmV0dXJuIGNhbGwoU2hvd01ldGhvZCk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDYWxscyB0aGUgUXVpdE1ldGhvZCB0byB0ZXJtaW5hdGUgdGhlIHByb2dyYW0uXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gUXVpdCgpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgIHJldHVybiBjYWxsKFF1aXRNZXRob2QpO1xyXG59XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQgeyBDYW5jZWxsYWJsZVByb21pc2UsIHR5cGUgQ2FuY2VsbGFibGVQcm9taXNlV2l0aFJlc29sdmVycyB9IGZyb20gXCIuL2NhbmNlbGxhYmxlLmpzXCI7XHJcbmltcG9ydCB7IG5ld1J1bnRpbWVDYWxsZXIsIG9iamVjdE5hbWVzIH0gZnJvbSBcIi4vcnVudGltZS5qc1wiO1xyXG5pbXBvcnQgeyBuYW5vaWQgfSBmcm9tIFwiLi9uYW5vaWQuanNcIjtcclxuXHJcbi8vIFNldHVwXHJcbndpbmRvdy5fd2FpbHMgPSB3aW5kb3cuX3dhaWxzIHx8IHt9O1xyXG5cclxudHlwZSBQcm9taXNlUmVzb2x2ZXJzID0gT21pdDxDYW5jZWxsYWJsZVByb21pc2VXaXRoUmVzb2x2ZXJzPGFueT4sIFwicHJvbWlzZVwiIHwgXCJvbmNhbmNlbGxlZFwiPlxyXG5cclxuY29uc3QgY2FsbCA9IG5ld1J1bnRpbWVDYWxsZXIob2JqZWN0TmFtZXMuQ2FsbCk7XHJcbmNvbnN0IGNhbmNlbENhbGwgPSBuZXdSdW50aW1lQ2FsbGVyKG9iamVjdE5hbWVzLkNhbmNlbENhbGwpO1xyXG5jb25zdCBjYWxsUmVzcG9uc2VzID0gbmV3IE1hcDxzdHJpbmcsIFByb21pc2VSZXNvbHZlcnM+KCk7XHJcblxyXG5jb25zdCBDYWxsQmluZGluZyA9IDA7XHJcbmNvbnN0IENhbmNlbE1ldGhvZCA9IDBcclxuXHJcbi8qKlxyXG4gKiBIb2xkcyBhbGwgcmVxdWlyZWQgaW5mb3JtYXRpb24gZm9yIGEgYmluZGluZyBjYWxsLlxyXG4gKiBNYXkgcHJvdmlkZSBlaXRoZXIgYSBtZXRob2QgSUQgb3IgYSBtZXRob2QgbmFtZSwgYnV0IG5vdCBib3RoLlxyXG4gKi9cclxuZXhwb3J0IHR5cGUgQ2FsbE9wdGlvbnMgPSB7XHJcbiAgICAvKiogVGhlIG51bWVyaWMgSUQgb2YgdGhlIGJvdW5kIG1ldGhvZCB0byBjYWxsLiAqL1xyXG4gICAgbWV0aG9kSUQ6IG51bWJlcjtcclxuICAgIC8qKiBUaGUgZnVsbHkgcXVhbGlmaWVkIG5hbWUgb2YgdGhlIGJvdW5kIG1ldGhvZCB0byBjYWxsLiAqL1xyXG4gICAgbWV0aG9kTmFtZT86IG5ldmVyO1xyXG4gICAgLyoqIEFyZ3VtZW50cyB0byBiZSBwYXNzZWQgaW50byB0aGUgYm91bmQgbWV0aG9kLiAqL1xyXG4gICAgYXJnczogYW55W107XHJcbn0gfCB7XHJcbiAgICAvKiogVGhlIG51bWVyaWMgSUQgb2YgdGhlIGJvdW5kIG1ldGhvZCB0byBjYWxsLiAqL1xyXG4gICAgbWV0aG9kSUQ/OiBuZXZlcjtcclxuICAgIC8qKiBUaGUgZnVsbHkgcXVhbGlmaWVkIG5hbWUgb2YgdGhlIGJvdW5kIG1ldGhvZCB0byBjYWxsLiAqL1xyXG4gICAgbWV0aG9kTmFtZTogc3RyaW5nO1xyXG4gICAgLyoqIEFyZ3VtZW50cyB0byBiZSBwYXNzZWQgaW50byB0aGUgYm91bmQgbWV0aG9kLiAqL1xyXG4gICAgYXJnczogYW55W107XHJcbn07XHJcblxyXG4vKipcclxuICogRXhjZXB0aW9uIGNsYXNzIHRoYXQgd2lsbCBiZSB0aHJvd24gaW4gY2FzZSB0aGUgYm91bmQgbWV0aG9kIHJldHVybnMgYW4gZXJyb3IuXHJcbiAqIFRoZSB2YWx1ZSBvZiB0aGUge0BsaW5rIFJ1bnRpbWVFcnJvciNuYW1lfSBwcm9wZXJ0eSBpcyBcIlJ1bnRpbWVFcnJvclwiLlxyXG4gKi9cclxuZXhwb3J0IGNsYXNzIFJ1bnRpbWVFcnJvciBleHRlbmRzIEVycm9yIHtcclxuICAgIC8qKlxyXG4gICAgICogQ29uc3RydWN0cyBhIG5ldyBSdW50aW1lRXJyb3IgaW5zdGFuY2UuXHJcbiAgICAgKiBAcGFyYW0gbWVzc2FnZSAtIFRoZSBlcnJvciBtZXNzYWdlLlxyXG4gICAgICogQHBhcmFtIG9wdGlvbnMgLSBPcHRpb25zIHRvIGJlIGZvcndhcmRlZCB0byB0aGUgRXJyb3IgY29uc3RydWN0b3IuXHJcbiAgICAgKi9cclxuICAgIGNvbnN0cnVjdG9yKG1lc3NhZ2U/OiBzdHJpbmcsIG9wdGlvbnM/OiBFcnJvck9wdGlvbnMpIHtcclxuICAgICAgICBzdXBlcihtZXNzYWdlLCBvcHRpb25zKTtcclxuICAgICAgICB0aGlzLm5hbWUgPSBcIlJ1bnRpbWVFcnJvclwiO1xyXG4gICAgfVxyXG59XHJcblxyXG4vKipcclxuICogR2VuZXJhdGVzIGEgdW5pcXVlIElEIHVzaW5nIHRoZSBuYW5vaWQgbGlicmFyeS5cclxuICpcclxuICogQHJldHVybnMgQSB1bmlxdWUgSUQgdGhhdCBkb2VzIG5vdCBleGlzdCBpbiB0aGUgY2FsbFJlc3BvbnNlcyBzZXQuXHJcbiAqL1xyXG5mdW5jdGlvbiBnZW5lcmF0ZUlEKCk6IHN0cmluZyB7XHJcbiAgICBsZXQgcmVzdWx0O1xyXG4gICAgZG8ge1xyXG4gICAgICAgIHJlc3VsdCA9IG5hbm9pZCgpO1xyXG4gICAgfSB3aGlsZSAoY2FsbFJlc3BvbnNlcy5oYXMocmVzdWx0KSk7XHJcbiAgICByZXR1cm4gcmVzdWx0O1xyXG59XHJcblxyXG4vKipcclxuICogQ2FsbCBhIGJvdW5kIG1ldGhvZCBhY2NvcmRpbmcgdG8gdGhlIGdpdmVuIGNhbGwgb3B0aW9ucy5cclxuICpcclxuICogSW4gY2FzZSBvZiBmYWlsdXJlLCB0aGUgcmV0dXJuZWQgcHJvbWlzZSB3aWxsIHJlamVjdCB3aXRoIGFuIGV4Y2VwdGlvblxyXG4gKiBhbW9uZyBSZWZlcmVuY2VFcnJvciAodW5rbm93biBtZXRob2QpLCBUeXBlRXJyb3IgKHdyb25nIGFyZ3VtZW50IGNvdW50IG9yIHR5cGUpLFxyXG4gKiB7QGxpbmsgUnVudGltZUVycm9yfSAobWV0aG9kIHJldHVybmVkIGFuIGVycm9yKSwgb3Igb3RoZXIgKG5ldHdvcmsgb3IgaW50ZXJuYWwgZXJyb3JzKS5cclxuICogVGhlIGV4Y2VwdGlvbiBtaWdodCBoYXZlIGEgXCJjYXVzZVwiIGZpZWxkIHdpdGggdGhlIHZhbHVlIHJldHVybmVkXHJcbiAqIGJ5IHRoZSBhcHBsaWNhdGlvbi0gb3Igc2VydmljZS1sZXZlbCBlcnJvciBtYXJzaGFsaW5nIGZ1bmN0aW9ucy5cclxuICpcclxuICogQHBhcmFtIG9wdGlvbnMgLSBBIG1ldGhvZCBjYWxsIGRlc2NyaXB0b3IuXHJcbiAqIEByZXR1cm5zIFRoZSByZXN1bHQgb2YgdGhlIGNhbGwuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gQ2FsbChvcHRpb25zOiBDYWxsT3B0aW9ucyk6IENhbmNlbGxhYmxlUHJvbWlzZTxhbnk+IHtcclxuICAgIGNvbnN0IGlkID0gZ2VuZXJhdGVJRCgpO1xyXG5cclxuICAgIGNvbnN0IHJlc3VsdCA9IENhbmNlbGxhYmxlUHJvbWlzZS53aXRoUmVzb2x2ZXJzPGFueT4oKTtcclxuICAgIGNhbGxSZXNwb25zZXMuc2V0KGlkLCB7IHJlc29sdmU6IHJlc3VsdC5yZXNvbHZlLCByZWplY3Q6IHJlc3VsdC5yZWplY3QgfSk7XHJcblxyXG4gICAgY29uc3QgcmVxdWVzdCA9IGNhbGwoQ2FsbEJpbmRpbmcsIE9iamVjdC5hc3NpZ24oeyBcImNhbGwtaWRcIjogaWQgfSwgb3B0aW9ucykpO1xyXG4gICAgbGV0IHJ1bm5pbmcgPSB0cnVlO1xyXG5cclxuICAgIHJlcXVlc3QudGhlbigocmVzKSA9PiB7XHJcbiAgICAgICAgcnVubmluZyA9IGZhbHNlO1xyXG4gICAgICAgIGNhbGxSZXNwb25zZXMuZGVsZXRlKGlkKTtcclxuICAgICAgICByZXN1bHQucmVzb2x2ZShyZXMpO1xyXG4gICAgfSwgKGVycikgPT4ge1xyXG4gICAgICAgIHJ1bm5pbmcgPSBmYWxzZTtcclxuICAgICAgICBjYWxsUmVzcG9uc2VzLmRlbGV0ZShpZCk7XHJcbiAgICAgICAgcmVzdWx0LnJlamVjdChlcnIpO1xyXG4gICAgfSk7XHJcblxyXG4gICAgY29uc3QgY2FuY2VsID0gKCkgPT4ge1xyXG4gICAgICAgIGNhbGxSZXNwb25zZXMuZGVsZXRlKGlkKTtcclxuICAgICAgICByZXR1cm4gY2FuY2VsQ2FsbChDYW5jZWxNZXRob2QsIHtcImNhbGwtaWRcIjogaWR9KS5jYXRjaCgoZXJyKSA9PiB7XHJcbiAgICAgICAgICAgIGNvbnNvbGUuZXJyb3IoXCJFcnJvciB3aGlsZSByZXF1ZXN0aW5nIGJpbmRpbmcgY2FsbCBjYW5jZWxsYXRpb246XCIsIGVycik7XHJcbiAgICAgICAgfSk7XHJcbiAgICB9O1xyXG5cclxuICAgIHJlc3VsdC5vbmNhbmNlbGxlZCA9ICgpID0+IHtcclxuICAgICAgICBpZiAocnVubmluZykge1xyXG4gICAgICAgICAgICByZXR1cm4gY2FuY2VsKCk7XHJcbiAgICAgICAgfSBlbHNlIHtcclxuICAgICAgICAgICAgcmV0dXJuIHJlcXVlc3QudGhlbihjYW5jZWwpO1xyXG4gICAgICAgIH1cclxuICAgIH07XHJcblxyXG4gICAgcmV0dXJuIHJlc3VsdC5wcm9taXNlO1xyXG59XHJcblxyXG4vKipcclxuICogQ2FsbHMgYSBib3VuZCBtZXRob2QgYnkgbmFtZSB3aXRoIHRoZSBzcGVjaWZpZWQgYXJndW1lbnRzLlxyXG4gKiBTZWUge0BsaW5rIENhbGx9IGZvciBkZXRhaWxzLlxyXG4gKlxyXG4gKiBAcGFyYW0gbWV0aG9kTmFtZSAtIFRoZSBuYW1lIG9mIHRoZSBtZXRob2QgaW4gdGhlIGZvcm1hdCAncGFja2FnZS5zdHJ1Y3QubWV0aG9kJy5cclxuICogQHBhcmFtIGFyZ3MgLSBUaGUgYXJndW1lbnRzIHRvIHBhc3MgdG8gdGhlIG1ldGhvZC5cclxuICogQHJldHVybnMgVGhlIHJlc3VsdCBvZiB0aGUgbWV0aG9kIGNhbGwuXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gQnlOYW1lKG1ldGhvZE5hbWU6IHN0cmluZywgLi4uYXJnczogYW55W10pOiBDYW5jZWxsYWJsZVByb21pc2U8YW55PiB7XHJcbiAgICByZXR1cm4gQ2FsbCh7IG1ldGhvZE5hbWUsIGFyZ3MgfSk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDYWxscyBhIG1ldGhvZCBieSBpdHMgbnVtZXJpYyBJRCB3aXRoIHRoZSBzcGVjaWZpZWQgYXJndW1lbnRzLlxyXG4gKiBTZWUge0BsaW5rIENhbGx9IGZvciBkZXRhaWxzLlxyXG4gKlxyXG4gKiBAcGFyYW0gbWV0aG9kSUQgLSBUaGUgSUQgb2YgdGhlIG1ldGhvZCB0byBjYWxsLlxyXG4gKiBAcGFyYW0gYXJncyAtIFRoZSBhcmd1bWVudHMgdG8gcGFzcyB0byB0aGUgbWV0aG9kLlxyXG4gKiBAcmV0dXJuIFRoZSByZXN1bHQgb2YgdGhlIG1ldGhvZCBjYWxsLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIEJ5SUQobWV0aG9kSUQ6IG51bWJlciwgLi4uYXJnczogYW55W10pOiBDYW5jZWxsYWJsZVByb21pc2U8YW55PiB7XHJcbiAgICByZXR1cm4gQ2FsbCh7IG1ldGhvZElELCBhcmdzIH0pO1xyXG59XHJcbiIsICIvLyBTb3VyY2U6IGh0dHBzOi8vZ2l0aHViLmNvbS9pbnNwZWN0LWpzL2lzLWNhbGxhYmxlXHJcblxyXG4vLyBUaGUgTUlUIExpY2Vuc2UgKE1JVClcclxuLy9cclxuLy8gQ29weXJpZ2h0IChjKSAyMDE1IEpvcmRhbiBIYXJiYW5kXHJcbi8vXHJcbi8vIFBlcm1pc3Npb24gaXMgaGVyZWJ5IGdyYW50ZWQsIGZyZWUgb2YgY2hhcmdlLCB0byBhbnkgcGVyc29uIG9idGFpbmluZyBhIGNvcHlcclxuLy8gb2YgdGhpcyBzb2Z0d2FyZSBhbmQgYXNzb2NpYXRlZCBkb2N1bWVudGF0aW9uIGZpbGVzICh0aGUgXCJTb2Z0d2FyZVwiKSwgdG8gZGVhbFxyXG4vLyBpbiB0aGUgU29mdHdhcmUgd2l0aG91dCByZXN0cmljdGlvbiwgaW5jbHVkaW5nIHdpdGhvdXQgbGltaXRhdGlvbiB0aGUgcmlnaHRzXHJcbi8vIHRvIHVzZSwgY29weSwgbW9kaWZ5LCBtZXJnZSwgcHVibGlzaCwgZGlzdHJpYnV0ZSwgc3VibGljZW5zZSwgYW5kL29yIHNlbGxcclxuLy8gY29waWVzIG9mIHRoZSBTb2Z0d2FyZSwgYW5kIHRvIHBlcm1pdCBwZXJzb25zIHRvIHdob20gdGhlIFNvZnR3YXJlIGlzXHJcbi8vIGZ1cm5pc2hlZCB0byBkbyBzbywgc3ViamVjdCB0byB0aGUgZm9sbG93aW5nIGNvbmRpdGlvbnM6XHJcbi8vXHJcbi8vIFRoZSBhYm92ZSBjb3B5cmlnaHQgbm90aWNlIGFuZCB0aGlzIHBlcm1pc3Npb24gbm90aWNlIHNoYWxsIGJlIGluY2x1ZGVkIGluIGFsbFxyXG4vLyBjb3BpZXMgb3Igc3Vic3RhbnRpYWwgcG9ydGlvbnMgb2YgdGhlIFNvZnR3YXJlLlxyXG4vL1xyXG4vLyBUSEUgU09GVFdBUkUgSVMgUFJPVklERUQgXCJBUyBJU1wiLCBXSVRIT1VUIFdBUlJBTlRZIE9GIEFOWSBLSU5ELCBFWFBSRVNTIE9SXHJcbi8vIElNUExJRUQsIElOQ0xVRElORyBCVVQgTk9UIExJTUlURUQgVE8gVEhFIFdBUlJBTlRJRVMgT0YgTUVSQ0hBTlRBQklMSVRZLFxyXG4vLyBGSVRORVNTIEZPUiBBIFBBUlRJQ1VMQVIgUFVSUE9TRSBBTkQgTk9OSU5GUklOR0VNRU5ULiBJTiBOTyBFVkVOVCBTSEFMTCBUSEVcclxuLy8gQVVUSE9SUyBPUiBDT1BZUklHSFQgSE9MREVSUyBCRSBMSUFCTEUgRk9SIEFOWSBDTEFJTSwgREFNQUdFUyBPUiBPVEhFUlxyXG4vLyBMSUFCSUxJVFksIFdIRVRIRVIgSU4gQU4gQUNUSU9OIE9GIENPTlRSQUNULCBUT1JUIE9SIE9USEVSV0lTRSwgQVJJU0lORyBGUk9NLFxyXG4vLyBPVVQgT0YgT1IgSU4gQ09OTkVDVElPTiBXSVRIIFRIRSBTT0ZUV0FSRSBPUiBUSEUgVVNFIE9SIE9USEVSIERFQUxJTkdTIElOIFRIRVxyXG4vLyBTT0ZUV0FSRS5cclxuXHJcbnZhciBmblRvU3RyID0gRnVuY3Rpb24ucHJvdG90eXBlLnRvU3RyaW5nO1xyXG52YXIgcmVmbGVjdEFwcGx5OiB0eXBlb2YgUmVmbGVjdC5hcHBseSB8IGZhbHNlIHwgbnVsbCA9IHR5cGVvZiBSZWZsZWN0ID09PSAnb2JqZWN0JyAmJiBSZWZsZWN0ICE9PSBudWxsICYmIFJlZmxlY3QuYXBwbHk7XHJcbnZhciBiYWRBcnJheUxpa2U6IGFueTtcclxudmFyIGlzQ2FsbGFibGVNYXJrZXI6IGFueTtcclxuaWYgKHR5cGVvZiByZWZsZWN0QXBwbHkgPT09ICdmdW5jdGlvbicgJiYgdHlwZW9mIE9iamVjdC5kZWZpbmVQcm9wZXJ0eSA9PT0gJ2Z1bmN0aW9uJykge1xyXG4gICAgdHJ5IHtcclxuICAgICAgICBiYWRBcnJheUxpa2UgPSBPYmplY3QuZGVmaW5lUHJvcGVydHkoe30sICdsZW5ndGgnLCB7XHJcbiAgICAgICAgICAgIGdldDogZnVuY3Rpb24gKCkge1xyXG4gICAgICAgICAgICAgICAgdGhyb3cgaXNDYWxsYWJsZU1hcmtlcjtcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgIH0pO1xyXG4gICAgICAgIGlzQ2FsbGFibGVNYXJrZXIgPSB7fTtcclxuICAgICAgICAvLyBlc2xpbnQtZGlzYWJsZS1uZXh0LWxpbmUgbm8tdGhyb3ctbGl0ZXJhbFxyXG4gICAgICAgIHJlZmxlY3RBcHBseShmdW5jdGlvbiAoKSB7IHRocm93IDQyOyB9LCBudWxsLCBiYWRBcnJheUxpa2UpO1xyXG4gICAgfSBjYXRjaCAoXykge1xyXG4gICAgICAgIGlmIChfICE9PSBpc0NhbGxhYmxlTWFya2VyKSB7XHJcbiAgICAgICAgICAgIHJlZmxlY3RBcHBseSA9IG51bGw7XHJcbiAgICAgICAgfVxyXG4gICAgfVxyXG59IGVsc2Uge1xyXG4gICAgcmVmbGVjdEFwcGx5ID0gbnVsbDtcclxufVxyXG5cclxudmFyIGNvbnN0cnVjdG9yUmVnZXggPSAvXlxccypjbGFzc1xcYi87XHJcbnZhciBpc0VTNkNsYXNzRm4gPSBmdW5jdGlvbiBpc0VTNkNsYXNzRnVuY3Rpb24odmFsdWU6IGFueSk6IGJvb2xlYW4ge1xyXG4gICAgdHJ5IHtcclxuICAgICAgICB2YXIgZm5TdHIgPSBmblRvU3RyLmNhbGwodmFsdWUpO1xyXG4gICAgICAgIHJldHVybiBjb25zdHJ1Y3RvclJlZ2V4LnRlc3QoZm5TdHIpO1xyXG4gICAgfSBjYXRjaCAoZSkge1xyXG4gICAgICAgIHJldHVybiBmYWxzZTsgLy8gbm90IGEgZnVuY3Rpb25cclxuICAgIH1cclxufTtcclxuXHJcbnZhciB0cnlGdW5jdGlvbk9iamVjdCA9IGZ1bmN0aW9uIHRyeUZ1bmN0aW9uVG9TdHIodmFsdWU6IGFueSk6IGJvb2xlYW4ge1xyXG4gICAgdHJ5IHtcclxuICAgICAgICBpZiAoaXNFUzZDbGFzc0ZuKHZhbHVlKSkgeyByZXR1cm4gZmFsc2U7IH1cclxuICAgICAgICBmblRvU3RyLmNhbGwodmFsdWUpO1xyXG4gICAgICAgIHJldHVybiB0cnVlO1xyXG4gICAgfSBjYXRjaCAoZSkge1xyXG4gICAgICAgIHJldHVybiBmYWxzZTtcclxuICAgIH1cclxufTtcclxudmFyIHRvU3RyID0gT2JqZWN0LnByb3RvdHlwZS50b1N0cmluZztcclxudmFyIG9iamVjdENsYXNzID0gJ1tvYmplY3QgT2JqZWN0XSc7XHJcbnZhciBmbkNsYXNzID0gJ1tvYmplY3QgRnVuY3Rpb25dJztcclxudmFyIGdlbkNsYXNzID0gJ1tvYmplY3QgR2VuZXJhdG9yRnVuY3Rpb25dJztcclxudmFyIGRkYUNsYXNzID0gJ1tvYmplY3QgSFRNTEFsbENvbGxlY3Rpb25dJzsgLy8gSUUgMTFcclxudmFyIGRkYUNsYXNzMiA9ICdbb2JqZWN0IEhUTUwgZG9jdW1lbnQuYWxsIGNsYXNzXSc7XHJcbnZhciBkZGFDbGFzczMgPSAnW29iamVjdCBIVE1MQ29sbGVjdGlvbl0nOyAvLyBJRSA5LTEwXHJcbnZhciBoYXNUb1N0cmluZ1RhZyA9IHR5cGVvZiBTeW1ib2wgPT09ICdmdW5jdGlvbicgJiYgISFTeW1ib2wudG9TdHJpbmdUYWc7IC8vIGJldHRlcjogdXNlIGBoYXMtdG9zdHJpbmd0YWdgXHJcblxyXG52YXIgaXNJRTY4ID0gISgwIGluIFssXSk7IC8vIGVzbGludC1kaXNhYmxlLWxpbmUgbm8tc3BhcnNlLWFycmF5cywgY29tbWEtc3BhY2luZ1xyXG5cclxudmFyIGlzRERBOiAodmFsdWU6IGFueSkgPT4gYm9vbGVhbiA9IGZ1bmN0aW9uIGlzRG9jdW1lbnREb3RBbGwoKSB7IHJldHVybiBmYWxzZTsgfTtcclxuaWYgKHR5cGVvZiBkb2N1bWVudCA9PT0gJ29iamVjdCcpIHtcclxuICAgIC8vIEZpcmVmb3ggMyBjYW5vbmljYWxpemVzIEREQSB0byB1bmRlZmluZWQgd2hlbiBpdCdzIG5vdCBhY2Nlc3NlZCBkaXJlY3RseVxyXG4gICAgdmFyIGFsbCA9IGRvY3VtZW50LmFsbDtcclxuICAgIGlmICh0b1N0ci5jYWxsKGFsbCkgPT09IHRvU3RyLmNhbGwoZG9jdW1lbnQuYWxsKSkge1xyXG4gICAgICAgIGlzRERBID0gZnVuY3Rpb24gaXNEb2N1bWVudERvdEFsbCh2YWx1ZSkge1xyXG4gICAgICAgICAgICAvKiBnbG9iYWxzIGRvY3VtZW50OiBmYWxzZSAqL1xyXG4gICAgICAgICAgICAvLyBpbiBJRSA2LTgsIHR5cGVvZiBkb2N1bWVudC5hbGwgaXMgXCJvYmplY3RcIiBhbmQgaXQncyB0cnV0aHlcclxuICAgICAgICAgICAgaWYgKChpc0lFNjggfHwgIXZhbHVlKSAmJiAodHlwZW9mIHZhbHVlID09PSAndW5kZWZpbmVkJyB8fCB0eXBlb2YgdmFsdWUgPT09ICdvYmplY3QnKSkge1xyXG4gICAgICAgICAgICAgICAgdHJ5IHtcclxuICAgICAgICAgICAgICAgICAgICB2YXIgc3RyID0gdG9TdHIuY2FsbCh2YWx1ZSk7XHJcbiAgICAgICAgICAgICAgICAgICAgcmV0dXJuIChcclxuICAgICAgICAgICAgICAgICAgICAgICAgc3RyID09PSBkZGFDbGFzc1xyXG4gICAgICAgICAgICAgICAgICAgICAgICB8fCBzdHIgPT09IGRkYUNsYXNzMlxyXG4gICAgICAgICAgICAgICAgICAgICAgICB8fCBzdHIgPT09IGRkYUNsYXNzMyAvLyBvcGVyYSAxMi4xNlxyXG4gICAgICAgICAgICAgICAgICAgICAgICB8fCBzdHIgPT09IG9iamVjdENsYXNzIC8vIElFIDYtOFxyXG4gICAgICAgICAgICAgICAgICAgICkgJiYgdmFsdWUoJycpID09IG51bGw7IC8vIGVzbGludC1kaXNhYmxlLWxpbmUgZXFlcWVxXHJcbiAgICAgICAgICAgICAgICB9IGNhdGNoIChlKSB7IC8qKi8gfVxyXG4gICAgICAgICAgICB9XHJcbiAgICAgICAgICAgIHJldHVybiBmYWxzZTtcclxuICAgICAgICB9O1xyXG4gICAgfVxyXG59XHJcblxyXG5mdW5jdGlvbiBpc0NhbGxhYmxlUmVmQXBwbHk8VD4odmFsdWU6IFQgfCB1bmtub3duKTogdmFsdWUgaXMgKC4uLmFyZ3M6IGFueVtdKSA9PiBhbnkgIHtcclxuICAgIGlmIChpc0REQSh2YWx1ZSkpIHsgcmV0dXJuIHRydWU7IH1cclxuICAgIGlmICghdmFsdWUpIHsgcmV0dXJuIGZhbHNlOyB9XHJcbiAgICBpZiAodHlwZW9mIHZhbHVlICE9PSAnZnVuY3Rpb24nICYmIHR5cGVvZiB2YWx1ZSAhPT0gJ29iamVjdCcpIHsgcmV0dXJuIGZhbHNlOyB9XHJcbiAgICB0cnkge1xyXG4gICAgICAgIChyZWZsZWN0QXBwbHkgYXMgYW55KSh2YWx1ZSwgbnVsbCwgYmFkQXJyYXlMaWtlKTtcclxuICAgIH0gY2F0Y2ggKGUpIHtcclxuICAgICAgICBpZiAoZSAhPT0gaXNDYWxsYWJsZU1hcmtlcikgeyByZXR1cm4gZmFsc2U7IH1cclxuICAgIH1cclxuICAgIHJldHVybiAhaXNFUzZDbGFzc0ZuKHZhbHVlKSAmJiB0cnlGdW5jdGlvbk9iamVjdCh2YWx1ZSk7XHJcbn1cclxuXHJcbmZ1bmN0aW9uIGlzQ2FsbGFibGVOb1JlZkFwcGx5PFQ+KHZhbHVlOiBUIHwgdW5rbm93bik6IHZhbHVlIGlzICguLi5hcmdzOiBhbnlbXSkgPT4gYW55IHtcclxuICAgIGlmIChpc0REQSh2YWx1ZSkpIHsgcmV0dXJuIHRydWU7IH1cclxuICAgIGlmICghdmFsdWUpIHsgcmV0dXJuIGZhbHNlOyB9XHJcbiAgICBpZiAodHlwZW9mIHZhbHVlICE9PSAnZnVuY3Rpb24nICYmIHR5cGVvZiB2YWx1ZSAhPT0gJ29iamVjdCcpIHsgcmV0dXJuIGZhbHNlOyB9XHJcbiAgICBpZiAoaGFzVG9TdHJpbmdUYWcpIHsgcmV0dXJuIHRyeUZ1bmN0aW9uT2JqZWN0KHZhbHVlKTsgfVxyXG4gICAgaWYgKGlzRVM2Q2xhc3NGbih2YWx1ZSkpIHsgcmV0dXJuIGZhbHNlOyB9XHJcbiAgICB2YXIgc3RyQ2xhc3MgPSB0b1N0ci5jYWxsKHZhbHVlKTtcclxuICAgIGlmIChzdHJDbGFzcyAhPT0gZm5DbGFzcyAmJiBzdHJDbGFzcyAhPT0gZ2VuQ2xhc3MgJiYgISgvXlxcW29iamVjdCBIVE1MLykudGVzdChzdHJDbGFzcykpIHsgcmV0dXJuIGZhbHNlOyB9XHJcbiAgICByZXR1cm4gdHJ5RnVuY3Rpb25PYmplY3QodmFsdWUpO1xyXG59O1xyXG5cclxuZXhwb3J0IGRlZmF1bHQgcmVmbGVjdEFwcGx5ID8gaXNDYWxsYWJsZVJlZkFwcGx5IDogaXNDYWxsYWJsZU5vUmVmQXBwbHk7XHJcbiIsICIvKlxyXG4gX1x0ICAgX19cdCAgXyBfX1xyXG58IHxcdCAvIC9fX18gXyhfKSAvX19fX1xyXG58IHwgL3wgLyAvIF9fIGAvIC8gLyBfX18vXHJcbnwgfC8gfC8gLyAvXy8gLyAvIChfXyAgKVxyXG58X18vfF9fL1xcX18sXy9fL18vX19fXy9cclxuVGhlIGVsZWN0cm9uIGFsdGVybmF0aXZlIGZvciBHb1xyXG4oYykgTGVhIEFudGhvbnkgMjAxOS1wcmVzZW50XHJcbiovXHJcblxyXG5pbXBvcnQgaXNDYWxsYWJsZSBmcm9tIFwiLi9jYWxsYWJsZS5qc1wiO1xyXG5cclxuLyoqXHJcbiAqIEV4Y2VwdGlvbiBjbGFzcyB0aGF0IHdpbGwgYmUgdXNlZCBhcyByZWplY3Rpb24gcmVhc29uXHJcbiAqIGluIGNhc2UgYSB7QGxpbmsgQ2FuY2VsbGFibGVQcm9taXNlfSBpcyBjYW5jZWxsZWQgc3VjY2Vzc2Z1bGx5LlxyXG4gKlxyXG4gKiBUaGUgdmFsdWUgb2YgdGhlIHtAbGluayBuYW1lfSBwcm9wZXJ0eSBpcyB0aGUgc3RyaW5nIGBcIkNhbmNlbEVycm9yXCJgLlxyXG4gKiBUaGUgdmFsdWUgb2YgdGhlIHtAbGluayBjYXVzZX0gcHJvcGVydHkgaXMgdGhlIGNhdXNlIHBhc3NlZCB0byB0aGUgY2FuY2VsIG1ldGhvZCwgaWYgYW55LlxyXG4gKi9cclxuZXhwb3J0IGNsYXNzIENhbmNlbEVycm9yIGV4dGVuZHMgRXJyb3Ige1xyXG4gICAgLyoqXHJcbiAgICAgKiBDb25zdHJ1Y3RzIGEgbmV3IGBDYW5jZWxFcnJvcmAgaW5zdGFuY2UuXHJcbiAgICAgKiBAcGFyYW0gbWVzc2FnZSAtIFRoZSBlcnJvciBtZXNzYWdlLlxyXG4gICAgICogQHBhcmFtIG9wdGlvbnMgLSBPcHRpb25zIHRvIGJlIGZvcndhcmRlZCB0byB0aGUgRXJyb3IgY29uc3RydWN0b3IuXHJcbiAgICAgKi9cclxuICAgIGNvbnN0cnVjdG9yKG1lc3NhZ2U/OiBzdHJpbmcsIG9wdGlvbnM/OiBFcnJvck9wdGlvbnMpIHtcclxuICAgICAgICBzdXBlcihtZXNzYWdlLCBvcHRpb25zKTtcclxuICAgICAgICB0aGlzLm5hbWUgPSBcIkNhbmNlbEVycm9yXCI7XHJcbiAgICB9XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBFeGNlcHRpb24gY2xhc3MgdGhhdCB3aWxsIGJlIHJlcG9ydGVkIGFzIGFuIHVuaGFuZGxlZCByZWplY3Rpb25cclxuICogaW4gY2FzZSBhIHtAbGluayBDYW5jZWxsYWJsZVByb21pc2V9IHJlamVjdHMgYWZ0ZXIgYmVpbmcgY2FuY2VsbGVkLFxyXG4gKiBvciB3aGVuIHRoZSBgb25jYW5jZWxsZWRgIGNhbGxiYWNrIHRocm93cyBvciByZWplY3RzLlxyXG4gKlxyXG4gKiBUaGUgdmFsdWUgb2YgdGhlIHtAbGluayBuYW1lfSBwcm9wZXJ0eSBpcyB0aGUgc3RyaW5nIGBcIkNhbmNlbGxlZFJlamVjdGlvbkVycm9yXCJgLlxyXG4gKiBUaGUgdmFsdWUgb2YgdGhlIHtAbGluayBjYXVzZX0gcHJvcGVydHkgaXMgdGhlIHJlYXNvbiB0aGUgcHJvbWlzZSByZWplY3RlZCB3aXRoLlxyXG4gKlxyXG4gKiBCZWNhdXNlIHRoZSBvcmlnaW5hbCBwcm9taXNlIHdhcyBjYW5jZWxsZWQsXHJcbiAqIGEgd3JhcHBlciBwcm9taXNlIHdpbGwgYmUgcGFzc2VkIHRvIHRoZSB1bmhhbmRsZWQgcmVqZWN0aW9uIGxpc3RlbmVyIGluc3RlYWQuXHJcbiAqIFRoZSB7QGxpbmsgcHJvbWlzZX0gcHJvcGVydHkgaG9sZHMgYSByZWZlcmVuY2UgdG8gdGhlIG9yaWdpbmFsIHByb21pc2UuXHJcbiAqL1xyXG5leHBvcnQgY2xhc3MgQ2FuY2VsbGVkUmVqZWN0aW9uRXJyb3IgZXh0ZW5kcyBFcnJvciB7XHJcbiAgICAvKipcclxuICAgICAqIEhvbGRzIGEgcmVmZXJlbmNlIHRvIHRoZSBwcm9taXNlIHRoYXQgd2FzIGNhbmNlbGxlZCBhbmQgdGhlbiByZWplY3RlZC5cclxuICAgICAqL1xyXG4gICAgcHJvbWlzZTogQ2FuY2VsbGFibGVQcm9taXNlPHVua25vd24+O1xyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ29uc3RydWN0cyBhIG5ldyBgQ2FuY2VsbGVkUmVqZWN0aW9uRXJyb3JgIGluc3RhbmNlLlxyXG4gICAgICogQHBhcmFtIHByb21pc2UgLSBUaGUgcHJvbWlzZSB0aGF0IGNhdXNlZCB0aGUgZXJyb3Igb3JpZ2luYWxseS5cclxuICAgICAqIEBwYXJhbSByZWFzb24gLSBUaGUgcmVqZWN0aW9uIHJlYXNvbi5cclxuICAgICAqIEBwYXJhbSBpbmZvIC0gQW4gb3B0aW9uYWwgaW5mb3JtYXRpdmUgbWVzc2FnZSBzcGVjaWZ5aW5nIHRoZSBjaXJjdW1zdGFuY2VzIGluIHdoaWNoIHRoZSBlcnJvciB3YXMgdGhyb3duLlxyXG4gICAgICogICAgICAgICAgICAgICBEZWZhdWx0cyB0byB0aGUgc3RyaW5nIGBcIlVuaGFuZGxlZCByZWplY3Rpb24gaW4gY2FuY2VsbGVkIHByb21pc2UuXCJgLlxyXG4gICAgICovXHJcbiAgICBjb25zdHJ1Y3Rvcihwcm9taXNlOiBDYW5jZWxsYWJsZVByb21pc2U8dW5rbm93bj4sIHJlYXNvbj86IGFueSwgaW5mbz86IHN0cmluZykge1xyXG4gICAgICAgIHN1cGVyKChpbmZvID8/IFwiVW5oYW5kbGVkIHJlamVjdGlvbiBpbiBjYW5jZWxsZWQgcHJvbWlzZS5cIikgKyBcIiBSZWFzb246IFwiICsgZXJyb3JNZXNzYWdlKHJlYXNvbiksIHsgY2F1c2U6IHJlYXNvbiB9KTtcclxuICAgICAgICB0aGlzLnByb21pc2UgPSBwcm9taXNlO1xyXG4gICAgICAgIHRoaXMubmFtZSA9IFwiQ2FuY2VsbGVkUmVqZWN0aW9uRXJyb3JcIjtcclxuICAgIH1cclxufVxyXG5cclxudHlwZSBDYW5jZWxsYWJsZVByb21pc2VSZXNvbHZlcjxUPiA9ICh2YWx1ZTogVCB8IFByb21pc2VMaWtlPFQ+IHwgQ2FuY2VsbGFibGVQcm9taXNlTGlrZTxUPikgPT4gdm9pZDtcclxudHlwZSBDYW5jZWxsYWJsZVByb21pc2VSZWplY3RvciA9IChyZWFzb24/OiBhbnkpID0+IHZvaWQ7XHJcbnR5cGUgQ2FuY2VsbGFibGVQcm9taXNlQ2FuY2VsbGVyID0gKGNhdXNlPzogYW55KSA9PiB2b2lkIHwgUHJvbWlzZUxpa2U8dm9pZD47XHJcbnR5cGUgQ2FuY2VsbGFibGVQcm9taXNlRXhlY3V0b3I8VD4gPSAocmVzb2x2ZTogQ2FuY2VsbGFibGVQcm9taXNlUmVzb2x2ZXI8VD4sIHJlamVjdDogQ2FuY2VsbGFibGVQcm9taXNlUmVqZWN0b3IpID0+IHZvaWQ7XHJcblxyXG5leHBvcnQgaW50ZXJmYWNlIENhbmNlbGxhYmxlUHJvbWlzZUxpa2U8VD4ge1xyXG4gICAgdGhlbjxUUmVzdWx0MSA9IFQsIFRSZXN1bHQyID0gbmV2ZXI+KG9uZnVsZmlsbGVkPzogKCh2YWx1ZTogVCkgPT4gVFJlc3VsdDEgfCBQcm9taXNlTGlrZTxUUmVzdWx0MT4gfCBDYW5jZWxsYWJsZVByb21pc2VMaWtlPFRSZXN1bHQxPikgfCB1bmRlZmluZWQgfCBudWxsLCBvbnJlamVjdGVkPzogKChyZWFzb246IGFueSkgPT4gVFJlc3VsdDIgfCBQcm9taXNlTGlrZTxUUmVzdWx0Mj4gfCBDYW5jZWxsYWJsZVByb21pc2VMaWtlPFRSZXN1bHQyPikgfCB1bmRlZmluZWQgfCBudWxsKTogQ2FuY2VsbGFibGVQcm9taXNlTGlrZTxUUmVzdWx0MSB8IFRSZXN1bHQyPjtcclxuICAgIGNhbmNlbChjYXVzZT86IGFueSk6IHZvaWQgfCBQcm9taXNlTGlrZTx2b2lkPjtcclxufVxyXG5cclxuLyoqXHJcbiAqIFdyYXBzIGEgY2FuY2VsbGFibGUgcHJvbWlzZSBhbG9uZyB3aXRoIGl0cyByZXNvbHV0aW9uIG1ldGhvZHMuXHJcbiAqIFRoZSBgb25jYW5jZWxsZWRgIGZpZWxkIHdpbGwgYmUgbnVsbCBpbml0aWFsbHkgYnV0IG1heSBiZSBzZXQgdG8gcHJvdmlkZSBhIGN1c3RvbSBjYW5jZWxsYXRpb24gZnVuY3Rpb24uXHJcbiAqL1xyXG5leHBvcnQgaW50ZXJmYWNlIENhbmNlbGxhYmxlUHJvbWlzZVdpdGhSZXNvbHZlcnM8VD4ge1xyXG4gICAgcHJvbWlzZTogQ2FuY2VsbGFibGVQcm9taXNlPFQ+O1xyXG4gICAgcmVzb2x2ZTogQ2FuY2VsbGFibGVQcm9taXNlUmVzb2x2ZXI8VD47XHJcbiAgICByZWplY3Q6IENhbmNlbGxhYmxlUHJvbWlzZVJlamVjdG9yO1xyXG4gICAgb25jYW5jZWxsZWQ6IENhbmNlbGxhYmxlUHJvbWlzZUNhbmNlbGxlciB8IG51bGw7XHJcbn1cclxuXHJcbmludGVyZmFjZSBDYW5jZWxsYWJsZVByb21pc2VTdGF0ZSB7XHJcbiAgICByZWFkb25seSByb290OiBDYW5jZWxsYWJsZVByb21pc2VTdGF0ZTtcclxuICAgIHJlc29sdmluZzogYm9vbGVhbjtcclxuICAgIHNldHRsZWQ6IGJvb2xlYW47XHJcbiAgICByZWFzb24/OiBDYW5jZWxFcnJvcjtcclxufVxyXG5cclxuLy8gUHJpdmF0ZSBmaWVsZCBuYW1lcy5cclxuY29uc3QgYmFycmllclN5bSA9IFN5bWJvbChcImJhcnJpZXJcIik7XHJcbmNvbnN0IGNhbmNlbEltcGxTeW0gPSBTeW1ib2woXCJjYW5jZWxJbXBsXCIpO1xyXG5jb25zdCBzcGVjaWVzOiB0eXBlb2YgU3ltYm9sLnNwZWNpZXMgPSBTeW1ib2wuc3BlY2llcyA/PyBTeW1ib2woXCJzcGVjaWVzUG9seWZpbGxcIik7XHJcblxyXG4vKipcclxuICogQSBwcm9taXNlIHdpdGggYW4gYXR0YWNoZWQgbWV0aG9kIGZvciBjYW5jZWxsaW5nIGxvbmctcnVubmluZyBvcGVyYXRpb25zIChzZWUge0BsaW5rIENhbmNlbGxhYmxlUHJvbWlzZSNjYW5jZWx9KS5cclxuICogQ2FuY2VsbGF0aW9uIGNhbiBvcHRpb25hbGx5IGJlIGJvdW5kIHRvIGFuIHtAbGluayBBYm9ydFNpZ25hbH1cclxuICogZm9yIGJldHRlciBjb21wb3NhYmlsaXR5IChzZWUge0BsaW5rIENhbmNlbGxhYmxlUHJvbWlzZSNjYW5jZWxPbn0pLlxyXG4gKlxyXG4gKiBDYW5jZWxsaW5nIGEgcGVuZGluZyBwcm9taXNlIHdpbGwgcmVzdWx0IGluIGFuIGltbWVkaWF0ZSByZWplY3Rpb25cclxuICogd2l0aCBhbiBpbnN0YW5jZSBvZiB7QGxpbmsgQ2FuY2VsRXJyb3J9IGFzIHJlYXNvbixcclxuICogYnV0IHdob2V2ZXIgc3RhcnRlZCB0aGUgcHJvbWlzZSB3aWxsIGJlIHJlc3BvbnNpYmxlXHJcbiAqIGZvciBhY3R1YWxseSBhYm9ydGluZyB0aGUgdW5kZXJseWluZyBvcGVyYXRpb24uXHJcbiAqIFRvIHRoaXMgcHVycG9zZSwgdGhlIGNvbnN0cnVjdG9yIGFuZCBhbGwgY2hhaW5pbmcgbWV0aG9kc1xyXG4gKiBhY2NlcHQgb3B0aW9uYWwgY2FuY2VsbGF0aW9uIGNhbGxiYWNrcy5cclxuICpcclxuICogSWYgYSBgQ2FuY2VsbGFibGVQcm9taXNlYCBzdGlsbCByZXNvbHZlcyBhZnRlciBoYXZpbmcgYmVlbiBjYW5jZWxsZWQsXHJcbiAqIHRoZSByZXN1bHQgd2lsbCBiZSBkaXNjYXJkZWQuIElmIGl0IHJlamVjdHMsIHRoZSByZWFzb25cclxuICogd2lsbCBiZSByZXBvcnRlZCBhcyBhbiB1bmhhbmRsZWQgcmVqZWN0aW9uLFxyXG4gKiB3cmFwcGVkIGluIGEge0BsaW5rIENhbmNlbGxlZFJlamVjdGlvbkVycm9yfSBpbnN0YW5jZS5cclxuICogVG8gZmFjaWxpdGF0ZSB0aGUgaGFuZGxpbmcgb2YgY2FuY2VsbGF0aW9uIHJlcXVlc3RzLFxyXG4gKiBjYW5jZWxsZWQgYENhbmNlbGxhYmxlUHJvbWlzZWBzIHdpbGwgX25vdF8gcmVwb3J0IHVuaGFuZGxlZCBgQ2FuY2VsRXJyb3Jgc1xyXG4gKiB3aG9zZSBgY2F1c2VgIGZpZWxkIGlzIHRoZSBzYW1lIGFzIHRoZSBvbmUgd2l0aCB3aGljaCB0aGUgY3VycmVudCBwcm9taXNlIHdhcyBjYW5jZWxsZWQuXHJcbiAqXHJcbiAqIEFsbCB1c3VhbCBwcm9taXNlIG1ldGhvZHMgYXJlIGRlZmluZWQgYW5kIHJldHVybiBhIGBDYW5jZWxsYWJsZVByb21pc2VgXHJcbiAqIHdob3NlIGNhbmNlbCBtZXRob2Qgd2lsbCBjYW5jZWwgdGhlIHBhcmVudCBvcGVyYXRpb24gYXMgd2VsbCwgcHJvcGFnYXRpbmcgdGhlIGNhbmNlbGxhdGlvbiByZWFzb25cclxuICogdXB3YXJkcyB0aHJvdWdoIHByb21pc2UgY2hhaW5zLlxyXG4gKiBDb252ZXJzZWx5LCBjYW5jZWxsaW5nIGEgcHJvbWlzZSB3aWxsIG5vdCBhdXRvbWF0aWNhbGx5IGNhbmNlbCBkZXBlbmRlbnQgcHJvbWlzZXMgZG93bnN0cmVhbTpcclxuICogYGBgdHNcclxuICogbGV0IHJvb3QgPSBuZXcgQ2FuY2VsbGFibGVQcm9taXNlKChyZXNvbHZlLCByZWplY3QpID0+IHsgLi4uIH0pO1xyXG4gKiBsZXQgY2hpbGQxID0gcm9vdC50aGVuKCgpID0+IHsgLi4uIH0pO1xyXG4gKiBsZXQgY2hpbGQyID0gY2hpbGQxLnRoZW4oKCkgPT4geyAuLi4gfSk7XHJcbiAqIGxldCBjaGlsZDMgPSByb290LmNhdGNoKCgpID0+IHsgLi4uIH0pO1xyXG4gKiBjaGlsZDEuY2FuY2VsKCk7IC8vIENhbmNlbHMgY2hpbGQxIGFuZCByb290LCBidXQgbm90IGNoaWxkMiBvciBjaGlsZDNcclxuICogYGBgXHJcbiAqIENhbmNlbGxpbmcgYSBwcm9taXNlIHRoYXQgaGFzIGFscmVhZHkgc2V0dGxlZCBpcyBzYWZlIGFuZCBoYXMgbm8gY29uc2VxdWVuY2UuXHJcbiAqXHJcbiAqIFRoZSBgY2FuY2VsYCBtZXRob2QgcmV0dXJucyBhIHByb21pc2UgdGhhdCBfYWx3YXlzIGZ1bGZpbGxzX1xyXG4gKiBhZnRlciB0aGUgd2hvbGUgY2hhaW4gaGFzIHByb2Nlc3NlZCB0aGUgY2FuY2VsIHJlcXVlc3RcclxuICogYW5kIGFsbCBhdHRhY2hlZCBjYWxsYmFja3MgdXAgdG8gdGhhdCBtb21lbnQgaGF2ZSBydW4uXHJcbiAqXHJcbiAqIEFsbCBFUzIwMjQgcHJvbWlzZSBtZXRob2RzIChzdGF0aWMgYW5kIGluc3RhbmNlKSBhcmUgZGVmaW5lZCBvbiBDYW5jZWxsYWJsZVByb21pc2UsXHJcbiAqIGJ1dCBhY3R1YWwgYXZhaWxhYmlsaXR5IG1heSB2YXJ5IHdpdGggT1Mvd2VidmlldyB2ZXJzaW9uLlxyXG4gKlxyXG4gKiBJbiBsaW5lIHdpdGggdGhlIHByb3Bvc2FsIGF0IGh0dHBzOi8vZ2l0aHViLmNvbS90YzM5L3Byb3Bvc2FsLXJtLWJ1aWx0aW4tc3ViY2xhc3NpbmcsXHJcbiAqIGBDYW5jZWxsYWJsZVByb21pc2VgIGRvZXMgbm90IHN1cHBvcnQgdHJhbnNwYXJlbnQgc3ViY2xhc3NpbmcuXHJcbiAqIEV4dGVuZGVycyBzaG91bGQgdGFrZSBjYXJlIHRvIHByb3ZpZGUgdGhlaXIgb3duIG1ldGhvZCBpbXBsZW1lbnRhdGlvbnMuXHJcbiAqIFRoaXMgbWlnaHQgYmUgcmVjb25zaWRlcmVkIGluIGNhc2UgdGhlIHByb3Bvc2FsIGlzIHJldGlyZWQuXHJcbiAqXHJcbiAqIENhbmNlbGxhYmxlUHJvbWlzZSBpcyBhIHdyYXBwZXIgYXJvdW5kIHRoZSBET00gUHJvbWlzZSBvYmplY3RcclxuICogYW5kIGlzIGNvbXBsaWFudCB3aXRoIHRoZSBbUHJvbWlzZXMvQSsgc3BlY2lmaWNhdGlvbl0oaHR0cHM6Ly9wcm9taXNlc2FwbHVzLmNvbS8pXHJcbiAqIChpdCBwYXNzZXMgdGhlIFtjb21wbGlhbmNlIHN1aXRlXShodHRwczovL2dpdGh1Yi5jb20vcHJvbWlzZXMtYXBsdXMvcHJvbWlzZXMtdGVzdHMpKVxyXG4gKiBpZiBzbyBpcyB0aGUgdW5kZXJseWluZyBpbXBsZW1lbnRhdGlvbi5cclxuICovXHJcbmV4cG9ydCBjbGFzcyBDYW5jZWxsYWJsZVByb21pc2U8VD4gZXh0ZW5kcyBQcm9taXNlPFQ+IGltcGxlbWVudHMgUHJvbWlzZUxpa2U8VD4sIENhbmNlbGxhYmxlUHJvbWlzZUxpa2U8VD4ge1xyXG4gICAgLy8gUHJpdmF0ZSBmaWVsZHMuXHJcbiAgICAvKiogQGludGVybmFsICovXHJcbiAgICBwcml2YXRlIFtiYXJyaWVyU3ltXSE6IFBhcnRpYWw8UHJvbWlzZVdpdGhSZXNvbHZlcnM8dm9pZD4+IHwgbnVsbDtcclxuICAgIC8qKiBAaW50ZXJuYWwgKi9cclxuICAgIHByaXZhdGUgcmVhZG9ubHkgW2NhbmNlbEltcGxTeW1dITogKHJlYXNvbjogQ2FuY2VsRXJyb3IpID0+IHZvaWQgfCBQcm9taXNlTGlrZTx2b2lkPjtcclxuXHJcbiAgICAvKipcclxuICAgICAqIENyZWF0ZXMgYSBuZXcgYENhbmNlbGxhYmxlUHJvbWlzZWAuXHJcbiAgICAgKlxyXG4gICAgICogQHBhcmFtIGV4ZWN1dG9yIC0gQSBjYWxsYmFjayB1c2VkIHRvIGluaXRpYWxpemUgdGhlIHByb21pc2UuIFRoaXMgY2FsbGJhY2sgaXMgcGFzc2VkIHR3byBhcmd1bWVudHM6XHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICBhIGByZXNvbHZlYCBjYWxsYmFjayB1c2VkIHRvIHJlc29sdmUgdGhlIHByb21pc2Ugd2l0aCBhIHZhbHVlXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICBvciB0aGUgcmVzdWx0IG9mIGFub3RoZXIgcHJvbWlzZSAocG9zc2libHkgY2FuY2VsbGFibGUpLFxyXG4gICAgICogICAgICAgICAgICAgICAgICAgYW5kIGEgYHJlamVjdGAgY2FsbGJhY2sgdXNlZCB0byByZWplY3QgdGhlIHByb21pc2Ugd2l0aCBhIHByb3ZpZGVkIHJlYXNvbiBvciBlcnJvci5cclxuICAgICAqICAgICAgICAgICAgICAgICAgIElmIHRoZSB2YWx1ZSBwcm92aWRlZCB0byB0aGUgYHJlc29sdmVgIGNhbGxiYWNrIGlzIGEgdGhlbmFibGUgX2FuZF8gY2FuY2VsbGFibGUgb2JqZWN0XHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICAoaXQgaGFzIGEgYHRoZW5gIF9hbmRfIGEgYGNhbmNlbGAgbWV0aG9kKSxcclxuICAgICAqICAgICAgICAgICAgICAgICAgIGNhbmNlbGxhdGlvbiByZXF1ZXN0cyB3aWxsIGJlIGZvcndhcmRlZCB0byB0aGF0IG9iamVjdCBhbmQgdGhlIG9uY2FuY2VsbGVkIHdpbGwgbm90IGJlIGludm9rZWQgYW55bW9yZS5cclxuICAgICAqICAgICAgICAgICAgICAgICAgIElmIGFueSBvbmUgb2YgdGhlIHR3byBjYWxsYmFja3MgaXMgY2FsbGVkIF9hZnRlcl8gdGhlIHByb21pc2UgaGFzIGJlZW4gY2FuY2VsbGVkLFxyXG4gICAgICogICAgICAgICAgICAgICAgICAgdGhlIHByb3ZpZGVkIHZhbHVlcyB3aWxsIGJlIGNhbmNlbGxlZCBhbmQgcmVzb2x2ZWQgYXMgdXN1YWwsXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICBidXQgdGhlaXIgcmVzdWx0cyB3aWxsIGJlIGRpc2NhcmRlZC5cclxuICAgICAqICAgICAgICAgICAgICAgICAgIEhvd2V2ZXIsIGlmIHRoZSByZXNvbHV0aW9uIHByb2Nlc3MgdWx0aW1hdGVseSBlbmRzIHVwIGluIGEgcmVqZWN0aW9uXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICB0aGF0IGlzIG5vdCBkdWUgdG8gY2FuY2VsbGF0aW9uLCB0aGUgcmVqZWN0aW9uIHJlYXNvblxyXG4gICAgICogICAgICAgICAgICAgICAgICAgd2lsbCBiZSB3cmFwcGVkIGluIGEge0BsaW5rIENhbmNlbGxlZFJlamVjdGlvbkVycm9yfVxyXG4gICAgICogICAgICAgICAgICAgICAgICAgYW5kIGJ1YmJsZWQgdXAgYXMgYW4gdW5oYW5kbGVkIHJlamVjdGlvbi5cclxuICAgICAqIEBwYXJhbSBvbmNhbmNlbGxlZCAtIEl0IGlzIHRoZSBjYWxsZXIncyByZXNwb25zaWJpbGl0eSB0byBlbnN1cmUgdGhhdCBhbnkgb3BlcmF0aW9uXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICAgICBzdGFydGVkIGJ5IHRoZSBleGVjdXRvciBpcyBwcm9wZXJseSBoYWx0ZWQgdXBvbiBjYW5jZWxsYXRpb24uXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICAgICBUaGlzIG9wdGlvbmFsIGNhbGxiYWNrIGNhbiBiZSB1c2VkIHRvIHRoYXQgcHVycG9zZS5cclxuICAgICAqICAgICAgICAgICAgICAgICAgICAgIEl0IHdpbGwgYmUgY2FsbGVkIF9zeW5jaHJvbm91c2x5XyB3aXRoIGEgY2FuY2VsbGF0aW9uIGNhdXNlXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICAgICB3aGVuIGNhbmNlbGxhdGlvbiBpcyByZXF1ZXN0ZWQsIF9hZnRlcl8gdGhlIHByb21pc2UgaGFzIGFscmVhZHkgcmVqZWN0ZWRcclxuICAgICAqICAgICAgICAgICAgICAgICAgICAgIHdpdGggYSB7QGxpbmsgQ2FuY2VsRXJyb3J9LCBidXQgX2JlZm9yZV9cclxuICAgICAqICAgICAgICAgICAgICAgICAgICAgIGFueSB7QGxpbmsgdGhlbn0ve0BsaW5rIGNhdGNofS97QGxpbmsgZmluYWxseX0gY2FsbGJhY2sgcnVucy5cclxuICAgICAqICAgICAgICAgICAgICAgICAgICAgIElmIHRoZSBjYWxsYmFjayByZXR1cm5zIGEgdGhlbmFibGUsIHRoZSBwcm9taXNlIHJldHVybmVkIGZyb20ge0BsaW5rIGNhbmNlbH1cclxuICAgICAqICAgICAgICAgICAgICAgICAgICAgIHdpbGwgb25seSBmdWxmaWxsIGFmdGVyIHRoZSBmb3JtZXIgaGFzIHNldHRsZWQuXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICAgICBVbmhhbmRsZWQgZXhjZXB0aW9ucyBvciByZWplY3Rpb25zIGZyb20gdGhlIGNhbGxiYWNrIHdpbGwgYmUgd3JhcHBlZFxyXG4gICAgICogICAgICAgICAgICAgICAgICAgICAgaW4gYSB7QGxpbmsgQ2FuY2VsbGVkUmVqZWN0aW9uRXJyb3J9IGFuZCBidWJibGVkIHVwIGFzIHVuaGFuZGxlZCByZWplY3Rpb25zLlxyXG4gICAgICogICAgICAgICAgICAgICAgICAgICAgSWYgdGhlIGByZXNvbHZlYCBjYWxsYmFjayBpcyBjYWxsZWQgYmVmb3JlIGNhbmNlbGxhdGlvbiB3aXRoIGEgY2FuY2VsbGFibGUgcHJvbWlzZSxcclxuICAgICAqICAgICAgICAgICAgICAgICAgICAgIGNhbmNlbGxhdGlvbiByZXF1ZXN0cyBvbiB0aGlzIHByb21pc2Ugd2lsbCBiZSBkaXZlcnRlZCB0byB0aGF0IHByb21pc2UsXHJcbiAgICAgKiAgICAgICAgICAgICAgICAgICAgICBhbmQgdGhlIG9yaWdpbmFsIGBvbmNhbmNlbGxlZGAgY2FsbGJhY2sgd2lsbCBiZSBkaXNjYXJkZWQuXHJcbiAgICAgKi9cclxuICAgIGNvbnN0cnVjdG9yKGV4ZWN1dG9yOiBDYW5jZWxsYWJsZVByb21pc2VFeGVjdXRvcjxUPiwgb25jYW5jZWxsZWQ/OiBDYW5jZWxsYWJsZVByb21pc2VDYW5jZWxsZXIpIHtcclxuICAgICAgICBsZXQgcmVzb2x2ZSE6ICh2YWx1ZTogVCB8IFByb21pc2VMaWtlPFQ+KSA9PiB2b2lkO1xyXG4gICAgICAgIGxldCByZWplY3QhOiAocmVhc29uPzogYW55KSA9PiB2b2lkO1xyXG4gICAgICAgIHN1cGVyKChyZXMsIHJlaikgPT4geyByZXNvbHZlID0gcmVzOyByZWplY3QgPSByZWo7IH0pO1xyXG5cclxuICAgICAgICBpZiAoKHRoaXMuY29uc3RydWN0b3IgYXMgYW55KVtzcGVjaWVzXSAhPT0gUHJvbWlzZSkge1xyXG4gICAgICAgICAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiQ2FuY2VsbGFibGVQcm9taXNlIGRvZXMgbm90IHN1cHBvcnQgdHJhbnNwYXJlbnQgc3ViY2xhc3NpbmcuIFBsZWFzZSByZWZyYWluIGZyb20gb3ZlcnJpZGluZyB0aGUgW1N5bWJvbC5zcGVjaWVzXSBzdGF0aWMgcHJvcGVydHkuXCIpO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgbGV0IHByb21pc2U6IENhbmNlbGxhYmxlUHJvbWlzZVdpdGhSZXNvbHZlcnM8VD4gPSB7XHJcbiAgICAgICAgICAgIHByb21pc2U6IHRoaXMsXHJcbiAgICAgICAgICAgIHJlc29sdmUsXHJcbiAgICAgICAgICAgIHJlamVjdCxcclxuICAgICAgICAgICAgZ2V0IG9uY2FuY2VsbGVkKCkgeyByZXR1cm4gb25jYW5jZWxsZWQgPz8gbnVsbDsgfSxcclxuICAgICAgICAgICAgc2V0IG9uY2FuY2VsbGVkKGNiKSB7IG9uY2FuY2VsbGVkID0gY2IgPz8gdW5kZWZpbmVkOyB9XHJcbiAgICAgICAgfTtcclxuXHJcbiAgICAgICAgY29uc3Qgc3RhdGU6IENhbmNlbGxhYmxlUHJvbWlzZVN0YXRlID0ge1xyXG4gICAgICAgICAgICBnZXQgcm9vdCgpIHsgcmV0dXJuIHN0YXRlOyB9LFxyXG4gICAgICAgICAgICByZXNvbHZpbmc6IGZhbHNlLFxyXG4gICAgICAgICAgICBzZXR0bGVkOiBmYWxzZVxyXG4gICAgICAgIH07XHJcblxyXG4gICAgICAgIC8vIFNldHVwIGNhbmNlbGxhdGlvbiBzeXN0ZW0uXHJcbiAgICAgICAgdm9pZCBPYmplY3QuZGVmaW5lUHJvcGVydGllcyh0aGlzLCB7XHJcbiAgICAgICAgICAgIFtiYXJyaWVyU3ltXToge1xyXG4gICAgICAgICAgICAgICAgY29uZmlndXJhYmxlOiBmYWxzZSxcclxuICAgICAgICAgICAgICAgIGVudW1lcmFibGU6IGZhbHNlLFxyXG4gICAgICAgICAgICAgICAgd3JpdGFibGU6IHRydWUsXHJcbiAgICAgICAgICAgICAgICB2YWx1ZTogbnVsbFxyXG4gICAgICAgICAgICB9LFxyXG4gICAgICAgICAgICBbY2FuY2VsSW1wbFN5bV06IHtcclxuICAgICAgICAgICAgICAgIGNvbmZpZ3VyYWJsZTogZmFsc2UsXHJcbiAgICAgICAgICAgICAgICBlbnVtZXJhYmxlOiBmYWxzZSxcclxuICAgICAgICAgICAgICAgIHdyaXRhYmxlOiBmYWxzZSxcclxuICAgICAgICAgICAgICAgIHZhbHVlOiBjYW5jZWxsZXJGb3IocHJvbWlzZSwgc3RhdGUpXHJcbiAgICAgICAgICAgIH1cclxuICAgICAgICB9KTtcclxuXHJcbiAgICAgICAgLy8gUnVuIHRoZSBhY3R1YWwgZXhlY3V0b3IuXHJcbiAgICAgICAgY29uc3QgcmVqZWN0b3IgPSByZWplY3RvckZvcihwcm9taXNlLCBzdGF0ZSk7XHJcbiAgICAgICAgdHJ5IHtcclxuICAgICAgICAgICAgZXhlY3V0b3IocmVzb2x2ZXJGb3IocHJvbWlzZSwgc3RhdGUpLCByZWplY3Rvcik7XHJcbiAgICAgICAgfSBjYXRjaCAoZXJyKSB7XHJcbiAgICAgICAgICAgIGlmIChzdGF0ZS5yZXNvbHZpbmcpIHtcclxuICAgICAgICAgICAgICAgIGNvbnNvbGUubG9nKFwiVW5oYW5kbGVkIGV4Y2VwdGlvbiBpbiBDYW5jZWxsYWJsZVByb21pc2UgZXhlY3V0b3IuXCIsIGVycik7XHJcbiAgICAgICAgICAgIH0gZWxzZSB7XHJcbiAgICAgICAgICAgICAgICByZWplY3RvcihlcnIpO1xyXG4gICAgICAgICAgICB9XHJcbiAgICAgICAgfVxyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ2FuY2VscyBpbW1lZGlhdGVseSB0aGUgZXhlY3V0aW9uIG9mIHRoZSBvcGVyYXRpb24gYXNzb2NpYXRlZCB3aXRoIHRoaXMgcHJvbWlzZS5cclxuICAgICAqIFRoZSBwcm9taXNlIHJlamVjdHMgd2l0aCBhIHtAbGluayBDYW5jZWxFcnJvcn0gaW5zdGFuY2UgYXMgcmVhc29uLFxyXG4gICAgICogd2l0aCB0aGUge0BsaW5rIENhbmNlbEVycm9yI2NhdXNlfSBwcm9wZXJ0eSBzZXQgdG8gdGhlIGdpdmVuIGFyZ3VtZW50LCBpZiBhbnkuXHJcbiAgICAgKlxyXG4gICAgICogSGFzIG5vIGVmZmVjdCBpZiBjYWxsZWQgYWZ0ZXIgdGhlIHByb21pc2UgaGFzIGFscmVhZHkgc2V0dGxlZDtcclxuICAgICAqIHJlcGVhdGVkIGNhbGxzIGluIHBhcnRpY3VsYXIgYXJlIHNhZmUsIGJ1dCBvbmx5IHRoZSBmaXJzdCBvbmVcclxuICAgICAqIHdpbGwgc2V0IHRoZSBjYW5jZWxsYXRpb24gY2F1c2UuXHJcbiAgICAgKlxyXG4gICAgICogVGhlIGBDYW5jZWxFcnJvcmAgZXhjZXB0aW9uIF9uZWVkIG5vdF8gYmUgaGFuZGxlZCBleHBsaWNpdGx5IF9vbiB0aGUgcHJvbWlzZXMgdGhhdCBhcmUgYmVpbmcgY2FuY2VsbGVkOl9cclxuICAgICAqIGNhbmNlbGxpbmcgYSBwcm9taXNlIHdpdGggbm8gYXR0YWNoZWQgcmVqZWN0aW9uIGhhbmRsZXIgZG9lcyBub3QgdHJpZ2dlciBhbiB1bmhhbmRsZWQgcmVqZWN0aW9uIGV2ZW50LlxyXG4gICAgICogVGhlcmVmb3JlLCB0aGUgZm9sbG93aW5nIGlkaW9tcyBhcmUgYWxsIGVxdWFsbHkgY29ycmVjdDpcclxuICAgICAqIGBgYHRzXHJcbiAgICAgKiBuZXcgQ2FuY2VsbGFibGVQcm9taXNlKChyZXNvbHZlLCByZWplY3QpID0+IHsgLi4uIH0pLmNhbmNlbCgpO1xyXG4gICAgICogbmV3IENhbmNlbGxhYmxlUHJvbWlzZSgocmVzb2x2ZSwgcmVqZWN0KSA9PiB7IC4uLiB9KS50aGVuKC4uLikuY2FuY2VsKCk7XHJcbiAgICAgKiBuZXcgQ2FuY2VsbGFibGVQcm9taXNlKChyZXNvbHZlLCByZWplY3QpID0+IHsgLi4uIH0pLnRoZW4oLi4uKS5jYXRjaCguLi4pLmNhbmNlbCgpO1xyXG4gICAgICogYGBgXHJcbiAgICAgKiBXaGVuZXZlciBzb21lIGNhbmNlbGxlZCBwcm9taXNlIGluIGEgY2hhaW4gcmVqZWN0cyB3aXRoIGEgYENhbmNlbEVycm9yYFxyXG4gICAgICogd2l0aCB0aGUgc2FtZSBjYW5jZWxsYXRpb24gY2F1c2UgYXMgaXRzZWxmLCB0aGUgZXJyb3Igd2lsbCBiZSBkaXNjYXJkZWQgc2lsZW50bHkuXHJcbiAgICAgKiBIb3dldmVyLCB0aGUgYENhbmNlbEVycm9yYCBfd2lsbCBzdGlsbCBiZSBkZWxpdmVyZWRfIHRvIGFsbCBhdHRhY2hlZCByZWplY3Rpb24gaGFuZGxlcnNcclxuICAgICAqIGFkZGVkIGJ5IHtAbGluayB0aGVufSBhbmQgcmVsYXRlZCBtZXRob2RzOlxyXG4gICAgICogYGBgdHNcclxuICAgICAqIGxldCBjYW5jZWxsYWJsZSA9IG5ldyBDYW5jZWxsYWJsZVByb21pc2UoKHJlc29sdmUsIHJlamVjdCkgPT4geyAuLi4gfSk7XHJcbiAgICAgKiBjYW5jZWxsYWJsZS50aGVuKCgpID0+IHsgLi4uIH0pLmNhdGNoKGNvbnNvbGUubG9nKTtcclxuICAgICAqIGNhbmNlbGxhYmxlLmNhbmNlbCgpOyAvLyBBIENhbmNlbEVycm9yIGlzIHByaW50ZWQgdG8gdGhlIGNvbnNvbGUuXHJcbiAgICAgKiBgYGBcclxuICAgICAqIElmIHRoZSBgQ2FuY2VsRXJyb3JgIGlzIG5vdCBoYW5kbGVkIGRvd25zdHJlYW0gYnkgdGhlIHRpbWUgaXQgcmVhY2hlc1xyXG4gICAgICogYSBfbm9uLWNhbmNlbGxlZF8gcHJvbWlzZSwgaXQgX3dpbGxfIHRyaWdnZXIgYW4gdW5oYW5kbGVkIHJlamVjdGlvbiBldmVudCxcclxuICAgICAqIGp1c3QgbGlrZSBub3JtYWwgcmVqZWN0aW9ucyB3b3VsZDpcclxuICAgICAqIGBgYHRzXHJcbiAgICAgKiBsZXQgY2FuY2VsbGFibGUgPSBuZXcgQ2FuY2VsbGFibGVQcm9taXNlKChyZXNvbHZlLCByZWplY3QpID0+IHsgLi4uIH0pO1xyXG4gICAgICogbGV0IGNoYWluZWQgPSBjYW5jZWxsYWJsZS50aGVuKCgpID0+IHsgLi4uIH0pLnRoZW4oKCkgPT4geyAuLi4gfSk7IC8vIE5vIGNhdGNoLi4uXHJcbiAgICAgKiBjYW5jZWxsYWJsZS5jYW5jZWwoKTsgLy8gVW5oYW5kbGVkIHJlamVjdGlvbiBldmVudCBvbiBjaGFpbmVkIVxyXG4gICAgICogYGBgXHJcbiAgICAgKiBUaGVyZWZvcmUsIGl0IGlzIGltcG9ydGFudCB0byBlaXRoZXIgY2FuY2VsIHdob2xlIHByb21pc2UgY2hhaW5zIGZyb20gdGhlaXIgdGFpbCxcclxuICAgICAqIGFzIHNob3duIGluIHRoZSBjb3JyZWN0IGlkaW9tcyBhYm92ZSwgb3IgdGFrZSBjYXJlIG9mIGhhbmRsaW5nIGVycm9ycyBldmVyeXdoZXJlLlxyXG4gICAgICpcclxuICAgICAqIEByZXR1cm5zIEEgY2FuY2VsbGFibGUgcHJvbWlzZSB0aGF0IF9mdWxmaWxsc18gYWZ0ZXIgdGhlIGNhbmNlbCBjYWxsYmFjayAoaWYgYW55KVxyXG4gICAgICogYW5kIGFsbCBoYW5kbGVycyBhdHRhY2hlZCB1cCB0byB0aGUgY2FsbCB0byBjYW5jZWwgaGF2ZSBydW4uXHJcbiAgICAgKiBJZiB0aGUgY2FuY2VsIGNhbGxiYWNrIHJldHVybnMgYSB0aGVuYWJsZSwgdGhlIHByb21pc2UgcmV0dXJuZWQgYnkgYGNhbmNlbGBcclxuICAgICAqIHdpbGwgYWxzbyB3YWl0IGZvciB0aGF0IHRoZW5hYmxlIHRvIHNldHRsZS5cclxuICAgICAqIFRoaXMgZW5hYmxlcyBjYWxsZXJzIHRvIHdhaXQgZm9yIHRoZSBjYW5jZWxsZWQgb3BlcmF0aW9uIHRvIHRlcm1pbmF0ZVxyXG4gICAgICogd2l0aG91dCBiZWluZyBmb3JjZWQgdG8gaGFuZGxlIHBvdGVudGlhbCBlcnJvcnMgYXQgdGhlIGNhbGwgc2l0ZS5cclxuICAgICAqIGBgYHRzXHJcbiAgICAgKiBjYW5jZWxsYWJsZS5jYW5jZWwoKS50aGVuKCgpID0+IHtcclxuICAgICAqICAgICAvLyBDbGVhbnVwIGZpbmlzaGVkLCBpdCdzIHNhZmUgdG8gZG8gc29tZXRoaW5nIGVsc2UuXHJcbiAgICAgKiB9LCAoZXJyKSA9PiB7XHJcbiAgICAgKiAgICAgLy8gVW5yZWFjaGFibGU6IHRoZSBwcm9taXNlIHJldHVybmVkIGZyb20gY2FuY2VsIHdpbGwgbmV2ZXIgcmVqZWN0LlxyXG4gICAgICogfSk7XHJcbiAgICAgKiBgYGBcclxuICAgICAqIE5vdGUgdGhhdCB0aGUgcmV0dXJuZWQgcHJvbWlzZSB3aWxsIF9ub3RfIGhhbmRsZSBpbXBsaWNpdGx5IGFueSByZWplY3Rpb25cclxuICAgICAqIHRoYXQgbWlnaHQgaGF2ZSBvY2N1cnJlZCBhbHJlYWR5IGluIHRoZSBjYW5jZWxsZWQgY2hhaW4uXHJcbiAgICAgKiBJdCB3aWxsIGp1c3QgdHJhY2sgd2hldGhlciByZWdpc3RlcmVkIGhhbmRsZXJzIGhhdmUgYmVlbiBleGVjdXRlZCBvciBub3QuXHJcbiAgICAgKiBUaGVyZWZvcmUsIHVuaGFuZGxlZCByZWplY3Rpb25zIHdpbGwgbmV2ZXIgYmUgc2lsZW50bHkgaGFuZGxlZCBieSBjYWxsaW5nIGNhbmNlbC5cclxuICAgICAqL1xyXG4gICAgY2FuY2VsKGNhdXNlPzogYW55KTogQ2FuY2VsbGFibGVQcm9taXNlPHZvaWQ+IHtcclxuICAgICAgICByZXR1cm4gbmV3IENhbmNlbGxhYmxlUHJvbWlzZTx2b2lkPigocmVzb2x2ZSkgPT4ge1xyXG4gICAgICAgICAgICAvLyBJTlZBUklBTlQ6IHRoZSByZXN1bHQgb2YgdGhpc1tjYW5jZWxJbXBsU3ltXSBhbmQgdGhlIGJhcnJpZXIgZG8gbm90IGV2ZXIgcmVqZWN0LlxyXG4gICAgICAgICAgICAvLyBVbmZvcnR1bmF0ZWx5IG1hY09TIEhpZ2ggU2llcnJhIGRvZXMgbm90IHN1cHBvcnQgUHJvbWlzZS5hbGxTZXR0bGVkLlxyXG4gICAgICAgICAgICBQcm9taXNlLmFsbChbXHJcbiAgICAgICAgICAgICAgICB0aGlzW2NhbmNlbEltcGxTeW1dKG5ldyBDYW5jZWxFcnJvcihcIlByb21pc2UgY2FuY2VsbGVkLlwiLCB7IGNhdXNlIH0pKSxcclxuICAgICAgICAgICAgICAgIGN1cnJlbnRCYXJyaWVyKHRoaXMpXHJcbiAgICAgICAgICAgIF0pLnRoZW4oKCkgPT4gcmVzb2x2ZSgpLCAoKSA9PiByZXNvbHZlKCkpO1xyXG4gICAgICAgIH0pO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQmluZHMgcHJvbWlzZSBjYW5jZWxsYXRpb24gdG8gdGhlIGFib3J0IGV2ZW50IG9mIHRoZSBnaXZlbiB7QGxpbmsgQWJvcnRTaWduYWx9LlxyXG4gICAgICogSWYgdGhlIHNpZ25hbCBoYXMgYWxyZWFkeSBhYm9ydGVkLCB0aGUgcHJvbWlzZSB3aWxsIGJlIGNhbmNlbGxlZCBpbW1lZGlhdGVseS5cclxuICAgICAqIFdoZW4gZWl0aGVyIGNvbmRpdGlvbiBpcyB2ZXJpZmllZCwgdGhlIGNhbmNlbGxhdGlvbiBjYXVzZSB3aWxsIGJlIHNldFxyXG4gICAgICogdG8gdGhlIHNpZ25hbCdzIGFib3J0IHJlYXNvbiAoc2VlIHtAbGluayBBYm9ydFNpZ25hbCNyZWFzb259KS5cclxuICAgICAqXHJcbiAgICAgKiBIYXMgbm8gZWZmZWN0IGlmIGNhbGxlZCAob3IgaWYgdGhlIHNpZ25hbCBhYm9ydHMpIF9hZnRlcl8gdGhlIHByb21pc2UgaGFzIGFscmVhZHkgc2V0dGxlZC5cclxuICAgICAqIE9ubHkgdGhlIGZpcnN0IHNpZ25hbCB0byBhYm9ydCB3aWxsIHNldCB0aGUgY2FuY2VsbGF0aW9uIGNhdXNlLlxyXG4gICAgICpcclxuICAgICAqIEZvciBtb3JlIGRldGFpbHMgYWJvdXQgdGhlIGNhbmNlbGxhdGlvbiBwcm9jZXNzLFxyXG4gICAgICogc2VlIHtAbGluayBjYW5jZWx9IGFuZCB0aGUgYENhbmNlbGxhYmxlUHJvbWlzZWAgY29uc3RydWN0b3IuXHJcbiAgICAgKlxyXG4gICAgICogVGhpcyBtZXRob2QgZW5hYmxlcyBgYXdhaXRgaW5nIGNhbmNlbGxhYmxlIHByb21pc2VzIHdpdGhvdXQgaGF2aW5nXHJcbiAgICAgKiB0byBzdG9yZSB0aGVtIGZvciBmdXR1cmUgY2FuY2VsbGF0aW9uLCBlLmcuOlxyXG4gICAgICogYGBgdHNcclxuICAgICAqIGF3YWl0IGxvbmdSdW5uaW5nT3BlcmF0aW9uKCkuY2FuY2VsT24oc2lnbmFsKTtcclxuICAgICAqIGBgYFxyXG4gICAgICogaW5zdGVhZCBvZjpcclxuICAgICAqIGBgYHRzXHJcbiAgICAgKiBsZXQgcHJvbWlzZVRvQmVDYW5jZWxsZWQgPSBsb25nUnVubmluZ09wZXJhdGlvbigpO1xyXG4gICAgICogYXdhaXQgcHJvbWlzZVRvQmVDYW5jZWxsZWQ7XHJcbiAgICAgKiBgYGBcclxuICAgICAqXHJcbiAgICAgKiBAcmV0dXJucyBUaGlzIHByb21pc2UsIGZvciBtZXRob2QgY2hhaW5pbmcuXHJcbiAgICAgKi9cclxuICAgIGNhbmNlbE9uKHNpZ25hbDogQWJvcnRTaWduYWwpOiBDYW5jZWxsYWJsZVByb21pc2U8VD4ge1xyXG4gICAgICAgIGlmIChzaWduYWwuYWJvcnRlZCkge1xyXG4gICAgICAgICAgICB2b2lkIHRoaXMuY2FuY2VsKHNpZ25hbC5yZWFzb24pXHJcbiAgICAgICAgfSBlbHNlIHtcclxuICAgICAgICAgICAgc2lnbmFsLmFkZEV2ZW50TGlzdGVuZXIoJ2Fib3J0JywgKCkgPT4gdm9pZCB0aGlzLmNhbmNlbChzaWduYWwucmVhc29uKSwge2NhcHR1cmU6IHRydWV9KTtcclxuICAgICAgICB9XHJcblxyXG4gICAgICAgIHJldHVybiB0aGlzO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQXR0YWNoZXMgY2FsbGJhY2tzIGZvciB0aGUgcmVzb2x1dGlvbiBhbmQvb3IgcmVqZWN0aW9uIG9mIHRoZSBgQ2FuY2VsbGFibGVQcm9taXNlYC5cclxuICAgICAqXHJcbiAgICAgKiBUaGUgb3B0aW9uYWwgYG9uY2FuY2VsbGVkYCBhcmd1bWVudCB3aWxsIGJlIGludm9rZWQgd2hlbiB0aGUgcmV0dXJuZWQgcHJvbWlzZSBpcyBjYW5jZWxsZWQsXHJcbiAgICAgKiB3aXRoIHRoZSBzYW1lIHNlbWFudGljcyBhcyB0aGUgYG9uY2FuY2VsbGVkYCBhcmd1bWVudCBvZiB0aGUgY29uc3RydWN0b3IuXHJcbiAgICAgKiBXaGVuIHRoZSBwYXJlbnQgcHJvbWlzZSByZWplY3RzIG9yIGlzIGNhbmNlbGxlZCwgdGhlIGBvbnJlamVjdGVkYCBjYWxsYmFjayB3aWxsIHJ1bixcclxuICAgICAqIF9ldmVuIGFmdGVyIHRoZSByZXR1cm5lZCBwcm9taXNlIGhhcyBiZWVuIGNhbmNlbGxlZDpfXHJcbiAgICAgKiBpbiB0aGF0IGNhc2UsIHNob3VsZCBpdCByZWplY3Qgb3IgdGhyb3csIHRoZSByZWFzb24gd2lsbCBiZSB3cmFwcGVkXHJcbiAgICAgKiBpbiBhIHtAbGluayBDYW5jZWxsZWRSZWplY3Rpb25FcnJvcn0gYW5kIGJ1YmJsZWQgdXAgYXMgYW4gdW5oYW5kbGVkIHJlamVjdGlvbi5cclxuICAgICAqXHJcbiAgICAgKiBAcGFyYW0gb25mdWxmaWxsZWQgVGhlIGNhbGxiYWNrIHRvIGV4ZWN1dGUgd2hlbiB0aGUgUHJvbWlzZSBpcyByZXNvbHZlZC5cclxuICAgICAqIEBwYXJhbSBvbnJlamVjdGVkIFRoZSBjYWxsYmFjayB0byBleGVjdXRlIHdoZW4gdGhlIFByb21pc2UgaXMgcmVqZWN0ZWQuXHJcbiAgICAgKiBAcmV0dXJucyBBIGBDYW5jZWxsYWJsZVByb21pc2VgIGZvciB0aGUgY29tcGxldGlvbiBvZiB3aGljaGV2ZXIgY2FsbGJhY2sgaXMgZXhlY3V0ZWQuXHJcbiAgICAgKiBUaGUgcmV0dXJuZWQgcHJvbWlzZSBpcyBob29rZWQgdXAgdG8gcHJvcGFnYXRlIGNhbmNlbGxhdGlvbiByZXF1ZXN0cyB1cCB0aGUgY2hhaW4sIGJ1dCBub3QgZG93bjpcclxuICAgICAqXHJcbiAgICAgKiAgIC0gaWYgdGhlIHBhcmVudCBwcm9taXNlIGlzIGNhbmNlbGxlZCwgdGhlIGBvbnJlamVjdGVkYCBoYW5kbGVyIHdpbGwgYmUgaW52b2tlZCB3aXRoIGEgYENhbmNlbEVycm9yYFxyXG4gICAgICogICAgIGFuZCB0aGUgcmV0dXJuZWQgcHJvbWlzZSBfd2lsbCByZXNvbHZlIHJlZ3VsYXJseV8gd2l0aCBpdHMgcmVzdWx0O1xyXG4gICAgICogICAtIGNvbnZlcnNlbHksIGlmIHRoZSByZXR1cm5lZCBwcm9taXNlIGlzIGNhbmNlbGxlZCwgX3RoZSBwYXJlbnQgcHJvbWlzZSBpcyBjYW5jZWxsZWQgdG9vO19cclxuICAgICAqICAgICB0aGUgYG9ucmVqZWN0ZWRgIGhhbmRsZXIgd2lsbCBzdGlsbCBiZSBpbnZva2VkIHdpdGggdGhlIHBhcmVudCdzIGBDYW5jZWxFcnJvcmAsXHJcbiAgICAgKiAgICAgYnV0IGl0cyByZXN1bHQgd2lsbCBiZSBkaXNjYXJkZWRcclxuICAgICAqICAgICBhbmQgdGhlIHJldHVybmVkIHByb21pc2Ugd2lsbCByZWplY3Qgd2l0aCBhIGBDYW5jZWxFcnJvcmAgYXMgd2VsbC5cclxuICAgICAqXHJcbiAgICAgKiBUaGUgcHJvbWlzZSByZXR1cm5lZCBmcm9tIHtAbGluayBjYW5jZWx9IHdpbGwgZnVsZmlsbCBvbmx5IGFmdGVyIGFsbCBhdHRhY2hlZCBoYW5kbGVyc1xyXG4gICAgICogdXAgdGhlIGVudGlyZSBwcm9taXNlIGNoYWluIGhhdmUgYmVlbiBydW4uXHJcbiAgICAgKlxyXG4gICAgICogSWYgZWl0aGVyIGNhbGxiYWNrIHJldHVybnMgYSBjYW5jZWxsYWJsZSBwcm9taXNlLFxyXG4gICAgICogY2FuY2VsbGF0aW9uIHJlcXVlc3RzIHdpbGwgYmUgZGl2ZXJ0ZWQgdG8gaXQsXHJcbiAgICAgKiBhbmQgdGhlIHNwZWNpZmllZCBgb25jYW5jZWxsZWRgIGNhbGxiYWNrIHdpbGwgYmUgZGlzY2FyZGVkLlxyXG4gICAgICovXHJcbiAgICB0aGVuPFRSZXN1bHQxID0gVCwgVFJlc3VsdDIgPSBuZXZlcj4ob25mdWxmaWxsZWQ/OiAoKHZhbHVlOiBUKSA9PiBUUmVzdWx0MSB8IFByb21pc2VMaWtlPFRSZXN1bHQxPiB8IENhbmNlbGxhYmxlUHJvbWlzZUxpa2U8VFJlc3VsdDE+KSB8IHVuZGVmaW5lZCB8IG51bGwsIG9ucmVqZWN0ZWQ/OiAoKHJlYXNvbjogYW55KSA9PiBUUmVzdWx0MiB8IFByb21pc2VMaWtlPFRSZXN1bHQyPiB8IENhbmNlbGxhYmxlUHJvbWlzZUxpa2U8VFJlc3VsdDI+KSB8IHVuZGVmaW5lZCB8IG51bGwsIG9uY2FuY2VsbGVkPzogQ2FuY2VsbGFibGVQcm9taXNlQ2FuY2VsbGVyKTogQ2FuY2VsbGFibGVQcm9taXNlPFRSZXN1bHQxIHwgVFJlc3VsdDI+IHtcclxuICAgICAgICBpZiAoISh0aGlzIGluc3RhbmNlb2YgQ2FuY2VsbGFibGVQcm9taXNlKSkge1xyXG4gICAgICAgICAgICB0aHJvdyBuZXcgVHlwZUVycm9yKFwiQ2FuY2VsbGFibGVQcm9taXNlLnByb3RvdHlwZS50aGVuIGNhbGxlZCBvbiBhbiBpbnZhbGlkIG9iamVjdC5cIik7XHJcbiAgICAgICAgfVxyXG5cclxuICAgICAgICAvLyBOT1RFOiBUeXBlU2NyaXB0J3MgYnVpbHQtaW4gdHlwZSBmb3IgdGhlbiBpcyBicm9rZW4sXHJcbiAgICAgICAgLy8gYXMgaXQgYWxsb3dzIHNwZWNpZnlpbmcgYW4gYXJiaXRyYXJ5IFRSZXN1bHQxICE9IFQgZXZlbiB3aGVuIG9uZnVsZmlsbGVkIGlzIG5vdCBhIGZ1bmN0aW9uLlxyXG4gICAgICAgIC8vIFdlIGNhbm5vdCBmaXggaXQgaWYgd2Ugd2FudCB0byBDYW5jZWxsYWJsZVByb21pc2UgdG8gaW1wbGVtZW50IFByb21pc2VMaWtlPFQ+LlxyXG5cclxuICAgICAgICBpZiAoIWlzQ2FsbGFibGUob25mdWxmaWxsZWQpKSB7IG9uZnVsZmlsbGVkID0gaWRlbnRpdHkgYXMgYW55OyB9XHJcbiAgICAgICAgaWYgKCFpc0NhbGxhYmxlKG9ucmVqZWN0ZWQpKSB7IG9ucmVqZWN0ZWQgPSB0aHJvd2VyOyB9XHJcblxyXG4gICAgICAgIGlmIChvbmZ1bGZpbGxlZCA9PT0gaWRlbnRpdHkgJiYgb25yZWplY3RlZCA9PSB0aHJvd2VyKSB7XHJcbiAgICAgICAgICAgIC8vIFNob3J0Y3V0IGZvciB0cml2aWFsIGFyZ3VtZW50cy5cclxuICAgICAgICAgICAgcmV0dXJuIG5ldyBDYW5jZWxsYWJsZVByb21pc2UoKHJlc29sdmUpID0+IHJlc29sdmUodGhpcyBhcyBhbnkpKTtcclxuICAgICAgICB9XHJcblxyXG4gICAgICAgIGNvbnN0IGJhcnJpZXI6IFBhcnRpYWw8UHJvbWlzZVdpdGhSZXNvbHZlcnM8dm9pZD4+ID0ge307XHJcbiAgICAgICAgdGhpc1tiYXJyaWVyU3ltXSA9IGJhcnJpZXI7XHJcblxyXG4gICAgICAgIHJldHVybiBuZXcgQ2FuY2VsbGFibGVQcm9taXNlPFRSZXN1bHQxIHwgVFJlc3VsdDI+KChyZXNvbHZlLCByZWplY3QpID0+IHtcclxuICAgICAgICAgICAgdm9pZCBzdXBlci50aGVuKFxyXG4gICAgICAgICAgICAgICAgKHZhbHVlKSA9PiB7XHJcbiAgICAgICAgICAgICAgICAgICAgaWYgKHRoaXNbYmFycmllclN5bV0gPT09IGJhcnJpZXIpIHsgdGhpc1tiYXJyaWVyU3ltXSA9IG51bGw7IH1cclxuICAgICAgICAgICAgICAgICAgICBiYXJyaWVyLnJlc29sdmU/LigpO1xyXG5cclxuICAgICAgICAgICAgICAgICAgICB0cnkge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICByZXNvbHZlKG9uZnVsZmlsbGVkISh2YWx1ZSkpO1xyXG4gICAgICAgICAgICAgICAgICAgIH0gY2F0Y2ggKGVycikge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICByZWplY3QoZXJyKTtcclxuICAgICAgICAgICAgICAgICAgICB9XHJcbiAgICAgICAgICAgICAgICB9LFxyXG4gICAgICAgICAgICAgICAgKHJlYXNvbj8pID0+IHtcclxuICAgICAgICAgICAgICAgICAgICBpZiAodGhpc1tiYXJyaWVyU3ltXSA9PT0gYmFycmllcikgeyB0aGlzW2JhcnJpZXJTeW1dID0gbnVsbDsgfVxyXG4gICAgICAgICAgICAgICAgICAgIGJhcnJpZXIucmVzb2x2ZT8uKCk7XHJcblxyXG4gICAgICAgICAgICAgICAgICAgIHRyeSB7XHJcbiAgICAgICAgICAgICAgICAgICAgICAgIHJlc29sdmUob25yZWplY3RlZCEocmVhc29uKSk7XHJcbiAgICAgICAgICAgICAgICAgICAgfSBjYXRjaCAoZXJyKSB7XHJcbiAgICAgICAgICAgICAgICAgICAgICAgIHJlamVjdChlcnIpO1xyXG4gICAgICAgICAgICAgICAgICAgIH1cclxuICAgICAgICAgICAgICAgIH1cclxuICAgICAgICAgICAgKTtcclxuICAgICAgICB9LCBhc3luYyAoY2F1c2U/KSA9PiB7XHJcbiAgICAgICAgICAgIC8vY2FuY2VsbGVkID0gdHJ1ZTtcclxuICAgICAgICAgICAgdHJ5IHtcclxuICAgICAgICAgICAgICAgIHJldHVybiBvbmNhbmNlbGxlZD8uKGNhdXNlKTtcclxuICAgICAgICAgICAgfSBmaW5hbGx5IHtcclxuICAgICAgICAgICAgICAgIGF3YWl0IHRoaXMuY2FuY2VsKGNhdXNlKTtcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgIH0pO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQXR0YWNoZXMgYSBjYWxsYmFjayBmb3Igb25seSB0aGUgcmVqZWN0aW9uIG9mIHRoZSBQcm9taXNlLlxyXG4gICAgICpcclxuICAgICAqIFRoZSBvcHRpb25hbCBgb25jYW5jZWxsZWRgIGFyZ3VtZW50IHdpbGwgYmUgaW52b2tlZCB3aGVuIHRoZSByZXR1cm5lZCBwcm9taXNlIGlzIGNhbmNlbGxlZCxcclxuICAgICAqIHdpdGggdGhlIHNhbWUgc2VtYW50aWNzIGFzIHRoZSBgb25jYW5jZWxsZWRgIGFyZ3VtZW50IG9mIHRoZSBjb25zdHJ1Y3Rvci5cclxuICAgICAqIFdoZW4gdGhlIHBhcmVudCBwcm9taXNlIHJlamVjdHMgb3IgaXMgY2FuY2VsbGVkLCB0aGUgYG9ucmVqZWN0ZWRgIGNhbGxiYWNrIHdpbGwgcnVuLFxyXG4gICAgICogX2V2ZW4gYWZ0ZXIgdGhlIHJldHVybmVkIHByb21pc2UgaGFzIGJlZW4gY2FuY2VsbGVkOl9cclxuICAgICAqIGluIHRoYXQgY2FzZSwgc2hvdWxkIGl0IHJlamVjdCBvciB0aHJvdywgdGhlIHJlYXNvbiB3aWxsIGJlIHdyYXBwZWRcclxuICAgICAqIGluIGEge0BsaW5rIENhbmNlbGxlZFJlamVjdGlvbkVycm9yfSBhbmQgYnViYmxlZCB1cCBhcyBhbiB1bmhhbmRsZWQgcmVqZWN0aW9uLlxyXG4gICAgICpcclxuICAgICAqIEl0IGlzIGVxdWl2YWxlbnQgdG9cclxuICAgICAqIGBgYHRzXHJcbiAgICAgKiBjYW5jZWxsYWJsZVByb21pc2UudGhlbih1bmRlZmluZWQsIG9ucmVqZWN0ZWQsIG9uY2FuY2VsbGVkKTtcclxuICAgICAqIGBgYFxyXG4gICAgICogYW5kIHRoZSBzYW1lIGNhdmVhdHMgYXBwbHkuXHJcbiAgICAgKlxyXG4gICAgICogQHJldHVybnMgQSBQcm9taXNlIGZvciB0aGUgY29tcGxldGlvbiBvZiB0aGUgY2FsbGJhY2suXHJcbiAgICAgKiBDYW5jZWxsYXRpb24gcmVxdWVzdHMgb24gdGhlIHJldHVybmVkIHByb21pc2VcclxuICAgICAqIHdpbGwgcHJvcGFnYXRlIHVwIHRoZSBjaGFpbiB0byB0aGUgcGFyZW50IHByb21pc2UsXHJcbiAgICAgKiBidXQgbm90IGluIHRoZSBvdGhlciBkaXJlY3Rpb24uXHJcbiAgICAgKlxyXG4gICAgICogVGhlIHByb21pc2UgcmV0dXJuZWQgZnJvbSB7QGxpbmsgY2FuY2VsfSB3aWxsIGZ1bGZpbGwgb25seSBhZnRlciBhbGwgYXR0YWNoZWQgaGFuZGxlcnNcclxuICAgICAqIHVwIHRoZSBlbnRpcmUgcHJvbWlzZSBjaGFpbiBoYXZlIGJlZW4gcnVuLlxyXG4gICAgICpcclxuICAgICAqIElmIGBvbnJlamVjdGVkYCByZXR1cm5zIGEgY2FuY2VsbGFibGUgcHJvbWlzZSxcclxuICAgICAqIGNhbmNlbGxhdGlvbiByZXF1ZXN0cyB3aWxsIGJlIGRpdmVydGVkIHRvIGl0LFxyXG4gICAgICogYW5kIHRoZSBzcGVjaWZpZWQgYG9uY2FuY2VsbGVkYCBjYWxsYmFjayB3aWxsIGJlIGRpc2NhcmRlZC5cclxuICAgICAqIFNlZSB7QGxpbmsgdGhlbn0gZm9yIG1vcmUgZGV0YWlscy5cclxuICAgICAqL1xyXG4gICAgY2F0Y2g8VFJlc3VsdCA9IG5ldmVyPihvbnJlamVjdGVkPzogKChyZWFzb246IGFueSkgPT4gKFByb21pc2VMaWtlPFRSZXN1bHQ+IHwgVFJlc3VsdCkpIHwgdW5kZWZpbmVkIHwgbnVsbCwgb25jYW5jZWxsZWQ/OiBDYW5jZWxsYWJsZVByb21pc2VDYW5jZWxsZXIpOiBDYW5jZWxsYWJsZVByb21pc2U8VCB8IFRSZXN1bHQ+IHtcclxuICAgICAgICByZXR1cm4gdGhpcy50aGVuKHVuZGVmaW5lZCwgb25yZWplY3RlZCwgb25jYW5jZWxsZWQpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQXR0YWNoZXMgYSBjYWxsYmFjayB0aGF0IGlzIGludm9rZWQgd2hlbiB0aGUgQ2FuY2VsbGFibGVQcm9taXNlIGlzIHNldHRsZWQgKGZ1bGZpbGxlZCBvciByZWplY3RlZCkuIFRoZVxyXG4gICAgICogcmVzb2x2ZWQgdmFsdWUgY2Fubm90IGJlIGFjY2Vzc2VkIG9yIG1vZGlmaWVkIGZyb20gdGhlIGNhbGxiYWNrLlxyXG4gICAgICogVGhlIHJldHVybmVkIHByb21pc2Ugd2lsbCBzZXR0bGUgaW4gdGhlIHNhbWUgc3RhdGUgYXMgdGhlIG9yaWdpbmFsIG9uZVxyXG4gICAgICogYWZ0ZXIgdGhlIHByb3ZpZGVkIGNhbGxiYWNrIGhhcyBjb21wbGV0ZWQgZXhlY3V0aW9uLFxyXG4gICAgICogdW5sZXNzIHRoZSBjYWxsYmFjayB0aHJvd3Mgb3IgcmV0dXJucyBhIHJlamVjdGluZyBwcm9taXNlLFxyXG4gICAgICogaW4gd2hpY2ggY2FzZSB0aGUgcmV0dXJuZWQgcHJvbWlzZSB3aWxsIHJlamVjdCBhcyB3ZWxsLlxyXG4gICAgICpcclxuICAgICAqIFRoZSBvcHRpb25hbCBgb25jYW5jZWxsZWRgIGFyZ3VtZW50IHdpbGwgYmUgaW52b2tlZCB3aGVuIHRoZSByZXR1cm5lZCBwcm9taXNlIGlzIGNhbmNlbGxlZCxcclxuICAgICAqIHdpdGggdGhlIHNhbWUgc2VtYW50aWNzIGFzIHRoZSBgb25jYW5jZWxsZWRgIGFyZ3VtZW50IG9mIHRoZSBjb25zdHJ1Y3Rvci5cclxuICAgICAqIE9uY2UgdGhlIHBhcmVudCBwcm9taXNlIHNldHRsZXMsIHRoZSBgb25maW5hbGx5YCBjYWxsYmFjayB3aWxsIHJ1bixcclxuICAgICAqIF9ldmVuIGFmdGVyIHRoZSByZXR1cm5lZCBwcm9taXNlIGhhcyBiZWVuIGNhbmNlbGxlZDpfXHJcbiAgICAgKiBpbiB0aGF0IGNhc2UsIHNob3VsZCBpdCByZWplY3Qgb3IgdGhyb3csIHRoZSByZWFzb24gd2lsbCBiZSB3cmFwcGVkXHJcbiAgICAgKiBpbiBhIHtAbGluayBDYW5jZWxsZWRSZWplY3Rpb25FcnJvcn0gYW5kIGJ1YmJsZWQgdXAgYXMgYW4gdW5oYW5kbGVkIHJlamVjdGlvbi5cclxuICAgICAqXHJcbiAgICAgKiBUaGlzIG1ldGhvZCBpcyBpbXBsZW1lbnRlZCBpbiB0ZXJtcyBvZiB7QGxpbmsgdGhlbn0gYW5kIHRoZSBzYW1lIGNhdmVhdHMgYXBwbHkuXHJcbiAgICAgKiBJdCBpcyBwb2x5ZmlsbGVkLCBoZW5jZSBhdmFpbGFibGUgaW4gZXZlcnkgT1Mvd2VidmlldyB2ZXJzaW9uLlxyXG4gICAgICpcclxuICAgICAqIEByZXR1cm5zIEEgUHJvbWlzZSBmb3IgdGhlIGNvbXBsZXRpb24gb2YgdGhlIGNhbGxiYWNrLlxyXG4gICAgICogQ2FuY2VsbGF0aW9uIHJlcXVlc3RzIG9uIHRoZSByZXR1cm5lZCBwcm9taXNlXHJcbiAgICAgKiB3aWxsIHByb3BhZ2F0ZSB1cCB0aGUgY2hhaW4gdG8gdGhlIHBhcmVudCBwcm9taXNlLFxyXG4gICAgICogYnV0IG5vdCBpbiB0aGUgb3RoZXIgZGlyZWN0aW9uLlxyXG4gICAgICpcclxuICAgICAqIFRoZSBwcm9taXNlIHJldHVybmVkIGZyb20ge0BsaW5rIGNhbmNlbH0gd2lsbCBmdWxmaWxsIG9ubHkgYWZ0ZXIgYWxsIGF0dGFjaGVkIGhhbmRsZXJzXHJcbiAgICAgKiB1cCB0aGUgZW50aXJlIHByb21pc2UgY2hhaW4gaGF2ZSBiZWVuIHJ1bi5cclxuICAgICAqXHJcbiAgICAgKiBJZiBgb25maW5hbGx5YCByZXR1cm5zIGEgY2FuY2VsbGFibGUgcHJvbWlzZSxcclxuICAgICAqIGNhbmNlbGxhdGlvbiByZXF1ZXN0cyB3aWxsIGJlIGRpdmVydGVkIHRvIGl0LFxyXG4gICAgICogYW5kIHRoZSBzcGVjaWZpZWQgYG9uY2FuY2VsbGVkYCBjYWxsYmFjayB3aWxsIGJlIGRpc2NhcmRlZC5cclxuICAgICAqIFNlZSB7QGxpbmsgdGhlbn0gZm9yIG1vcmUgZGV0YWlscy5cclxuICAgICAqL1xyXG4gICAgZmluYWxseShvbmZpbmFsbHk/OiAoKCkgPT4gdm9pZCkgfCB1bmRlZmluZWQgfCBudWxsLCBvbmNhbmNlbGxlZD86IENhbmNlbGxhYmxlUHJvbWlzZUNhbmNlbGxlcik6IENhbmNlbGxhYmxlUHJvbWlzZTxUPiB7XHJcbiAgICAgICAgaWYgKCEodGhpcyBpbnN0YW5jZW9mIENhbmNlbGxhYmxlUHJvbWlzZSkpIHtcclxuICAgICAgICAgICAgdGhyb3cgbmV3IFR5cGVFcnJvcihcIkNhbmNlbGxhYmxlUHJvbWlzZS5wcm90b3R5cGUuZmluYWxseSBjYWxsZWQgb24gYW4gaW52YWxpZCBvYmplY3QuXCIpO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgaWYgKCFpc0NhbGxhYmxlKG9uZmluYWxseSkpIHtcclxuICAgICAgICAgICAgcmV0dXJuIHRoaXMudGhlbihvbmZpbmFsbHksIG9uZmluYWxseSwgb25jYW5jZWxsZWQpO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgcmV0dXJuIHRoaXMudGhlbihcclxuICAgICAgICAgICAgKHZhbHVlKSA9PiBDYW5jZWxsYWJsZVByb21pc2UucmVzb2x2ZShvbmZpbmFsbHkoKSkudGhlbigoKSA9PiB2YWx1ZSksXHJcbiAgICAgICAgICAgIChyZWFzb24/KSA9PiBDYW5jZWxsYWJsZVByb21pc2UucmVzb2x2ZShvbmZpbmFsbHkoKSkudGhlbigoKSA9PiB7IHRocm93IHJlYXNvbjsgfSksXHJcbiAgICAgICAgICAgIG9uY2FuY2VsbGVkLFxyXG4gICAgICAgICk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBXZSB1c2UgdGhlIGBbU3ltYm9sLnNwZWNpZXNdYCBzdGF0aWMgcHJvcGVydHksIGlmIGF2YWlsYWJsZSxcclxuICAgICAqIHRvIGRpc2FibGUgdGhlIGJ1aWx0LWluIGF1dG9tYXRpYyBzdWJjbGFzc2luZyBmZWF0dXJlcyBmcm9tIHtAbGluayBQcm9taXNlfS5cclxuICAgICAqIEl0IGlzIGNyaXRpY2FsIGZvciBwZXJmb3JtYW5jZSByZWFzb25zIHRoYXQgZXh0ZW5kZXJzIGRvIG5vdCBvdmVycmlkZSB0aGlzLlxyXG4gICAgICogT25jZSB0aGUgcHJvcG9zYWwgYXQgaHR0cHM6Ly9naXRodWIuY29tL3RjMzkvcHJvcG9zYWwtcm0tYnVpbHRpbi1zdWJjbGFzc2luZ1xyXG4gICAgICogaXMgZWl0aGVyIGFjY2VwdGVkIG9yIHJldGlyZWQsIHRoaXMgaW1wbGVtZW50YXRpb24gd2lsbCBoYXZlIHRvIGJlIHJldmlzZWQgYWNjb3JkaW5nbHkuXHJcbiAgICAgKlxyXG4gICAgICogQGlnbm9yZVxyXG4gICAgICogQGludGVybmFsXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyBnZXQgW3NwZWNpZXNdKCkge1xyXG4gICAgICAgIHJldHVybiBQcm9taXNlO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ3JlYXRlcyBhIENhbmNlbGxhYmxlUHJvbWlzZSB0aGF0IGlzIHJlc29sdmVkIHdpdGggYW4gYXJyYXkgb2YgcmVzdWx0c1xyXG4gICAgICogd2hlbiBhbGwgb2YgdGhlIHByb3ZpZGVkIFByb21pc2VzIHJlc29sdmUsIG9yIHJlamVjdGVkIHdoZW4gYW55IFByb21pc2UgaXMgcmVqZWN0ZWQuXHJcbiAgICAgKlxyXG4gICAgICogRXZlcnkgb25lIG9mIHRoZSBwcm92aWRlZCBvYmplY3RzIHRoYXQgaXMgYSB0aGVuYWJsZSBfYW5kXyBjYW5jZWxsYWJsZSBvYmplY3RcclxuICAgICAqIHdpbGwgYmUgY2FuY2VsbGVkIHdoZW4gdGhlIHJldHVybmVkIHByb21pc2UgaXMgY2FuY2VsbGVkLCB3aXRoIHRoZSBzYW1lIGNhdXNlLlxyXG4gICAgICpcclxuICAgICAqIEBncm91cCBTdGF0aWMgTWV0aG9kc1xyXG4gICAgICovXHJcbiAgICBzdGF0aWMgYWxsPFQ+KHZhbHVlczogSXRlcmFibGU8VCB8IFByb21pc2VMaWtlPFQ+Pik6IENhbmNlbGxhYmxlUHJvbWlzZTxBd2FpdGVkPFQ+W10+O1xyXG4gICAgc3RhdGljIGFsbDxUIGV4dGVuZHMgcmVhZG9ubHkgdW5rbm93bltdIHwgW10+KHZhbHVlczogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTx7IC1yZWFkb25seSBbUCBpbiBrZXlvZiBUXTogQXdhaXRlZDxUW1BdPjsgfT47XHJcbiAgICBzdGF0aWMgYWxsPFQgZXh0ZW5kcyBJdGVyYWJsZTx1bmtub3duPiB8IEFycmF5TGlrZTx1bmtub3duPj4odmFsdWVzOiBUKTogQ2FuY2VsbGFibGVQcm9taXNlPHVua25vd24+IHtcclxuICAgICAgICBsZXQgY29sbGVjdGVkID0gQXJyYXkuZnJvbSh2YWx1ZXMpO1xyXG4gICAgICAgIGNvbnN0IHByb21pc2UgPSBjb2xsZWN0ZWQubGVuZ3RoID09PSAwXHJcbiAgICAgICAgICAgID8gQ2FuY2VsbGFibGVQcm9taXNlLnJlc29sdmUoY29sbGVjdGVkKVxyXG4gICAgICAgICAgICA6IG5ldyBDYW5jZWxsYWJsZVByb21pc2U8dW5rbm93bj4oKHJlc29sdmUsIHJlamVjdCkgPT4ge1xyXG4gICAgICAgICAgICAgICAgdm9pZCBQcm9taXNlLmFsbChjb2xsZWN0ZWQpLnRoZW4ocmVzb2x2ZSwgcmVqZWN0KTtcclxuICAgICAgICAgICAgfSwgKGNhdXNlPyk6IFByb21pc2U8dm9pZD4gPT4gY2FuY2VsQWxsKHByb21pc2UsIGNvbGxlY3RlZCwgY2F1c2UpKTtcclxuICAgICAgICByZXR1cm4gcHJvbWlzZTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIENyZWF0ZXMgYSBDYW5jZWxsYWJsZVByb21pc2UgdGhhdCBpcyByZXNvbHZlZCB3aXRoIGFuIGFycmF5IG9mIHJlc3VsdHNcclxuICAgICAqIHdoZW4gYWxsIG9mIHRoZSBwcm92aWRlZCBQcm9taXNlcyByZXNvbHZlIG9yIHJlamVjdC5cclxuICAgICAqXHJcbiAgICAgKiBFdmVyeSBvbmUgb2YgdGhlIHByb3ZpZGVkIG9iamVjdHMgdGhhdCBpcyBhIHRoZW5hYmxlIF9hbmRfIGNhbmNlbGxhYmxlIG9iamVjdFxyXG4gICAgICogd2lsbCBiZSBjYW5jZWxsZWQgd2hlbiB0aGUgcmV0dXJuZWQgcHJvbWlzZSBpcyBjYW5jZWxsZWQsIHdpdGggdGhlIHNhbWUgY2F1c2UuXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyBhbGxTZXR0bGVkPFQ+KHZhbHVlczogSXRlcmFibGU8VCB8IFByb21pc2VMaWtlPFQ+Pik6IENhbmNlbGxhYmxlUHJvbWlzZTxQcm9taXNlU2V0dGxlZFJlc3VsdDxBd2FpdGVkPFQ+PltdPjtcclxuICAgIHN0YXRpYyBhbGxTZXR0bGVkPFQgZXh0ZW5kcyByZWFkb25seSB1bmtub3duW10gfCBbXT4odmFsdWVzOiBUKTogQ2FuY2VsbGFibGVQcm9taXNlPHsgLXJlYWRvbmx5IFtQIGluIGtleW9mIFRdOiBQcm9taXNlU2V0dGxlZFJlc3VsdDxBd2FpdGVkPFRbUF0+PjsgfT47XHJcbiAgICBzdGF0aWMgYWxsU2V0dGxlZDxUIGV4dGVuZHMgSXRlcmFibGU8dW5rbm93bj4gfCBBcnJheUxpa2U8dW5rbm93bj4+KHZhbHVlczogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTx1bmtub3duPiB7XHJcbiAgICAgICAgbGV0IGNvbGxlY3RlZCA9IEFycmF5LmZyb20odmFsdWVzKTtcclxuICAgICAgICBjb25zdCBwcm9taXNlID0gY29sbGVjdGVkLmxlbmd0aCA9PT0gMFxyXG4gICAgICAgICAgICA/IENhbmNlbGxhYmxlUHJvbWlzZS5yZXNvbHZlKGNvbGxlY3RlZClcclxuICAgICAgICAgICAgOiBuZXcgQ2FuY2VsbGFibGVQcm9taXNlPHVua25vd24+KChyZXNvbHZlLCByZWplY3QpID0+IHtcclxuICAgICAgICAgICAgICAgIHZvaWQgUHJvbWlzZS5hbGxTZXR0bGVkKGNvbGxlY3RlZCkudGhlbihyZXNvbHZlLCByZWplY3QpO1xyXG4gICAgICAgICAgICB9LCAoY2F1c2U/KTogUHJvbWlzZTx2b2lkPiA9PiBjYW5jZWxBbGwocHJvbWlzZSwgY29sbGVjdGVkLCBjYXVzZSkpO1xyXG4gICAgICAgIHJldHVybiBwcm9taXNlO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogVGhlIGFueSBmdW5jdGlvbiByZXR1cm5zIGEgcHJvbWlzZSB0aGF0IGlzIGZ1bGZpbGxlZCBieSB0aGUgZmlyc3QgZ2l2ZW4gcHJvbWlzZSB0byBiZSBmdWxmaWxsZWQsXHJcbiAgICAgKiBvciByZWplY3RlZCB3aXRoIGFuIEFnZ3JlZ2F0ZUVycm9yIGNvbnRhaW5pbmcgYW4gYXJyYXkgb2YgcmVqZWN0aW9uIHJlYXNvbnNcclxuICAgICAqIGlmIGFsbCBvZiB0aGUgZ2l2ZW4gcHJvbWlzZXMgYXJlIHJlamVjdGVkLlxyXG4gICAgICogSXQgcmVzb2x2ZXMgYWxsIGVsZW1lbnRzIG9mIHRoZSBwYXNzZWQgaXRlcmFibGUgdG8gcHJvbWlzZXMgYXMgaXQgcnVucyB0aGlzIGFsZ29yaXRobS5cclxuICAgICAqXHJcbiAgICAgKiBFdmVyeSBvbmUgb2YgdGhlIHByb3ZpZGVkIG9iamVjdHMgdGhhdCBpcyBhIHRoZW5hYmxlIF9hbmRfIGNhbmNlbGxhYmxlIG9iamVjdFxyXG4gICAgICogd2lsbCBiZSBjYW5jZWxsZWQgd2hlbiB0aGUgcmV0dXJuZWQgcHJvbWlzZSBpcyBjYW5jZWxsZWQsIHdpdGggdGhlIHNhbWUgY2F1c2UuXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyBhbnk8VD4odmFsdWVzOiBJdGVyYWJsZTxUIHwgUHJvbWlzZUxpa2U8VD4+KTogQ2FuY2VsbGFibGVQcm9taXNlPEF3YWl0ZWQ8VD4+O1xyXG4gICAgc3RhdGljIGFueTxUIGV4dGVuZHMgcmVhZG9ubHkgdW5rbm93bltdIHwgW10+KHZhbHVlczogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTxBd2FpdGVkPFRbbnVtYmVyXT4+O1xyXG4gICAgc3RhdGljIGFueTxUIGV4dGVuZHMgSXRlcmFibGU8dW5rbm93bj4gfCBBcnJheUxpa2U8dW5rbm93bj4+KHZhbHVlczogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTx1bmtub3duPiB7XHJcbiAgICAgICAgbGV0IGNvbGxlY3RlZCA9IEFycmF5LmZyb20odmFsdWVzKTtcclxuICAgICAgICBjb25zdCBwcm9taXNlID0gY29sbGVjdGVkLmxlbmd0aCA9PT0gMFxyXG4gICAgICAgICAgICA/IENhbmNlbGxhYmxlUHJvbWlzZS5yZXNvbHZlKGNvbGxlY3RlZClcclxuICAgICAgICAgICAgOiBuZXcgQ2FuY2VsbGFibGVQcm9taXNlPHVua25vd24+KChyZXNvbHZlLCByZWplY3QpID0+IHtcclxuICAgICAgICAgICAgICAgIHZvaWQgUHJvbWlzZS5hbnkoY29sbGVjdGVkKS50aGVuKHJlc29sdmUsIHJlamVjdCk7XHJcbiAgICAgICAgICAgIH0sIChjYXVzZT8pOiBQcm9taXNlPHZvaWQ+ID0+IGNhbmNlbEFsbChwcm9taXNlLCBjb2xsZWN0ZWQsIGNhdXNlKSk7XHJcbiAgICAgICAgcmV0dXJuIHByb21pc2U7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBDcmVhdGVzIGEgUHJvbWlzZSB0aGF0IGlzIHJlc29sdmVkIG9yIHJlamVjdGVkIHdoZW4gYW55IG9mIHRoZSBwcm92aWRlZCBQcm9taXNlcyBhcmUgcmVzb2x2ZWQgb3IgcmVqZWN0ZWQuXHJcbiAgICAgKlxyXG4gICAgICogRXZlcnkgb25lIG9mIHRoZSBwcm92aWRlZCBvYmplY3RzIHRoYXQgaXMgYSB0aGVuYWJsZSBfYW5kXyBjYW5jZWxsYWJsZSBvYmplY3RcclxuICAgICAqIHdpbGwgYmUgY2FuY2VsbGVkIHdoZW4gdGhlIHJldHVybmVkIHByb21pc2UgaXMgY2FuY2VsbGVkLCB3aXRoIHRoZSBzYW1lIGNhdXNlLlxyXG4gICAgICpcclxuICAgICAqIEBncm91cCBTdGF0aWMgTWV0aG9kc1xyXG4gICAgICovXHJcbiAgICBzdGF0aWMgcmFjZTxUPih2YWx1ZXM6IEl0ZXJhYmxlPFQgfCBQcm9taXNlTGlrZTxUPj4pOiBDYW5jZWxsYWJsZVByb21pc2U8QXdhaXRlZDxUPj47XHJcbiAgICBzdGF0aWMgcmFjZTxUIGV4dGVuZHMgcmVhZG9ubHkgdW5rbm93bltdIHwgW10+KHZhbHVlczogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTxBd2FpdGVkPFRbbnVtYmVyXT4+O1xyXG4gICAgc3RhdGljIHJhY2U8VCBleHRlbmRzIEl0ZXJhYmxlPHVua25vd24+IHwgQXJyYXlMaWtlPHVua25vd24+Pih2YWx1ZXM6IFQpOiBDYW5jZWxsYWJsZVByb21pc2U8dW5rbm93bj4ge1xyXG4gICAgICAgIGxldCBjb2xsZWN0ZWQgPSBBcnJheS5mcm9tKHZhbHVlcyk7XHJcbiAgICAgICAgY29uc3QgcHJvbWlzZSA9IG5ldyBDYW5jZWxsYWJsZVByb21pc2U8dW5rbm93bj4oKHJlc29sdmUsIHJlamVjdCkgPT4ge1xyXG4gICAgICAgICAgICB2b2lkIFByb21pc2UucmFjZShjb2xsZWN0ZWQpLnRoZW4ocmVzb2x2ZSwgcmVqZWN0KTtcclxuICAgICAgICB9LCAoY2F1c2U/KTogUHJvbWlzZTx2b2lkPiA9PiBjYW5jZWxBbGwocHJvbWlzZSwgY29sbGVjdGVkLCBjYXVzZSkpO1xyXG4gICAgICAgIHJldHVybiBwcm9taXNlO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ3JlYXRlcyBhIG5ldyBjYW5jZWxsZWQgQ2FuY2VsbGFibGVQcm9taXNlIGZvciB0aGUgcHJvdmlkZWQgY2F1c2UuXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyBjYW5jZWw8VCA9IG5ldmVyPihjYXVzZT86IGFueSk6IENhbmNlbGxhYmxlUHJvbWlzZTxUPiB7XHJcbiAgICAgICAgY29uc3QgcCA9IG5ldyBDYW5jZWxsYWJsZVByb21pc2U8VD4oKCkgPT4ge30pO1xyXG4gICAgICAgIHAuY2FuY2VsKGNhdXNlKTtcclxuICAgICAgICByZXR1cm4gcDtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIENyZWF0ZXMgYSBuZXcgQ2FuY2VsbGFibGVQcm9taXNlIHRoYXQgY2FuY2Vsc1xyXG4gICAgICogYWZ0ZXIgdGhlIHNwZWNpZmllZCB0aW1lb3V0LCB3aXRoIHRoZSBwcm92aWRlZCBjYXVzZS5cclxuICAgICAqXHJcbiAgICAgKiBJZiB0aGUge0BsaW5rIEFib3J0U2lnbmFsLnRpbWVvdXR9IGZhY3RvcnkgbWV0aG9kIGlzIGF2YWlsYWJsZSxcclxuICAgICAqIGl0IGlzIHVzZWQgdG8gYmFzZSB0aGUgdGltZW91dCBvbiBfYWN0aXZlXyB0aW1lIHJhdGhlciB0aGFuIF9lbGFwc2VkXyB0aW1lLlxyXG4gICAgICogT3RoZXJ3aXNlLCBgdGltZW91dGAgZmFsbHMgYmFjayB0byB7QGxpbmsgc2V0VGltZW91dH0uXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyB0aW1lb3V0PFQgPSBuZXZlcj4obWlsbGlzZWNvbmRzOiBudW1iZXIsIGNhdXNlPzogYW55KTogQ2FuY2VsbGFibGVQcm9taXNlPFQ+IHtcclxuICAgICAgICBjb25zdCBwcm9taXNlID0gbmV3IENhbmNlbGxhYmxlUHJvbWlzZTxUPigoKSA9PiB7fSk7XHJcbiAgICAgICAgaWYgKEFib3J0U2lnbmFsICYmIHR5cGVvZiBBYm9ydFNpZ25hbCA9PT0gJ2Z1bmN0aW9uJyAmJiBBYm9ydFNpZ25hbC50aW1lb3V0ICYmIHR5cGVvZiBBYm9ydFNpZ25hbC50aW1lb3V0ID09PSAnZnVuY3Rpb24nKSB7XHJcbiAgICAgICAgICAgIEFib3J0U2lnbmFsLnRpbWVvdXQobWlsbGlzZWNvbmRzKS5hZGRFdmVudExpc3RlbmVyKCdhYm9ydCcsICgpID0+IHZvaWQgcHJvbWlzZS5jYW5jZWwoY2F1c2UpKTtcclxuICAgICAgICB9IGVsc2Uge1xyXG4gICAgICAgICAgICBzZXRUaW1lb3V0KCgpID0+IHZvaWQgcHJvbWlzZS5jYW5jZWwoY2F1c2UpLCBtaWxsaXNlY29uZHMpO1xyXG4gICAgICAgIH1cclxuICAgICAgICByZXR1cm4gcHJvbWlzZTtcclxuICAgIH1cclxuXHJcbiAgICAvKipcclxuICAgICAqIENyZWF0ZXMgYSBuZXcgQ2FuY2VsbGFibGVQcm9taXNlIHRoYXQgcmVzb2x2ZXMgYWZ0ZXIgdGhlIHNwZWNpZmllZCB0aW1lb3V0LlxyXG4gICAgICogVGhlIHJldHVybmVkIHByb21pc2UgY2FuIGJlIGNhbmNlbGxlZCB3aXRob3V0IGNvbnNlcXVlbmNlcy5cclxuICAgICAqXHJcbiAgICAgKiBAZ3JvdXAgU3RhdGljIE1ldGhvZHNcclxuICAgICAqL1xyXG4gICAgc3RhdGljIHNsZWVwKG1pbGxpc2Vjb25kczogbnVtYmVyKTogQ2FuY2VsbGFibGVQcm9taXNlPHZvaWQ+O1xyXG4gICAgLyoqXHJcbiAgICAgKiBDcmVhdGVzIGEgbmV3IENhbmNlbGxhYmxlUHJvbWlzZSB0aGF0IHJlc29sdmVzIGFmdGVyXHJcbiAgICAgKiB0aGUgc3BlY2lmaWVkIHRpbWVvdXQsIHdpdGggdGhlIHByb3ZpZGVkIHZhbHVlLlxyXG4gICAgICogVGhlIHJldHVybmVkIHByb21pc2UgY2FuIGJlIGNhbmNlbGxlZCB3aXRob3V0IGNvbnNlcXVlbmNlcy5cclxuICAgICAqXHJcbiAgICAgKiBAZ3JvdXAgU3RhdGljIE1ldGhvZHNcclxuICAgICAqL1xyXG4gICAgc3RhdGljIHNsZWVwPFQ+KG1pbGxpc2Vjb25kczogbnVtYmVyLCB2YWx1ZTogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTxUPjtcclxuICAgIHN0YXRpYyBzbGVlcDxUID0gdm9pZD4obWlsbGlzZWNvbmRzOiBudW1iZXIsIHZhbHVlPzogVCk6IENhbmNlbGxhYmxlUHJvbWlzZTxUPiB7XHJcbiAgICAgICAgcmV0dXJuIG5ldyBDYW5jZWxsYWJsZVByb21pc2U8VD4oKHJlc29sdmUpID0+IHtcclxuICAgICAgICAgICAgc2V0VGltZW91dCgoKSA9PiByZXNvbHZlKHZhbHVlISksIG1pbGxpc2Vjb25kcyk7XHJcbiAgICAgICAgfSk7XHJcbiAgICB9XHJcblxyXG4gICAgLyoqXHJcbiAgICAgKiBDcmVhdGVzIGEgbmV3IHJlamVjdGVkIENhbmNlbGxhYmxlUHJvbWlzZSBmb3IgdGhlIHByb3ZpZGVkIHJlYXNvbi5cclxuICAgICAqXHJcbiAgICAgKiBAZ3JvdXAgU3RhdGljIE1ldGhvZHNcclxuICAgICAqL1xyXG4gICAgc3RhdGljIHJlamVjdDxUID0gbmV2ZXI+KHJlYXNvbj86IGFueSk6IENhbmNlbGxhYmxlUHJvbWlzZTxUPiB7XHJcbiAgICAgICAgcmV0dXJuIG5ldyBDYW5jZWxsYWJsZVByb21pc2U8VD4oKF8sIHJlamVjdCkgPT4gcmVqZWN0KHJlYXNvbikpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ3JlYXRlcyBhIG5ldyByZXNvbHZlZCBDYW5jZWxsYWJsZVByb21pc2UuXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyByZXNvbHZlKCk6IENhbmNlbGxhYmxlUHJvbWlzZTx2b2lkPjtcclxuICAgIC8qKlxyXG4gICAgICogQ3JlYXRlcyBhIG5ldyByZXNvbHZlZCBDYW5jZWxsYWJsZVByb21pc2UgZm9yIHRoZSBwcm92aWRlZCB2YWx1ZS5cclxuICAgICAqXHJcbiAgICAgKiBAZ3JvdXAgU3RhdGljIE1ldGhvZHNcclxuICAgICAqL1xyXG4gICAgc3RhdGljIHJlc29sdmU8VD4odmFsdWU6IFQpOiBDYW5jZWxsYWJsZVByb21pc2U8QXdhaXRlZDxUPj47XHJcbiAgICAvKipcclxuICAgICAqIENyZWF0ZXMgYSBuZXcgcmVzb2x2ZWQgQ2FuY2VsbGFibGVQcm9taXNlIGZvciB0aGUgcHJvdmlkZWQgdmFsdWUuXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyByZXNvbHZlPFQ+KHZhbHVlOiBUIHwgUHJvbWlzZUxpa2U8VD4pOiBDYW5jZWxsYWJsZVByb21pc2U8QXdhaXRlZDxUPj47XHJcbiAgICBzdGF0aWMgcmVzb2x2ZTxUID0gdm9pZD4odmFsdWU/OiBUIHwgUHJvbWlzZUxpa2U8VD4pOiBDYW5jZWxsYWJsZVByb21pc2U8QXdhaXRlZDxUPj4ge1xyXG4gICAgICAgIGlmICh2YWx1ZSBpbnN0YW5jZW9mIENhbmNlbGxhYmxlUHJvbWlzZSkge1xyXG4gICAgICAgICAgICAvLyBPcHRpbWlzZSBmb3IgY2FuY2VsbGFibGUgcHJvbWlzZXMuXHJcbiAgICAgICAgICAgIHJldHVybiB2YWx1ZTtcclxuICAgICAgICB9XHJcbiAgICAgICAgcmV0dXJuIG5ldyBDYW5jZWxsYWJsZVByb21pc2U8YW55PigocmVzb2x2ZSkgPT4gcmVzb2x2ZSh2YWx1ZSkpO1xyXG4gICAgfVxyXG5cclxuICAgIC8qKlxyXG4gICAgICogQ3JlYXRlcyBhIG5ldyBDYW5jZWxsYWJsZVByb21pc2UgYW5kIHJldHVybnMgaXQgaW4gYW4gb2JqZWN0LCBhbG9uZyB3aXRoIGl0cyByZXNvbHZlIGFuZCByZWplY3QgZnVuY3Rpb25zXHJcbiAgICAgKiBhbmQgYSBnZXR0ZXIvc2V0dGVyIGZvciB0aGUgY2FuY2VsbGF0aW9uIGNhbGxiYWNrLlxyXG4gICAgICpcclxuICAgICAqIFRoaXMgbWV0aG9kIGlzIHBvbHlmaWxsZWQsIGhlbmNlIGF2YWlsYWJsZSBpbiBldmVyeSBPUy93ZWJ2aWV3IHZlcnNpb24uXHJcbiAgICAgKlxyXG4gICAgICogQGdyb3VwIFN0YXRpYyBNZXRob2RzXHJcbiAgICAgKi9cclxuICAgIHN0YXRpYyB3aXRoUmVzb2x2ZXJzPFQ+KCk6IENhbmNlbGxhYmxlUHJvbWlzZVdpdGhSZXNvbHZlcnM8VD4ge1xyXG4gICAgICAgIGxldCByZXN1bHQ6IENhbmNlbGxhYmxlUHJvbWlzZVdpdGhSZXNvbHZlcnM8VD4gPSB7IG9uY2FuY2VsbGVkOiBudWxsIH0gYXMgYW55O1xyXG4gICAgICAgIHJlc3VsdC5wcm9taXNlID0gbmV3IENhbmNlbGxhYmxlUHJvbWlzZTxUPigocmVzb2x2ZSwgcmVqZWN0KSA9PiB7XHJcbiAgICAgICAgICAgIHJlc3VsdC5yZXNvbHZlID0gcmVzb2x2ZTtcclxuICAgICAgICAgICAgcmVzdWx0LnJlamVjdCA9IHJlamVjdDtcclxuICAgICAgICB9LCAoY2F1c2U/OiBhbnkpID0+IHsgcmVzdWx0Lm9uY2FuY2VsbGVkPy4oY2F1c2UpOyB9KTtcclxuICAgICAgICByZXR1cm4gcmVzdWx0O1xyXG4gICAgfVxyXG59XHJcblxyXG4vKipcclxuICogUmV0dXJucyBhIGNhbGxiYWNrIHRoYXQgaW1wbGVtZW50cyB0aGUgY2FuY2VsbGF0aW9uIGFsZ29yaXRobSBmb3IgdGhlIGdpdmVuIGNhbmNlbGxhYmxlIHByb21pc2UuXHJcbiAqIFRoZSBwcm9taXNlIHJldHVybmVkIGZyb20gdGhlIHJlc3VsdGluZyBmdW5jdGlvbiBkb2VzIG5vdCByZWplY3QuXHJcbiAqL1xyXG5mdW5jdGlvbiBjYW5jZWxsZXJGb3I8VD4ocHJvbWlzZTogQ2FuY2VsbGFibGVQcm9taXNlV2l0aFJlc29sdmVyczxUPiwgc3RhdGU6IENhbmNlbGxhYmxlUHJvbWlzZVN0YXRlKSB7XHJcbiAgICBsZXQgY2FuY2VsbGF0aW9uUHJvbWlzZTogdm9pZCB8IFByb21pc2VMaWtlPHZvaWQ+ID0gdW5kZWZpbmVkO1xyXG5cclxuICAgIHJldHVybiAocmVhc29uOiBDYW5jZWxFcnJvcik6IHZvaWQgfCBQcm9taXNlTGlrZTx2b2lkPiA9PiB7XHJcbiAgICAgICAgaWYgKCFzdGF0ZS5zZXR0bGVkKSB7XHJcbiAgICAgICAgICAgIHN0YXRlLnNldHRsZWQgPSB0cnVlO1xyXG4gICAgICAgICAgICBzdGF0ZS5yZWFzb24gPSByZWFzb247XHJcbiAgICAgICAgICAgIHByb21pc2UucmVqZWN0KHJlYXNvbik7XHJcblxyXG4gICAgICAgICAgICAvLyBBdHRhY2ggYW4gZXJyb3IgaGFuZGxlciB0aGF0IGlnbm9yZXMgdGhpcyBzcGVjaWZpYyByZWplY3Rpb24gcmVhc29uIGFuZCBub3RoaW5nIGVsc2UuXHJcbiAgICAgICAgICAgIC8vIEluIHRoZW9yeSwgYSBzYW5lIHVuZGVybHlpbmcgaW1wbGVtZW50YXRpb24gYXQgdGhpcyBwb2ludFxyXG4gICAgICAgICAgICAvLyBzaG91bGQgYWx3YXlzIHJlamVjdCB3aXRoIG91ciBjYW5jZWxsYXRpb24gcmVhc29uLFxyXG4gICAgICAgICAgICAvLyBoZW5jZSB0aGUgaGFuZGxlciB3aWxsIG5ldmVyIHRocm93LlxyXG4gICAgICAgICAgICB2b2lkIFByb21pc2UucHJvdG90eXBlLnRoZW4uY2FsbChwcm9taXNlLnByb21pc2UsIHVuZGVmaW5lZCwgKGVycikgPT4ge1xyXG4gICAgICAgICAgICAgICAgaWYgKGVyciAhPT0gcmVhc29uKSB7XHJcbiAgICAgICAgICAgICAgICAgICAgdGhyb3cgZXJyO1xyXG4gICAgICAgICAgICAgICAgfVxyXG4gICAgICAgICAgICB9KTtcclxuICAgICAgICB9XHJcblxyXG4gICAgICAgIC8vIElmIHJlYXNvbiBpcyBub3Qgc2V0LCB0aGUgcHJvbWlzZSByZXNvbHZlZCByZWd1bGFybHksIGhlbmNlIHdlIG11c3Qgbm90IGNhbGwgb25jYW5jZWxsZWQuXHJcbiAgICAgICAgLy8gSWYgb25jYW5jZWxsZWQgaXMgdW5zZXQsIG5vIG5lZWQgdG8gZ28gYW55IGZ1cnRoZXIuXHJcbiAgICAgICAgaWYgKCFzdGF0ZS5yZWFzb24gfHwgIXByb21pc2Uub25jYW5jZWxsZWQpIHsgcmV0dXJuOyB9XHJcblxyXG4gICAgICAgIGNhbmNlbGxhdGlvblByb21pc2UgPSBuZXcgUHJvbWlzZTx2b2lkPigocmVzb2x2ZSkgPT4ge1xyXG4gICAgICAgICAgICB0cnkge1xyXG4gICAgICAgICAgICAgICAgcmVzb2x2ZShwcm9taXNlLm9uY2FuY2VsbGVkIShzdGF0ZS5yZWFzb24hLmNhdXNlKSk7XHJcbiAgICAgICAgICAgIH0gY2F0Y2ggKGVycikge1xyXG4gICAgICAgICAgICAgICAgUHJvbWlzZS5yZWplY3QobmV3IENhbmNlbGxlZFJlamVjdGlvbkVycm9yKHByb21pc2UucHJvbWlzZSwgZXJyLCBcIlVuaGFuZGxlZCBleGNlcHRpb24gaW4gb25jYW5jZWxsZWQgY2FsbGJhY2suXCIpKTtcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgIH0pLmNhdGNoKChyZWFzb24/KSA9PiB7XHJcbiAgICAgICAgICAgIFByb21pc2UucmVqZWN0KG5ldyBDYW5jZWxsZWRSZWplY3Rpb25FcnJvcihwcm9taXNlLnByb21pc2UsIHJlYXNvbiwgXCJVbmhhbmRsZWQgcmVqZWN0aW9uIGluIG9uY2FuY2VsbGVkIGNhbGxiYWNrLlwiKSk7XHJcbiAgICAgICAgfSk7XHJcblxyXG4gICAgICAgIC8vIFVuc2V0IG9uY2FuY2VsbGVkIHRvIHByZXZlbnQgcmVwZWF0ZWQgY2FsbHMuXHJcbiAgICAgICAgcHJvbWlzZS5vbmNhbmNlbGxlZCA9IG51bGw7XHJcblxyXG4gICAgICAgIHJldHVybiBjYW5jZWxsYXRpb25Qcm9taXNlO1xyXG4gICAgfVxyXG59XHJcblxyXG4vKipcclxuICogUmV0dXJucyBhIGNhbGxiYWNrIHRoYXQgaW1wbGVtZW50cyB0aGUgcmVzb2x1dGlvbiBhbGdvcml0aG0gZm9yIHRoZSBnaXZlbiBjYW5jZWxsYWJsZSBwcm9taXNlLlxyXG4gKi9cclxuZnVuY3Rpb24gcmVzb2x2ZXJGb3I8VD4ocHJvbWlzZTogQ2FuY2VsbGFibGVQcm9taXNlV2l0aFJlc29sdmVyczxUPiwgc3RhdGU6IENhbmNlbGxhYmxlUHJvbWlzZVN0YXRlKTogQ2FuY2VsbGFibGVQcm9taXNlUmVzb2x2ZXI8VD4ge1xyXG4gICAgcmV0dXJuICh2YWx1ZSkgPT4ge1xyXG4gICAgICAgIGlmIChzdGF0ZS5yZXNvbHZpbmcpIHsgcmV0dXJuOyB9XHJcbiAgICAgICAgc3RhdGUucmVzb2x2aW5nID0gdHJ1ZTtcclxuXHJcbiAgICAgICAgaWYgKHZhbHVlID09PSBwcm9taXNlLnByb21pc2UpIHtcclxuICAgICAgICAgICAgaWYgKHN0YXRlLnNldHRsZWQpIHsgcmV0dXJuOyB9XHJcbiAgICAgICAgICAgIHN0YXRlLnNldHRsZWQgPSB0cnVlO1xyXG4gICAgICAgICAgICBwcm9taXNlLnJlamVjdChuZXcgVHlwZUVycm9yKFwiQSBwcm9taXNlIGNhbm5vdCBiZSByZXNvbHZlZCB3aXRoIGl0c2VsZi5cIikpO1xyXG4gICAgICAgICAgICByZXR1cm47XHJcbiAgICAgICAgfVxyXG5cclxuICAgICAgICBpZiAodmFsdWUgIT0gbnVsbCAmJiAodHlwZW9mIHZhbHVlID09PSAnb2JqZWN0JyB8fCB0eXBlb2YgdmFsdWUgPT09ICdmdW5jdGlvbicpKSB7XHJcbiAgICAgICAgICAgIGxldCB0aGVuOiBhbnk7XHJcbiAgICAgICAgICAgIHRyeSB7XHJcbiAgICAgICAgICAgICAgICB0aGVuID0gKHZhbHVlIGFzIGFueSkudGhlbjtcclxuICAgICAgICAgICAgfSBjYXRjaCAoZXJyKSB7XHJcbiAgICAgICAgICAgICAgICBzdGF0ZS5zZXR0bGVkID0gdHJ1ZTtcclxuICAgICAgICAgICAgICAgIHByb21pc2UucmVqZWN0KGVycik7XHJcbiAgICAgICAgICAgICAgICByZXR1cm47XHJcbiAgICAgICAgICAgIH1cclxuXHJcbiAgICAgICAgICAgIGlmIChpc0NhbGxhYmxlKHRoZW4pKSB7XHJcbiAgICAgICAgICAgICAgICB0cnkge1xyXG4gICAgICAgICAgICAgICAgICAgIGxldCBjYW5jZWwgPSAodmFsdWUgYXMgYW55KS5jYW5jZWw7XHJcbiAgICAgICAgICAgICAgICAgICAgaWYgKGlzQ2FsbGFibGUoY2FuY2VsKSkge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICBjb25zdCBvbmNhbmNlbGxlZCA9IChjYXVzZT86IGFueSkgPT4ge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgUmVmbGVjdC5hcHBseShjYW5jZWwsIHZhbHVlLCBbY2F1c2VdKTtcclxuICAgICAgICAgICAgICAgICAgICAgICAgfTtcclxuICAgICAgICAgICAgICAgICAgICAgICAgaWYgKHN0YXRlLnJlYXNvbikge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgLy8gSWYgYWxyZWFkeSBjYW5jZWxsZWQsIHByb3BhZ2F0ZSBjYW5jZWxsYXRpb24uXHJcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICAvLyBUaGUgcHJvbWlzZSByZXR1cm5lZCBmcm9tIHRoZSBjYW5jZWxsZXIgYWxnb3JpdGhtIGRvZXMgbm90IHJlamVjdFxyXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgLy8gc28gaXQgY2FuIGJlIGRpc2NhcmRlZCBzYWZlbHkuXHJcbiAgICAgICAgICAgICAgICAgICAgICAgICAgICB2b2lkIGNhbmNlbGxlckZvcih7IC4uLnByb21pc2UsIG9uY2FuY2VsbGVkIH0sIHN0YXRlKShzdGF0ZS5yZWFzb24pO1xyXG4gICAgICAgICAgICAgICAgICAgICAgICB9IGVsc2Uge1xyXG4gICAgICAgICAgICAgICAgICAgICAgICAgICAgcHJvbWlzZS5vbmNhbmNlbGxlZCA9IG9uY2FuY2VsbGVkO1xyXG4gICAgICAgICAgICAgICAgICAgICAgICB9XHJcbiAgICAgICAgICAgICAgICAgICAgfVxyXG4gICAgICAgICAgICAgICAgfSBjYXRjaCB7fVxyXG5cclxuICAgICAgICAgICAgICAgIGNvbnN0IG5ld1N0YXRlOiBDYW5jZWxsYWJsZVByb21pc2VTdGF0ZSA9IHtcclxuICAgICAgICAgICAgICAgICAgICByb290OiBzdGF0ZS5yb290LFxyXG4gICAgICAgICAgICAgICAgICAgIHJlc29sdmluZzogZmFsc2UsXHJcbiAgICAgICAgICAgICAgICAgICAgZ2V0IHNldHRsZWQoKSB7IHJldHVybiB0aGlzLnJvb3Quc2V0dGxlZCB9LFxyXG4gICAgICAgICAgICAgICAgICAgIHNldCBzZXR0bGVkKHZhbHVlKSB7IHRoaXMucm9vdC5zZXR0bGVkID0gdmFsdWU7IH0sXHJcbiAgICAgICAgICAgICAgICAgICAgZ2V0IHJlYXNvbigpIHsgcmV0dXJuIHRoaXMucm9vdC5yZWFzb24gfVxyXG4gICAgICAgICAgICAgICAgfTtcclxuXHJcbiAgICAgICAgICAgICAgICBjb25zdCByZWplY3RvciA9IHJlamVjdG9yRm9yKHByb21pc2UsIG5ld1N0YXRlKTtcclxuICAgICAgICAgICAgICAgIHRyeSB7XHJcbiAgICAgICAgICAgICAgICAgICAgUmVmbGVjdC5hcHBseSh0aGVuLCB2YWx1ZSwgW3Jlc29sdmVyRm9yKHByb21pc2UsIG5ld1N0YXRlKSwgcmVqZWN0b3JdKTtcclxuICAgICAgICAgICAgICAgIH0gY2F0Y2ggKGVycikge1xyXG4gICAgICAgICAgICAgICAgICAgIHJlamVjdG9yKGVycik7XHJcbiAgICAgICAgICAgICAgICB9XHJcbiAgICAgICAgICAgICAgICByZXR1cm47IC8vIElNUE9SVEFOVCFcclxuICAgICAgICAgICAgfVxyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgaWYgKHN0YXRlLnNldHRsZWQpIHsgcmV0dXJuOyB9XHJcbiAgICAgICAgc3RhdGUuc2V0dGxlZCA9IHRydWU7XHJcbiAgICAgICAgcHJvbWlzZS5yZXNvbHZlKHZhbHVlKTtcclxuICAgIH07XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBSZXR1cm5zIGEgY2FsbGJhY2sgdGhhdCBpbXBsZW1lbnRzIHRoZSByZWplY3Rpb24gYWxnb3JpdGhtIGZvciB0aGUgZ2l2ZW4gY2FuY2VsbGFibGUgcHJvbWlzZS5cclxuICovXHJcbmZ1bmN0aW9uIHJlamVjdG9yRm9yPFQ+KHByb21pc2U6IENhbmNlbGxhYmxlUHJvbWlzZVdpdGhSZXNvbHZlcnM8VD4sIHN0YXRlOiBDYW5jZWxsYWJsZVByb21pc2VTdGF0ZSk6IENhbmNlbGxhYmxlUHJvbWlzZVJlamVjdG9yIHtcclxuICAgIHJldHVybiAocmVhc29uPykgPT4ge1xyXG4gICAgICAgIGlmIChzdGF0ZS5yZXNvbHZpbmcpIHsgcmV0dXJuOyB9XHJcbiAgICAgICAgc3RhdGUucmVzb2x2aW5nID0gdHJ1ZTtcclxuXHJcbiAgICAgICAgaWYgKHN0YXRlLnNldHRsZWQpIHtcclxuICAgICAgICAgICAgdHJ5IHtcclxuICAgICAgICAgICAgICAgIGlmIChyZWFzb24gaW5zdGFuY2VvZiBDYW5jZWxFcnJvciAmJiBzdGF0ZS5yZWFzb24gaW5zdGFuY2VvZiBDYW5jZWxFcnJvciAmJiBPYmplY3QuaXMocmVhc29uLmNhdXNlLCBzdGF0ZS5yZWFzb24uY2F1c2UpKSB7XHJcbiAgICAgICAgICAgICAgICAgICAgLy8gU3dhbGxvdyBsYXRlIHJlamVjdGlvbnMgdGhhdCBhcmUgQ2FuY2VsRXJyb3JzIHdob3NlIGNhbmNlbGxhdGlvbiBjYXVzZSBpcyB0aGUgc2FtZSBhcyBvdXJzLlxyXG4gICAgICAgICAgICAgICAgICAgIHJldHVybjtcclxuICAgICAgICAgICAgICAgIH1cclxuICAgICAgICAgICAgfSBjYXRjaCB7fVxyXG5cclxuICAgICAgICAgICAgdm9pZCBQcm9taXNlLnJlamVjdChuZXcgQ2FuY2VsbGVkUmVqZWN0aW9uRXJyb3IocHJvbWlzZS5wcm9taXNlLCByZWFzb24pKTtcclxuICAgICAgICB9IGVsc2Uge1xyXG4gICAgICAgICAgICBzdGF0ZS5zZXR0bGVkID0gdHJ1ZTtcclxuICAgICAgICAgICAgcHJvbWlzZS5yZWplY3QocmVhc29uKTtcclxuICAgICAgICB9XHJcbiAgICB9XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBDYW5jZWxzIGFsbCB2YWx1ZXMgaW4gYW4gYXJyYXkgdGhhdCBsb29rIGxpa2UgY2FuY2VsbGFibGUgdGhlbmFibGVzLlxyXG4gKiBSZXR1cm5zIGEgcHJvbWlzZSB0aGF0IGZ1bGZpbGxzIG9uY2UgYWxsIGNhbmNlbGxhdGlvbiBwcm9jZWR1cmVzIGZvciB0aGUgZ2l2ZW4gdmFsdWVzIGhhdmUgc2V0dGxlZC5cclxuICovXHJcbmZ1bmN0aW9uIGNhbmNlbEFsbChwYXJlbnQ6IENhbmNlbGxhYmxlUHJvbWlzZTx1bmtub3duPiwgdmFsdWVzOiBhbnlbXSwgY2F1c2U/OiBhbnkpOiBQcm9taXNlPHZvaWQ+IHtcclxuICAgIGNvbnN0IHJlc3VsdHM6IFByb21pc2U8dm9pZD5bXSA9IFtdO1xyXG5cclxuICAgIGZvciAoY29uc3QgdmFsdWUgb2YgdmFsdWVzKSB7XHJcbiAgICAgICAgbGV0IGNhbmNlbDogQ2FuY2VsbGFibGVQcm9taXNlQ2FuY2VsbGVyO1xyXG4gICAgICAgIHRyeSB7XHJcbiAgICAgICAgICAgIGlmICghaXNDYWxsYWJsZSh2YWx1ZS50aGVuKSkgeyBjb250aW51ZTsgfVxyXG4gICAgICAgICAgICBjYW5jZWwgPSB2YWx1ZS5jYW5jZWw7XHJcbiAgICAgICAgICAgIGlmICghaXNDYWxsYWJsZShjYW5jZWwpKSB7IGNvbnRpbnVlOyB9XHJcbiAgICAgICAgfSBjYXRjaCB7IGNvbnRpbnVlOyB9XHJcblxyXG4gICAgICAgIGxldCByZXN1bHQ6IHZvaWQgfCBQcm9taXNlTGlrZTx2b2lkPjtcclxuICAgICAgICB0cnkge1xyXG4gICAgICAgICAgICByZXN1bHQgPSBSZWZsZWN0LmFwcGx5KGNhbmNlbCwgdmFsdWUsIFtjYXVzZV0pO1xyXG4gICAgICAgIH0gY2F0Y2ggKGVycikge1xyXG4gICAgICAgICAgICBQcm9taXNlLnJlamVjdChuZXcgQ2FuY2VsbGVkUmVqZWN0aW9uRXJyb3IocGFyZW50LCBlcnIsIFwiVW5oYW5kbGVkIGV4Y2VwdGlvbiBpbiBjYW5jZWwgbWV0aG9kLlwiKSk7XHJcbiAgICAgICAgICAgIGNvbnRpbnVlO1xyXG4gICAgICAgIH1cclxuXHJcbiAgICAgICAgaWYgKCFyZXN1bHQpIHsgY29udGludWU7IH1cclxuICAgICAgICByZXN1bHRzLnB1c2goXHJcbiAgICAgICAgICAgIChyZXN1bHQgaW5zdGFuY2VvZiBQcm9taXNlICA/IHJlc3VsdCA6IFByb21pc2UucmVzb2x2ZShyZXN1bHQpKS5jYXRjaCgocmVhc29uPykgPT4ge1xyXG4gICAgICAgICAgICAgICAgUHJvbWlzZS5yZWplY3QobmV3IENhbmNlbGxlZFJlamVjdGlvbkVycm9yKHBhcmVudCwgcmVhc29uLCBcIlVuaGFuZGxlZCByZWplY3Rpb24gaW4gY2FuY2VsIG1ldGhvZC5cIikpO1xyXG4gICAgICAgICAgICB9KVxyXG4gICAgICAgICk7XHJcbiAgICB9XHJcblxyXG4gICAgcmV0dXJuIFByb21pc2UuYWxsKHJlc3VsdHMpIGFzIGFueTtcclxufVxyXG5cclxuLyoqXHJcbiAqIFJldHVybnMgaXRzIGFyZ3VtZW50LlxyXG4gKi9cclxuZnVuY3Rpb24gaWRlbnRpdHk8VD4oeDogVCk6IFQge1xyXG4gICAgcmV0dXJuIHg7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBUaHJvd3MgaXRzIGFyZ3VtZW50LlxyXG4gKi9cclxuZnVuY3Rpb24gdGhyb3dlcihyZWFzb24/OiBhbnkpOiBuZXZlciB7XHJcbiAgICB0aHJvdyByZWFzb247XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBBdHRlbXB0cyB2YXJpb3VzIHN0cmF0ZWdpZXMgdG8gY29udmVydCBhbiBlcnJvciB0byBhIHN0cmluZy5cclxuICovXHJcbmZ1bmN0aW9uIGVycm9yTWVzc2FnZShlcnI6IGFueSk6IHN0cmluZyB7XHJcbiAgICB0cnkge1xyXG4gICAgICAgIGlmIChlcnIgaW5zdGFuY2VvZiBFcnJvciB8fCB0eXBlb2YgZXJyICE9PSAnb2JqZWN0JyB8fCBlcnIudG9TdHJpbmcgIT09IE9iamVjdC5wcm90b3R5cGUudG9TdHJpbmcpIHtcclxuICAgICAgICAgICAgcmV0dXJuIFwiXCIgKyBlcnI7XHJcbiAgICAgICAgfVxyXG4gICAgfSBjYXRjaCB7fVxyXG5cclxuICAgIHRyeSB7XHJcbiAgICAgICAgcmV0dXJuIEpTT04uc3RyaW5naWZ5KGVycik7XHJcbiAgICB9IGNhdGNoIHt9XHJcblxyXG4gICAgdHJ5IHtcclxuICAgICAgICByZXR1cm4gT2JqZWN0LnByb3RvdHlwZS50b1N0cmluZy5jYWxsKGVycik7XHJcbiAgICB9IGNhdGNoIHt9XHJcblxyXG4gICAgcmV0dXJuIFwiPGNvdWxkIG5vdCBjb252ZXJ0IGVycm9yIHRvIHN0cmluZz5cIjtcclxufVxyXG5cclxuLyoqXHJcbiAqIEdldHMgdGhlIGN1cnJlbnQgYmFycmllciBwcm9taXNlIGZvciB0aGUgZ2l2ZW4gY2FuY2VsbGFibGUgcHJvbWlzZS4gSWYgbmVjZXNzYXJ5LCBpbml0aWFsaXNlcyB0aGUgYmFycmllci5cclxuICovXHJcbmZ1bmN0aW9uIGN1cnJlbnRCYXJyaWVyPFQ+KHByb21pc2U6IENhbmNlbGxhYmxlUHJvbWlzZTxUPik6IFByb21pc2U8dm9pZD4ge1xyXG4gICAgbGV0IHB3cjogUGFydGlhbDxQcm9taXNlV2l0aFJlc29sdmVyczx2b2lkPj4gPSBwcm9taXNlW2JhcnJpZXJTeW1dID8/IHt9O1xyXG4gICAgaWYgKCEoJ3Byb21pc2UnIGluIHB3cikpIHtcclxuICAgICAgICBPYmplY3QuYXNzaWduKHB3ciwgcHJvbWlzZVdpdGhSZXNvbHZlcnM8dm9pZD4oKSk7XHJcbiAgICB9XHJcbiAgICBpZiAocHJvbWlzZVtiYXJyaWVyU3ltXSA9PSBudWxsKSB7XHJcbiAgICAgICAgcHdyLnJlc29sdmUhKCk7XHJcbiAgICAgICAgcHJvbWlzZVtiYXJyaWVyU3ltXSA9IHB3cjtcclxuICAgIH1cclxuICAgIHJldHVybiBwd3IucHJvbWlzZSE7XHJcbn1cclxuXHJcbi8vIFBvbHlmaWxsIFByb21pc2Uud2l0aFJlc29sdmVycy5cclxubGV0IHByb21pc2VXaXRoUmVzb2x2ZXJzID0gUHJvbWlzZS53aXRoUmVzb2x2ZXJzO1xyXG5pZiAocHJvbWlzZVdpdGhSZXNvbHZlcnMgJiYgdHlwZW9mIHByb21pc2VXaXRoUmVzb2x2ZXJzID09PSAnZnVuY3Rpb24nKSB7XHJcbiAgICBwcm9taXNlV2l0aFJlc29sdmVycyA9IHByb21pc2VXaXRoUmVzb2x2ZXJzLmJpbmQoUHJvbWlzZSk7XHJcbn0gZWxzZSB7XHJcbiAgICBwcm9taXNlV2l0aFJlc29sdmVycyA9IGZ1bmN0aW9uIDxUPigpOiBQcm9taXNlV2l0aFJlc29sdmVyczxUPiB7XHJcbiAgICAgICAgbGV0IHJlc29sdmUhOiAodmFsdWU6IFQgfCBQcm9taXNlTGlrZTxUPikgPT4gdm9pZDtcclxuICAgICAgICBsZXQgcmVqZWN0ITogKHJlYXNvbj86IGFueSkgPT4gdm9pZDtcclxuICAgICAgICBjb25zdCBwcm9taXNlID0gbmV3IFByb21pc2U8VD4oKHJlcywgcmVqKSA9PiB7IHJlc29sdmUgPSByZXM7IHJlamVjdCA9IHJlajsgfSk7XHJcbiAgICAgICAgcmV0dXJuIHsgcHJvbWlzZSwgcmVzb2x2ZSwgcmVqZWN0IH07XHJcbiAgICB9XHJcbn1cclxuIiwgIi8qXHJcbiBfXHQgICBfX1x0ICBfIF9fXHJcbnwgfFx0IC8gL19fXyBfKF8pIC9fX19fXHJcbnwgfCAvfCAvIC8gX18gYC8gLyAvIF9fXy9cclxufCB8LyB8LyAvIC9fLyAvIC8gKF9fICApXHJcbnxfXy98X18vXFxfXyxfL18vXy9fX19fL1xyXG5UaGUgZWxlY3Ryb24gYWx0ZXJuYXRpdmUgZm9yIEdvXHJcbihjKSBMZWEgQW50aG9ueSAyMDE5LXByZXNlbnRcclxuKi9cclxuXHJcbmltcG9ydCB7bmV3UnVudGltZUNhbGxlciwgb2JqZWN0TmFtZXN9IGZyb20gXCIuL3J1bnRpbWUuanNcIjtcclxuXHJcbmNvbnN0IGNhbGwgPSBuZXdSdW50aW1lQ2FsbGVyKG9iamVjdE5hbWVzLkNsaXBib2FyZCk7XHJcblxyXG5jb25zdCBDbGlwYm9hcmRTZXRUZXh0ID0gMDtcclxuY29uc3QgQ2xpcGJvYXJkVGV4dCA9IDE7XHJcblxyXG4vKipcclxuICogU2V0cyB0aGUgdGV4dCB0byB0aGUgQ2xpcGJvYXJkLlxyXG4gKlxyXG4gKiBAcGFyYW0gdGV4dCAtIFRoZSB0ZXh0IHRvIGJlIHNldCB0byB0aGUgQ2xpcGJvYXJkLlxyXG4gKiBAcmV0dXJuIEEgUHJvbWlzZSB0aGF0IHJlc29sdmVzIHdoZW4gdGhlIG9wZXJhdGlvbiBpcyBzdWNjZXNzZnVsLlxyXG4gKi9cclxuZXhwb3J0IGZ1bmN0aW9uIFNldFRleHQodGV4dDogc3RyaW5nKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICByZXR1cm4gY2FsbChDbGlwYm9hcmRTZXRUZXh0LCB7dGV4dH0pO1xyXG59XHJcblxyXG4vKipcclxuICogR2V0IHRoZSBDbGlwYm9hcmQgdGV4dFxyXG4gKlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCByZXNvbHZlcyB3aXRoIHRoZSB0ZXh0IGZyb20gdGhlIENsaXBib2FyZC5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBUZXh0KCk6IFByb21pc2U8c3RyaW5nPiB7XHJcbiAgICByZXR1cm4gY2FsbChDbGlwYm9hcmRUZXh0KTtcclxufVxyXG4iLCAiLypcclxuIF9cdCAgIF9fXHQgIF8gX19cclxufCB8XHQgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuZXhwb3J0IGludGVyZmFjZSBTaXplIHtcclxuICAgIC8qKiBUaGUgd2lkdGggb2YgYSByZWN0YW5ndWxhciBhcmVhLiAqL1xyXG4gICAgV2lkdGg6IG51bWJlcjtcclxuICAgIC8qKiBUaGUgaGVpZ2h0IG9mIGEgcmVjdGFuZ3VsYXIgYXJlYS4gKi9cclxuICAgIEhlaWdodDogbnVtYmVyO1xyXG59XHJcblxyXG5leHBvcnQgaW50ZXJmYWNlIFJlY3Qge1xyXG4gICAgLyoqIFRoZSBYIGNvb3JkaW5hdGUgb2YgdGhlIG9yaWdpbi4gKi9cclxuICAgIFg6IG51bWJlcjtcclxuICAgIC8qKiBUaGUgWSBjb29yZGluYXRlIG9mIHRoZSBvcmlnaW4uICovXHJcbiAgICBZOiBudW1iZXI7XHJcbiAgICAvKiogVGhlIHdpZHRoIG9mIHRoZSByZWN0YW5nbGUuICovXHJcbiAgICBXaWR0aDogbnVtYmVyO1xyXG4gICAgLyoqIFRoZSBoZWlnaHQgb2YgdGhlIHJlY3RhbmdsZS4gKi9cclxuICAgIEhlaWdodDogbnVtYmVyO1xyXG59XHJcblxyXG5leHBvcnQgaW50ZXJmYWNlIFNjcmVlbiB7XHJcbiAgICAvKiogVW5pcXVlIGlkZW50aWZpZXIgZm9yIHRoZSBzY3JlZW4uICovXHJcbiAgICBJRDogc3RyaW5nO1xyXG4gICAgLyoqIEh1bWFuLXJlYWRhYmxlIG5hbWUgb2YgdGhlIHNjcmVlbi4gKi9cclxuICAgIE5hbWU6IHN0cmluZztcclxuICAgIC8qKiBUaGUgc2NhbGUgZmFjdG9yIG9mIHRoZSBzY3JlZW4gKERQSS85NikuIDEgPSBzdGFuZGFyZCBEUEksIDIgPSBIaURQSSAoUmV0aW5hKSwgZXRjLiAqL1xyXG4gICAgU2NhbGVGYWN0b3I6IG51bWJlcjtcclxuICAgIC8qKiBUaGUgWCBjb29yZGluYXRlIG9mIHRoZSBzY3JlZW4uICovXHJcbiAgICBYOiBudW1iZXI7XHJcbiAgICAvKiogVGhlIFkgY29vcmRpbmF0ZSBvZiB0aGUgc2NyZWVuLiAqL1xyXG4gICAgWTogbnVtYmVyO1xyXG4gICAgLyoqIENvbnRhaW5zIHRoZSB3aWR0aCBhbmQgaGVpZ2h0IG9mIHRoZSBzY3JlZW4uICovXHJcbiAgICBTaXplOiBTaXplO1xyXG4gICAgLyoqIENvbnRhaW5zIHRoZSBib3VuZHMgb2YgdGhlIHNjcmVlbiBpbiB0ZXJtcyBvZiBYLCBZLCBXaWR0aCwgYW5kIEhlaWdodC4gKi9cclxuICAgIEJvdW5kczogUmVjdDtcclxuICAgIC8qKiBDb250YWlucyB0aGUgcGh5c2ljYWwgYm91bmRzIG9mIHRoZSBzY3JlZW4gaW4gdGVybXMgb2YgWCwgWSwgV2lkdGgsIGFuZCBIZWlnaHQgKGJlZm9yZSBzY2FsaW5nKS4gKi9cclxuICAgIFBoeXNpY2FsQm91bmRzOiBSZWN0O1xyXG4gICAgLyoqIENvbnRhaW5zIHRoZSBhcmVhIG9mIHRoZSBzY3JlZW4gdGhhdCBpcyBhY3R1YWxseSB1c2FibGUgKGV4Y2x1ZGluZyB0YXNrYmFyIGFuZCBvdGhlciBzeXN0ZW0gVUkpLiAqL1xyXG4gICAgV29ya0FyZWE6IFJlY3Q7XHJcbiAgICAvKiogQ29udGFpbnMgdGhlIHBoeXNpY2FsIFdvcmtBcmVhIG9mIHRoZSBzY3JlZW4gKGJlZm9yZSBzY2FsaW5nKS4gKi9cclxuICAgIFBoeXNpY2FsV29ya0FyZWE6IFJlY3Q7XHJcbiAgICAvKiogVHJ1ZSBpZiB0aGlzIGlzIHRoZSBwcmltYXJ5IG1vbml0b3Igc2VsZWN0ZWQgYnkgdGhlIHVzZXIgaW4gdGhlIG9wZXJhdGluZyBzeXN0ZW0uICovXHJcbiAgICBJc1ByaW1hcnk6IGJvb2xlYW47XHJcbiAgICAvKiogVGhlIHJvdGF0aW9uIG9mIHRoZSBzY3JlZW4uICovXHJcbiAgICBSb3RhdGlvbjogbnVtYmVyO1xyXG59XHJcblxyXG5pbXBvcnQgeyBuZXdSdW50aW1lQ2FsbGVyLCBvYmplY3ROYW1lcyB9IGZyb20gXCIuL3J1bnRpbWUuanNcIjtcclxuY29uc3QgY2FsbCA9IG5ld1J1bnRpbWVDYWxsZXIob2JqZWN0TmFtZXMuU2NyZWVucyk7XHJcblxyXG5jb25zdCBnZXRBbGwgPSAwO1xyXG5jb25zdCBnZXRQcmltYXJ5ID0gMTtcclxuY29uc3QgZ2V0Q3VycmVudCA9IDI7XHJcbmNvbnN0IGdldEJ5SUQgPSAzO1xyXG5jb25zdCBnZXRCeUluZGV4ID0gNDtcclxuXHJcbi8qKlxyXG4gKiBHZXRzIGFsbCBzY3JlZW5zLlxyXG4gKlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCByZXNvbHZlcyB0byBhbiBhcnJheSBvZiBTY3JlZW4gb2JqZWN0cy5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBHZXRBbGwoKTogUHJvbWlzZTxTY3JlZW5bXT4ge1xyXG4gICAgcmV0dXJuIGNhbGwoZ2V0QWxsKTtcclxufVxyXG5cclxuLyoqXHJcbiAqIEdldHMgdGhlIHByaW1hcnkgc2NyZWVuLlxyXG4gKlxyXG4gKiBAcmV0dXJucyBBIHByb21pc2UgdGhhdCByZXNvbHZlcyB0byB0aGUgcHJpbWFyeSBzY3JlZW4uXHJcbiAqL1xyXG5leHBvcnQgZnVuY3Rpb24gR2V0UHJpbWFyeSgpOiBQcm9taXNlPFNjcmVlbj4ge1xyXG4gICAgcmV0dXJuIGNhbGwoZ2V0UHJpbWFyeSk7XHJcbn1cclxuXHJcbi8qKlxyXG4gKiBHZXRzIHRoZSBjdXJyZW50IGFjdGl2ZSBzY3JlZW4uXHJcbiAqXHJcbiAqIEByZXR1cm5zIEEgcHJvbWlzZSB0aGF0IHJlc29sdmVzIHdpdGggdGhlIGN1cnJlbnQgYWN0aXZlIHNjcmVlbi5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBHZXRDdXJyZW50KCk6IFByb21pc2U8U2NyZWVuPiB7XHJcbiAgICByZXR1cm4gY2FsbChnZXRDdXJyZW50KTtcclxufVxyXG5cclxuLyoqXHJcbiAqIEdldHMgYSBzY3JlZW4gYnkgaXRzIHVuaXF1ZSBkaXNwbGF5IElELlxyXG4gKlxyXG4gKiBAcGFyYW0gaWQgLSBUaGUgdW5pcXVlIGlkZW50aWZpZXIgb2YgdGhlIHNjcmVlbi5cclxuICogQHJldHVybnMgQSBwcm9taXNlIHRoYXQgcmVzb2x2ZXMgdG8gdGhlIG1hdGNoaW5nIFNjcmVlbi5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBHZXRCeUlEKGlkOiBzdHJpbmcpOiBQcm9taXNlPFNjcmVlbj4ge1xyXG4gICAgcmV0dXJuIGNhbGwoZ2V0QnlJRCwgeyBpZCB9KTtcclxufVxyXG5cclxuLyoqXHJcbiAqIEdldHMgYSBzY3JlZW4gYnkgaXRzIGluZGV4IGluIHRoZSBzY3JlZW4gbGlzdC5cclxuICpcclxuICogQHBhcmFtIGluZGV4IC0gVGhlIHplcm8tYmFzZWQgaW5kZXggb2YgdGhlIHNjcmVlbi5cclxuICogQHJldHVybnMgQSBwcm9taXNlIHRoYXQgcmVzb2x2ZXMgdG8gdGhlIG1hdGNoaW5nIFNjcmVlbi5cclxuICovXHJcbmV4cG9ydCBmdW5jdGlvbiBHZXRCeUluZGV4KGluZGV4OiBudW1iZXIpOiBQcm9taXNlPFNjcmVlbj4ge1xyXG4gICAgcmV0dXJuIGNhbGwoZ2V0QnlJbmRleCwgeyBpbmRleCB9KTtcclxufVxyXG4iLCAiLypcclxuIF8gICAgIF9fICAgICBfIF9fXHJcbnwgfCAgLyAvX19fIF8oXykgL19fX19cclxufCB8IC98IC8gLyBfXyBgLyAvIC8gX19fL1xyXG58IHwvIHwvIC8gL18vIC8gLyAoX18gIClcclxufF9fL3xfXy9cXF9fLF8vXy9fL19fX18vXHJcblRoZSBlbGVjdHJvbiBhbHRlcm5hdGl2ZSBmb3IgR29cclxuKGMpIExlYSBBbnRob255IDIwMTktcHJlc2VudFxyXG4qL1xyXG5cclxuaW1wb3J0IHsgbmV3UnVudGltZUNhbGxlciwgb2JqZWN0TmFtZXMgfSBmcm9tIFwiLi9ydW50aW1lLmpzXCI7XHJcblxyXG5jb25zdCBjYWxsID0gbmV3UnVudGltZUNhbGxlcihvYmplY3ROYW1lcy5JT1MpO1xyXG5cclxuLy8gTWV0aG9kIElEc1xyXG5jb25zdCBIYXB0aWNzSW1wYWN0ID0gMDtcclxuY29uc3QgRGV2aWNlSW5mbyA9IDE7XHJcblxyXG5leHBvcnQgbmFtZXNwYWNlIEhhcHRpY3Mge1xyXG4gICAgZXhwb3J0IHR5cGUgSW1wYWN0U3R5bGUgPSBcImxpZ2h0XCJ8XCJtZWRpdW1cInxcImhlYXZ5XCJ8XCJzb2Z0XCJ8XCJyaWdpZFwiO1xyXG4gICAgZXhwb3J0IGZ1bmN0aW9uIEltcGFjdChzdHlsZTogSW1wYWN0U3R5bGUgPSBcIm1lZGl1bVwiKTogUHJvbWlzZTx2b2lkPiB7XHJcbiAgICAgICAgcmV0dXJuIGNhbGwoSGFwdGljc0ltcGFjdCwgeyBzdHlsZSB9KTtcclxuICAgIH1cclxufVxyXG5cclxuZXhwb3J0IG5hbWVzcGFjZSBEZXZpY2Uge1xyXG4gICAgZXhwb3J0IGludGVyZmFjZSBJbmZvIHtcclxuICAgICAgICBtb2RlbDogc3RyaW5nO1xyXG4gICAgICAgIHN5c3RlbU5hbWU6IHN0cmluZztcclxuICAgICAgICBzeXN0ZW1WZXJzaW9uOiBzdHJpbmc7XHJcbiAgICAgICAgaXNTaW11bGF0b3I6IGJvb2xlYW47XHJcbiAgICB9XHJcbiAgICBleHBvcnQgZnVuY3Rpb24gSW5mbygpOiBQcm9taXNlPEluZm8+IHtcclxuICAgICAgICByZXR1cm4gY2FsbChEZXZpY2VJbmZvKTtcclxuICAgIH1cclxufVxyXG4iXSwKICAibWFwcGluZ3MiOiAiOzs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7Ozs7O0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBO0FBQUE7QUFBQTtBQUFBOzs7QUNBQTtBQUFBO0FBQUE7QUFBQTtBQUFBOzs7QUNBQTtBQUFBO0FBQUE7QUFBQTs7O0FDNkJBLElBQU0sY0FDRjtBQUVHLFNBQVMsT0FBTyxPQUFlLElBQVk7QUFDOUMsTUFBSSxLQUFLO0FBRVQsTUFBSSxJQUFJLE9BQU87QUFDZixTQUFPLEtBQUs7QUFFUixVQUFNLFlBQWEsS0FBSyxPQUFPLElBQUksS0FBTSxDQUFDO0FBQUEsRUFDOUM7QUFDQSxTQUFPO0FBQ1g7OztBQzdCQSxJQUFNLGFBQWEsT0FBTyxTQUFTLFNBQVM7QUFHNUMsSUFBTSxrQkFBa0IsTUFBTTtBQU12QixJQUFNLGNBQWMsT0FBTyxPQUFPO0FBQUEsRUFDckMsTUFBTTtBQUFBLEVBQ04sV0FBVztBQUFBLEVBQ1gsYUFBYTtBQUFBLEVBQ2IsUUFBUTtBQUFBLEVBQ1IsYUFBYTtBQUFBLEVBQ2IsUUFBUTtBQUFBLEVBQ1IsUUFBUTtBQUFBLEVBQ1IsU0FBUztBQUFBLEVBQ1QsUUFBUTtBQUFBLEVBQ1IsU0FBUztBQUFBLEVBQ1QsWUFBWTtBQUFBLEVBQ1osS0FBSztBQUNULENBQUM7QUFDTSxJQUFJLFdBQVcsT0FBTztBQXVCN0IsSUFBSSxrQkFBMkM7QUFzQnhDLFNBQVMsYUFBYSxXQUEwQztBQUNuRSxvQkFBa0I7QUFDdEI7QUFLTyxTQUFTLGVBQXdDO0FB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QUEsSUFDMUMsMkJBQTJCO0FBQUEsSUFDM0IsK0JBQStCO0FBQUEsSUFDL0Isb0JBQW9CO0FBQUEsSUFDcEIsNEJBQTRCO0FBQUEsSUFDNUIsc0JBQXNCO0FBQUEsSUFDdEIsc0JBQXNCO0FBQUEsSUFDdEIsK0JBQStCO0FBQUEsSUFDL0IsNkJBQTZCO0FBQUEsSUFDN0IsZ0NBQWdDO0FBQUEsSUFDaEMscUJBQXFCO0FBQUEsSUFDckIsNkJBQTZCO0FBQUEsSUFDN0IsMEJBQTBCO0FBQUEsSUFDMUIsdUJBQXVCO0FBQUEsSUFDdkIsdUJBQXVCO0FBQUEsSUFDdkIsZ0JBQWdCO0FBQUEsSUFDaEIsc0JBQXNCO0FBQUEsSUFDdEIsY0FBYztBQUFBLElBQ2Qsb0JBQW9CO0FBQUEsSUFDcEIsb0JBQW9CO0FBQUEsSUFDcEIsc0JBQXNCO0FBQUEsSUFDdEIsYUFBYTtBQUFBLElBQ2IsY0FBYztBQUFBLElBQ2QsbUJBQW1CO0FBQUEsSUFDbkIsbUJBQW1CO0FBQUEsSUFDbkIseUJBQXlCO0FBQUEsSUFDekIsZUFBZTtBQUFBLElBQ2YsaUJBQWlCO0FBQUEsSUFDakIsdUJBQXVCO0FBQUEsSUFDdkIscUJBQXFCO0FBQUEsSUFDckIscUJBQXFCO0FBQUEsSUFDckIsdUJBQXVCO0FBQUEsSUFDdkIsY0FBYztBQUFBLElBQ2QsZUFBZTtBQUFBLElBQ2Ysb0JBQW9CO0FBQUEsSUFDcEIsb0JBQW9CO0FBQUEsSUFDcEIsMEJBQTBCO0FBQUEsSUFDMUIsZ0JBQWdCO0FBQUEsSUFDaEIsNEJBQTRCO0FBQUEsSUFDNUIsNEJBQTRCO0FBQUEsSUFDNUIseURBQXlEO0FBQUEsSUFDekQsc0NBQXNDO0FBQUEsSUFDdEMsb0JBQW9CO0FBQUEsSUFDcEIscUJBQXFCO0FBQUEsSUFDckIscUJBQXFCO0FBQUEsSUFDckIsc0JBQXNCO0FBQUEsSUFDdEIsZ0NBQWdDO0FBQUEsSUFDaEMsa0NBQWtDO0FBQUEsSUFDbEMsbUNBQW1DO0FBQUEsSUFDbkMsb0NBQW9DO0FBQUEsSUFDcEMsK0JBQStCO0FBQUEsSUFDL0IsNkJBQTZCO0FBQUEsSUFDN0IsdUJBQXVCO0FBQUEsSUFDdkIsaUNBQWlDO0FBQUEsSUFDakMsOEJBQThCO0FBQUEsSUFDOUIsNEJBQTRCO0FBQUEsSUFDNUIsc0NBQXNDO0FBQUEsSUFDdEMsNEJBQTRCO0FBQUEsSUFDNUIsc0JBQXNCO0FBQUEsSUFDdEIsa0NBQWtDO0FBQUEsSUFDbEMsc0JBQXNCO0FBQUEsSUFDdEIsd0JBQXdCO0FBQUEsSUFDeEIsd0JBQXdCO0FBQUEsSUFDeEIsbUJBQW1CO0FBQUEsSUFDbkIsMEJBQTBCO0FBQUEsSUFDMUIsOEJBQThCO0FBQUEsSUFDOUIseUJBQXlCO0FBQUEsSUFDekIsNkJBQTZCO0FBQUEsSUFDN0IsaUJBQWlCO0FBQUEsSUFDakIsZ0JBQWdCO0FBQUEsSUFDaEIsc0JBQXNCO0FBQUEsSUFDdEIsZUFBZTtBQUFBLElBQ2YseUJBQXlCO0FBQUEsSUFDekIsd0JBQXdCO0FBQUEsSUFDeEIsb0JBQW9CO0FBQUEsSUFDcEIscUJBQXFCO0FBQUEsSUFDckIsaUJBQWlCO0FBQUEsSUFDakIsaUJBQWlCO0FBQUEsSUFDakIsc0JBQXNCO0FBQUEsSUFDdEIsbUNBQW1DO0FBQUEsSUFDbkMscUNBQXFDO0FBQUEsSUFDckMsdUJBQXVCO0FBQUEsSUFDdkIsc0JBQXNCO0FBQUEsSUFDdEIsd0JBQXdCO0FBQUEsSUFDeEIsZUFBZTtBQUFBLElBQ2YsMkJBQTJCO0FBQUEsSUFDM0IsMEJBQTBCO0FBQUEsSUFDMUIsNkJBQTZCO0FBQUEsSUFDN0IsWUFBWTtBQUFBLElBQ1osZ0JBQWdCO0FBQUEsSUFDaEIsa0JBQWtCO0FBQUEsSUFDbEIsZ0JBQWdCO0FBQUEsSUFDaEIsa0JBQWtCO0FBQUEsSUFDbEIsbUJBQW1CO0FBQUEsSUFDbkIsWUFBWTtBQUFBLElBQ1oscUJBQXFCO0FBQUEsSUFDckIsc0JBQXNCO0FBQUEsSUFDdEIsc0JBQXNCO0FBQUEsSUFDdEIsOEJBQThCO0FBQUEsSUFDOUIsaUJBQWlCO0FBQUEsSUFDakIseUJBQXlCO0FBQUEsSUFDekIsMkJBQTJCO0FBQUEsSUFDM0IsK0JB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QUFhO0FBRVosSUFBVTtBQUFBLENBQVYsQ0FBVUMsYUFBVjtBQUVJLFdBQVMsT0FBTyxRQUFxQixVQUF5QjtBQUNqRSxXQUFPRCxPQUFLLGVBQWUsRUFBRSxNQUFNLENBQUM7QUFBQSxFQUN4QztBQUZPLEVBQUFDLFNBQVM7QUFBQSxHQUZIO0FBT1YsSUFBVTtBQUFBLENBQVYsQ0FBVUMsWUFBVjtBQU9JLFdBQVNDLFFBQXNCO0FBQ2xDLFdBQU9ILE9BQUssVUFBVTtBQUFBLEVBQzFCO0FBRk8sRUFBQUUsUUFBUyxPQUFBQztBQUFBLEdBUEg7OztBdkJkakIsT0FBTyxTQUFTLE9BQU8sVUFBVSxDQUFDO0FBd0RsQyxPQUFPLE9BQU8sU0FBZ0I7QUFDOUIsT0FBTyxPQUFPLFdBQVc7QUFLekIsT0FBTyxPQUFPLHlCQUF5QixlQUFPLHVCQUF1QixLQUFLLGNBQU07QUFHaEYsT0FBTyxPQUFPLGtCQUFrQjtBQUNoQyxPQUFPLE9BQU8sa0JBQWtCO0FBQ2hDLE9BQU8sT0FBTyxpQkFBaUI7QUFFeEIsT0FBTyxxQkFBcUI7QUFPNUIsU0FBUyxtQkFBbUIsS0FBNEI7QUFDM0QsU0FBTyxNQUFNLEtBQUssRUFBRSxRQUFRLE9BQU8sQ0FBQyxFQUMvQixLQUFLLGNBQVk7QUFDZCxRQUFJLFNBQVMsSUFBSTtBQUdiLFlBQU0sZUFBZSxTQUFTLFFBQVEsSUFBSSxjQUFjLEtBQUssSUFBSSxZQUFZO0FBQzdFLFVBQUksWUFBWSxTQUFTLFlBQVksR0FBRztBQUNwQyxjQUFNLFNBQVMsU0FBUyxjQUFjLFFBQVE7QUFDOUMsZUFBTyxNQUFNO0FBQ2IsaUJBQVMsS0FBSyxZQUFZLE1BQU07QUFBQSxNQUNwQztBQUFBLElBQ0o7QUFBQSxFQUNKLENBQUMsRUFDQSxNQUFNLE1BQU07QUFBQSxFQUFDLENBQUM7QUFDdkI7QUFHQSxtQkFBbUIsa0JBQWtCOyIsCiAgIm5hbWVzIjogWyJfYSIsICJFcnJvciIsICJjYWxsIiwgIkVycm9yIiwgIl9hIiwgIkFycmF5IiwgIk1hcCIsICJBcnJheSIsICJNYXAiLCAia2V5IiwgImNhbGwiLCAiX2EiLCAiX2EiLCAicmVzaXphYmxlIiwgIl9hIiwgImNhbGwiLCAiX2EiLCAiY2FsbCIsICJfYSIsICJjYWxsIiwgIkhpZGVNZXRob2QiLCAiU2hvd01ldGhvZCIsICJpc0RvY3VtZW50RG90QWxsIiwgIl9hIiwgInJlYXNvbiIsICJ2YWx1ZSIsICJjYWxsIiwgImNhbGwiLCAiY2FsbCIsICJjYWxsIiwgIkhhcHRpY3MiLCAiRGV2aWNlIiwgIkluZm8iXQp9Cg==
