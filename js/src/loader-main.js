// This code is injected into .coinkit pages in order to load their actual content.

window.stop();

console.log("loading begins");

chrome.runtime.sendMessage(
  {
    getFile: {
      hostname: window.location.hostname,
      pathname: window.location.pathname
    }
  },
  response => {
    console.log("loading complete");
    if (!response) {
      document.write(
        "error: received empty response from extension. check extension logs"
      );
      return;
    }
    if (response.error) {
      document.write("error: " + response.error);
    } else {
      document.write(response);
    }
  }
);
