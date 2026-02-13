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
    const wsUrl = `${protocol}//${window.location.host}/ws?room=${currentRoom}`;
    
    console.log("Connecting to:", wsUrl);
    socket = new WebSocket(wsUrl);

    socket.onopen = function(e) {
        statusSpan.textContent = "Connected";
        const joinMsg = {
            type: "system",
            content: `${currentUser} joined the room`,
            sender: "System",
            room: currentRoom
        };
        // socket.send(JSON.stringify(joinMsg)); 
    };

    socket.onmessage = function(event) {
        try {
            const msg = JSON.parse(event.data);
            appendMessage(msg);
        } catch (e) {
            console.log("Received non-JSON message:", event.data);
            appendMessage({
                content: event.data,
                sender: "Unknown",
                type: "message"
            });
        }
    };

    socket.onclose = function(event) {
        statusSpan.textContent = "Disconnected";
        if (event.wasClean) {
            console.log(`[close] Connection closed cleanly, code=${event.code}`);
        } else {
            console.log('[close] Connection died');
            alert("Connection lost. Please refresh.");
        }
    };

    socket.onerror = function(error) {
        console.log(`[error]`, error);
    };
}

function sendMessage() {
    const input = document.getElementById('msg-input');
    const content = input.value.trim();

    if (!content || !socket) return;

    const messagePayload = {
        type: "message",
        sender: currentUser,
        content: content,
        room: currentRoom
    };

    socket.send(JSON.stringify(messagePayload));

    input.value = "";
    
}

function appendMessage(msg) {
    const div = document.createElement('div');
    div.classList.add('message');

    if (msg.sender === currentUser) {
        div.classList.add('my-message');
        div.textContent = msg.content;
    } else {
        div.classList.add('other-message');
        div.innerHTML = `<span class="sender-name">${msg.sender}</span>${msg.content}`;
    }

    messageLog.appendChild(div);
    messageLog.scrollTop = messageLog.scrollHeight;
}
document.getElementById('msg-input').addEventListener('keypress', function (e) {
    if (e.key === 'Enter') {
        sendMessage();
    }
});