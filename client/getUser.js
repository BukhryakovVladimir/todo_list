getUser();

async function getUser() {
    let username;
    let login = document.getElementById("indexLogin");
    let show = document.getElementById("show");
    let olist = document.getElementById("olist");
    let unauthorized = document.getElementById("unauthorized");
    let logout = document.getElementById("indexLogout");
    const response = await fetch(
      "http://localhost:3000/api/user",
      {
        headers: { "Content-Type": "application/json" },
        credentials: "include",
      }
    );
    username = await response.json();
    
    console.log(username);
    
    if (username === "" || typeof username === "undefined") {
        login.style.display = "block";
        addForm.style.display = "none";
        deleteForm.style.display = "none";
        show.style.display = "none";
        olist.style.display = "none";
        unauthorized.style.display = "block";
        logout.style.display = "none";
    } else {
        login.style.display = "none";
        addForm.style.display = "block";
        deleteForm.style.display = "block";
        show.style.display = "block";
        olist.style.display = "block";
        unauthorized.style.display = "none";
        logout.style.display = "block";
    }
}