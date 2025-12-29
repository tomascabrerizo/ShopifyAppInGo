<script>
  import { setContext } from "svelte";
  import { view, VIEW_HOME, VIEW_SETTINGS, VIEW_ORDER } from "./stores/views";
  import Home from "./views/Home.svelte";
  import Order from "./views/Order.svelte";
  import Settings from "./views/Settings.svelte";

  import createApp from "@shopify/app-bridge";
  import { authenticatedFetch } from "@shopify/app-bridge/utilities";

  const urlParams = new URLSearchParams(window.location.search);
  const app = createApp({
    apiKey: "2298e2e615e8c35791b251ed4504b203",
    host: urlParams.get("host"),
    forceRedirect: true,
  });
  setContext("shopify", {
    app: app,
    fetch: authenticatedFetch(app),
  });
</script>

<main>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <s-button onclick={() => view.home()}> home </s-button>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <s-button onclick={() => view.settings()}> settings </s-button>
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <s-button onclick={() => view.order(0)}> order </s-button>

  {#if $view.name === VIEW_HOME}
    <Home />
  {:else if $view.name === VIEW_SETTINGS}
    <Settings />
  {:else if $view.name === VIEW_ORDER}
    <Order />
  {/if}
</main>
