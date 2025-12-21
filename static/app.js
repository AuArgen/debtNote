document.addEventListener('DOMContentLoaded', () => {
    const app = document.getElementById('app');
    const payDebtModal = document.getElementById('pay-debt-modal');
    const clientDetailsModal = document.getElementById('client-details-modal');
    let currentDebtToPay = null;

    // --- Router ---
    const routes = {
        '/': 'home-page',
        '/clients': 'clients-page',
        '/history': 'history-page'
    };

    const render = (path) => {
        const templateId = routes[path] || 'home-page';
        const template = document.getElementById(templateId);
        if (template) {
            app.innerHTML = template.innerHTML;
            if (path === '/') initHomePage();
            if (path === '/clients') initClientsPage();
            if (path === '/history') initHistoryPage();
        } else {
            app.innerHTML = '<h2>404 - Page Not Found</h2>';
        }
    };

    window.addEventListener('popstate', () => render(window.location.pathname));
    document.body.addEventListener('click', e => {
        // Handle navigation links
        if (e.target.matches('a[href^="/"]')) {
            e.preventDefault();
            history.pushState(null, null, e.target.href);
            render(window.location.pathname);
        }
        
        // Handle Image Click (Zoom)
        if (e.target.classList.contains('zoomable-image')) {
            const src = e.target.src;
            if (src && !src.includes('placeholder.com')) {
                openImageInNewTab(src);
            }
        }
    });

    // --- Helper: Open Image in New Window ---
    function openImageInNewTab(base64Data) {
        const newWindow = window.open();
        if (newWindow) {
            newWindow.document.write(`
                <html>
                    <head>
                        <title>Сүрөт</title>
                        <style>
                            body { margin: 0; display: flex; justify-content: center; align-items: center; background-color: #222; height: 100vh; }
                            img { max-width: 100%; max-height: 100vh; object-fit: contain; box-shadow: 0 0 20px rgba(0,0,0,0.5); }
                        </style>
                    </head>
                    <body>
                        <img src="${base64Data}" alt="Full size photo">
                    </body>
                </html>
            `);
            newWindow.document.close();
        }
    }

    // --- Helper: Get Rating Badge (Translation) ---
    function getRatingBadge(rating) {
        switch(rating) {
            case 'good':
                return '<span class="text-green-600 font-bold">Жакшы</span>';
            case 'bad':
                return '<span class="text-orange-500 font-bold">Начар</span>';
            case 'untrusted':
                return '<span class="text-red-600 font-bold">Ишенич жок</span>';
            default:
                return '<span class="text-gray-400">-</span>';
        }
    }

    // --- Common Functions ---
    function createPagination(containerId, currentPage, totalItems, itemsPerPage, onPageClick) {
        const container = document.getElementById(containerId);
        if (!container) return;
        container.innerHTML = '';
        
        const totalPages = Math.ceil(totalItems / itemsPerPage);
        if (totalPages <= 1) return;

        const nav = document.createElement('nav');
        nav.className = 'flex justify-center space-x-2 mt-4';

        // Previous Button
        if (currentPage > 1) {
            const prevBtn = document.createElement('button');
            prevBtn.textContent = '<';
            prevBtn.className = 'px-3 py-1 rounded bg-gray-200 hover:bg-gray-300';
            prevBtn.addEventListener('click', () => onPageClick(currentPage - 1));
            nav.appendChild(prevBtn);
        }

        // Page Numbers (Show limited range)
        let startPage = Math.max(1, currentPage - 2);
        let endPage = Math.min(totalPages, currentPage + 2);

        if (startPage > 1) {
            const firstBtn = document.createElement('button');
            firstBtn.textContent = '1';
            firstBtn.className = 'px-3 py-1 rounded bg-gray-200 hover:bg-gray-300';
            firstBtn.addEventListener('click', () => onPageClick(1));
            nav.appendChild(firstBtn);
            if (startPage > 2) {
                const dots = document.createElement('span');
                dots.textContent = '...';
                dots.className = 'px-2 py-1';
                nav.appendChild(dots);
            }
        }

        for (let i = startPage; i <= endPage; i++) {
            const button = document.createElement('button');
            button.textContent = i;
            button.className = `px-3 py-1 rounded ${i === currentPage ? 'bg-blue-500 text-white' : 'bg-gray-200 hover:bg-gray-300'}`;
            button.addEventListener('click', () => onPageClick(i));
            nav.appendChild(button);
        }

        if (endPage < totalPages) {
            if (endPage < totalPages - 1) {
                const dots = document.createElement('span');
                dots.textContent = '...';
                dots.className = 'px-2 py-1';
                nav.appendChild(dots);
            }
            const lastBtn = document.createElement('button');
            lastBtn.textContent = totalPages;
            lastBtn.className = 'px-3 py-1 rounded bg-gray-200 hover:bg-gray-300';
            lastBtn.addEventListener('click', () => onPageClick(totalPages));
            nav.appendChild(lastBtn);
        }

        // Next Button
        if (currentPage < totalPages) {
            const nextBtn = document.createElement('button');
            nextBtn.textContent = '>';
            nextBtn.className = 'px-3 py-1 rounded bg-gray-200 hover:bg-gray-300';
            nextBtn.addEventListener('click', () => onPageClick(currentPage + 1));
            nav.appendChild(nextBtn);
        }

        container.appendChild(nav);
    }

    // --- Home Page Logic ---
    function initHomePage() {
        const addDebtForm = document.getElementById('add-debt-form');
        const fullnameInput = document.getElementById('fullname');
        const phoneInput = document.getElementById('phone');
        const addressInput = document.getElementById('address');
        const clientSuggestions = document.getElementById('client-suggestions');
        const photoElement = document.getElementById('photo');
        const photoDataElement = document.getElementById('photo-data');
        const webcamElement = document.getElementById('webcam');
        const canvasElement = document.getElementById('canvas');
        const startCameraButton = document.getElementById('start-camera');
        const capturePhotoButton = document.getElementById('capture-photo');
        const recapturePhotoButton = document.getElementById('recapture-photo');
        
        // Search and Filter Elements
        const searchInput = document.getElementById('search-active');
        const dateFilter = document.getElementById('filter-date-active');
        
        let searchTimeout = null;
        let stream = null;
        let isExistingClient = false;

        // Smart select logic
        fullnameInput.addEventListener('input', () => {
            clearTimeout(searchTimeout);
            const query = fullnameInput.value;
            if (query.length < 2) {
                clientSuggestions.classList.add('hidden');
                return;
            }
            searchTimeout = setTimeout(async () => {
                const response = await fetch(`/api/clients/search?q=${encodeURIComponent(query)}`);
                const clients = await response.json();
                clientSuggestions.innerHTML = '';
                if (clients && clients.length > 0) {
                    clients.forEach(client => {
                        const item = document.createElement('div');
                        item.className = 'p-2 hover:bg-gray-100 cursor-pointer flex items-center space-x-2 border-b last:border-b-0';
                        
                        // Active Debt Indicator
                        const debtIndicator = client.has_active_debt 
                            ? '<span class="w-3 h-3 bg-red-500 rounded-full mr-2" title="Активдүү карызы бар"></span>' 
                            : '<span class="w-3 h-3 bg-green-500 rounded-full mr-2" title="Карызы жок"></span>';

                        // Reputation Badge
                        let reputationBadge = '';
                        switch(client.reputation) {
                            case 'untrusted':
                                reputationBadge = '<span class="px-2 py-0.5 text-xs font-bold text-white bg-black rounded">Ишенич жок</span>';
                                break;
                            case 'bad':
                                reputationBadge = '<span class="px-2 py-0.5 text-xs font-bold text-white bg-orange-500 rounded">Начар</span>';
                                break;
                            case 'good':
                                reputationBadge = '<span class="px-2 py-0.5 text-xs font-bold text-white bg-green-600 rounded">Жакшы</span>';
                                break;
                            default:
                                reputationBadge = '<span class="px-2 py-0.5 text-xs font-bold text-gray-600 bg-gray-200 rounded">Жаңы</span>';
                        }

                        item.innerHTML = `
                            <img src="${client.photo_data || 'https://via.placeholder.com/40'}" class="w-10 h-10 rounded-full object-cover zoomable-image">
                            <div class="flex-grow ml-2">
                                <div class="flex items-center justify-between">
                                    <div class="font-bold">${client.fullname}</div>
                                    <div class="flex items-center">${debtIndicator}${reputationBadge}</div>
                                </div>
                                <div class="text-sm text-gray-500">${client.phone}</div>
                            </div>
                        `;
                        item.addEventListener('click', (e) => {
                            if (!e.target.classList.contains('zoomable-image')) {
                                selectClient(client);
                            }
                        });
                        clientSuggestions.appendChild(item);
                    });
                    clientSuggestions.classList.remove('hidden');
                } else {
                    clientSuggestions.classList.add('hidden');
                }
            }, 300);
        });

        // Close suggestions when clicking outside or tabbing away
        fullnameInput.addEventListener('blur', () => {
            setTimeout(() => {
                clientSuggestions.classList.add('hidden');
            }, 200);
        });

        function selectClient(client) {
            fullnameInput.value = client.fullname;
            phoneInput.value = client.phone;
            addressInput.value = client.address;
            if (client.photo_data) {
                photoElement.src = client.photo_data;
                photoDataElement.value = client.photo_data;
                photoElement.classList.remove('hidden');
                photoElement.classList.add('zoomable-image', 'cursor-pointer');
            }
            clientSuggestions.classList.add('hidden');
            phoneInput.readOnly = true;
            addressInput.readOnly = true;
            isExistingClient = true;
        }
        
        fullnameInput.addEventListener('focus', () => {
            if (phoneInput.readOnly) {
                addDebtForm.reset();
                phoneInput.readOnly = false;
                addressInput.readOnly = false;
                photoElement.src = '';
                photoElement.classList.add('hidden');
                photoElement.classList.remove('zoomable-image', 'cursor-pointer');
                photoDataElement.value = '';
                isExistingClient = false;
            }
        });

        // Camera logic
        startCameraButton.addEventListener('click', async () => {
            if (navigator.mediaDevices && navigator.mediaDevices.getUserMedia) {
                try {
                    stream = await navigator.mediaDevices.getUserMedia({ 
                        video: { 
                            width: { ideal: 1280 },
                            height: { ideal: 720 }
                        } 
                    });
                    webcamElement.srcObject = stream;
                    webcamElement.classList.remove('hidden');
                    startCameraButton.classList.add('hidden');
                    capturePhotoButton.classList.remove('hidden');
                } catch (error) {
                    console.error("Camera error", error);
                    alert("Камераны иштетүүдө ката кетти.");
                }
            }
        });

        capturePhotoButton.addEventListener('click', () => {
            canvasElement.width = webcamElement.videoWidth;
            canvasElement.height = webcamElement.videoHeight;
            canvasElement.getContext('2d').drawImage(webcamElement, 0, 0, canvasElement.width, canvasElement.height);
            const photoDataUrl = canvasElement.toDataURL('image/jpeg', 0.95);
            
            photoElement.src = photoDataUrl;
            photoDataElement.value = photoDataUrl;
            webcamElement.classList.add('hidden');
            photoElement.classList.remove('hidden');
            photoElement.classList.add('zoomable-image', 'cursor-pointer');
            capturePhotoButton.classList.add('hidden');
            recapturePhotoButton.classList.remove('hidden');
            if (stream) stream.getTracks().forEach(track => track.stop());
        });

        recapturePhotoButton.addEventListener('click', () => {
            photoElement.classList.add('hidden');
            photoElement.classList.remove('zoomable-image', 'cursor-pointer');
            recapturePhotoButton.classList.add('hidden');
            startCameraButton.click();
        });

        // Form submission
        addDebtForm.addEventListener('submit', async (event) => {
            event.preventDefault();
            const formData = new FormData(addDebtForm);
            const data = Object.fromEntries(formData.entries());
            data.amount = parseFloat(data.amount);

            if (isNaN(data.amount)) {
                alert('Сумманы туура жазыңыз');
                return;
            }

            if (!isExistingClient && !data.photo_data) {
                alert('Жаңы клиент үчүн сүрөткө тартуу милдеттүү.');
                return;
            }

            const response = await fetch('/api/debts/add', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(data),
            });

            if (response.ok) {
                alert('Карыз ийгиликтүү кошулду!');
                addDebtForm.reset();
                phoneInput.readOnly = false;
                addressInput.readOnly = false;
                photoElement.src = '';
                photoElement.classList.add('hidden');
                recapturePhotoButton.classList.add('hidden');
                startCameraButton.classList.remove('hidden');
                if (stream) stream.getTracks().forEach(track => track.stop());
                webcamElement.classList.add('hidden');
                isExistingClient = false;
                loadActiveDebts(1);
            } else {
                const error = await response.json();
                alert(`Ката: ${error.message || 'Белгисиз ката.'}`);
            }
        });

        // Search and Filter Listeners
        searchInput.addEventListener('input', () => loadActiveDebts(1));
        dateFilter.addEventListener('change', () => loadActiveDebts(1));

        loadActiveDebts(1);
    }

    async function loadActiveDebts(page) {
        const searchInput = document.getElementById('search-active');
        const dateFilter = document.getElementById('filter-date-active');
        const search = searchInput ? searchInput.value : '';
        const date = dateFilter ? dateFilter.value : '';

        const container = document.getElementById('active-debts-container');
        let url = `/api/debts?status=active&page=${page}&limit=20`;
        if (search) url += `&search=${encodeURIComponent(search)}`;
        if (date) url += `&date=${date}`;

        const response = await fetch(url);
        const result = await response.json();
        const debts = result.data;
        const total = result.total;
        
        let content = `
            <table class="min-w-full bg-white border border-gray-200">
                <thead class="bg-gray-100">
                    <tr>
                        <th class="py-2 px-4 border-b text-left">№</th>
                        <th class="py-2 px-4 border-b text-left">Сүрөт</th>
                        <th class="py-2 px-4 border-b text-left">ФИО</th>
                        <th class="py-2 px-4 border-b text-left">Телефон</th>
                        <th class="py-2 px-4 border-b text-left">Сумма</th>
                        <th class="py-2 px-4 border-b text-left">Комментарий</th>
                        <th class="py-2 px-4 border-b text-left">Дата</th>
                        <th class="py-2 px-4 border-b text-left">Аракет</th>
                    </tr>
                </thead>
                <tbody id="active-debts-table-body"></tbody>
            </table>
            <div id="pagination-active"></div>`;
        container.innerHTML = content;
        
        const tableBody = document.getElementById('active-debts-table-body');
        tableBody.innerHTML = '';
        
        if (debts && debts.length > 0) {
            debts.forEach((debt, index) => {
                const rowNumber = (page - 1) * 20 + index + 1;
                tableBody.innerHTML += `
                    <tr class="border-b hover:bg-gray-50">
                        <td class="py-2 px-4">${rowNumber}</td>
                        <td class="py-2 px-4">
                            <img src="${debt.photo_data || 'https://via.placeholder.com/80'}" alt="${debt.fullname}" class="w-[80px] h-[80px] object-cover rounded-md zoomable-image cursor-pointer" title="Чоңойтуу үчүн басыңыз">
                        </td>
                        <td class="py-2 px-4 font-semibold">${debt.fullname}</td>
                        <td class="py-2 px-4">${debt.phone}</td>
                        <td class="py-2 px-4 font-bold text-red-600">${debt.amount} сом</td>
                        <td class="py-2 px-4 text-sm text-gray-600 italic">${debt.comment || '-'}</td>
                        <td class="py-2 px-4 text-sm">${new Date(debt.created_at).toLocaleDateString()}</td>
                        <td class="py-2 px-4">
                            <button data-debt-id="${debt.debt_id}" data-client-name="${debt.fullname}" class="pay-debt-btn px-3 py-1 bg-green-500 text-white rounded hover:bg-green-600 text-sm">Жабуу</button>
                        </td>
                    </tr>`;
            });
        } else {
            tableBody.innerHTML = '<tr><td colspan="8" class="text-center py-4">Активдүү карыздар жок.</td></tr>';
        }
        createPagination('pagination-active', page, total, 20, loadActiveDebts);
    }

    // --- Clients Page Logic ---
    function initClientsPage() {
        const searchInput = document.getElementById('search-clients');
        const dateFilter = document.getElementById('filter-date-clients');
        
        searchInput.addEventListener('input', () => loadClients(1));
        dateFilter.addEventListener('change', () => loadClients(1));

        loadClients(1);
    }

    async function loadClients(page) {
        const searchInput = document.getElementById('search-clients');
        const dateFilter = document.getElementById('filter-date-clients');
        const search = searchInput ? searchInput.value : '';
        const date = dateFilter ? dateFilter.value : '';

        const container = document.getElementById('clients-container');
        let url = `/api/clients?page=${page}&limit=200`;
        if (search) url += `&search=${encodeURIComponent(search)}`;
        if (date) url += `&date=${date}`;

        const response = await fetch(url);
        const result = await response.json();
        const clients = result.data;
        const total = result.total;
        
        let content = `
            <table class="min-w-full bg-white">
                <thead class="bg-gray-200">
                    <tr>
                        <th class="py-2 px-4 border-b text-left">№</th>
                        <th class="py-2 px-4 border-b text-left">Сүрөт</th>
                        <th class="py-2 px-4 border-b text-left">Аты-жөнү</th>
                        <th class="py-2 px-4 border-b text-left">Телефон</th>
                        <th class="py-2 px-4 border-b text-left">Дареги</th>
                        <th class="py-2 px-4 border-b text-left">Кошулган күнү</th>
                    </tr>
                </thead>
                <tbody id="clients-table-body"></tbody>
            </table>
            <div id="pagination-clients"></div>`;
        container.innerHTML = content;

        const tableBody = document.getElementById('clients-table-body');
        tableBody.innerHTML = '';
        if (clients && clients.length > 0) {
            clients.forEach((client, index) => {
                const rowNumber = (page - 1) * 200 + index + 1;
                const row = document.createElement('tr');
                row.className = 'border-b hover:bg-gray-100 cursor-pointer';
                row.innerHTML = `
                    <td class="py-2 px-4">${rowNumber}</td>
                    <td class="py-2 px-4">
                        <img src="${client.photo_data || 'https://via.placeholder.com/80'}" class="w-[80px] h-[80px] rounded-full object-cover zoomable-image cursor-pointer" title="Чоңойтуу үчүн басыңыз">
                    </td>
                    <td class="py-2 px-4 font-semibold text-blue-600">${client.fullname}</td>
                    <td class="py-2 px-4">${client.phone}</td>
                    <td class="py-2 px-4">${client.address}</td>
                    <td class="py-2 px-4">${new Date(client.created_at).toLocaleDateString()}</td>
                `;
                
                // Add click event to open details, but ignore if clicking the image
                row.addEventListener('click', (e) => {
                    if (!e.target.classList.contains('zoomable-image')) {
                        openClientDetails(client);
                    }
                });
                
                tableBody.appendChild(row);
            });
        } else {
            tableBody.innerHTML = '<tr><td colspan="6" class="text-center py-4">Клиенттер жок.</td></tr>';
        }
        createPagination('pagination-clients', page, total, 200, loadClients);
    }

    // --- Client Details Modal Logic ---
    async function openClientDetails(client) {
        document.getElementById('details-client-name').textContent = client.fullname;
        
        // Load Active Debts
        const activeRes = await fetch(`/api/debts?client_id=${client.id}&status=active&limit=100`);
        const activeResult = await activeRes.json();
        renderClientDebtsTable('client-active-debts', activeResult.data, true);

        // Load History Debts
        const historyRes = await fetch(`/api/debts?client_id=${client.id}&status=paid&limit=100`);
        const historyResult = await historyRes.json();
        renderClientDebtsTable('client-history-debts', historyResult.data, false);

        clientDetailsModal.classList.remove('hidden');
        clientDetailsModal.classList.add('flex');
    }

    function renderClientDebtsTable(containerId, debts, isActive) {
        const container = document.getElementById(containerId);
        if (!debts || debts.length === 0) {
            container.innerHTML = '<p class="text-gray-500 italic">Маалымат жок.</p>';
            return;
        }

        let html = `
            <table class="min-w-full bg-white border">
                <thead class="bg-gray-100">
                    <tr>
                        <th class="py-2 px-4 text-left">Сумма</th>
                        <th class="py-2 px-4 text-left">Дата</th>
                        <th class="py-2 px-4 text-left">Коммент</th>
                        ${!isActive ? '<th class="py-2 px-4 text-left">Төлөнгөн</th><th class="py-2 px-4 text-left">Баа</th>' : ''}
                        ${isActive ? '<th class="py-2 px-4 text-left">Аракет</th>' : ''}
                    </tr>
                </thead>
                <tbody>`;
        
        debts.forEach(debt => {
            html += `
                <tr class="border-b">
                    <td class="py-2 px-4 font-bold text-red-600">${debt.amount} сом</td>
                    <td class="py-2 px-4">${new Date(debt.created_at).toLocaleDateString()}</td>
                    <td class="py-2 px-4 text-sm">${debt.comment || '-'}</td>
                    ${!isActive ? `<td class="py-2 px-4">${new Date(debt.paid_at).toLocaleDateString()}</td><td class="py-2 px-4">${getRatingBadge(debt.rating)}</td>` : ''}
                    ${isActive ? `<td class="py-2 px-4"><button data-debt-id="${debt.debt_id}" data-client-name="${debt.fullname}" class="pay-debt-btn px-3 py-1 bg-green-500 text-white rounded text-sm hover:bg-green-600">Жабуу</button></td>` : ''}
                </tr>`;
        });
        html += '</tbody></table>';
        container.innerHTML = html;
    }

    document.getElementById('close-details-modal').addEventListener('click', () => {
        clientDetailsModal.classList.add('hidden');
        clientDetailsModal.classList.remove('flex');
    });

    // --- History Page Logic ---
    function initHistoryPage() {
        const searchInput = document.getElementById('search-history');
        const dateFilter = document.getElementById('filter-date-history');
        
        searchInput.addEventListener('input', () => loadHistory(1));
        dateFilter.addEventListener('change', () => loadHistory(1));

        loadHistory(1);
    }

    async function loadHistory(page) {
        const searchInput = document.getElementById('search-history');
        const dateFilter = document.getElementById('filter-date-history');
        const search = searchInput ? searchInput.value : '';
        const date = dateFilter ? dateFilter.value : '';

        const container = document.getElementById('history-container');
        let url = `/api/debts?status=paid&page=${page}&limit=200`;
        if (search) url += `&search=${encodeURIComponent(search)}`;
        if (date) url += `&date=${date}`;

        const response = await fetch(url);
        const result = await response.json();
        const debts = result.data;
        const total = result.total;

        let content = `
            <table class="min-w-full bg-white">
                <thead class="bg-gray-200">
                    <tr>
                        <th class="py-2 px-4">Аты-жөнү</th>
                        <th class="py-2 px-4">Сумма</th>
                        <th class="py-2 px-4">Алынган күнү</th>
                        <th class="py-2 px-4">Төлөнгөн күнү</th>
                        <th class="py-2 px-4">Баа</th>
                        <th class="py-2 px-4">Коммент</th>
                    </tr>
                </thead>
                <tbody id="history-table-body"></tbody>
            </table>
            <div id="pagination-history"></div>`;
        container.innerHTML = content;

        const tableBody = document.getElementById('history-table-body');
        tableBody.innerHTML = '';
        if (debts && debts.length > 0) {
            debts.forEach(debt => {
                tableBody.innerHTML += `
                    <tr class="border-b">
                        <td class="py-2 px-4">${debt.fullname}</td>
                        <td class="py-2 px-4">${debt.amount} сом</td>
                        <td class="py-2 px-4">${new Date(debt.created_at).toLocaleDateString()}</td>
                        <td class="py-2 px-4">${debt.paid_at ? new Date(debt.paid_at).toLocaleDateString() : '-'}</td>
                        <td class="py-2 px-4">${getRatingBadge(debt.rating)}</td>
                        <td class="py-2 px-4 text-sm text-gray-500">${debt.comment || '-'}</td>
                    </tr>`;
            });
        } else {
            tableBody.innerHTML = '<tr><td colspan="6" class="text-center py-4">Тарых жок.</td></tr>';
        }
        createPagination('pagination-history', page, total, 200, loadHistory);
    }
    
    // --- Modal Logic ---
    document.body.addEventListener('click', (event) => {
        if (event.target.classList.contains('pay-debt-btn')) {
            currentDebtToPay = event.target.dataset.debtId;
            document.getElementById('modal-client-name').textContent = event.target.dataset.clientName;
            payDebtModal.style.display = 'flex';
        }
    });

    document.getElementById('cancel-pay').addEventListener('click', () => {
        payDebtModal.style.display = 'none';
        currentDebtToPay = null;
    });

    document.getElementById('confirm-pay').addEventListener('click', async () => {
        if (!currentDebtToPay) return;
        const rating = document.getElementById('rating').value;
        const response = await fetch('/api/debts/pay', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ debt_id: parseInt(currentDebtToPay), rating: rating }),
        });
        if (response.ok) {
            alert('Карыз ийгиликтүү жабылды!');
            payDebtModal.style.display = 'none';
            // Refresh current view
            const currentPath = window.location.pathname;
            if (currentPath === '/') loadActiveDebts(1);
            // If inside client details modal, refresh that too
            if (!clientDetailsModal.classList.contains('hidden')) {
                clientDetailsModal.classList.add('hidden'); 
                clientDetailsModal.classList.remove('flex');
                if (currentPath === '/clients') loadClients(1);
            }
        } else {
            alert('Карызды жабууда ката кетти.');
        }
        currentDebtToPay = null;
    });

    // --- Initial Render ---
    render(window.location.pathname);
});
