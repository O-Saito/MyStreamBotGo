const logins = document.getElementById('logins');
const kickConnections = document.querySelector('#chat-connection-kick');
const twitchConnections = document.querySelector('#chat-connection-twitch');

const twitchLastEventsub = document.getElementById('twitch-last-eventsub');

const txtStreamGame = document.getElementById('stream-game')
const streamGameList = document.getElementById('stream-games')

const chatDiv = document.querySelector("#chat .content");
const eventDiv = document.getElementById('events');

let ws = new WebSocket(`ws://${location.host}/ws`);

let twitchData = null;
let emoteMap = {};

(async () => {
    emoteMap = await loadAllEmotes();
    console.log(emoteMap);
})()

console.log("Conectando ao WebSocket...");

const twitchNotificationHandlers = {
    "channel.follow": (data) => {
        eventDiv.innerHTML += `<div class="event"><span><span class="user">${data.payload.event.user_name}</span> seguiu ${twitchData == null || twitchData.userId == data.payload.event.broadcaster_user_id ? data.payload.event.broadcaster_user_name : 'você'}!</span></div>`;
    },
    "channel.raid": (data) => {
        eventDiv.innerHTML += `<div class="event"><span><span class="user">${data.payload.event.from_broadcaster_user_name}</span> está raidando ${twitchData == null || twitchData.userId == data.payload.event.to_broadcaster_user_id ? data.payload.event.to_broadcaster_user_name : 'você'} com ${data.payload.event.viewers} viewers!</span></div>`;
    },
};

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
        updateTwitchConnectionDetails(data.twitch);
        updateKickConnectionDetails(data.kick.connected_as);
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
        div.dataset.userId = msg.userId;
        div.dataset.channel = msg.channel;
        div.dataset.messageId = msg.messageId;
        div.dataset.user = msg.user;
        div.dataset.source = msg.source;

        let span = document.createElement("span");

        let source = document.createElement("span");
        let channel = document.createElement("span");
        let badges = document.createElement('span');
        let user = document.createElement('span');
        let message = document.createElement('span');

        source.className = "chat-source";
        channel.className = "chat-channel";
        badges.className = "chat-badges";
        user.className = "chat-user";
        message.className = "chat-message";


        span.append(source, channel, badges, user, message);

        source.textContent = msg.source;
        channel.textContent = msg.channel;
        user.textContent = msg.user;
        message.innerHTML = parseText(msg);

        if (msg.metadata) {
            user.style.color = msg.metadata.color;
            if (msg.metadata['display-name'])
                user.textContent = msg.metadata['display-name']
            if (msg.metadata['badges-info'])
                badges.innerHTML = Object.getOwnPropertyNames(msg.metadata['badges-info']).map(x => msg.metadata['badges-info'][x] && msg.metadata['badges-info'][x][0] != null && msg.metadata['badges-info'][x][0] != '' ? `<img src="${msg.metadata['badges-info'][x][0].image_url_1x}" alt="${msg.metadata['badges-info'][x][0].title}" />` : '').join('');

            if (msg.metadata['room']) {
                channel.innerHTML = `<img src="${msg.metadata['room'].profile_image_url}" alt="${msg.metadata['room'].display_name}" />`;
            }
            if (msg.metadata['source-room']) {
                channel.innerHTML = `<img src="${msg.metadata['source-room'].profile_image_url}" alt="${msg.metadata['source-room'].display_name}" />`;
            }

            if (msg.metadata["source-room-id"])
                div.dataset.sourceRoom = msg.metadata["source-room-id"];
        }
        if (msg.source == "twitch") {
            source.innerHTML = `<img src="https://assets.twitch.tv/assets/favicon-32-e29e246c157142c94346.png" />`
        }
        if (msg.source == "kick") {
            source.innerHTML = `<img src="https://kick.com/favicon.ico" />`
        }
        //span.textContent = `[${msg.source}] ${msg.channel ? `(${msg.channel})` : ""} ${msg.user || ""}: ${msg.message || msg.event_type}`;

        // criar botões de ação
        if (msg.source === "twitch") {
            let delBtn = document.createElement("button");
            delBtn.innerHTML = `<i class="fa fa-trash" aria-hidden="true"></i>`;
            delBtn.onclick = deleteTwitchMessage;
            let banBtn = document.createElement("button");
            banBtn.innerHTML = `<i class="fa fa-ban" aria-hidden="true"></i>`;
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
    'result-query-stream-games': function (data) {
        streamGameList.innerHTML = data.map(x => `<option data-id="${x.id}" data-name="${x.name}" >${x.name}</option>`).join('');
    },
    'twitch-eventsub-keepalive': function (data) {
        console.log('twitch-eventsub-keepalive', data);
        let date = new Date(data.metadata.message_timestamp);
        console.log(date);
        twitchLastEventsub.innerHTML = `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}:${date.getSeconds().toString().padStart(2, '0')}`;
    },
    'twitch-eventsub-notification': function (data) {
        if (twitchNotificationHandlers[data.metadata.subscription_type]){
            twitchNotificationHandlers[data.metadata.subscription_type](data);
            return;
        } 
        console.log('twitch-eventsub-notification não tratado!', data);
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
    const messageDiv = e.target.closest('.message');
    const user = messageDiv.dataset.user;
    const message = messageDiv.dataset.messageId;

    if (!user || !message) return;

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
        msg.classList.add('message-deleted');
        msg.querySelectorAll('button').forEach(x => { x.remove(); });
    }
}

