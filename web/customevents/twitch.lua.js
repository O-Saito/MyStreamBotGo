import w from './../wrapper.js'

const filename = "twitch.lua";

const e = w.subscribe(filename);

const mydiv = document.getElementById(filename);
const mybody = mydiv.querySelector('.content-body');

mybody.style.overflowY = "auto"
mybody.innerHTML = `
    <pre id="${filename}-response" style="background:#1e1e1e;color:#d4d4d4;padding:8px;margin-bottom:12px;min-height:20px; overflow-y:auto; max-height: 200px;"></pre>
    <table id="${filename}-table" style="border-collapse:collapse;width:100%;">
        <thead>
            <tr style="border-bottom:1px solid #444;">
                <th style="padding:6px;text-align:left;">Function</th>
                <th style="padding:6px;">Arg 1</th>
                <th style="padding:6px;">Arg 2</th>
                <th style="padding:6px;">Arg 3</th>
                <th style="padding:6px;">Arg 4</th>
                <th style="padding:6px;">Arg 5</th>
                <th style="padding:6px;">Action</th>
            </tr>
        </thead>
        <tbody></tbody>
    </table>
`;

const responseDisplay = document.getElementById(`${filename}-response`);
const tbody = document.getElementById(`${filename}-table`).querySelector('tbody');

e.send("get_functions", {}, (d) => {
    console.log(filename, d);
    const functions = d;
    if (!Array.isArray(functions)) return;

    functions.forEach(fnName => {
        const tr = document.createElement('tr');
        tr.style.borderBottom = "1px solid #333";

        const nameTd = document.createElement('td');
        nameTd.textContent = fnName;
        nameTd.style.padding = "4px 8px";
        tr.appendChild(nameTd);

        const inputs = [];
        for (let i = 0; i < 5; i++) {
            const td = document.createElement('td');
            td.style.textAlign = "center";
            const input = document.createElement('input');
            input.type = "text";
            input.style.width = "100%";
            input.style.padding = "4px";
            td.appendChild(input);
            tr.appendChild(td);
            inputs.push(input);
        }

        const actionTd = document.createElement('td');
        actionTd.style.textAlign = "center";
        const btn = document.createElement('button');
        btn.textContent = "Send";
        btn.style.padding = "4px 12px";
        btn.onclick = function () {
            const args = inputs.map(inp => inp.value === "" ? null : inp.value);
            e.send(fnName, args, (response) => {
                responseDisplay.textContent = `${fnName} ${args}
${JSON.stringify(response, null, 2)}`
            });
        };
        actionTd.appendChild(btn);
        tr.appendChild(actionTd);

        tbody.appendChild(tr);
    });
});

let fs = null;
const openPanels = new Map();
export default (funcs) => {
    fs = funcs;
    console.log(fs);

    w.on("user-message", appendMessage);
    w.on("self-message", appendMessage);

    function appendMessage(data) {
        if (!openPanels.has(data.userId)) return;
        const message = document.createElement('div');
        message.innerHTML = fs.checkuserReference(fs.parseText(data));
        const panel = openPanels.get(data.userId);
        const chatBox = panel.querySelector(`.chat-history`);
        chatBox.appendChild(message);
        chatBox.scrollTop = chatDiv.scrollHeight;
    }

    document.addEventListener('click', (event) => {
        const chatUser = event.target.closest('.chat-user');
        if (!chatUser) return;
        const messageDiv = chatUser.closest('.message[data-source="twitch"]');
        if (!messageDiv) return;

        const username = chatUser.textContent.trim();
        const userId = messageDiv.dataset.userId;
        const color = chatUser.style.color || '#a970ff';
        const title = `${filename} - <span style="color:${color}">${username}</span>`;

        if (openPanels.has(userId)) {
            openPanels.get(userId).focus();
            return;
        }

        const loading = '<div style="padding:20px;text-align:center;">Loading...</div>';
        const panel = fs.showPanel(title, loading, { pos: { x: event.clientX, y: event.clientY }, onClose: () => openPanels.delete(userId) });

        e.send('get_user_data_by_id', [userId], (d) => {
            if (!d) {
                panel.querySelector('.panel-body').innerHTML = '<div style="padding:12px;">User not found.</div>';
                return;
            }

            const createdAt = new Date(d.CreatedAt).toLocaleString();
            const content = `
                <div class="user-info" style="display:flex;gap:12px;margin-bottom:16px;">
                    <img src="${d.ProfileImageURL}" style="width:70px;height:70px;border-radius:50%;border:2px solid ${color};" />
                    <div>
                        <div style="font-weight:bold;font-size:16px;">${d.DisplayName}</div>
                        <div style="color:#888;font-size:12px;">@${d.Login} · ID: ${d.ID}</div>
                        <div style="color:#aaa;font-size:12px;">Created: ${createdAt}</div>
                        ${d.BroadcasterType ? `<div style="color:#a970ff;font-size:12px;text-transform:uppercase;">${d.BroadcasterType}</div>` : ''}
                    </div>
                </div>
                ${d.Description ? `<div style="margin-bottom:16px;padding:8px;background:#2a2a2a;border-radius:4px;font-size:13px;">${d.Description}</div>` : ''}
                <div style="margin-bottom:16px;">
                    <div style="font-weight:bold;font-size:14px;margin-bottom:8px;">Chat History</div>
                    <div class="chat-history" style="min-height:100px;padding:8px;background:#1a1a1a;border-radius:4px;font-size:12px;overflow-y: auto;max-height: 100px;">
                    </div>
                </div>
                <button class="open-viewercard" style="padding:6px 14px;background:#6441a5;color:#fff;border:none;border-radius:4px;cursor:pointer;">Open Viewer Card</button>
            `;

            panel.querySelector('.panel-body').innerHTML = content;
            panel.querySelector(`.open-viewercard`).onclick = () => {
                window.open(`https://www.twitch.tv/popout/scavote/viewercard/${d.Login}`, '_blank');
            };
            const chat = document.querySelector('#chat .content');
            const messages = chat.querySelectorAll(`.message[data-user-id="${userId}"]`);
            const panelChat = panel.querySelector(`.chat-history`);
            for (let i = 0; i < messages.length; i++) {
                const msgElement = messages[i];
                const message = document.createElement('div');
                message.innerHTML = msgElement.querySelector('.chat-message').innerHTML;
                panelChat.appendChild(message);
                panelChat.scrollTop = panelChat.scrollHeight;
            }
        });
    });
};