const ChatModule = (() => {
    let ws = null;
    let currentGroupId = null;
    let currentUserId = null;
    let reconnectTimeout = null;
    let reconnectAttempts = 0;

    function getToken() {
        return localStorage.getItem(TOKEN_KEY);
    }

    function getUserId() {
        return state.user?.id || null;
    }

    function connect(groupId) {
        disconnect();
        currentGroupId = groupId;
        currentUserId = getUserId();
        reconnectAttempts = 0;

        const token = getToken();
        if (!token) return;

        const wsBase = API_BASE_URL.replace(/^http/, 'ws');
        const wsUrl = `${wsBase}/ws/groups/${encodeURIComponent(groupId)}?token=${encodeURIComponent(token)}`;

        try {
            ws = new WebSocket(wsUrl);
        } catch (_) {
            renderConnectionStatus('error');
            return;
        }

        ws.onmessage = (event) => {
            let data;
            try {
                data = JSON.parse(event.data);
            } catch (_) {
                return;
            }
            handleServerMessage(data);
        };

        ws.onerror = () => {
            renderConnectionStatus('error');
        };

        ws.onopen = () => {
            clearTimeout(reconnectTimeout);
            reconnectAttempts = 0;
            renderConnectionStatus('conectado');
        };

        ws.onclose = () => {
            renderConnectionStatus('desconectado');
            reconnectAttempts++;
            if (currentGroupId && reconnectAttempts <= 5) {
                reconnectTimeout = setTimeout(() => {
                    if (currentGroupId) connect(currentGroupId);
                }, 3000);
            }
        };
    }

    function disconnect() {
        clearTimeout(reconnectTimeout);
        currentGroupId = null;
        if (ws) {
            ws.onopen = null;
            ws.onmessage = null;
            ws.onerror = null;
            ws.onclose = null;
            ws.close();
            ws = null;
        }
    }

    function sendMessage(content) {
        if (!ws || ws.readyState !== WebSocket.OPEN) return false;
        const trimmed = String(content).trim();
        if (!trimmed) return false;
        ws.send(JSON.stringify({ type: 'message', content: trimmed }));
        return true;
    }

    function sendReadAck() {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'read_ack' }));
        }
    }

    function handleServerMessage(data) {
        if (data.type === 'history') {
            renderHistory(Array.isArray(data.messages) ? data.messages : []);
            sendReadAck();
        } else if (data.type === 'message') {
            if (data.message) appendMessage(data.message);
            sendReadAck();
        } else if (data.type === 'system') {
            if (data.content) appendSystemMessage(data.content);
        }
    }

    function renderHistory(messages) {
        const container = document.getElementById('chatMessages');
        if (!container) return;
        container.innerHTML = '';
        messages.forEach(msg => appendMessage(msg, false));
        scrollToBottom();
    }

    function appendMessage(msg, scroll) {
        if (scroll === undefined) scroll = true;
        const container = document.getElementById('chatMessages');
        if (!container) return;

        const isMine = msg.senderId === currentUserId;
        let time = '';
        if (msg.createdAt) {
            try {
                time = new Date(msg.createdAt).toLocaleTimeString('es-ES', { hour: '2-digit', minute: '2-digit' });
            } catch (_) {}
        }

        const el = document.createElement('div');
        el.className = 'chat-msg ' + (isMine ? 'chat-msg--mine' : 'chat-msg--theirs');

        let senderHtml = '';
        if (!isMine && msg.senderName) {
            senderHtml = '<span class="chat-msg__sender">' + escapeHtml(msg.senderName) + '</span>';
        }

        el.innerHTML = senderHtml +
            '<div class="chat-msg__bubble">' +
                '<span class="chat-msg__text">' + escapeHtml(msg.content || '') + '</span>' +
                (time ? '<span class="chat-msg__time">' + time + '</span>' : '') +
            '</div>';

        container.appendChild(el);
        if (scroll) scrollToBottom();
    }

    function appendSystemMessage(content) {
        const container = document.getElementById('chatMessages');
        if (!container) return;
        const el = document.createElement('div');
        el.className = 'chat-msg chat-msg--system';
        el.textContent = content;
        container.appendChild(el);
        scrollToBottom();
    }

    function renderConnectionStatus(status) {
        const el = document.getElementById('chatStatus');
        if (!el) return;
        if (status === 'conectado') {
            el.textContent = '';
        } else if (status === 'desconectado') {
            el.textContent = 'Reconectando...';
        } else {
            el.textContent = 'Error de conexion';
        }
        el.className = 'chat-status chat-status--' + status;
    }

    function scrollToBottom() {
        const container = document.getElementById('chatMessages');
        if (container) container.scrollTop = container.scrollHeight;
    }

    function escapeHtml(str) {
        return String(str)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#39;');
    }

    return { connect, disconnect, sendMessage };
})();
