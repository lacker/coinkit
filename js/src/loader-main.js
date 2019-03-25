// This code is injected into .coinkit pages in order to load their actual content.

// Stops the process of loading the nonexistent .coinkit url
window.stop();

// Starts displaying a blank document, rather than whatever document was previously
// shown in the browser
document.open();

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
