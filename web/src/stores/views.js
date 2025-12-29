import { writable } from "svelte/store";

export const VIEW_HOME = "home"
export const VIEW_SETTINGS = "settings"
export const VIEW_ORDER = "orders"

// TODO: maybe get last view from session storage
const { subscribe, set } = writable({
    name: VIEW_HOME,
    params: {},
})

function home() {
    set({ name: VIEW_HOME, params: {} })
}

function settings() {
    set({ name: VIEW_SETTINGS, params: {} })
}

function order(id) {
    set({ name: VIEW_ORDER, params: { id: id } })
}

export const view = {
    subscribe,
    home,
    settings,
    order,
}


