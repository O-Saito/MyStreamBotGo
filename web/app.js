const logins = document.getElementById('logins');
const chatConnections = document.getElementById('chat-connections');
const kickConnections = chatConnections.querySelector('#chat-connection-kick');
const twitchConnections = chatConnections.querySelector('#chat-connection-twitch');

const txtStreamGame = document.getElementById('stream-game')
const streamGameList = document.getElementById('stream-games')

const chatDiv = document.getElementById("chat");
let ws = new WebSocket(`ws://${location.host}/ws`);

console.log("Conectando ao WebSocket...");

const handlers = {
    'init': function (data) {
        console.log(data);
        if (data.twitch_connected_chat && data.twitch_connected_chat.length > 0) {
            for (let i = 0; i < data.twitch_connected_chat.length; i++) {
                const c = data.twitch_connected_chat[i];
                handlers['twitch-chat-connection']({ id: c, name: c });
            }
        }
        if (data.kick_connected_chat && data.kick_connected_chat.length > 0) {
            for (let i = 0; i < data.kick_connected_chat.length; i++) {
                const c = data.kick_connected_chat[i];
                handlers['kick-chat-connection']({ id: c.ID, name: c.Slug });
            }
        }
        updateTwitchConnectionDetails(data.twitch.connected_as)
        updateKickConnectionDetails(data.kick.connected_as)

    },
    'twitch-connection': updateTwitchConnectionDetails,
    'kick-connection': updateKickConnectionDetails,
    'twitch-chat-connection': function (data) {
        const li = document.createElement('li');
        li.innerHTML = `(${data.id}) ${data.name}`;
        twitchConnections.querySelector('.list').append(li);
    },
    'kick-chat-connection': function (data) {
        const li = document.createElement('li');
        li.innerHTML = `(${data.id}) ${data.name}`;
        kickConnections.querySelector('.list').append(li);
    },
    "user-message": function (msg) {
        console.log("Nova mensagem de usuário:", msg);
        let div = document.createElement("div");
        div.className = "message";
        div.dataset.userId = msg.metadata['userId'];
        div.dataset.channel = msg.channel;
        div.dataset.messageId = msg.messageId;
        div.dataset.user = msg.user;
        div.dataset.source = msg.source;

        let span = document.createElement("span");
        span.textContent = `[${msg.source}] ${msg.channel ? `(${msg.channel})` : ""} ${msg.user || ""}: ${msg.message || msg.event_type}`;

        // criar botões de ação
        if (msg.source === "twitch") {
            let delBtn = document.createElement("button");
            delBtn.textContent = "Excluir";
            delBtn.onclick = deleteTwitchMessage;
            let banBtn = document.createElement("button");
            banBtn.textContent = "Banir";
            banBtn.onclick = banTwitchUser;

            div.appendChild(delBtn);
            //div.appendChild(banBtn);
        }

        if (msg.source == "kick") {
            let delBtn = document.createElement("button");
            delBtn.textContent = "Excluir";
            delBtn.onclick = deleteKickMessage;
            //div.appendChild(delBtn);
        }

        div.appendChild(span);
        chatDiv.appendChild(div);
        chatDiv.scrollTop = chatDiv.scrollHeight;
    },
    'user-message-update': function (data) {
        updateMessageDelete(data?.metadata['target-msg-id']);
    },
    'user-message-delete': function (data) {
        updateMessageDelete(data);
    },
    'result-query-stream-games': function(data) {
        streamGameList.innerHTML = data.map(x => `<option data-id="${x.id}" data-name="${x.name}" >${x.name}</option>`).join('');
    }
}

ws.onmessage = function (e) {
    console.log("Mensagem recebida:", e.data);
    const data = JSON.parse(e.data);
    if (!handlers[data.type]) {
        console.log("Handler não encontrado para o tipo:", data.type);
        return;
    }
    handlers[data.type](data.data);
};

ws.onopen = function () {
    console.log("Conectado ao WebSocket.");
    ws.send("init");
}

// TWITCH FUNCS

function connectToTwitchChat(slug) {
    ws.send(JSON.stringify({ type: 'connect-chat-twitch', data: { channel: slug } }))
}

