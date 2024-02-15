import {Sidenav, Datatable, Navbar, Dropdown, Modal, Select, initMDB} from "/assets/js/libs/mdb.es.min.js";

initMDB({Sidenav, Datatable, Navbar, Dropdown, Modal, Select});

const sidenav = document.getElementById("main-sidenav");

let innerWidth = null;

const setMode = (e) => {
  const sidenavInstance = Sidenav.getOrCreateInstance(sidenav);
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
    new Select(select);
  });

  document.querySelectorAll('.dropdown').forEach((dropdown) => {
    new Dropdown(dropdown);
  });
})

