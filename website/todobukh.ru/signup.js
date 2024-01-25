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

let signupForm = document.getElementById("signupForm");

signupForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let username = document.getElementById("signupUsername").value;
  let password = document.getElementById("signupPassword").value;

  let invalidUsername = document.getElementById("invalidUsername");
  let invalidPassword = document.getElementById("invalidPassword");

  fetch(`https://${ipAddress}:${port}/api/signup`, {
    method: "POST",
    headers: { Accept: "application/json", "Content-Type": "application/json" },
    body: JSON.stringify({ username: username, password: password }),
  })
    .then((response) => {
      if (response.ok) {
        // Успешная регистрация, редирект на login.html.
        window.location.href = "login.html";
      } else {
        // Проверка http статуса 401 Unauthorized.
        return response.json(); // Парсинг JSON'а
      }
    })
    .then((responseJson) => {
      switch (responseJson.slice(0, 8)) {
        case "Username":
          invalidUsername.innerHTML = responseJson;
          invalidPassword.innerHTML = "Invalid Password placeholder";

          invalidUsername.style.visibility = "visible";
          invalidPassword.style.visibility = "hidden";
          break;
        case "Password":
          invalidUsername.innerHTML = "Invalid Username placeholder";
          invalidPassword.innerHTML = responseJson;

          invalidUsername.style.visibility = "hidden";
          invalidPassword.style.visibility = "visible";
          break;
      }
    })
    .catch((error) => {
      console.error("There was a problem with the fetch operation:", error);
    });
});
