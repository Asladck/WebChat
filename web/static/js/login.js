function parseJwt(token) {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) throw new Error('Invalid token');
        let payload = parts[1];
        payload = payload.replace(/-/g, '+').replace(/_/g, '/');
        while (payload.length % 4) payload += '=';
        return JSON.parse(atob(payload));
    } catch (e) {
        console.error("parseJwt error:", e);
        return null;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById("loginForm");
    const messageDiv = document.getElementById("message");
    const submitBtn = document.getElementById("loginSubmit");

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        messageDiv.innerHTML = '';

        const emailEl = document.getElementById("email");
        const usernameEl = document.getElementById("username");
        const passwordEl = document.getElementById("password");
        let valid = true;

        if (!emailEl.checkValidity()) { emailEl.classList.add('is-invalid'); valid = false; } else emailEl.classList.remove('is-invalid');
        if (!usernameEl.checkValidity()) { usernameEl.classList.add('is-invalid'); valid = false; } else usernameEl.classList.remove('is-invalid');
        if (!passwordEl.checkValidity()) { passwordEl.classList.add('is-invalid'); valid = false; } else passwordEl.classList.remove('is-invalid');

        if (!valid) {
            messageDiv.innerHTML = '<div class="alert alert-danger">Please fix form errors.</div>';
            return;
        }

        const email = emailEl.value.trim();
        const username = usernameEl.value.trim();
        const password = passwordEl.value.trim();

        submitBtn.disabled = true;
        submitBtn.textContent = 'Signing in...';

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
                localStorage.setItem("access_token", data.access_token);
                localStorage.setItem("refresh_token", data.refresh_token);

                const tokenPayload = parseJwt(data.access_token);
                if (tokenPayload && tokenPayload.username) {
                    localStorage.setItem("username", tokenPayload.username);
                }

                messageDiv.innerHTML = '<div class="alert alert-success">Signed in. Redirecting…</div>';
                setTimeout(()=> window.location.href = "/", 600);
            }  else {
                messageDiv.innerHTML = `<div class="alert alert-danger">${data && data.message ? data.message : 'Login failed'}</div>`;
            }
        } catch (err) {
            console.error("Login error:", err);
            messageDiv.innerHTML = '<div class="alert alert-danger">Something went wrong.</div>';
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Sign in';
        }
    });
});
