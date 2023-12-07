let loginForm = document.getElementById("loginForm");

loginForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let username = document.getElementById("loginUsername").value;
  let password = document.getElementById("loginPassword").value;

  fetch("http://localhost:3000/api/login", {
    method: "POST",
    headers: { "Accept": "application/json", "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ username: username, password: password }),
  })
  .then(response => {
    if (response.ok) {
      // Successful authentication, redirect to index.html
      window.location.href = "index.html";
    } else {
      // Check for 401 Unauthorized status
      return response.json(); // Parse JSON response
    }
  })
  .then(responseJson => {
    console.log(responseJson);
    if (responseJson === "Username not found") {
      console.log(responseJson);
      document.getElementById("invalidUsername").style.visibility = "visible";
      document.getElementById("invalidPassword").style.visibility = "hidden";
    } else if (responseJson === "Incorrect password") {
      console.log(responseJson);
      document.getElementById("invalidPassword").style.visibility = "visible";
      document.getElementById("invalidUsername").style.visibility = "hidden";
    }
  })
  .catch(error => {
    console.error('There was a problem with the fetch operation:', error);
  });
});