function updateTwitchConnectionDetails(data) {
    twitchData = data;
    logins.querySelector('#login-twitch').innerHTML = data.userLogin == '' ? `<a href="/twitch/login" target="_blank">Logar Twitch</a>` : `Twitch: <img src="${data.profile_image_url}" /> ${data.userLogin}`;
    if (!data.userLogin || data.userLogin == '') {
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
    if (txtStreamGame.value == '' || txtStreamGame.value.length < 3) {
        return;
    }
    ws.send(JSON.stringify({ type: 'query-stream-game', data: { q: txtStreamGame.value } }));
});

function parseText(data) {
    let text = data.message;

    text = parseMessageWithEmotes(text, emoteMap);

    if (data.source == "twitch" && data.metadata) {
        let emotes = data.metadata.emotes.split('/')
        if (emotes && emotes[0] != '') {
            emotes = emotes.map(x => {
                let v = x.split(':');
                let id = v[0];
                let url = `https://static-cdn.jtvnw.net/emoticons/v2/${id}/default/light/1.0`;
                let indexes = v[1].split('-').map(y => parseInt(y));
                let name = data.message.slice(indexes[0], indexes[1] + 1);
                return {
                    id: id,
                    url: url,
                    name: name,
                    indexes: indexes,
                }
            });
        }

        for (let i = 0; i < emotes.length; i++) {
            const emote = emotes[i];
            text = text.replaceAll(emote.name, `<img src="${emote.url}" />`)
        }

    }


    return text;
}


async function loadBTTVEmotes(twitchId) {
    const [global, channel] = await Promise.all([
        fetch("https://api.betterttv.net/3/cached/emotes/global").then(r => r.json()),
        //fetch(`https://api.betterttv.net/3/cached/users/twitch/${twitchId}`).then(r => r.json())
    ]);

    const all = [
        ...global,
        //...(channel.channelEmotes || []), 
        //...(channel.sharedEmotes || [])
    ];
    const map = {};
    for (const e of all) {
        map[e.code] = `https://cdn.betterttv.net/emote/${e.id}/3x`;
    }
    return map;
}

async function loadFFZEmotes(login) {
    const [global, channel] = await Promise.all([
        fetch("https://api.frankerfacez.com/v1/set/global").then(r => r.json()),
        //fetch(`https://api.frankerfacez.com/v1/room/${login}`).then(r => r.json())
    ]);

    const sets = [
        ...(Object.values(global.sets || {})),
        //...(Object.values(channel.sets || {}))
    ];

    const map = {};
    for (const set of sets) {
        for (const e of set.emoticons) {
            const url = e.urls["4"] || e.urls["2"] || e.urls["1"];
            map[e.name] = url.startsWith("//") ? "https:" + url : url;
        }
    }
    return map;
}

async function load7TVEmotes(twitchId) {
    const [global, user] = await Promise.all([
        fetch("https://7tv.io/v3/emote-sets/global").then(r => r.json()),
        //fetch(`https://7tv.io/v3/users/twitch/${twitchId}`).then(r => r.json())
    ]);

    const all = [
        ...(global.emotes || []),
        //...((user.emote_set && user.emote_set.emotes) || [])
    ];

    const map = {};
    for (const e of all) {
        const base = e.data?.host?.url || e.host?.url;
        map[e.name] = `${base}/3x.webp`;
    }
    return map;
}

async function loadAllEmotes(twitchId, login) {
    const [bttv, ffz, stv] = await Promise.all([
        loadBTTVEmotes(twitchId),
        loadFFZEmotes(login),
        load7TVEmotes(twitchId)
    ]);

    return { ...bttv, ...ffz, ...stv };
}

function parseMessageWithEmotes(message, emoteMap) {
    return message
        .split(/\s+/)
        .map(word => {
            const url = emoteMap[word];
            if (url)
                return `<img src="${url}" alt="${word}" title="${word}" class="emote">`;
            return word;
        })
        .join(" ");
}