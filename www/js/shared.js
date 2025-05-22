$(document).ready(function() {
  $("#menu-placeholder").load("components/menu.html", function() {
    // Highlight current page in menu
    const currentPage = window.location.pathname.split('/').pop();
    $(`.nav-link[href="${currentPage}"]`).addClass('active');
  });
});