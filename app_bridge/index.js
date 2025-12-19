import createApp from '@shopify/app-bridge';
import { Redirect } from '@shopify/app-bridge/actions';

const config = {
    apiKey: "2298e2e615e8c35791b251ed4504b203",
    host: new URLSearchParams(location.search).get("host"),
    forceRedirect: false
};

const app = createApp(config);
const redirect = Redirect.create(app);

function ExitIFrame() {
    redirect.dispatch(Redirect.Action.REMOTE, window.location.href);
}

function NoIFrameRedirect() {
    const params = new URLSearchParams(window.location.search);
    const redirectUriParams = new URLSearchParams(params).toString();
    const redirectUri = `/api/auth?${redirectUriParams}`;
    window.location.href = redirectUri;
}

window.addEventListener("DOMContentLoaded", () => {
    const params = new URLSearchParams(window.location.search);
    const embedded = params.get("embedded");
    const inIframe = window.top !== window.self;
    if (embedded === "1" && inIframe) {
        ExitIFrame();
        return
    }
    NoIFrameRedirect()
})
