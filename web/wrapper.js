//@ts-check

/**
 * @typedef SocketMessage
 * @property {string} type
 * @property {string} filter
 * @property {string|undefined} responseClientID
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

let ws = new WebSocket(`ws://${location.host}/ws`);
let reconnectInterval = 1000; // Initial delay in milliseconds
let maxReconnectInterval = 30000; // Maximum delay
let reconnectAttempts = 0;

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
            console.error('twitch-eventsub-notification não tratado!', data);
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

function webSocketConnect() {
    ws = new WebSocket(`ws://${location.host}/ws`);
    /** @param {MessageEvent<any>} e  */
    ws.onmessage = function (e) {
        //console.log("WS Message:", e.data);
        /** @type {SocketMessage} */
        const data = JSON.parse(e.data);
        if (data.filter && data.filter != "" && eventsHandlers[data.filter] && eventsHandlers[data.filter][data.type]) {
            eventsHandlers[data.filter][data.type].forEach(f => {
                f(data.data);
            });
            return;
        }

        if (data.responseClientID && waitingResponseFunctions[data.responseClientID]) {
            const f = waitingResponseFunctions[data.responseClientID];
            delete waitingResponseFunctions[data.responseClientID];
            f(data.data);
            return;
        }

        if (!handlers[data.type]) {
            console.warn("Handler not found:", data.type);
            return;
        }
        handlers[data.type].forEach(f => { f(data.data); });
    };

    ws.onclose = () => {
        console.warn('WebSocket disconnected. Attempting to reconnect...');
        reconnectAttempts++;
        const delay = Math.min(reconnectInterval * Math.pow(2, reconnectAttempts - 1), maxReconnectInterval);
        setTimeout(webSocketConnect, delay);
    };
}

/** @type {Object.<string, Function>} */
const waitingResponseFunctions = {};

export default {
    connect: () => {
        if (connectCalled) {
            console.error("Connected was called before!");
            return;
        }
        connectCalled = true;
        webSocketConnect();
        ws.onopen = function () {
            console.log("Connecting to WebSocket.");
            ws.send("init");
        }
    },
    ignoreBroadcast: () => {
        ws.send(JSON.stringify({ type: "upgrade-conn", data: { "conn": 'ignore-broadcast' } }));
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
     * @param {Function|undefined} func 
     */
    send: (eventType, data, func) => {
        if (func) {
            const id = crypto.randomUUID();
            waitingResponseFunctions[id] = func;
            ws.send(JSON.stringify({ type: eventType, data: data, responseClientID: id }));
            return
        }
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
            console.warn("(Exec) Handler não encontrado para o tipo:", eventType);
            return;
        }

        handlers[eventType].forEach(f => {
            f(data);
        });
    },
    /**
     * @returns {Object.<string, string>}
     */
    getEmotes: () => emoteMap,
    /**
     * @param {string} twitchId 
     * @param {string} login 
     */
    loadEmote: async (twitchId, login) => {
        const emotes = await loadAllEmotes(twitchId, login);
        emoteMap = { ...emoteMap, ...emotes };
        return emotes;
    },
}
