// This code runs in our content script, in the context of every web page.

console.log("XXX this is content main take 3");

window.addEventListener(
  "message",
  event => {
    if (event.source != window || event.data.type != "toCoinkit") {
      return;
    }

    console.log("XXX content script received", event.data);
  },
  false
);
