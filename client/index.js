let form = document.getElementById("form");

form.addEventListener("submit", e => {
    e.preventDefault();

    let input = document.getElementById("input").value;

    fetch("http://localhost:3000/api", {
        method: "POST",
        headers: {"Content-Type": "application/json"},
        mode: "no-cors",
        body: input
    })

    console.log(JSON.stringify(response.body));
});

let button = document.getElementById("show");
var str = {};
button.addEventListener("click", e => {
    fetch('http://localhost:3000/api/get', {
    method: 'GET',
    headers: {
        'Content-Type': 'application/json',
    },
})
   .then(response => response.json())
   .then(data => console.log(data))
})
