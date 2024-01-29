import {Sidenav, Dropdown, Navbar, Select, Modal, initMDB} from "/assets/js/libs/mdb.es.min.js";

initMDB({Sidenav, Dropdown, Navbar, Select, Modal});

const sidenav = document.getElementById("main-sidenav");
const sidenavInstance = Sidenav.getInstance(sidenav);

let innerWidth = null;

const setMode = (e) => {
  // Check necessary for Android devices
  if (window.innerWidth === innerWidth) {
    return;
  }

  innerWidth = window.innerWidth;

  if (window.innerWidth < 1400) {
    sidenavInstance.changeMode("over");
    sidenavInstance.hide();
  } else {
    sidenavInstance.changeMode("side");
    sidenavInstance.show();
  }
};

setMode();

// Event listeners
window.addEventListener("resize", setMode);

// Make all the selects pretty even with the dynamic content
document.body.addEventListener("htmx:afterSwap", function (evt) {
  document.querySelectorAll('.select').forEach((select) => {
    const res = Select.getInstance(select);
    if (res) {
      res.dispose();
    }
    new Select(select);
  });
})

document.body.addEventListener("htmx:afterSettle", function (evt) {
  document.querySelectorAll('.select').forEach((select) => {
    select.classList.add('select-initialized');
  });
});
