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
        console.log("not a link");
        return;
    }
    const link = target.href;
    console.log("Link:", link);
    e.preventDefault();
    navigateTo(link, true);
}
function handlePopState(_) {
    console.log(`Going back to ${document.location.href}`);
    navigateTo(document.location.href, false);
}
function navigateTo(link, push) {
    return __awaiter(this, void 0, void 0, function* () {
        const response = yield fetch(link);
        const doc = new DOMParser().parseFromString(yield response.text(), "text/html");
        document.replaceChild(doc.all[0], document.all[0]);
        if (push) {
            history.pushState(null, "", link);
        }
    });
}
function addEventListeners() {
    addEventListener('click', handleClick);
    addEventListener('popstate', handlePopState);
}
function initGospel() {
    console.log("initializing gospel");
    const submittables = document.querySelectorAll("[gospel-onSubmit]");
    for (const [_, submittable] of submittables.entries()) {
        console.log("adding onSubmit handler...");
        submittable.onsubmit = (_) => {
            console.log("submitting...");
            // e.preventDefault();
        };
    }
    addEventListeners();
}
export function init() {
    document.addEventListener('DOMContentLoaded', initGospel, false);
}
//# sourceMappingURL=index.js.map