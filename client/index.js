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
    credentials: "include",
    //mode: "no-cors", //probably can delete that
    body: addInput,
  });

  setTimeout(() => {
    getList();
  }, 200);
});

let button = document.getElementById("show");

button.addEventListener("click", (e) => {
  getList();
});

function create(data) {
  let str;
  let ol = document.getElementById("olist");
  let parsedData;
  for (let i = 0; i < data.length; i++) {
    // !!
    // !! Add on click function to button, inside that function it fetches the database by number, that number will be counter using loop here !!
    // !!
    // span #completed
    parsedData = JSON.parse(JSON.stringify(data[i]));
    if (parsedData.is_completed) {
      str = `<li> 
      <span class="completedTask">${parsedData.task_description}</span> 
      <button class="regularButton" id="crossmark" onclick="setIsCompletedFalse(${i+1})">
      <img src="../style/crossmark.png" title="Undo Task Completed" />
      </button>
      <button class="regularButton" id="wastebin" onclick="deleteTask(${i+1})"><img src="../style/wastebin.png" title="Remove Task"/></button> 
      </li>`;
    } else {
      str = `<li> 
      <span class="regularTask">${parsedData.task_description}</span> 
      <button class="regularButton" id="checkmark" onclick="setIsCompletedTrue(${i+1})">
      <img src="../style/checkmark.png" title="Task Completed" />
      </button>
      <button class="regularButton" id="wastebin" onclick="deleteTask(${i+1})"><img src="../style/wastebin.png" title="Remove Task"/></button> 
      </li>`;
    }
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
    credentials: "include",
  })
    .then((response) => response.json())
    .then((data) => {
      create(data);
      console.log(data);
    });
}

function setIsCompletedTrue(rowNum) {
  fetch("http://localhost:3000/api/setIsCompletedTrue", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: rowNum,
  });

  setTimeout(() => {
    getList();
  }, 200);
}

function setIsCompletedFalse(rowNum) {
  fetch("http://localhost:3000/api/setIsCompletedFalse", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: rowNum,
  });

  setTimeout(() => {
    getList();
  }, 200);
}

function deleteTask(rowNum) {
    fetch("http://localhost:3000/api/delete", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
      body: rowNum,
    });

    setTimeout(() => {
      getList();
    }, 200);
}
