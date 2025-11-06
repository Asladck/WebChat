let currentSocket = null;
let reconnectAttempts = 0;
const maxReconnectAttempts = 5;
let socketInitialized = false;

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

function getInitials(name){
    if (!name) return "U";
    const parts = name.trim().split(/\s+/);
    if (parts.length === 1) return parts[0].slice(0,2).toUpperCase();
    return (parts[0][0] + parts[parts.length-1][0]).toUpperCase();
}

function formatTime(ts){
    try {
        const d = ts ? new Date(ts) : new Date();
        return d.toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'});
    } catch(e){ return new Date().toLocaleTimeString(); }
}

function safeText(s){ const el = document.createElement('div'); el.textContent = s; return el.textContent; }

// Заменяем два обработчика на один: валидируем токен, устанавливаем UI, затем запускаем WS.
document.addEventListener("DOMContentLoaded", async () => {
    const accessToken = localStorage.getItem("access_token");
    console.info("chat: access_token present?", !!accessToken);
    if (!accessToken) {
        console.warn("chat: no access_token -> redirect to /sign-in");
        window.location.href = "/sign-in";
        return;
    }

    const payload = parseJwt(accessToken);
    console.info("chat: token payload:", payload);
    if (!payload) {
        console.warn("chat: access token parse failed, trying refresh...");
        const newToken = await refreshAccessToken();
        if (newToken) {
            console.info("chat: refresh succeeded, using new access token");
            initializeChatWebSocket(newToken);
            const payload2 = parseJwt(newToken) || {};
            const username2 = payload2.username || "User";
            document.getElementById("profileLink")?.setAttribute('href', `/profile/${username2}`);
            document.getElementById("navbarUsername") && (document.getElementById("navbarUsername").textContent = username2);
            document.getElementById("usernameSmall") && (document.getElementById("usernameSmall").textContent = username2);
            return;
        }
        console.warn("chat: refresh failed -> clearing tokens and redirect");
        localStorage.removeItem("access_token");
        localStorage.removeItem("refresh_token");
        window.location.href = "/sign-in";
        return;
    }

    const username = payload.username || "User";
    document.getElementById("profileLink")?.setAttribute('href', `/profile/${username}`);
    document.getElementById("navbarUsername") && (document.getElementById("navbarUsername").textContent = username);
    document.getElementById("usernameSmall") && (document.getElementById("usernameSmall").textContent = username);

    initializeChatWebSocket(accessToken);
});

