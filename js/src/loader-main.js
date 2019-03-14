// This code is injected into .coinkit pages in order to load their actual content.

window.stop();

chrome.runtime.sendMessage(
  {
    getFile: {
      hostname: window.location.hostname,
      pathname: window.location.pathname
    }
  },
  response => {
    document.write(response);
  }
);
