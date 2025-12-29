<script>
    import { onMount, getContext } from "svelte";

    const shopify = getContext("shopify");
    $inspect(shopify);

    let services = $state([]);
    $inspect(services);

    onMount(() => {
        (async () => {
            try {
                const res = await shopify.fetch("/api/carrier-service", {
                    method: "GET",
                });
                const data = await res.json();
                services = data;
            } catch (e) {
                console.error("failed to get carrier services:", e);
            }
        })();
    });

    async function deleteCarrierService(id) {
        const body = {
            id: id,
        };
        try {
            const res = await shopify.fetch("/api/carrier-service", {
                method: "DELETE",
                body: JSON.stringify(body),
            });
            const data = await res.json();
        } catch (e) {
            console.error("failed to delete carrier service:", e);
        }
    }

    async function createCarrierService(name, callbackUrl) {
        const body = {
            name: name,
            callbackUrl: callbackUrl,
        };
        try {
            const res = await shopify.fetch("/api/carrier-service", {
                method: "POST",
                body: JSON.stringify(body),
            });
            const data = await res.json();
        } catch (e) {
            console.error("failed to create carrier service:", e);
        }
    }
</script>

<h1>Settings</h1>

<s-section padding="none">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">Name</s-table-header>
            <s-table-header listSlot="inline">CallbackUrl</s-table-header>
            <s-table-header listSlot="inline" format="numeric"
                ><s-button variant="primary" commandFor="carrierServiceModal"
                    >New service</s-button
                ></s-table-header
            >
        </s-table-header-row>
        <s-table-body>
            {#each services as service}
                <s-table-row>
                    <s-table-cell>{service.name}</s-table-cell>
                    <s-table-cell>{service.callbackUrl}</s-table-cell>
                    <s-table-cell>
                        <!-- svelte-ignore a11y_click_events_have_key_events -->
                        <!-- svelte-ignore a11y_no_static_element_interactions -->
                        <s-button
                            variant="primary"
                            onclick={() => deleteCarrierService(service.id)}
                        >
                            <s-icon type="delete"></s-icon>
                        </s-button>
                    </s-table-cell>
                </s-table-row>
            {/each}
        </s-table-body>
    </s-table>
</s-section>

<s-modal id="carrierServiceModal" heading="New service">
    <s-stack gap="base">
        <s-text-field label="Name" value="" placeholder="name"></s-text-field>
        <s-url-field label="Callback url" placeholder="url"></s-url-field>
    </s-stack>

    <s-button
        slot="secondary-actions"
        commandFor="carrierServiceModal"
        command="--hide"
    >
        Close
    </s-button>
    <s-button
        slot="primary-action"
        variant="primary"
        commandFor="carrierServiceModal"
        command="--hide"
    >
        Create
    </s-button>
</s-modal>
