function formatPrice(value) {
    return new Intl.NumberFormat("es-AR", {
        style: "currency",
        currency: "ARS",
        minimumFractionDigits: 2,
    }).format(value / 100);
}

export const utils = {
    formatPrice,
}
