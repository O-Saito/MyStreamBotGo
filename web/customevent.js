export function loadDyModules(modules) {
    if (!modules?.length) {
        document.getElementById('dyevents-panel').style.display = 'none';
        document.getElementById('dyevents-list').style.display = 'none';
        return;
    }

    const sidebar = document.getElementById('dyevents-sidebar');
    const dyEventsDiv = document.getElementById('dyevents-list');

    modules.forEach(m => {
        const sidebarItem = document.createElement('div');
        sidebarItem.className = 'sidebar-item';
        sidebarItem.textContent = m.name;
        sidebarItem.addEventListener('click', () => selectDyEvent(m.name));
        sidebar.appendChild(sidebarItem);

        const div = document.createElement('div');
        div.id = m.name;
        div.className = 'dyevent-content';
        div.innerHTML = `
            <h3>${m.name}</h3>
            <div><label>${m.paused ? 'Paused' : 'Running'}</label></div>
            <div class="content-body"></div>
        `;

        import(`./customevents/${m.name}.js?t=${Date.now()}`)
            .then(module => module.default?.())
            .catch(err => console.error('Module loading failed:', m.name, err));

        dyEventsDiv.appendChild(div);
    });

    if (modules.length > 0) {
        selectDyEvent(modules[0].name);
    }
}

export function selectDyEvent(name) {
    document.querySelectorAll('.sidebar-item').forEach(item => {
        item.classList.toggle('active', item.textContent === name);
    });
    document.querySelectorAll('.dyevent-content').forEach(content => {
        content.style.display = content.id === name ? 'block' : 'none';
    });
}
