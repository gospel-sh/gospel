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
    const link = target.href;
    e.preventDefault();
    navigateTo(link, true);
}
function handlePopState(_) {
    navigateTo(document.location.href, false);
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
        // we need this instead of using form.action as that can be overwritten
        // if a field named 'action' is present in the form...
        const action = Object.getOwnPropertyDescriptor(HTMLFormElement.prototype, 'action').get.call(form);
        const response = yield fetch(action, {
            body: formData,
            method: form.method,
        });
        replaceDom(response.url, yield response.text(), response.redirected);
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
}
function addEventListeners() {
    addEventListener('click', handleClick);
    addEventListener('popstate', handlePopState);
}
function initDocument() {
    const submittables = document.querySelectorAll("[gospel-onSubmit]");
    for (const [_, submittable] of submittables.entries()) {
        console.log("adding onSubmit handler...");
        submittable.onsubmit = handleOnSubmit;
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