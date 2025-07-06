let currentSocket = null;

document.addEventListener("DOMContentLoaded", () => {
    const accessToken = localStorage.getItem("access_token");
    if (accessToken) {
        initializeChatWebSocket(accessToken);
    } else {
        alert("No access token. Please login.");
        window.location.href = "/sign-in";
    }
});

function initializeChatWebSocket(accessToken) {
    if (currentSocket) {
        currentSocket.close();
    }

    const socket = new WebSocket(`ws://${window.location.host}/ws?token=${encodeURIComponent(accessToken)}`);
    currentSocket = socket;

    const chat = document.getElementById('chat');
    const messageInput = document.getElementById('messageInput');
    const sendButton = document.getElementById('sendButton');

    function addMessage({ time, address, message, isMyMessage }) {
        const msg = document.createElement('div');
        msg.className = `message p-2 rounded ${isMyMessage ? 'bg-primary' : 'bg-light'}`;
        msg.innerHTML = `
            <div><strong>${address}</strong> (${time})</div>
            <div>${message}</div>
        `;
        chat.appendChild(msg);
        chat.scrollTop = chat.scrollHeight;
    }

    function sendMessage() {
        const text = messageInput.value.trim();
        if (!text) return;

        socket.send(JSON.stringify({ message: text }));
        addMessage({
            time: new Date().toLocaleTimeString(),
            address: "You",
            message: text,
            isMyMessage: true
        });

        messageInput.value = '';
    }

    sendButton.onclick = sendMessage;
    messageInput.onkeypress = (e) => {
        if (e.key === "Enter") sendMessage();
    };

    socket.onmessage = (event) => {
        try {
            const data = JSON.parse(event.data);
            if (!data.isMyMessage) {
                addMessage({
                    time: data.time,
                    address: data.address,
                    message: data.message,
                    isMyMessage: false
                });
            }
        } catch (e) {
            console.warn("Non-JSON message:", event.data);
        }
    };

    socket.onerror = async () => {
        console.error("WebSocket error. Trying to refresh token...");

        const newToken = await refreshAccessToken();
        if (newToken) {
            setTimeout(() => initializeChatWebSocket(newToken), 1000);
        } else {
            addMessage({
                time: new Date().toLocaleTimeString(),
                address: "System",
                message: "Session expired. Please login again.",
                isMyMessage: false
            });
            localStorage.removeItem("access_token");
            localStorage.removeItem("refresh_token");
            setTimeout(() => window.location.href = "/sign-in", 1500);
        }
    };

    socket.onclose = () => {
        addMessage({
            time: new Date().toLocaleTimeString(),
            address: "System",
            message: "Connection closed",
            isMyMessage: false
        });
    };
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

        const data = await res.json();
        if (res.ok) {
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
