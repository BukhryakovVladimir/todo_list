let signupForm = document.getElementById("signupForm");

signupForm.addEventListener("submit", (e) => {
    e.preventDefault();
  
    let username = document.getElementById("signupUsername").value;
    let password = document.getElementById("signupPassword").value;
    
    console.log(JSON.stringify({username: username, password: password}))
    fetch("http://localhost:3000/api/signup", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      //mode: "no-cors", //probably can delete that
      body: JSON.stringify({username: username, password: password})
    });
  
});
  