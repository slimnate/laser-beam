// Remove 'noscript' class when js is enabled
document.getElementsByClassName("noscript").item(0).className = "";

// function to reinitialize interactive components after htmx swap
function handleAfterSettle() {
  initFlowbite();
}
