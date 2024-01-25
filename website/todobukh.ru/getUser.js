// let config;
// let ipAddress;
// let port;

fetch("./config.json")
  .then((response) => response.json())
  .then((data) => {
    config = data;
    // Use config as a regular JSON object
    ipAddress = config.ipAddress;
    port = config.port;

    getUser();
  })
  .catch((error) => console.error("Error fetching config.json:", error));

async function getUser() {
  let username;
  let login = document.getElementById("indexLogin");
  let show = document.getElementById("show");
  let olist = document.getElementById("olist");
  let unauthorized = document.getElementById("unauthorized");
  let logout = document.getElementById("indexLogout");
  let signedAs = document.getElementById("signedAs");
  const response = await fetch(`https://${ipAddress}:${port}/api/user`, {
    headers: { "Content-Type": "application/json" },
    credentials: "include",
  });
  username = await response.json();



  if (username === "" || typeof username === "undefined") {
    login.style.display = "block";
    addForm.style.display = "none";
    updateForm.style.display = "none";
    show.style.display = "none";
    olist.style.display = "none";
    unauthorized.style.display = "block";
    logout.style.display = "none";
  } else {
    login.style.display = "none";
    addForm.style.display = "block";
    updateForm.style.display = "block";
    show.style.display = "block";
    olist.style.display = "block";
    unauthorized.style.display = "none";
    logout.style.display = "block";
    signedAs.innerHTML = "Currently signed as: " + username;
  }
}
