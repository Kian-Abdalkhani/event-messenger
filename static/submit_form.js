// Show live character count
const messageField = document.getElementById("message");
const charCount = document.getElementById("charCount");

messageField.addEventListener("input", function (event) {
  const currentLength = this.value.length;
  const maxLength = 500;

  // Update character count
  charCount.textContent = this.value.length + " / 500 characters";

  // Conditional to add or remove red css styling based on character count
  if (currentLength > maxLength) {
    charCount.classList.add("over-limit");
  } else {
    charCount.classList.remove("over-limit");
  }
});
