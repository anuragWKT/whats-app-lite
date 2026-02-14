let socket = null;
let currentUser = "";
let currentRoom = "";

const loginView = document.getElementById('login-view');
const chatView = document.getElementById('chat-view');
const messageLog = document.getElementById('message-log');
const statusSpan = document.getElementById('status');
const roomDisplay = document.getElementById('room-display');

function joinChat() {
    const usernameInput = document.getElementById('username').value.trim();
    const roomInput = document.getElementById('room').value.trim();

    if (!usernameInput || !roomInput) {
        alert("Please enter both username and room.");
        return;
    }

    currentUser = usernameInput;
    currentRoom = roomInput;

    loginView.style.display = 'none';
    chatView.style.display = 'flex';
    roomDisplay.textContent = `Room: ${currentRoom}`;

    connectWebSocket();
}

function connectWebSocket() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws?room=${currentRoom}&username=${currentUser}`;
    
    console.log("Connecting to:", wsUrl);
    socket = new WebSocket(wsUrl);

    socket.onopen = function(e) {
        statusSpan.textContent = "Connected";
    };

    socket.onmessage = function(event) {
        try {
            const msg = JSON.parse(event.data);
            appendMessage(msg);
        } catch (e) {
            console.log("Received non-JSON message:", event.data);
        }
    };

    socket.onclose = function(event) {
        statusSpan.textContent = "Disconnected";
        if (!event.wasClean) {
            alert("Connection lost. Please refresh.");
        }
    };

    socket.onerror = function(error) {
        console.log(`[error]`, error);
    };
}

function sendMessage() {
    const input = document.getElementById('msg-input');
    const rawText = input.value.trim();

    if (!rawText || !socket) return;

    let type = "message";
    let content = rawText;
    let recipient = ""; 

    if (rawText.startsWith("/msg ")) {
        const parts = rawText.split(" ");
        if (parts.length >= 3) {
            recipient = parts[1];
            content = parts.slice(2).join(" ");
        }
    }

    const messagePayload = {
        type: type,
        sender: currentUser,
        content: content,
        room: currentRoom,
        recipient: recipient
    };

    socket.send(JSON.stringify(messagePayload));
    input.value = "";
}

function appendMessage(msg) {
    const div = document.createElement('div');
    div.classList.add('message');

    if (msg.recipient && msg.recipient !== "") {
        div.classList.add('private-message');
        div.innerHTML = `
            <span class="sender-name"> Private from ${msg.sender} to ${msg.recipient}</span>
            ${msg.content}
        `;
        div.style.backgroundColor = "#fff3cd";
        div.style.border = "1px solid #ffeeba";
    } 

    else {
        if (msg.sender === currentUser) {
            div.classList.add('my-message');
            div.textContent = msg.content;
        } else {
            div.classList.add('other-message');
            div.innerHTML = `<span class="sender-name">${msg.sender}</span>${msg.content}`;
        }
    }

    messageLog.appendChild(div);
    messageLog.scrollTop = messageLog.scrollHeight;
}

document.getElementById('msg-input').addEventListener('keypress', function (e) {
    if (e.key === 'Enter') {
        sendMessage();
    }
});