window.onload = () => {
  getList();
};

let addForm = document.getElementById("addForm");
// add button to delete element from to-do list by it's respectful number
addForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let addInput = document.getElementById("addInput").value;

  fetch("http://localhost:3000/api/write", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    //mode: "no-cors", //probably can delete that
    body: addInput
  });

  setTimeout(() => {
    getList();
  }, 200)
});

let deleteForm = document.getElementById("deleteForm");

deleteForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let deleteInput = document.getElementById("deleteInput").value;

  fetch("http://localhost:3000/api/delete", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: deleteInput
  });

  setTimeout(() => {
    getList();
  }, 200)

});

let button = document.getElementById("show");

button.addEventListener("click", (e) => {
  getList();
});

function create(data) {
  let str;
  let ol = document.getElementById("olist");
  for (let i = 0; i < data.length; i++) {
    str = `<li> ${JSON.parse(JSON.stringify(data[i]))} </li>`;
    ol.insertAdjacentHTML("beforeend", str);
  }
}

function clearList() {
  var element = document.getElementsByTagName("li"),
    index;

  for (index = element.length - 1; index >= 0; index--) {
    element[index].parentNode.removeChild(element[index]);
  }
}

function getList() {
  clearList();
  fetch("http://localhost:3000/api/read", {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then((response) => response.json())
    .then((data) => create(data));
}

