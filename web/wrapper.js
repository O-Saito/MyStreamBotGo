//@ts-check

/**
 * @typedef SocketMessage
 * @property {string} type
 * @property {string} filter
 * @property {any} data
 * 
 * @typedef EventSubcriber
 * @property {Function} on
 * @property {Function} send
 */

/** @type {Object.<string, string>} */
let emoteMap = {};
(async () => {
    emoteMap = await loadAllEmotes();
    console.log(emoteMap);
})()

const ws = new WebSocket(`ws://${location.host}/ws`);

/** @type {Object.<string, Array<Function>>} */
const handlers = {
    'init': [
        /** @param {any} data */
        (data) => {
            if (!data.custom_events_modules || data.custom_events_modules.length == 0)
                return;
            data.custom_events_modules.forEach((/** @type {{ name: string; }} */ x) => {
                eventsToSubscribe.push(x.name);
            });
        },
    ],
    'twitch-eventsub-notification': [
        /** @param {any} data */
        (data) => {
            if (twitchNotificationHandlers[data.metadata.subscription_type]) {
                twitchNotificationHandlers[data.metadata.subscription_type].forEach(f => { f(data); });
                return;
            }
            console.log('twitch-eventsub-notification não tratado!', data);
        },
    ],
}

/** @type {Object.<string, Array<Function>>} */
const twitchNotificationHandlers = {};

/** @type {Object.<string, Object.<string, Array<Function>>>} */
const eventsHandlers = {};

/** @type {Array<string>} eventsToSubscribe */
const eventsToSubscribe = [];

let connectCalled = false;

/**
 * 
 * @param {MessageEvent<any>} e 
 * @returns 
 */
ws.onmessage = function (e) {
    console.log("Mensagem recebida:", e.data);

    /** @type {SocketMessage} */
    const data = JSON.parse(e.data);

    if (data.filter && data.filter != "" && eventsHandlers[data.filter] && eventsHandlers[data.filter][data.type]) {
        eventsHandlers[data.filter][data.type].forEach(f => {
            f(data.data);
        });
        return;
    }

    if (!handlers[data.type]) {
        console.log("Handler não encontrado para o tipo:", data.type);
        return;
    }

    handlers[data.type].forEach(f => {
        f(data.data);
    });
};

/**
 * @param {string} [twitchId] 
 * @returns {Promise<Object.<string, string>>}
 */
async function loadBTTVEmotes(twitchId = "") {
    let url = "https://api.betterttv.net/3/cached/emotes/global";
    if (twitchId != "") {
        url = `https://api.betterttv.net/3/cached/users/twitch/${twitchId}`
    }
    const emotes = await fetch(url).then(r => r.json());

    /** @type {Object.<string, string>} */
    const map = {};
    for (const e of emotes) {
        map[e.code] = `https://cdn.betterttv.net/emote/${e.id}/3x`;
    }
    return map;
}

/**
 * @param {string} [login] 
 * @returns {Promise<Object.<string, string>>}
 */
async function loadFFZEmotes(login = "") {
    let url = "https://api.frankerfacez.com/v1/set/global";
    if (login != "") {
        url = `https://api.frankerfacez.com/v1/room/${login}`
    }
    const emotes = await fetch(url).then(r => r.json());

    /** @type {Object.<string, string>} */
    const map = {};
    for (const set of Object.values(emotes.sets || {})) {
        for (const e of set.emoticons) {
            const url = e.urls["4"] || e.urls["2"] || e.urls["1"];
            map[e.name] = url.startsWith("//") ? "https:" + url : url;
        }
    }
    return map;
}

/**
 * @param {string} [twitchId] 
 * @returns {Promise<Object.<string, string>>}
 */
async function load7TVEmotes(twitchId = "") {
    let url = "https://7tv.io/v3/emote-sets/global";
    if (twitchId != "") {
        url = `https://7tv.io/v3/users/twitch/${twitchId}`
    }
    const emotes = await fetch(url).then(r => r.json());

    /** @type {Object.<string, string>} */
    const map = {};
    for (const e of emotes.emotes) {
        const base = e.data?.host?.url || e.host?.url;
        map[e.name] = `${base}/3x.webp`;
    }
    return map;
}

/**
 * @param {string} [twitchId] 
 * @param {string} [login] 
 * @returns {Promise<Object.<string, string>>}
 */
async function loadAllEmotes(twitchId, login) {
    const [bttv, ffz, stv] = await Promise.all([
        loadBTTVEmotes(twitchId),
        loadFFZEmotes(login),
        load7TVEmotes(twitchId)
    ]);

    return { ...bttv, ...ffz, ...stv };
}

/**
 * 
 * @param {string} message 
 * @param {Object.<string, string>} emoteMap 
 * @returns {string}
 */
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

export default {
    connect: () => {
        if(connectCalled) {
            return;
        }
        connectCalled = true;
        ws.onopen = function () {
            console.log("Conectado ao WebSocket.");
            ws.send("init");
        }
    },
    /**
     * @param {string} event 
     * @returns {EventSubcriber | null}
     */
    subscribe: (event) => {
        if (eventsToSubscribe.find(x => x == event) == null) {
            console.error(`Evento não encontrado ${event}`);
            return null;
        }

        ws.send(JSON.stringify({ type: "upgrade-conn", data: { "conn": event } }));

        if (!eventsHandlers[event]) eventsHandlers[event] = {};

        return {
            /**
             * @param {string} type 
             * @param {Function} f 
             */
            on: (type, f) => {
                if (!eventsHandlers[event][type]) eventsHandlers[event][type] = [];
                eventsHandlers[event][type].push(f);
            },
            /**
             * @param {string} type 
             * @param {any} data 
             */
            send: (type, data) => {
                ws.send(JSON.stringify({ type: type, filter: event, data: data }));
            }
        };
    },
    /**
     * @param {string} eventType 
     * @param {Object} data 
     */
    send: (eventType, data) => {
        ws.send(JSON.stringify({ type: eventType, data: data }));
    },
    /**
     * @param {string} eventType 
     * @param {Function} func 
     */
    on: (eventType, func) => {
        if (handlers[eventType] == null) {
            handlers[eventType] = [];
        }

        handlers[eventType].push(func);
    },
    /**
     * @param {string} eventType 
     * @param {Function} func 
     */
    onTwitch: (eventType, func) => {
        if (twitchNotificationHandlers[eventType] == null) {
            twitchNotificationHandlers[eventType] = [];
        }

        twitchNotificationHandlers[eventType].push(func);
    },
    /**
     * @param {string} eventType 
     * @param {Object} data 
     */
    exec: (eventType, data) => {
        if (!handlers[eventType]) {
            console.log("(Exec) Handler não encontrado para o tipo:", eventType);
            return;
        }

        handlers[eventType].forEach(f => {
            f(data);
        });
    },
    getEmotes: () => emoteMap
}
