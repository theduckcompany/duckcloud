
import { Sidenav, Datatable, Dropdown, Select } from "/assets/js/libs/mdb.es.min.js";

export function SetupSideNav() {
  const sidenav = document.getElementById("main-sidenav");

  let innerWidth = null;

  const setMode = (e) => {
    const sidenavInstance = Sidenav.getOrCreateInstance(sidenav);
    // Check necessary for Android devices
    if (window.innerWidth === innerWidth) {
      return;
    }

    innerWidth = window.innerWidth;

    if (window.innerWidth < 960) {
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
}


export function SetupBoostrapElems() {
  // Make all the selects pretty even with the dynamic content
  document.body.addEventListener("htmx:afterSwap", function(evt) {
    document.querySelectorAll('.select').forEach((select) => {
      Select.getOrCreateInstance(select);
    });

    document.querySelectorAll('.dropdown').forEach((dropdown) => {
      Dropdown.getOrCreateInstance(dropdown);
    });

    document.querySelectorAll('.datatable').forEach((datatable) => {
      Datatable.getOrCreateInstance(datatable);
    });
  })
}

