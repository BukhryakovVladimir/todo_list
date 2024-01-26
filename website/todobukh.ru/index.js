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

    
  }).then(window.onload = () => { getList();})
  .catch((error) => console.error("Error fetching config.json:", error));



let addForm = document.getElementById("addForm");
// add button to delete element from to-do list by it's respectful number
addForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let addInput = document.getElementById("addInput").value;

  fetch(`https://${ipAddress}:${port}/api/write`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    //mode: "no-cors", //probably can delete that
    body: addInput,
  }).then((response) => {
    if (response.ok) {
      // Успешная запись задачи в БД, вывод обновлённого списка.
      getList();
    }
  }).catch((error) => {
    console.error("There was a problem with the fetch operation:", error);
  });

});

let updateForm = document.getElementById("updateForm");
// add button to delete element from to-do list by it's respectful number
updateForm.addEventListener("submit", (e) => {
  e.preventDefault();

  let updateInputNumber = document.getElementById("updateInputNumber").value;
  let updateInputTask = document.getElementById("updateInputTask").value;

  fetch(`https://${ipAddress}:${port}/api/update`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    //mode: "no-cors", //probably can delete that
    body: JSON.stringify({task_number: updateInputNumber, task_description: updateInputTask}),
  }).then((response) => {
    if (response.ok) {
      // Успешное одновление задачи в БД, вывод обновлённого списка.
      getList();
    }
  }).catch((error) => {
    console.error("There was a problem with the fetch operation:", error);
  });

});

let button = document.getElementById("show");

button.addEventListener("click", (e) => {
  getList();
});

function create(data) {
  addInput.value = "";
  updateInputNumber.value = "";
  updateInputTask.value = "";
  let str;
  let ol = document.getElementById("olist");
  let parsedData;
  if (data !== null) {
      for (let i = 0; i < data.length; i++) {
        // !!
        // !! Add on click function to button, inside that function it fetches the database by number, that number will be counter using loop here !!
        // !!
        // span #completed
        parsedData = JSON.parse(JSON.stringify(data[i]));
        if (parsedData.is_completed) {
          str = `<li> 
              <span class="completedTask">${parsedData.task_description}</span> 
              <button class="regularButton" id="crossmark" onclick="setIsCompletedFalse(${
                i + 1
              })">
              <img src="../style/crossmark.png" title="Undo Task Completed" />
              </button>
              <button class="regularButton" id="wastebin" onclick="deleteTask(${
                i + 1
              })"><img src="../style/wastebin.svg" title="Remove Task"/></button> 
              </li>`;
        } else {
          str = `<li> 
              <span class="regularTask">${parsedData.task_description}</span> 
              <button class="regularButton" id="checkmark" onclick="setIsCompletedTrue(${
                i + 1
              })">
              <img src="../style/checkmark.png" title="Task Completed" />
              </button>
              <button class="regularButton" id="wastebin" onclick="deleteTask(${
                i + 1
              })"><img src="../style/wastebin.svg" title="Remove Task"/></button> 
              </li>`;
        }
        ol.insertAdjacentHTML("beforeend", str);
      }
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
  fetch(`https://${ipAddress}:${port}/api/read`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
  })
    .then((response) => response.json())
    .then((data) => {
      create(data);
    });
}

function setIsCompletedTrue(rowNum) {
  fetch(`https://${ipAddress}:${port}/api/setIsCompletedTrue`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: rowNum,
  }).then((response) => {
    if (response.ok) {
      // Успешная установка флага о выполнении задачи в значение true в БД, вывод обновлённого списка.
      getList();
    }
  }).catch((error) => {
    console.error("There was a problem with the fetch operation:", error);
  });
}

function setIsCompletedFalse(rowNum) {
  fetch(`https://${ipAddress}:${port}/api/setIsCompletedFalse`, {
    method: "PUT",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: rowNum,
  }).then((response) => {
    if (response.ok) {
      // Успешная установка флага о выполнении задачи в значение false в БД, вывод обновлённого списка.
      getList();
    }
  }).catch((error) => {
    console.error("There was a problem with the fetch operation:", error);
  });
}

function deleteTask(rowNum) {
  fetch(`https://${ipAddress}:${port}/api/delete`, {
    method: "DELETE",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: rowNum,
  }).then((response) => {
    if (response.ok) {
      // Успешная удаление задачи в БД, вывод обновлённого списка.
      getList();
    }
  }).catch((error) => {
    console.error("There was a problem with the fetch operation:", error);
  });
}
