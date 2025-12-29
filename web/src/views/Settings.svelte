<script>
    import { onMount, getContext } from "svelte";

    const shopify = getContext("shopify");
    $inspect(shopify);

    onMount(() => {
        (async () => {
            const res = await shopify.fetch("/api/carrier-service", {
                method: "GET",
            });

            const data = await res.json();
            console.log(data);

            // const body = {
            //     name: "halfpipe",
            //     callbackUrl: "https://halfpipe.dev",
            // };

            // const res = await shopify.fetch("/api/carrier-service", {
            //     method: "POST",
            //     body: JSON.stringify(body),
            // });

            //const data = await res.json();
            //console.log(data);
        })();
    });
</script>

<h1>Settings</h1>
