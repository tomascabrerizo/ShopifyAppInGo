<script>
  import createApp from "@shopify/app-bridge";
  import { authenticatedFetch } from "@shopify/app-bridge/utilities";
  import { onMount } from "svelte";

  let app;
  let fetch;

  let orders = $state([]);
  $inspect(orders);

  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search);
    app = createApp({
      apiKey: "2298e2e615e8c35791b251ed4504b203",
      host: urlParams.get("host"),
      forceRedirect: true,
    });
    fetch = authenticatedFetch(app);

    (async () => {
      const res = await fetch("/api/orders");
      const data = await res.json();
      orders = data;
    })();
  });
</script>

<main>
  <s-page heading="Pedidos" inlineSize="large">
    <s-table>
      <s-table-header-row>
        <s-table-header listSlot="primary">Pedido</s-table-header>
        <s-table-header listSlot="secondary">Estado de la orden</s-table-header>
        <s-table-header listSlot="inline">Fecha</s-table-header>
        <s-table-header listSlot="inline">Cliente</s-table-header>
        <s-table-header listSlot="inline">Estado de pago</s-table-header>
        <s-table-header listSlot="inline">Total</s-table-header>
      </s-table-header-row>
      <s-table-body>
        {#each orders as order}
          <s-table-row key={order.order_id}>
            <s-table-cell>
              <s-stack direction="inline" gap="small" alignItems="center">
                <s-checkbox checked={false} id={order.order_id}></s-checkbox>
                <s-text type="strong">{order.order_id}</s-text>
              </s-stack>
            </s-table-cell>

            <s-table-cell>
              <s-badge tone={!order.fulfilled ? "warning" : "success"}
                >{order.fulfilled ? "Confirmado" : "Sin confirmar"}</s-badge
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
</main>
