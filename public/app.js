const API_BASE = 'http://localhost:8080';

// Elements - Views
const loginView = document.getElementById('login-view');
const dashboardView = document.getElementById('dashboard-view');
const loginForm = document.getElementById('login-form');
const loginError = document.getElementById('login-error');
const logoutBtn = document.getElementById('logout-btn');
const userRoleBadge = document.getElementById('user-role-badge');
const tabBtns = document.querySelectorAll('.tab-btn');
const listViews = document.querySelectorAll('.list-view');
const addEmployeeBtn = document.getElementById('add-employee-btn');

// Elements - Add Modal
const addModal = document.getElementById('add-modal');
const closeAddModal = document.getElementById('close-add-modal');
const addForm = document.getElementById('add-form');
const addRoleSelect = document.getElementById('add-role');
const addError = document.getElementById('add-error');
const managerSelectGroup = document.getElementById('manager-select-group');
const boardSelectGroup = document.getElementById('board-select-group');

// Elements - Edit Modal
const editModal = document.getElementById('edit-modal');
const closeEditModal = document.getElementById('close-edit-modal');
const editForm = document.getElementById('edit-form');
const editError = document.getElementById('edit-error');
const editManagerGroup = document.getElementById('edit-manager-select-group');
const editBoardGroup = document.getElementById('edit-board-select-group');
const editSalaryGroup = document.getElementById('edit-salary-group');

// Elements - Relations Modal
const relationsModal = document.getElementById('relations-modal');
const closeRelationsModal = document.getElementById('close-relations-modal');

// State
let token = localStorage.getItem('token');
let roleId = parseInt(localStorage.getItem('roleId'));
let userId = parseInt(localStorage.getItem('userId'));

let globalManagers = [];
let globalBoardMembers = [];
let myWorkerId = null;
let myManagerId = null;
let myBoardMemId = null;

// Init
if (token) {
    showDashboard();
}

