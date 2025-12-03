function copyToClipboard() {
  const input = document.getElementById("funnelUrl");
  const button = document.getElementById("copyBtn");
  const feedback = document.getElementById("copyFeedback");

  // Select and copy the text
  input.select();
  input.setSelectionRange(0, 99999); // For mobile devices

  try {
    // Modern clipboard API
    navigator.clipboard
      .writeText(input.value)
      .then(() => {
        showCopyFeedback();
      })
      .catch(() => {
        // Fallback for older browsers
        document.execCommand("copy");
        showCopyFeedback();
      });
  } catch (err) {
    // Fallback for older browsers
    document.execCommand("copy");
    showCopyFeedback();
  }

  function showCopyFeedback() {
    // Update button
    button.textContent = "Copied!";
    button.classList.add("copied");

    // Show feedback message
    feedback.classList.add("show");

    // Reset after 2 seconds
    setTimeout(() => {
      button.textContent = "Copy Link";
      button.classList.remove("copied");
      feedback.classList.remove("show");
    }, 2000);
  }
}

// Allow copying by clicking the input field
document.getElementById("funnelUrl").addEventListener("click", function () {
  this.select();
});