function deleteTwitchMessage(e) {
    e.preventDefault();
    const messageDiv = e.target.parentElement;
    const user = messageDiv.dataset.user;
    const message = messageDiv.dataset.messageId;

    fetch("/admin/delete/twitch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user, message })
    }).then(res => console.log("Mensagem excluída:", res.status));
}

function banTwitchUser(e) {
    e.preventDefault();
    const messageDiv = e.target.parentElement;
    const user = messageDiv.dataset.user;
    const message = messageDiv.dataset.messageId;

    fetch("/admin/ban/twitch", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ user })
    }).then(res => console.log("Usuário banido:", user));
}

// KICK FUNCS

function connectToKickChat(slug) {
    fetch(`https://kick.com/api/v2/channels/${slug}`).then(x => x.json()).then(d => {
        console.log(d);
        if (d.chatroom) {
            ws.send(JSON.stringify({ type: 'connect-chat-kick', data: { channel: slug, roomId: d.chatroom.id.toString() } }))
        }
    });
}

function deleteKickMessage(e) {
    e.preventDefault();

    //https://kick.com/api/v2/chatrooms/76514426/messages/92504d5f-a107-4c70-85bd-315eb2b88dec
    const messageDiv = e.target.parentElement;
    const user = messageDiv.dataset.user;
    const message = messageDiv.dataset.messageId;
    const channel = messageDiv.dataset.channel;

    fetch(`https://kick.com/api/v2/chatrooms/${channel.split('.')[1]}/messages/${message}`, {
        method: "DELETE",
        headers: {
            "Authorization": "Bearer MTLJYZUWMZETMJAYYS0ZOWU1LTK4MJYTNGZIMGRHODK4NTK2",
            "Content-Type": "application/json"
        }
    }).then(x => x.json()).then(d => {
        console.log(d);

    });
}


// VISUALS

document.getElementById('btn-mandar-msg').onclick = function (e) {
    e.preventDefault();
    const t = prompt("Informe a mensagem a ser enviada para o canal principal de todas plataformas conectadas: ");
    if (!t || t == '') return;
    ws.send(JSON.stringify({ type: 'send-chat-message', data: { text: t } }))
}

function updateMessageDelete(messageId) {
    const msg = document.querySelector(`div[data-message-id="${messageId}"]`);
    if (msg) {
        msg.style.textDecoration = 'line-through';
        msg.querySelectorAll('button').forEach(x => { x.remove(); });
    }
}

function updateTwitchConnectionDetails(user) {
    logins.querySelector('#login-twitch').innerHTML = user == '' ? `<a href="/twitch/login" target="_blank">Logar Twitch</a>` : `Twitch: ${user}`;
    if (!user || user == '') {
        twitchConnections.style.display = 'none';
        return;
    }

    twitchConnections.style.display = 'block';
    twitchConnections.querySelector('.btn-novo-chat').onclick = function (e) {
        e.preventDefault();
        const t = prompt('Informe o usuário Twitch');
        if (t == '' || !t) return;
        connectToTwitchChat(t)
    }
}

function updateKickConnectionDetails(user) {
    logins.querySelector('#login-kick').innerHTML = user == '' ? `<a href="/kick/login" target="_blank">Logar Kick</a>` : `Kick: ${user}`;
    if (!user || user == '') {
        kickConnections.style.display = "none";
        return;
    }
    // tenta conectar ao chat da kick
    if (kickConnections.querySelectorAll('.list li').length == 0)
        connectToKickChat(user);
    kickConnections.style.display = "block";

    kickConnections.querySelector('.btn-novo-chat').onclick = function (e) {
        e.preventDefault();
        const t = prompt('Informe o usuário Kick');
        if (t == '' || !t) return;
        connectToKickChat(t);
    }
}

txtStreamGame.addEventListener('input', function (e) {
    console.log('input', txtStreamGame.value);
    if(txtStreamGame.value == '' || txtStreamGame.value.length < 3) {
        return;
    }
    ws.send(JSON.stringify({ type: 'query-stream-game', data: { q: txtStreamGame.value } }));
});
