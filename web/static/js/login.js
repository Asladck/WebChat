document.getElementById("loginForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const email = document.getElementById("email").value.trim();
    const username = document.getElementById("username").value.trim();
    const password = document.getElementById("password").value.trim();
    const messageDiv = document.getElementById("message");

    try {
        const res = await fetch("/auth/sign-in", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ email, username, password_hash: password })
        });

        const data = await res.json();

        if (res.ok) {
            const tokenPayload = JSON.parse(atob(data.access_token.split('.')[1]));
            localStorage.setItem("access_token", data.access_token);
            localStorage.setItem("refresh_token", data.refresh_token);
            localStorage.setItem("username", tokenPayload.username);
            console.log(localStorage.getItem("access_token"));
            window.location.href = "/";
        } else {
            messageDiv.innerText = data.message || "Login failed.";
            messageDiv.classList.remove("success");
            messageDiv.classList.add("failed");
        }
    } catch (err) {
        console.error("Login error:", err);
        messageDiv.innerText = "Something went wrong.";
        messageDiv.classList.remove("success");
        messageDiv.classList.add("failed");
    }
});
