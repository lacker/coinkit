// This code is injected into .coinkit pages in order to load their actual content.

import WebTorrent from "webtorrent";

console.log("running loader-main.js");
window.stop();

let client = new WebTorrent();

fetch("http://localhost:4444")
  .then(response => {
    return response.json();
  })
  .then(json => {
    console.log(json.magnet);

    client.add(json.magnet, torrent => {
      torrent.on("done", () => {
        let file = torrent.files[0];
        console.log("length:", file.length, "downloaded:", file.downloaded);
        file.getBlob((err, blob) => {
          let reader = new FileReader();
          reader.onload = e => {
            document.write(e.target.result);
          };
          reader.readAsText(blob);
        });
      });
    });
  });
