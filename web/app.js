let socket = null;
let currentUser = "";
let currentRoom = "";
let currentPassword = "";
let typingTimeout = null;
let isTyping = false;
let manualClose = false;
const typingIndicators = {};

const loginView = document.getElementById("login-view");
const roomsView = document.getElementById("rooms-view");
const roomsList = document.getElementById("rooms-list");
const chatView = document.getElementById("chat-view");
const messageLog = document.getElementById("message-log");
const statusSpan = document.getElementById("status");
const roomDisplay = document.getElementById("room-display");
const errorMessage = document.getElementById("error-message");

function showError(message) {
  errorMessage.textContent = message;
  errorMessage.classList.add("show");
  setTimeout(() => {
    errorMessage.classList.remove("show");
  }, 4000);
}

function toggleForm() {
  const loginForm = document.getElementById("login-form");
  const registerForm = document.getElementById("register-form");

  loginForm.classList.toggle("active");
  registerForm.classList.toggle("active");

  errorMessage.classList.remove("show");
}

async function handleRegister() {
  const username = document.getElementById("register-username").value.trim();
  const password = document.getElementById("register-password").value.trim();
  const confirmPassword = document
    .getElementById("register-confirm-password")
    .value.trim();

  if (!username || !password || !confirmPassword) {
    showError("Please fill in all fields");
    return;
  }

  if (password !== confirmPassword) {
    showError("Passwords do not match");
    return;
  }

  if (password.length < 4) {
    showError("Password must be at least 4 characters");
    return;
  }

  try {
    const response = await fetch("/register", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ username, password }),
    });

    if (response.ok) {
      showError("Registration successful! Please login.");
      setTimeout(() => {
        toggleForm();
        document.getElementById("register-username").value = "";
        document.getElementById("register-password").value = "";
        document.getElementById("register-confirm-password").value = "";
      }, 1500);
    } else {
      const data = await response.text();
      showError(data || "Registration failed");
    }
  } catch (error) {
    showError("Error: " + error.message);
  }
}

async function handleLogin() {
  const username = document.getElementById("login-username").value.trim();
  const password = document.getElementById("login-password").value.trim();

  if (!username || !password) {
    showError("Please fill in all fields");
    return;
  }

  currentUser = username;
  currentPassword = password;

  const ok = await fetchRooms();
  if (!ok) {
    return;
  }

  loginView.style.display = "none";
  roomsView.style.display = "flex";
  chatView.style.display = "none";
}

async function fetchRooms() {
  try {
    const response = await fetch(
      `/rooms?username=${encodeURIComponent(currentUser)}&password=${encodeURIComponent(currentPassword)}`,
    );
    if (!response.ok) {
      showError("Invalid username or password. Please try again.");
      return false;
    }
    const data = await response.json();
    renderRooms(data.rooms || []);
    return true;
  } catch (error) {
    showError("Failed to load rooms. Please try again.");
    return false;
  }
}

function renderRooms(rooms) {
  roomsList.innerHTML = "";

  if (rooms.length === 0) {
    const emptyState = document.createElement("div");
    emptyState.classList.add("room-card");
    emptyState.innerHTML = `<h4>No rooms yet</h4><p>Create your first room!</p>`;
    roomsList.appendChild(emptyState);
    return;
  }

  rooms.forEach((room) => {
    const card = document.createElement("div");
    card.classList.add("room-card");
    card.innerHTML = `<h4>${room}</h4>`;

    const button = document.createElement("button");
    button.textContent = "Join";
    button.onclick = () => joinRoom(room);
    card.appendChild(button);

    roomsList.appendChild(card);
  });
}

function createRoom() {
  const input = document.getElementById("new-room-name");
  const roomName = input.value.trim();
  if (!roomName) {
    showError("Please enter a room name");
    return;
  }
  input.value = "";
  joinRoom(roomName);
}

function joinRoom(roomName) {
  currentRoom = roomName;
  roomsView.style.display = "none";
  chatView.style.display = "flex";
  roomDisplay.textContent = `Room: ${currentRoom}`;
  statusSpan.textContent = "Connecting...";
  connectWebSocket();
}

function leaveRoom() {
  if (socket) {
    manualClose = true;
    socket.close();
    socket = null;
  }
  messageLog.innerHTML = "";
  roomsView.style.display = "flex";
  chatView.style.display = "none";
  fetchRooms();
}

