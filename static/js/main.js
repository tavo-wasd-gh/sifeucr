function fetchAndSwap(event) {
    event.preventDefault();

    const el = event.currentTarget;
    const form = el.tagName === 'FORM' ? el : el.closest('form');
    if (!form) return console.error("No form found for element");

    let action = form.getAttribute('action') || window.location.href;
    const method = (form.getAttribute('method') || 'GET').toUpperCase();
    const targetSelector = el.getAttribute('target');
    const swap = el.getAttribute('swap') || 'innerHTML';

    let targetEl = null;
    if (targetSelector?.startsWith('#')) {
        targetEl = document.querySelector(targetSelector);
    } else if (targetSelector) {
        targetEl = document.querySelector(targetSelector);
        if (!targetEl) {
            console.warn(`No element found for selector: '${targetSelector}'`);
        }
    }

    if (!targetEl && swap !== 'none' && swap !== 'delete') {
        console.error('Target element not found for selector:', targetSelector);
        return;
    }

    const loadingEls = form.querySelectorAll('[loading-state], [loading-state=""]');
    loadingEls.forEach(node => node.setAttribute('aria-busy', 'true'));
    if (form.hasAttribute('loading-state')) form.setAttribute('aria-busy', 'true');

    const formData = new FormData(form);

    for (const [key, value] of formData.entries()) {
        const input = form.querySelector(`[name="${key}"]`);
        if (!input || !value) continue;

        if (input.type === 'date' || input.type === 'datetime-local' || input.type === 'datetime') {
            const date = new Date(value);
            if (!isNaN(date)) {
                const unixTime = Math.floor(date.getTime() / 1000);
                formData.set(key, unixTime);
            }
        }
    }

    const fetchOptions = {
        method,
        headers: {
            'X-CSRF-Token': getCSRFTokenFromMeta()
        }
    };

    if (method !== 'GET') {
        fetchOptions.body = formData;
    } else {
        const params = new URLSearchParams(formData).toString();
        action += (action.includes('?') ? '&' : '?') + params;
    }

    fetch(action, fetchOptions)
        .then(async res => {
            const newCsrfToken = res.headers.get('X-CSRF-Token');
            if (newCsrfToken) {
                updateCSRFTokenMeta(newCsrfToken);
            }

            const responseText = await res.text();
            if (res.status !== 200) {
                console.warn(`Fetch returned status ${res.status}, skipping swap.`);
                return;
            }

            switch (swap) {
                case 'innerHTML':
                    targetEl.innerHTML = responseText;
                    break;
                case 'outerHTML':
                    targetEl.outerHTML = responseText;
                    break;
                case 'textContent':
                    targetEl.textContent = responseText;
                    break;
                case 'beforebegin':
                    targetEl.insertAdjacentHTML('beforebegin', responseText);
                    break;
                case 'afterbegin':
                    targetEl.insertAdjacentHTML('afterbegin', responseText);
                    break;
                case 'beforeend':
                    targetEl.insertAdjacentHTML('beforeend', responseText);
                    break;
                case 'afterend':
                    targetEl.insertAdjacentHTML('afterend', responseText);
                    break;
                case 'delete':
                    targetEl?.remove();
                    break;
                case 'none':
                    break;
                default:
                    console.warn('Unknown swap type:', swap);
            }
        })
        .catch(err => {
            console.error('Request failed:', err);
        })
        .finally(() => {
            loadingEls.forEach(node => node.setAttribute('aria-busy', 'false'));
            if (form.hasAttribute('loading-state')) form.setAttribute('aria-busy', 'false');
        });
}

function getCSRFTokenFromMeta() {
    const meta = document.querySelector('meta[name="csrf_token"]');
    return meta?.getAttribute('content') || '';
}

function updateCSRFTokenMeta(token) {
    let meta = document.querySelector('meta[name="csrf_token"]');
    if (!meta) {
        meta = document.createElement('meta');
        meta.setAttribute('name', 'csrf_token');
        document.head.appendChild(meta);
    }
    meta.setAttribute('content', token);
}

function filterTable(inputChanged) {
    const table = inputChanged.closest("table");
    const rows = table.tBodies[0].rows;
    const inputs = table.querySelectorAll("thead input");

    for (let i = 0; i < rows.length; i++) {
        let row = rows[i];
        let visible = true;

        for (let j = 0; j < inputs.length; j++) {
            const input = inputs[j];
            const filterValue = input.value.trim();
            if (!filterValue) continue;

            const colIndex = input.closest('th').cellIndex;
            const cell = row.cells[colIndex];
            if (!cell) continue;

            const cellText = (cell.textContent || cell.innerText).trim();
            const inputType = input.type;
            const filterAttr = input.getAttribute("filter");

            if ((inputType === "date" || inputType === "datetime-local" || inputType === "datetime") && filterValue) {
                const cellDate = new Date(cellText);
                const filterDate = new Date(filterValue);
                if (isNaN(cellDate) || isNaN(filterDate)) {
                    visible = false;
                    break;
                }

                if (filterAttr === "before" && !(cellDate <= filterDate)) {
                    visible = false;
                    break;
                }
                if (filterAttr === "after" && !(cellDate >= filterDate)) {
                    visible = false;
                    break;
                }
            } else if (inputType === "number" && filterValue !== "") {
                const cleanFilterValue = filterValue.replace(/[^\d.]/g, '').replace(/^0+/, '');
                const cleanedCellText = cellText.replace(/[^\d.]/g, '').replace(/^0+/, '');

                if (!cleanedCellText.includes(cleanFilterValue)) {
                    visible = false;
                    break;
                }
            } else if (filterAttr === "number" && filterValue !== "") {
                const cleanFilterValue = filterValue.replace(/[^\d.]/g, '').replace(/^0+/, '');
                const cleanedCellText = cellText.replace(/[^\d.]/g, '').replace(/^0+/, '');

                if (!cleanedCellText.includes(cleanFilterValue)) {
                    visible = false;
                    break;
                }
            } else if (filterValue) {
                const hasUpperCase = /[A-Z]/.test(filterValue);
                const cellToCompare = hasUpperCase ? cellText : cellText.toLowerCase();
                const filterToCompare = hasUpperCase ? filterValue : filterValue.toLowerCase();
                if (!cellToCompare.includes(filterToCompare)) {
                    visible = false;
                    break;
                }
            }
        }

        row.style.display = visible ? "" : "none";
    }
}

function datetimeLocalToUnix(datetimeValue) {
    const date = new Date(datetimeValue);
    return Math.floor(date.getTime() / 1000);
}

function toggleModal(id) {
    const modal = document.getElementById(id);
    if (!modal || modal.tagName !== 'DIALOG') {
        console.warn(`No <dialog> found with ID: ${id}`);
        return;
    }

    if (modal.hasAttribute('open')) {
        modal.removeAttribute('open');
    } else {
        modal.setAttribute('open', '');
    }
}
