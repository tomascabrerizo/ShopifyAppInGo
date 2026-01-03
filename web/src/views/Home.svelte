<script>
    import { getContext } from "svelte";
    import { view } from "../stores/views";
    import { orders } from "../stores/orders";
    import { utils } from "../utils";

    const shopify = getContext("shopify");
</script>

<s-section heading="Pedidos" padding="base">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">Pedido</s-table-header>
            <s-table-header listSlot="secondary"
                >Estado de la orden</s-table-header
            >
            <s-table-header listSlot="inline">Fecha</s-table-header>
            <s-table-header listSlot="inline">Cliente</s-table-header>
            <s-table-header listSlot="inline">Estado de pago</s-table-header>
            <s-table-header listSlot="inline" format="currency"
                >Total</s-table-header
            >
        </s-table-header-row>
        <s-table-body>
            {#each $orders as order}
                <s-table-row key={order.order_id}>
                    <s-table-cell>
                        <!-- svelte-ignore a11y_no_static_element_interactions -->
                        <s-stack
                            direction="inline"
                            gap="small"
                            alignItems="center"
                        >
                            <s-checkbox checked={false} id={order.order_id}
                            ></s-checkbox>
                            <!-- svelte-ignore a11y_click_events_have_key_events -->
                            <s-link
                                onclick={(e) => {
                                    e.preventDefault();
                                    view.order(order.order_id);
                                }}>{order.order_id}</s-link
                            >
                        </s-stack>
                    </s-table-cell>

                    <s-table-cell>
                        <s-badge tone={!order.fulfilled ? "warning" : "success"}
                            >{order.fulfilled
                                ? "Confirmado"
                                : "Sin confirmar"}</s-badge
                        >
                    </s-table-cell>

                    <s-table-cell>{order.created_at}</s-table-cell>

                    <s-table-cell>{order.shipping_address.name}</s-table-cell>

                    <s-table-cell>
                        <s-badge tone={!order.paid ? "warning" : "success"}
                            >{order.paid ? "pagado" : "sin pagar"}</s-badge
                        >
                    </s-table-cell>

                    <s-table-cell
                        >{utils.formatPrice(order.total_price)}</s-table-cell
                    >
                </s-table-row>
            {/each}
        </s-table-body>
    </s-table>
</s-section>
