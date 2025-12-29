<script>
    import { onMount, getContext } from "svelte";

    const shopify = getContext("shopify");
    $inspect(shopify);

    let orders = $state([]);
    $inspect(orders);

    onMount(() => {
        (async () => {
            const res = await shopify.fetch("/api/orders");
            const data = await res.json();
            orders = data;
        })();
    });
</script>

<h1>Home</h1>
<s-page heading="Pedidos" inlineSize="large">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">Pedido</s-table-header>
            <s-table-header listSlot="secondary"
                >Estado de la orden</s-table-header
            >
            <s-table-header listSlot="inline">Fecha</s-table-header>
            <s-table-header listSlot="inline">Cliente</s-table-header>
            <s-table-header listSlot="inline">Estado de pago</s-table-header>
            <s-table-header listSlot="inline">Total</s-table-header>
        </s-table-header-row>
        <s-table-body>
            {#each orders as order}
                <s-table-row key={order.order_id}>
                    <s-table-cell>
                        <s-stack
                            direction="inline"
                            gap="small"
                            alignItems="center"
                        >
                            <s-checkbox checked={false} id={order.order_id}
                            ></s-checkbox>
                            <s-text type="strong">{order.order_id}</s-text>
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
                        >{new Intl.NumberFormat("es-AR", {
                            style: "currency",
                            currency: "ARS",
                            minimumFractionDigits: 2,
                        }).format(order.total_price / 100)}</s-table-cell
                    >
                </s-table-row>
            {/each}
        </s-table-body>
    </s-table>
</s-page>
