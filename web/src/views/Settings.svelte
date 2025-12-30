<script>
    import { onMount, getContext } from "svelte";
    const shopify = getContext("shopify");

    let services = $state({
        loading: true,
        data: [],
        formData: {
            name: "",
            callbackUrl: "",
        },
    });

    onMount(() => {
        (async () => {
            try {
                const res = await shopify.fetch("/api/carrier-service", {
                    method: "GET",
                });
                const data = await res.json();
                services.data = data;
                services.loading = false;
            } catch (e) {
                console.error("failed to get carrier services:", e);
            }
        })();
    });

    function printUserErros(userErrors) {
        for (const { message } of userErrors) {
            console.error(message);
        }
    }

    async function deleteCarrierService(id) {
        const encodedId = encodeURIComponent(id);
        try {
            const res = await shopify.fetch(
                "/api/carrier-service/" + encodedId,
                {
                    method: "DELETE",
                },
            );
            const { deletedId, userErrors } = await res.json();
            if (userErrors.length > 0) {
                printUserErros(userErrors);
                return;
            }
            services.data = services.data.filter(
                (service) => service.id !== deletedId,
            );
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
            const { carrierService, userErrors } = await res.json();
            if (userErrors.length > 0) {
                printUserErros(userErrors);
                return;
            }
            services.data = [carrierService, ...services.data];
        } catch (e) {
            console.error("failed to create carrier service:", e);
        }
    }
</script>

<h1>Settings</h1>

<s-section heading="Carrier service" padding="base">
    <s-table>
        <s-table-header-row>
            <s-table-header listSlot="primary">Name</s-table-header>
            <s-table-header listSlot="inline">CallbackUrl</s-table-header>
            <s-table-header listSlot="inline" format="numeric">
                <s-button variant="primary" commandFor="carrierServiceModal">
                    New service
                </s-button>
            </s-table-header>
        </s-table-header-row>
        <s-table-body>
            {#if services.loading}
                <s-table-row>
                    <s-table-cell>
                        <s-spinner
                            size="base"
                            accessibility-label="Loading services"
                        ></s-spinner>
                    </s-table-cell>
                </s-table-row>
            {/if}
            {#each services.data as service}
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
        <s-text-field
            label="Name"
            placeholder="name"
            value={services.formData.name}
            onchange={(e) => {
                services.formData.name = e.target.value;
            }}
        ></s-text-field>
        <s-url-field
            label="Callback url"
            placeholder="url"
            value={services.formData.callbackUrl}
            onchange={(e) => {
                services.formData.callbackUrl = e.target.value;
            }}
        ></s-url-field>
    </s-stack>

    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <s-button
        slot="secondary-actions"
        commandFor="carrierServiceModal"
        command="--hide"
        onclick={() => {
            services.formData = {
                name: "",
                callbackUrl: "",
            };
        }}
    >
        Close
    </s-button>
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <s-button
        slot="primary-action"
        variant="primary"
        commandFor="carrierServiceModal"
        command="--hide"
        onclick={() => {
            createCarrierService(
                services.formData.name,
                services.formData.callbackUrl,
            );
        }}
    >
        Create
    </s-button>
</s-modal>
