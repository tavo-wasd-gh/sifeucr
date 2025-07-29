document.addEventListener("DOMContentLoaded", (event) => {

    // Auto-size textareas
    document.addEventListener('input', function (event) {
        if (event.target.tagName.toLowerCase() === 'textarea') {
            event.target.style.height = 'auto';
            event.target.style.height = event.target.scrollHeight + 'px';
        }
    });
});

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
    if (targetSelector) {
        targetEl = document.querySelector(targetSelector);
        if (!targetEl && swap !== 'none' && swap !== 'delete') {
            console.error('Target element not found for selector:', targetSelector);
            return;
        }
    }

    const loadingEls = form.querySelectorAll('[loading-state], [loading-state=""]');
    loadingEls.forEach(node => node.setAttribute('aria-busy', 'true'));
    if (form.hasAttribute('loading-state')) form.setAttribute('aria-busy', 'true');

    const formData = new FormData(form);

    for (const [key, value] of formData.entries()) {
        const input = form.querySelector(`[name="${key}"]`);
        if (!input || !value) continue;

        if (['date', 'datetime-local', 'datetime'].includes(input.type)) {
            const date = new Date(value);
            if (!isNaN(date)) {
                formData.set(key, Math.floor(date.getTime() / 1000));
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
            if (newCsrfToken) updateCSRFTokenMeta(newCsrfToken);

            const responseText = await res.text();
            if (res.status !== 200) {
                console.warn(`Fetch returned status ${res.status}, skipping swap.`);
                return;
            }

            const range = document.createRange();
            const fragment = range.createContextualFragment(responseText);

            // ---- Handle OOB elements BEFORE main insert
            const oobNodes = Array.from(fragment.querySelectorAll('[data-oob]'));
            oobNodes.forEach(oob => {
                const selector = oob.getAttribute('data-oob');
                const swapType = oob.getAttribute('swap') || 'innerHTML';
                const target = document.querySelector(selector);

                if (!target) {
                    console.warn(`OOB target not found: ${selector}`);
                    return;
                }

                const oobClone = oob.cloneNode(true);

                switch (swapType) {
                    case 'innerHTML':
                        target.innerHTML = '';
                        target.appendChild(oobClone);
                        break;
                    case 'outerHTML':
                        target.replaceWith(oobClone);
                        break;
                    case 'textContent':
                        target.textContent = oob.textContent;
                        break;
                    case 'beforebegin':
                    case 'afterbegin':
                    case 'beforeend':
                    case 'afterend':
                        target.insertAdjacentHTML(swapType, oob.outerHTML);
                        break;
                    case 'delete':
                        target.remove();
                        break;
                    case 'none':
                        break;
                    default:
                        console.warn('Unknown OOB swap type:', swapType);
                }

                // Remove OOB node from the fragment so it doesn’t get inserted again
                oob.remove();
            });

            // ---- Apply main content swap
            if (swap !== 'none') {
                switch (swap) {
                    case 'innerHTML':
                        targetEl.innerHTML = '';
                        targetEl.appendChild(fragment);
                        break;
                    case 'outerHTML':
                        targetEl.replaceWith(fragment);
                        break;
                    case 'textContent':
                        targetEl.textContent = responseText;
                        break;
                    case 'beforebegin':
                    case 'afterbegin':
                    case 'beforeend':
                    case 'afterend':
                        targetEl.insertAdjacentHTML(swap, responseText);
                        break;
                    case 'delete':
                        targetEl?.remove();
                        break;
                    default:
                        console.warn('Unknown swap type:', swap);
                }
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
            const tags = cell.getAttribute("tags") || cell.getAttribute("data-tags") || "";
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
                const tagsToCompare = hasUpperCase ? tags : tags.toLowerCase();
                const filterToCompare = hasUpperCase ? filterValue : filterValue.toLowerCase();

                if (
                    !cellToCompare.includes(filterToCompare) &&
                    !tagsToCompare.includes(filterToCompare)
                ) {
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
