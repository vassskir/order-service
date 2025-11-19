const API_BASE_URL = window.location.origin; 
const ORDER_API_URL = `${API_BASE_URL}/order`;

const searchInput = document.getElementById('orderId');
const searchButton = document.querySelector('.search-button');
const resultsSection = document.getElementById('results');
const loadingElement = document.getElementById('loading');
const errorElement = document.getElementById('error');
const errorMessage = document.getElementById('errorMessage');

const orderIdDisplay = document.getElementById('orderIdDisplay');
const basicInfoElement = document.getElementById('basicInfo');
const deliveryInfoElement = document.getElementById('deliveryInfo');
const paymentInfoElement = document.getElementById('paymentInfo');
const itemsListElement = document.getElementById('itemsList');

async function searchOrder() {
    const orderId = searchInput.value.trim();
    
    if (!orderId) {
        showError('Пожалуйста, введите ID заказа');
        return;
    }

    showLoading();
    hideResults();
    hideError();

    try {
        const order = await fetchOrder(orderId);
        displayOrder(order);
        showResults();
    } catch (error) {
        showError(error.message);
    } finally {
        hideLoading();
    }
}

async function fetchOrder(orderId) {
    const response = await fetch(`${ORDER_API_URL}/${orderId}`);
    
    if (!response.ok) {
        if (response.status === 404) {
            throw new Error('Заказ с таким ID не найден');
        } else if (response.status === 400) {
            throw new Error('Неверный формат ID заказа');
        } else {
            throw new Error('Произошла ошибка при поиске заказа');
        }
    }
    
    return await response.json();
}

function displayOrder(order) {
    displayOrderId(order.order_uid);
    displayBasicInfo(order);
    displayDeliveryInfo(order.delivery, order.delivery_service);
    displayPaymentInfo(order.payment);
    displayItems(order.items);
}

function displayOrderId(orderId) {
    orderIdDisplay.textContent = orderId;
}

function displayBasicInfo(order) {
    const basicInfo = [
        { label: 'Трек номер', value: order.track_number },
        { label: 'Язык', value: order.locale },
        { label: 'ID клиента', value: order.customer_id },
        { label: 'Дата создания', value: formatDate(order.date_created) },
        { label: 'Внутренняя подпись', value: order.internal_signature || '—' },
        { label: 'Shard Key', value: order.shardkey },
        { label: 'SM ID', value: order.sm_id },
        { label: 'OOF Shard', value: order.oof_shard }
    ];

    basicInfoElement.innerHTML = basicInfo.map(info => `
        <div class="info-item">
            <span class="info-label">${info.label}:</span>
            <span class="info-value">${info.value}</span>
        </div>
    `).join('');
}

function displayDeliveryInfo(delivery, deliveryService) {
    const deliveryInfo = [
        { label: 'Служба доставки', value: deliveryService },
        { label: 'Получатель', value: delivery.name },
        { label: 'Телефон', value: formatPhone(delivery.phone) },
        { label: 'Email', value: delivery.email },
        { label: 'Адрес', value: delivery.address },
        { label: 'Город', value: delivery.city },
        { label: 'Регион', value: delivery.region },
        { label: 'Индекс', value: delivery.zip }
    ];

    deliveryInfoElement.innerHTML = deliveryInfo.map(info => `
        <div class="info-item">
            <span class="info-label">${info.label}:</span>
            <span class="info-value">${info.value}</span>
        </div>
    `).join('');
}

function displayPaymentInfo(payment) {
    const paymentInfo = [
        { label: 'Транзакция', value: payment.transaction },
        { label: 'Валюта', value: payment.currency },
        { label: 'Провайдер', value: payment.provider },
        { label: 'Сумма', value: formatCurrency(payment.amount, payment.currency) },
        { label: 'Банк', value: payment.bank },
        { label: 'Стоимость доставки', value: formatCurrency(payment.delivery_cost, payment.currency) },
        { label: 'Стоимость товаров', value: formatCurrency(payment.goods_total, payment.currency) },
        { label: 'Комиссия', value: formatCurrency(payment.custom_fee, payment.currency) },
        { label: 'Дата оплаты', value: formatTimestamp(payment.payment_dt) },
        { label: 'Request ID', value: payment.request_id || '—' }
    ];

    paymentInfoElement.innerHTML = paymentInfo.map(info => `
        <div class="info-item">
            <span class="info-label">${info.label}:</span>
            <span class="info-value">${info.value}</span>
        </div>
    `).join('');
}

function displayItems(items) {
    if (!items || items.length === 0) {
        itemsListElement.innerHTML = '<p>Товары не найдены</p>';
        return;
    }

    itemsListElement.innerHTML = items.map(item => `
        <div class="item-card">
            <div class="item-image">🛍️</div>
            <div class="item-details">
                <div class="item-name">${item.name}</div>
                <div class="item-meta">
                    <span>Бренд: ${item.brand}</span>
                    <span>Цена: ${formatCurrency(item.price, 'USD')}</span>
                    <span>Скидка: ${item.sale}%</span>
                    <span>Итого: ${formatCurrency(item.total_price, 'USD')}</span>
                </div>
                <div class="item-meta">
                    <span>Артикул: ${item.chrt_id}</span>
                    <span>RID: ${item.rid}</span>
                    <span>NM ID: ${item.nm_id}</span>
                    <span>Статус: ${item.status}</span>
                    <span>Размер: ${item.size}</span>
                </div>
            </div>
        </div>
    `).join('');
}

function formatDate(dateString) {
    try {
        const date = new Date(dateString);
        return date.toLocaleDateString('ru-RU', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    } catch {
        return dateString;
    }
}

function formatTimestamp(timestamp) {
    try {
        const date = new Date(timestamp * 1000);
        return date.toLocaleDateString('ru-RU', {
            year: 'numeric',
            month: 'long',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        });
    } catch {
        return timestamp;
    }
}

function formatPhone(phone) {
    return phone.replace(/(\d{1})(\d{3})(\d{3})(\d{2})(\d{2})/, '+$1 ($2) $3-$4-$5');
}

function formatCurrency(amount, currency) {
    return new Intl.NumberFormat('ru-RU', {
        style: 'currency',
        currency: currency || 'USD'
    }).format(amount / 100); 
}

function showLoading() {
    loadingElement.classList.remove('hidden');
}

function hideLoading() {
    loadingElement.classList.add('hidden');
}

function showResults() {
    resultsSection.classList.remove('hidden');
}

function hideResults() {
    resultsSection.classList.add('hidden');
}

function showError(message) {
    errorMessage.textContent = message;
    errorElement.classList.remove('hidden');
}

function hideError() {
    errorElement.classList.add('hidden');
}

function retrySearch() {
    searchInput.value = ''; 
    searchInput.focus(); 
    hideError();
}

searchButton.addEventListener('click', searchOrder);

searchInput.addEventListener('keypress', (event) => {
    if (event.key === 'Enter') {
        searchOrder();
    }
});

window.searchOrder = searchOrder;
window.retrySearch = retrySearch;

document.addEventListener('DOMContentLoaded', () => {
    hideLoading();
    hideResults();
    hideError();
});