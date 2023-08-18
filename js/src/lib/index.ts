function handleClick(e: Event){
	let target = e.target;
	if (target instanceof HTMLElement && target.tagName !== 'a'){
		target = target.closest('a');
	}

	if (target === null){
		return;
	}

	const a = target as HTMLAnchorElement

	if (a.dataset.plain !== undefined)
		return;

	const link = a.href;

	if (link === "")
		return;

	const url = new URL(link)

	if (url.origin !== document.location.origin)
		return;

	e.preventDefault();
	navigateTo(link, true);

}

function handlePopState(_: PopStateEvent){
	navigateTo(document.location.href, false);
}

async function handleOnSubmit(e: SubmitEvent) {

	e.preventDefault();

	const form = e.target as HTMLFormElement

	// we create the form data
	const formData = new FormData(form);

	if (e.submitter){

		const button = e.submitter as HTMLButtonElement

		if (button.name !== "")
			formData.append(button.name, button.value);
	}

	// we need this instead of using form.action as that can be overwritten
	// if a field named 'action' is present in the form...
	let action = Object.getOwnPropertyDescriptor(HTMLFormElement.prototype, 'action')!.get!.call(form)

	const params : RequestInit = {
		method: form.method,
	}

	if (form.method == 'get'){
		// for a get request, we convert the formData to query parameters
		const url = new URL(action);

		// @ts-ignore
		url.search = (new URLSearchParams(formData)).toString();

		action = url.toString();
	} else {
		// for all other methods, we submit the form data in the request body
		params["body"] = formData
	}

	const response = await fetch(action, params)

	// we only push to history if we were redirected or if this is a 'get' form request...
	replaceDom(response.url, await response.text(), response.redirected || form.method == 'get');

}

async function navigateTo(link: string, push: boolean){
	const response = await fetch(link);
	replaceDom(response.url, await response.text(), push || response.redirected);
}

function replaceDom(link: string, text: string, push: boolean) {

	const doc = new DOMParser().parseFromString(text, "text/html");

	// we capture the scroll position
	const scrollX = window.scrollX;
	const scrollY = window.scrollY;	

	document.replaceChild(doc.all[0], document.all[0]);

	// we restore the scroll position
	window.scroll(scrollX, scrollY);

	if (push){
		history.pushState(null, "", link);
	}

	// we add the event handlers...
	initDocument();

	// we execute scripts...

	const scripts = document.getElementsByTagName("script");

	for(const script of scripts){
		if (script.type === "" || script.type === "application/javascript"){
			// we try to execute the script...
			try {
				eval(script.innerText);
			} catch(e){
				console.error(`Cannot execute script: ${e}`)
			}
		}
	}

}

function addEventListeners(){

	addEventListener('click', handleClick);
	addEventListener('popstate', handlePopState);

}

function initDocument(){
	const forms = document.getElementsByTagName("form");

	for(const form of forms){
		console.log(`adding onSubmit handler to ${form.id}...`);
		(form as HTMLFormElement).onsubmit = handleOnSubmit;
	}
}


function initGospel() {
	initDocument();
	addEventListeners();
}

export function init() {

	document.addEventListener('DOMContentLoaded', initGospel, false);
}