function connectWebSocket() {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.host}/ws?room=${currentRoom}&username=${currentUser}&password=${encodeURIComponent(currentPassword)}`;

  console.log("Connecting to:", wsUrl);
  socket = new WebSocket(wsUrl);
  let connectionAttempted = false;

  socket.onopen = function (e) {
    console.log("WebSocket opened successfully");
    connectionAttempted = true;
    statusSpan.textContent = "Connected";
  };

  socket.onmessage = function (event) {
    try {
      const msg = JSON.parse(event.data);
      if (msg.type === "typing") {
        handleTypingIndicator(msg);
      } else {
        appendMessage(msg);
      }
    } catch (e) {
      console.log("Received non-JSON message:", event.data);
    }
  };

  socket.onclose = function (event) {
    console.log(
      "WebSocket closed. Clean:",
      event.wasClean,
      "Code:",
      event.code,
    );
    statusSpan.textContent = "Disconnected";

    if (manualClose) {
      manualClose = false;
      return;
    }

    if (!event.wasClean || event.code !== 1000) {
      roomsView.style.display = "flex";
      chatView.style.display = "none";
      if (!connectionAttempted) {
        showError("Invalid username or password. Please try again.");
        loginView.style.display = "flex";
        roomsView.style.display = "none";
      } else {
        showError("Connection lost. Please rejoin the room.");
      }
    }
  };

  socket.onerror = function (error) {
    console.log(`[WebSocket error]`, error);
    statusSpan.textContent = "Connection Error";
    if (!connectionAttempted) {
      loginView.style.display = "flex";
      roomsView.style.display = "none";
      chatView.style.display = "none";
      showError("Invalid username or password. Please try again.");
    }
  };
}

function sendTypingIndicator() {
  if (!socket || !currentUser) return;

  const typingPayload = {
    type: "typing",
    sender: currentUser,
    room: currentRoom,
    content: "",
  };

  socket.send(JSON.stringify(typingPayload));
}

function handleTypingIndicator(msg) {
  if (msg.sender === currentUser) return;

  const typingKey = msg.sender;
  const existingTyping = document.getElementById(`typing-${typingKey}`);

  if (existingTyping) {
    existingTyping.remove();
  }

  const typingDiv = document.createElement("div");
  typingDiv.id = `typing-${typingKey}`;
  typingDiv.classList.add("typing-indicator");
  typingDiv.innerHTML = `<span class="sender-name">${msg.sender}</span><span class="dots">typing<span>.</span><span>.</span><span>.</span></span>`;
  messageLog.appendChild(typingDiv);
  messageLog.scrollTop = messageLog.scrollHeight;

  if (typingIndicators[typingKey]) {
    clearTimeout(typingIndicators[typingKey]);
  }
  typingIndicators[typingKey] = setTimeout(() => {
    const el = document.getElementById(`typing-${typingKey}`);
    if (el) el.remove();
    delete typingIndicators[typingKey];
  }, 3000);
}

function sendMessage() {
  const input = document.getElementById("msg-input");
  const rawText = input.value.trim();

  if (!rawText || !socket) return;

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
    type: "message",
    sender: currentUser,
    content: content,
    room: currentRoom,
    recipient: recipient,
  };

  socket.send(JSON.stringify(messagePayload));
  input.value = "";
  isTyping = false;

  const typingKey = currentUser;
  if (typingIndicators[typingKey]) {
    clearTimeout(typingIndicators[typingKey]);
    delete typingIndicators[typingKey];
  }
}

function appendMessage(msg) {
  const typingKey = msg.sender;
  const existingTyping = document.getElementById(`typing-${typingKey}`);
  if (existingTyping) {
    existingTyping.remove();
  }
  if (typingIndicators[typingKey]) {
    clearTimeout(typingIndicators[typingKey]);
    delete typingIndicators[typingKey];
  }

  const div = document.createElement("div");
  div.classList.add("message");

  if (msg.recipient && msg.recipient !== "") {
    div.classList.add("private-message");
    div.innerHTML = `
            <span class="sender-name"> Private from ${msg.sender} to ${msg.recipient}</span>
            ${msg.content}
        `;
    div.style.backgroundColor = "#fff3cd";
    div.style.border = "1px solid #ffeeba";
  } else {
    if (msg.sender === currentUser) {
      div.classList.add("my-message");
      div.textContent = msg.content;
    } else {
      div.classList.add("other-message");
      div.innerHTML = `<span class="sender-name">${msg.sender}</span>${msg.content}`;
    }
  }

  messageLog.appendChild(div);
  messageLog.scrollTop = messageLog.scrollHeight;
}

const msgInput = document.getElementById("msg-input");

msgInput.addEventListener("input", function () {
  if (!isTyping && socket) {
    isTyping = true;
    sendTypingIndicator();
  }

  clearTimeout(typingTimeout);
  typingTimeout = setTimeout(() => {
    isTyping = false;
  }, 1000);
});

msgInput.addEventListener("keypress", function (e) {
  if (e.key === "Enter") {
    sendMessage();
  }
});