// ========================
// CORE AUTHENTICATION
// ========================
loginForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const email = document.getElementById('email').value;
    const password = document.getElementById('password').value;

    try {
        const res = await fetch(`${API_BASE}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await res.json();
        if (res.ok) {
            token = data.token;
            roleId = data.role_id;
            userId = data.user_id;
            localStorage.setItem('token', token);
            localStorage.setItem('roleId', roleId);
            localStorage.setItem('userId', userId);
            loginError.classList.add('hidden');
            showDashboard();
        } else {
            loginError.textContent = data.error || 'Login failed';
            loginError.classList.remove('hidden');
        }
    } catch (err) {
        loginError.textContent = 'Connection error. Is the server running?';
        loginError.classList.remove('hidden');
    }
});

logoutBtn.addEventListener('click', () => {
    token = null;
    roleId = null;
    userId = null;
    localStorage.removeItem('token');
    localStorage.removeItem('roleId');
    localStorage.removeItem('userId');
    loginView.classList.remove('hidden');
    dashboardView.classList.add('hidden');
    loginForm.reset();
});

// ========================
// DASHBOARD & DATA LOADING
// ========================
async function showDashboard() {
    loginView.classList.add('hidden');
    dashboardView.classList.remove('hidden');
    
    const roles = { 1: 'HR', 2: 'Admin', 3: 'Worker', 4: 'Manager', 5: 'Board Member' };
    userRoleBadge.textContent = roles[roleId] || 'User';

    // Show 'Add' button only for Admin/HR
    if (roleId === 1 || roleId === 2) {
        addEmployeeBtn.classList.remove('hidden');
    } else {
        addEmployeeBtn.classList.add('hidden');
    }

    await loadData();
}

async function fetchAPI(endpoint, method = 'GET', body = null) {
    const options = {
        method,
        headers: { 
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json'
        }
    };
    if (body) options.body = JSON.stringify(body);
    
    const res = await fetch(`${API_BASE}${endpoint}`, options);
    if (res.status === 401) {
        logoutBtn.click();
        throw new Error('Unauthorized');
    }
    
    if (res.status === 204 || method === 'DELETE') {
        if(!res.ok) throw new Error('Delete failed');
        return true;
    }
    
    const data = await res.json();
    if (!res.ok) throw new Error(data.error || 'Request failed');
    return data;
}

async function loadData() {
    try {
        const [workers, managers, boardMembers] = await Promise.all([
            fetchAPI('/workers').catch(() => []),
            fetchAPI('/managers').catch(() => []),
            fetchAPI('/board-members').catch(() => [])
        ]);

        globalManagers = managers.length ? managers : [];
        globalBoardMembers = boardMembers.length ? boardMembers : [];
        
        // Resolve Identity for Authz
        myWorkerId = null;
        myManagerId = null;
        myBoardMemId = null;
        
        if (roleId === 3) {
            const me = workers.find(w => w.user_id === userId);
            if (me) myWorkerId = me.worker_id;
        } else if (roleId === 4) {
            const me = managers.find(m => m.user_id === userId);
            if (me) myManagerId = me.manager_id;
        } else if (roleId === 5) {
            const me = boardMembers.find(b => b.user_id === userId);
            if (me) myBoardMemId = me.board_mem_id;
        }

        renderList('workers-content', workers, 'workers', 'worker_id');
        renderList('managers-content', managers, 'managers', 'manager_id');
        renderList('board-content', boardMembers, 'board-members', 'board_mem_id');
    } catch (e) {
        console.error("Failed to load data", e);
    }
}

function renderList(containerId, items, type, idField) {
    const container = document.getElementById(containerId);
    if (!items || items.error || !Array.isArray(items)) {
        container.innerHTML = `<div style="padding: 24px; color: var(--text-muted)">No access or error loading data.</div>`;
        return;
    }
    if (items.length === 0) {
        container.innerHTML = `<div style="padding: 24px; color: var(--text-muted)">No records found.</div>`;
        return;
    }

    container.innerHTML = items.map(item => {
        let canEdit = false;
        let canDelete = false;

        // AUTHZ RULES LOGIC
        if (roleId === 1 || roleId === 2) {
            canEdit = true;
            canDelete = true;
        } else if (roleId === 3 && type === 'workers' && item.user_id === userId) {
            canEdit = true;
        } else if (roleId === 4) {
            if (type === 'managers' && item.user_id === userId) canEdit = true;
            if (type === 'workers' && item.manager_id === myManagerId) canEdit = true;
        } else if (roleId === 5) {
            if (type === 'board-members' && item.user_id === userId) canEdit = true;
            if (type === 'managers' && item.board_mem_id === myBoardMemId) canEdit = true;
        }

        let actionBtns = '';
        if (canEdit) {
            const itemJSON = encodeURIComponent(JSON.stringify(item));
            actionBtns += `<button class="btn-edit" onclick="openEditModal('${type}', '${itemJSON}')">Edit</button> `;
        }
        if (canDelete) {
            actionBtns += `<button class="btn-danger" onclick="deleteUser('${type}', ${item[idField]})">Delete</button>`;
        }
        
        if (!canEdit && !canDelete) {
            actionBtns = `<span style="color:var(--text-muted); font-size:12px;">View Only</span>`;
        }

        let gridClass = '';
        let extraCols = '';
        if (type === 'workers') {
            gridClass = 'workers-grid';
            extraCols = `
                <span>${item.manager_name || 'None'}</span>
                <span>${item.board_mem_name || 'None'}</span>
            `;
        } else if (type === 'managers') {
            gridClass = 'managers-grid';
            const workersJSON = encodeURIComponent(JSON.stringify(item.workers || []));
            const workersCount = item.workers ? item.workers.length : 0;
            extraCols = `
                <span>${item.board_mem_name || 'None'}</span>
                <div>
                    <button class="btn-relation-view" onclick="openRelationsModal('Workers under ${item.name}', '${workersJSON}', 'worker')">
                        View (${workersCount})
                    </button>
                </div>
            `;
        } else if (type === 'board-members') {
            gridClass = 'board-grid';
            const managersJSON = encodeURIComponent(JSON.stringify(item.managers || []));
            const managersCount = item.managers ? item.managers.length : 0;
            const workersJSON = encodeURIComponent(JSON.stringify(item.workers || []));
            const workersCount = item.workers ? item.workers.length : 0;
            extraCols = `
                <div>
                    <button class="btn-relation-view" onclick="openRelationsModal('Managers under ${item.name}', '${managersJSON}', 'manager')">
                        View (${managersCount})
                    </button>
                </div>
                <div>
                    <button class="btn-relation-view" onclick="openRelationsModal('Workers under ${item.name}', '${workersJSON}', 'worker')">
                        View (${workersCount})
                    </button>
                </div>
            `;
        }

        const salaryDisplay = (item.salary !== null && item.salary !== undefined)
            ? `$${Number(item.salary).toLocaleString()}`
            : `<span class="salary-confidential">Confidential</span>`;

        return `
            <div class="list-item ${gridClass}">
                <span style="font-weight: 500">${item.name}</span>
                <span style="color: var(--text-muted)">${item.email}</span>
                <span>${salaryDisplay}</span>
                ${extraCols}
                <div class="actions-cell">
                    ${actionBtns}
                </div>
            </div>
        `;
    }).join('');
}

// Tab Switching
tabBtns.forEach(btn => {
    btn.addEventListener('click', () => {
        tabBtns.forEach(b => b.classList.remove('active'));
        listViews.forEach(v => v.classList.add('hidden'));
        btn.classList.add('active');
        document.getElementById(btn.dataset.target).classList.remove('hidden');
    });
});

// ========================
// CREATE USER (POST)
// ========================
addEmployeeBtn.addEventListener('click', () => {
    addError.classList.add('hidden');
    addForm.reset();
    
    // Populate role options dynamically based on the current user's role
    if (roleId === 2) { // Admin
        addRoleSelect.innerHTML = `
            <option value="1">HR</option>
            <option value="5">Board Member</option>
            <option value="4">Manager</option>
            <option value="3">Worker</option>
        `;
    } else { // HR
        addRoleSelect.innerHTML = `
            <option value="5">Board Member</option>
            <option value="4">Manager</option>
            <option value="3">Worker</option>
        `;
    }

    handleRoleChange(); 
    addModal.classList.remove('hidden');
});

closeAddModal.addEventListener('click', () => addModal.classList.add('hidden'));

addRoleSelect.addEventListener('change', handleRoleChange);

function handleRoleChange() {
    const role = addRoleSelect.value;
    managerSelectGroup.classList.add('hidden');
    boardSelectGroup.classList.add('hidden');
    
    const managerSelect = document.getElementById('add-manager-id');
    const boardSelect = document.getElementById('add-board-id');
    
    if (role === '3') { 
        managerSelectGroup.classList.remove('hidden');
        managerSelect.innerHTML = globalManagers.map(m => `<option value="${m.manager_id}">${m.name}</option>`).join('');
    } else if (role === '4') { 
        boardSelectGroup.classList.remove('hidden');
        boardSelect.innerHTML = globalBoardMembers.map(b => `<option value="${b.board_mem_id}">${b.name}</option>`).join('');
    }
}

addForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const payload = {
        name: document.getElementById('add-name').value,
        email: document.getElementById('add-email').value,
        password: document.getElementById('add-password').value,
        role_id: parseInt(addRoleSelect.value),
        salary: parseFloat(document.getElementById('add-salary').value)
    };

    if (payload.role_id === 3) payload.manager_id = parseInt(document.getElementById('add-manager-id').value);
    if (payload.role_id === 4) payload.board_mem_id = parseInt(document.getElementById('add-board-id').value);

    try {
        await fetchAPI('/admin/users', 'POST', payload);
        addModal.classList.add('hidden');
        loadData(); 
    } catch (err) {
        addError.textContent = err.message;
        addError.classList.remove('hidden');
    }
});

// ========================
// DELETE USER (DELETE)
// ========================
window.deleteUser = async (type, id) => {
    if (!confirm(`Are you sure you want to delete this ${type.replace('-', ' ')}?`)) return;
    try {
        await fetchAPI(`/${type}/${id}`, 'DELETE');
        loadData(); 
    } catch (err) {
        alert("Failed to delete: " + err.message);
    }
};

// ========================
// EDIT USER (PUT)
// ========================
closeEditModal.addEventListener('click', () => editModal.classList.add('hidden'));

window.openEditModal = (type, itemJSON) => {
    const item = JSON.parse(decodeURIComponent(itemJSON));
    editError.classList.add('hidden');
    
    let id = 0;
    if (type === 'workers') id = item.worker_id;
    else if (type === 'managers') id = item.manager_id;
    else if (type === 'board-members') id = item.board_mem_id;
    
    document.getElementById('edit-id').value = id;
    document.getElementById('edit-type').value = type;
    document.getElementById('edit-name').value = item.name;
    document.getElementById('edit-email').value = item.email;
    document.getElementById('edit-password').value = ''; // Reset password field
    document.getElementById('edit-salary').value = (item.salary !== null && item.salary !== undefined) ? item.salary : '';
    
    // Authz Form Logic: Only Admin/HR can see/edit Salary or Assignments
    if (roleId === 1 || roleId === 2) {
        editSalaryGroup.classList.remove('hidden');
        
        editManagerGroup.classList.add('hidden');
        editBoardGroup.classList.add('hidden');
        
        if (type === 'workers') {
            editManagerGroup.classList.remove('hidden');
            const sel = document.getElementById('edit-manager-id');
            sel.innerHTML = globalManagers.map(m => `<option value="${m.manager_id}" ${m.manager_id === item.manager_id ? 'selected' : ''}>${m.name}</option>`).join('');
        } else if (type === 'managers') {
            editBoardGroup.classList.remove('hidden');
            const sel = document.getElementById('edit-board-id');
            sel.innerHTML = globalBoardMembers.map(b => `<option value="${b.board_mem_id}" ${b.board_mem_id === item.board_mem_id ? 'selected' : ''}>${b.name}</option>`).join('');
        }
    } else {
        editSalaryGroup.classList.add('hidden');
        editManagerGroup.classList.add('hidden');
        editBoardGroup.classList.add('hidden');
    }
    
    editModal.classList.remove('hidden');
};

editForm.addEventListener('submit', async (e) => {
    e.preventDefault();
    const type = document.getElementById('edit-type').value;
    const id = document.getElementById('edit-id').value;
    const pwd = document.getElementById('edit-password').value;
    
    const payload = {
        name: document.getElementById('edit-name').value,
        email: document.getElementById('edit-email').value
    };

    if (pwd) {
        payload.password = pwd;
    }
    
    // Only append admin-only fields if they are an admin
    if (roleId === 1 || roleId === 2) {
        payload.salary = parseFloat(document.getElementById('edit-salary').value);
        if (type === 'workers') payload.manager_id = parseInt(document.getElementById('edit-manager-id').value);
        if (type === 'managers') payload.board_mem_id = parseInt(document.getElementById('edit-board-id').value);
    }
    
    try {
        await fetchAPI(`/${type}/${id}`, 'PUT', payload);
        editModal.classList.add('hidden');
        loadData(); 
    } catch (err) {
        editError.textContent = err.message;
        editError.classList.remove('hidden');
    }
});

// ========================
// RELATIONS MODAL & CLOSE HANDLERS
// ========================
closeRelationsModal.addEventListener('click', () => relationsModal.classList.add('hidden'));

window.openRelationsModal = (title, listJSON, type) => {
    const list = JSON.parse(decodeURIComponent(listJSON));
    const titleEl = document.getElementById('relations-title');
    const contentEl = document.getElementById('relations-content');
    
    titleEl.textContent = title;
    
    if (!list || list.length === 0) {
        contentEl.innerHTML = `<div style="padding: 24px; color: var(--text-muted); text-align: center; font-size: 14px;">No related personnel found.</div>`;
    } else {
        const badgeClass = type === 'manager' ? 'badge-manager' : 'badge-worker';
        contentEl.innerHTML = `
            <div style="display: flex; flex-direction: column; gap: 10px; padding: 5px;">
                ${list.map(name => `
                    <div style="display: flex; align-items: center; justify-content: space-between; padding: 10px 14px; background: rgba(255,255,255,0.03); border: 1px solid var(--glass-border); border-radius: 8px;">
                        <span style="font-weight: 500; color: var(--text-primary);">${name}</span>
                        <span class="badge ${badgeClass}">${type.charAt(0).toUpperCase() + type.slice(1)}</span>
                    </div>
                `).join('')}
            </div>
        `;
    }
    
    relationsModal.classList.remove('hidden');
};

// Click outside modals to close
window.addEventListener('click', (e) => {
    if (e.target === relationsModal) relationsModal.classList.add('hidden');
    if (e.target === addModal) addModal.classList.add('hidden');
    if (e.target === editModal) editModal.classList.add('hidden');
});
