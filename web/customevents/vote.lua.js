import w from './../wrapper.js'

const vote = w.subscribe("vote.lua");
if (vote != null) {
    vote.on("user_vote_update", onVoteUpdate);
    vote.on("config", onVoteUpdate);
    vote.send("setup", { opcoes: ["Sim", "Não"] });
}

const stalkerDiv = document.getElementById('vote.lua');
const c = stalkerDiv.querySelector('.content-body');

c.innerHTML = `
    <button id="vote.lua-iniciar">Iniciar</button>
    <div id="vote.lua-time"></div>
    <div id="vote.lua-options">
    </div>
`;

const btn = document.getElementById('vote.lua-iniciar');
const options = document.getElementById('vote.lua-options');
const timer = document.getElementById('vote.lua-time');

btn.onclick = function () {
    vote.send('start', {opcoes: ["Sim", "Não"]});
}

function onVoteUpdate(data) {
    const v = data.votos;
    const names = Object.getOwnPropertyNames(v);

    timer.style.color = !data.ended ? "green" : "red";
    timer.innerHTML = `${data.tempo}s`;

    options.innerHTML = `${names.sort((a,b) => v[a].indexValue - v[b].indexValue).map(n=> `<div>${v[n].indexValue} - ${v[n].descr}: ${v[n].count}<div>`).join('')}`;
}


export default () => { };