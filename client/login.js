// let loginForm = document.getElementById("loginForm");
// loginForm.addEventListener("submit", (e) => {
//   e.preventDefault();
  
//   let xhr = new XMLHttpRequest();

//   let buddy = JSON.stringify({
//     username: document.getElementById("loginUsername").value,
//     password: document.getElementById("loginPassword").value
//   })

//   xhr.open("POST", "http://localhost:3000/api/login")
//   xhr.setRequestHeader('Content-type', 'application/json')
//   xhr.withCredentials = true;
//   xhr.send(buddy);
// });


let loginForm = document.getElementById("loginForm");

loginForm.addEventListener("submit", (e) => {
  e.preventDefault();
    console.log("try");
  let username = document.getElementById("loginUsername").value;
  let password = document.getElementById("loginPassword").value;

  console.log(JSON.stringify({ username: username, password: password }));
  fetch("http://localhost:3000/api/login", {
    method: "POST",
    headers: {"Accept": "application/json", "Content-Type": "application/json" },
    credentials: "include",
    //mode: "no-cors", //probably can delete that
    body: JSON.stringify({ username: username, password: password }),
  }).then(setTimeout(() => {window.location.href = "index.html"}, 2000));
});

