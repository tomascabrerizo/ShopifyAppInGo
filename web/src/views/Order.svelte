<script>
    import { getContext, onMount } from "svelte";
    import { view } from "../stores/views";
    import { orders } from "../stores/orders";
    import { utils } from "../utils";

    const shopify = getContext("shopify");
    $inspect(shopify);

    let order = $derived.by(() => {
        const id = $view.params.id;
        return $orders.find((order) => order.order_id === id);
    });
    let fulfillments = $state({
        loading: true,
        data: [],
    });

    onMount(() => {
        const id = `gid://shopify/Order/${order.order_id}`;
        const encodedId = encodeURIComponent(id);
        (async () => {
            try {
                const res = await shopify.fetch(
                    `/api/orders/${encodedId}/fulfillments`,
                );
                const data = await res.json();
                fulfillments = {
                    loading: false,
                    data: data.nodes,
                };
            } catch (e) {
                console.error("failed to fetch order fulfillment:", e);
            }
        })();
    });
</script>

<s-section heading="Order" padding="base">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">id</s-table-header>
            <s-table-header listSlot="secondary">currency</s-table-header>
            <s-table-header listSlot="inline">carrier name</s-table-header>
            <s-table-header listSlot="inline">carrier code</s-table-header>
            <s-table-header listSlot="inline" format="currency"
                >shipping price</s-table-header
            >
            <s-table-header listSlot="inline" format="currency"
                >total price</s-table-header
            >
        </s-table-header-row>
        <s-table-body>
            <s-table-row>
                <s-table-cell>{order.order_id}</s-table-cell>
                <s-table-cell>{order.currency}</s-table-cell>
                <s-table-cell>{order.carrier_name}</s-table-cell>
                <s-table-cell>{order.carrier_code}</s-table-cell>
                <s-table-cell
                    >{utils.formatPrice(order.carrier_price)}</s-table-cell
                >
                <s-table-cell
                    >{utils.formatPrice(order.total_price)}</s-table-cell
                >
            </s-table-row>
        </s-table-body>
    </s-table>
</s-section>
<s-section heading="Items" padding="base">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">id</s-table-header>
            <s-table-header listSlot="secondary">name</s-table-header>
            <s-table-header listSlot="inline">grams</s-table-header>
            <s-table-header listSlot="inline">quantity</s-table-header>
            <s-table-header listSlot="inline" format="currency"
                >price</s-table-header
            >
        </s-table-header-row>
        <s-table-body>
            {#each order.items as item}
                <s-table-row key={item.item_id}>
                    <s-table-cell>{item.item_id}</s-table-cell>
                    <s-table-cell>{item.name}</s-table-cell>
                    <s-table-cell>{item.grams}</s-table-cell>
                    <s-table-cell>{item.quantity}</s-table-cell>
                    <s-table-cell>{utils.formatPrice(item.price)}</s-table-cell>
                </s-table-row>
            {/each}
        </s-table-body>
    </s-table>
</s-section>
<s-section heading="Destination" padding="base">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">address</s-table-header>
            <s-table-header listSlot="inline">zip</s-table-header>
            <s-table-header listSlot="inline">country</s-table-header>
            <s-table-header listSlot="inline">province</s-table-header>
            <s-table-header listSlot="inline">city</s-table-header>
            <s-table-header listSlot="inline">name</s-table-header>
            <s-table-header listSlot="inline">lastname</s-table-header>
            <s-table-header listSlot="inline">phone</s-table-header>
            <s-table-header listSlot="inline">email</s-table-header>
        </s-table-header-row>
        <s-table-body>
            <s-table-row key={order.order_id}>
                <s-table-cell>{order.shipping_address.address1}</s-table-cell>
                <s-table-cell>{order.shipping_address.zip}</s-table-cell>
                <s-table-cell>{order.shipping_address.country}</s-table-cell>
                <s-table-cell>{order.shipping_address.province}</s-table-cell>
                <s-table-cell>{order.shipping_address.city}</s-table-cell>
                <s-table-cell>{order.shipping_address.name}</s-table-cell>
                <s-table-cell>{order.shipping_address.last_name}</s-table-cell>
                <s-table-cell>{order.shipping_address.phone}</s-table-cell>
                <s-table-cell>{order.shipping_address.email}</s-table-cell>
            </s-table-row>
        </s-table-body>
    </s-table>
</s-section>

<s-stack gap="base">
    <s-heading>Fulfillments:</s-heading>
    {#if fulfillments.loading}
        <s-spinner size="base" accessibility-label="Loading fulfillments"
        ></s-spinner>
    {/if}

    {#each fulfillments.data as fulfillment}
        {@const location = fulfillment.assignedLocation.location}
        {@const address = location.address}
        {@const actions = fulfillment.supportedActions.map((a) => a.action)}
        <s-section padding="base">
            <s-stack gap="base">
                <s-grid gap="small-200" gridTemplateColumns="1fr auto">
                    <s-heading>{address.address1}:</s-heading>
                    <s-button variant="primary">Generar orden de envio</s-button
                    >
                </s-grid>
                <s-divider color="strong"></s-divider>
                <s-table>
                    <s-table-header-row>
                        <s-table-header listSlot="primary"
                            >actions</s-table-header
                        >
                        <s-table-header listSlot="inline"
                            >address</s-table-header
                        >
                        <s-table-header listSlot="inline">zip</s-table-header>
                        <s-table-header listSlot="inline"
                            >country</s-table-header
                        >
                        <s-table-header listSlot="inline"
                            >province</s-table-header
                        >
                        <s-table-header listSlot="inline">city</s-table-header>
                        <s-table-header listSlot="inline">name</s-table-header>
                    </s-table-header-row>
                    <s-table-body>
                        <s-table-row key={fulfillment.id}>
                            <s-table-cell>
                                <s-stack gap="small">
                                    {#each actions as action}
                                        <s-badge
                                            tone={action ===
                                            "CREATE_FULFILLMENT"
                                                ? "success"
                                                : "info"}>{action}</s-badge
                                        >
                                    {/each}
                                </s-stack>
                            </s-table-cell>
                            <s-table-cell>{address.address1}</s-table-cell>
                            <s-table-cell>{address.zip}</s-table-cell>
                            <s-table-cell>{address.country}</s-table-cell>
                            <s-table-cell>{address.province}</s-table-cell>
                            <s-table-cell>{address.city}</s-table-cell>
                            <s-table-cell>{location.name}</s-table-cell>
                        </s-table-row>
                    </s-table-body>
                </s-table>
                <s-divider color="strong"></s-divider>
                <s-heading>Line items:</s-heading>
                <s-table>
                    <s-table-header-row>
                        <s-table-header listSlot="primary">title</s-table-header
                        >
                        <s-table-header listSlot="inline"
                            >remaining quantity</s-table-header
                        >
                        <s-table-header listSlot="inline"
                            >total quantity</s-table-header
                        >
                    </s-table-header-row>
                    <s-table-body>
                        {#each fulfillment.lineItems.nodes as lineItem}
                            {@const product = lineItem.lineItem.product}
                            <s-table-row key={lineItem.id}>
                                <s-table-cell>{product.title}</s-table-cell>
                                <s-table-cell
                                    >{lineItem.remainingQuantity}</s-table-cell
                                >
                                <s-table-cell
                                    >{lineItem.totalQuantity}</s-table-cell
                                >
                            </s-table-row>
                        {/each}
                    </s-table-body>
                </s-table>
            </s-stack>
        </s-section>
    {/each}
</s-stack>

<!-- <pre>{JSON.stringify(order, null, 2)}</pre> -->
