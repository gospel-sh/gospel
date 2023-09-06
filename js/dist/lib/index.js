var __awaiter = (this && this.__awaiter) || function (thisArg, _arguments, P, generator) {
    function adopt(value) { return value instanceof P ? value : new P(function (resolve) { resolve(value); }); }
    return new (P || (P = Promise))(function (resolve, reject) {
        function fulfilled(value) { try { step(generator.next(value)); } catch (e) { reject(e); } }
        function rejected(value) { try { step(generator["throw"](value)); } catch (e) { reject(e); } }
        function step(result) { result.done ? resolve(result.value) : adopt(result.value).then(fulfilled, rejected); }
        step((generator = generator.apply(thisArg, _arguments || [])).next());
    });
};
function handleClick(e) {
    let target = e.target;
    if (target instanceof HTMLElement && target.tagName !== 'a') {
        target = target.closest('a');
    }
    if (target === null) {
        return;
    }
    const a = target;
    if (a.dataset.plain !== undefined)
        return;
    const link = a.href;
    if (link === "")
        return;
    const url = new URL(link);
    if (url.origin !== document.location.origin)
        return;
    e.preventDefault();
    navigateTo(link, true);
}
function handlePopState(_) {
    navigateTo(document.location.href, false);
}
function getFormProperty(form, name) {
    // this is necessary as properties can be overriden by input fields, e.g.
    // if we have an input with id 'action' or 'method', it will override the
    // values of these form attributes...
    return Object.getOwnPropertyDescriptor(HTMLFormElement.prototype, name).get.call(form);
}
function handleOnSubmit(e) {
    return __awaiter(this, void 0, void 0, function* () {
        e.preventDefault();
        const form = e.target;
        // we create the form data
        const formData = new FormData(form);
        if (e.submitter) {
            const button = e.submitter;
            if (button.name !== "")
                formData.append(button.name, button.value);
        }
        let action = getFormProperty(form, 'action');
        let method = getFormProperty(form, 'method');
        const params = {
            method: method,
        };
        if (method == 'get') {
            // for a get request, we convert the formData to query parameters
            const url = new URL(action);
            // @ts-ignore
            url.search = (new URLSearchParams(formData)).toString();
            action = url.toString();
        }
        else {
            // for all other methods, we submit the form data in the request body
            params["body"] = formData;
        }
        const response = yield fetch(action, params);
        // we only push to history if we were redirected or if this is a 'get' form request...
        replaceDom(response.url, yield response.text(), response.redirected || method == 'get');
    });
}
function navigateTo(link, push) {
    return __awaiter(this, void 0, void 0, function* () {
        const response = yield fetch(link);
        replaceDom(response.url, yield response.text(), push || response.redirected);
    });
}
function replaceDom(link, text, push) {
    const doc = new DOMParser().parseFromString(text, "text/html");
    // we capture the scroll position
    const scrollX = window.scrollX;
    const scrollY = window.scrollY;
    document.replaceChild(doc.all[0], document.all[0]);
    // we restore the scroll position
    window.scroll(scrollX, scrollY);
    if (push) {
        history.pushState(null, "", link);
    }
    // we add the event handlers...
    initDocument();
    // we execute scripts...
    const scripts = document.getElementsByTagName("script");
    for (const script of scripts) {
        if (script.type === "" || script.type === "application/javascript") {
            // we try to execute the script...
            try {
                eval(script.innerText);
            }
            catch (e) {
                console.error(`Cannot execute script: ${e}`);
            }
        }
    }
}
function addEventListeners() {
    addEventListener('click', handleClick);
    addEventListener('popstate', handlePopState);
}
function initDocument() {
    const forms = document.getElementsByTagName("form");
    for (const form of forms) {
        console.log(`adding onSubmit handler to ${form.id}...`);
        form.onsubmit = handleOnSubmit;
    }
}
function initGospel() {
    initDocument();
    addEventListeners();
}
export function init() {
    document.addEventListener('DOMContentLoaded', initGospel, false);
}
//# sourceMappingURL=index.js.map