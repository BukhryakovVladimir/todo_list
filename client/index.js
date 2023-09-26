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
   .then(data => create(data))
})  

function create(data) {
    let str
    let ul = document.getElementById("ulist")
    for (let i = 0; i < data.length; i++) {
        str = "<li>" + JSON.stringify(data[i]) + "</li>"
        ul.insertAdjacentHTML("beforeend", str);
        console.log(str);
    }
}