// This code is injected into .coinkit pages in order to load their actual content.

console.log("running loader-main.js");

window.stop();

fetch("http://localhost:4444")
  .then(response => {
    return response.json();
  })
  .then(json => {
    console.log(json);
  });
