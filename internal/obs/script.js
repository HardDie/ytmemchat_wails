/* <====== Shared state ======> */
var appClosed = false;

/* Pages set these before this file runs (var / function, shared across classic scripts). */
/* socketReadyLog — string logged on open. */
/* handleSocketMessage — WebSocket onmessage. */
/* onConnectionLost — optional; overlay tears down playback. */

/* <====== Errors / recover ======> */
const RECOVER_COOLDOWN_MS = 2000;
let recovering = false;

function recoverFromUnhandledError(label, detail) {
	console.error(label, detail);
	if (recovering) return;
	recovering = true;
	try {
		if (typeof onConnectionLost === 'function') {
			onConnectionLost();
		}
		connectWebSocket();
	} catch (e) {
		console.error(e);
	}
	setTimeout(function() { recovering = false; }, RECOVER_COOLDOWN_MS);
}

window.onerror = function(msg, url, line) {
	recoverFromUnhandledError("CRITICAL ERROR:", msg + " at " + line);
	return true;
};

window.onunhandledrejection = function(event) {
	event.preventDefault();
	recoverFromUnhandledError("PROMISE REJECTION:", event.reason);
};

/* <====== Connection ======> */
const statusEl = document.getElementById('connection-status');
const STATUS_LOST = '⚠️ WEB SOCKET DISCONNECTED - RETRYING...';
const STATUS_CLOSED = 'ytmemchat is closed — reconnecting when it starts…';
const RETRY_INTERVAL = 500;
const CONNECTING_TIMEOUT_MS = 2500;
let reconnectTimer = null;
let connectingTimer = null;
let websocket = null;
const wsUri = ((window.location.protocol === 'https:') ? 'wss://' : 'ws://') + window.location.host + window.location.pathname.replace(/\/$/, '') + '/ws';

function showConnectionStatus() {
	statusEl.textContent = appClosed ? STATUS_CLOSED : STATUS_LOST;
	statusEl.classList.toggle('app-closed', appClosed);
	statusEl.style.display = 'block';
}

function scheduleReconnect() {
	if (reconnectTimer) return;
	reconnectTimer = setTimeout(function() {
		reconnectTimer = null;
		connectWebSocket();
	}, RETRY_INTERVAL);
}

function clearConnectingTimer() {
	if (connectingTimer) {
		clearTimeout(connectingTimer);
		connectingTimer = null;
	}
}

function connectWebSocket() {
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
	clearConnectingTimer();
	if (websocket) {
		console.log("Cleaning up old WebSocket before reconnecting...");
		websocket.onopen = null;
		websocket.onmessage = null;
		websocket.onerror = null;
		websocket.onclose = null;
		websocket.close();
		websocket = null;
	}
	websocket = new WebSocket(wsUri);
	const thisSocket = websocket;
	connectingTimer = setTimeout(function() {
		connectingTimer = null;
		if (websocket === thisSocket && thisSocket.readyState === WebSocket.CONNECTING) {
			thisSocket.close();
		}
	}, CONNECTING_TIMEOUT_MS);
	websocket.onopen = function() {
		clearConnectingTimer();
		console.log(socketReadyLog);
		appClosed = false;
		statusEl.classList.remove('app-closed');
		statusEl.style.display = 'none';
	};
	websocket.onmessage = handleSocketMessage;
	websocket.onclose = function() {
		clearConnectingTimer();
		websocket = null;
		if (typeof onConnectionLost === 'function') {
			onConnectionLost();
		}
		showConnectionStatus();
		console.log("Connection lost. Retrying in " + RETRY_INTERVAL + "ms...");
		scheduleReconnect();
	};
	websocket.onerror = function(err) {
		console.error("Socket encountered error: ", err);
		if (websocket) websocket.close();
	};
}

connectWebSocket();
