document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById("registerForm");
    const messageDiv = document.getElementById("message");
    const submitBtn = document.getElementById("registerSubmit");

    form.addEventListener("submit", async (e) => {
        e.preventDefault();
        messageDiv.innerHTML = '';

        const nameEl = document.getElementById("name");
        const emailEl = document.getElementById("email");
        const usernameEl = document.getElementById("username");
        const passwordEl = document.getElementById("password");

        let valid = true;
        [nameEl, emailEl, usernameEl, passwordEl].forEach(el=>{
            if (!el.checkValidity()) { el.classList.add('is-invalid'); valid = false; } else el.classList.remove('is-invalid');
        });
        if (!valid) {
            messageDiv.innerHTML = '<div class="alert alert-danger">Please fix form errors.</div>';
            return;
        }

        const name = nameEl.value.trim();
        const email = emailEl.value.trim();
        const username = usernameEl.value.trim();
        const password = passwordEl.value.trim();

        submitBtn.disabled = true;
        submitBtn.textContent = 'Creating...';

        try {
            const res = await fetch("/auth/sign-up", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({ name, email, username, password })
            });

            const data = await res.json().catch(()=>null);

            if (res.ok) {
                messageDiv.innerHTML = '<div class="alert alert-success">Registration successful! Redirecting to sign in…</div>';
                setTimeout(() => window.location.href = "/sign-in", 900);
            } else {
                messageDiv.innerHTML = `<div class="alert alert-danger">${data && data.message ? data.message : 'Registration failed.'}</div>`;
            }
        } catch (err) {
            console.error("Registration error:", err);
            messageDiv.innerHTML = '<div class="alert alert-danger">Something went wrong.</div>';
        } finally {
            submitBtn.disabled = false;
            submitBtn.textContent = 'Create account';
        }
    });
});