function initializeChatWebSocket(accessToken) {
    if (socketInitialized) {
        console.info("initializeChatWebSocket: already initialized");
        return;
    }
    socketInitialized = true;

    if (currentSocket) {
        try { currentSocket.close(1000, "reconnect"); } catch (e) { /* ignore */ }
    }

    const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
    console.info("initializeChatWebSocket: connecting with token (first 20 chars):", accessToken ? accessToken.slice(0,20) : null);
    const socket = new WebSocket(`${scheme}//${window.location.host}/ws?token=${encodeURIComponent(accessToken)}`);
    currentSocket = socket;

    const chat = document.getElementById('chat');
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');

    function addMessage({ time, address, message, isMyMessage }) {
        // avoid duplicates: basic dedupe by last message text
        const last = chat.lastElementChild;
        if (last && last.textContent && last.textContent.includes(message) && last.classList.contains('fade-in')) {
            // if the same message just added, skip
        }

        const row = document.createElement('div');
        row.className = 'msg-row fade-in ' + (isMyMessage ? 'me' : 'other');

        // avatar on left for others, right for me
        if (!isMyMessage) {
            const avatar = document.createElement('div');
            avatar.className = 'avatar';
            avatar.textContent = getInitials(address || 'Anon');
            row.appendChild(avatar);
        }

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble ' + (isMyMessage ? 'me' : 'other');

        const meta = document.createElement('div');
        meta.className = 'meta';
        meta.textContent = `${address} • ${formatTime(time)}`;

        const content = document.createElement('div');
        content.textContent = message; // safe via textContent

        bubble.appendChild(meta);
        bubble.appendChild(content);
        row.appendChild(bubble);

        if (isMyMessage) {
            // add avatar for own messages on right
            const avatar = document.createElement('div');
            avatar.className = 'avatar';
            avatar.textContent = getInitials("You");
            row.appendChild(avatar);
        }

        chat.appendChild(row);

        // auto-scroll: use last child into view
        try {
            row.scrollIntoView({ behavior: 'smooth', block: 'end' });
        } catch(e){
            chat.scrollTo({ top: chat.scrollHeight, behavior: 'smooth' });
        }
    }

    function updateSendState(){
        if (!messageInput) return;
        const val = messageInput.value.trim();
        if (sendButton) sendButton.disabled = (val.length === 0 || (socket && socket.readyState !== WebSocket.OPEN));
    }

    function sendMessage() {
        const text = messageInput.value.trim();
        if (!text || socket.readyState !== WebSocket.OPEN) return;
        try {
            socket.send(JSON.stringify({ message: text }));
            addMessage({
                time: new Date().toISOString(),
                address: "You",
                message: text,
                isMyMessage: true
            });
            messageInput.value = '';
            updateSendState();
            messageInput.focus();
        } catch (e) {
            console.error("sendMessage failed", e);
        }
    }

    if (sendButton) sendButton.onclick = sendMessage;
    if (messageInput) {
        messageInput.addEventListener('input', updateSendState);
        messageInput.addEventListener('keypress', (e) => { if (e.key === "Enter") { e.preventDefault(); sendMessage(); } });
    }
    updateSendState();

    socket.onopen = () => { reconnectAttempts = 0; console.info("WebSocket opened"); updateSendState(); };

    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (data.message) {
                addMessage({
                    time: data.time || new Date().toISOString(),
                    address: data.address || 'Anon',
                    message: data.message,
                    isMyMessage: !!data.isMyMessage
                });
            }
        } catch (e) {
            console.warn("Non-JSON or malformed message:", event.data);
        }
    };

    socket.onerror = async (ev) => {
        console.error("WebSocket error:", ev);
        const newToken = await refreshAccessToken();
        if (newToken) {
            socketInitialized = false;
            setTimeout(() => initializeChatWebSocket(newToken), 1000);
        } else {
            addMessage({
                time: new Date().toISOString(),
                address: "System",
                message: "Session expired. Please login again.",
                isMyMessage: false
            });
            localStorage.removeItem("access_token");
            localStorage.removeItem("refresh_token");
            setTimeout(() => window.location.href = "/sign-in", 1500);
        }
    };

    socket.onclose = async (event) => {
        console.warn("WebSocket closed:", event.code, event.reason);
        currentSocket = null;
        socketInitialized = false;

        if (event.code === 1001) {
            console.info("WebSocket client going away (likely navigation). Not forcing reconnect.");
            return;
        }

        if (event.code === 1000) {
            addMessage({
                time: new Date().toISOString(),
                address: "System",
                message: "Connection closed",
                isMyMessage: false
            });
            return;
        }

        if (reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++;
            const delay = Math.min(30, Math.pow(2, reconnectAttempts));
            console.info(`Reconnecting in ${delay}s (attempt ${reconnectAttempts}/${maxReconnectAttempts})...`);
            const newToken = await refreshAccessToken();
            setTimeout(() => {
                const tokenToUse = newToken || localStorage.getItem("access_token");
                if (tokenToUse) initializeChatWebSocket(tokenToUse);
                else window.location.href = "/sign-in";
            }, delay * 1000);
        } else {
            addMessage({
                time: new Date().toISOString(),
                address: "System",
                message: "Can't reconnect. Please refresh the page or login again.",
                isMyMessage: false
            });
        }
    };

    window.addEventListener('beforeunload', () => {
        if (currentSocket && currentSocket.readyState === WebSocket.OPEN) {
            try { currentSocket.close(1000, 'window unload'); } catch (e) { /* ignore */ }
        }
    });
}

async function refreshAccessToken() {
    const refreshToken = localStorage.getItem("refresh_token");
    if (!refreshToken) return null;

    try {
        const res = await fetch("/auth/refresh", {
            method: "POST",
            headers: {
                "Content-Type": "application/json"
            },
            body: JSON.stringify({ refresh_token: refreshToken })
        });

        const data = await res.json().catch(()=>null);
        if (res.ok && data && data.access_token) {
            localStorage.setItem("access_token", data.access_token);
            return data.access_token;
        } else {
            return null;
        }
    } catch (e) {
        console.error("Refresh failed:", e);
        return null;
    }
}

function logout() {
    localStorage.clear();
    window.location.href = "/sign-in";
}