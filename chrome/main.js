function loadBalance() {
  passphrase = document.getElementById("passphrase").value;
  fetch("http://localhost:9090/" + passphrase).then(function (response) {
    return response.json();
  }).then(function (json) {
    document.getElementById("results").textContent = JSON.stringify(json);
  });
}

window.addEventListener("load", function () {
  document.getElementById("load_balance").addEventListener("click", function () {
    loadBalance();
  });

  loadBalance();
});
