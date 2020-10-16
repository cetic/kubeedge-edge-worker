function openNav() {

}

function closeNav() {
  $(".icon").toggleClass("close");
  const element = document.querySelector(".icon");
  if (element.classList.contains("close")){
    document.getElementById("mySidenav").style.width = "250px";
  } else {
    document.getElementById("mySidenav").style.width = "55px";
  }
}
