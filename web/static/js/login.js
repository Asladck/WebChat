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
            body: JSON.stringify({ email, username, password_hash: password }),
        });
        let data;
        try{
            data = await res.json();
        }catch (jsonErr){
            throw new Error("Invalid JSON response")
        }
        if (res.ok && data.access_token && data.refresh_token) {
            // Сохраняем токены
            localStorage.setItem("access_token", data.access_token);
            localStorage.setItem("refresh_token", data.refresh_token);

            // Декодируем токен
            const payloadBase64 = data.access_token.split('.')[1];
            const tokenPayload = JSON.parse(atob(payloadBase64));
            localStorage.setItem("username", tokenPayload.username);

            console.log("Login successful:", tokenPayload);

            // Переход на главную
            window.location.href = "/";
        }  else {
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
