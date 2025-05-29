const socket = new WebSocket('ws://' + window.location.host + '/ws');
const chat = document.getElementById('chat');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');

function addMessage(data) {
    const messageElement = document.createElement('div');
    messageElement.className = `message p-2 rounded ${data.isMyMessage ? 'bg-primary text-white' : 'bg-light'}`;

    messageElement.innerHTML = `
        <div class="address">${data.address }</div>
        <div class="time">${data.time}</div>
        <div class="text">${data.message}</div>
    `;

    chat.appendChild(messageElement);
    chat.scrollTop = chat.scrollHeight;
}

function sendMessage() {
    const message = messageInput.value.trim();
    if (message) {
        socket.send(JSON.stringify({ message }));
        addMessage({
            time: new Date().toLocaleTimeString(),
            address: "You",
            message: message,
            isMyMessage: true
        });
        messageInput.value = '';
    }
}

sendButton.addEventListener('click', sendMessage);
messageInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter') sendMessage();
});

socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (!data.isMyMessage) {
        addMessage({
            time: data.time,
            address: data.address,
            message: data.message,
            isMyMessage: false
        });
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