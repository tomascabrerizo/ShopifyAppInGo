<script>
  import createApp from "@shopify/app-bridge";
  import { authenticatedFetch } from "@shopify/app-bridge/utilities";
  import { onMount } from "svelte";

  let app;
  let fetch;
  onMount(() => {
    const urlParams = new URLSearchParams(window.location.search);
    app = createApp({
      apiKey: "2298e2e615e8c35791b251ed4504b203",
      host: urlParams.get("host"),
      forceRedirect: true,
    });
    fetch = authenticatedFetch(app);

    (async () => {
      await fetch("/test");
    })();
  });
</script>

<main>
  <h1>Shopify Svelte and go</h1>
</main>
