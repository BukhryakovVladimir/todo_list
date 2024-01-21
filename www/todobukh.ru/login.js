let config;
let ipAddress;
let port;

fetch("./config.json")
  .then((response) => response.json())
  .then((data) => {
    config = data;
    // Use config as a regular JSON object
    ipAddress = config.ipAddress;
    port = config.port;
  })
  .catch((error) => console.error("Error fetching config.json:", error));

let loginForm = document.getElementById("loginForm");

loginForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let username = document.getElementById("loginUsername").value;
  let password = document.getElementById("loginPassword").value;

  fetch(`https://${ipAddress}:${port}/api/login`, {
    method: "POST",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ username: username, password: password }),
  })
    .then((response) => {
      if (response.ok) {
        // Successful authentication, redirect to index.html
        window.location.href = "index.html";
      } else {
        // Check for 401 Unauthorized status
        return response.json(); // Parse JSON response
      }
    })
    .then((responseJson) => {
      if (responseJson === "Username not found") {
        document.getElementById("invalidUsername").style.visibility = "visible";
        document.getElementById("invalidPassword").style.visibility = "hidden";
      } else if (responseJson === "Incorrect password") {
        document.getElementById("invalidPassword").style.visibility = "visible";
        document.getElementById("invalidUsername").style.visibility = "hidden";
      }
    })
    .catch((error) => {
      console.error("There was a problem with the fetch operation:", error);
    });
});
