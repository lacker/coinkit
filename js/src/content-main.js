// This code runs in our content script, in the context of every web page.

window.addEventListener(
  "message",
  event => {
    if (event.source != window || event.data.type != "toCoinkit") {
      return;
    }

    chrome.runtime.sendMessage(event.data.message, response => {
      let data = {
        id: event.data.id,
        type: "fromCoinkit",
        message: response
      };
      window.postMessage(data, "*");
    });
  },
  false
);
