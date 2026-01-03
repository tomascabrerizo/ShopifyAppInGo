import { writable } from "svelte/store";

const { subscribe, update, set } = writable([])

function add(order) {
    update(orders => [order, ...orders]);
}

function remove(orderId) {
    update(orders => orders.filter(order => order.order_id !== orderId));
}

function updateOrder(newOrder) {
    update(orders => orders.map(order => order.order_id === newOrder.order_id ? newOrder : order));
}

export const orders = {
    subscribe,
    set,
    add,
    remove,
    update: updateOrder,
}