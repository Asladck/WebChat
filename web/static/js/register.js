document.getElementById("registerForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const name = document.getElementById("name").value.trim();
    const email = document.getElementById("email").value.trim();
    const username = document.getElementById("username").value.trim();
    const password = document.getElementById("password").value.trim();
    const messageDiv = document.getElementById("message");

    try {
        const res = await fetch("/auth/sign-up", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ name, email, username, password })
        });

        const data = await res.json();

        if (res.ok) {
            messageDiv.innerText = "Registration successful!";
            messageDiv.classList.add("success");
            messageDiv.classList.remove("failed");

            setTimeout(() => window.location.href = "/sign-in", 1000);
        } else {
            messageDiv.innerText = data.message || "Registration failed.";
            messageDiv.classList.remove("success");
            messageDiv.classList.add("failed");

        }
    } catch (err) {
        console.error("Registration error:", err);
        messageDiv.innerText = "Something went wrong.";
        messageDiv.classList.remove("success");
        messageDiv.classList.add("failed");
    }
});
penis
