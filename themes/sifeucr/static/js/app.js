function showDialog() {
  const dialog = document.getElementById('dialog');
  const overlay = document.getElementById('overlay');

  overlay.classList.add('active');
  dialog.classList.add('active');

  overlay.addEventListener('click', (event) => {
    if (!dialog.contains(event.target)) {
      closeDialog();
    }
  });
}

function closeDialog() {
	document.getElementById('overlay').classList.remove('active');
	document.getElementById('dialog').classList.remove('active');
	document.getElementById('dialog-content').innerHTML = '<div class="spinner"></div>';
}
